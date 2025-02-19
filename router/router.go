package router

import (
	"md-api-gateway/middleware"
	"md-api-gateway/proxy"
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter sets up the API Gateway router
func NewRouter() *mux.Router {
	r := mux.NewRouter()
	r.Use(middleware.LoggingMiddleware)

	for _, service := range proxy.ServiceMap {
		for route, roles := range service.Routes {
			handler := createProxyHandler(route)

			// Apply AuthMiddleware only if roles are defined for the route
			if len(roles) > 0 {
				handler = middleware.AuthMiddleware(handler).(http.HandlerFunc)
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

// func NewRouter() http.Handler {
// 	r := http.NewServeMux()

// 	// Define routes for the API
// 	r.Handle("/health", middleware.AuthMiddleware(http.HandlerFunc(HealthCheckHandler)))
// 	// r.Handle("/api/v1/auth/", http.StripPrefix("/api/v1/auth", proxy.NewProxy("http://localhost:9001")))
// 	// r.Handle("/api/v1/tenant/", middleware.AuthMiddleware(http.StripPrefix("/api/v1/tenant", proxy.NewProxy("http://localhost:9002"))))
// 	// r.Handle("/api/v1/geo-track/", middleware.AuthMiddleware(http.StripPrefix("/api/v1/geo-track", proxy.NewProxy("http://localhost:9003"))))
// 	// r.Handle("/api/v1/geo-stream/", middleware.AuthMiddleware(http.StripPrefix("/api/v1/geo-stream", proxy.NewProxy("http://localhost:9004"))))

// 	return r
// }
// func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte(`{"message": "Health check passed"}`))
// }
