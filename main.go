/*
 * ELF tree - Tree viewer for ELF library dependency
 *
 * Copyright (C) 2017  Namhyung Kim <namhyung@gmail.com>
 *
 * Released under MIT license.
 */

// Command elftree prints the shared library dependencies of an ELF binary.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/luncliff/elftree/internal/elftree"
)

// version is overridable at build time with
// `-ldflags "-X main.version=v1.2.3"`. When empty, the version recorded by
// `go install` is used instead.
var version string

const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type options struct {
	showPath bool
	verbose  bool
	format   string
	depth    int
	sections bool
	segments bool
	dynamic  bool
	symbols  bool
	version  bool
}

func newFlagSet(opts *options, out io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet("elftree", flag.ContinueOnError)
	fs.SetOutput(out)

	// Each option has a POSIX style short name and a GNU style long name.
	// Go's flag package accepts both `-name` and `--name` spellings.
	fs.BoolVar(&opts.showPath, "path", false, "show the resolved path of each library")
	fs.BoolVar(&opts.showPath, "p", false, "alias for --path")
	fs.BoolVar(&opts.verbose, "verbose", false, "show a summary of the ELF file")
	fs.BoolVar(&opts.verbose, "v", false, "alias for --verbose")
	fs.StringVar(&opts.format, "format", "tree", "output format: tree, flat or json")
	fs.StringVar(&opts.format, "f", "tree", "alias for --format")
	fs.IntVar(&opts.depth, "depth", 0, "limit the dependency depth (0 means no limit)")
	fs.IntVar(&opts.depth, "d", 0, "alias for --depth")
	fs.BoolVar(&opts.sections, "sections", false, "show the section headers")
	fs.BoolVar(&opts.segments, "segments", false, "show the program headers")
	fs.BoolVar(&opts.dynamic, "dynamic", false, "show the .dynamic section entries")
	fs.BoolVar(&opts.symbols, "symbols", false, "show the symbol tables")
	fs.BoolVar(&opts.version, "version", false, "show the version and exit")

	fs.Usage = func() {
		fmt.Fprintln(out, "elftree - show library dependencies of an ELF binary")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Usage:")
		fmt.Fprintln(out, "  elftree [options] <file>")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Options:")
		fs.PrintDefaults()
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Examples:")
		fmt.Fprintln(out, "  elftree /bin/ls")
		fmt.Fprintln(out, "  elftree --path --depth 2 /bin/ls")
		fmt.Fprintln(out, "  elftree --format json /lib/x86_64-linux-gnu/libc.so.6")
	}
	return fs
}

func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "(devel)"
}

func run(args []string, stdout, stderr io.Writer) int {
	var opts options

	fs := newFlagSet(&opts, stderr)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}

	if opts.version {
		fmt.Fprintf(stdout, "elftree %s\n", buildVersion())
		return exitOK
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(stderr, "elftree: exactly one ELF file is required")
		fs.Usage()
		return exitUsage
	}

	format := strings.ToLower(opts.format)
	switch format {
	case "tree", "flat", "json":
	default:
		fmt.Fprintf(stderr, "elftree: unknown output format %q (want tree, flat or json)\n", opts.format)
		return exitUsage
	}
	if opts.depth < 0 {
		fmt.Fprintln(stderr, "elftree: --depth must not be negative")
		return exitUsage
	}

	graph, err := elftree.Load(rest[0])
	if err != nil {
		fmt.Fprintf(stderr, "elftree: %v\n", err)
		return exitError
	}

	printOpts := elftree.PrintOptions{ShowPath: opts.showPath, MaxDepth: opts.depth}
	switch format {
	case "json":
		err = graph.WriteJSON(stdout, printOpts)
	case "flat":
		err = graph.WriteFlat(stdout, printOpts)
	default:
		err = graph.WriteTree(stdout, printOpts)
	}
	if err != nil {
		fmt.Fprintf(stderr, "elftree: %v\n", err)
		return exitError
	}

	if format == "json" {
		return exitOK
	}

	for _, step := range []struct {
		enabled bool
		write   func(io.Writer) error
	}{
		{opts.verbose, graph.WriteSummary},
		{opts.segments, graph.WriteSegments},
		{opts.sections, graph.WriteSections},
		{opts.dynamic, graph.WriteDynamic},
		{opts.symbols, graph.WriteSymbols},
	} {
		if !step.enabled {
			continue
		}
		if err := step.write(stdout); err != nil {
			fmt.Fprintf(stderr, "elftree: %v\n", err)
			return exitError
		}
	}
	return exitOK
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
