package graph

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"news-swipe/backend/graph/model"

	"github.com/pemistahl/lingua-go"
)

var filterContextKey = &contextKey{"filter"}
var languageContextKey = &contextKey{"language"}

type contextKey struct {
	name string
}

func FilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawFilter := r.Header.Get("Filter")
		if rawFilter == "" {
			next.ServeHTTP(w, r)
			return
		}
		filters := strings.Split(rawFilter, ",")
		sourcesFilter := make([]model.Source, 0)

		for _, filter := range filters {
			switch strings.TrimSpace(filter) {
			case model.SourceFaz.String():
				sourcesFilter = append(sourcesFilter, model.SourceFaz)
			case model.SourceDieZeit.String():
				sourcesFilter = append(sourcesFilter, model.SourceDieZeit)
			case model.SourceHandelsblatt.String():
				sourcesFilter = append(sourcesFilter, model.SourceHandelsblatt)
			case model.SourceSueddeutsche.String():
				sourcesFilter = append(sourcesFilter, model.SourceSueddeutsche)
			case model.SourceTagesschau.String():
				sourcesFilter = append(sourcesFilter, model.SourceTagesschau)
			case model.SourceWelt.String():
				sourcesFilter = append(sourcesFilter, model.SourceWelt)
			case model.SourceTaz.String():
				sourcesFilter = append(sourcesFilter, model.SourceTaz)
			}
		}
		ctx := context.WithValue(r.Context(), filterContextKey, sourcesFilter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LanguageMiddleware parses the Accept-Language header and stores the preferred language in context
func LanguageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptLanguage := r.Header.Get("Accept-Language")
		preferredLang := parseAcceptLanguage(acceptLanguage)

		ctx := context.WithValue(r.Context(), languageContextKey, preferredLang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// parseAcceptLanguage parses the Accept-Language header and returns the preferred language
// Defaults to German if no valid language is found
func parseAcceptLanguage(header string) model.Language {
	if header == "" {
		return model.FromLingua(lingua.German)
	}

	// Parse Accept-Language header (format: "en-US,en;q=0.9,de;q=0.8")
	languages := strings.Split(header, ",")

	type langWithQuality struct {
		lang    model.Language
		quality float32
	}

	candidates := []langWithQuality{}

	for _, lang := range languages {
		lang = strings.TrimSpace(lang)

		// Split by quality factor
		parts := strings.Split(lang, ";")
		langCode := strings.TrimSpace(parts[0])

		// Extract just the language code (before any dash)
		if dashIndex := strings.Index(langCode, "-"); dashIndex != -1 {
			langCode = langCode[:dashIndex]
		}

		// Parse quality factor (default is 1.0)
		quality := float32(1.0)
		if len(parts) > 1 {
			qPart := strings.TrimSpace(parts[1])
			if strings.HasPrefix(qPart, "q=") {
				var q float32
				_, err := fmt.Sscanf(qPart, "q=%f", &q)
				if err == nil {
					quality = q
				}
			}
		}

		// Convert to our Language type
		langCode = strings.ToLower(langCode)
		var modelLang model.Language

		switch langCode {
		case "en":
			modelLang = model.FromLingua(lingua.English)
		case "de":
			modelLang = model.FromLingua(lingua.German)
		case "fr":
			modelLang = model.FromLingua(lingua.French)
		case "es":
			modelLang = model.FromLingua(lingua.Spanish)
		case "it":
			modelLang = model.FromLingua(lingua.Italian)
		case "pt":
			modelLang = model.FromLingua(lingua.Portuguese)
		case "nl":
			modelLang = model.FromLingua(lingua.Dutch)
		case "pl":
			modelLang = model.FromLingua(lingua.Polish)
		case "ru":
			modelLang = model.FromLingua(lingua.Russian)
		case "tr":
			modelLang = model.FromLingua(lingua.Turkish)
		case "zh":
			modelLang = model.FromLingua(lingua.Chinese)
		case "ja":
			modelLang = model.FromLingua(lingua.Japanese)
		case "ar":
			modelLang = model.FromLingua(lingua.Arabic)
		default:
			continue // Skip unknown languages
		}

		candidates = append(candidates, langWithQuality{lang: modelLang, quality: quality})
	}

	// If no valid languages found, default to German
	if len(candidates) == 0 {
		return model.FromLingua(lingua.German)
	}

	// Find the language with highest quality factor
	bestLang := candidates[0]
	for _, candidate := range candidates[1:] {
		if candidate.quality > bestLang.quality {
			bestLang = candidate
		}
	}

	return bestLang.lang
}

// GetLanguageFromContext retrieves the language from the request context
// Returns German as default if language is not found in context
func GetLanguageFromContext(ctx context.Context) model.Language {
	if lang, ok := ctx.Value(languageContextKey).(model.Language); ok {
		return lang
	}
	return model.FromLingua(lingua.German)
}
