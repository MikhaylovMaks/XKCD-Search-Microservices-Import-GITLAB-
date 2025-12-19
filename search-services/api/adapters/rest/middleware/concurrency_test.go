package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrency(t *testing.T) {
	var requestCount int64
	var concurrentRequests int64

	// Create a test handler that simulates some work
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)

		// Track concurrent requests
		atomic.AddInt64(&concurrentRequests, 1)
		defer atomic.AddInt64(&concurrentRequests, -1)

		// Simulate work that takes some time
		time.Sleep(10 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	// Test with limit of 2
	limit := 2
	concurrencyMiddleware := Concurrency(testHandler, limit)

	// Send 5 concurrent requests
	numRequests := 5
	var wg sync.WaitGroup
	wg.Add(numRequests)

	results := make([]int, numRequests)
	for i := range numRequests {
		go func(index int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			concurrencyMiddleware.ServeHTTP(w, req)

			results[index] = w.Code
		}(i)
	}

	wg.Wait()

	// Check results
	successCount := 0
	busyCount := 0
	for _, code := range results {
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusServiceUnavailable:
			busyCount++
		}
	}

	// Should have some successful requests and some busy responses
	assert.True(t, successCount > 0, "should have at least one successful request")
	assert.True(t, busyCount > 0, "should have at least one busy response")
	assert.Equal(t, numRequests, successCount+busyCount, "all requests should be accounted for")
}

func TestConcurrency_LimitOne(t *testing.T) {
	var concurrentRequests int64
	var maxConcurrent int64

	// Create a test handler that tracks concurrency
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&concurrentRequests, 1)
		defer atomic.AddInt64(&concurrentRequests, -1)

		// Track the maximum concurrent requests
		for {
			max := atomic.LoadInt64(&maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt64(&maxConcurrent, max, current) {
				break
			}
		}

		// Simulate longer work to ensure concurrency control
		time.Sleep(50 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
	})

	// Test with limit of 1
	limit := 1
	concurrencyMiddleware := Concurrency(testHandler, limit)

	// Send 3 concurrent requests
	numRequests := 3
	var wg sync.WaitGroup
	wg.Add(numRequests)

	results := make([]int, numRequests)

	for i := range numRequests {
		go func(index int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			concurrencyMiddleware.ServeHTTP(w, req)

			results[index] = w.Code
		}(i)
	}

	wg.Wait()

	// With limit of 1, only one request should succeed, others should get 503
	successCount := 0
	blockedCount := 0
	for _, code := range results {
		switch code {
		case http.StatusOK:
			successCount++
		case http.StatusServiceUnavailable:
			blockedCount++
		}
	}
	assert.Equal(t, 1, successCount, "exactly one request should succeed with limit of 1")
	assert.Equal(t, 2, blockedCount, "two requests should be blocked with limit of 1")

	// Maximum concurrent requests should never exceed 1
	assert.Equal(t, int64(1), atomic.LoadInt64(&maxConcurrent), "should never exceed concurrency limit")
}

func TestConcurrency_ZeroLimit(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test with limit of 0 (should reject all requests)
	limit := 0
	concurrencyMiddleware := Concurrency(testHandler, limit)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	concurrencyMiddleware.ServeHTTP(w, req)

	// With limit of 0, request should be immediately rejected
	assert.Equal(t, http.StatusServiceUnavailable, w.Code, "request should be rejected with limit of 0")
}
