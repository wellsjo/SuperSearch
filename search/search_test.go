package search

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Creates one or many directories of dummy files used for testing search
// numDirs will create that many directories, each of which contains
// numFiles files
func setupSearchFolder(numFiles int) string {
	tmpdir, err := ioutil.TempDir("", "ss-test")
	if err != nil {
		log.Fatal(err)
	}
	for j := 0; j < numFiles; j++ {
		createSearchFile(tmpdir, strconv.Itoa(j))
	}
	return tmpdir
}

func createSearchFile(dir, file string) {
	tmpFile, err := ioutil.TempFile(dir, file)
	if err != nil {
		log.Fatal(err)
	}
	var content []byte
	for i := 0; i < 10000; i++ {
		content = append(content, []byte("The quick brown fox jumped over the lazy dog.\n")...)
	}
	if _, err := tmpFile.Write(content); err != nil {
		log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}
}

func TestSearch(t *testing.T) {
	numFiles := 3
	tmpDir := setupSearchFolder(numFiles)
	defer os.Remove(tmpDir)

	s := New(&Options{
		Pattern:  "fox",
		Location: tmpDir,
		Debug:    true,
		Quiet:    true,
	})

	s.Run()

	assert.Equal(t, numFiles*10000, int(*s.matches), "there should be 1000 matches")
}
