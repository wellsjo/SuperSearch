package search

import (
	"log"
	"os"
)

func Debug(a string, s ...interface{}) {
	if *Opts.debug {
		log.Printf(a, s...)
	}
}

func Fail(s ...interface{}) {
	log.Println(s...)
	os.Exit(1)
}
