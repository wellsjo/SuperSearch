package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	dirname := "." + string(filepath.Separator)
	args := os.Args[1:]
	fmt.Println("searching", dirname, "for", args)
	numFiles := 0
	done := make(chan bool)
	filepath.Walk(dirname, func(f string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			numFiles++
			go search(f, done)
		}
		return nil
	})
	for i := 0; i < numFiles; i++ {
		<-done
	}
}

func search(f string, done chan bool) {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	done <- true
}
