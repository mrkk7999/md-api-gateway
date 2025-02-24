package router

import (
	"md-api-gateway/caches"
	"md-api-gateway/middleware"
	"md-api-gateway/proxy"
	"md-api-gateway/utils/token"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// NewRouter sets up the API Gateway router
func NewRouter(cache caches.Cache, log *logrus.Logger) *mux.Router {
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return middleware.LoggingMiddleware(next, log)
	})

	for _, service := range proxy.ServiceMap {
		for route, roles := range service.Routes {

			handler := createProxyHandler(route)

			// Apply AuthMiddleware only if roles are defined for the route
			if len(roles) > 0 {
				handler = middleware.AuthMiddleware(handler, cache).(http.HandlerFunc)
			}

			// Add specific handler for sign-out route
			if route == "/api/v1/sign-out" {
				r.HandleFunc(route, signOutHandler(cache)).Methods("POST")
			}

			r.HandleFunc(route, handler.ServeHTTP)

		}
	}

	return r
}

// createProxyHandler returns a handler function for proxying requests
func createProxyHandler(route string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target, found := proxy.GetServiceTarget(route)
		if !found {
			http.Error(w, "Service Not Found", http.StatusNotFound)
			return
		}

		proxyHandler := proxy.ReverseProxy(target)
		proxyHandler.ServeHTTP(w, r)
	}
}

// signOutHandler handles the sign-out route and invalidates the cache
func signOutHandler(cache caches.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := token.ExtractToken(r.Header.Get("Authorization"))
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Invalidate the cache for the token
		err := cache.InvalidateHash(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "Failed to invalidate cache", http.StatusInternalServerError)
			return
		}

		// Forward the request to the actual sign-out service
		target, found := proxy.GetServiceTarget("/api/v1/sign-out")
		if !found {
			http.Error(w, "Service Not Found", http.StatusNotFound)
			return
		}

		proxyHandler := proxy.ReverseProxy(target)
		proxyHandler.ServeHTTP(w, r)
	}
}
