package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"news-swipe/backend/graph/model"

	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

var (
	stopwordsCache      map[string]bool
	stopwordsCacheOnce  sync.Once
	titleFormatterCache cases.Caser
	formatterOnce       sync.Once
	stringBuilderPool   = sync.Pool{
		New: func() interface{} {
			return &strings.Builder{}
		},
	}
)

func getStopwords() map[string]bool {
	stopwordsCacheOnce.Do(func() {
		stopwordsCache = make(map[string]bool)
		maps.Copy(stopwordsCache, StopwordsEn)
		maps.Copy(stopwordsCache, StopwordsDe)
	})
	return stopwordsCache
}

func getTitleFormatter() cases.Caser {
	formatterOnce.Do(func() {
		titleFormatterCache = cases.Title(language.German)
	})
	return titleFormatterCache
}

func GenerateKeywordsFromArticles(db *gorm.DB) error {
	cutoff := time.Now().AddDate(0, 0, -14)

	var articles []model.Article
	if err := db.Select("id, title, description, published_at").
		Where("published_at >= ?", cutoff).
		Find(&articles).Error; err != nil {
		return err
	}
	if len(articles) == 0 {
		return nil
	}

	articles = deduplicateByTitle(articles)
	config := DefaultSimilarityConfig()
	clusters := clusterArticles(articles, 0.36, config)
	keywords := extractAndMergeKeywords(clusters)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM article_keywords").Error; err != nil {
			return err
		}

		if err := tx.Exec("DELETE FROM key_words").Error; err != nil {
			return err
		}

		if len(keywords) == 0 {
			return nil
		}

		return persistKeywordsInTx(tx, keywords)
	})
}
func deduplicateByTitle(articles []model.Article) []model.Article {
	seen := make(map[string]bool, len(articles))
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
	clusters := make([][]model.Article, 0, len(articles)/5)
	visited := make(map[string]bool, len(articles))

	for i, a1 := range articles {
		if visited[a1.ID] {
			continue
		}

		cluster := make([]model.Article, 1, 10)
		cluster[0] = a1
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

type keywordData struct {
	keyword        string
	totalFrequency int
	articleSet     map[string]bool
	articles       []string
}

func extractAndMergeKeywords(clusters [][]model.Article) map[string]*keywordData {
	stopwords := getStopwords()
	titleFormatter := getTitleFormatter()

	keywordToData := make(map[string]*keywordData)

	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}

		articleIDs := make([]string, len(cluster))
		articleSet := make(map[string]bool, len(cluster))
		for i, art := range cluster {
			articleIDs[i] = art.ID
			articleSet[art.ID] = true
		}
		sort.Strings(articleIDs)

		candidates := extractCandidates(cluster, stopwords)
		topKeywords := selectTopKeywords(candidates, 8)

		for _, kw := range topKeywords {
			cleaned := cleanKeyword(kw)
			if len(cleaned) < 4 || isCommonEntity(strings.ToLower(cleaned)) {
				continue
			}

			formatted := titleFormatter.String(cleaned)

			if existing, exists := keywordToData[formatted]; exists {
				for _, artID := range articleIDs {
					if !existing.articleSet[artID] {
						existing.articleSet[artID] = true
						existing.articles = append(existing.articles, artID)
					}
				}
				existing.totalFrequency += candidates[kw]
			} else {
				articleSetCopy := make(map[string]bool, len(articleSet))
				for k, v := range articleSet {
					articleSetCopy[k] = v
				}
				articlesCopy := make([]string, len(articleIDs))
				copy(articlesCopy, articleIDs)

				keywordToData[formatted] = &keywordData{
					keyword:        formatted,
					totalFrequency: candidates[kw],
					articleSet:     articleSetCopy,
					articles:       articlesCopy,
				}
			}
		}
	}

	for _, kw := range keywordToData {
		sort.Strings(kw.articles)
	}

	keywordToData = deduplicateByArticleSet(keywordToData)
	keywordToData = removeSubsets(keywordToData)

	return keywordToData
}

func deduplicateByArticleSet(allKeywords map[string]*keywordData) map[string]*keywordData {
	hashToKeywords := make(map[string][]*keywordData)

	for _, kw := range allKeywords {
		hash := hashArticleSet(kw.articles)
		hashToKeywords[hash] = append(hashToKeywords[hash], kw)
	}

	result := make(map[string]*keywordData, len(allKeywords))

	for _, keywords := range hashToKeywords {
		if len(keywords) == 1 {
			kw := keywords[0]
			result[kw.keyword] = kw
			continue
		}

		sort.Slice(keywords, func(i, j int) bool {
			if keywords[i].totalFrequency != keywords[j].totalFrequency {
				return keywords[i].totalFrequency > keywords[j].totalFrequency
			}
			return keywords[i].keyword < keywords[j].keyword
		})

		best := keywords[0]
		result[best.keyword] = best
	}

	return result
}

func removeSubsets(allKeywords map[string]*keywordData) map[string]*keywordData {
	type kwItem struct {
		keyword string
		data    *keywordData
	}

	items := make([]kwItem, 0, len(allKeywords))
	for k, v := range allKeywords {
		items = append(items, kwItem{k, v})
	}

	sort.Slice(items, func(i, j int) bool {
		if len(items[i].data.articles) != len(items[j].data.articles) {
			return len(items[i].data.articles) > len(items[j].data.articles)
		}
		if items[i].data.totalFrequency != items[j].data.totalFrequency {
			return items[i].data.totalFrequency > items[j].data.totalFrequency
		}
		return items[i].keyword < items[j].keyword
	})

	result := make(map[string]*keywordData, len(allKeywords))
	toRemove := make(map[string]bool)

	for i, item := range items {
		if toRemove[item.keyword] {
			continue
		}

		result[item.keyword] = item.data

		for j := i + 1; j < len(items); j++ {
			other := items[j]
			if toRemove[other.keyword] {
				continue
			}

			if isSubset(other.data.articleSet, item.data.articleSet) {
				toRemove[other.keyword] = true
			}
		}
	}

	return result
}

func isSubset(smaller, larger map[string]bool) bool {
	if len(smaller) > len(larger) {
		return false
	}

	for articleID := range smaller {
		if !larger[articleID] {
			return false
		}
	}
	return true
}

func hashArticleSet(articleIDs []string) string {
	sb := stringBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	defer stringBuilderPool.Put(sb)

	for _, id := range articleIDs {
		sb.WriteString(id)
		sb.WriteByte(',')
	}

	hash := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(hash[:16])
}

func extractCandidates(cluster []model.Article, stopwords map[string]bool) map[string]int {
	candidates := make(map[string]int, 50)

	for _, article := range cluster {
		sb := stringBuilderPool.Get().(*strings.Builder)
		sb.Reset()
		sb.WriteString(article.Title)
		sb.WriteByte(' ')
		sb.WriteString(article.Description)
		text := sb.String()
		stringBuilderPool.Put(sb)

		words := tokenize(text)

		for i := 0; i < len(words); i++ {
			w := words[i]
			if len(w) > 3 && !stopwords[w] && !isCommonEntity(w) {
				candidates[w] += 3
			}

			if i < len(words)-1 {
				w2 := words[i+1]
				if len(w2) > 3 && !stopwords[w2] {
					sb := stringBuilderPool.Get().(*strings.Builder)
					sb.Reset()
					sb.WriteString(w)
					sb.WriteByte(' ')
					sb.WriteString(w2)
					bigram := sb.String()
					stringBuilderPool.Put(sb)
					candidates[bigram] += 5
				}
			}
		}
	}
	return candidates
}

func selectTopKeywords(candidates map[string]int, max int) []string {
	if len(candidates) == 0 {
		return nil
	}

	type scoredKw struct {
		keyword string
		score   int
	}

	list := make([]scoredKw, 0, len(candidates))
	for k, v := range candidates {
		list = append(list, scoredKw{k, v})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score
		}
		return list[i].keyword < list[j].keyword
	})

	limit := max
	if len(list) < max {
		limit = len(list)
	}

	selected := make([]string, limit)
	for i := 0; i < limit; i++ {
		selected[i] = list[i].keyword
	}
	return selected
}

func persistKeywordsInTx(tx *gorm.DB, keywords map[string]*keywordData) error {
	if len(keywords) == 0 {
		return nil
	}

	kwList := make([]*model.KeyWords, 0, len(keywords))
	kwNames := make([]string, 0, len(keywords))

	for name := range keywords {
		kwList = append(kwList, &model.KeyWords{
			GormModel:  model.GormModel{ID: uuid.NewString()},
			Keyword:    name,
			LastUpdate: time.Now(),
		})
		kwNames = append(kwNames, name)
	}

	if err := tx.Omit("Articles").CreateInBatches(kwList, 100).Error; err != nil {
		return err
	}

	var dbKws []model.KeyWords
	if err := tx.Select("id, keyword").Where("keyword IN ?", kwNames).Find(&dbKws).Error; err != nil {
		return err
	}

	idMap := make(map[string]string, len(dbKws))
	for _, dk := range dbKws {
		idMap[dk.Keyword] = dk.ID
	}

	joinRows := make([]map[string]interface{}, 0, len(keywords)*10)
	for name, data := range keywords {
		kwID, ok := idMap[name]
		if !ok {
			continue
		}

		for _, artID := range data.articles {
			joinRows = append(joinRows, map[string]interface{}{
				"key_words_id": kwID,
				"article_id":   artID,
			})
		}
	}

	if len(joinRows) > 0 {
		const batchSize = 500
		for i := 0; i < len(joinRows); i += batchSize {
			end := i + batchSize
			if end > len(joinRows) {
				end = len(joinRows)
			}

			if err := tx.Table("article_keywords").Create(joinRows[i:end]).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
func tokenize(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	result := make([]string, len(words))
	for i, w := range words {
		result[i] = cleanWord(w)
	}
	return result
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

var commonEntities = map[string]bool{
	"news":    true,
	"article": true,
	"report":  true,
	"says":    true,
	"heute":   true,
}

func isCommonEntity(word string) bool {
	return commonEntities[word]
}
