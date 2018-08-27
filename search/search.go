package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/juju/errors"
	"golang.org/x/exp/mmap"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
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

type Options struct {
	Usage       string
	Pattern     string
	Location    string
	Quiet       bool `short:"q" long:"quiet" description:"Doesn't log any matches, just the results summary"`
	Concurrency int  `short:"c" long:"concurrency" description:"The number of files to process in parallel" default:"8"`

	Hidden bool `long:"hidden" description:"Search hidden files"`

	Unrestricted bool `short:"U" long:"unrestricted" description:"Search all files (ignore .gitignore)"`
	Debug        bool `short:"D" long:"debug" description:"Show verbose debug information"`
}

type SuperSearch struct {
	opts *Options

	searchRegexp  *regexp.Regexp
	searchQueue   chan *string
	numMatches    *uint64
	filesMatched  *uint64
	filesSearched *uint64

	wg *sync.WaitGroup
}

func New(opts *Options) *SuperSearch {
	debug("Searching %q for %q", opts.Location, opts.Pattern)
	debug("Concurrency: %v", concurrency)
	var (
		numMatches, filesMatched, filesSearched uint64
	)
	return &SuperSearch{
		searchRegexp:  regexp.MustCompile(opts.Pattern),
		opts:          opts,
		numMatches:    &numMatches,
		filesMatched:  &filesMatched,
		filesSearched: &filesSearched,

		// Allow enough files in the buffer so that there will always be plenty
		// for the worker threads. This is an arbitrary large number.
		searchQueue: make(chan *string, 4096),
		wg:          new(sync.WaitGroup),
	}
}

func (ss *SuperSearch) Run() error {
	ss.wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go ss.worker()
	}
	ss.findFiles()
	close(ss.searchQueue)
	ss.wg.Wait()
	if !ss.opts.Quiet {
		ss.printResults()
	}
	return nil
}

func (ss *SuperSearch) findFiles() error {
	fi, err := os.Stat(ss.opts.Location)
	if err != nil {
		return err
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		ss.scanDir(&ss.opts.Location)
	case mode.IsRegular():
		ss.searchQueue <- &ss.opts.Location
	}
	return nil
}

// Recursively go through directory, sending all files into searchQueue
func (ss *SuperSearch) scanDir(dir *string) error {
	debug("Scanning directory %v", *dir)
	ignores, _ := NewGitIgnoreFromFile(*dir + "/.gitignore")
	dirInfo, err := ioutil.ReadDir(*dir)
	if err != nil {
		return errors.Annotate(err, "io error: failed to read directory")
	}
	for _, fi := range dirInfo {
		if fi.Name()[0] == '.' {
			debug("Skipping hidden file %v", fi.Name())
			continue
		}
		if ignores.Match(fi.Name()) {
			debug("skipping gitignore match %v", fi.Name())
			continue
		}
		path := filepath.Join(*dir, fi.Name())
		if fi.IsDir() {
			ss.scanDir(&path)
		} else if fi.Mode().IsRegular() {
			ss.searchQueue <- &path
			debug("Queuing %v", path)
		}
	}
	debug("Finished scanning directory %v", *dir)
	return nil
}

// These run in parallel, taking files off of the searchQueue channel until it
// is finished
func (ss *SuperSearch) worker() {
	debug("Started worker")
	var output strings.Builder
	for path := range ss.searchQueue {
		ss.searchFile(path, &output)
	}
	if !ss.opts.Quiet && output.Len() > 0 {
		fmt.Print(output.String())
	}
	ss.wg.Done()
}

func (ss *SuperSearch) searchFile(path *string, output *strings.Builder) error {
	file, err := mmap.Open(*path)
	if err != nil {
		return errors.Annotate(err, "Failed to open file with mmap")
	}
	defer file.Close()

	atomic.AddUint64(ss.filesSearched, 1)

	if isBin(file) {
		debug("Skipping binary file")
		return nil
	}

	if file.Len() == 0 {
		debug("Skipping empty file")
		return nil
	}

	lastIndex := 0
	lineNo := 1
	buf := make([]byte, file.Len())
	bytesRead, err := file.ReadAt(buf, 0)

	if err != nil {
		return errors.Annotate(err, fmt.Sprint("Failed to read file", *path+".", "Read", bytesRead, "bytes."))
	}

	matchFound := false

	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' {
			var line = buf[lastIndex:i]
			ixs := ss.searchRegexp.FindAllIndex(line, -1)

			if ixs != nil {
				if !matchFound {
					matchFound = true
					atomic.AddUint64(ss.filesMatched, 1)
					output.Write([]byte(highlightFile.Sprintf("%v\n", *path)))
				}
				// Increase match counter
				atomic.AddUint64(ss.numMatches, 1)
				// Print line number, followed by each match
				output.Write([]byte(highlightNumber.Sprintf("%v:", lineNo)))
				lastIndex := 0

				for _, i := range ixs {
					output.Write([]byte(fmt.Sprint(string(line[lastIndex:i[0]]))))
					output.Write([]byte(highlightMatch.Sprint(string(line[i[0]:i[1]]))))
					lastIndex = i[1]
				}
				output.Write([]byte(fmt.Sprintln(string(line[lastIndex:]))))
			}

			lastIndex = i + 1
			lineNo++
		}
	}

	if matchFound {
		output.Write([]byte("\n"))
	}

	return nil
}

// Cheap (at the expense of being janky) way to determine if a file is binary
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

func (ss *SuperSearch) printResults() {
	p := message.NewPrinter(language.English)
	matchesPlural := "s"
	if *ss.numMatches == 1 {
		matchesPlural = ""
	}
	filesPlural := "s"
	if *ss.filesMatched == 1 {
		filesPlural = ""
	}
	p.Printf("%v matche%s found in %v file%s (%v total)",
		*ss.numMatches, matchesPlural, *ss.filesMatched,
		filesPlural, *ss.filesSearched)
}
