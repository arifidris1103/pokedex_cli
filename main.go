package main

import (
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
	cfg := config{
		pokeapiClient: pokeapi.NewClient(time.Hour),
		pokedex:       pokedex.New(),
	}
	startRepl(&cfg)
}
