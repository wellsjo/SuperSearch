package search

import (
	"flag"
	"path/filepath"
)

type Options struct {
	pattern     string
	location    string
	concurrency *int
	debug       *bool
}

var opts *Options

func init() {
	debug := flag.Bool("D", false, "Debug mode.")
	concurrency := flag.Int("c", 64, "Concurrency (number of files processed in parallel)")
	flag.Parse()
	var (
		pattern  string
		location string
	)
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
	opts = &Options{
		pattern,
		location,
		concurrency,
		debug,
	}
}

func GetOptions() *Options {
	return opts
}
