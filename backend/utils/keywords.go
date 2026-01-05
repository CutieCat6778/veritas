package utils

import (
	"maps"
	"sort"
	"strings"
	"time"
	"unicode"

	"news-swipe/backend/graph/model"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GenerateKeywordsFromArticles(db *gorm.DB) error {
	cutoff := time.Now().AddDate(0, 0, -14)

	var articles []model.Article
	if err := db.Where("published_at >= ?", cutoff).Find(&articles).Error; err != nil {
		return err
	}
	if len(articles) == 0 {
		return nil
	}

	// Deduplicate by title (case-insensitive)
	articles = deduplicateByTitle(articles)

	// Cluster similar articles
	config := DefaultSimilarityConfig()
	clusters := clusterArticles(articles, 0.37, config)

	// Extract and clean keywords
	keywords := extractKeywordsFromClusters(clusters)
	if len(keywords) == 0 {
		return nil
	}

	// Merge similar/redundant keywords
	keywords = mergeRedundantKeywords(keywords)

	// Persist to database
	return persistKeywords(db, keywords, clusters)
}

func deduplicateByTitle(articles []model.Article) []model.Article {
	seen := make(map[string]bool)
	result := make([]model.Article, 0, len(articles))

	for _, a := range articles {
		titleKey := strings.ToLower(strings.TrimSpace(a.Title))
		if !seen[titleKey] {
			seen[titleKey] = true
			result = append(result, a)
		}
	}
	return result
}

func clusterArticles(articles []model.Article, threshold float64, config SimilarityConfig) [][]model.Article {
	var clusters [][]model.Article
	visited := make(map[string]bool)

	for i, a1 := range articles {
		if visited[a1.ID] {
			continue
		}

		cluster := []model.Article{a1}
		visited[a1.ID] = true

		for j := i + 1; j < len(articles); j++ {
			a2 := articles[j]
			if !visited[a2.ID] && IsSimilar(a1, a2, threshold, config) {
				cluster = append(cluster, a2)
				visited[a2.ID] = true
			}
		}

		if len(cluster) > 1 {
			clusters = append(clusters, cluster)
		}
	}
	return clusters
}

func extractKeywordsFromClusters(clusters [][]model.Article) map[string]*model.KeyWords {
	keywords := make(map[string]*model.KeyWords)
	titleFormatter := cases.Title(language.German)
	stopwords := getCombinedStopwords()

	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}

		candidates := extractCandidates(cluster, stopwords)
		topKeywords := selectTopKeywords(candidates, 5) // Get more initially

		for _, kw := range topKeywords {
			cleaned := cleanKeyword(kw)
			if len(cleaned) < 4 {
				continue
			}

			formatted := titleFormatter.String(cleaned)
			if _, exists := keywords[formatted]; !exists {
				keywords[formatted] = &model.KeyWords{
					GormModel:  model.GormModel{ID: uuid.NewString()},
					Keyword:    formatted,
					LastUpdate: time.Now(),
				}
			}
		}
	}
	return keywords
}

func mergeRedundantKeywords(keywords map[string]*model.KeyWords) map[string]*model.KeyWords {
	// Convert to slice for processing
	type kwItem struct {
		key string
		kw  *model.KeyWords
	}

	items := make([]kwItem, 0, len(keywords))
	for k, v := range keywords {
		items = append(items, kwItem{k, v})
	}

	// Sort by length (shortest first) - we keep the shortest/simplest form
	sort.Slice(items, func(i, j int) bool {
		return len(items[i].key) < len(items[j].key)
	})

	merged := make(map[string]*model.KeyWords)
	toSkip := make(map[string]bool)

	for i, item := range items {
		if toSkip[item.key] {
			continue
		}

		// Check if this keyword is redundant with any shorter keyword we've already kept
		isRedundant := false
		for existingKey := range merged {
			if isKeywordRedundant(item.key, existingKey) {
				isRedundant = true
				break
			}
		}

		if !isRedundant {
			merged[item.key] = item.kw

			// Mark longer variations as redundant
			for j := i + 1; j < len(items); j++ {
				if isKeywordRedundant(items[j].key, item.key) {
					toSkip[items[j].key] = true
				}
			}
		}
	}

	return merged
}

func isKeywordRedundant(longer, shorter string) bool {
	longerLower := strings.ToLower(longer)
	shorterLower := strings.ToLower(shorter)

	// Same keyword
	if longerLower == shorterLower {
		return true
	}

	// Extract core words (ignore stopwords and common prepositions)
	longerWords := extractCoreWords(longerLower)
	shorterWords := extractCoreWords(shorterLower)

	// If shorter is empty after filtering, not redundant
	if len(shorterWords) == 0 {
		return false
	}

	// Check if all core words from shorter appear in longer
	matchCount := 0
	for _, sw := range shorterWords {
		for _, lw := range longerWords {
			if sw == lw {
				matchCount++
				break
			}
		}
	}

	// If all core words from shorter are in longer, it's redundant
	// Example: "Trump" vs "Donald Trump" -> redundant
	// Example: "Trump" vs "On Trump" -> redundant
	return matchCount == len(shorterWords)
}

func extractCoreWords(text string) []string {
	// Remove common connecting words/prepositions
	connectingWords := map[string]bool{
		"on": true, "in": true, "at": true, "to": true, "for": true,
		"of": true, "and": true, "or": true, "the": true, "a": true,
		"is": true, "was": true, "are": true, "were": true, "be": true,
		"an": true, "as": true, "by": true, "with": true, "from": true,
		"über": true, "von": true, "im": true, "am": true, "der": true,
		"die": true, "das": true, "den": true, "dem": true, "des": true,
		"ein": true, "eine": true, "einer": true, "einem": true,
		"ist": true, "sind": true, "war": true, "waren": true,
	}

	words := strings.Fields(text)
	core := make([]string, 0, len(words))

	for _, w := range words {
		w = cleanWord(w)
		if len(w) > 2 && !connectingWords[w] {
			core = append(core, w)
		}
	}

	return core
}

func extractCandidates(cluster []model.Article, stopwords map[string]bool) map[string]int {
	candidates := make(map[string]int)

	for _, article := range cluster {
		text := article.Title + " " + article.Description
		words := tokenize(text)

		// Score unigrams and bigrams
		for i := 0; i < len(words); i++ {
			w := words[i]
			if len(w) > 3 && !stopwords[w] && !isCommonEntity(w) {
				candidates[w] += 3
			}

			// Bigrams
			if i < len(words)-1 {
				w2 := words[i+1]
				if len(w2) > 3 && !stopwords[w2] {
					bigram := w + " " + w2
					candidates[bigram] += 5
				}
			}
		}
	}
	return candidates
}

func selectTopKeywords(candidates map[string]int, max int) []string {
	// Sort by score descending
	type scoredKw struct {
		keyword string
		score   int
	}

	list := make([]scoredKw, 0, len(candidates))
	for k, v := range candidates {
		list = append(list, scoredKw{k, v})
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].score > list[j].score
	})

	// Select top keywords (less strict filtering here since we'll merge later)
	selected := make([]string, 0, max)
	for _, item := range list {
		if len(selected) >= max {
			break
		}
		selected = append(selected, item.keyword)
	}
	return selected
}

func persistKeywords(db *gorm.DB, keywords map[string]*model.KeyWords, clusters [][]model.Article) error {
	kwList := make([]*model.KeyWords, 0, len(keywords))
	kwNames := make([]string, 0, len(keywords))

	for name, kw := range keywords {
		kwList = append(kwList, kw)
		kwNames = append(kwNames, name)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// Upsert keywords
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "keyword"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_update", "updated_at"}),
		}).Omit("Articles").CreateInBatches(kwList, 100).Error; err != nil {
			return err
		}

		// Get actual DB IDs
		var dbKws []model.KeyWords
		if err := tx.Where("keyword IN ?", kwNames).Find(&dbKws).Error; err != nil {
			return err
		}

		idMap := make(map[string]string, len(dbKws))
		for _, dk := range dbKws {
			idMap[dk.Keyword] = dk.ID
		}

		// Create join table entries
		joinRows := buildJoinRows(clusters, idMap, keywords)
		if len(joinRows) > 0 {
			if err := tx.Table("article_keywords").Clauses(clause.OnConflict{
				DoNothing: true,
			}).CreateInBatches(joinRows, 500).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func buildJoinRows(clusters [][]model.Article, idMap map[string]string, validKeywords map[string]*model.KeyWords) []map[string]interface{} {
	var joinRows []map[string]interface{}
	stopwords := getCombinedStopwords()
	titleFormatter := cases.Title(language.German)

	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}

		candidates := extractCandidates(cluster, stopwords)
		topKeywords := selectTopKeywords(candidates, 5)

		for _, kw := range topKeywords {
			cleaned := cleanKeyword(kw)
			formatted := titleFormatter.String(cleaned)

			// Only use keywords that survived merging
			if _, exists := validKeywords[formatted]; !exists {
				continue
			}

			if kwID, ok := idMap[formatted]; ok {
				for _, art := range cluster {
					joinRows = append(joinRows, map[string]interface{}{
						"key_words_id": kwID,
						"article_id":   art.ID,
					})
				}
			}
		}
	}
	return joinRows
}

// Helper functions

func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	for i, w := range words {
		words[i] = cleanWord(w)
	}
	return words
}

func cleanWord(w string) string {
	return strings.TrimFunc(w, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func cleanKeyword(kw string) string {
	return strings.TrimFunc(kw, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func getCombinedStopwords() map[string]bool {
	combined := make(map[string]bool)
	maps.Copy(combined, StopwordsEn)
	maps.Copy(combined, StopwordsDe)
	return combined
}

func isCommonEntity(word string) bool {
	common := map[string]bool{
		"news":    true,
		"article": true,
		"report":  true,
		"says":    true,
		"heute":   true,
	}
	return common[word]
}
