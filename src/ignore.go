package search

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"

	"golang.org/x/exp/mmap"
)

func GetGlobalIgnorePatterns() {
	home := os.Getenv("HOME")
	for _, f := range globalIgnoreFiles {
		globalIgnorePatterns = append(globalIgnorePatterns, LoadIgnorePatterns(filepath.Join(home, f))...)
	}
}

// Loads ignore patterns from a file
func LoadIgnorePatterns(file string) []*regexp.Regexp {
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
