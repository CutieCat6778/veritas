package utils

import (
	"fmt"
	"log"
	"time"
)

type Source string

const (
	GraphQL  Source = "GraphQL"
	Cron     Source = "CronJob"
	Main     Source = "Main"
	System   Source = "System"
	Database Source = "Database"
	Cache    Source = "Cache"
	Scraper  Source = "Scraper"
	Server   Source = "Server"
)

func Log(source Source, content ...any) {
	if len(content) == 0 {
		log.Printf("%s: [%s]", time.Now(), source)
		return
	}

	// Format as key-value pairs
	msg := ""
	for i := 0; i < len(content); i += 2 {
		if i > 0 {
			msg += ", "
		}
		if i+1 < len(content) {
			msg += fmt.Sprintf("%v=%v", content[i], content[i+1])
		} else {
			msg += fmt.Sprintf("%v", content[i])
		}
	}

	log.Printf("%s: [%s] %s", time.Now(), source, msg)
}
