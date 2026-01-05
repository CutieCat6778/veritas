package utils

import (
	"math"
	"strings"
	"time"

	"news-swipe/backend/graph/model"

	"github.com/pemistahl/lingua-go"
)

type SimilarityConfig struct {
	TitleWeight     float64
	DescWeight      float64
	TimeWeight      float64
	SourceWeight    float64
	SameDayBonus    float64
	SameWeekBonus   float64
	SameMonthBonus  float64
	DifferentSource float64
	MinTitleLength  int
	MinDescLength   int
}

func DefaultSimilarityConfig() SimilarityConfig {
	return SimilarityConfig{
		TitleWeight:     0.45,
		DescWeight:      0.35,
		TimeWeight:      0.15,
		SourceWeight:    0.05,
		SameDayBonus:    1.0,
		SameWeekBonus:   0.8,
		SameMonthBonus:  0.5,
		DifferentSource: 0.3,
		MinTitleLength:  10,
		MinDescLength:   20,
	}
}

var languageDetector = lingua.NewLanguageDetectorBuilder().
	FromLanguages(lingua.English, lingua.German).
	Build()

var StopwordsEn = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
	"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
	"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
	"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	"have": true, "this": true, "but": true, "or": true, "not": true, "been": true,
	"were": true, "they": true, "their": true, "can": true, "had": true,
}

var StopwordsDe = map[string]bool{
	"aber": true, "als": true, "am": true, "an": true, "auch": true, "auf": true,
	"aus": true, "bei": true, "bin": true, "bis": true, "bist": true, "da": true,
	"das": true, "dass": true, "dem": true, "den": true, "der": true, "des": true,
	"die": true, "dies": true, "diese": true, "diesem": true, "diesen": true,
	"dieser": true, "doch": true, "du": true, "durch": true, "ein": true,
	"eine": true, "einem": true, "einen": true, "einer": true, "eines": true,
	"er": true, "es": true, "für": true, "hab": true, "habe": true, "haben": true,
	"hat": true, "hatte": true, "hatten": true, "hier": true, "ich": true,
	"ihm": true, "ihn": true, "ihr": true, "im": true, "in": true, "ins": true,
	"ist": true, "ja": true, "kann": true, "machen": true, "mein": true,
	"mit": true, "nach": true, "nicht": true, "noch": true, "nur": true,
	"oder": true, "ohne": true, "sehr": true, "sein": true, "seine": true,
	"seinem": true, "seinen": true, "seiner": true, "sich": true, "sie": true,
	"sind": true, "so": true, "über": true, "um": true, "und": true, "uns": true,
	"von": true, "vor": true, "war": true, "waren": true, "warum": true,
	"was": true, "weil": true, "wenn": true, "wer": true, "wie": true,
	"wird": true, "wir": true, "wo": true, "wurde": true, "wurden": true,
	"zu": true, "zum": true, "zur": true,
}

func DetectArticleLanguage(title, description string) lingua.Language {
	text := title + " " + description
	if lang, exists := languageDetector.DetectLanguageOf(text); exists {
		return lang
	}
	return lingua.English
}

func ArticleSimilarity(a1, a2 model.Article, config SimilarityConfig) float64 {
	titleSim := enhancedStringSimilarity(a1.Title, a2.Title, a1.Language.ToLingua(), config.MinTitleLength)
	descSim := enhancedStringSimilarity(a1.Description, a2.Description, a1.Language.ToLingua(), config.MinDescLength)
	timeSim := timeSimilarityBucketed(a1.PublishedAt, a2.PublishedAt, config)
	sourceSim := sourceSimilarity(a1.Source, a2.Source, config)

	return (titleSim*config.TitleWeight +
		descSim*config.DescWeight +
		timeSim*config.TimeWeight +
		sourceSim*config.SourceWeight)
}

func enhancedStringSimilarity(s1, s2 string, lang lingua.Language, minLength int) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}
	if len(s1) < minLength || len(s2) < minLength {
		return 0.0
	}

	levDist := levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))
	levSim := 1.0 - (float64(levDist) / maxLen)

	var stopwords map[string]bool
	switch lang {
	case lingua.German:
		stopwords = StopwordsDe
	case lingua.English:
		stopwords = StopwordsEn
	default:
		stopwords = StopwordsEn
	}

	words1 := filterWithStopwords(strings.Fields(s1), stopwords)
	words2 := filterWithStopwords(strings.Fields(s2), stopwords)

	if len(words1) == 0 || len(words2) == 0 {
		return levSim * 0.5
	}

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, w := range words1 {
		set1[w] = true
	}
	for _, w := range words2 {
		set2[w] = true
	}

	intersection := 0
	for w := range set1 {
		if set2[w] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	jaccardSim := 0.0
	if union > 0 {
		jaccardSim = float64(intersection) / float64(union)
	}

	return (levSim*0.4 + jaccardSim*0.6)
}

func filterWithStopwords(words []string, stopwords map[string]bool) []string {
	filtered := make([]string, 0, len(words))
	for _, w := range words {
		if !stopwords[w] && len(w) > 2 {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for j := 0; j <= len(s2); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(s1); i++ {
		curr[0] = i
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func timeSimilarityBucketed(t1, t2 time.Time, config SimilarityConfig) float64 {
	if t1.After(t2) {
		t1, t2 = t2, t1
	}

	diff := t2.Sub(t1)
	hours := diff.Hours()

	if hours < 24 {
		return config.SameDayBonus
	}
	if hours < 24*7 {
		return config.SameWeekBonus
	}
	if hours < 24*30 {
		return config.SameMonthBonus
	}

	days := hours / 24
	return math.Max(0.0, 1.0-days/365.0)
}

func sourceSimilarity(s1, s2 model.Source, config SimilarityConfig) float64 {
	if s1 == s2 {
		return 0.2
	}
	return config.DifferentSource
}

func IsSimilar(a1, a2 model.Article, threshold float64, config SimilarityConfig) bool {
	calc := ArticleSimilarity(a1, a2, config)
	return calc >= threshold
}
