package search

import (
	// "bufio"
	"bytes"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/exp/mmap"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"unicode/utf8"
)

var (
	ignoreFilePatterns   = []string{}
	globalIgnoreFiles    = [...]string{".gitignore_global"}
	ignoreFiles          = [...]string{".gitignore"}
	globalIgnorePatterns = []*regexp.Regexp{}
	concurrency          = 64
	highlightMatch       = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile        = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber      = color.New(color.FgGreen).Add(color.Bold)
)

func init() {
	home := os.Getenv("HOME")
	for _, f := range globalIgnoreFiles {
		globalIgnorePatterns = append(globalIgnorePatterns, getIgnorePatterns(filepath.Join(home, f))...)
	}
}

type SuperSearch struct {
	searchRegexp *regexp.Regexp
	location     string
	sem          chan bool
	wg           *sync.WaitGroup
}

func NewSuperSearch() *SuperSearch {
	opts := GetOptions()
	Debug("Searching", opts.location, "for", opts.pattern)
	Debug("Concurrency", *opts.concurrency)
	return &SuperSearch{
		searchRegexp: regexp.MustCompile(opts.pattern),
		location:     opts.location,
		sem:          make(chan bool, *opts.concurrency),
		wg:           new(sync.WaitGroup),
	}
}

func (ss *SuperSearch) Run() {
	fi, err := os.Stat(ss.location)
	if err != nil {
		fmt.Println(err)
		return
	}
	ss.wg.Add(1)
	switch mode := fi.Mode(); {
	case mode.IsDir():
		ss.ScanDir(ss.location)
	case mode.IsRegular():
		ss.SearchFile(ss.location)
	}
	ss.wg.Wait()
}

func (ss *SuperSearch) ScanDir(dir string) {
	Debug("Scanning directory", dir)
	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, fi := range dirInfo {
		if fi.Name()[0] == '.' {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		if fi.IsDir() {
			ss.wg.Add(1)
			go ss.ScanDir(path)
		} else if fi.Mode().IsRegular() {
			ss.wg.Add(1)
			go func() {
				ss.sem <- true
				ss.SearchFile(path)
			}()
		}
	}
	Debug("Goroutine ScanDir", dir, "finished")
	ss.wg.Done()
}

// Loads ignore patterns from a file
func getIgnorePatterns(file string) []*regexp.Regexp {
	Debug("loading ignore patterns from", file)
	reader, err := mmap.Open(file)
	if err != nil {
		Debug("Failed to open ignore file", file)
		panic(err) // TODO remove this
	}
	var ignores []*regexp.Regexp
	for lastIndex, curIndex := 0, 0; curIndex < reader.Len(); curIndex++ {
		if reader.At(curIndex) == '\n' {
			var line = make([]byte, curIndex-lastIndex+1)
			_, err := reader.ReadAt(line, int64(lastIndex))
			lastIndex = curIndex + 1
			if err != nil {
				panic(err) // TODO remove this
			}
			// Ignore comments, whitespace
			line = bytes.TrimSpace(line)
			if line[0] == '#' || len(line) == 0 {
				continue
			}
			Debug("Adding ignore pattern", string(line))
			ignores = append(ignores, regexp.MustCompile(string(line)))
		}
	}
	return ignores
}

func (ss *SuperSearch) SearchFile(path string) {
	Debug("Searching file", path)
	reader, err := mmap.Open(path)
	if err != nil {
		panic(err)
	}
	if !isBin(reader) {
		lastIndex := 0
		lineNo := 1
		output := ""
		for b := 0; b < reader.Len(); b++ {
			if reader.At(b) == '\n' {
				var line = make([]byte, b-lastIndex+1)
				bytesRead, err := reader.ReadAt(line, int64(lastIndex))
				ss.processLine(line, &lineNo, &output)
				lastIndex = b + 1
				if err != nil {
					Debug("bytesRead", bytesRead)
					panic(err)
				}
				lineNo++
			}
		}
		if len(output) > 0 {
			highlightFile.Println(path)
			fmt.Println(output)
		}
		err = reader.Close()
		if err != nil {
			panic(err)
		}
	}
	<-ss.sem
	ss.wg.Done()
}

func isBin(r *mmap.ReaderAt) bool {
	var bytes = make([]byte, 4)
	r.ReadAt(bytes, 0)
	return !utf8.Valid(bytes)
}

func (ss *SuperSearch) processLine(line []byte, lineNo *int, output *string) {
	// TODO maybe move this out to processFile
	ixs := ss.searchRegexp.FindAllIndex(line, -1)
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
// func search(f string, run chan bool) {
// 	readLines(f)
// 	<-run
// }

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
