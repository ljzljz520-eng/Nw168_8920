package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"coffeeadvisor/internal/recommendation"
)

func main() {
	dataDir := flag.String("data", "./data", "directory containing the CSV data files")
	userID := flag.String("user", "demo", "user id to recommend for")
	limit := flag.Int("limit", 3, "maximum number of coffee bean recommendations")
	flag.Parse()

	service, err := recommendation.NewServiceFromCSV(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	report, err := service.Recommend(*userID, *limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
