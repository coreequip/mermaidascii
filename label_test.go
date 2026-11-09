package mermaidascii

import (
	"strings"
	"testing"
)

// A label on a straight edge between two neighbours used to be dropped: the
// only spots tried lay inside the target box. Every label must arrive, in
// both directions and at any spacing.
func TestEdgeLabelsAreNeverDropped(t *testing.T) {
	sources := map[string]string{
		"LR":   "flowchart LR\n    A[Eingang] --> B{Gültig?}\n    B -->|ja| C[Bericht]\n    B -->|nein| D[Ablehnen]",
		"TD":   "flowchart TD\n    A[Eingang] --> B{Gültig?}\n    B -->|ja| C[Bericht]\n    B -->|nein| D[Ablehnen]",
		"long": "flowchart LR\n    A[Start] -->|eine lange Beschriftung| B[Ziel]",
	}
	for name, src := range sources {
		for _, opts := range [][]Option{nil, {WithSpacing(2, 2)}, {WithSpacing(1, 1), WithPadding(0)}, {WithASCII()}} {
			out := mustRender(t, src, opts...)
			for _, label := range []string{"ja", "nein", "eine lange Beschriftung"} {
				if strings.Contains(src, "|"+label+"|") && !strings.Contains(out, label) {
					t.Errorf("%s, %d options: label %q missing\n%s", name, len(opts), label, out)
				}
			}
		}
	}
}
