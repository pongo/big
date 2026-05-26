package main

import (
	"errors"
	"fmt"
	"os"

	"big/internal/scan"
	"big/internal/tui"
	tea "charm.land/bubbletea/v2"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: big [path]")
		return 1
	}

	root := "."
	if len(args) == 1 {
		root = args[0]
	}

	scanner := scan.NewScanner(nil)
	entries, err := scanner.ScanRoot(root)
	if err != nil {
		printScanError(err)
		return 1
	}

	program := tea.NewProgram(
		tui.NewModel(root, entries),
	)
	if _, err = program.Run(); err != nil {
		clearTerminalScreen()
		fmt.Fprintf(os.Stderr, "tui error: %v\n", err)
		return 1
	}
	clearTerminalScreen()

	return 0
}

func clearTerminalScreen() {
	fmt.Fprint(os.Stdout, "\x1b[2J\x1b[H")
}

func printScanError(err error) {
	switch {
	case errors.Is(err, scan.ErrPathNotFound):
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	case errors.Is(err, scan.ErrNotDirectory):
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	default:
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}
