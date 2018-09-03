package search

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/wellsjo/SuperSearch/src/gitignore"
	"github.com/wellsjo/SuperSearch/src/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	// Setting maxConcurrency to # cpu cores gives best benchmark results
	maxConcurrency = runtime.NumCPU()
	separator      = string(filepath.Separator)

	highlightMatch  = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile   = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber = color.New(color.FgGreen).Add(color.Bold)

	utf8BOMMarker = []byte{0xEF, 0xBB, 0xBF}
	pdfMarker     = []byte{'%', 'P', 'D', 'F', '-'}
	regexChars    = "[{(*+.?^|\\"
)

type Options struct {
	Usage    string
	Pattern  string
	Location string

	IgnoreCase   bool `short:"i" long:"ignore-case" description:"Ignore case sensitivity when matching"`
	Hidden       bool `long:"hidden" description:"Search hidden files"`
	Unrestricted bool `short:"U" long:"unrestricted" description:"Search all files (ignore .gitignore)"`

	Quiet     bool `short:"q" long:"quiet" description:"Doesn't log any matches, just the results summary"`
	Debug     bool `short:"D" long:"debug" description:"Show verbose debug information"`
	ShowStats bool `long:"stats" description:"Show stats (# matches, files searched, time taken, etc.)"`
}

type searchFile struct {
	path  string
	index uint64
	size  int64
}

type printFile struct {
	output string
	index  uint64
}

type SuperSearch struct {
	opts *Options

	isRegex      bool
	searchRegexp *regexp.Regexp

	searchQueue chan *searchFile
	workerQueue chan *searchFile
	// printQueue  chan *printFile

	// Map of file indexes to empty structs, which the printLoop goroutine uses
	// to determine what to print next
	skipFiles *sync.Map

	// These are used for --stats; some of these aren't tracked by default
	numMatches    uint64
	filesMatched  uint64
	filesSearched uint64
	numWorkers    uint64
	duration      time.Duration

	workDir string
	wg      *sync.WaitGroup
}

func New(opts *Options) *SuperSearch {
	log.Debug("Searching %q for %q", opts.Location, opts.Pattern)

	if opts.Debug {
		log.DebugMode = true
	}

	if opts.IgnoreCase {
		log.Debug("Using case insensitive search %v", opts.Pattern)
	}

	var rgx *regexp.Regexp
	isRgx := false
	if isRegex(opts.Pattern) {
		log.Debug("Using regex search")
		rgx = regexp.MustCompile(opts.Pattern)
		isRgx = true
	} else {
		log.Debug("Using Boyer-Moore string search")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fail(err.Error())
	}

	return &SuperSearch{
		searchRegexp: rgx,
		isRegex:      isRgx,
		opts:         opts,
		workDir:      wd,

		searchQueue: make(chan *searchFile),
		workerQueue: make(chan *searchFile),
		// printQueue:  make(chan *printFile),

		skipFiles: new(sync.Map),

		wg: new(sync.WaitGroup),
	}
}

// Main program logic
func (ss *SuperSearch) Run() {
	var start time.Time
	if ss.opts.ShowStats {
		start = time.Now()
	}

	// processFiles takes files from findFiles and delegates them to workers
	// over the searchQueue channel. Workers then search the files and send
	// results over to printLoop, which concatonates as many of the results
	// as it can before printing.
	go ss.processFiles()
	go ss.printLoop()

	// Synchronously finds files and send them into searchQueue,
	// which are then processed by the processFiles goroutine
	ss.findFiles()

	// 1 is added to the WaitGroup for every file processed, then
	// Done() is called after each file is searched.
	ss.wg.Wait()

	// All files have been processed, so we can close these
	close(ss.searchQueue)

	// Wait for printing to finish before exiting
	// ss.wg.Add(1)
	// close(ss.printQueue)
	// ss.wg.Wait()

	if ss.opts.ShowStats {
		ss.duration = time.Since(start)
		ss.printStats()
	}
}

// This runs in its own goroutine, and recieves files from ss.findFiles()
// through searchQueue. When a file is recieved, it either gives it to a ready
// worker, or creates a new worker. This will block if all workers are busy
// and numWorkers == maxWorkers.
func (ss *SuperSearch) processFiles() {
PROCESSLOOP:
	for {
		p := <-ss.searchQueue

		if p == nil {
			break PROCESSLOOP
		}

		if p.size == 0 {
			log.Debug("Skipping empty file %v", p.path)
			continue
		}

		log.Debug("Processing %v", p.path)

		select {
		case ss.workerQueue <- p:
			// no-op

		default:
			if int(ss.numWorkers) < maxConcurrency {
				log.Debug("Workers busy; Creating new worker")
				ss.newWorker()
			} else {
				log.Debug("Workers busy and can't create more; Waiting...")
			}
			ss.workerQueue <- p
		}
	}

	// At this point, all jobs have been given to the workerQueue, and therefore
	// accepted by workers. Closing this will free up the workers.
	close(ss.workerQueue)
}

// This runs in its own goroutine, receiving output strings and indexes.
// As the print loop receives output, it is cached until the current index
// is received. Once that happens, the printer will attempt to concatonate
// the next n subsequent outputs into one string builder for efficiency
// while maintaining order.
func (ss *SuperSearch) printLoop() {
	var (
		// Mapping of indexes to output strings
		// TODO make this global syncmap..get around a channel
		print = make(map[uint64]string)

		// The current print index. The printer will wait until this
		// is recieved before attemptint to print.
		i uint64 = 1

		output strings.Builder
	)

	for {
		output.Reset()

		// Add as many outputs together as we can before printing
		for {
			out, ok := print[i]
			if ok {
				log.DebugGreen("Adding %v to string builder", i)
				output.WriteString(out)
				i++
			}
			if _, ok := ss.skipFiles.Load(i); ok {
				log.DebugGreen("Printer skipping %v", i)
				i++
			} else {
				break
			}
		}

		if !ss.opts.Quiet && output.Len() > 0 {
			fmt.Print(output.String())
		}
	}

	log.Debug("Print loop done\n%v", print)
	fmt.Println(*ss.skipFiles)
	ss.wg.Done()
}

func (ss *SuperSearch) findFiles() {
	fi, err := os.Stat(ss.opts.Location)
	if err != nil {
		log.Fail("invalid location input %v", ss.opts.Location)
	}
	usr, err := user.Current()
	if err != nil {
		log.Fail(err.Error())
	}

	switch mode := fi.Mode(); {

	case mode.IsDir():
		var m gitignore.Matcher
		if !ss.opts.Unrestricted {
			ps, _ := gitignore.ReadIgnoreFile(filepath.Join(usr.HomeDir, ".gitignore_global"))
			m = gitignore.NewMatcher(ps)
		}
		ss.scanDir(ss.opts.Location, m)

	case mode.IsRegular():
		log.Debug("Queuing %v", ss.opts.Location)
		ss.wg.Add(1)
		ss.searchQueue <- &searchFile{
			path:  ss.opts.Location,
			index: 1,
			size:  fi.Size(),
		}
	}
}

// Recursively go through directory, sending all files into searchQueue
func (ss *SuperSearch) scanDir(dir string, m gitignore.Matcher) {
	log.Debug("Scanning directory %v", dir)

	dirInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	if !ss.opts.Unrestricted {
		ps2, _ := gitignore.ReadIgnoreFile(filepath.Join(dir, ".gitignore"))
		if len(ps2) > 0 {
			m.AddPatterns(ps2)
		}
	}

	for _, fi := range dirInfo {
		if !ss.opts.Hidden && fi.Name()[0] == '.' {
			log.Debug("Skipping hidden file %v", fi.Name())
			continue
		}
		path := filepath.Join(dir, fi.Name())
		if !ss.opts.Unrestricted && m.Match(strings.Split(path, separator)[1:], fi.IsDir()) {
			log.Debug("Skipping gitignore match: %v", path)
			continue
		}
		if fi.IsDir() {
			ss.scanDir(path, m)
		} else if fi.Mode().IsRegular() {
			atomic.AddUint64(&ss.filesSearched, 1)
			log.Debug("Queuing %v", path)
			ss.wg.Add(1)
			ss.searchQueue <- &searchFile{
				path:  path,
				index: ss.filesSearched,
				size:  fi.Size(),
			}
		}
	}

	log.Debug("Finished scanning directory %v", dir)
}

// These run in parallel, taking files off of the searchQueue channel until it
// is finished
func (ss *SuperSearch) newWorker() {

	if ss.opts.Debug {
		atomic.AddUint64(&ss.numWorkers, 1)
	}

	workerNum := ss.numWorkers
	log.Debug("Starting worker %v", ss.numWorkers)

	go func() {
		for {
			log.Debug("Worker %v waiting", workerNum)

			sf := <-ss.workerQueue
			if sf == nil {
				break
			}

			log.Debug("Worker %v searching %v", workerNum, sf.path)

			matchFound := ss.searchFile(sf)

			if !matchFound {
				ss.skipFiles.Store(sf.index, struct{}{})
			}
		}

		log.Debug("Worker %v finished", workerNum)
	}()
}

func (ss *SuperSearch) searchFile(sf *searchFile) bool {
	defer ss.wg.Done()

	file, err := os.Open(sf.path)
	if err != nil {
		log.Debug("Failed to open file %v", sf.path)
		return false
	}
	defer file.Close()

	buf := make([]byte, sf.size)
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		return false
	}

	if isBinary(buf) {
		log.Debug("Skipping binary file")
		return false
	}

	if ss.isRegex {
		return ss.searchFileRegex(sf, buf)
	} else {
		return ss.searchFileBoyerMoore(sf, buf)
	}
}

func (ss *SuperSearch) searchFileRegex(sf *searchFile, buf []byte) bool {

	var output strings.Builder
	matchFound := false
	lastIndex := 0
	lineNo := 1

	// TODO regex whole file instead of by line?
	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' {
			var line = buf[lastIndex:i]
			ixs := ss.searchRegexp.FindAllIndex(line, -1)

			// Skip binary files
			if !matchFound && !utf8.Valid(line) {
				return false
			}

			// We found matches
			if ixs != nil {

				if ss.opts.ShowStats {
					atomic.AddUint64(&ss.numMatches, uint64(len(ixs)))
				}

				if !matchFound {
					matchFound = true
					if ss.opts.ShowStats {
						atomic.AddUint64(&ss.filesMatched, 1)
					}
					// Print the file name if we find a match
					output.WriteString(highlightFile.Sprintf("%v\n", sf.path))
				}

				// Print line number, followed by each match
				output.WriteString(highlightNumber.Sprintf("%v:", lineNo))
				lastIndex = 0

				// Loop through match indexes, output highlighted match
				for _, i := range ixs {
					output.Write(line[lastIndex:i[0]])
					output.WriteString(highlightMatch.Sprint(string(line[i[0]:i[1]])))
					lastIndex = i[1]
				}
				output.Write(line[lastIndex:])
				output.WriteRune('\n')
			}

			lastIndex = i + 1
			lineNo++
		}
	}

	if matchFound {
		output.WriteRune('\n')
		fmt.Print(output.String())
		// ss.printQueue <- &printFile{
		// 	output: output.String(),
		// 	index:  sf.index,
		// }
		return true
	}

	return false
}

func (ss *SuperSearch) searchFileBoyerMoore(sf *searchFile, buf []byte) bool {
	matches := BoyerMooreSearch(string(buf), ss.opts.Pattern)

	if len(matches) == 0 {
		return false
	}

	fName := strings.Replace(sf.path, ss.workDir, "", -1)
	if fName[0] == '/' {
		fName = fName[1:]
	}

	atomic.AddUint64(&ss.numMatches, uint64(len(matches)))
	log.Debug("Found matches at %v in %v", matches, fName)

	var output strings.Builder
	lineNo := 1
	matchIndex := 0
	printingLine := false
	lastIndex := 0
	done := false

	output.WriteString(highlightFile.Sprintf("%v\n", fName))

	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' || i == len(buf)-1 {
			if printingLine {
				output.Write(buf[lastIndex:i])
				if done {
					break
				}
				output.WriteRune('\n')
				printingLine = false
			}

			lineNo++
			lastIndex = i + 1
		}

		if done {
			continue
		}

		if i == matches[matchIndex] {

			// Print line number, followed by each match
			if !printingLine {
				output.WriteString(highlightNumber.Sprintf("%v:", lineNo))
			}

			log.DebugGreen("%v %v", lastIndex, i)
			output.Write(buf[lastIndex:i])

			output.WriteString(highlightMatch.Sprint(string(buf[i : i+len(ss.opts.Pattern)])))

			matchIndex++
			if matchIndex == len(matches) {
				done = true
			}

			lastIndex = i + len(ss.opts.Pattern)
			printingLine = true
		}
	}

	output.WriteString("\n")
	fmt.Print(output.String())

	return true
}

func isBinary(buf []byte) bool {
	if len(buf) > 2 && bytes.Compare(buf[:3], utf8BOMMarker) == 0 {
		log.Debug("UTF-8 BOM found")
		return false
	}
	if len(buf) > 4 && bytes.Compare(buf[:5], pdfMarker) == 0 {
		log.Debug("PDF found")
		return true
	}
	maxCheck := 32
	if len(buf) < 32 {
		maxCheck = len(buf)
	}
	return !utf8.Valid(buf[:maxCheck])
}

func isRegex(pattern string) bool {
	return strings.ContainsAny(pattern, regexChars)
}

func (ss *SuperSearch) printStats() {
	p := message.NewPrinter(language.English)
	p.Printf("%v matches\n%v files contained matches\n%v files searched\n%v seconds",
		ss.numMatches, ss.filesMatched, ss.filesSearched, ss.duration.Seconds())
}
