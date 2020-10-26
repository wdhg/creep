package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

const (
	linksQuery  = "a, link"
	href        = "href"
	urlSelector = `[a-zA-Z]+:\/\/[a-zA-Z\.-]+(\/[\S)]*)?`
)

func isURL(address string) bool {
	matches, err := regexp.MatchString(urlSelector, address)
	return err == nil && matches
}

func findURLs(doc *goquery.Document) []string {
	links := []string{}
	doc.Find(linksQuery).Each(func(_ int, s *goquery.Selection) {
		link, exists := s.Attr(href)
		if !exists {
			return
		}
		if isURL(link) {
			links = append(links, link)
		} else {
			links = append(links, fmt.Sprintf("%s/%s", doc.Url.String(), link))
		}
	})
	return links
}

func scrapeURLs(address string) ([]string, error) {
	res, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	return findURLs(doc), nil
}

func main() {
	links, err := scrapeURLs("https://news.ycombinator.com")
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range links {
		fmt.Println(v)
	}
}
