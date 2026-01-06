package welt

import (
	"encoding/xml"
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/common"
	"time"

	"github.com/pemistahl/lingua-go"
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
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	GUID        string   `xml:"guid"`
	Categories  []string `xml:"category"`
	Creator     string   `xml:"creator"`
	Premium     string   `xml:"premium"`
	Topic       string   `xml:"topic"`
	Media       []Media  `xml:"content"`
	Keywords    string   `xml:"keywords"`
}

type Media struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Credit string `xml:"credit"`
}

// ScrapeWeltRSS fetches and processes the WELT.de RSS feed
func Scrape() ([]model.Article, error) {
	var rss RSS
	if err := common.FetchRSSFeed("https://www.welt.de/feeds/topnews.rss", &rss); err != nil {
		return nil, err
	}
	return parseRSStoArticles(rss)
}

func parseRSStoArticles(rss RSS) ([]model.Article, error) {
	var articles []model.Article
	seenGUIDs := make(map[string]bool) // Track GUIDs to avoid duplicates

	for _, item := range rss.Channel.Items {

		if item.Premium == "true" {
			continue
		}

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

		// Get banner image from media:content
		banner := ""
		for _, media := range item.Media {
			if media.Type == "image/jpeg" {
				banner = media.URL
				break
			}
		}

		// Clean up CDATA and extra whitespace
		title := common.CleanCDATA(item.Title)
		description := common.CleanCDATA(item.Description)

		// Skip items with no description
		if description == "" {
			continue
		}

		// Clean GUID for ID (use as-is since it's a numeric ID)
		id := item.GUID

		// Combine categories and topic
		categories := make([]string, 0, len(item.Categories)+1)
		categories = append(categories, item.Categories...)
		if item.Topic != "" {
			categories = append(categories, item.Topic)
		}

		// Create Article
		article := model.Article{
			GormModel: model.GormModel{
				ID: fmt.Sprintf("%s-%s", model.SourceWelt, id),
			},
			Title:       title,
			Source:      model.SourceWelt, // Assuming SourceWelt is defined in model
			PublishedAt: pubDate,
			URI:         item.Link,
			Views:       0, // Not provided in RSS feed
			Description: description,
			Banner:      banner,
			Category:    categories,
			Language:    model.FromLingua(lingua.German),
		}

		articles = append(articles, article)
	}

	return articles, nil
}

func parsePubDate(pubDate string) (time.Time, error) {
	// RSS pubDate format: Thu, 08 May 2025 17:24:49 GMT
	parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 MST", pubDate)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
