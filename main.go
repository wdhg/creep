package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

func scrapeURLs(address string) ([]string, error) {
	res, err := http.Get(address)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	links := []string{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		matches, err := regexp.MatchString(`https:\/\/[\S]*`, link)
		if err != nil {
			log.Fatal(err)
		}
		if matches {
			links = append(links, link)
		} else {
			links = append(links, fmt.Sprintf("%s/%s", address, link))
		}
	})
	return links, nil
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
