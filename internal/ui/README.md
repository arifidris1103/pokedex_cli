# internal/ui: Pokédex terminal UI

A [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) UI for the Pokédex. The UI only draws. All data comes through the `Backend` interface in `backend.go`, and each method matches a REPL command:

| REPL command | Backend method |
| --- | --- |
| `map` / `mapb` | `ListLocationAreas(pageURL)` |
| `explore <area>` | `ExploreArea(area)` |
| `catch <pokemon>` | `GetPokemon(name)` + `Catch(name)` |
| `inspect <pokemon>` | `Inspect(name)` |
| `pokedex` | `Pokedex()` |
| optional | `Sprite(url)`, `Evolutions(name)` |

Return `ui.ErrNotImplemented` from anything you haven't built yet; the UI shows "not implemented yet" instead of crashing.

## Setup (once)

```bash
go get charm.land/bubbletea/v2 charm.land/lipgloss/v2 github.com/charmbracelet/x/ansi
go mod tidy
```

## Run

```bash
go run ./cmd/pokedex-tui-demo   # offline demo backend: try the whole UI now
go run ./cmd/pokedex-tui        # your real backend (wired in cmd/pokedex-tui/backend.go)
go test ./internal/ui/          # UI tests (use the demo backend, no network)
```

The terminal must be at least 90×24.

## Rules for the backend

- **Don't print to stdout** (e.g. `fmt.Println("cache hit!")`). The UI owns the screen, so any stray print corrupts it.
- **Be safe for concurrent use.** The UI calls the backend from goroutines, so guard shared maps (like the caught-Pokémon map) with a mutex.

## Files

```
backend.go   the Backend interface + data types (the contract)
model.go     state, keys, background calls, catch animation, console commands
view.go      layout and drawing
styles.go    colours, panels, helpers
sprite.go    PNG → half-block pixel art (2 pixels per character cell)
demo.go      offline demo backend (19 real Pokémon)
model_test.go
```
