package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/juju/errors"
	"golang.org/x/exp/mmap"
)

var (
	ignoreFilePatterns   = []string{}
	globalIgnoreFiles    = [...]string{".gitignore_global"}
	ignoreFiles          = [...]string{".gitignore"}
	globalIgnorePatterns = []*regexp.Regexp{}

	// Set concurrency to # cores
	concurrency = runtime.NumCPU()

	highlightMatch  = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile   = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber = color.New(color.FgGreen).Add(color.Bold)
)

type PrintData struct {
	file string
	data string
}

type SuperSearch struct {
	searchRegexp *regexp.Regexp
	location     *string
	searchFiles  chan *string
	printData    chan *PrintData
	// filesFinished chan *string

	// Signal channels
	done chan struct{}

	// Global wait group
	wg *sync.WaitGroup

	err chan error
}

func New() *SuperSearch {
	Debug("Searching", Opts.location, "for", Opts.pattern)
	Debug("Concurrency", *Opts.concurrency)
	return &SuperSearch{
		searchRegexp: regexp.MustCompile(Opts.pattern),
		location:     &Opts.location,
		printData:    make(chan *PrintData),

		// Allow enough files in the buffer so that there will always be plenty
		// for the worker threads
		searchFiles: make(chan *string, *Opts.concurrency*2),

		// filesFinished: make(chan *string),
		done: make(chan struct{}),

		wg:  new(sync.WaitGroup),
		err: make(chan error),
	}
}

func (ss *SuperSearch) Run() {
	for i := 0; i < *Opts.concurrency; i++ {
		go func(i int) {
			ss.worker(&i)
		}(i)
	}
	ss.findFiles()
}

// func (ss *SuperSearch) printer() {
// 	var dataToPrint = make(map[string][]string)
// 	var finishedFiles = make(map[*string]bool)
// 	var curFile string
// printLoop:
// 	for {
// 		select {
// 		case pd := <-ss.printData:
// 			dataToPrint[pd.file] = append(dataToPrint[pd.file], pd.data)
// 		case finished := <-ss.filesFinished:
// 			finishedFiles[finished] = true
// 		case <-ss.done:
// 			break printLoop
// 		default:
// 			if len(dataToPrint[curFile]) > 0 {
// 				fmt.Print(strings.Join(dataToPrint[curFile], ""))
// 				delete(dataToPrint, curFile)
// 			}
// 			if finishedFiles[curFile] {
// 				delete(finishedFiles, curFile)
// 				fmt.Println()
// 				curFile = ""
// 			}
// 			if curFile == "" {
// 				for i := range dataToPrint {
// 					curFile = i
// 					highlightFile.Println(curFile)
// 					break
// 				}
// 			}
// 		}
// 	}
// 	ss.wg.Done()
// }

func (ss *SuperSearch) findFiles() {
	fi, err := os.Stat(*ss.location)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		ss.err <- ss.ScanDir(ss.location)
	case mode.IsRegular():
		ss.searchFiles <- ss.location
	}
}

func (ss *SuperSearch) ScanDir(dir *string) error {
	Debug("Scanning directory %v", dir)
	dirInfo, err := ioutil.ReadDir(*dir)
	if err != nil {
		return errors.Annotate(err, "io error: failed to read directory")
	}
	for _, fi := range dirInfo {
		if fi.Name()[0] == '.' {
			continue
		}
		path := filepath.Join(*dir, fi.Name())
		if fi.IsDir() {
			ss.ScanDir(&path)
		} else if fi.Mode().IsRegular() {
			ss.searchFiles <- &path
			Debug("Queuing %v", path)
		}
	}
	Debug("Scan dir finished %v", dir)
	return nil
}

func (ss *SuperSearch) worker(num *int) {
	Debug("Started worker %d", *num)
	for {
		select {
		case next := <-ss.searchFiles:
			ss.searchFile(next)
		}
	}
}

func (ss *SuperSearch) searchFile(path *string) {
	file, err := mmap.Open(*path)
	if err != nil {
		Fail("Failed to open file with mmap", path)
	}
	defer file.Close()
	if !isBin(file) && file.Len() > 0 {
		lastIndex := 0
		lineNo := 1
		buf := make([]byte, file.Len())
		bytesRead, err := file.ReadAt(buf, 0)
		if err != nil {
			Fail("Failed to read file", *path+".", "Read", bytesRead, "bytes.")
		}
		for i := 0; i < len(buf); i++ {
			if buf[i] == '\n' {
				var line = buf[lastIndex:i]
				ixs := ss.searchRegexp.FindAllIndex(line, -1)
				var output string
				if ixs != nil {
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
					ss.printData <- &PrintData{
						file: *path,
						data: output,
					}
				}
				lastIndex = i + 1
				lineNo++
			}
		}
	}
	Debug("Closing file search goroutine", path)
	// ss.filesFinished <- path
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
