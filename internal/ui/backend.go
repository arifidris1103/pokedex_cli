// Package ui is a Bubble Tea terminal UI for the Pokédex.
//
// The UI never talks to PokéAPI itself. Everything goes through the Backend
// interface below, which mirrors the REPL commands (map, mapb, explore,
// catch, inspect, pokedex). Your backend implements it; the UI just draws.
package ui

import "errors"

// ErrNotImplemented lets a backend method say "not built yet".
// The UI shows a friendly message instead of crashing.
var ErrNotImplemented = errors.New("not implemented yet")

// LocationPage is one page of location areas (the `map` / `mapb` commands).
type LocationPage struct {
	Areas []string
	Next  *string // URL of the next page, nil on the last page
	Prev  *string // URL of the previous page, nil on the first page
}

// Stat is one base stat, e.g. {"hp", 45}.
type Stat struct {
	Name  string
	Value int
}

// Pokemon is everything the details panel can show.
type Pokemon struct {
	ID             int
	Name           string
	BaseExperience int
	Height         int // decimetres, as PokéAPI returns it
	Weight         int // hectograms, as PokéAPI returns it
	Types          []string
	Stats          []Stat
	SpriteURL      string // sprites.front_default
}

// Backend is what the UI needs from you. Each method maps to a REPL command.
//
// The UI calls these from background goroutines, so implementations must be
// safe for concurrent use (a mutex around shared maps is enough), and must
// not print to stdout: stray output corrupts the screen.
type Backend interface {
	// ListLocationAreas returns a page of areas. pageURL is nil for the
	// first page, otherwise LocationPage.Next or LocationPage.Prev. (map / mapb)
	ListLocationAreas(pageURL *string) (LocationPage, error)

	// ExploreArea returns the names of the Pokémon found in an area. (explore)
	ExploreArea(area string) ([]string, error)

	// GetPokemon fetches a Pokémon's data, caught or not. The UI uses it to
	// show a silhouette before you catch it.
	GetPokemon(name string) (Pokemon, error)

	// Catch throws a Pokéball. It returns true if the Pokémon was caught,
	// and a caught Pokémon must then appear in Pokedex(). (catch)
	Catch(name string) (bool, error)

	// Inspect returns a caught Pokémon; ok is false if it isn't caught. (inspect)
	Inspect(name string) (p Pokemon, ok bool)

	// Pokedex lists every caught Pokémon. (pokedex)
	Pokedex() []Pokemon
}

// ── Optional extras ───────────────────────────────────────────────────────
// Implement these on the same type and the UI picks them up automatically.

// SpriteFetcher downloads image bytes (PNG) for a sprite URL.
// Without it the details panel shows no picture.
type SpriteFetcher interface {
	Sprite(url string) ([]byte, error)
}

// EvoStage is one step of an evolution line.
type EvoStage struct {
	Name string
	How  string // e.g. "Lv. 16", "Water Stone", "Trade"; empty for the first stage
}

// EvolutionProvider returns the evolution line a Pokémon belongs to,
// e.g. bulbasaur → [bulbasaur, ivysaur (Lv. 16), venusaur (Lv. 32)].
type EvolutionProvider interface {
	Evolutions(name string) ([]EvoStage, error)
}
