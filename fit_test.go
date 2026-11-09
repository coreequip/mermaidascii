package mermaidascii

import (
	"strings"
	"testing"
)

const wideFlowchart = `flowchart LR
    A[Rechnung empfangen] --> B{Format erkannt?}
    B -->|ja| C[Schema validieren]
    B -->|nein| X[Ablehnen]
    C --> D[Schematron prüfen]
    D --> E[Bericht erzeugen]
    E --> F[PDF ausliefern]
    F --> G[Archivieren]`

func mustRender(t *testing.T, src string, opts ...Option) string {
	t.Helper()
	out, err := Render(src, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestMaxWidthLeavesFittingDiagramAlone(t *testing.T) {
	src := "flowchart LR\n    A[Start] --> B[Ende]"
	if got, want := mustRender(t, src, WithMaxWidth(200)), mustRender(t, src); got != want {
		t.Errorf("a diagram that fits must not change:\n%s\nwant\n%s", got, want)
	}
}

func TestMaxWidthZeroMeansNoLimit(t *testing.T) {
	if got, want := mustRender(t, wideFlowchart, WithMaxWidth(0)), mustRender(t, wideFlowchart); got != want {
		t.Error("WithMaxWidth(0) must behave like no limit")
	}
}

func TestMaxWidthNarrowsWideDiagram(t *testing.T) {
	full := Width(mustRender(t, wideFlowchart))
	got := mustRender(t, wideFlowchart, WithMaxWidth(110))
	if w := Width(got); w > 110 {
		t.Fatalf("width %d exceeds the limit of 110 (unlimited: %d)\n%s", w, full, got)
	}
	// Narrowing must not lose any words of the labels.
	for _, word := range []string{"Rechnung", "empfangen", "Schematron", "prüfen", "Archivieren", "Ablehnen"} {
		if !strings.Contains(got, word) {
			t.Errorf("%q is missing after narrowing:\n%s", word, got)
		}
	}
}

// The steps are tried in order and the first that fits wins, so a generous
// limit only wraps labels and keeps the default spacing.
func TestMaxWidthStopsAtFirstStepThatFits(t *testing.T) {
	wrappedOnly := Width(mustRender(t, wideFlowchart, WithMaxWidth(130)))
	tightest := Width(mustRender(t, wideFlowchart, WithMaxWidth(1)))
	if wrappedOnly <= tightest {
		t.Errorf("limit 130 gave width %d, the tightest step %d: spacing was reduced although wrapping sufficed", wrappedOnly, tightest)
	}
}

func TestMaxWidthUnreachableReturnsNarrowest(t *testing.T) {
	full := Width(mustRender(t, wideFlowchart))
	got := Width(mustRender(t, wideFlowchart, WithMaxWidth(1)))
	if got >= full {
		t.Errorf("an unreachable limit should still give the narrowest drawing: got %d, unlimited %d", got, full)
	}
}

func TestMaxWidthIgnoredForSequenceDiagrams(t *testing.T) {
	src := "sequenceDiagram\n    Alice->>Bob: Eine recht lange Nachricht an Bob"
	if got, want := mustRender(t, src, WithMaxWidth(10)), mustRender(t, src); got != want {
		t.Error("sequence diagrams have no step to narrow them and must stay unchanged")
	}
}

func TestWrapLabel(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Rechnung empfangen", "Rechnung\nempfangen"},
		{"Archivieren", "Archivieren"},
		// Short words share a line as long as it is no wider than the
		// longest word.
		{"PDF an Kunden", "PDF an\nKunden"},
		// Existing breaks stay where the author put them.
		{"Schema prüfen\nEN 16931", "Schema\nprüfen\nEN\n16931"},
		{"", ""},
	}
	for _, tc := range tests {
		if got := wrapLabel(tc.in); got != tc.want {
			t.Errorf("wrapLabel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWidth(t *testing.T) {
	if got := Width("ab\n┌──┐\n"); got != 4 {
		t.Errorf("Width = %d, want 4 (runes, not bytes)", got)
	}
}
