package pokeapi

// Species is the slice of /pokemon-species we need: where its evolution chain lives.
type Species struct {
	Name           string `json:"name"`
	EvolutionChain struct {
		URL string `json:"url"`
	} `json:"evolution_chain"`
}

// EvolutionChain is /evolution-chain/{id}.
type EvolutionChain struct {
	ID    int       `json:"id"`
	Chain ChainLink `json:"chain"`
}

// ChainLink is one species in the chain plus everything it evolves into.
// It contains a slice of itself, so one type covers chains of any depth
// and any number of branches (Eevee has 8).
type ChainLink struct {
	IsBaby           bool              `json:"is_baby"`
	Species          NamedAPIResource  `json:"species"`
	EvolutionDetails []EvolutionDetail `json:"evolution_details"`
	EvolvesTo        []ChainLink       `json:"evolves_to"`
}

// EvolutionDetail says how a species evolves. Fields that are often null
// in the JSON are pointers, so "not set" (nil) differs from zero.
type EvolutionDetail struct {
	IsDefault    bool              `json:"is_default"`
	Trigger      NamedAPIResource  `json:"trigger"`
	MinLevel     *int              `json:"min_level"`
	MinHappiness *int              `json:"min_happiness"`
	MinAffection *int              `json:"min_affection"`
	Item         *NamedAPIResource `json:"item"`
	HeldItem     *NamedAPIResource `json:"held_item"`
	KnownMove    *NamedAPIResource `json:"known_move"`
	Location     *NamedAPIResource `json:"location"`
	TimeOfDay    string            `json:"time_of_day"`
}

// GetSpecies takes the full URL from Pokemon.Species.URL.
func (c *Client) GetSpecies(url string) (Species, error) {
	return get[Species](c, url)
}

// GetEvolutionChain takes the full URL from Species.EvolutionChain.URL.
func (c *Client) GetEvolutionChain(url string) (EvolutionChain, error) {
	return get[EvolutionChain](c, url)
}
