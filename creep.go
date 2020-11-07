package main

import (
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

type Crawler struct {
	client   http.Client
	maxCount int
	store    *addressStore
	logging  bool
	lock     sync.Mutex
	wg       sync.WaitGroup
}

func newCrawler(timeout int64, maxCount int, threadCount int, logging bool) *Crawler {
	return &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		maxCount: maxCount,
		store:    newAddressStore(threadCount),
		logging:  logging,
		lock:     sync.Mutex{},
		wg:       sync.WaitGroup{},
	}
}

func isURL(address string) bool {
	matches, err := regexp.MatchString(urlSelector, address)
	return err == nil && matches
}

// sanitiseAddress removes the query and any trailing forward slashes
func sanitiseAddress(address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	for u.Path != "" && u.Path[len(u.Path)-1] == '/' {
		u.Path = u.Path[0 : len(u.Path)-1]
	}
	return u.String(), nil
}

func findAddresses(doc *goquery.Document) []string {
	addresses := []string{}
	doc.Find(linksQuery).Each(func(_ int, s *goquery.Selection) {
		rawAddress, exists := s.Attr(href)
		if !exists {
			return
		}
		if isURL(rawAddress) {
			address, err := sanitiseAddress(rawAddress)
			if err != nil {
				return
			}
			addresses = append(addresses, address)
		}
	})
	return addresses
}

func (crawler *Crawler) storeAddresses(addresses []string) {
	for _, address := range addresses {
		crawler.store.add(address)
	}
}

func (crawler *Crawler) scrapeNext() {
	address := crawler.store.next()
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

func (crawler *Crawler) crawl() {
	defer crawler.wg.Done()
	for crawler.store.count < crawler.maxCount {
		crawler.scrapeNext()
	}
}

func (crawler *Crawler) run() {
	crawler.wg.Add(*threadCount)
	for i := 0; i < *threadCount; i++ {
		go crawler.crawl()
	}
	crawler.wg.Wait()
}

func (crawler *Crawler) dump(output string) error {
	if output == "" {
		crawler.store.dumpToTerminal()
	} else {
		err := crawler.store.dumpToFile(output)
		if err != nil {
			return err
		}
	}
	return nil
}
