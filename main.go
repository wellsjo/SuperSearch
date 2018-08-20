package main

import (
	"flag"
	"path/filepath"

	"github.com/wellsjo/SuperSearch/search"
)

func main() {
	var (
		pattern  string
		location string
	)

	debug := flag.Bool("D", false, "Debug mode.")
	flag.Parse()
	args := flag.Args()

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
		Pattern:  pattern,
		Location: location,
		Debug:    *debug,
	})

	ss.Run()
}
