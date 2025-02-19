package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the incoming request details and completion time
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Println("-----------------------------------------------------------")
		log.Printf("Incoming request: Method=%s URL=%s Timestamp=%s", r.Method, r.URL.Path, startTime.Format(time.RFC3339))

		// Call next handler
		next.ServeHTTP(w, r)

		endTime := time.Now()
		log.Printf("Request completed: Method=%s URL=%s CompletedAt=%s Duration=%v", r.Method, r.URL.Path, endTime.Format(time.RFC3339), endTime.Sub(startTime))
		log.Println("-----------------------------------------------------------")
	})
}
