package main

import (
	"slices"
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"  Catch   PIKACHU  ", []string{"catch", "pikachu"}},
		{"", []string{}},
		{"\texplore canalave-city-area\n", []string{"explore", "canalave-city-area"}},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if !slices.Equal(actual, c.expected) {
			t.Errorf("cleanInput(%q) = %q, want %q", c.input, actual, c.expected)
		}
	}
}

func TestEveryCommandIsFindable(t *testing.T) {
	for _, c := range getCommands() {
		if got, ok := findCommand(c.name); !ok || got.name != c.name {
			t.Errorf("findCommand(%q) failed", c.name)
		}
	}
	if _, ok := findCommand("fly"); ok {
		t.Error("fly is not a command")
	}
}
