package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	colBG     = lipgloss.Color("#1E2948")
	colBar    = lipgloss.Color("#2A3862")
	colTrack  = lipgloss.Color("#253258")
	colBorder = lipgloss.Color("#8A93B8")
	colText   = lipgloss.Color("#E2E6F2")
	colDim    = lipgloss.Color("#8A93B8")
	colYellow = lipgloss.Color("#FADE28")
	colOrange = lipgloss.Color("#FF9F1C")
	colGreen  = lipgloss.Color("#4ADE80")
	colRed    = lipgloss.Color("#FF4D6D")

	stText   = lipgloss.NewStyle().Foreground(colText)
	stDim    = lipgloss.NewStyle().Foreground(colDim)
	stYellow = lipgloss.NewStyle().Foreground(colYellow)
	stTitle  = lipgloss.NewStyle().Foreground(colYellow).Bold(true)
	stOrange = lipgloss.NewStyle().Foreground(colOrange)
	stGreen  = lipgloss.NewStyle().Foreground(colGreen)
	stRed    = lipgloss.NewStyle().Foreground(colRed)
	stSel    = lipgloss.NewStyle().Background(colYellow).Foreground(lipgloss.Color("#000000")).Bold(true)
	stSelDim = lipgloss.NewStyle().Background(colBar).Foreground(colText)
)

var typeColors = map[string]string{
	"normal": "#A8A77A", "fire": "#EE8130", "water": "#6390F0", "grass": "#22C55E",
	"electric": "#F7D02C", "ice": "#96D9D6", "fighting": "#C22E28", "poison": "#C052C4",
	"ground": "#E2BF65", "flying": "#A98FF3", "psychic": "#F95587", "bug": "#A6B91A",
	"rock": "#B6A136", "ghost": "#735797", "dragon": "#6F35FC", "dark": "#705746",
	"steel": "#B7B7CE", "fairy": "#D685AD",
}

func typeBadge(t string) string {
	c, ok := typeColors[t]
	if !ok {
		c = "#999999"
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(c)).Foreground(lipgloss.Color("#000000")).Render(" " + title(t) + " ")
}

func statColor(v int) color.Color {
	switch {
	case v < 60:
		return lipgloss.Color("#FF1744")
	case v < 90:
		return lipgloss.Color("#FFA500")
	case v < 110:
		return colYellow
	case v < 130:
		return colGreen
	default:
		return lipgloss.Color("#22D3EE")
	}
}

// title turns "tentacool" / "canalave-city-area" into "Tentacool" / "Canalave City Area".
func title(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool { return r == '-' || r == ' ' })
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// fit pads or truncates a (possibly styled) line to exactly w cells.
func fit(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) > w {
		s = ansi.Truncate(s, w, "…")
	}
	return s + strings.Repeat(" ", w-ansi.StringWidth(s))
}

// panel draws a box of exactly w×h cells with a title in the top border.
// Focused panels get a double yellow border, like the reference design.
// titleText may already be styled; plain text gets the default title style.
func panel(titleText, right string, body []string, w, h int, focused bool) string {
	if w < 6 || h < 2 {
		return ""
	}
	tl, tr, bl, br, hz, vt := "┌", "┐", "└", "┘", "─", "│"
	bs := lipgloss.NewStyle().Foreground(colBorder)
	ts := lipgloss.NewStyle().Foreground(colBorder).Bold(true)
	if focused {
		tl, tr, bl, br, hz, vt = "╔", "╗", "╚", "╝", "═", "║"
		bs = lipgloss.NewStyle().Foreground(colYellow)
		ts = stTitle
	}
	inner := w - 2

	tseg, rseg := "", ""
	if titleText != "" {
		tseg = " " + titleText + " "
		if !strings.Contains(titleText, "\x1b") {
			tseg = ts.Render(tseg)
		}
	}
	if right != "" {
		rseg = stDim.Render(" " + right + " ")
	}
	// top = corner + hz + title + fill + right + hz + corner  (exactly w cells)
	if inner-2-ansi.StringWidth(tseg)-ansi.StringWidth(rseg) < 0 {
		rseg = ""
	}
	if inner-2-ansi.StringWidth(tseg) < 0 {
		tseg = ansi.Truncate(tseg, max(0, inner-2), "")
	}
	fill := inner - 2 - ansi.StringWidth(tseg) - ansi.StringWidth(rseg)
	top := bs.Render(tl+hz) + tseg + bs.Render(strings.Repeat(hz, fill)) + rseg + bs.Render(hz+tr)

	lines := []string{top}
	for i := range h - 2 {
		line := ""
		if i < len(body) {
			line = body[i]
		}
		lines = append(lines, bs.Render(vt)+fit(line, inner)+bs.Render(vt))
	}
	lines = append(lines, bs.Render(bl+strings.Repeat(hz, inner)+br))
	return strings.Join(lines, "\n")
}

// listWindow returns the slice bounds that keep `sel` visible in `height` rows.
func listWindow(n, sel, height int) (start, end int) {
	if height <= 0 || n == 0 {
		return 0, 0
	}
	start = max(0, min(sel-height/2, n-height))
	return start, min(n, start+height)
}
