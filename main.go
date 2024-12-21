package main

import (
	"encoding/xml"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Category    string `xml:"category"`
	PubDate     string `xml:"pubDate"`
}

func main() {
	hpUrl := "https://www.upfc.jp/helloproject/news_list.php?@rst=all"
	resp, err := http.Get(hpUrl)
	if err != nil {
		log.Fatalf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Error: status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	var items []Item
	baseURL, err := url.Parse(hpUrl)
	if err != nil {
		log.Fatalf("Failed to parse base URL: %v", err)
	}

	doc.Find("div[data-category='ctg00'] ul.news_ul li").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		link, _ := s.Find("a").Attr("href")

		link = html.UnescapeString(link)

		parsedLink, err := url.Parse(link)
		if err != nil {
			log.Printf("Failed to parse link: %v", err)
			return
		}

		var absoluteURL string
		if parsedLink.IsAbs() {
			absoluteURL = link
		} else {
			absoluteURL = baseURL.ResolveReference(parsedLink).String()
		}

		dateTag := s.Find("a").Find(".news__date")

		print(fmt.Sprintf("dateTag: %v", dateTag.Text()))

		categoryTag := dateTag.Find(".news__ctg")
		category := categoryTag.Text()
		categoryTag.Remove()

		date := dateTag.Text()

		description := strings.TrimSpace(s.Find(".news__txt").Text())

		item := Item{
			Title:       title,
			Link:        absoluteURL,
			Category:    category,
			Description: description,
			PubDate:     date,
		}
		items = append(items, item)
	})

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "Hello! Project News",
			Link:        hpUrl,
			Description: "Latest news from Hello! Project",
			Items:       items,
		},
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		log.Fatalf("Failed to generate XML: %v", err)
	}

	fmt.Println(xml.Header + string(output))
}
