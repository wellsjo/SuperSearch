package search

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func setupSearchFiles(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "super_search_test")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	var content []byte
	for i := 0; i < 10000; i++ {
		content = append(content, []byte("The quick brown fox jumped over the lazy dog.\n")...)
	}
	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestSearch(t *testing.T) {
	s := NewSuperSearch()
}
