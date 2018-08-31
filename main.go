package main

import (
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/wellsjo/SuperSearch/src/log"
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
		log.Fail(err.Error())
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
			log.Fail(err.Error())
		}
	}

	if opts.Debug {
		log.DebugMode = true
	}

	search.New(&search.Options{
		Pattern:      pattern,
		Location:     location,
		Quiet:        opts.Quiet,
		Hidden:       opts.Hidden,
		Unrestricted: opts.Unrestricted,
		Debug:        opts.Debug,
	}).Run()
}
