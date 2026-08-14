package recommendation

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

type Service struct {
	dataset Dataset
}

func NewService(dataset Dataset) *Service {
	return &Service{dataset: dataset}
}

func NewServiceFromCSV(dir string) (*Service, error) {
	dataset, err := LoadDataset(dir)
	if err != nil {
		return nil, err
	}
	return NewService(dataset), nil
}

func (s *Service) Recommend(userID string, limit int) (Report, error) {
	profile := s.dataset.Profiles[userID]
	if profile == nil {
		return Report{}, fmt.Errorf("user %q has no preference data", userID)
	}
	if limit <= 0 {
		return Report{}, fmt.Errorf("recommendation limit must be positive")
	}

	ranked := make([]Recommendation, 0, len(s.dataset.Beans))
	for _, bean := range s.dataset.Beans {
		recommendation, matched := scoreBean(*profile, bean)
		if matched {
			ranked = append(ranked, recommendation)
		}
	}
	if len(ranked) == 0 {
		return Report{UserID: userID, Beans: []Recommendation{}, Extras: []Recommendation{}}, nil
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].ID < ranked[j].ID
		}
		return ranked[i].Score > ranked[j].Score
	})
	for i := range ranked {
		ranked[i].Rank = i + 1
	}

	extras := companionRecommendations(*profile, ranked[0], s.dataset.Beans)
	return Report{
		UserID: userID,
		Beans:  takeTop(ranked, limit),
		Extras: extras,
	}, nil
}

func scoreBean(profile Profile, bean Bean) (Recommendation, bool) {
	var score float64
	var matchedFlavorWeight float64
	var totalFlavorWeight float64
	var reasons []string

	for _, preference := range profile.Flavors {
		totalFlavorWeight += preference.Weight
		if slices.Contains(bean.Flavors, preference.Flavor) {
			matchedFlavorWeight += preference.Weight
			score += preference.Weight * 10
			reasons = append(reasons, "matches flavor: "+preference.Flavor)
		}
	}

	equipmentMatched := intersects(profile.Equipment, bean.BrewMethods)
	if equipmentMatched {
		score += 12
		reasons = append(reasons, "works with equipment: "+strings.Join(profile.Equipment, ", "))
	}

	repurchaseCount := profile.Repurchases[bean.ID]
	if repurchaseCount > 0 {
		score += float64(repurchaseCount) * 8
		reasons = append(reasons, fmt.Sprintf("repurchased %d times", repurchaseCount))
	}

	favorite := profile.Favorites[bean.ID]
	if favorite {
		score += 20
		reasons = append(reasons, "previously saved as a favorite")
	}

	cuppingSimilarity := 0.0
	if profile.HasCupping {
		cuppingSimilarity = (similarity(profile.Cupping.Acidity, bean.Acidity) +
			similarity(profile.Cupping.Body, bean.Body) +
			similarity(profile.Cupping.Sweetness, bean.Sweetness)) / 3
		score += cuppingSimilarity * 15
		reasons = append(reasons, fmt.Sprintf("cupping profile match: %.0f%%", cuppingSimilarity*100))
	}

	matched := matchedFlavorWeight > 0 && equipmentMatched
	if !matched {
		return Recommendation{}, false
	}

	covered := matchedFlavorWeight
	available := totalFlavorWeight
	if len(profile.Equipment) > 0 {
		available++
		if equipmentMatched {
			covered++
		}
	}
	if profile.HasCupping {
		available += 3
		covered += cuppingSimilarity * 3
	}
	if len(profile.Repurchases) > 0 {
		available++
		if repurchaseCount > 0 {
			covered++
		}
	}
	if len(profile.Favorites) > 0 {
		available++
		if favorite {
			covered++
		}
	}

	return Recommendation{
		Kind:     "coffee_bean",
		ID:       bean.ID,
		Name:     bean.Name,
		Score:    round(score),
		Coverage: round(covered / available * 100),
		Reasons:  reasons,
	}, true
}

func companionRecommendations(profile Profile, top Recommendation, beans []Bean) []Recommendation {
	var selected Bean
	for _, bean := range beans {
		if bean.ID == top.ID {
			selected = bean
			break
		}
	}

	extras := make([]Recommendation, 0, 2)
	if selected.FilterPaperID != "" && selected.FilterPaper != "" {
		extras = append(extras, Recommendation{
			Rank:     len(extras) + 1,
			Kind:     "filter_paper",
			ID:       selected.FilterPaperID,
			Name:     selected.FilterPaper,
			Score:    top.Score,
			Coverage: top.Coverage,
			Reasons:  []string{"paired with " + selected.Name + " for " + strings.Join(profile.Equipment, ", ")},
		})
	}
	if selected.GrindID != "" && selected.GrindService != "" {
		extras = append(extras, Recommendation{
			Rank:     len(extras) + 1,
			Kind:     "grind_service",
			ID:       selected.GrindID,
			Name:     selected.GrindService,
			Score:    top.Score,
			Coverage: top.Coverage,
			Reasons:  []string{"ground for " + strings.Join(profile.Equipment, ", ") + " and " + selected.Name},
		})
	}
	return extras
}

func takeTop(candidates []Recommendation, limit int) []Recommendation {
	end := min(limit, len(candidates))
	return candidates[:end]
}

func intersects(left, right []string) bool {
	for _, item := range left {
		if slices.Contains(right, item) {
			return true
		}
	}
	return false
}

func similarity(want, got float64) float64 {
	return max(0, 1-math.Abs(want-got)/10)
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}
