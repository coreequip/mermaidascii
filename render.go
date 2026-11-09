package mermaidascii

import "unicode/utf8"

// renderFlowchart lays a graph out and draws it.
func renderFlowchart(g *Graph, cfg Config) string {
	if len(g.Nodes) == 0 {
		return ""
	}
	l := doLayout(g, cfg)
	routes := l.routeEdges()
	l.growForPaths(routes)
	for _, r := range routes {
		l.placeLabel(r)
	}
	l.toDrawing()

	cs := charsetFor(cfg)
	c := newCanvas(l.width, l.height)

	// Edges first, boxes on top: where a line runs into a box, the box wins,
	// and the junction character is set afterwards.
	for _, r := range routes {
		l.drawRoute(c, r, cs)
	}
	l.drawJunctions(c, cs)
	for _, p := range l.order {
		c.draw(drawNode(p.node.Label, p.node.Shape, p.box, cs, cfg.UseASCII), p.draw.x, p.draw.y)
	}
	for _, r := range routes {
		l.drawEnds(c, r, cs)
		l.drawLabel(c, r)
	}
	return c.String()
}

// growForPaths makes sure every grid cell an edge passes through has a size,
// so that the drawing is wide and tall enough to hold the detour.
//
// A cell no node occupies gets half the node spacing. One character would be
// enough to draw a line through, but a detour squeezed into a single column
// runs so close past the boxes that it reads as if it touched them.
func (l *layout) growForPaths(routes []*route) {
	for _, r := range routes {
		for _, p := range r.path {
			if _, ok := l.colWidth[p.x]; !ok {
				l.colWidth[p.x] = l.cfg.PaddingX / 2
			}
			if _, ok := l.rowHeight[p.y]; !ok {
				l.rowHeight[p.y] = l.cfg.PaddingY / 2
			}
		}
	}
}

// center is the character position at the middle of a grid cell.
func (l *layout) center(p gridPos) gridPos {
	return gridPos{
		x: l.charX(p.x) + l.colWidth[p.x]/2,
		y: l.charY(p.y) + l.rowHeight[p.y]/2,
	}
}

// drawRoute draws the line of one edge.
//
// After mergePath every segment runs along one axis, so the segments are
// plain runs of characters; the interesting part is the joints between them,
// where a corner character has to bridge two directions.
func (l *layout) drawRoute(c *canvas, r *route, cs charset) {
	if len(r.path) < 2 {
		return
	}
	h, v := cs.line(r.edge.Style)

	pts := make([]gridPos, len(r.path))
	for i, p := range r.path {
		pts[i] = l.center(p)
	}

	for i := 0; i < len(pts)-1; i++ {
		from, to := pts[i], pts[i+1]
		if from.y == to.y {
			for x := minInt(from.x, to.x); x <= maxInt(from.x, to.x); x++ {
				c.set(x, from.y, h)
			}
			continue
		}
		if from.x != to.x {
			// Only reachable through the straight-line fallback, when no
			// route was found at all. Draw the horizontal leg first so the
			// edge is visible rather than absent.
			for x := minInt(from.x, to.x); x <= maxInt(from.x, to.x); x++ {
				c.set(x, from.y, h)
			}
		}
		for y := minInt(from.y, to.y); y <= maxInt(from.y, to.y); y++ {
			c.set(to.x, y, v)
		}
	}

	// Corners last, over the lines that meet in them.
	for i := 1; i < len(pts)-1; i++ {
		in := bearing(pts[i-1], pts[i])
		out := bearing(pts[i], pts[i+1])
		if in != out {
			c.set(pts[i].x, pts[i].y, cornerFor(in, out, cs))
		}
	}
}

// bearing is the compass direction of a straight segment.
func bearing(from, to gridPos) dir {
	switch {
	case to.x > from.x:
		return dirRight
	case to.x < from.x:
		return dirLeft
	case to.y > from.y:
		return dirDown
	default:
		return dirUp
	}
}

// cornerFor is the corner character joining an incoming and an outgoing
// direction. The character is named after the two arms it shows: a line
// arriving from the left and leaving downwards needs arms to the left and
// down, which is ┐.
func cornerFor(in, out dir, cs charset) rune {
	switch {
	case in == dirRight && out == dirDown, in == dirUp && out == dirLeft:
		return cs.cornerTR
	case in == dirRight && out == dirUp, in == dirDown && out == dirLeft:
		return cs.cornerBR
	case in == dirLeft && out == dirDown, in == dirUp && out == dirRight:
		return cs.cornerTL
	case in == dirLeft && out == dirUp, in == dirDown && out == dirRight:
		return cs.cornerBL
	}
	return cs.cross
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// drawEnds sets the arrowheads and the junctions where an edge meets a box.
//
// The two ends are not symmetric. Where an edge leaves its source, the box
// border opens into a junction (├, ┬) — the line grows out of the box. Where
// it arrives, the border stays intact and the arrowhead sits in front of it,
// so a box is never eaten into by an incoming edge.
func (l *layout) drawEnds(c *canvas, r *route, cs charset) {
	if len(r.path) < 2 {
		return
	}
	first := l.center(r.path[0])
	last := l.center(r.path[len(r.path)-1])

	if tee := teeFor(r.startDir, cs); tee != 0 {
		c.set(first.x, first.y, tee)
	}
	if r.edge.ArrowStart {
		p := stepInside(first, r.startDir)
		c.set(p.x, p.y, arrowFor(opposite(r.startDir), cs))
	}
	if r.edge.ArrowEnd {
		p := stepInside(last, r.endDir)
		c.set(p.x, p.y, arrowFor(r.endDir, cs))
	}
}

// stepInside moves one cell away from a box, along the side the edge is
// attached to — the cell the arrowhead belongs in.
func stepInside(p gridPos, d dir) gridPos {
	switch d {
	case dirLeft:
		return gridPos{p.x - 1, p.y}
	case dirRight:
		return gridPos{p.x + 1, p.y}
	case dirUp:
		return gridPos{p.x, p.y - 1}
	case dirDown:
		return gridPos{p.x, p.y + 1}
	}
	return p
}

// arrowFor is the arrowhead that points into a box from the given side.
//
// An edge attached on the left of a box arrives travelling rightwards, so it
// is drawn with a right-pointing head.
func arrowFor(d dir, cs charset) rune {
	switch d {
	case dirLeft:
		return cs.arrowRight
	case dirRight:
		return cs.arrowLeft
	case dirUp:
		return cs.arrowDown
	case dirDown:
		return cs.arrowUp
	}
	return cs.arrowRight
}

// teeFor is the junction drawn where a line leaves or meets a box without an
// arrowhead — the box border opens towards the line.
func teeFor(d dir, cs charset) rune {
	switch d {
	case dirLeft:
		return cs.teeLeft
	case dirRight:
		return cs.teeRight
	case dirUp:
		return cs.teeUp
	case dirDown:
		return cs.teeDown
	}
	return cs.cross
}

// placeLabel picks the stretch of path an edge label goes on and widens the
// grid so the label fits there. This is determineLabelLine from upstream.
//
// A stretch qualifies when its columns add up to the label's width. The
// last qualifying one wins, and the first stretch only as a fallback: near
// the source, edges from the same node still run together. If no stretch is
// wide enough, the widest is taken. Either way its middle column grows to
// the label plus a cell on each side, so the label never has to be dropped.
func (l *layout) placeLabel(r *route) {
	n := labelWidth(r.edge.Label)
	if n == 0 || len(r.path) < 2 {
		return
	}

	best, found := -1, false
	for _, firstToo := range []bool{false, true} {
		for i := len(r.path) - 1; i >= 1; i-- {
			if (i > 1 || firstToo) && l.stretchWidth(r.path[i-1], r.path[i]) >= n {
				best, found = i, true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		widest := -1
		for i := 1; i < len(r.path); i++ {
			if w := l.stretchWidth(r.path[i-1], r.path[i]); w > widest {
				best, widest = i, w
			}
		}
	}

	a, b := r.path[best-1], r.path[best]
	col := l.labelColumn(min(a.x, b.x), max(a.x, b.x))
	l.colWidth[col] = max(l.colWidth[col], n+2)
	r.labelLine = [2]gridPos{a, b}
}

// labelColumn picks the column between lo and hi to widen for a label: the
// free column closest to the middle.
//
// Deliberately not upstream, which always takes the middle column. When
// that column belongs to a node, the node grows with it: upstream keeps
// drawing the box at its own size and leaves a gap before the edge's
// attachment, and since boxes here fill their grid cells, the box would
// widen and the label land on its border. A column no node spans widens
// only the space between nodes. Only if there is none does the middle one
// serve, as upstream.
func (l *layout) labelColumn(lo, hi int) int {
	mid := lo + (hi-lo)/2
	best := -1
	for x := lo; x <= hi; x++ {
		if l.nodeInColumn(x) {
			continue
		}
		if best < 0 || abs(x-mid) < abs(best-mid) {
			best = x
		}
	}
	if best < 0 {
		return mid
	}
	return best
}

// nodeInColumn reports whether any node spans grid column x.
func (l *layout) nodeInColumn(x int) bool {
	for _, p := range l.order {
		if x >= p.pos.x && x < p.pos.x+3 {
			return true
		}
	}
	return false
}

// stretchWidth adds up the widths of the columns a stretch of path spans.
func (l *layout) stretchWidth(a, b gridPos) int {
	w := 0
	for x := min(a.x, b.x); x <= max(a.x, b.x); x++ {
		w += l.colWidth[x]
	}
	return w
}

// drawLabel writes an edge label centred on its stretch of path, over the
// line itself. This is drawTextOnLine from upstream.
//
// On a vertical stretch, an edge running down puts its label in the upper
// part and one running up in the lower part, so that the labels of two
// edges between the same pair of nodes do not land on each other.
func (l *layout) drawLabel(c *canvas, r *route) {
	if r.edge.Label == "" || len(r.path) < 2 {
		return
	}
	a, b := l.center(r.labelLine[0]), l.center(r.labelLine[1])
	minX, maxX := min(a.x, b.x), max(a.x, b.x)
	minY, maxY := min(a.y, b.y), max(a.y, b.y)
	midX := minX + (maxX-minX)/2
	midY := minY + (maxY-minY)/2

	if minX == maxX {
		startY, endY := r.path[0].y, r.path[len(r.path)-1].y
		shift := max(1, (maxY-minY)/4)
		switch {
		case endY < startY:
			midY += shift
		case endY > startY:
			midY -= shift
		}
	}

	lines := labelLines(r.edge.Label)
	top := midY - (len(lines)-1)/2
	for i, line := range lines {
		n := utf8.RuneCountInString(line)
		x := midX - n/2
		if minY == maxY {
			// Not upstream: on a horizontal stretch the label keeps a cell
			// of line on either side, clear of the attachment at the start
			// and the arrowhead at the end, which otherwise it may cover.
			x = max(minX+2, min(x, maxX-2-n))
		}
		c.text(x, top+i, line)
	}
}

// overlapsNode reports whether a run of n characters starting at (x, y) would
// touch any node's box.
func (l *layout) overlapsNode(x, y, n int) bool {
	for _, p := range l.order {
		if y < p.draw.y || y >= p.draw.y+p.box.h {
			continue
		}
		if x < p.draw.x+p.box.w && x+n > p.draw.x {
			return true
		}
	}
	return false
}

// drawJunctions sets the character where bundled edges meet.
//
// Which one it is follows from the lines that actually arrive there, not from
// the bundle's shape: three lines meeting need a T, four a cross, and reading
// them off the canvas gets every case right without enumerating them.
func (l *layout) drawJunctions(c *canvas, cs charset) {
	for _, b := range l.bundles {
		p := l.center(b.junction)
		up, down := isLine(c.at(p.x, p.y-1), cs), isLine(c.at(p.x, p.y+1), cs)
		left, right := isLine(c.at(p.x-1, p.y), cs), isLine(c.at(p.x+1, p.y), cs)

		switch {
		case up && down && left && right:
			c.set(p.x, p.y, cs.cross)
		case up && down && right:
			c.set(p.x, p.y, orCross(cs.teeRight, cs))
		case up && down && left:
			c.set(p.x, p.y, orCross(cs.teeLeft, cs))
		case left && right && down:
			c.set(p.x, p.y, orCross(cs.teeDown, cs))
		case left && right && up:
			c.set(p.x, p.y, orCross(cs.teeUp, cs))
		}
	}
}

// orCross falls back to the cross where a character set has no T-pieces.
// ASCII has none, because a "+" cannot be told apart from a corner — but
// where lines genuinely cross, the "+" is exactly right.
func orCross(r rune, cs charset) rune {
	if r == 0 {
		return cs.cross
	}
	return r
}

// isLine reports whether a character is part of a drawn line.
func isLine(r rune, cs charset) bool {
	switch r {
	case cs.horizontal, cs.vertical, cs.dottedH, cs.dottedV, cs.thickH, cs.thickV,
		cs.cornerTL, cs.cornerTR, cs.cornerBL, cs.cornerBR,
		cs.teeUp, cs.teeDown, cs.teeLeft, cs.teeRight, cs.cross,
		cs.arrowUp, cs.arrowDown, cs.arrowLeft, cs.arrowRight:
		return r != 0
	}
	return false
}
