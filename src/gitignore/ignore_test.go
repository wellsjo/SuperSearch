package gitignore

import (
	"testing"
)

func TestExtension(t *testing.T) {
	ig := NewGitIgnore()
	file := "foo.js"
	pattern := "*.js"
	ig.AddIgnorePattern(pattern)
	if !ig.Test(file) {
		t.Errorf("Extension test failed for %s", pattern)
	}

	ig = NewGitIgnore()
	pattern = "*.min.js"
	file = "foo.min.js"
	ig.AddIgnorePattern(pattern)
	if !ig.Test(file) {
		t.Errorf("Extension test failed for %s", pattern)
	}
	if ig.Test(".min.js") {
		t.Errorf("Extension test failed for %s", pattern)
	}
}

func TestRegex(t *testing.T) {
	ig := NewGitIgnore()
	ig.AddIgnorePattern("foo/*/baz")
	if !ig.Test("foo/bar/baz") || !ig.Test("foo/bar/baz/") {
		t.Errorf("Regex test failed for *")
	}
}
