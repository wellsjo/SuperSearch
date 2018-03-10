package ignore

import (
	"bytes"
	"log"
	// "path/filepath"
	"regexp"
	"strings"

	"github.com/danwakefield/fnmatch"
	"golang.org/x/exp/mmap"
)

// TODO can we get rid of ']'?
const functionChars string = "!*?[\\"
const regexChars string = "$(*+.?[^{|"

type GitIgnore struct {
	ignorePatterns []string
}

func NewGitIgnore() *GitIgnore {
	return &GitIgnore{
		ignorePatterns: make([]string, 0),
		// extensions: make(map[string]bool),
	}
}

// Loads ignore patterns from a file
func NewGitIgnoreFromFile(file string) []*regexp.Regexp {
	reader, err := mmap.Open(file)
	if err != nil {
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
			ignores = append(ignores, regexp.MustCompile(string(line)))
		}
	}
	return ignores
}

func (ig *GitIgnore) AddIgnorePattern(pattern string) {
	ig.ignorePatterns = append(ig.ignorePatterns, pattern)
}

func (ig *GitIgnore) Match(filename string) bool {
	for _, p := range ig.ignorePatterns {
		log.Println("testing", filename, "against", p)
		if fnmatch.Match(p, filename, 0) {
			return true
		}
	}
	return false
	// if ig.extensions[extension(filename)] {
	// 	return true
	// } else {
	// 	for r := range ig.regexes {
	// 		matched, err := filepath.Match(r, filename)
	// 		if err != nil {
	// 			panic(err) // TODO change this
	// 		}
	// 		if mached {
	// 			return true
	// 		}
	// 	}
	// }
	// return false
}

// Return extension of filename: foo.js -> js
func extension(filename string) string {
	if filename[0] == '.' {
		filename = filename[1:]
	}
	i := strings.Index(filename, ".")
	if i == -1 {
		return ""
	} else {
		return filename[i+1:]
	}
}

func isFnMatch(filename string) bool {
	return strings.ContainsAny(filename, functionChars)
}

func isRegex(query string) bool {
	return strings.ContainsAny(query, regexChars)
}
