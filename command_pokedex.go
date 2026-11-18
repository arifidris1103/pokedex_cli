package main

import "fmt"

func callbackPokedex(cfg *config, args ...string) error {
	caught := cfg.pokedex.All()
	if len(caught) == 0 {
		fmt.Println("Your Pokedex is empty. Go catch something!")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for _, p := range caught {
		fmt.Printf(" - %s\n", p.Name)
	}
	return nil
}
