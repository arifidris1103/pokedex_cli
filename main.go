package main

import (
	"fmt"
	"os"
	"time"

	"github.com/arifidris1103/pokedex_cli/internal/pokeapi"
	"github.com/arifidris1103/pokedex_cli/internal/pokedex"
)

type config struct {
	pokeapiClient       pokeapi.Client
	pokedex             *pokedex.Pokedex
	nextLocationAreaURL *string
	prevLocationAreaURL *string
}

func main() {
	dex, err := pokedex.OpenDefault()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}
	cfg := config{
		pokeapiClient: pokeapi.NewClient(time.Hour),
		pokedex:       dex,
	}
	startRepl(&cfg)
}
