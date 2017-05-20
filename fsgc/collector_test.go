package fsgc

import (
	"testing"
	"regexp"
	"time"
	"fmt"
	"os"
	"path/filepath"
)

type fileInfo struct {
	os.FileInfo
	modTime time.Time
}

func (f fileInfo) ModTime() time.Time {
	return f.modTime
}

var settings = CollectorSettings{
	MarkerRegexp: regexp.MustCompile(`.*ttl=(?P<value>[0-9]+)(?P<suffix>h?)`),
	SuffixToDuration: map[string]time.Duration{
		"h": time.Hour,
		"":  24 * time.Hour,
	},
}

var collector = Collector{
	Root:              ".",
	CollectorSettings: settings,
}

func TestCollector_retrieveTTL(t *testing.T) {
	tests := []struct {
		input   string
		wantTTL time.Duration
		wantOK  bool
	}{
		{"/path/to/dir.ttl=1", 24 * time.Hour, true},
		{"/path/to/dittl=2qdqd", 48 * time.Hour, true},
		{"/ttl=3", 72 * time.Hour, true},
		{"ttl=3/adad/ad", 72 * time.Hour, true},
		{"/ttl=3/adad/ad/ttl=1", 24 * time.Hour, true},
		{"/path/to/ttl=1", 24 * time.Hour, true},
		{"/path/to/ttl=", 0, false},
		{"path/to/dir", 0, false},
		{"/path/tt/l=3ir", 0, false},
	}

	for _, test := range tests {
		gotTTL, gotOK := collector.retrieveTTL(test.input)
		if gotTTL != test.wantTTL || gotOK != test.wantOK {
			t.Errorf("%v, %v = retrieveTTL(%q), expected %v", gotTTL, gotOK, test.input, test.wantTTL)
		}
	}
}

func ExampleCollector_Collect() {
	oldOsRemoveAll := osRemoveAll
	oldFilepathWalk := filepathWalk
	defer func() { osRemoveAll = oldOsRemoveAll }()
	defer func() { filepathWalk = oldFilepathWalk }()

	osRemoveAll = func(path string) error {
		fmt.Println(path)
		return nil
	}

	now := time.Now()

	tests := []struct {
		fileInfo
		path string
	}{
		{fileInfo{modTime: now.Add(-25 * time.Hour)}, "/a/b/ttl=1"},
		{fileInfo{modTime: now.Add(-25 * time.Hour)}, "/a/b/ttl=2"},
		{fileInfo{modTime: now.Add(-3 * time.Hour)}, "/a/b/ttl=2h"},
		{fileInfo{modTime: now.Add(-1 * time.Hour)}, "/a/b/ttl=2h"},
		{fileInfo{modTime: now.Add(-2 * time.Hour)}, "/a/b/ttl=2/doc-ttl=1h.md"},
	}

	filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
		for _, test := range tests {
			walkFn(test.path, test, nil)
		}
		return nil
	}
	collector.Collect()

	// Output:
	// /a/b/ttl=1
	// /a/b/ttl=2h
	// /a/b/ttl=2/doc-ttl=1h.md
}
