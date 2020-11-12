package main

import (
	"flag"
	"log"
	"time"
)

func main() {
	threadCount := flag.Int("threads", 10, "Number of threads")
	help := flag.Bool("help", false, "Show help")
	logging := flag.Bool("log", false, "Enables logging")
	maxCount := flag.Int("max", 100, "Number of urls to scrape")
	output := flag.String("output", "", "Name of output file to write to. If not set will output to terminal")
	hostRegex := flag.String("hostname", ".*", "Restricts crawling to specified hostname regex")
	start := flag.String("start", "https://news.ycombinator.com", "The site to start crawling from")
	timeout := flag.Int64("timeout", 5000, "Timeout for each http request (ms)")
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	crawler, err := newCrawler(*start, *timeout, *maxCount, *hostRegex, *logging)
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
	err = crawler.dump(*output)
	if err != nil {
		log.Fatal(err)
	}
}
