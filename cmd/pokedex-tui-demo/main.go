// Command pokedex-tui-demo runs the UI on an offline demo backend.
//
//	go run ./cmd/pokedex-tui-demo
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/arifidris1103/pokedex_cli/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.New(ui.NewDemoBackend())).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
