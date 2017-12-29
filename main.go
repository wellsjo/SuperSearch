package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	// "mime"
	"os"
	"path/filepath"
	"regexp"
	// "strings"
	"unicode/utf8"
)

var (
	searchPattern   string
	searchLocation  string
	ignorePatterns  = []string{".git*"}
	ignoreRegexps   []*regexp.Regexp
	searchRegexp    *regexp.Regexp
	highlightMatch  = color.New(color.BgYellow).Add(color.FgBlack).Add(color.Bold)
	highlightFile   = color.New(color.FgCyan).Add(color.Bold)
	highlightNumber = color.New(color.FgWhite).Add(color.Bold)
)

func init() {
	for _, p := range ignorePatterns {
		ignoreRegexps = append(ignoreRegexps, regexp.MustCompile(p))
	}
	if len(os.Args) > 1 {
		searchPattern = os.Args[1]
	}
	searchRegexp = regexp.MustCompile(searchPattern)
	if len(os.Args) > 2 {
		searchLocation = os.Args[2]
	}
	if searchLocation == "" {
		searchLocation = "." + string(filepath.Separator)
	}
}

func main() {
	if len(searchPattern) == 0 {
		return
	}
	log.Println("searching", searchLocation, "for", searchPattern)
	numFiles := 0
	done := make(chan bool)
	filepath.Walk(searchLocation, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return nil
		}
		for _, r := range ignoreRegexps {
			if r.MatchString(path) {
				return nil
			}
		}
		if fi.Mode().IsRegular() {
			numFiles++
			go search(path, done)
		}
		return nil
	})
	for i := 0; i < numFiles; i++ {
		<-done
	}
}

func readLinesBuffer(file string) error {
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()
	reader := bufio.NewReader(fi)
	lineNo := 1
	var s string
	for {
		line, err := reader.ReadSlice('\n')
		// TODO come up with faster way to do this
		if !utf8.Valid(line) {
			return nil
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err, file, lineNo)
				panic("buff full")
				// continue
			}
		}
		s += *(processLine(line, &lineNo))
		lineNo++
	}
	if s != "" {
		highlightFile.Println(file)
		fmt.Print(s)
	}
	return nil
}

func processLine(line []byte, lineNo *int) *string {
	ixs := searchRegexp.FindAllIndex(line, -1)
	rs := ""
	if ixs != nil {
		rs = highlightNumber.Sprint(*lineNo, ":")
		lastIndex := 0
		for _, i := range ixs {
			rs += fmt.Sprint(string(line[lastIndex:i[0]]))
			rs += highlightMatch.Sprint(string(line[i[0]:i[1]]))
			lastIndex = i[1]
		}
		rs += fmt.Sprint(string(line[lastIndex:]))
	}
	return &rs
}

func search(f string, done chan bool) error {
	err := readLinesBuffer(f)
	done <- true
	return err
}
