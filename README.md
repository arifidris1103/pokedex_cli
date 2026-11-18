# pokedex_cli

A Pokédex for the terminal, written in Go. Explore locations from the Pokémon games, find the wild Pokémon there, try to catch them and inspect the ones you've caught. All data comes live from [PokéAPI](https://pokeapi.co).

It comes in two forms:

- **A REPL**: type commands at a `>` prompt.
- **A full-screen UI** built with [Bubble Tea](https://github.com/charmbracelet/bubbletea): browse areas, watch the Pokéball wobble, see pixel-art sprites (a silhouette until you catch it), colour-coded stats and evolution lines.

## Run

Requires Go 1.27+.

```bash
go run .                        # REPL
go run ./cmd/pokedex-tui        # terminal UI (needs at least 90×24)
go run ./cmd/pokedex-tui-demo   # terminal UI on offline demo data
go test ./...
```

## REPL commands

| Command | What it does |
| --- | --- |
| `map` / `mapb` | next / previous page of location areas |
| `explore <area>` | list the Pokémon found in an area |
| `catch <pokemon>` | throw a Pokéball (harder for higher base experience, 5–95%) |
| `inspect <pokemon>` | show stats and types of a caught Pokémon |
| `pokedex` | list everything you've caught |
| `help` / `exit` | |

## How it's built

```
main.go, repl.go, command_*.go   the REPL
pokecache/                       thread-safe response cache with expiry
internal/pokeapi/                PokéAPI client; every request goes through one
                                 generic, cached get[T] helper
internal/pokedex/                caught Pokémon (mutex-protected) + the catch rule
internal/ui/                     Bubble Tea UI, talks to the backend only through
                                 the ui.Backend interface
cmd/pokedex-tui/                 wires pokeapi + pokedex into ui.Backend
cmd/pokedex-tui-demo/            the UI on an offline demo backend
```

The REPL and the UI share the same client, Pokédex and catch rule. The UI calls the backend from goroutines, which is why the cache and the Pokédex are guarded by mutexes.
