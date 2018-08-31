package log

import (
	"log"
	"os"

	"github.com/fatih/color"
)

var (
	DebugMode bool

	highlightError = color.New(color.FgRed).Add(color.Bold)
)

func Debug(a string, s ...interface{}) {
	if DebugMode {
		log.Printf(a, s...)
	}
}

func Fail(a string, s ...interface{}) {
	highlightError.Sprintf(a, s...)
	os.Exit(1)
}
