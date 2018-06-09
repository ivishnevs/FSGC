package fsgc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	FSGC_CONF = "fsgc.json"
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

// Policy is a map from a regexp pattern to a maximum count of files matching
// that pattern which should be retained.  For each pattern, all matching files
// are sorted in reverse order lexicographically and the first @count are
// retained, while all remaining ones are removed.
//
// For example, given the following policy file:
//
// { "master.*.tar.gz": 4 }
//
// And the files master1.tar.gz, master2.tar.gz, master3.tar.gz.
//
// master1.tar.gz and master.2.tar.gz are GC'd while master3.tar.gz is retained.
type Policy map[string]int

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

func (c Collector) processPath(path string, fi os.FileInfo, err error) error {
	if err != nil {
		log.Println(err)
		return filepath.SkipDir
	}
	if fi.IsDir() {
		policy := c.getDirPolicy(path)
		if policy != nil {
			log.Println("Found FSGC policy for dir ", path)
			c.gcDirFiles(path, policy)
		}
		return nil
	}
	if ttl, ok := c.retrieveTTL(path); ok {
		if time.Since(fi.ModTime()) > ttl {
			log.Println("Deleting: ", path)
			if err := osRemoveAll(path); err != nil {
				log.Printf("Cannot delete %v: %s\n", path, err.Error())
			}
		}
	}

	return nil
}

// Returns GC policy for the directory at @path.
func (c Collector) getDirPolicy(path string) Policy {
	fsgcConf := filepath.Join(path, FSGC_CONF)
	if _, err := os.Stat(fsgcConf); os.IsNotExist(err) {
		// Nothing needs to be done for such directories.
		return nil
	}
	data, err := ioutil.ReadFile(fsgcConf)
	if err != nil {
		// Log error and continue.
		log.Printf("Could not read file %s: %v", fsgcConf, err)
		return nil
	}
	var policy Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		// Log error and continue.
		log.Printf("Could not parse json from file %s: %v", fsgcConf, err)
		return nil
	}
	return policy
}

// byModTime implements sort.Interface for []os.FileInfo based on ModTime().
type byModTime []os.FileInfo

func (a byModTime) Len() int           { return len(a) }
func (a byModTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byModTime) Less(i, j int) bool { return a[i].ModTime().Before(a[j].ModTime()) }

// Given @policy, GC files from directory at @path.
func (c Collector) gcDirFiles(path string, policy Policy) {
	// Read all files in directory and sort in reverse order by mtime.
	dirEnts, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("Could not read directory %s: %v", path, err)
		return
	}
	sort.Sort(sort.Reverse(byModTime(dirEnts)))
	log.Printf("Directory has %d entries", len(dirEnts))

	// For each regexp pattern identified in @policy, delete all files beyond the
	// first count as per @policy.
	actual := make(Policy)
	remove := make(map[string]bool)
	for _, entry := range dirEnts {
		if entry.IsDir() {
			continue
		}
		f := entry.Name()
		for pattern, count := range policy {
			if regexp.MustCompile(pattern).Match([]byte(f)) {
				if _, ok := actual[pattern]; ! ok {
					actual[pattern] = 1
				} else {
					actual[pattern] += 1
				}
				if actual[pattern] > count {
					remove[f] = true
				}
			}
		}
	}
	log.Printf("Found %d files for removal", len(remove))

	for f, _ := range remove {
		log.Printf("Deleting file %s", f)
		err = nil
		err := os.Remove(filepath.Join(path, f))
		if err != nil {
			log.Printf("Could not delete file %s: %v", f, err)
		}
	}
}

func (c Collector) Collect() {
	if err := filepathWalk(c.Root, c.processPath); err != nil {
		log.Println(err)
	}
}
