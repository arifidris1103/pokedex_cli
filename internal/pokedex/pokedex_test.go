package pokedex

import (
	"testing"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
)

func TestCatchChance(t *testing.T) {
	cases := []struct {
		name    string
		baseExp int
		want    int
	}{
		{"magikarp", 40, 90},
		{"bulbasaur", 64, 84},
		{"gyarados", 189, 53},
		{"chansey", 395, 5}, // would be 1% without the floor
		{"blissey", 608, 5}, // would be negative without the floor
		{"missing base exp", 0, 95},
	}
	for _, c := range cases {
		if got := CatchChance(c.baseExp); got != c.want {
			t.Errorf("%s (%d): got %d%%, want %d%%", c.name, c.baseExp, got, c.want)
		}
	}
}

func TestRollMatchesChance(t *testing.T) {
	const throws = 20000
	caught := 0
	for range throws {
		if Roll(189) { // gyarados: 53%
			caught++
		}
	}
	if pct := caught * 100 / throws; pct < 50 || pct > 56 {
		t.Errorf("caught %d%% of the time, want about 53%%", pct)
	}
}

func TestPokedexSortedByID(t *testing.T) {
	d := New()
	d.Add(pokeapi.Pokemon{ID: 129, Name: "magikarp"})
	d.Add(pokeapi.Pokemon{ID: 1, Name: "bulbasaur"})
	d.Add(pokeapi.Pokemon{ID: 25, Name: "pikachu"})

	if _, ok := d.Get("pikachu"); !ok {
		t.Fatal("pikachu should be caught")
	}
	if _, ok := d.Get("mew"); ok {
		t.Fatal("mew should not be caught")
	}

	all := d.All()
	want := []string{"bulbasaur", "pikachu", "magikarp"}
	for i, p := range all {
		if p.Name != want[i] {
			t.Fatalf("got order %v, want %v", names(all), want)
		}
	}
}

func names(ps []pokeapi.Pokemon) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Name
	}
	return out
}
