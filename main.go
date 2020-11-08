package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	help := flag.Bool("h", false, "Show help")
	logging := flag.Bool("l", false, "Enables logging")
	maxCount := flag.Int("n", 100, "Number of urls to scrape")
	start := flag.String("s", "https://news.ycombinator.com", "The site to start crawling from")
	hostRegex := flag.String("r", ".*", "Restricts crawling to specified hostname regex")
	output := flag.String("o", "", "Name of output file to write to. If not set will output to terminal")
	timeout := flag.Int64("t", 5000, "Timeout for each http request (ms)")
	threadCount := flag.Int("c", 10, "Number of threads")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	crawler, err := newCrawler(*start, *timeout, *threadCount, *hostRegex, *logging)
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
