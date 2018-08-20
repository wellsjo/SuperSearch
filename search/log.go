package search

import (
	"log"
	"os"
)

func Debug(a string, s ...interface{}) {
	log.Printf(a, s...)
}

func Fail(s ...interface{}) {
	log.Println(s...)
	os.Exit(1)
}
