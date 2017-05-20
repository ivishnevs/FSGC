package fsgc

import (
	"regexp"
	"time"
	"strconv"
	"log"
	"path/filepath"
	"os"
)

var osRemoveAll = os.RemoveAll
var filepathWalk = filepath.Walk

type CollectorSettings struct {
	MarkerRegexp     *regexp.Regexp
	SuffixToDuration map[string]time.Duration
}

type Collector struct {
	Root string
	CollectorSettings
}

func (c Collector) retrieveTTL(str string) (ttl time.Duration, ok bool) {
	tmp := make(map[string]string)
	match := c.MarkerRegexp.FindStringSubmatch(str)
	for i, name := range c.MarkerRegexp.SubexpNames() {
		if i != 0 && len(match) > i {
			tmp[name] = match[i]
		}
	}

	suffix, ok := tmp["suffix"]
	if ok {
		value, err := strconv.Atoi(tmp["value"])
		if err != nil {
			log.Printf("The string '%v' cannot be converted to int", tmp["value"])
		}
		ttl = time.Duration(value) * c.SuffixToDuration[suffix]
	}
	return
}

func (c Collector) processPath(path string, f os.FileInfo, err error) error {
	if err != nil {
		return filepath.SkipDir
	}
	if ttl, ok := c.retrieveTTL(filepath.Base(path)); ok {
		if time.Since(f.ModTime()) > ttl {
			if err := osRemoveAll(path); err != nil {
				log.Printf("Cannot delete %v\n", path)
			}
		}
	}

	return nil
}

func (c Collector) Collect() {
	if err := filepathWalk(c.Root, c.processPath); err != nil {
		log.Println(err)
	}
}
