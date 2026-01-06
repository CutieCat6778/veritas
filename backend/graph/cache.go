package graph

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"news-swipe/backend/utils"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCacheMiddleware wraps the GraphQL handler with Redis caching
func RedisCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only cache GET requests
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching if Redis is not available
		if utils.RedisClient == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key based on query and variables
		cacheKey, err := generateCacheKey(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()

		// Try to get from cache
		cached, err := utils.RedisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache hit
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write([]byte(cached))
			return
		} else if err != redis.Nil {
			// Redis error, continue without cache
			utils.Log(utils.GraphQL, "Redis cache error:", err)
		}

		// Cache miss - create response recorder
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &strings.Builder{},
		}

		// Call next handler
		next.ServeHTTP(recorder, r)

		// Cache the response if status is 200
		if recorder.statusCode == http.StatusOK && recorder.body.Len() > 0 {
			// Default cache duration (can be overridden by cache-control headers)
			cacheDuration := 5 * time.Minute

			// Try to parse cache-control header
			if cacheControl := recorder.Header().Get("Cache-Control"); cacheControl != "" {
				if duration := parseCacheControl(cacheControl); duration > 0 {
					cacheDuration = duration
				}
			}

			// Store in Redis
			err = utils.RedisClient.Set(ctx, cacheKey, recorder.body.String(), cacheDuration).Err()
			if err != nil {
				utils.Log(utils.GraphQL, "Failed to cache response:", err)
			}
		}

		// Copy headers from recorder to actual response
		for k, v := range recorder.Header() {
			w.Header()[k] = v
		}

		// Write the actual response
		w.Header().Set("X-Cache", "MISS")
		w.WriteHeader(recorder.statusCode)
		w.Write([]byte(recorder.body.String()))
	})
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *strings.Builder
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	// Don't write to the underlying ResponseWriter yet - just record the status
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	// Only record the body, don't write to the underlying ResponseWriter yet
	return r.body.Write(b)
}

func generateCacheKey(r *http.Request) (string, error) {
	var query string
	var variables map[string]interface{}

	if r.Method == http.MethodGet {
		query = r.URL.Query().Get("query")
		variablesStr := r.URL.Query().Get("variables")
		if variablesStr != "" {
			if err := json.Unmarshal([]byte(variablesStr), &variables); err != nil {
				return "", err
			}
		}
	} else if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return "", err
		}
		// Restore body for next handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		var reqBody struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := json.Unmarshal(body, &reqBody); err != nil {
			return "", err
		}
		query = reqBody.Query
		variables = reqBody.Variables
	}

	// Create cache key from query and variables
	keyData := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: variables,
	}

	keyBytes, err := json.Marshal(keyData)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(keyBytes)
	return fmt.Sprintf("gql:cache:%x", hash), nil
}

func parseCacheControl(cacheControl string) time.Duration {
	// Parse "max-age=XXX" from cache-control header
	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "max-age=") {
			var seconds int
			fmt.Sscanf(part, "max-age=%d", &seconds)
			return time.Duration(seconds) * time.Second
		}
	}
	return 0
}
