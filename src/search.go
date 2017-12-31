package search

import (
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/exp/mmap"
	"os"
	"regexp"
	"sync"
	"unicode/utf8"

	"github.com/MichaelTJones/walk"
	"github.com/wellsjo/search/src/log"
)

var (
	ignorePatternFiles = [...]string{".gitignore"}
	ignorePatterns     = []string{}
	ignoreRegexps      []*regexp.Regexp
	searchRegexp       *regexp.Regexp
	concurrency        = 20
	highlightMatch     = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile      = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber    = color.New(color.FgGreen).Add(color.Bold)
)

func init() {
	for _, p := range ignorePatterns {
		ignoreRegexps = append(ignoreRegexps, regexp.MustCompile(p))
	}
}

func Search(pattern string, location string) {
	searchRegexp = regexp.MustCompile(pattern)
	run := make(chan bool, concurrency)
	var wg sync.WaitGroup
	walk.Walk(location, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return nil
		}
		for _, r := range ignoreRegexps {
			if r.MatchString(path) {
				return nil
			}
		}
		if fi.Mode().IsRegular() {
			wg.Add(1)
			run <- true
			go func() {
				search(path, run)
				wg.Done()
			}()
		}
		return nil
	})
	wg.Wait()
}

// Read a file by lines using a buffer
// func processFileBuffer(file string) error {
// 	fi, err := os.Open(file)
// 	if err != nil {
// 		return err
// 	}
// 	defer fi.Close()
// 	reader := bufio.NewReader(fi)
// 	lineNo := 1
// 	var s string
// 	for {
// 		line, err := reader.ReadSlice('\n')
// 		// TODO come up with faster way to do this
// 		// if !utf8.Valid(line) {
// 		// 	return nil
// 		// }
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			} else {
// 				fmt.Println(err, file, lineNo)
// 				panic("buff full")
// 				// continue
// 			}
// 		}
// 		s += *(processLine(line, &lineNo))
// 		lineNo++
// 	}
// 	if s != "" {
// 		highlightFile.Println(file)
// 		fmt.Print(s)
// 	}
// return nil
// }

func readLines(path string, f func()) {
	reader, err := mmap.Open(file)
	if err != nil {
		panic(err)
	}
	if isBin(reader) {
		return
	}
	lastIndex := 0
	lineNo := 1
	output := ""
	for b := 0; b < reader.Len(); b++ {
		if reader.At(b) == '\n' {
			var line = make([]byte, b-lastIndex+1)
			bytesRead, err := reader.ReadAt(line, int64(lastIndex))
			processLine(line, &lineNo, &output)
			lastIndex = b + 1
			if err != nil {
				log.Debug("bytesRead", bytesRead)
				panic(err)
			}
			lineNo++
		}
	}
	if len(output) > 0 {
		highlightFile.Println(file)
		fmt.Println(output)
	}
	err = reader.Close()
	if err != nil {
		panic(err)
	}
}

func isBin(r *mmap.ReaderAt) bool {
	var bytes = make([]byte, 4)
	r.ReadAt(bytes, 0)
	return !utf8.Valid(bytes)
}

func processLine(line []byte, lineNo *int, output *string) {
	// TODO maybe move this out to processFile
	ixs := searchRegexp.FindAllIndex(line, -1)
	if ixs != nil {
		*output += highlightNumber.Sprint(*lineNo, ":")
		lastIndex := 0
		for _, i := range ixs {
			*output += fmt.Sprint(string(line[lastIndex:i[0]]))
			*output += highlightMatch.Sprint(string(line[i[0]:i[1]]))
			lastIndex = i[1]
		}
		*output += fmt.Sprint(string(line[lastIndex:]))
	}
}

// func search(f string, done chan bool) error {
// func search(f string, done chan bool, run chan bool) {
func search(f string, run chan bool) {
	processFileMMAP(f)
	<-run
}
