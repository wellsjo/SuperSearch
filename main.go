package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"

	"github.com/wellsjo/SuperSearch/search"
)

var opts struct {
	Debug bool `short:"D" long:"debug" description:"Show verbose debug information"`

	Concurrency int `short:"c" long:"concurrency" description:"The number of files to process in parallel" default:"8"`
}

func main() {
	var (
		pattern  string
		location string
	)

	parser := flags.NewParser(&opts, flags.Default)
	args, err := parser.Parse()

	if err != nil {
		log.Fatal(err)
	}

	if len(args) == 0 {
		parser.WriteHelp(os.Stdout)
		return
	}

	if len(args) > 0 {
		pattern = args[0]
	}
	if len(args) > 1 {
		location = args[1]
	}
	if location == "" {
		location = "." + string(filepath.Separator)
	}
	if pattern == "" {
		flag.PrintDefaults()
		return
	}

	ss := search.New(&search.Options{
		Pattern:     pattern,
		Location:    location,
		Debug:       opts.Debug,
		Concurrency: opts.Concurrency,
	})

	ss.Run()
}
