package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/wellsjo/SuperSearch/src/logger"
	"github.com/wellsjo/SuperSearch/src/search"
)

func main() {
	var (
		pattern  string
		location string
		opts     search.Options
	)

	parser := flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] PATTERN [PATH]"
	args, err := parser.Parse()

	if err != nil {
		logger.Fail(err.Error())
	}

	if len(args) == 0 {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	} else {
		pattern = args[0]
	}

	if len(args) > 1 {
		location = args[1]
	}

	if pattern == "" {
		parser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	if location == "" {
		location, err = os.Getwd()
		if err != nil {
			logger.Fail(err.Error())
		}
	}

	search.New(&search.Options{
		Pattern:  pattern,
		Location: location,

		IgnoreCase:   opts.IgnoreCase,
		Quiet:        opts.Quiet,
		Hidden:       opts.Hidden,
		Unrestricted: opts.Unrestricted,
		Debug:        opts.Debug,
		ShowStats:    opts.ShowStats,
	}).Run()
}
