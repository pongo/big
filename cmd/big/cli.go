package main

import (
	"fmt"
	"io"

	flag "github.com/spf13/pflag"
)

type cliOptions struct {
	args       []string
	minAgeDays int
}

func parseCLI(args []string, output io.Writer) (cliOptions, error) {
	options := cliOptions{}

	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)

	if err := flags.Parse(args); err != nil {
		return cliOptions{}, err
	}

	options.args = flags.Args()
	if len(options.args) > 1 {
		return cliOptions{}, fmt.Errorf("too many arguments")
	}
	if options.minAgeDays < 0 {
		return cliOptions{}, fmt.Errorf("invalid age: must be greater than or equal to 0")
	}

	return options, nil
}

func newCLIFlagSet(output io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet("big", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.SortFlags = false
	flags.Usage = func() {
		_, _ = fmt.Fprintln(flags.Output(), "Usage: big [flags] [path]")
		_, _ = fmt.Fprintln(flags.Output())
		_, _ = fmt.Fprintln(flags.Output(), "Arguments:")
		_, _ = fmt.Fprintln(flags.Output(), "      path   scan root (defaults to current working directory)")
		_, _ = fmt.Fprintln(flags.Output())
		_, _ = fmt.Fprintln(flags.Output(), "Flags:")
		flags.PrintDefaults()
	}
	return flags
}

func addCLIFlags(flags *flag.FlagSet, options *cliOptions) {
	flags.IntVar(&options.minAgeDays, "age", 0, "minimum root entry age in whole days; 0 includes any age")
}

func printCLIHelp(output io.Writer) {
	options := cliOptions{}
	flags := newCLIFlagSet(output)
	addCLIFlags(flags, &options)
	flags.Usage()
}
