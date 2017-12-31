package log

import (
	"log"
)

func Debug(s ...interface{}) {
	log.Println(s...)
}
