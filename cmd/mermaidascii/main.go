// Command mermaidascii renders a Mermaid diagram as ASCII or Unicode art.
//
//	mermaidascii diagram.mmd
//	cat diagram.mmd | mermaidascii
//	mermaidascii -ascii -x 3 diagram.mmd
//	mermaidascii -max-width 80 diagram.mmd
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/coreequip/mermaidascii"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mermaidascii:", err)
		os.Exit(1)
	}
}

func run() error {
	ascii := flag.Bool("ascii", false, "draw with +-| instead of box-drawing characters")
	x := flag.Int("x", 5, "horizontal space between nodes")
	y := flag.Int("y", 5, "vertical space between nodes")
	pad := flag.Int("padding", 1, "space between a label and its border")
	maxW := flag.Int("max-width", 0, "narrow a wider flowchart to this many columns (0: no limit)")
	flag.Parse()

	src, err := readInput(flag.Args())
	if err != nil {
		return err
	}

	opts := []mermaidascii.Option{
		mermaidascii.WithSpacing(*x, *y),
		mermaidascii.WithPadding(*pad),
		mermaidascii.WithMaxWidth(*maxW),
	}
	if *ascii {
		opts = append(opts, mermaidascii.WithASCII())
	}

	out, err := mermaidascii.Render(string(src), opts...)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

// readInput takes the file named on the command line, or standard input.
func readInput(args []string) ([]byte, error) {
	if len(args) == 0 {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}
