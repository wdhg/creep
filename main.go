package main

import (
	"flag"
	"fmt"
	"log"
	"time"
)

var (
	help        = flag.Bool("h", false, "Show help")
	start       = flag.String("s", "https://news.ycombinator.com", "The site to start crawling from")
	maxCount    = flag.Int("n", 1000, "Number of urls to scrape")
	threadCount = flag.Int("tc", 10, "Number of threads")
	timeout     = flag.Int64("t", 5000, "Timeout for each http request (ms)")
	logging     = flag.Bool("l", false, "Enables logging")
	output      = flag.String("o", "", "Name of output file to write to. If not set will output to terminal")
)

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	crawler := newCrawler(*timeout, *logging)
	crawler.storeAddress(*start)
	startTime := time.Now()
	if crawler.logging {
		log.Printf("Crawling for %d urls...\n", *maxCount)
	}
	for crawler.count < *maxCount {
		crawler.scrapeBatch(*threadCount)
	}
	if crawler.logging {
		log.Printf("Found %d urls in %.3f seconds\n", crawler.count, time.Since(startTime).Seconds())
	}
	if *output == "" {
		fmt.Println(crawler.dump())
	} else {
		err := crawler.dumpToFile(*output)
		if err != nil {
			log.Fatal(err)
		}
	}
}
