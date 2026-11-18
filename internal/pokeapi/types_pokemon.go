package pokeapi

// Pokemon keeps only the fields the app uses. The full JSON is huge (every
// move in every game); fields left out here are simply skipped when decoding.
type Pokemon struct {
	ID             int              `json:"id"`
	Name           string           `json:"name"`
	BaseExperience int              `json:"base_experience"`
	Height         int              `json:"height"` // decimetres
	Weight         int              `json:"weight"` // hectograms
	Species        NamedAPIResource `json:"species"`
	Sprites        struct {
		FrontDefault string `json:"front_default"`
		FrontShiny   string `json:"front_shiny"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int              `json:"base_stat"`
		Stat     NamedAPIResource `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int              `json:"slot"`
		Type NamedAPIResource `json:"type"`
	} `json:"types"`
}
