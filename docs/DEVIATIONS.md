# Deviations from the reference implementation

The golden files in `testdata/golden/` are the output of beautiful-mermaid.
Where this port differs, the difference belongs here — either as a gap still
to close, or as a deliberate choice.

## Gaps still to close

Each of these is a skipped fixture in `TestGolden`. Closing one means deleting
its line from `notYet` in `golden_test.go` and watching the test go green.

**Edge bundling** — `flow-multiroot`, `flow-edge-labels`.
When several edges run into the same node, the reference merges them before
they arrive: they meet at a junction and share the last stretch of line. This
port draws each edge separately, so the lines run in parallel instead of
joining. Upstream this is `edge-bundling.ts`.

**Shape-specific sizes** — `flow-shapes`.
Stadium and hexagon nodes are drawn slightly wider upstream, and a stadium is
one row taller. Every shape here is measured as a rectangle; only the corner
characters differ.

**Self edges** — `flow-selfloop`.
`A --> A` needs a small loop routed around the node. Currently the path
collapses onto the node itself.

**Back edges** — `flow-cycle`, `flow-direction-bt`.
An edge that runs against the flow picks a different attachment point than
upstream, so the detour comes out on the other side of the diagram. The
picture is correct, just mirrored.

**Edge label placement** — `flow-edge-labels`.
Labels are placed beside the line at its midpoint, and dropped when nothing
fits there. Upstream puts them further along the line and makes room for
them, so a label never goes missing. Here, a horizontal edge between two
close neighbours has no free row and loses its label.

## Deliberate choices

**Label width is counted in runes.** `labelWidth` counts characters, so
"Größe" is five wide. Characters that occupy two terminal columns — CJK, most
emoji — still count as one and make their boxes too narrow. Fixing this needs
an East Asian width table, which upstream gets from a dependency and this
library does without. Latin text, which is what these diagrams carry, is
unaffected.

**Trailing whitespace is trimmed.** The canvas is a fixed rectangle, so every
line ends in spaces. They carry no meaning, and keeping them makes diffs and
golden comparisons noisy. `tools/gen-fixtures.sh` trims the reference output
the same way, so the comparison stays fair.

**Unsupported statements are skipped, not reported.** `subgraph`, `classDef`,
`style` and `click` lines are ignored so that a diagram using them still
renders. A diagram that silently loses its subgraph boxes is more useful than
an error, and `Render` returning a picture is the whole point.
