package main

import (
	"errors"
	"fmt"
)

func callbackMap(cfg *config, args ...string) error {
	// Both URLs nil means we haven't loaded a page yet, so start at page 1.
	if cfg.nextLocationAreaURL == nil && cfg.prevLocationAreaURL != nil {
		return errors.New("You are on the last page")
	}
	return showLocationPage(cfg, cfg.nextLocationAreaURL)
}

func callbackMapb(cfg *config, args ...string) error {
	if cfg.prevLocationAreaURL == nil {
		return errors.New("You are on the first page")
	}
	return showLocationPage(cfg, cfg.prevLocationAreaURL)
}

// showLocationPage prints one page of areas and remembers its neighbours.
func showLocationPage(cfg *config, pageURL *string) error {
	resp, err := cfg.pokeapiClient.ListLocationAreas(pageURL)
	if err != nil {
		return err
	}
	fmt.Println("Location areas:")
	for _, area := range resp.Results {
		fmt.Printf(" - %s\n", area.Name)
	}
	cfg.nextLocationAreaURL = resp.Next
	cfg.prevLocationAreaURL = resp.Previous
	return nil
}
