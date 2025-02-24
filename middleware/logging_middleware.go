package middleware

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

// LoggingMiddleware logs the incoming request details and completion time
func LoggingMiddleware(next http.Handler, log *logrus.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("#######################################################################")
		log.Infof("Incoming request: Method=%s URL=%s  ", r.Method, r.URL.Path)

		// Call next handler
		next.ServeHTTP(w, r)
		log.Println("#######################################################################")
	})
}
