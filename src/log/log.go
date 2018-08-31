package log

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	// Set at compile time with -ldflags
	debugMode string
	// Set at run time
	DebugMode bool

	highlightError = color.New(color.FgRed).Add(color.Bold)
)

func init() {
	if debugMode == "true" {
		DebugMode = true
	}
}

func Debug(a string, s ...interface{}) {
	if DebugMode {
		log.Printf(a, s...)
	}
}

func Fail(a string, s ...interface{}) {
	highlightError.Sprintf(a, s...)
	os.Exit(1)
}
