// Package pokedex stores caught Pokémon and holds the catch rule.
// It's safe for concurrent use, so the REPL and the UI can share the code.
//
// A Pokédex opened with Open is saved to a JSON file after every catch, so
// catches survive a restart and the REPL and the UI share one Pokédex.
package pokedex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
)

type Pokedex struct {
	mu     sync.Mutex
	caught map[string]pokeapi.Pokemon
	path   string // where to save; empty means memory only
}

// New returns an empty Pokédex that lives only in memory.
func New() *Pokedex {
	return &Pokedex{caught: make(map[string]pokeapi.Pokemon)}
}

// Open loads the Pokédex saved at path (a missing file just means nothing
// caught yet). Every Add after this saves the file again.
func Open(path string) (*Pokedex, error) {
	caught, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return &Pokedex{caught: caught, path: path}, nil
}

// DefaultPath is $POKEDEX_FILE if set, otherwise ~/.pokedex.json.
func DefaultPath() (string, error) {
	if p := os.Getenv("POKEDEX_FILE"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pokedex.json"), nil
}

// OpenDefault opens the Pokédex at DefaultPath. If that fails, it still
// returns a working (memory-only) Pokédex, plus an error saying why your
// catches won't be saved.
func OpenDefault() (*Pokedex, error) {
	path, err := DefaultPath()
	if err == nil {
		var d *Pokedex
		if d, err = Open(path); err == nil {
			return d, nil
		}
	}
	return New(), fmt.Errorf("couldn't open your saved Pokedex (%v); catches this session won't be saved", err)
}

// Add stores a caught Pokémon and, for a Pokédex from Open, saves the file.
// If saving fails, the Pokémon is still caught for this session.
func (d *Pokedex) Add(p pokeapi.Pokemon) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.caught[p.Name] = p
	if d.path == "" {
		return nil
	}
	if err := d.save(); err != nil {
		return fmt.Errorf("couldn't save your Pokedex: %w", err)
	}
	return nil
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

// save writes the Pokédex to d.path. The caller must hold d.mu.
//
// It first merges in whatever is already in the file, so if the REPL and the
// UI are open at the same time, neither wipes out the other's catches. It
// writes to a temporary file and renames it into place, so a crash halfway
// through can never leave a half-written (corrupt) Pokédex behind.
func (d *Pokedex) save() error {
	onDisk, err := readFile(d.path)
	if err != nil {
		return err // don't overwrite a file we can't read
	}
	for name, p := range d.caught {
		onDisk[name] = p
	}
	d.caught = onDisk

	data, err := json.MarshalIndent(d.caught, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(d.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".pokedex-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), d.path)
}

// readFile loads a saved Pokédex; a missing file is an empty Pokédex.
func readFile(path string) (map[string]pokeapi.Pokemon, error) {
	caught := make(map[string]pokeapi.Pokemon)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return caught, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &caught); err != nil {
		return nil, fmt.Errorf("%s is not a valid Pokedex file: %w", path, err)
	}
	return caught, nil
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
