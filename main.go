package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	threadCount := flag.Int("c", 10, "Number of threads")
	help := flag.Bool("h", false, "Show help")
	logging := flag.Bool("l", false, "Enables logging")
	maxCount := flag.Int("n", 100, "Number of urls to scrape")
	output := flag.String("o", "", "Name of output file to write to. If not set will output to terminal")
	hostRegex := flag.String("r", ".*", "Restricts crawling to specified hostname regex")
	start := flag.String("s", "https://news.ycombinator.com", "The site to start crawling from")
	timeout := flag.Int64("t", 5000, "Timeout for each http request (ms)")
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
	crawler.run(*start, *maxCount, *threadCount)
	if crawler.logging {
		log.Printf("Found %d urls in %.3f seconds\n", crawler.store.count, time.Since(startTime).Seconds())
	}
	err = crawler.dump(*output)
	if err != nil {
		log.Fatal(err)
	}
}
