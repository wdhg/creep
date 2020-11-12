package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sync"
	"time"
)

const (
	selectorURL = `href=["'](https?:\/\/[a-zA-Z0-9\.\-]+[^"']*)["']`
)

type Crawler struct {
	client     http.Client
	store      *addressStore
	wg         sync.WaitGroup
	reURL      *regexp.Regexp
	reHostname *regexp.Regexp
	logging    bool
}

// newCrawler creates a Crawler with a new addressStore and pre-compiled regexps
func newCrawler(start string, timeout int64, maxCount int, selectorHost string, logging bool) (*Crawler, error) {
	reURL, err := regexp.Compile(selectorURL)
	if err != nil {
		return nil, err
	}
	reHostname, err := regexp.Compile(selectorHost)
	if err != nil {
		return nil, err
	}
	crawler := &Crawler{
		client: http.Client{
			Timeout: time.Duration(timeout) * time.Millisecond,
		},
		store:      newAddressStore(maxCount),
		wg:         sync.WaitGroup{},
		reURL:      reURL,
		reHostname: reHostname,
		logging:    logging,
	}
	crawler.store.add(start)
	return crawler, nil
}

// dump dumps all the found addresses either to the terminal or to a file
func (crawler *Crawler) dump(filename string) error {
	file, err := getDumpFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = crawler.store.dumpTo(file)
	return err
}

// run causes the Crawler to run with `threadCount` extra threads until
// `maxCount` addresses have been found
func (crawler *Crawler) run(maxCount int, threadCount int) {
	crawler.wg.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go crawler.crawl(maxCount)
	}
	crawler.wg.Wait()
}

// crawl constantly GETs pages, scrapes any addresses in them, and adds them to
// the `addressStore` until `maxCount` addresses are found
func (crawler *Crawler) crawl(maxCount int) {
	full := false
	for !full {
		full = crawler.scrape()
	}
	crawler.wg.Done()
}

// scrapeNext gets an unvisited address, GETs it, and scrapes any addresses from
// its content
func (crawler *Crawler) scrape() bool {
	address := crawler.store.next()
	if crawler.logging {
		log.Printf("Scraping %s...\n", address)
	}
	res, err := crawler.client.Get(address)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}
	return crawler.storeAddresses(string(data))
}

// storeAddresses scrapes all url addresses from the content using regex and
// stores them in store
func (crawler *Crawler) storeAddresses(content string) bool {
	for _, submatch := range crawler.reURL.FindAllStringSubmatch(content, -1) {
		address, err := sanitiseAddress(submatch[1])
		if err != nil {
			continue
		}
		u, err := url.Parse(address)
		if !crawler.reHostname.MatchString(u.Hostname()) {
			continue
		}
		full := crawler.store.add(address)
		if full {
			return true
		}
	}
	return false
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

func getDumpFile(filename string) (*os.File, error) {
	if filename == "" {
		return os.Stdout, nil
	}
	return os.Create(filename)
}
