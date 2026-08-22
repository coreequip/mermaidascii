package mermaidascii

// charset holds every character the renderer draws with, in one place, so
// that the ASCII and Unicode modes differ in data rather than in branches.
type charset struct {
	horizontal rune
	vertical   rune
	// dotted and thick are the line styles of `-.->` and `==>`.
	dottedH, dottedV rune
	thickH, thickV   rune

	arrowUp, arrowDown, arrowLeft, arrowRight rune

	// Corners of a path bend, named after the two directions it joins.
	cornerTL, cornerTR, cornerBL, cornerBR rune
	// Junctions where a line meets a box or another line.
	teeUp, teeDown, teeLeft, teeRight rune
	cross                             rune
}

var unicodeSet = charset{
	horizontal: '─', vertical: '│',
	dottedH: '┄', dottedV: '┆',
	thickH: '━', thickV: '┃',
	arrowUp: '▲', arrowDown: '▼', arrowLeft: '◄', arrowRight: '►',
	cornerTL: '┌', cornerTR: '┐', cornerBL: '└', cornerBR: '┘',
	teeUp: '┴', teeDown: '┬', teeLeft: '┤', teeRight: '├',
	cross: '┼',
}

var asciiSet = charset{
	horizontal: '-', vertical: '|',
	dottedH: '.', dottedV: ':',
	thickH: '=', thickV: '|',
	arrowUp: '^', arrowDown: 'v', arrowLeft: '<', arrowRight: '>',
	cornerTL: '+', cornerTR: '+', cornerBL: '+', cornerBR: '+',
	// No junctions in ASCII mode. A '+' would be indistinguishable from a
	// corner, so the box border is left whole and the line simply starts
	// next to it.
	teeUp: 0, teeDown: 0, teeLeft: 0, teeRight: 0,
	cross: '+',
}

// line returns the horizontal and vertical character for an edge style.
func (c charset) line(s EdgeStyle) (h, v rune) {
	switch s {
	case Dotted:
		return c.dottedH, c.dottedV
	case Thick:
		return c.thickH, c.thickV
	default:
		return c.horizontal, c.vertical
	}
}

// corners are the four corner characters of a node outline. Every shape is
// drawn as a box; the corners are what tells them apart. Diagonal strokes
// would break the character grid, so a diamond is a box with ◇ at the
// corners rather than an actual diamond.
type corners struct{ tl, tr, bl, br rune }

var shapeCorners = map[Shape]struct{ unicode, ascii corners }{
	Rectangle:     {corners{'┌', '┐', '└', '┘'}, corners{'+', '+', '+', '+'}},
	Rounded:       {corners{'╭', '╮', '╰', '╯'}, corners{'.', '.', '\'', '\''}},
	Circle:        {corners{'◯', '◯', '◯', '◯'}, corners{'o', 'o', 'o', 'o'}},
	Diamond:       {corners{'◇', '◇', '◇', '◇'}, corners{'<', '>', '<', '>'}},
	Hexagon:       {corners{'⌜', '⌝', '⌞', '⌟'}, corners{'*', '*', '*', '*'}},
	Stadium:       {corners{'(', ')', '(', ')'}, corners{'(', ')', '(', ')'}},
	Subroutine:    {corners{'╟', '╢', '╟', '╢'}, corners{'|', '|', '|', '|'}},
	Cylinder:      {corners{'╭', '╮', '╰', '╯'}, corners{'.', '.', '\'', '\''}},
	Trapezoid:     {corners{'/', '\\', '└', '┘'}, corners{'/', '\\', '+', '+'}},
	Parallelogram: {corners{'/', '/', '/', '/'}, corners{'/', '/', '/', '/'}},
}

// cornersFor returns the corner characters of a shape, falling back to the
// rectangle for anything unknown.
func cornersFor(s Shape, useASCII bool) corners {
	c, ok := shapeCorners[s]
	if !ok {
		c = shapeCorners[Rectangle]
	}
	if useASCII {
		return c.ascii
	}
	return c.unicode
}
