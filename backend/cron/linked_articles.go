package cron

import (
	"fmt"
	"news-swipe/backend/graph/model"
	"news-swipe/backend/scrapper/faz"
	"news-swipe/backend/scrapper/handelsblatt"
	"news-swipe/backend/scrapper/sueddeutsche"
	"news-swipe/backend/scrapper/tagesschau"
	"news-swipe/backend/scrapper/taz"
	"news-swipe/backend/scrapper/welt"
	"news-swipe/backend/scrapper/zeit"
	"news-swipe/backend/utils"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type scraperResult struct {
	source   string
	articles []model.Article
	err      error
}

type scraper struct {
	name string
	fn   func() ([]model.Article, error)
}

func FilterLinked(db *gorm.DB) error {
	startTime := time.Now()
	utils.CronJobRunsTotal.WithLabelValues("filter_linked").Inc()

	// Scrape articles from all sources concurrently
	articles, errors := scrapeAllSources()

	// Log results
	if len(articles) == 0 && len(errors) > 0 {
		utils.CronJobErrorsTotal.WithLabelValues("filter_linked").Inc()
		return fmt.Errorf("all scrapers failed: %v", errors)
	}

	utils.Log(utils.Database, "Scraped articles", "count", len(articles), "errors", len(errors))

	// Detect languages
	detectLanguages(articles)

	// Save to database
	err := saveToDatabase(db, articles)

	// Record metrics
	utils.CronJobDuration.WithLabelValues("filter_linked").Observe(time.Since(startTime).Seconds())
	if err != nil {
		utils.CronJobErrorsTotal.WithLabelValues("filter_linked").Inc()
	}

	return err
}

func scrapeAllSources() ([]model.Article, []error) {
	scrapers := []scraper{
		{"Zeit", zeit.Scrape},
		{"FAZ", faz.Scrape},
		{"Tagesschau", tagesschau.Scrape},
		{"Süddeutsche", sueddeutsche.Scrape},
		{"Welt", welt.Scrape},
		{"Handelsblatt", handelsblatt.Scrape},
		{"TAZ", taz.Scrape},
	}

	results := make(chan scraperResult, len(scrapers))
	var wg sync.WaitGroup

	// Launch all scrapers concurrently
	for _, s := range scrapers {
		wg.Add(1)
		go runScraper(&wg, results, s.name, s.fn)
	}

	// Close results channel when all scrapers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	return collectResults(results)
}

func runScraper(wg *sync.WaitGroup, results chan<- scraperResult, name string, scrapeFn func() ([]model.Article, error)) {
	defer wg.Done()

	startTime := time.Now()
	utils.ScraperRequestsTotal.WithLabelValues(name).Inc()

	articles, err := scrapeFn()

	// Record metrics
	utils.ScraperDuration.WithLabelValues(name).Observe(time.Since(startTime).Seconds())
	if err != nil {
		utils.ScraperErrorsTotal.WithLabelValues(name).Inc()
	} else {
		utils.ArticlesScraped.WithLabelValues(name).Add(float64(len(articles)))
	}

	results <- scraperResult{
		source:   name,
		articles: articles,
		err:      err,
	}
}

func collectResults(results <-chan scraperResult) ([]model.Article, []error) {
	var allArticles []model.Article
	var errors []error

	for result := range results {
		if result.err != nil {
			utils.Log(utils.Scraper, result.source+" scraping failed", "error", result.err)
			errors = append(errors, fmt.Errorf("%s: %w", result.source, result.err))
		} else {
			allArticles = append(allArticles, result.articles...)
			utils.Log(utils.Scraper, result.source+" scraping succeeded", "count", len(result.articles))
		}
	}

	return allArticles, errors
}

func detectLanguages(articles []model.Article) {
	for i := range articles {
		articles[i].Language = model.FromLingua(
			utils.DetectArticleLanguage(articles[i].Title, articles[i].Description),
		)
	}
}

func saveToDatabase(db *gorm.DB, articles []model.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// Load existing articles
	var existing []model.Article
	if err := db.Find(&existing).Error; err != nil {
		return err
	}

	// Link similar articles
	linkSimilarArticles(articles, existing)

	// Collect all unique articles to upsert
	articlesMap := collectUniqueArticles(articles)

	// Upsert articles in batches
	if err := upsertArticles(db, articlesMap); err != nil {
		return err
	}

	// Persist associations
	return persistAssociations(db, articles)
}

func linkSimilarArticles(newArticles []model.Article, existingArticles []model.Article) {
	config := utils.DefaultSimilarityConfig()
	threshold := 0.3

	// Link new articles to existing articles
	for i := range newArticles {
		for j := range existingArticles {
			if utils.IsSimilar(newArticles[i], existingArticles[j], threshold, config) && newArticles[i].ID != existingArticles[j].ID {
				newArticles[i].LinkedTo = append(newArticles[i].LinkedTo, &existingArticles[j])
			}
		}

		// Link new articles to each other
		for j := i + 1; j < len(newArticles); j++ {
			if utils.IsSimilar(newArticles[i], newArticles[j], threshold, config) && newArticles[i].ID != newArticles[j].ID {
				newArticles[i].LinkedTo = append(newArticles[i].LinkedTo, &newArticles[j])
			}
		}
	}
}

func collectUniqueArticles(articles []model.Article) map[string]*model.Article {
	articlesMap := make(map[string]*model.Article)

	// Add all new articles
	for i := range articles {
		a := &articles[i]
		if a.ID == "" {
			continue
		}
		articlesMap[a.ID] = a

		// Include linked articles
		for _, linked := range a.LinkedTo {
			if linked.ID != "" {
				articlesMap[linked.ID] = linked
			}
		}
	}

	return articlesMap
}

func upsertArticles(db *gorm.DB, articlesMap map[string]*model.Article) error {
	if len(articlesMap) == 0 {
		return nil
	}

	articles := make([]*model.Article, 0, len(articlesMap))
	for _, a := range articlesMap {
		articles = append(articles, a)
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).CreateInBatches(articles, 100).Error
}

func persistAssociations(db *gorm.DB, articles []model.Article) error {
	for i := range articles {
		if len(articles[i].LinkedTo) == 0 {
			continue
		}

		// Filter valid links
		validLinks := make([]*model.Article, 0, len(articles[i].LinkedTo))
		for _, link := range articles[i].LinkedTo {
			if link.ID != "" {
				validLinks = append(validLinks, link)
			}
		}

		if len(validLinks) > 0 {
			if err := db.Model(&articles[i]).Association("LinkedTo").Append(validLinks); err != nil {
				return err
			}
		}
	}
	return nil
}
