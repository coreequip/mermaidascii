package mermaidascii

import (
	"strings"
	"unicode/utf8"
)

// Width returns the width of a drawing in cells: the rune count of its
// longest line.
func Width(drawing string) int {
	w := 0
	for _, line := range strings.Split(drawing, "\n") {
		if n := utf8.RuneCountInString(line); n > w {
			w = n
		}
	}
	return w
}

// narrowSpacing is the node spacing the second narrowing step falls back to.
// Below two cells the line between neighbouring boxes is barely longer than
// its attachment and arrowhead and no longer reads as an edge.
const narrowSpacing = 2

// narrowingSteps are tried in order on a flowchart that is wider than
// Config.MaxWidth. Each step adds to the ones before it, from the change
// that costs the least legibility to the one that costs the most.
var narrowingSteps = []func(g *Graph, cfg *Config){
	// Labels break at word boundaries, which mostly trades width for height.
	func(g *Graph, _ *Config) {
		for _, n := range g.Nodes {
			n.Label = wrapLabel(n.Label)
		}
	},
	// Less room between the nodes.
	func(_ *Graph, cfg *Config) {
		cfg.PaddingX = min(cfg.PaddingX, narrowSpacing)
		cfg.PaddingY = min(cfg.PaddingY, narrowSpacing)
	},
	// Labels touch their borders.
	func(_ *Graph, cfg *Config) { cfg.Padding = 0 },
}

// renderFlowchartWithin draws src and, if a width limit is set and the
// drawing exceeds it, narrows it step by step until it fits. If no step is
// enough, the narrowest drawing is returned; the caller can tell by Width.
func renderFlowchartWithin(src string, cfg Config) (string, error) {
	g, err := Parse(src)
	if err != nil {
		return "", err
	}
	best := renderFlowchart(g, cfg)
	if cfg.MaxWidth <= 0 || Width(best) <= cfg.MaxWidth {
		return best, nil
	}

	for i := range narrowingSteps {
		// Layout keeps state on the graph, so every attempt starts from a
		// fresh parse with all steps up to this one applied.
		g, _ := Parse(src)
		c := cfg
		for _, step := range narrowingSteps[:i+1] {
			step(g, &c)
		}
		out := renderFlowchart(g, c)
		if Width(out) < Width(best) {
			best = out
		}
		if Width(out) <= cfg.MaxWidth {
			return out, nil
		}
	}
	return best, nil
}

// wrapLabel breaks a node label at word boundaries so that no line is wider
// than its longest word. Line breaks the author wrote stay in place.
func wrapLabel(label string) string {
	longest := 0
	for _, w := range strings.Fields(label) {
		longest = max(longest, utf8.RuneCountInString(w))
	}

	var out []string
	for _, line := range labelLines(label) {
		cur := ""
		for _, w := range strings.Fields(line) {
			switch {
			case cur == "":
				cur = w
			case utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) <= longest:
				cur += " " + w
			default:
				out = append(out, cur)
				cur = w
			}
		}
		out = append(out, cur)
	}
	return strings.Join(out, "\n")
}
