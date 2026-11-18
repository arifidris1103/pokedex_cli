package pokedex

import (
	"os"
	"path/filepath"
	"slices"
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

func TestSaveAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")

	d, err := Open(path) // no file yet: starts empty
	if err != nil || len(d.All()) != 0 {
		t.Fatalf("new Pokedex: %v, %d caught", err, len(d.All()))
	}
	if err := d.Add(pokeapi.Pokemon{ID: 25, Name: "pikachu", BaseExperience: 112}); err != nil {
		t.Fatal(err)
	}

	again, err := Open(path) // like restarting the program
	if err != nil {
		t.Fatal(err)
	}
	p, ok := again.Get("pikachu")
	if !ok || p.ID != 25 || p.BaseExperience != 112 {
		t.Fatalf("pikachu after reopening: %+v, %v", p, ok)
	}
}

// The REPL and the UI open at the same time must not erase each other's catches.
func TestTwoProgramsAtOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	repl, _ := Open(path)
	tui, _ := Open(path)

	if err := repl.Add(pokeapi.Pokemon{ID: 129, Name: "magikarp"}); err != nil {
		t.Fatal(err)
	}
	if err := tui.Add(pokeapi.Pokemon{ID: 120, Name: "staryu"}); err != nil {
		t.Fatal(err)
	}

	d, _ := Open(path)
	if got := names(d.All()); !slices.Equal(got, []string{"staryu", "magikarp"}) {
		t.Fatalf("saved file has %v", got)
	}
}

func TestCorruptFileIsNotOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pokedex.json")
	os.WriteFile(path, []byte("{not json"), 0o644)

	if _, err := Open(path); err == nil {
		t.Fatal("expected an error for a corrupt file")
	}
	t.Setenv("POKEDEX_FILE", path)
	d, err := OpenDefault()
	if err == nil {
		t.Fatal("OpenDefault should report the corrupt file")
	}
	d.Add(pokeapi.Pokemon{ID: 1, Name: "bulbasaur"}) // memory only
	if data, _ := os.ReadFile(path); string(data) != "{not json" {
		t.Fatalf("corrupt file was overwritten: %q", data)
	}
	if _, ok := d.Get("bulbasaur"); !ok {
		t.Fatal("catching should still work for this session")
	}
}

func TestDefaultPath(t *testing.T) {
	t.Setenv("POKEDEX_FILE", "/tmp/custom.json")
	if p, _ := DefaultPath(); p != "/tmp/custom.json" {
		t.Fatalf("got %q", p)
	}
	t.Setenv("POKEDEX_FILE", "")
	home, _ := os.UserHomeDir()
	if p, _ := DefaultPath(); p != filepath.Join(home, ".pokedex.json") {
		t.Fatalf("got %q", p)
	}
}
