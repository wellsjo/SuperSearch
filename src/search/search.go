package search

import (
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
	"golang.org/x/exp/mmap"
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
)

type Options struct {
	Usage    string
	Pattern  string
	Location string

	Quiet        bool `short:"q" long:"quiet" description:"Doesn't log any matches, just the results summary"`
	Hidden       bool `long:"hidden" description:"Search hidden files"`
	Unrestricted bool `short:"U" long:"unrestricted" description:"Search all files (ignore .gitignore)"`
	Debug        bool `short:"D" long:"debug" description:"Show verbose debug information"`
	Stats        bool `long:"stats" description:"Show stats (# matches, files searched, time taken, etc.)"`
}

type searchFile struct {
	path  string
	index uint64
}

type printFile struct {
	output string
	index  uint64
}

type SuperSearch struct {
	opts *Options

	searchRegexp *regexp.Regexp

	searchQueue chan *searchFile
	workerQueue chan *searchFile
	printQueue  chan *printFile

	numMatches    uint64
	filesMatched  uint64
	filesSearched uint64

	wg         *sync.WaitGroup
	duration   time.Duration
	numWorkers uint64
}

func New(opts *Options) *SuperSearch {
	if opts.Debug {
		log.DebugMode = true
	}
	log.Debug("Searching %q for %q", opts.Location, opts.Pattern)
	return &SuperSearch{
		searchRegexp: regexp.MustCompile(opts.Pattern),
		opts:         opts,

		searchQueue: make(chan *searchFile),
		workerQueue: make(chan *searchFile),
		printQueue:  make(chan *printFile),
		wg:          new(sync.WaitGroup),
	}
}

// Main program logic
func (ss *SuperSearch) Run() {
	start := time.Now()
	go ss.processFiles()
	go ss.printLoop()
	ss.findFiles()
	ss.wg.Wait()

	// Workers have finished
	close(ss.searchQueue)

	ss.wg.Add(1)

	close(ss.printQueue)

	ss.wg.Wait()

	ss.duration = time.Since(start)
	if ss.opts.Stats {
		ss.printStats()
	}
}

// This runs in its own goroutine, and recieves files from ss.findFiles() through searchQueue.
// When a file is recieved, it either gives it to a ready worker, or creates a new worker.
// This will block if all workers are busy and numWorkers == maxWorkers.
func (ss *SuperSearch) processFiles() {
PROCESSLOOP:
	for {
		log.Debug("Waiting on search queue..")
		p := <-ss.searchQueue
		if p == nil {
			break PROCESSLOOP
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
	log.Debug("Closing worker queue...")
	close(ss.workerQueue)
}

// This runs in its own goroutine.
func (ss *SuperSearch) printLoop() {
	var (
		i     uint64 = 1
		print        = make(map[uint64]string)
	)
	for {
		p := <-ss.printQueue
		if p == nil {
			break
		}
		print[p.index] = p.output
		var output strings.Builder
		for {
			out, ok := print[i]
			if ok {
				log.Debug("Adding output to string builder")
				output.WriteString(out)
				i++
			} else {
				break
			}
		}
		if output.Len() > 0 {
			fmt.Print(output.String())
		}
	}
	log.Debug("Print loop done")
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
		log.Debug("Testing %v against ignore rules", path)
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
			}
		}
	}

	log.Debug("Finished scanning directory %v", dir)
}

// These run in parallel, taking files off of the searchQueue channel until it
// is finished
func (ss *SuperSearch) newWorker() {
	atomic.AddUint64(&ss.numWorkers, 1)
	workerNum := ss.numWorkers
	log.Debug("Started worker %v", ss.numWorkers)
	go func() {
		for {
			log.Debug("Worker %v ready...", workerNum)
			sf := <-ss.workerQueue
			if sf == nil {
				break
			}
			log.Debug("Worker %v searching %v", workerNum, sf.path)
			ss.searchFile(sf)
		}
		log.Debug("Worker %v finished", workerNum)
	}()
}

func (ss *SuperSearch) searchFile(sf *searchFile) {
	log.Debug("Opening %v", sf.path)
	defer ss.wg.Done()

	file, err := mmap.Open(sf.path)
	if err != nil {
		return
	}
	defer file.Close()

	if isBin(file) {
		log.Debug("Skipping binary file")
		return
	}

	if file.Len() == 0 {
		log.Debug("Skipping empty file")
		return
	}

	var output strings.Builder
	matchFound := false
	lastIndex := 0
	lineNo := 1
	buf := make([]byte, file.Len())
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		return
	}

	for i := 0; i < len(buf); i++ {
		if buf[i] == '\n' {
			var line = buf[lastIndex:i]
			ixs := ss.searchRegexp.FindAllIndex(line, -1)

			if ixs != nil {
				if !matchFound {
					matchFound = true
					if ss.opts.Stats {
						atomic.AddUint64(&ss.filesMatched, 1)
					}
					output.WriteString(highlightFile.Sprintf("%v\n", sf.path))
				}

				if ss.opts.Stats {
					atomic.AddUint64(&ss.numMatches, 1)
				}

				// Print line number, followed by each match
				output.WriteString(highlightNumber.Sprintf("%v:", lineNo))
				lastIndex := 0

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
	}

	if output.Len() > 0 {
		ss.printQueue <- &printFile{
			output: output.String(),
			index:  sf.index,
		}
	}
}

// Determine if file is binary by checking if it is valid utf8
func isBin(file *mmap.ReaderAt) bool {
	var (
		offsetStart = file.Len() / 3
		offsetEnd   = file.Len() / 2
	)
	var buf = make([]byte, offsetEnd-offsetStart)
	file.ReadAt(buf, int64(offsetStart))
	return !utf8.Valid(buf)
}

func (ss *SuperSearch) printStats() {
	p := message.NewPrinter(language.English)
	p.Printf("%v matches\n%v files contained matches\n%v files searched\n%v seconds",
		ss.numMatches, ss.filesMatched, ss.filesSearched, ss.duration.Seconds())
}
