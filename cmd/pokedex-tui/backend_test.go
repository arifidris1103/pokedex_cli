package main

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
	"github.com/arifidris1103/pokedex_cli/internal/ui"
)

// chain builds a ChainLink from compact JSON, the same shape PokéAPI returns.
func chain(t *testing.T, js string) pokeapi.ChainLink {
	t.Helper()
	var c pokeapi.ChainLink
	if err := json.Unmarshal([]byte(js), &c); err != nil {
		t.Fatal(err)
	}
	return c
}

const bulbasaurChain = `{"species":{"name":"bulbasaur"},"evolution_details":[],"evolves_to":[
	{"species":{"name":"ivysaur"},"evolution_details":[{"min_level":16,"trigger":{"name":"level-up"}}],"evolves_to":[
		{"species":{"name":"venusaur"},"evolution_details":[{"min_level":32,"trigger":{"name":"level-up"}}],"evolves_to":[]}]}]}`

const pichuChain = `{"species":{"name":"pichu"},"evolution_details":[],"evolves_to":[
	{"species":{"name":"pikachu"},"evolution_details":[{"min_happiness":220,"trigger":{"name":"level-up"}}],"evolves_to":[
		{"species":{"name":"raichu"},"evolution_details":[{"item":{"name":"thunder-stone"},"trigger":{"name":"use-item"}}],"evolves_to":[]}]}]}`

const eeveeChain = `{"species":{"name":"eevee"},"evolution_details":[],"evolves_to":[
	{"species":{"name":"vaporeon"},"evolution_details":[{"item":{"name":"water-stone"},"trigger":{"name":"use-item"}}],"evolves_to":[]},
	{"species":{"name":"jolteon"},"evolution_details":[{"item":{"name":"thunder-stone"},"trigger":{"name":"use-item"}}],"evolves_to":[]},
	{"species":{"name":"umbreon"},"evolution_details":[{"min_happiness":160,"time_of_day":"night","trigger":{"name":"level-up"}}],"evolves_to":[]}]}`

func stages(names ...string) []ui.EvoStage {
	out := make([]ui.EvoStage, len(names)/2)
	for i := range out {
		out[i] = ui.EvoStage{Name: names[2*i], How: names[2*i+1]}
	}
	return out
}

func TestFindPath(t *testing.T) {
	cases := []struct {
		name, chain, target string
		want                []ui.EvoStage
	}{
		{"first stage shows the whole line", bulbasaurChain, "bulbasaur",
			stages("bulbasaur", "", "ivysaur", "Lv. 16", "venusaur", "Lv. 32")},
		{"last stage", bulbasaurChain, "venusaur",
			stages("bulbasaur", "", "ivysaur", "Lv. 16", "venusaur", "Lv. 32")},
		{"middle stage", pichuChain, "pikachu",
			stages("pichu", "", "pikachu", "Friendship", "raichu", "Thunder Stone")},
		{"branch: only jolteon's branch", eeveeChain, "jolteon",
			stages("eevee", "", "jolteon", "Thunder Stone")},
		{"branch root follows the first branch", eeveeChain, "eevee",
			stages("eevee", "", "vaporeon", "Water Stone")},
		{"friendship beats time of day", eeveeChain, "umbreon",
			stages("eevee", "", "umbreon", "Friendship")},
	}
	for _, c := range cases {
		got, ok := findPath(chain(t, c.chain), c.target)
		if !ok || !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v (found=%v), want %v", c.name, got, ok, c.want)
		}
	}

	if _, ok := findPath(chain(t, bulbasaurChain), "pikachu"); ok {
		t.Error("pikachu is not in bulbasaur's chain")
	}
}

func TestEvolutionHowTrade(t *testing.T) {
	var d []pokeapi.EvolutionDetail
	js := `[{"trigger":{"name":"trade"},"held_item":{"name":"metal-coat"}}]`
	if err := json.Unmarshal([]byte(js), &d); err != nil {
		t.Fatal(err)
	}
	if got := evolutionHow(d); got != "Trade w/ Metal Coat" {
		t.Fatalf("got %q", got)
	}
}
