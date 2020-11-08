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
	selectorURL    = `href=["'](https?:\/\/[a-zA-Z0-9\.\-]+[^"']*)["']`
	selectorDomain = `[a-zA-Z0-9\.\-]*\b([a-zA-Z0-9\-]+\.[a-zA-Z0-9\-]+)\b`
)

var (
	reURL    = regexp.MustCompile(selectorURL)
	reDomain = regexp.MustCompile(selectorDomain)
)

type Crawler struct {
	client           http.Client
	store            *addressStore
	wg               sync.WaitGroup
	restricted       bool
	restrictedDomain string
	logging          bool
}

func newCrawler(start string, timeout int64, queueSize int, restricted bool, logging bool) (*Crawler, error) {
	crawler := &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		store:            newAddressStore(queueSize),
		wg:               sync.WaitGroup{},
		restricted:       restricted,
		restrictedDomain: domain(start),
		logging:          logging,
	}
	crawler.store.add(start)
	return crawler, nil
}

func domain(address string) string {
	return reDomain.FindStringSubmatch(address)[1]
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
		if crawler.restricted && domain(address) != crawler.restrictedDomain {
			continue
		}
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
