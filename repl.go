package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	usage       string
	description string
	callback    func(*config, ...string) error
}

// getCommands lists every command in the order `help` shows them.
func getCommands() []cliCommand {
	return []cliCommand{
		{"help", "help", "Prints the help menu", callbackHelp},
		{"map", "map", "Lists the next page of location areas", callbackMap},
		{"mapb", "mapb", "Lists the previous page of location areas", callbackMapb},
		{"explore", "explore <location_area>", "Lists the pokemon in an area", callbackExplore},
		{"catch", "catch <pokemon>", "Throws a Pokeball at a pokemon", callbackCatch},
		{"inspect", "inspect <pokemon>", "Shows a pokemon you've caught", callbackInspect},
		{"pokedex", "pokedex", "Lists every pokemon you've caught", callbackPokedex},
		{"exit", "exit", "Turns off the Pokedex", callbackExit},
	}
}

func findCommand(name string) (cliCommand, bool) {
	for _, c := range getCommands() {
		if c.name == name {
			return c, true
		}
	}
	return cliCommand{}, false
}

func startRepl(cfg *config) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			fmt.Println()
			return
		}

		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}

		command, ok := findCommand(words[0])
		if !ok {
			fmt.Println("invalid command")
			continue
		}
		if err := command.callback(cfg, words[1:]...); err != nil {
			fmt.Println(err)
		}
	}
}

func cleanInput(text string) []string {
	return strings.Fields(strings.ToLower(text))
}
