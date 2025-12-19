package middleware

import (
	"net/http"
)

func Concurrency(next http.HandlerFunc, limit int) http.HandlerFunc {
	limiter := make(chan struct{}, limit)
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case limiter <- struct{}{}:
			defer func() { <-limiter }()
			next.ServeHTTP(w, r)
		default:
			http.Error(w, "try later", http.StatusServiceUnavailable)
		}
	}
}
