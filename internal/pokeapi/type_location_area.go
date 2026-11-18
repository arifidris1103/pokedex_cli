package pokeapi

type LocationAreaResp struct {
	Count    int                `json:"count"`
	Next     *string            `json:"next"`
	Previous *string            `json:"previous"`
	Results  []NamedAPIResource `json:"results"`
}

// LocationArea keeps only the fields the app uses; the JSON has many more.
type LocationArea struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PokemonEncounters []struct {
		Pokemon NamedAPIResource `json:"pokemon"`
	} `json:"pokemon_encounters"`
}
