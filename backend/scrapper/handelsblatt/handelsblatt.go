package handelsblatt

import (
	"encoding/xml"
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/common"
	"time"
)

// RSS feed structures
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	GUID        string    `xml:"guid"`
	Categories  []string  `xml:"category"`
	Creator     string    `xml:"creator"`
	Enclosure   Enclosure `xml:"enclosure"`
}

type Enclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

// ScrapeHandelsblattRSS fetches and processes the Handelsblatt RSS feed
func Scrape() ([]model.Article, error) {
	var rss RSS
	if err := common.FetchRSSFeed("https://www.handelsblatt.com/contentexport/feed/schlagzeilen", &rss); err != nil {
		return nil, err
	}
	return parseRSStoArticles(rss)
}

func parseRSStoArticles(rss RSS) ([]model.Article, error) {
	var articles []model.Article
	seenGUIDs := make(map[string]bool) // Track GUIDs to avoid duplicates

	for _, item := range rss.Channel.Items {
		// Skip duplicate items based on GUID
		if seenGUIDs[item.GUID] {
			continue
		}
		seenGUIDs[item.GUID] = true

		// Parse publication date
		pubDate, err := parsePubDate(item.PubDate)
		if err != nil {
			continue
		}

		// Get banner image from enclosure
		banner := ""
		if item.Enclosure.Type == "image/jpeg" {
			banner = item.Enclosure.URL
		}

		// Clean up CDATA and extra whitespace
		title := common.CleanCDATA(item.Title)
		description := common.CleanCDATA(item.Description)

		// Skip items with no description
		if description == "" {
			continue
		}

		// Clean GUID for ID (use as-is since it's a URL)
		id := item.GUID

		// Handle categories (with bounds checking to prevent panic)
		categories := []string{}
		if len(item.Categories) > 0 {
			categories = []string{item.Categories[0]}
		}

		// Create Article
		article := model.Article{
			GormModel: model.GormModel{
				ID: fmt.Sprintf("%s-%s", model.SourceHandelsblatt, id),
			},
			Title:       title,
			Source:      model.SourceHandelsblatt, // Assuming SourceHandelsblatt is defined in model
			PublishedAt:   pubDate,
			URI:         item.Link,
			Views:       0, // Not provided in RSS feed
			Description: description,
			Banner:      banner,
			Category:    categories,
		}

		articles = append(articles, article)
	}

	return articles, nil
}

func parsePubDate(pubDate string) (time.Time, error) {
	// RSS pubDate format: Thu, 08 May 2025 19:51:16 +0200
	parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", pubDate)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
