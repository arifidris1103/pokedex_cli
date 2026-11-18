package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"
)

// decodeSprite decodes a PNG and crops away its transparent border.
func decodeSprite(data []byte) (image.Image, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X-1, b.Min.Y-1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if opaque(img, x, y) {
				minX, minY = min(minX, x), min(minY, y)
				maxX, maxY = max(maxX, x), max(maxY, y)
			}
		}
	}
	if maxX < minX {
		return img, nil
	}
	if sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	}); ok {
		return sub.SubImage(image.Rect(minX, minY, maxX+1, maxY+1)), nil
	}
	return img, nil
}

func opaque(img image.Image, x, y int) bool {
	_, _, _, a := img.At(x, y).RGBA()
	return a >= 0x8000
}

// renderSprite draws img into a w×h cell box using half blocks: every cell
// holds two stacked pixels (▀ with fg = top pixel, bg = bottom pixel).
// silhouette paints every pixel one dark colour, for Pokémon not caught yet.
func renderSprite(img image.Image, w, h int, silhouette bool) string {
	if img == nil || w <= 0 || h <= 0 {
		return ""
	}
	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()
	if iw == 0 || ih == 0 {
		return ""
	}
	// Fit into w × 2h pixels. Upscale by whole numbers only so pixel art stays crisp.
	scale := min(float64(w)/float64(iw), float64(2*h)/float64(ih))
	if scale >= 1 {
		scale = float64(int(scale))
	}
	outW := max(1, min(w, int(float64(iw)*scale)))
	outH := max(1, min(2*h, int(float64(ih)*scale)))

	type px struct {
		r, g, b uint8
		ok      bool
	}
	sample := func(x, y int) px {
		if y >= outH {
			return px{}
		}
		sx := b.Min.X + min(iw-1, int((float64(x)+0.5)/scale))
		sy := b.Min.Y + min(ih-1, int((float64(y)+0.5)/scale))
		r, g, bl, a := img.At(sx, sy).RGBA()
		if a < 0x8000 {
			return px{}
		}
		if silhouette {
			return px{20, 28, 54, true}
		}
		return px{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), true}
	}

	var sb strings.Builder
	rows := (outH + 1) / 2
	lines := make([]string, 0, rows)
	for cy := range rows {
		sb.Reset()
		for cx := range outW {
			top, bot := sample(cx, 2*cy), sample(cx, 2*cy+1)
			switch {
			case top.ok && bot.ok:
				fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", top.r, top.g, top.b, bot.r, bot.g, bot.b)
			case top.ok:
				fmt.Fprintf(&sb, "\x1b[49m\x1b[38;2;%d;%d;%dm▀", top.r, top.g, top.b)
			case bot.ok:
				fmt.Fprintf(&sb, "\x1b[49m\x1b[38;2;%d;%d;%dm▄", bot.r, bot.g, bot.b)
			default:
				sb.WriteString("\x1b[0m ")
			}
		}
		sb.WriteString("\x1b[0m")
		lines = append(lines, sb.String())
	}
	return strings.Join(lines, "\n")
}
