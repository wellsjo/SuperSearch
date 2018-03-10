package search

import (
	"log"
)

func Debug(s ...interface{}) {
	if *Opts.debug {
		log.Println(s...)
	}
}
