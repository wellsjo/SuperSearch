package gitignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellsjo/SuperSearch/src/log"
)

const (
	commentPrefix = "#"
	coreSection   = "core"
	eol           = "\n"
	excludesfile  = "excludesfile"
	gitDir        = ".git"
	gitignoreFile = ".gitignore"
	gitconfigFile = ".gitconfig"
	systemFile    = "/etc/gitconfig"
)

// ReadIgnoreFile reads a specific git ignore file.
// func ReadIgnoreFile(path []string, ignoreFile string) (ps []Pattern, err error) {
func ReadIgnoreFile(p string) (ps []Pattern, err error) {
	log.Debug("Loading ignore file %v", p)
	parts := strings.Split(p, string(filepath.Separator))
	path := parts[1 : len(parts)-1]

	f, err := os.Open(p)
	if err == nil {
		defer f.Close()

		if data, err := ioutil.ReadAll(f); err == nil {
			for _, s := range strings.Split(string(data), eol) {
				log.Debug("gitignore processing %v", s)
				if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
					ps = append(ps, ParsePattern(s, path))
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}
