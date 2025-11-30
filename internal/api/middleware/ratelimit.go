package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*clientLimit
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}

type clientLimit struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	log.Printf("[INFO] Rate limiter initialized: %d requests per minute", requestsPerMinute)
	rl := &RateLimiter{
		requests: make(map[string]*clientLimit),
		limit:    requestsPerMinute,
		window:   time.Minute,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.RemoteAddr
		}

		rl.mu.Lock()
		client, exists := rl.requests[apiKey]
		now := time.Now()

		if !exists || now.After(client.resetTime) {
			rl.requests[apiKey] = &clientLimit{
				count:     1,
				resetTime: now.Add(rl.window),
			}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if client.count >= rl.limit {
			rl.mu.Unlock()
			log.Printf("[WARN] Rate limit exceeded for client: %s, path: %s", apiKey, r.URL.Path)
			w.Header().Set("X-RateLimit-Limit", string(rune(rl.limit)))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		client.count++
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cleanedCount := 0
		for key, client := range rl.requests {
			if now.After(client.resetTime) {
				delete(rl.requests, key)
				cleanedCount++
			}
		}
		rl.mu.Unlock()
		if cleanedCount > 0 {
			log.Printf("[DEBUG] Rate limiter cleanup: removed %d expired entries", cleanedCount)
		}
	}
}
