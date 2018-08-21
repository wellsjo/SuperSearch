package search

import (
	"log"
	"os"
)

var (
	// Set at compile time with -ldflags
	debug string
	// Set at run time
	DebugMode bool
)

func init() {
	if debug == "true" {
		DebugMode = true
	}
}

func Debug(a string, s ...interface{}) {
	if DebugMode {
		log.Printf(a, s...)
	}
}

func Fail(s ...interface{}) {
	log.Println(s...)
	os.Exit(1)
}
