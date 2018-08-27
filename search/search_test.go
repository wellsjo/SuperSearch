package search

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// var testDir string
var numFiles = 10
var linesPerFile = 100
var testDir = setupSearchFolder(numFiles, linesPerFile)

func TestMain(m *testing.M) {
	defer os.Remove(testDir)
	os.Exit(m.Run())
}

// Creates one or many directories of dummy files used for testing search
// numDirs will create that many directories, each of which contains
// numFiles files
func setupSearchFolder(numFiles, linesPerFile int) string {
	tmpdir, err := ioutil.TempDir("", "ss-test")
	if err != nil {
		log.Fatal(err)
	}
	for j := 0; j < numFiles; j++ {
		createSearchFile(tmpdir, strconv.Itoa(j), linesPerFile)
	}
	return tmpdir
}

func createSearchFile(dir, file string, lines int) {
	tmpFile, err := ioutil.TempFile(dir, file)
	if err != nil {
		log.Fatal(err)
	}
	var content []byte
	for i := 0; i < linesPerFile; i++ {
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
	fmt.Printf("test searching %v", testDir)
	s := New(&Options{
		Pattern:  "fox",
		Location: testDir,
		Quiet:    true,
	})
	s.Run()
	assert.Equal(t, numFiles*linesPerFile, int(*s.numMatches),
		fmt.Sprintf("there should be %d matches", numFiles*linesPerFile))
}

func BenchmarkSearchConcurrency1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 1,
		})
		s.Run()
	}
}

func BenchmarkSearchConcurrency2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 2,
		})
		s.Run()
	}
}

func BenchmarkSearchConcurrency4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 4,
		})
		s.Run()
	}
}

func BenchmarkSearchConcurrency8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 8,
		})
		s.Run()
	}
}

func BenchmarkSearchConcurrency16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 16,
		})
		s.Run()
	}
}

func BenchmarkSearchConcurrency32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:     "fox",
			Location:    testDir,
			Quiet:       true,
			Concurrency: 32,
		})
		s.Run()
	}
}
