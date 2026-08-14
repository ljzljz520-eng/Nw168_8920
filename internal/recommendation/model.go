package recommendation

type Bean struct {
	ID            string
	Name          string
	Flavors       []string
	BrewMethods   []string
	Acidity       float64
	Body          float64
	Sweetness     float64
	FilterPaperID string
	FilterPaper   string
	GrindID       string
	GrindService  string
}

type FlavorPreference struct {
	Flavor string
	Weight float64
}

type CuppingPreference struct {
	Acidity   float64
	Body      float64
	Sweetness float64
}

type Profile struct {
	UserID      string
	Flavors     []FlavorPreference
	Equipment   []string
	Repurchases map[string]int
	Favorites   map[string]bool
	Cupping     CuppingPreference
	HasCupping  bool
}

type Dataset struct {
	Beans    []Bean
	Profiles map[string]*Profile
}

type Recommendation struct {
	Rank     int      `json:"rank"`
	Kind     string   `json:"kind"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Score    float64  `json:"score"`
	Coverage float64  `json:"coverage"`
	Reasons  []string `json:"reasons"`
}

type Report struct {
	UserID string           `json:"user_id"`
	Beans  []Recommendation `json:"beans"`
	Extras []Recommendation `json:"extras"`
}
