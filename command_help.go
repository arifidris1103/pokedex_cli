package main

import "fmt"

func callbackHelp(cfg *config, args ...string) error {
	fmt.Println()
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, c := range getCommands() {
		fmt.Printf("  %-24s %s\n", c.usage, c.description)
	}
	fmt.Println()
	return nil
}
