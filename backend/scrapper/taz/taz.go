package taz

import (
	"encoding/xml"
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/common"
	"strings"
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
	Title        string       `xml:"title"`
	Link         []string     `xml:"link"`
	Description  string       `xml:"description"`
	PubDate      string       `xml:"pubDate"`
	GUID         string       `xml:"guid"`
	MediaContent MediaContent `xml:"content"`
}

type MediaContent struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Medium string `xml:"medium,attr"`
}

// ScrapeTazRSS fetches and processes the TAZ.de RSS feed
func Scrape() ([]model.Article, error) {
	var rss RSS
	if err := common.FetchRSSFeed("https://taz.de/!p4608;rss/", &rss); err != nil {
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

		// Get banner image from media:content
		banner := ""
		if item.MediaContent.Type == "image/jpeg" && item.MediaContent.Medium == "image" {
			banner = item.MediaContent.URL
		}

		// Clean up CDATA and extra whitespace
		title := common.CleanCDATA(item.Title)
		description := common.CleanCDATA(item.Description)

		// Remove trailing <a href...> link from description
		if idx := strings.Index(description, "<a href"); idx != -1 {
			description = strings.TrimSpace(description[:idx])
		}

		// Skip items with no description
		if description == "" {
			continue
		}

		// Clean GUID for ID (use as-is since it's a URL)
		id := item.GUID

		// TAZ.de RSS does not provide categories or authors
		categories := []string{}

		// Get URI with bounds checking to prevent panic
		uri := ""
		if len(item.Link) > 0 {
			uri = item.Link[0]
		}

		// Create Article
		article := model.Article{
			GormModel: model.GormModel{
				ID: fmt.Sprintf("%s-%s", model.SourceTaz, id),
			},
			Title:       title,
			Source:      model.SourceTaz, // Assuming SourceTaz is defined in model
			PublishedAt: pubDate,
			URI:         uri,
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
	// RSS pubDate format: 8 May 2025 20:19:00 +0200
	parsedTime, err := time.Parse("2 Jan 2006 15:04:05 -0700", pubDate)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
