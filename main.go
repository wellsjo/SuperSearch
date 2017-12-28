package main

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
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
	log.Println("searching", searchLocation, "for", searchPattern)
	numFiles := 0
	done := make(chan bool)
	filepath.Walk(searchLocation, func(f string, fi os.FileInfo, err error) error {
		for _, r := range ignoreRegexps {
			if r.MatchString(f) {
				return nil
			}
		}
		if !fi.IsDir() {
			numFiles++
			go search(f, done)
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
		if err != nil {
			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}
		s += processLine(line, &lineNo)
		lineNo++
	}
	if s != "" {
		highlightFile.Println(file)
		fmt.Print(s)
	}
	return nil
}

func processLine(line []byte, lineNo *int) string {
	ixs := searchRegexp.FindAllIndex(line, -1)
	if ixs == nil {
		return ""
	}
	var rs = highlightNumber.Sprint(*lineNo, ":")
	lastIndex := 0
	for _, i := range ixs {
		rs += fmt.Sprint(string(line[lastIndex:i[0]]))
		rs += highlightMatch.Sprint(string(line[i[0]:i[1]]))
		lastIndex = i[1]
	}
	return rs + fmt.Sprint(string(line[lastIndex:]))
}

func search(f string, done chan bool) error {
	// then := time.Now()
	err := readLinesBuffer(f)
	// now := time.Now()
	// dur := now.Sub(then)
	// log.Println("dur", dur, '\n')
	done <- true
	return err
}