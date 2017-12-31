package main

import (
	"os"
	"path/filepath"

	"github.com/wellsjo/search/src"
	"github.com/wellsjo/search/src/log"
)

func main() {
	var (
		pattern  string
		location string
	)
	if len(os.Args) > 1 {
		pattern = os.Args[1]
	}
	if len(os.Args) > 2 {
		location = os.Args[2]
	}
	if location == "" {
		location = "." + string(filepath.Separator)
	}
	if len(pattern) == 0 {
		log.Debug("No search pattern provided.")
		return
	}
	log.Debug("Searching", location, "for", pattern)
	search.Search(pattern, location)
}
