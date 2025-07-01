package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ef-tech/github-issue-tool/pkg/config"
)

type Options struct {
	File      string
	DryRun    bool
	ShowHelp  bool
	ShowVersion bool
	Repository string
	NoSort    bool
	LabelOnly bool
}

func ParseFlags() (*Options, error) {
	opts := &Options{}

	flag.StringVar(&opts.File, "file", "", "Path to the issues file (required)")
	flag.StringVar(&opts.File, "f", "", "Path to the issues file (short form)")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "Perform a dry run without creating issues")
	flag.BoolVar(&opts.DryRun, "n", false, "Perform a dry run (short form)")
	flag.BoolVar(&opts.ShowHelp, "help", false, "Show help message")
	flag.BoolVar(&opts.ShowHelp, "h", false, "Show help message (short form)")
	flag.BoolVar(&opts.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&opts.ShowVersion, "v", false, "Show version information (short form)")
	flag.StringVar(&opts.Repository, "repo", "", "Target repository (owner/name format)")
	flag.BoolVar(&opts.NoSort, "no-sort", false, "Create issues in file order without dependency sorting")
	flag.BoolVar(&opts.LabelOnly, "label-only", false, "Create only labels without creating issues")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s - %s\n\n", config.AppName, config.AppDesc)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  %s [options]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Create issues from a file\n")
		fmt.Fprintf(os.Stderr, "  %s --file issues.txt\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  # Dry run to preview what will be created\n")
		fmt.Fprintf(os.Stderr, "  %s --file issues.txt --dry-run\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  # Using short form options\n")
		fmt.Fprintf(os.Stderr, "  %s -f issues.txt -n\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  # Create issues in file order without dependency sorting\n")
		fmt.Fprintf(os.Stderr, "  %s --file issues.txt --no-sort\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "  # Create only labels without creating issues\n")
		fmt.Fprintf(os.Stderr, "  %s --file issues.txt --label-only\n", filepath.Base(os.Args[0]))
	}

	flag.Parse()

	if opts.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	if opts.ShowVersion {
		fmt.Printf("%s version %s\n", config.AppName, config.AppVersion)
		os.Exit(0)
	}

	if opts.File == "" {
		return nil, fmt.Errorf("--file flag is required")
	}

	return opts, nil
}