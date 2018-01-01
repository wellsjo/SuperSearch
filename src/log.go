package search

import (
	"log"
)

func Debug(s ...interface{}) {
	opts := GetOptions()
	if *opts.debug {
		log.Println(s...)
	}
}
