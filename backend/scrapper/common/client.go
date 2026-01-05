package common

import (
	"net/http"
	"time"
)

// SharedClient is a configured HTTP client with appropriate timeouts
// to prevent indefinite hangs when scraping RSS feeds.
var SharedClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	},
}
