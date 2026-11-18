package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	minWidth  = 90
	minHeight = 24
	leftW     = 30
	wildW     = 26
	consoleH  = 8
)

func (m *Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.BackgroundColor = colBG
	v.WindowTitle = "Pokédex"
	return v
}

func (m *Model) render() string {
	if m.width == 0 {
		return ""
	}
	if m.width < minWidth || m.height < minHeight {
		msg := fmt.Sprintf("Make the terminal at least %d×%d (now %d×%d)", minWidth, minHeight, m.width, m.height)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, stYellow.Render(msg))
	}

	bodyH := m.height - 1 - consoleH - 1
	detailsW := m.width - leftW - wildW

	header := m.renderHeader()
	body := lipgloss.JoinHorizontal(lipgloss.Top,
		m.renderLeft(leftW, bodyH),
		m.renderWild(wildW, bodyH),
		m.renderDetails(detailsW, bodyH),
	)
	console := m.renderConsole(m.width, consoleH)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, console, m.renderBar())
}

// renderBar centres the key hints on a full-width bar (manually, so a long
// hint line truncates instead of wrapping onto a second row).
func (m *Model) renderBar() string {
	hint := "↑/↓ move · ←/→ page · enter explore/catch · tab switch · p pokédex · : command · ? help · q quit"
	if ansi.StringWidth(hint) > m.width {
		hint = ansi.Truncate(hint, m.width, "…")
	}
	left := (m.width - ansi.StringWidth(hint)) / 2
	return lipgloss.NewStyle().Background(colBar).Foreground(colText).
		Render(strings.Repeat(" ", left) + fit(hint, m.width-left))
}

func (m *Model) renderHeader() string {
	left := stTitle.Render(" Pokédex") + stYellow.Render(" — Explorer")
	right := stOrange.Render(fmt.Sprintf("caught %d ", len(m.dex)))
	gap := max(1, m.width-ansi.StringWidth(left)-ansi.StringWidth(right))
	return left + strings.Repeat(" ", gap) + right
}

// ── Left: Locations / Pokédex tabs ────────────────────────────────────────

func (m *Model) renderLeft(w, h int) string {
	focused := m.focus == focusLeft
	tab := func(label string, active bool) string {
		switch {
		case active && focused:
			return stTitle.Render(label)
		case active:
			return lipgloss.NewStyle().Foreground(colText).Bold(true).Render(label)
		default:
			return stDim.Render(label)
		}
	}
	titleText := tab("Locations", m.tab == tabLocations) + stDim.Render(" │ ") +
		tab(fmt.Sprintf("Pokédex %d", len(m.dex)), m.tab == tabPokedex)

	rows := h - 2
	var body []string
	right := ""
	if m.tab == tabLocations {
		if m.pageNum > 0 {
			right = fmt.Sprintf("p.%d", m.pageNum)
		}
		switch {
		case m.pageLoading && len(m.page.Areas) == 0:
			body = []string{stDim.Render(" loading…")}
		case len(m.page.Areas) == 0:
			body = []string{stDim.Render(" no areas loaded")}
		default:
			body = m.listLines(len(m.page.Areas), m.areaSel, rows, w-2, focused, func(i int) string {
				return " " + m.page.Areas[i]
			})
			if m.pageLoading {
				body[len(body)-1] = stDim.Render(" loading page…")
			}
		}
	} else {
		if len(m.dex) == 0 {
			body = []string{
				stDim.Render(" Nothing caught yet."),
				"",
				stDim.Render(" Explore an area, then press"),
				stDim.Render(" c on a Pokémon to catch it."),
			}
		} else {
			body = m.listLines(len(m.dex), m.dexSel, rows, w-2, focused, func(i int) string {
				p := m.dex[i]
				return fmt.Sprintf(" %s %s", stDim.Render(fmt.Sprintf("#%04d", p.ID)), title(p.Name))
			})
		}
	}
	return panel(titleText, right, body, w, h, focused)
}

// listLines renders a scrolling list with a highlighted row.
func (m *Model) listLines(n, sel, rows, width int, focused bool, item func(int) string) []string {
	start, end := listWindow(n, sel, rows)
	lines := make([]string, 0, rows)
	for i := start; i < end; i++ {
		text := item(i)
		if i == sel {
			st := stSelDim
			if focused {
				st = stSel
			}
			lines = append(lines, st.Render(fit("▶"+ansi.Strip(text), width)))
			continue
		}
		lines = append(lines, " "+text)
	}
	return lines
}

// ── Middle: wild Pokémon in the explored area ─────────────────────────────

func (m *Model) renderWild(w, h int) string {
	focused := m.focus == focusWild
	var body []string
	right := ""
	switch {
	case m.exploring:
		body = []string{stDim.Render(" exploring…")}
	case m.area == "":
		body = []string{
			stDim.Render(" Pick a location and"),
			stDim.Render(" press Enter to explore."),
		}
	case len(m.wild) == 0:
		body = []string{stDim.Render(" " + m.area), "", stDim.Render(" No Pokémon here.")}
	default:
		right = fmt.Sprint(len(m.wild))
		body = append(body, stDim.Render(" "+m.area), "")
		body = append(body, m.listLines(len(m.wild), m.wildSel, h-4, w-2, focused, func(i int) string {
			name := m.wild[i]
			if _, caught := m.be.Inspect(name); caught {
				return stGreen.Render("●") + " " + title(name)
			}
			return stDim.Render("○") + " " + title(name)
		})...)
	}
	return panel("Wild Pokémon", right, body, w, h, focused)
}

// ── Right: details ────────────────────────────────────────────────────────

func (m *Model) renderDetails(w, h int) string {
	if m.help {
		return panel("Help", "any key closes", m.helpLines(), w, h, true)
	}
	inner, rows := w-2, h-2
	name := m.current()
	if name == "" {
		return panel("Details", "", []string{"", stDim.Render("  Nothing selected.")}, w, h, false)
	}

	_, caught := m.be.Inspect(name)
	pl := m.pokemon[name]

	// Bottom lines: catch animation or evolution line.
	var footer []string
	if m.throwing == name {
		wobble := []string{"◓", "◑", "◒", "◐"}[m.throwFrame%4]
		dots := strings.Repeat("●", min(m.throwFrame, 3)) + strings.Repeat("○", max(0, 3-m.throwFrame))
		footer = append(footer, stYellow.Render(fmt.Sprintf("  %s  Throwing a Pokéball at %s…  %s", wobble, title(name), dots)))
	} else if caught {
		if line := m.evoLine(name); line != "" {
			footer = append(footer, "  "+line)
		}
	}

	spriteW := min(inner*2/5, 36)
	spriteH := rows - len(footer) - 1
	sprite := ""
	if pl != nil && !pl.loading && pl.err == nil {
		sprite = m.sprite(pl.val.SpriteURL, spriteW-2, spriteH, !caught)
	}
	info := m.infoLines(name, pl, caught, inner-spriteW-2)

	spriteLines := strings.Split(sprite, "\n")
	pad := max(0, (spriteH-len(spriteLines))/2)
	var body []string
	for i := range rows - len(footer) {
		left := ""
		if j := i - pad; j >= 0 && j < len(spriteLines) {
			left = spriteLines[j]
		}
		right := ""
		if i < len(info) {
			right = info[i]
		}
		body = append(body, " "+fit(left, spriteW)+" "+right)
	}
	body = append(body, footer...)

	label := "Details"
	if caught {
		label = "Details · caught"
	}
	return panel(label, "", body, w, h, false)
}

func (m *Model) sprite(url string, w, h int, silhouette bool) string {
	il, ok := m.images[url]
	if !ok || il.loading || il.err != nil || il.val == nil {
		return ""
	}
	key := fmt.Sprintf("%s|%d|%d|%v", url, w, h, silhouette)
	if s, ok := m.spriteS[key]; ok {
		return s
	}
	s := renderSprite(il.val, w, h, silhouette)
	m.spriteS[key] = s
	return s
}

var statLabels = map[string]string{
	"hp": "HP", "attack": "Attack", "defense": "Defense",
	"special-attack": "Sp. Atk", "special-defense": "Sp. Def", "speed": "Speed",
}

func (m *Model) infoLines(name string, pl *load[Pokemon], caught bool, w int) []string {
	lines := []string{""}
	head := stTitle.Render(title(name))
	switch {
	case pl == nil || pl.loading:
		return append(lines, head, stDim.Render("loading…"))
	case pl.err != nil:
		return append(lines, head, stRed.Render(errText(pl.err)))
	}
	p := pl.val
	if !caught {
		return append(lines,
			head+stDim.Render("   #????"),
			stDim.Render("Not caught yet"),
			"",
			stDim.Render("Data unlocks when you catch it."),
			stYellow.Render("Press c to throw a Pokéball."),
		)
	}

	lines = append(lines, head+stDim.Render(fmt.Sprintf("   #%04d", p.ID)))
	badges := make([]string, len(p.Types))
	for i, t := range p.Types {
		badges[i] = typeBadge(t)
	}
	lines = append(lines,
		"Types: "+strings.Join(badges, " "),
		fmt.Sprintf("Height: %.1f m   Weight: %.1f kg", float64(p.Height)/10, float64(p.Weight)/10),
		stDim.Render(fmt.Sprintf("Base experience: %d", p.BaseExperience)),
		"",
	)
	barW := max(4, min(w-15, 30))
	total := 0
	for _, s := range p.Stats {
		total += s.Value
		label, ok := statLabels[s.Name]
		if !ok {
			label = title(s.Name)
		}
		filled := max(1, min(barW, s.Value*barW/255))
		bar := lipgloss.NewStyle().Foreground(statColor(s.Value)).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Background(colTrack).Render(strings.Repeat(" ", barW-filled))
		lines = append(lines, fmt.Sprintf("%-8s %3d  %s", label, s.Value, bar))
	}
	if len(p.Stats) > 0 {
		lines = append(lines, stDim.Render(fmt.Sprintf("%-8s %3d", "Total", total)))
	}
	return lines
}

func (m *Model) evoLine(name string) string {
	el, ok := m.evo[name]
	if !ok || el.loading || el.err != nil || len(el.val) < 2 {
		return ""
	}
	parts := make([]string, len(el.val))
	for i, s := range el.val {
		label := title(s.Name)
		switch _, caught := m.be.Inspect(s.Name); {
		case s.Name == name:
			label = stTitle.Render(label)
		case caught:
			label = stGreen.Render(label)
		default:
			label = stDim.Render(label)
		}
		if s.How != "" {
			label = stOrange.Render(s.How) + " " + stDim.Render("→") + " " + label
		}
		parts[i] = label
	}
	return stDim.Render("Evolution: ") + strings.Join(parts, stDim.Render("  "))
}

func errText(err error) string {
	if err == ErrNotImplemented {
		return "not implemented in the backend yet"
	}
	return err.Error()
}

func (m *Model) helpLines() []string {
	keys := [][2]string{
		{"↑ ↓  j k", "move"},
		{"← →  b n", "previous / next page (mapb / map)"},
		{"enter", "explore area · catch · inspect"},
		{"c", "catch the selected Pokémon"},
		{"tab", "switch between lists"},
		{"p  1", "Pokédex tab · Locations tab"},
		{":", "type a command (↑ for history)"},
		{"q", "quit"},
	}
	lines := []string{"", stTitle.Render("  Keys")}
	for _, k := range keys {
		lines = append(lines, "  "+stYellow.Render(fmt.Sprintf("%-10s", k[0]))+" "+k[1])
	}
	lines = append(lines, "", stTitle.Render("  Commands"))
	for _, c := range commandHelp {
		lines = append(lines, "  "+stYellow.Render(fmt.Sprintf("%-18s", c[0]))+" "+stDim.Render(c[1]))
	}
	return lines
}

// ── Console (the REPL) ────────────────────────────────────────────────────

func (m *Model) renderConsole(w, h int) string {
	focused := m.focus == focusInput
	rows := h - 2
	logRows := rows - 1
	start := max(0, len(m.log)-logRows)
	var body []string
	for _, l := range m.log[start:] {
		switch l.kind {
		case logCmd:
			body = append(body, stYellow.Render(" > ")+stYellow.Render(l.text))
		case logOK:
			body = append(body, "   "+stGreen.Render(l.text))
		case logErr:
			body = append(body, "   "+stRed.Render(l.text))
		default:
			body = append(body, "   "+stText.Render(l.text))
		}
	}
	for len(body) < logRows {
		body = append([]string{""}, body...)
	}
	if focused {
		body = append(body, stTitle.Render(" > ")+m.input+stSel.Render(" "))
	} else {
		body = append(body, stDim.Render(" press : to type a command — map, explore <area>, catch <pokemon>, inspect, pokedex"))
	}
	return panel("Console", "", body, w, h, focused)
}
