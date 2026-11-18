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
	if _, err := tea.NewProgram(ui.New(newBackend(&client, pokedex.New()))).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
