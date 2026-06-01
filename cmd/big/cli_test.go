package main

import (
	"bytes"
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
)

func TestParseCLIAcceptsAgeBeforePath(t *testing.T) {
	options, err := parseCLI([]string{"--age", "30", "C:\\Users\\pavel\\Downloads"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}

	if options.minAgeDays != 30 {
		t.Fatalf("min age days = %d, want %d", options.minAgeDays, 30)
	}
	if len(options.args) != 1 || options.args[0] != "C:\\Users\\pavel\\Downloads" {
		t.Fatalf("args = %#v, want downloads path", options.args)
	}
}

func TestParseCLIAcceptsAgeAfterPath(t *testing.T) {
	options, err := parseCLI([]string{"C:\\Users\\pavel\\Downloads", "--age", "30"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseCLI returned error: %v", err)
	}

	if options.minAgeDays != 30 {
		t.Fatalf("min age days = %d, want %d", options.minAgeDays, 30)
	}
	if len(options.args) != 1 || options.args[0] != "C:\\Users\\pavel\\Downloads" {
		t.Fatalf("args = %#v, want downloads path", options.args)
	}
}

func TestParseCLIRejectsNegativeAge(t *testing.T) {
	_, err := parseCLI([]string{"--age", "-1"}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("parseCLI returned nil error")
	}
	if err.Error() != "invalid age: must be greater than or equal to 0" {
		t.Fatalf("error = %q, want invalid age", err)
	}
}

func TestParseCLIHelpUsesRootEntryAgeLanguage(t *testing.T) {
	var output bytes.Buffer
	_, err := parseCLI([]string{"--help"}, &output)
	if err != flag.ErrHelp {
		t.Fatalf("parseCLI error = %v, want ErrHelp", err)
	}

	help := output.String()
	if !strings.Contains(help, "Usage: big [flags] [path]") {
		t.Fatalf("help does not contain usage: %q", help)
	}
	if !strings.Contains(help, "minimum root entry age in whole days; 0 includes any age") {
		t.Fatalf("help does not contain age flag text: %q", help)
	}
}
