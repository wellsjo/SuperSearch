package ignore

import (
	"testing"
)

func TestExtension(t *testing.T) {
	ig := NewGitIgnore()
	file := "foo.js"
	pattern := "*.js"
	ig.AddIgnorePattern(pattern)
	if !ig.Match(file) {
		t.Errorf("Extension test failed for %s with %s", file, pattern)
	}

	ig = NewGitIgnore()
	pattern = "*.min.js"
	file = "foo.min.js"
	ig.AddIgnorePattern(pattern)
	if !ig.Match(file) {
		t.Errorf("Extension test failed for %s with %s", file, pattern)
	}
	if ig.Match("min.js") {
		t.Errorf("Extension test failed for %s with %s", file, pattern)
	}
}

func TestRegex(t *testing.T) {
	ig := NewGitIgnore()
	pattern := "foo/**/baz"
	ig.AddIgnorePattern(pattern)
	if !ig.Match("foo/bar/baz") {
		t.Errorf("Extension test failed for %s", pattern)
	}
	if !ig.Match("foo/bar/baz/") {
		t.Errorf("Extension test failed for %s", pattern)
	}
}
