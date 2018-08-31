package search

import (
	"bytes"
	"path/filepath"
	"strings"

	"github.com/wellsjo/SuperSearch/src/log"
	"golang.org/x/exp/mmap"
)

type GitIgnore struct {
	ignorePatterns []string
}

func NewGitIgnore() *GitIgnore {
	return &GitIgnore{
		ignorePatterns: make([]string, 0),
	}
}

// Loads ignore patterns from a file
func NewGitIgnoreFromFile(file string) (*GitIgnore, error) {
	reader, err := mmap.Open(file)
	if err != nil {
		return &GitIgnore{}, err
	}
	var ignores []string
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
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			ignores = append(ignores, (string(line)))
		}
	}
	return &GitIgnore{
		ignorePatterns: ignores,
	}, nil
}

func (ig *GitIgnore) AddIgnorePattern(pattern string) {
	ig.ignorePatterns = append(ig.ignorePatterns, pattern)
}

func (ig *GitIgnore) Match(filename string) bool {
	log.Debug("Testing %v against .gitignore", filename)
	for _, p := range ig.ignorePatterns {
		if matched, _ := filepath.Match(p, filename); matched {
			return true
		}
	}
	return false
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
