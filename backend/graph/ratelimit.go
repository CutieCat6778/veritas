package graph

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"news-swipe/backend/utils"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

var rateLimitConfig = RateLimitConfig{
	RequestsPerMinute: 60,
	BurstSize:         10,
}

func init() {
	// Read rate limit config from environment
	if rpm := os.Getenv("RATE_LIMIT_RPM"); rpm != "" {
		if val, err := strconv.Atoi(rpm); err == nil {
			rateLimitConfig.RequestsPerMinute = val
		}
	}
	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if val, err := strconv.Atoi(burst); err == nil {
			rateLimitConfig.BurstSize = val
		}
	}
}

// RateLimitMiddleware implements Redis-based rate limiting
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting if Redis is not available
		if utils.RedisClient == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get client IP
		clientIP := getClientIP(r)
		if clientIP == "" {
			// If we can't determine IP, allow the request
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		allowed, retryAfter := checkRateLimit(ctx, clientIP)

		if !allowed {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitConfig.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(retryAfter).Unix()))
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		remaining := getRemainingRequests(ctx, clientIP)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitConfig.RequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		next.ServeHTTP(w, r)
	})
}

func checkRateLimit(ctx context.Context, clientIP string) (bool, time.Duration) {
	key := fmt.Sprintf("ratelimit:%s", clientIP)
	now := time.Now()
	windowStart := now.Add(-time.Minute)

	pipe := utils.RedisClient.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count requests in current window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiry
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		utils.Log(utils.GraphQL, "Rate limit check error:", err)
		// On error, allow the request
		return true, 0
	}

	count := countCmd.Val()

	// Check if limit exceeded (count is before adding current request)
	if count >= int64(rateLimitConfig.RequestsPerMinute) {
		// Calculate retry after
		oldestCmd := utils.RedisClient.ZRange(ctx, key, 0, 0)
		oldest, err := oldestCmd.Result()
		if err == nil && len(oldest) > 0 {
			var oldestTime int64
			fmt.Sscanf(oldest[0], "%d", &oldestTime)
			retryAfter := time.Unix(0, oldestTime).Add(time.Minute).Sub(now)
			if retryAfter < 0 {
				retryAfter = time.Second
			}
			return false, retryAfter
		}
		return false, time.Minute
	}

	return true, 0
}

func getRemainingRequests(ctx context.Context, clientIP string) int {
	key := fmt.Sprintf("ratelimit:%s", clientIP)
	count, err := utils.RedisClient.ZCard(ctx, key).Result()
	if err != nil {
		return rateLimitConfig.RequestsPerMinute
	}

	remaining := rateLimitConfig.RequestsPerMinute - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func getClientIP(r *http.Request) string {
	// Try X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		return forwarded
	}

	// Try X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
