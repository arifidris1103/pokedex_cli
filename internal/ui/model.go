package ui

import (
	"errors"
	"fmt"
	"image"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

type focus int

const (
	focusLeft focus = iota
	focusWild
	focusInput
)

type leftTab int

const (
	tabLocations leftTab = iota
	tabPokedex
)

// load tracks one background request.
type load[T any] struct {
	val     T
	err     error
	loading bool
}

type logKind int

const (
	logCmd logKind = iota
	logInfo
	logOK
	logErr
)

type logLine struct {
	text string
	kind logKind
}

// Messages produced by background commands.
type (
	pageMsg struct {
		page LocationPage
		dir  int // -1 previous, 0 first, +1 next
		err  error
	}
	exploreMsg struct {
		area  string
		names []string
		err   error
	}
	pokemonMsg struct {
		name string
		p    Pokemon
		err  error
	}
	spriteMsg struct {
		url string
		img image.Image
		err error
	}
	evoMsg struct {
		name   string
		stages []EvoStage
		err    error
	}
	catchMsg struct {
		name   string
		caught bool
		err    error
	}
	throwTickMsg struct{}
	settleMsg    struct{ seq int }
)

const (
	settleDelay = 150 * time.Millisecond // wait for scrolling to stop before fetching
	throwFrames = 4                      // minimum wobbles before the result shows
)

// Model is the Bubble Tea model. Create it with New.
type Model struct {
	be      Backend
	sprites SpriteFetcher     // nil if the backend doesn't provide sprites
	evos    EvolutionProvider // nil if the backend doesn't provide evolutions

	width, height int
	focus, back   focus // back = where Esc returns to from the console
	tab           leftTab
	help          bool

	// Locations (map / mapb)
	page        LocationPage
	pageNum     int
	pageLoading bool
	areaSel     int

	// Wild Pokémon in the explored area (explore)
	area      string
	wild      []string
	wildSel   int
	exploring bool

	// Caught Pokémon (pokedex)
	dex    []Pokemon
	dexSel int

	pokemon map[string]*load[Pokemon]
	images  map[string]*load[image.Image]
	evo     map[string]*load[[]EvoStage]
	spriteS map[string]string // rendered sprite cache

	// Catch animation
	throwing    string
	throwFrame  int
	throwResult *catchMsg

	// Console (the REPL inside the UI)
	input   string
	history []string
	histPos int
	log     []logLine

	seq int
}

// New builds the UI around your backend.
func New(be Backend) *Model {
	m := &Model{
		be:      be,
		pokemon: map[string]*load[Pokemon]{},
		images:  map[string]*load[image.Image]{},
		evo:     map[string]*load[[]EvoStage]{},
		spriteS: map[string]string{},
	}
	m.sprites, _ = be.(SpriteFetcher)
	m.evos, _ = be.(EvolutionProvider)
	m.dex = m.sortedDex()
	return m
}

func (m *Model) Init() tea.Cmd {
	m.addLog(logInfo, "Welcome to the Pokédex! Press ? for help, : to type commands.")
	m.addLog(logCmd, "map")
	m.pageLoading = true
	return m.cmdPage(nil, 0)
}

// ── Background commands ────────────────────────────────────────────────────

func (m *Model) cmdPage(url *string, dir int) tea.Cmd {
	be := m.be
	return func() tea.Msg {
		p, err := be.ListLocationAreas(url)
		return pageMsg{page: p, dir: dir, err: err}
	}
}

func (m *Model) cmdExplore(area string) tea.Cmd {
	be := m.be
	return func() tea.Msg {
		names, err := be.ExploreArea(area)
		return exploreMsg{area: area, names: names, err: err}
	}
}

func (m *Model) cmdCatch(name string) tea.Cmd {
	be := m.be
	return func() tea.Msg {
		ok, err := be.Catch(name)
		return catchMsg{name: name, caught: ok, err: err}
	}
}

func throwTick() tea.Cmd {
	return tea.Tick(350*time.Millisecond, func(time.Time) tea.Msg { return throwTickMsg{} })
}

func (m *Model) settle() tea.Cmd {
	m.seq++
	seq := m.seq
	return tea.Tick(settleDelay, func(time.Time) tea.Msg { return settleMsg{seq} })
}

// ensure starts any fetches the selected Pokémon still needs.
func (m *Model) ensure(name string) tea.Cmd {
	if name == "" {
		return nil
	}
	var cmds []tea.Cmd
	pl, ok := m.pokemon[name]
	if !ok {
		m.pokemon[name] = &load[Pokemon]{loading: true}
		be := m.be
		cmds = append(cmds, func() tea.Msg {
			p, err := be.GetPokemon(name)
			return pokemonMsg{name: name, p: p, err: err}
		})
	} else if !pl.loading && pl.err == nil {
		cmds = append(cmds, m.ensureSprite(pl.val.SpriteURL))
	}
	if _, caught := m.be.Inspect(name); caught && m.evos != nil {
		if _, ok := m.evo[name]; !ok {
			m.evo[name] = &load[[]EvoStage]{loading: true}
			evos := m.evos
			cmds = append(cmds, func() tea.Msg {
				s, err := evos.Evolutions(name)
				return evoMsg{name: name, stages: s, err: err}
			})
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) ensureSprite(url string) tea.Cmd {
	if m.sprites == nil || url == "" {
		return nil
	}
	if _, ok := m.images[url]; ok {
		return nil
	}
	m.images[url] = &load[image.Image]{loading: true}
	sf := m.sprites
	return func() tea.Msg {
		data, err := sf.Sprite(url)
		if err != nil {
			return spriteMsg{url: url, err: err}
		}
		img, err := decodeSprite(data)
		return spriteMsg{url: url, img: img, err: err}
	}
}

// ── Selection helpers ──────────────────────────────────────────────────────

// current is the Pokémon the details panel shows.
func (m *Model) current() string {
	if m.tab == tabPokedex && (m.focus == focusLeft || len(m.wild) == 0) {
		if m.dexSel < len(m.dex) {
			return m.dex[m.dexSel].Name
		}
		return ""
	}
	if m.wildSel < len(m.wild) {
		return m.wild[m.wildSel]
	}
	return ""
}

func (m *Model) sortedDex() []Pokemon {
	d := slices.Clone(m.be.Pokedex())
	slices.SortFunc(d, func(a, b Pokemon) int { return a.ID - b.ID })
	return d
}

func clamp(v, n int) int { return max(0, min(v, n-1)) }

func (m *Model) move(delta int) tea.Cmd {
	switch {
	case m.focus == focusWild:
		m.wildSel = clamp(m.wildSel+delta, len(m.wild))
		return m.settle()
	case m.tab == tabPokedex:
		m.dexSel = clamp(m.dexSel+delta, len(m.dex))
		return m.settle()
	default:
		m.areaSel = clamp(m.areaSel+delta, len(m.page.Areas))
		return nil
	}
}

func (m *Model) addLog(kind logKind, format string, args ...any) {
	m.log = append(m.log, logLine{text: fmt.Sprintf(format, args...), kind: kind})
	if len(m.log) > 200 {
		m.log = m.log[len(m.log)-200:]
	}
}

func (m *Model) logErr(action string, err error) {
	if errors.Is(err, ErrNotImplemented) {
		m.addLog(logErr, "%s isn't implemented in the backend yet", action)
		return
	}
	m.addLog(logErr, "%s failed: %v", action, err)
}

// ── Actions (shared by keys and console commands) ─────────────────────────

func (m *Model) nextPage() tea.Cmd {
	if m.pageLoading {
		return nil
	}
	if m.pageNum > 0 && m.page.Next == nil {
		m.addLog(logErr, "You are on the last page")
		return nil
	}
	m.pageLoading = true
	return m.cmdPage(m.page.Next, +1)
}

func (m *Model) prevPage() tea.Cmd {
	if m.pageLoading {
		return nil
	}
	if m.page.Prev == nil {
		m.addLog(logErr, "You are on the first page")
		return nil
	}
	m.pageLoading = true
	return m.cmdPage(m.page.Prev, -1)
}

func (m *Model) explore(area string) tea.Cmd {
	if area == "" || m.exploring {
		return nil
	}
	m.exploring = true
	m.addLog(logInfo, "Exploring %s...", area)
	return m.cmdExplore(area)
}

func (m *Model) catch(name string) tea.Cmd {
	if name == "" {
		return nil
	}
	if m.throwing != "" {
		m.addLog(logErr, "Wait, a Pokéball is already in the air!")
		return nil
	}
	m.throwing, m.throwFrame, m.throwResult = name, 0, nil
	m.addLog(logInfo, "Throwing a Pokeball at %s...", name)
	return tea.Batch(m.cmdCatch(name), throwTick(), m.ensure(name))
}

func (m *Model) finishThrow() tea.Cmd {
	r := m.throwResult
	m.throwing, m.throwResult = "", nil
	switch {
	case r.err != nil:
		m.logErr("catch", r.err)
	case r.caught:
		m.addLog(logOK, "%s was caught!", r.name)
		m.addLog(logInfo, "You may now inspect it with the inspect command.")
		m.dex = m.sortedDex()
		return m.ensure(r.name)
	default:
		m.addLog(logErr, "%s escaped!", r.name)
	}
	return nil
}

func (m *Model) inspect(name string) {
	p, ok := m.be.Inspect(name)
	if !ok {
		m.addLog(logErr, "you have not caught that pokemon")
		return
	}
	m.tab, m.focus = tabPokedex, focusLeft
	if i := slices.IndexFunc(m.dex, func(d Pokemon) bool { return d.Name == name }); i >= 0 {
		m.dexSel = i
	}
	m.addLog(logInfo, "Name: %s · Height: %d · Weight: %d", p.Name, p.Height, p.Weight)
	stats := make([]string, 0, len(p.Stats))
	for _, s := range p.Stats {
		stats = append(stats, fmt.Sprintf("%s %d", s.Name, s.Value))
	}
	m.addLog(logInfo, "Stats: %s · Types: %s", strings.Join(stats, ", "), strings.Join(p.Types, ", "))
}

var commandHelp = [][2]string{
	{"help", "show this list"},
	{"map", "next page of location areas"},
	{"mapb", "previous page of location areas"},
	{"explore <area>", "list the Pokémon in an area"},
	{"catch <pokemon>", "throw a Pokéball"},
	{"inspect <pokemon>", "show a caught Pokémon"},
	{"pokedex", "list everything you've caught"},
	{"exit", "quit"},
}

// run executes one console line, exactly like the REPL.
func (m *Model) run(line string) tea.Cmd {
	words := strings.Fields(strings.ToLower(line))
	if len(words) == 0 {
		return nil
	}
	m.addLog(logCmd, "%s", strings.Join(words, " "))
	arg := ""
	if len(words) > 1 {
		arg = words[1]
	}
	needArg := func(what string) bool {
		if arg == "" {
			m.addLog(logErr, "usage: %s <%s>", words[0], what)
			return false
		}
		return true
	}

	switch words[0] {
	case "help":
		for _, c := range commandHelp {
			m.addLog(logInfo, " - %s: %s", c[0], c[1])
		}
	case "exit", "quit":
		return tea.Quit
	case "map":
		m.tab = tabLocations
		return m.nextPage()
	case "mapb":
		m.tab = tabLocations
		return m.prevPage()
	case "explore":
		if needArg("area") {
			return m.explore(arg)
		}
	case "catch":
		if needArg("pokemon") {
			return m.catch(arg)
		}
	case "inspect":
		if needArg("pokemon") {
			m.inspect(arg)
			return m.settle()
		}
	case "pokedex":
		m.tab, m.focus = tabPokedex, focusLeft
		if len(m.dex) == 0 {
			m.addLog(logInfo, "Your Pokedex is empty. Go catch something!")
		} else {
			names := make([]string, len(m.dex))
			for i, p := range m.dex {
				names[i] = p.Name
			}
			m.addLog(logInfo, "Your Pokedex: %s", strings.Join(names, ", "))
		}
		return m.settle()
	default:
		m.addLog(logErr, "invalid command")
	}
	return nil
}

// ── Update ─────────────────────────────────────────────────────────────────

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		clear(m.spriteS)
		return m, nil

	case tea.KeyPressMsg:
		return m, m.onKey(msg)

	case pageMsg:
		m.pageLoading = false
		if msg.err != nil {
			m.logErr("map", msg.err)
			return m, nil
		}
		m.page = msg.page
		switch msg.dir {
		case 0:
			m.pageNum = 1
		default:
			m.pageNum += msg.dir
		}
		m.areaSel = 0
		m.addLog(logInfo, "Location areas: page %d (%d areas)", m.pageNum, len(msg.page.Areas))
		return m, nil

	case exploreMsg:
		m.exploring = false
		if msg.err != nil {
			m.logErr("explore", msg.err)
			return m, nil
		}
		m.area, m.wild, m.wildSel = msg.area, msg.names, 0
		m.addLog(logOK, "Found %d Pokémon in %s", len(msg.names), msg.area)
		if len(msg.names) > 0 {
			m.focus = focusWild
			if m.tab == tabPokedex {
				m.tab = tabLocations
			}
		}
		return m, m.settle()

	case settleMsg:
		if msg.seq != m.seq {
			return m, nil
		}
		return m, m.ensure(m.current())

	case pokemonMsg:
		l := m.pokemon[msg.name]
		if l == nil {
			l = &load[Pokemon]{}
			m.pokemon[msg.name] = l
		}
		l.loading, l.val, l.err = false, msg.p, msg.err
		if msg.err != nil {
			return m, nil
		}
		return m, m.ensureSprite(msg.p.SpriteURL)

	case spriteMsg:
		m.images[msg.url] = &load[image.Image]{val: msg.img, err: msg.err}
		return m, nil

	case evoMsg:
		m.evo[msg.name] = &load[[]EvoStage]{val: msg.stages, err: msg.err}
		return m, nil

	case catchMsg:
		m.throwResult = &msg
		if m.throwFrame >= throwFrames {
			return m, m.finishThrow()
		}
		return m, nil

	case throwTickMsg:
		if m.throwing == "" {
			return m, nil
		}
		m.throwFrame++
		if m.throwFrame >= throwFrames && m.throwResult != nil {
			return m, m.finishThrow()
		}
		return m, throwTick()
	}
	return m, nil
}

func (m *Model) onKey(k tea.KeyPressMsg) tea.Cmd {
	key := k.String()
	if key == "ctrl+c" {
		return tea.Quit
	}
	if m.help {
		m.help = false
		return nil
	}

	if m.focus == focusInput {
		switch key {
		case "esc":
			m.input, m.focus = "", m.back
		case "enter":
			line := strings.TrimSpace(m.input)
			m.input, m.focus = "", m.back
			if line != "" {
				m.history = append(m.history, line)
			}
			m.histPos = len(m.history)
			return m.run(line)
		case "backspace":
			if r := []rune(m.input); len(r) > 0 {
				m.input = string(r[:len(r)-1])
			}
		case "ctrl+u":
			m.input = ""
		case "up":
			if m.histPos > 0 {
				m.histPos--
				m.input = m.history[m.histPos]
			}
		case "down":
			if m.histPos < len(m.history)-1 {
				m.histPos++
				m.input = m.history[m.histPos]
			} else {
				m.histPos, m.input = len(m.history), ""
			}
		default:
			m.input += k.Text
		}
		return nil
	}

	switch key {
	case "q":
		return tea.Quit
	case "?":
		m.help = true
	case ":", "/":
		m.back, m.focus = m.focus, focusInput
	case "tab", "shift+tab":
		if m.focus == focusLeft && len(m.wild) > 0 {
			m.focus = focusWild
		} else {
			m.focus = focusLeft
		}
		return m.settle()
	case "p", "2":
		m.tab, m.focus = tabPokedex, focusLeft
		m.dex = m.sortedDex()
		return m.settle()
	case "1":
		m.tab, m.focus = tabLocations, focusLeft
	case "up", "k":
		return m.move(-1)
	case "down", "j":
		return m.move(1)
	case "pgup":
		return m.move(-10)
	case "pgdown":
		return m.move(10)
	case "home", "g":
		return m.move(-1 << 20)
	case "end", "G":
		return m.move(1 << 20)
	case "right", "l", "n":
		if m.tab == tabLocations {
			m.addLog(logCmd, "map")
			return m.nextPage()
		}
	case "left", "h", "b":
		if m.tab == tabLocations {
			m.addLog(logCmd, "mapb")
			return m.prevPage()
		}
	case "enter", "e":
		switch {
		case m.focus == focusLeft && m.tab == tabLocations && m.areaSel < len(m.page.Areas):
			area := m.page.Areas[m.areaSel]
			m.addLog(logCmd, "explore %s", area)
			return m.explore(area)
		case m.focus == focusWild && key == "enter":
			return m.run("catch " + m.current())
		case m.focus == focusLeft && m.tab == tabPokedex && key == "enter":
			return m.run("inspect " + m.current())
		}
	case "c":
		if name := m.current(); name != "" {
			return m.run("catch " + name)
		}
	}
	return nil
}
