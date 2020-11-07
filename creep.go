package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
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
	count    int
	maxCount int
	found    map[string]bool
	logging  bool
	lock     sync.Mutex
	wg       sync.WaitGroup
}

func newCrawler(timeout int64, maxCount int, logging bool) *Crawler {
	return &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		count:    0,
		maxCount: maxCount,
		found:    make(map[string]bool),
		logging:  logging,
		lock:     sync.Mutex{},
		wg:       sync.WaitGroup{},
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

func (crawler *Crawler) getNext() (string, bool) {
	crawler.lock.Lock()
	defer crawler.lock.Unlock()
	for address, hasBeenScraped := range crawler.found {
		if !hasBeenScraped {
			crawler.found[address] = true
			return address, true
		}
	}
	return "", false
}

func (crawler *Crawler) crawl() {
	defer crawler.wg.Done()
	for crawler.count < crawler.maxCount {
		address, ok := crawler.getNext()
		if ok {
			crawler.scrape(address)
		}
	}
}

func (crawler *Crawler) dumpAddresses() string {
	builder := strings.Builder{}
	for address := range crawler.found {
		builder.WriteString(address)
		builder.WriteString("\n")
	}
	return builder.String()
}

func (crawler *Crawler) dumpToTerminal() {
	fmt.Printf(crawler.dumpAddresses())
}

func (crawler *Crawler) dumpToFile(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.WriteString(crawler.dumpAddresses())
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}
