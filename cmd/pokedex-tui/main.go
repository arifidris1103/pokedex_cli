// Command pokedex-tui runs the UI on your real backend (internal/pokeapi).
//
//	go run ./cmd/pokedex-tui
package main

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
	"github.com/arifidris1103/pokedex_cli/internal/pokedex"
	"github.com/arifidris1103/pokedex_cli/internal/ui"
)

func main() {
	client := pokeapi.NewClient(time.Hour)
	dex, err := pokedex.OpenDefault()
	if err != nil {
		// Shown after you quit; the UI covers the screen while it runs.
		defer fmt.Fprintln(os.Stderr, "warning:", err)
	}
	if _, err := tea.NewProgram(ui.New(newBackend(&client, dex))).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
