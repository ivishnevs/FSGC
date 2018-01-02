package main

import (
	"log"
	"regexp"
	"time"
	"./fsgc"
	"flag"
)

func main() {
	root := flag.String("root", ".", "the root for cleaning up")
	flag.Parse()
	log.Printf("Starting FSGC on %s", *root)

	settings := fsgc.CollectorSettings{
		MarkerRegexp: regexp.MustCompile(`.*ttl=(?P<value>[0-9]+)(?P<suffix>h?).*`),
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
