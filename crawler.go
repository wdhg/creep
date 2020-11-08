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

type Crawler struct {
	client  http.Client
	store   *addressStore
	wg      sync.WaitGroup
	reURL   *regexp.Regexp
	reHost  *regexp.Regexp
	logging bool
}

func newCrawler(start string, timeout int64, queueSize int, selectorHost string, logging bool) (*Crawler, error) {
	crawler := &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		store:   newAddressStore(queueSize),
		wg:      sync.WaitGroup{},
		reURL:   regexp.MustCompile(selectorURL),
		reHost:  regexp.MustCompile(selectorHost),
		logging: logging,
	}
	crawler.store.add(start)
	return crawler, nil
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

func (crawler *Crawler) findAddresses(html string) []string {
	addresses := []string{}
	for _, submatch := range crawler.reURL.FindAllStringSubmatch(html, -1) {
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
		u, err := url.Parse(address)
		if err != nil {
			continue
		}
		if crawler.reHost.MatchString(u.Hostname()) {
			crawler.store.add(address)
		}
	}
}

func (crawler *Crawler) scrapeNext() {
	address, ok := crawler.store.next()
	if !ok {
		return
	}
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
	crawler.storeAddresses(crawler.findAddresses(string(data)))
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
