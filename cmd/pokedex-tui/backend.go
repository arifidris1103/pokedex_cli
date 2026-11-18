package main

// This file connects your pokeapi.Client and Pokédex to the UI's Backend
// interface. It's the only place the UI and your backend meet: each method
// calls the backend, converts the result into the UI's types, and returns.

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
	"github.com/arifidris1103/pokedex_cli/internal/pokedex"
	"github.com/arifidris1103/pokedex_cli/internal/ui"
)

type backend struct {
	client *pokeapi.Client
	dex    *pokedex.Pokedex
}

func newBackend(c *pokeapi.Client, dex *pokedex.Pokedex) *backend {
	return &backend{client: c, dex: dex}
}

// map / mapb
func (b *backend) ListLocationAreas(pageURL *string) (ui.LocationPage, error) {
	resp, err := b.client.ListLocationAreas(pageURL)
	if err != nil {
		return ui.LocationPage{}, err
	}
	areas := make([]string, len(resp.Results))
	for i, r := range resp.Results {
		areas[i] = r.Name
	}
	return ui.LocationPage{Areas: areas, Next: resp.Next, Prev: resp.Previous}, nil
}

// explore
func (b *backend) ExploreArea(area string) ([]string, error) {
	la, err := b.client.GetLocationArea(area)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(la.PokemonEncounters))
	for _, enc := range la.PokemonEncounters {
		names = append(names, enc.Pokemon.Name)
	}
	return names, nil
}

// Pokémon data, caught or not (the UI shows a silhouette until it's caught)
func (b *backend) GetPokemon(name string) (ui.Pokemon, error) {
	p, err := b.client.GetPokemon(name)
	if err != nil {
		return ui.Pokemon{}, err
	}
	return toUI(p), nil
}

// catch: same rule as the REPL, via pokedex.Roll
func (b *backend) Catch(name string) (bool, error) {
	p, err := b.client.GetPokemon(name)
	if err != nil {
		return false, err
	}
	if !pokedex.Roll(p.BaseExperience) {
		return false, nil
	}
	if err := b.dex.Add(p); err != nil {
		return true, fmt.Errorf("caught %s, but %w", name, err)
	}
	return true, nil
}

// inspect
func (b *backend) Inspect(name string) (ui.Pokemon, bool) {
	p, ok := b.dex.Get(name)
	if !ok {
		return ui.Pokemon{}, false
	}
	return toUI(p), true
}

// pokedex
func (b *backend) Pokedex() []ui.Pokemon {
	caught := b.dex.All()
	result := make([]ui.Pokemon, 0, len(caught))
	for _, p := range caught {
		result = append(result, toUI(p))
	}
	return result
}

// toUI converts your PokéAPI type into the UI's simpler type.
func toUI(p pokeapi.Pokemon) ui.Pokemon {
	types := make([]string, 0, len(p.Types))
	for _, t := range p.Types {
		types = append(types, t.Type.Name)
	}
	stats := make([]ui.Stat, 0, len(p.Stats))
	for _, s := range p.Stats {
		stats = append(stats, ui.Stat{Name: s.Stat.Name, Value: s.BaseStat})
	}
	return ui.Pokemon{
		ID:             p.ID,
		Name:           p.Name,
		BaseExperience: p.BaseExperience,
		Height:         p.Height,
		Weight:         p.Weight,
		Types:          types,
		Stats:          stats,
		SpriteURL:      p.Sprites.FrontDefault,
	}
}

// ── Optional extras ───────────────────────────────────────────────────────

var spriteClient = &http.Client{Timeout: 10 * time.Second}

// Sprite downloads a sprite PNG for the details panel.
func (b *backend) Sprite(url string) ([]byte, error) {
	resp, err := spriteClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch sprite: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// Evolutions returns the evolution line that contains this Pokémon:
// Pokémon → species → evolution chain → the branch with this species in it.
func (b *backend) Evolutions(name string) ([]ui.EvoStage, error) {
	p, err := b.client.GetPokemon(name)
	if err != nil {
		return nil, err
	}
	species, err := b.client.GetSpecies(p.Species.URL)
	if err != nil {
		return nil, err
	}
	chain, err := b.client.GetEvolutionChain(species.EvolutionChain.URL)
	if err != nil {
		return nil, err
	}
	// Search by species name: forms like "deoxys-normal" belong to species "deoxys".
	path, ok := findPath(chain.Chain, p.Species.Name)
	if !ok {
		return nil, fmt.Errorf("%s not found in its evolution chain", name)
	}
	return path, nil
}

// findPath walks the chain tree and returns the line that goes through
// target: everything before it, then on down its first branch.
// Chains can branch (Eevee has 8), so only the branch with target is kept.
func findPath(link pokeapi.ChainLink, target string) ([]ui.EvoStage, bool) {
	stage := ui.EvoStage{Name: link.Species.Name, How: evolutionHow(link.EvolutionDetails)}

	if link.Species.Name == target {
		path := []ui.EvoStage{stage}
		for len(link.EvolvesTo) > 0 {
			link = link.EvolvesTo[0]
			path = append(path, ui.EvoStage{Name: link.Species.Name, How: evolutionHow(link.EvolutionDetails)})
		}
		return path, true
	}

	for _, child := range link.EvolvesTo {
		if rest, ok := findPath(child, target); ok {
			return append([]ui.EvoStage{stage}, rest...), true
		}
	}
	return nil, false
}

// evolutionHow turns evolution details into short text like "Lv. 16",
// "Water Stone" or "Trade". The base form has no details, so it gets "".
func evolutionHow(details []pokeapi.EvolutionDetail) string {
	if len(details) == 0 {
		return ""
	}
	d := details[0]
	for _, candidate := range details {
		if candidate.IsDefault {
			d = candidate
			break
		}
	}

	switch d.Trigger.Name {
	case "level-up":
		switch {
		case d.MinLevel != nil:
			return fmt.Sprintf("Lv. %d", *d.MinLevel)
		case d.MinHappiness != nil:
			return "Friendship"
		case d.MinAffection != nil:
			return "Affection"
		case d.KnownMove != nil:
			return "Knows " + pretty(d.KnownMove.Name)
		case d.Location != nil:
			return "At " + pretty(d.Location.Name)
		case d.HeldItem != nil:
			return "Hold " + pretty(d.HeldItem.Name)
		case d.TimeOfDay != "":
			return "Level up (" + d.TimeOfDay + ")"
		default:
			return "Level up"
		}
	case "use-item":
		if d.Item != nil {
			return pretty(d.Item.Name)
		}
		return "Use item"
	case "trade":
		if d.HeldItem != nil {
			return "Trade w/ " + pretty(d.HeldItem.Name)
		}
		return "Trade"
	default:
		return pretty(d.Trigger.Name)
	}
}

// pretty turns "water-stone" into "Water Stone".
func pretty(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
