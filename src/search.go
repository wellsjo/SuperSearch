package search

import (
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
		ss.sem <- true
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

func (ss *SuperSearch) SearchFile(path string) {
	Debug("Goroutine created. Searching file", path)
	file, err := mmap.Open(path)
	if err != nil {
		Debug("Failed to open file with mmap", path)
		panic(err)
	}
	if !isBin(file) && file.Len() > 0 {
		lastIndex := 0
		lineNo := 1
		buf := make([]byte, file.Len())
		bytesRead, err := file.ReadAt(buf, 0)
		if err != nil {
			Debug("Failed to read file", path+".", "Read", bytesRead, "bytes.")
			panic(err)
		}
		filePrinted := false
		for i := 0; i < len(buf); i++ {
			if buf[i] == '\n' {
				var line = buf[lastIndex:i]
				ixs := ss.searchRegexp.FindAllIndex(line, -1)
				var output string
				if ixs != nil {
					if !filePrinted {
						highlightFile.Println(path)
						filePrinted = true
					}
					output = highlightNumber.Sprint(lineNo, ":")
					lastIndex := 0
					for _, i := range ixs {
						output += fmt.Sprint(string(line[lastIndex:i[0]]))
						output += highlightMatch.Sprint(string(line[i[0]:i[1]]))
						lastIndex = i[1]
					}
					output += fmt.Sprintln(string(line[lastIndex:]))
				}
				if len(output) > 0 {
					fmt.Print(output)
				}
				lastIndex = i + 1
				lineNo++
			}
		}
		if filePrinted {
			fmt.Println()
		}
	}
	err = file.Close()
	if err != nil {
		panic(err)
	}
	Debug("Closing file search goroutine", path)
	<-ss.sem
	ss.wg.Done()
}

func isBin(file *mmap.ReaderAt) bool {
	var offsetLen int64 = int64(file.Len()) / 4
	var offset int64 = 0
	var buf = make([]byte, 4)
	for i := 0; i < 4; i++ {
		file.ReadAt(buf, offset)
		if !utf8.Valid(buf) {
			return true
		}
		offset += offsetLen
	}
	return false
}
