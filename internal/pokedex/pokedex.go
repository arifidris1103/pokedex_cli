// Package pokedex stores caught Pokémon and holds the catch rule.
// It's safe for concurrent use, so the REPL and the UI can share the code.
package pokedex

import (
	"math/rand/v2"
	"slices"
	"sync"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
)

type Pokedex struct {
	mu     sync.Mutex
	caught map[string]pokeapi.Pokemon
}

func New() *Pokedex {
	return &Pokedex{caught: make(map[string]pokeapi.Pokemon)}
}

func (d *Pokedex) Add(p pokeapi.Pokemon) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.caught[p.Name] = p
}

// Get returns the caught Pokémon with the given name, if any.
func (d *Pokedex) Get(name string) (pokeapi.Pokemon, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	p, ok := d.caught[name]
	return p, ok
}

// All returns every caught Pokémon, sorted by Pokédex number.
func (d *Pokedex) All() []pokeapi.Pokemon {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]pokeapi.Pokemon, 0, len(d.caught))
	for _, p := range d.caught {
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b pokeapi.Pokemon) int { return a.ID - b.ID })
	return out
}

// CatchChance is the percent chance (5–95) of catching a Pokémon:
// the higher its base experience, the harder it is.
func CatchChance(baseExperience int) int {
	return max(5, min(95, 100-baseExperience/4))
}

// Roll throws a Pokéball: true means caught.
func Roll(baseExperience int) bool {
	return rand.IntN(100) < CatchChance(baseExperience)
}
