package common

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// FetchRSSFeed fetches and parses an RSS feed from the given URL.
// It uses the SharedClient with timeout protection and validates HTTP response.
func FetchRSSFeed(url string, target interface{}) error {
	resp, err := SharedClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := xml.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to parse XML: %w", err)
	}

	return nil
}

// CleanCDATA removes CDATA tags and trims whitespace from a string.
// This is commonly needed when parsing RSS feeds that wrap content in CDATA sections.
func CleanCDATA(s string) string {
	s = strings.ReplaceAll(s, "<![CDATA[", "")
	s = strings.ReplaceAll(s, "]]>", "")
	return strings.TrimSpace(s)
}
