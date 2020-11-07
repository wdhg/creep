package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

const (
	selectorURL = `href=["'](https?:\/\/[a-zA-Z0-9\.\-]+[^"']*)["']`
)

var (
	reURL = regexp.MustCompile(selectorURL)
)

type Crawler struct {
	client  http.Client
	store   *addressStore
	logging bool
	lock    sync.Mutex
	wg      sync.WaitGroup
}

func newCrawler(timeout int64, queueSize int, logging bool) *Crawler {
	return &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		store:   newAddressStore(queueSize),
		logging: logging,
		lock:    sync.Mutex{},
		wg:      sync.WaitGroup{},
	}
}

// sanitiseAddress removes the query, fragment, and any trailing forward slashes
func sanitiseAddress(address string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	u.RawFragment = ""
	for u.Path != "" && u.Path[len(u.Path)-1] == '/' {
		u.Path = u.Path[0 : len(u.Path)-1]
	}
	return u.String(), nil
}

func findAddresses(html string) []string {
	addresses := []string{}
	for _, submatch := range reURL.FindAllStringSubmatch(html, -1) {
		address, err := sanitiseAddress(submatch[1])
		if err != nil {
			continue
		}
		addresses = append(addresses, address)
	}
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
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	crawler.storeAddresses(findAddresses(string(data)))
}

func (crawler *Crawler) crawl(maxCount int) {
	defer crawler.wg.Done()
	for crawler.store.count < maxCount {
		crawler.scrapeNext()
	}
}

func (crawler *Crawler) run(maxCount int, threadCount int) {
	crawler.wg.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go crawler.crawl(maxCount)
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
