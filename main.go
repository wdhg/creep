package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	linksQuery  = "a, link"
	href        = "href"
	urlSelector = `[a-zA-Z]+:\/\/[a-zA-Z\.-]+(\/[\S)]*)?`
)

var (
	help        = flag.Bool("h", false, "Show help")
	start       = flag.String("s", "https://news.ycombinator.com", "The site to start crawling from")
	maxCount    = flag.Int("n", 1000, "Number of urls to scrape")
	threadCount = flag.Int("tc", 10, "Number of threads")
	timeout     = flag.Int64("t", 5000, "Timeout for each http request (ms)")
	logging     = flag.Bool("l", true, "Enable / disable logging")
)

type Crawler struct {
	client  http.Client
	count   int
	found   map[string]bool
	logging bool
	lock    sync.Mutex
	wg      sync.WaitGroup
}

func newCrawler(timeout int64, logging bool) *Crawler {
	return &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		count:   0,
		found:   make(map[string]bool),
		logging: logging,
		lock:    sync.Mutex{},
		wg:      sync.WaitGroup{},
	}
}

func isURL(address string) bool {
	matches, err := regexp.MatchString(urlSelector, address)
	return err == nil && matches
}

func findAddresses(doc *goquery.Document) []string {
	addresses := []string{}
	doc.Find(linksQuery).Each(func(_ int, s *goquery.Selection) {
		address, exists := s.Attr(href)
		if !exists {
			return
		}
		if isURL(address) {
			addresses = append(addresses, address)
		} else {
			addresses = append(addresses, fmt.Sprintf("%s/%s", doc.Url.String(), address))
		}
	})
	return addresses
}

func getHostname(address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	return u.Hostname(), nil
}

func (crawler *Crawler) storeAddress(address string) {
	crawler.lock.Lock()
	defer crawler.lock.Unlock()
	if _, ok := crawler.found[address]; !ok {
		crawler.found[address] = false
		crawler.count++
	}
}

func (crawler *Crawler) storeAddresses(addresses []string) {
	for _, address := range addresses {
		crawler.storeAddress(address)
	}
}

func (crawler *Crawler) scrape(address string) {
	defer crawler.wg.Done()
	crawler.lock.Lock()
	crawler.found[address] = true
	crawler.lock.Unlock()
	if crawler.logging {
		log.Printf("Scraping %s...\n", address)
	}
	res, err := crawler.client.Get(address)
	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}
	crawler.storeAddresses(findAddresses(doc))
}

func (crawler *Crawler) scrapeNext() bool {
	crawler.lock.Lock()
	defer crawler.lock.Unlock()
	for address, hasBeenScraped := range crawler.found {
		if !hasBeenScraped {
			crawler.found[address] = true
			go crawler.scrape(address)
			return true
		}
	}
	return false
}

func (crawler *Crawler) scrapeBatch(threads int) {
	if crawler.logging {
		log.Println("Running batch")
		defer log.Println("Batch done")
	}
	crawler.wg.Add(threads)
	for i := 0; i < threads; i++ {
		ok := crawler.scrapeNext()
		if !ok {
			crawler.wg.Done()
		}
	}
	crawler.wg.Wait()
}

func main() {
	flag.Parse()
	if *help {
		flag.Usage()
		return
	}
	crawler := newCrawler(*timeout, *logging)
	crawler.storeAddress(*start)
	startTime := time.Now()
	for crawler.count < *maxCount {
		crawler.scrapeBatch(*threadCount)
	}
	if crawler.logging {
		log.Printf("Found %d urls in %.3f seconds\n", crawler.count, time.Since(startTime).Seconds())
	}
}
