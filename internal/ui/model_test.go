package ui

import (
	"image"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func newTestModel(roll float64) (*Model, *DemoBackend) {
	be := NewDemoBackend()
	be.Delay, be.Offline = 0, true
	be.Rand = func() float64 { return roll }
	m := New(be)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 36})
	return m, be
}

// drain runs a command and feeds every resulting message back into the
// model, like the Bubble Tea runtime does.
func drain(t *testing.T, m *Model, cmd tea.Cmd) {
	t.Helper()
	var run func(tea.Cmd, int)
	run = func(c tea.Cmd, depth int) {
		if c == nil || depth > 50 {
			return
		}
		switch msg := c().(type) {
		case nil, tea.QuitMsg:
		case tea.BatchMsg:
			for _, sub := range msg {
				run(sub, depth+1)
			}
		default:
			_, next := m.Update(msg)
			run(next, depth+1)
		}
	}
	run(cmd, 0)
}

func lastLog(m *Model) string {
	if len(m.log) == 0 {
		return ""
	}
	return m.log[len(m.log)-1].text
}

func logContains(m *Model, s string) bool {
	for _, l := range m.log {
		if strings.Contains(l.text, s) {
			return true
		}
	}
	return false
}

func TestInitLoadsFirstPage(t *testing.T) {
	m, _ := newTestModel(0)
	drain(t, m, m.Init())
	if m.pageNum != 1 || len(m.page.Areas) != demoPageSize || m.page.Next == nil {
		t.Fatalf("page %d with %d areas, next=%v", m.pageNum, len(m.page.Areas), m.page.Next)
	}
}

func TestMapAndMapb(t *testing.T) {
	m, _ := newTestModel(0)
	drain(t, m, m.Init())

	drain(t, m, m.run("map"))
	if m.pageNum != 2 || m.page.Areas[0] != demoAreas[20] {
		t.Fatalf("after map: page %d starts with %q", m.pageNum, m.page.Areas[0])
	}
	drain(t, m, m.run("map"))
	if lastLog(m) != "You are on the last page" {
		t.Fatalf("expected last-page message, got %q", lastLog(m))
	}
	drain(t, m, m.run("mapb"))
	if m.pageNum != 1 {
		t.Fatalf("after mapb: page %d", m.pageNum)
	}
	drain(t, m, m.run("mapb"))
	if lastLog(m) != "You are on the first page" {
		t.Fatalf("expected first-page message, got %q", lastLog(m))
	}
}

func TestExploreThenCatch(t *testing.T) {
	m, be := newTestModel(0) // roll 0 = always caught
	drain(t, m, m.run("explore canalave-city-area"))
	if len(m.wild) != 11 || m.focus != focusWild {
		t.Fatalf("explore: %d wild, focus %v", len(m.wild), m.focus)
	}
	drain(t, m, m.run("catch tentacool"))
	if !logContains(m, "tentacool was caught!") {
		t.Fatalf("log: %+v", m.log)
	}
	if _, ok := be.Inspect("tentacool"); !ok || len(m.dex) != 1 {
		t.Fatalf("pokedex not updated: %+v", m.dex)
	}
	if m.throwing != "" {
		t.Fatal("throw animation still running")
	}
}

func TestCatchEscapes(t *testing.T) {
	m, _ := newTestModel(0.999)
	drain(t, m, m.run("catch gyarados"))
	if lastLog(m) != "gyarados escaped!" || len(m.dex) != 0 {
		t.Fatalf("got %q, dex %d", lastLog(m), len(m.dex))
	}
}

func TestInspect(t *testing.T) {
	m, _ := newTestModel(0)
	drain(t, m, m.run("inspect magikarp"))
	if lastLog(m) != "you have not caught that pokemon" {
		t.Fatalf("got %q", lastLog(m))
	}
	drain(t, m, m.run("catch magikarp"))
	drain(t, m, m.run("inspect magikarp"))
	if m.tab != tabPokedex || m.current() != "magikarp" {
		t.Fatalf("inspect should open the Pokédex on magikarp, got tab %v / %q", m.tab, m.current())
	}
}

func TestBadCommands(t *testing.T) {
	m, _ := newTestModel(0)
	drain(t, m, m.run("fly"))
	if lastLog(m) != "invalid command" {
		t.Fatalf("got %q", lastLog(m))
	}
	drain(t, m, m.run("catch"))
	if lastLog(m) != "usage: catch <pokemon>" {
		t.Fatalf("got %q", lastLog(m))
	}
}

// A backend that hasn't built explore yet should get a friendly message.
type halfBackend struct{ *DemoBackend }

func (halfBackend) ExploreArea(string) ([]string, error) { return nil, ErrNotImplemented }

func TestNotImplementedIsFriendly(t *testing.T) {
	be := NewDemoBackend()
	be.Delay = 0
	m := New(halfBackend{be})
	drain(t, m, m.run("explore eterna-city-area"))
	if lastLog(m) != "explore isn't implemented in the backend yet" {
		t.Fatalf("got %q", lastLog(m))
	}
}

func TestKeysDriveTheUI(t *testing.T) {
	m, _ := newTestModel(0)
	drain(t, m, m.Init())
	press := func(s string) {
		k := tea.KeyPressMsg{Text: s}
		switch s {
		case "enter":
			k = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "down":
			k = tea.KeyPressMsg{Code: tea.KeyDown}
		}
		_, cmd := m.Update(k)
		drain(t, m, cmd)
	}
	press("down")
	press("enter") // explore the 2nd area
	if m.area != demoAreas[1] {
		t.Fatalf("explored %q", m.area)
	}
	press("c") // catch the first wild Pokémon
	if len(m.dex) != 1 {
		t.Fatalf("dex %d", len(m.dex))
	}
	for _, r := range ":pokedex" {
		press(string(r))
	}
	press("enter")
	if m.tab != tabPokedex || m.focus != focusLeft {
		t.Fatalf("console command didn't switch tabs")
	}
}

// Every frame must be exactly the terminal size, or the screen tears.
func TestFrameSize(t *testing.T) {
	for _, size := range [][2]int{{120, 36}, {90, 24}, {200, 50}} {
		m, _ := newTestModel(0)
		m.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		drain(t, m, m.Init())
		drain(t, m, m.run("explore canalave-city-area"))
		drain(t, m, m.run("catch staryu"))
		for _, help := range []bool{false, true} {
			m.help = help
			lines := strings.Split(m.render(), "\n")
			if len(lines) != size[1] {
				t.Fatalf("%v: %d lines", size, len(lines))
			}
			for i, l := range lines {
				if w := ansi.StringWidth(l); w != size[0] {
					t.Fatalf("%v help=%v: line %d is %d wide: %q", size, help, i, w, ansi.Strip(l))
				}
			}
		}
	}
}

func TestRenderSprite(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := range 4 {
		for y := range 4 {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	out := renderSprite(img, 8, 4, false) // upscales 2× → 8 wide, 4 rows
	lines := strings.Split(out, "\n")
	if len(lines) != 4 || ansi.StringWidth(lines[0]) != 8 {
		t.Fatalf("got %d lines, width %d", len(lines), ansi.StringWidth(lines[0]))
	}
}
