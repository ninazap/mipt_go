package api

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware middleware для логирования запросов
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		next.ServeHTTP(w, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s %v", r.Method, r.URL.Path, duration)
	})
}