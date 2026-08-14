package recommendation

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func LoadDataset(dir string) (Dataset, error) {
	dataset := Dataset{Profiles: make(map[string]*Profile)}

	beans, err := loadBeans(filepath.Join(dir, "beans.csv"))
	if err != nil {
		return Dataset{}, err
	}
	dataset.Beans = beans

	loaders := []func(string, *Dataset) error{
		loadFlavors,
		loadEquipment,
		loadRepurchases,
		loadFavorites,
		loadCuppingPreferences,
	}
	files := []string{
		"flavors.csv",
		"equipment.csv",
		"repurchases.csv",
		"favorites.csv",
		"cupping_preferences.csv",
	}
	for i, load := range loaders {
		if err := load(filepath.Join(dir, files[i]), &dataset); err != nil {
			return Dataset{}, err
		}
	}

	return dataset, nil
}

func loadBeans(path string) ([]Bean, error) {
	records, err := readCSV(path, []string{
		"id", "name", "flavors", "brew_methods", "acidity", "body", "sweetness",
		"filter_paper_id", "filter_paper", "grind_id", "grind_service",
	})
	if err != nil {
		return nil, err
	}

	beans := make([]Bean, 0, len(records))
	seen := make(map[string]bool, len(records))
	for row, record := range records {
		id := normalize(record[0])
		if id == "" || strings.TrimSpace(record[1]) == "" {
			return nil, fmt.Errorf("%s row %d: bean id and name are required", path, row+2)
		}
		if seen[id] {
			return nil, fmt.Errorf("%s row %d: duplicate bean id %q", path, row+2, id)
		}
		seen[id] = true
		acidity, err := parseFloat(path, row, "acidity", record[4])
		if err != nil {
			return nil, err
		}
		body, err := parseFloat(path, row, "body", record[5])
		if err != nil {
			return nil, err
		}
		sweetness, err := parseFloat(path, row, "sweetness", record[6])
		if err != nil {
			return nil, err
		}
		beans = append(beans, Bean{
			ID:            id,
			Name:          strings.TrimSpace(record[1]),
			Flavors:       splitList(record[2]),
			BrewMethods:   splitList(record[3]),
			Acidity:       acidity,
			Body:          body,
			Sweetness:     sweetness,
			FilterPaperID: normalize(record[7]),
			FilterPaper:   strings.TrimSpace(record[8]),
			GrindID:       normalize(record[9]),
			GrindService:  strings.TrimSpace(record[10]),
		})
	}
	return beans, nil
}

func loadFlavors(path string, dataset *Dataset) error {
	records, err := readCSV(path, []string{"user_id", "flavor", "weight"})
	if err != nil {
		return err
	}
	for row, record := range records {
		profile, err := profileFor(path, row, record[0], dataset)
		if err != nil {
			return err
		}
		weight, err := parseFloat(path, row, "weight", record[2])
		if err != nil {
			return err
		}
		flavor := normalize(record[1])
		if flavor == "" || weight <= 0 {
			return fmt.Errorf("%s row %d: flavor and a positive weight are required", path, row+2)
		}
		profile.Flavors = append(profile.Flavors, FlavorPreference{Flavor: flavor, Weight: weight})
	}
	return nil
}

func loadEquipment(path string, dataset *Dataset) error {
	records, err := readCSV(path, []string{"user_id", "equipment"})
	if err != nil {
		return err
	}
	for row, record := range records {
		profile, err := profileFor(path, row, record[0], dataset)
		if err != nil {
			return err
		}
		equipment := normalize(record[1])
		if equipment == "" {
			return fmt.Errorf("%s row %d: equipment is required", path, row+2)
		}
		profile.Equipment = append(profile.Equipment, equipment)
	}
	return nil
}

func loadRepurchases(path string, dataset *Dataset) error {
	records, err := readCSV(path, []string{"user_id", "bean_id", "count"})
	if err != nil {
		return err
	}
	for row, record := range records {
		profile, err := profileFor(path, row, record[0], dataset)
		if err != nil {
			return err
		}
		count, err := strconv.Atoi(strings.TrimSpace(record[2]))
		if err != nil || count < 0 {
			return fmt.Errorf("%s row %d: count must be a non-negative integer", path, row+2)
		}
		beanID := normalize(record[1])
		if beanID == "" {
			return fmt.Errorf("%s row %d: bean id is required", path, row+2)
		}
		profile.Repurchases[beanID] += count
	}
	return nil
}

func loadFavorites(path string, dataset *Dataset) error {
	records, err := readCSV(path, []string{"user_id", "bean_id"})
	if err != nil {
		return err
	}
	for row, record := range records {
		profile, err := profileFor(path, row, record[0], dataset)
		if err != nil {
			return err
		}
		beanID := normalize(record[1])
		if beanID == "" {
			return fmt.Errorf("%s row %d: bean id is required", path, row+2)
		}
		profile.Favorites[beanID] = true
	}
	return nil
}

func loadCuppingPreferences(path string, dataset *Dataset) error {
	records, err := readCSV(path, []string{"user_id", "acidity", "body", "sweetness"})
	if err != nil {
		return err
	}
	for row, record := range records {
		profile, err := profileFor(path, row, record[0], dataset)
		if err != nil {
			return err
		}
		acidity, err := parseFloat(path, row, "acidity", record[1])
		if err != nil {
			return err
		}
		body, err := parseFloat(path, row, "body", record[2])
		if err != nil {
			return err
		}
		sweetness, err := parseFloat(path, row, "sweetness", record[3])
		if err != nil {
			return err
		}
		profile.Cupping = CuppingPreference{Acidity: acidity, Body: body, Sweetness: sweetness}
		profile.HasCupping = true
	}
	return nil
}

func readCSV(path string, expectedHeader []string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = len(expectedHeader)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read %s header: %w", path, err)
	}
	for i := range expectedHeader {
		if strings.TrimSpace(header[i]) != expectedHeader[i] {
			return nil, fmt.Errorf("%s column %d: got %q, want %q", path, i+1, header[i], expectedHeader[i])
		}
	}

	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		records = append(records, record)
	}
	return records, nil
}

func profileFor(path string, row int, userID string, dataset *Dataset) (*Profile, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("%s row %d: user id is required", path, row+2)
	}
	profile := dataset.Profiles[userID]
	if profile == nil {
		profile = &Profile{
			UserID:      userID,
			Repurchases: make(map[string]int),
			Favorites:   make(map[string]bool),
		}
		dataset.Profiles[userID] = profile
	}
	return profile, nil
}

func parseFloat(path string, row int, field, value string) (float64, error) {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || parsed < 0 || parsed > 10 {
		return 0, fmt.Errorf("%s row %d: %s must be a number from 0 to 10", path, row+2, field)
	}
	return parsed, nil
}

func splitList(value string) []string {
	parts := strings.Split(value, ";")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := normalize(part); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
