package faz

import (
	"encoding/xml"
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/common"
	"regexp"
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
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	PubDate     string   `xml:"pubDate"`
	GUID        string   `xml:"guid"`
	Categories  []string `xml:"category"`
	Creator     string   `xml:"dc:creator"`
	Media       []Media  `xml:"content"`
}

type Media struct {
	URL    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Medium string `xml:"medium,attr"`
	Height string `xml:"height,attr"`
	Width  string `xml:"width,attr"`
}

// ScrapeFAZRSS fetches and processes the FAZ.NET RSS feed
func Scrape() ([]model.Article, error) {
	var rss RSS
	if err := common.FetchRSSFeed("https://www.faz.net/rss/aktuell/", &rss); err != nil {
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
		for _, media := range item.Media {
			if media.Type == "image/jpeg" && media.Medium == "image" {
				banner = media.URL
				break
			}
		}

		// Clean up CDATA and extract description
		title := common.CleanCDATA(item.Title)
		description := common.CleanCDATA(item.Description)

		// Remove HTML tags and image from description
		re := regexp.MustCompile(`<p><img[^>]+></p>`)
		description = re.ReplaceAllString(description, "")
		re = regexp.MustCompile(`<p>(.*?)</p>`)
		matches := re.FindStringSubmatch(description)
		if len(matches) > 1 {
			description = strings.TrimSpace(matches[1])
		}

		// Skip items with no description
		if description == "" {
			continue
		}

		// Convert categories to []string
		categories := make([]string, len(item.Categories))
		for i, cat := range item.Categories {
			catCopy := cat
			categories[i] = catCopy
		}

		// Extract ID from GUID (assuming it's the article ID in the URL)
		idParts := strings.Split(item.GUID, "-")
		var id string
		if len(idParts) > 0 {
			id = idParts[len(idParts)-1]
			id = strings.TrimSuffix(id, ".html")
		} else {
			id = item.GUID // Fallback to full GUID
		}

		// Create Article
		article := model.Article{
			GormModel: model.GormModel{
				ID: fmt.Sprintf("%s-%s", model.SourceFaz, id),
			},
			Title:       title,
			Source:      model.SourceFaz, // Assuming SourceFAZ is defined in model
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
	// RSS pubDate format: Wed, 07 May 2025 17:46:21 +0200
	parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", pubDate)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}
