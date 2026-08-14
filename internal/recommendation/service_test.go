package recommendation_test

import (
	"os"
	"path/filepath"
	"testing"

	"coffeeadvisor/internal/recommendation"
)

func TestServiceReturnsThreeRankedCoffeeBeanRecommendations(t *testing.T) {
	service := serviceFromFixture(t)

	report, err := service.Recommend("buyer-1", 3)
	if err != nil {
		t.Fatalf("recommend for buyer-1: %v", err)
	}
	if got, want := len(report.Beans), 3; got != want {
		t.Fatalf("got %d coffee bean recommendations, want %d", got, want)
	}

	wantNames := []string{"Aurora Highlands", "Solstice Estate", "Harbor Decaf"}
	for i, want := range wantNames {
		if got := report.Beans[i].Name; got != want {
			t.Errorf("recommendation %d is %q, want %q", i+1, got, want)
		}
		if got, wantRank := report.Beans[i].Rank, i+1; got != wantRank {
			t.Errorf("%q has rank %d, want %d", report.Beans[i].Name, got, wantRank)
		}
		if i > 0 && report.Beans[i-1].Score < report.Beans[i].Score {
			t.Errorf("recommendation %d has score %.2f after score %.2f", i+1, report.Beans[i].Score, report.Beans[i-1].Score)
		}
	}
}

func TestServiceReturnsReasonsCoverageAndCompanionProducts(t *testing.T) {
	service := serviceFromFixture(t)

	report, err := service.Recommend("buyer-1", 3)
	if err != nil {
		t.Fatalf("recommend for buyer-1: %v", err)
	}
	for _, bean := range report.Beans {
		if len(bean.Reasons) == 0 {
			t.Errorf("%q has no recommendation reasons", bean.Name)
		}
		if bean.Coverage <= 0 || bean.Coverage > 100 {
			t.Errorf("%q has coverage %.2f, want a value above 0 and at most 100", bean.Name, bean.Coverage)
		}
	}
	if got, want := len(report.Extras), 2; got != want {
		t.Fatalf("got %d companion product recommendations, want %d", got, want)
	}
	wantKinds := []string{"filter_paper", "grind_service"}
	for i, want := range wantKinds {
		if got := report.Extras[i].Kind; got != want {
			t.Errorf("companion recommendation %d has kind %q, want %q", i+1, got, want)
		}
		if len(report.Extras[i].Reasons) == 0 {
			t.Errorf("companion recommendation %d has no recommendation reasons", i+1)
		}
	}
}

func serviceFromFixture(t *testing.T) *recommendation.Service {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"beans.csv": "id,name,flavors,brew_methods,acidity,body,sweetness,filter_paper_id,filter_paper,grind_id,grind_service\n" +
			"aurora,Aurora Highlands,berry;citrus,v60;aeropress,7,5,8,v60-02,V60 02 Paper,grind-v60,V60 Medium-Fine Grind\n" +
			"solstice,Solstice Estate,berry;chocolate,v60;chemex,6,6,7,v60-02,V60 02 Paper,grind-v60,V60 Medium-Fine Grind\n" +
			"harbor,Harbor Decaf,citrus;chocolate,v60;french-press,5,7,6,v60-02,V60 02 Paper,grind-v60,V60 Medium-Fine Grind\n",
		"flavors.csv": "user_id,flavor,weight\n" +
			"buyer-1,berry,5\n" +
			"buyer-1,citrus,4\n" +
			"buyer-1,chocolate,2\n",
		"equipment.csv": "user_id,equipment\n" +
			"buyer-1,v60\n",
		"repurchases.csv": "user_id,bean_id,count\n" +
			"buyer-1,aurora,3\n" +
			"buyer-1,solstice,1\n",
		"favorites.csv": "user_id,bean_id\n" +
			"buyer-1,aurora\n",
		"cupping_preferences.csv": "user_id,acidity,body,sweetness\n" +
			"buyer-1,7,5,8\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	service, err := recommendation.NewServiceFromCSV(dir)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return service
}
