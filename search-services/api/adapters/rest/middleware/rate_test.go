package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRate(t *testing.T) {
	var requestCount int64

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Test with rate limit of 2 requests per second
	rps := 2
	rateMiddleware := Rate(testHandler, rps)

	// Send 5 requests quickly
	numRequests := 5
	results := make([]int, numRequests)
	start := time.Now()

	for i := range numRequests {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		rateMiddleware.ServeHTTP(w, req)
		results[i] = w.Code
	}

	duration := time.Since(start)

	// With rate limit of 2 RPS, 5 requests should take at least 2 seconds
	// (first 2 requests immediate, then 3 more with delays)
	minExpectedDuration := 2 * time.Second

	assert.True(t, duration >= minExpectedDuration,
		"duration %v should be at least %v for rate limit of %d RPS",
		duration, minExpectedDuration, rps)

	// All requests should succeed
	for _, code := range results {
		assert.Equal(t, http.StatusOK, code, "all requests should succeed")
	}
}

func TestRate_ConcurrentRequests(t *testing.T) {
	var requestCount int64
	var mu sync.Mutex

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestCount++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})

	// Test with rate limit of 5 requests per second
	rps := 5
	rateMiddleware := Rate(testHandler, rps)

	// Send 10 concurrent requests
	numRequests := 10
	var wg sync.WaitGroup
	wg.Add(numRequests)

	results := make([]int, numRequests)
	start := time.Now()

	for i := range numRequests {
		go func(index int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			rateMiddleware.ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// With rate limit of 5 RPS, 10 requests should take at least 1.8 seconds
	// (first 5 requests immediate, then 5 more with delays)
	minExpectedDuration := 1800 * time.Millisecond

	assert.True(t, duration >= minExpectedDuration,
		"duration %v should be at least %v for rate limit of %d RPS",
		duration, minExpectedDuration, rps)

	// All requests should succeed
	for _, code := range results {
		assert.Equal(t, http.StatusOK, code, "all requests should succeed")
	}
}

func TestRate_ZeroRate(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with rate limit of 0 (should block all requests)
	rps := 0
	rateMiddleware := Rate(testHandler, rps)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// With rate limit of 0, request should be rejected
	rateMiddleware.ServeHTTP(w, req)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code, "request should be rejected with zero rate limit")
}

// TestRateLimiterDirectly tests the rate limiter directly
func TestRateLimiterDirectly(t *testing.T) {
	// Test creating a rate limiter directly
	limiter := rate.NewLimiter(rate.Limit(1), 1)

	// First request should succeed immediately
	assert.True(t, limiter.Allow(), "first request should be allowed")

	// Second request should be denied (bursts = 1)
	assert.False(t, limiter.Allow(), "second request should be denied")

	// Wait for token refill
	time.Sleep(1100 * time.Millisecond) // slightly more than 1 second

	// Third request should succeed
	assert.True(t, limiter.Allow(), "third request should be allowed after waiting")
}
