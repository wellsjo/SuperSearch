package gitignore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wellsjo/SuperSearch/src/logger"
)

var testDir string

var ignoreFile1 = ".gitignore"

func TestMain(m *testing.M) {
	testDir, err := ioutil.TempDir("", "ss-test")
	if err != nil {
		logger.Fail(err.Error())
	}
	ioutil.WriteFile(filepath.Join(testDir, "test"), []byte{}, 0644)
	defer func() {
		os.Remove(testDir)
	}()
	os.Exit(m.Run())
}

func TestGitignore(t *testing.T) {
	// patterns := gitignore.ReadIgnoreFile(testDir)
}
