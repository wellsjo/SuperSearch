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
var numFiles1 = 10
var linesPerFile1 = 10
var numFiles2 = 100
var linesPerFile2 = 1000
var testDir = setupSearchFolder(numFiles1, linesPerFile1)
var testDir2 = setupSearchFolder(numFiles2, linesPerFile2)

func TestMain(m *testing.M) {
	defer func() {
		os.Remove(testDir)
		os.Remove(testDir2)
	}()
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
	for i := 0; i < lines; i++ {
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
	s := New(&Options{
		Pattern:      "fox",
		Location:     testDir,
		Quiet:        true,
		Unrestricted: true,
		ShowStats:    true,
	})
	s.Run()
	assert.Equal(t, numFiles1*linesPerFile1, int(s.numMatches),
		fmt.Sprintf("there should be %d matches", numFiles1*linesPerFile1))
}

func BenchmarkSearchDynamicConcurrency(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:  "fox",
			Location: testDir,
			Quiet:    true,
		})
		s.Run()
	}
}

func BenchmarkSearchDynamicConcurrencyLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:  "fox",
			Location: testDir2,
			Quiet:    true,
		})
		s.Run()
	}
}

func BenchmarkSearchStatsOff(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(&Options{
			Pattern:   "fox",
			Location:  testDir,
			Quiet:     true,
			ShowStats: false,
		})
		s.Run()
	}
}

// func BenchmarkBufferSize0(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:   "fox",
// 			Location:  testDir,
// 			Quiet:     true,
// 			ShowStats: false,
// 			bufSize:   0,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkBufferSize1(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:   "fox",
// 			Location:  testDir,
// 			Quiet:     true,
// 			ShowStats: false,
// 			bufSize:   1,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkBufferSize2(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:   "fox",
// 			Location:  testDir,
// 			Quiet:     true,
// 			ShowStats: false,
// 			bufSize:   2,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkBufferSize4(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:   "fox",
// 			Location:  testDir,
// 			Quiet:     true,
// 			ShowStats: false,
// 			bufSize:   4,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall2(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 2,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall4(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 4,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall8(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 8,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall16(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 16,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall32(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 32,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencySmall64(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir,
// 			Quiet:       true,
// 			Concurrency: 64,
// 		})
// 		s.Run()
// 	}
// }

// // Large dir
// func BenchmarkSearchConcurrencyLarge1(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 1,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge2(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 2,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge4(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 4,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge8(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 8,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge16(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 16,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge32(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 32,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchConcurrencyLarge64(b *testing.B) {
// 	b.Skip()
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:     "fox",
// 			Location:    testDir2,
// 			Quiet:       true,
// 			Concurrency: 64,
// 		})
// 		s.Run()
// 	}
// }

// The following tests determined that maxWorkers should be numCPU
// However these don't compile anymore

// func BenchmarkSearchDynamicConcurrencyMax4(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:    "fox",
// 			Location:   testDir2,
// 			Quiet:      true,
// 			MaxWorkers: 4,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchDynamicConcurrencyMax8(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:    "fox",
// 			Location:   testDir2,
// 			Quiet:      true,
// 			MaxWorkers: 8,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchDynamicConcurrencyMax9(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:    "fox",
// 			Location:   testDir2,
// 			Quiet:      true,
// 			MaxWorkers: 9,
// 		})
// 		s.Run()
// 	}
// }

// func BenchmarkSearchDynamicConcurrencyMax16(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		s := New(&Options{
// 			Pattern:    "fox",
// 			Location:   testDir2,
// 			Quiet:      true,
// 			MaxWorkers: 16,
// 		})
// 		s.Run()
// 	}
// }
