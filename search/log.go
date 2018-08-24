package search

import (
	"log"
)

var (
	// Set at compile time with -ldflags
	debugMode string
	// Set at run time
	DebugMode bool
)

func init() {
	if debugMode == "true" {
		DebugMode = true
	}
}

func debug(a string, s ...interface{}) {
	if DebugMode {
		log.Printf(a, s...)
	}
}
