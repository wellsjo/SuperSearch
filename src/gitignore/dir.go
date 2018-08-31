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

// ReadPatterns reads gitignore patterns recursively traversing through the directory
// structure. The result is in the ascending order of priority (last higher).
// func ReadPatterns(path []string) (ps []Pattern, err error) {
// 	ps, _ = ReadIgnoreFile(path, gitignoreFile)

// 	var fis []os.FileInfo
// 	fis, err = ioutil.ReadDir(filepath.Join(path...))
// 	if err != nil {
// 		return
// 	}

// 	for _, fi := range fis {
// 		if fi.IsDir() && fi.Name() != gitDir {
// 			var subps []Pattern
// 			subps, err = ReadPatterns(append(path, fi.Name()))
// 			if err != nil {
// 				return
// 			}

// 			if len(subps) > 0 {
// 				ps = append(ps, subps...)
// 			}
// 		}
// 	}

// 	return
// }

// func LoadPatterns(path string) (ps []Pattern, err error) {
// 	log.Debug("Loading patterns from %v", path)
// 	f, err := os.Open(path)
// 	if err != nil {
// 		if os.IsNotExist(err) {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}

// 	defer gioutil.CheckClose(f, &err)

// 	b, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		return
// 	}

// 	d := config.NewDecoder(bytes.NewBuffer(b))

// 	raw := config.New()
// 	if err = d.Decode(raw); err != nil {
// 		return
// 	}

// 	s := raw.Section(coreSection)
// 	efo := s.Options.Get(excludesfile)
// 	if efo == "" {
// 		return nil, nil
// 	}

// 	ps, err = ReadIgnoreFile(nil, efo)
// 	if os.IsNotExist(err) {
// 		return nil, nil
// 	}

// 	return
// }

// LoadGlobalPatterns loads gitignore patterns from from the gitignore file
// declared in a user's ~/.gitconfig file.  If the ~/.gitconfig file does not
// exist the function will return nil.  If the core.excludesfile property
// is not declared, the function will return nil.  If the file pointed to by
// the core.excludesfile property does not exist, the function will return nil.
//
// The function assumes fs is rooted at the root filesystem.
// func LoadGlobalPatterns() (ps []Pattern, err error) {
// 	usr, err := user.Current()
// 	if err != nil {
// 		return
// 	}

// 	return LoadPatterns(filepath.Join(usr.HomeDir, gitconfigFile))
// }

// LoadSystemPatterns loads gitignore patterns from from the gitignore file
// declared in a system's /etc/gitconfig file.  If the ~/.gitconfig file does
// not exist the function will return nil.  If the core.excludesfile property
// is not declared, the function will return nil.  If the file pointed to by
// the core.excludesfile property does not exist, the function will return nil.
//
// The function assumes fs is rooted at the root filesystem.
// func LoadSystemPatterns(fs billy.Filesystem) (ps []Pattern, err error) {
// 	return LoadPatterns(systemFile)
// }
