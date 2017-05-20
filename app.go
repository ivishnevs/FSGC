package main

import (
	"regexp"
	"time"
	"./fsgc"
	"flag"
)

func main() {
	root := flag.String("root", ".", "the root for cleaning up")
	flag.Parse()

	settings := fsgc.CollectorSettings{
		MarkerRegexp: regexp.MustCompile(`.*ttl=(?P<value>[0-9]+)(?P<suffix>h?)`),
		SuffixToDuration: map[string]time.Duration{
			"h": time.Hour,
			"":  24 * time.Hour,
		},
	}
	collector := fsgc.Collector{
		Root:              *root,
		CollectorSettings: settings,
	}
	collector.Collect()
}
