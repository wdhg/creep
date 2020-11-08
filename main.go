package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	help := flag.Bool("h", false, "Show help")
	start := flag.String("s", "https://news.ycombinator.com", "The site to start crawling from")
	restrictToStart := flag.Bool("r", false, "Restricts crawling to specified start site")
	maxCount := flag.Int("n", 100, "Number of urls to scrape")
	threadCount := flag.Int("tc", 10, "Number of threads")
	timeout := flag.Int64("t", 5000, "Timeout for each http request (ms)")
	logging := flag.Bool("l", false, "Enables logging")
	output := flag.String("o", "", "Name of output file to write to. If not set will output to terminal")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	crawler, err := newCrawler(*start, *timeout, *threadCount, *restrictToStart, *logging)
	if err != nil {
		log.Fatalln(err)
	}
	startTime := time.Now()
	if crawler.logging {
		log.Printf("Crawling for %d urls...\n", *maxCount)
	}
	crawler.run(*maxCount, *threadCount)
	if crawler.logging {
		log.Printf("Found %d urls in %.3f seconds\n", crawler.store.count, time.Since(startTime).Seconds())
	}
	crawler.dump(*output)
}
