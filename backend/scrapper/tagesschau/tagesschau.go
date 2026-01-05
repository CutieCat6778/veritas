package tagesschau

import (
	"encoding/xml"
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/common"
	"regexp"
	"strings"
	"time"
)

// RDF represents the root of the RDF/XML feed.
type RDF struct {
	XMLName xml.Name `xml:"RDF"`
	Items   []Item   `xml:"item"`
}

// Item represents an <item> element in the RDF feed.
type Item struct {
	About          string `xml:"about,attr"`
	Title          string `xml:"title"`
	Link           string `xml:"link"`
	Description    string `xml:"description"`
	PubDate        string `xml:"pubDate"`
	GUID           string `xml:"guid"`
	DCDate         string `xml:"date"`
	ContentEncoded string `xml:"encoded"`
	DCIdentifier   string `xml:"identifier"`
}

// scrapeTagesschau fetches and parses the Tagesschau RDF/XML feed.
func Scrape() ([]model.Article, error) {
	// Fetch and parse the XML feed
	var rdf RDF
	err := common.FetchRSSFeed("https://www.tagesschau.de/infoservices/alle-meldungen-100~rdf.xml", &rdf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Convert items to articles
	articles := make([]model.Article, 0, len(rdf.Items))
	for _, item := range rdf.Items {
		// Parse the creation date
		pubDate, err := parsePubDate(item.PubDate)
		if err != nil {
			continue
		}

		// Extract banner image URL from content:encoded
		banner := extractImageURL(item.ContentEncoded)

		article := model.Article{
			GormModel: model.GormModel{
				ID: fmt.Sprintf("%s-%s", model.SourceTagesschau, item.GUID),
			},
			Title:       item.Title,
			Source:      model.SourceTagesschau,
			PublishedAt: pubDate,
			URI:         item.Link,
			Views:       0, // Not available in XML
			Description: item.Description,
			Banner:      banner,
			Category:    []string{},
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// extractImageURL extracts the src attribute of the first <img> tag in the HTML content.
func extractImageURL(content string) string {
	// Simple regex to match <img src="...">
	re := regexp.MustCompile(`<img[^>]+src=["'](.*?)["']`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func parsePubDate(pubDate string) (time.Time, error) {
	// RSS pubDate format: Mon, 05 May 2025 17:19:18 CEST
	// Convert CEST to a known format
	pubDate = strings.Replace(pubDate, "CEST", "+0200", 1)

	// Parse the date
	parsedTime, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", pubDate)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime, nil
}
