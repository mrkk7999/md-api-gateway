package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// ServiceConfig holds service information
type ServiceConfig struct {
	Target string
	Routes map[string][]string
}

// ServiceMap stores all registered services
var ServiceMap = make(map[string]ServiceConfig)

// LoadConfig loads the service mappings from auth.yml (convert it to a Go struct first)
func LoadConfig(config map[string]ServiceConfig) {
	ServiceMap = config
}

// GetServiceTarget finds the appropriate backend service based on the request path
func GetServiceTarget(path string) (string, bool) {
	for _, service := range ServiceMap {
		for route := range service.Routes {

			// Normalize strings
			normalizedPath := strings.TrimSpace(strings.TrimSpace(path))
			normalizedRoute := strings.TrimSpace(strings.TrimSpace(route))

			if strings.HasPrefix(normalizedPath, normalizedRoute) {
				return service.Target, true
			}
		}
	}
	return "", false
}

// ReverseProxy creates a reverse proxy for the given service target
func ReverseProxy(target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("##########")
			fmt.Println("##########")
			log.Println(r.URL.Path)
			fmt.Println(r.URL.Path)
			fmt.Println("##########")
			log.Println("##########")
			http.Error(w, "Bad Gateway: Invalid target URL", http.StatusBadGateway)
		})
	}

	return httputil.NewSingleHostReverseProxy(targetURL)
}

// // NewProxy creates a reverse proxy to the specified target URL
// func NewProxy(target string) http.Handler {
// 	// Parse the target URL to create a proxy
// 	targetURL, err := url.Parse(target)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create and return the reverse proxy
// 	proxy := httputil.NewSingleHostReverseProxy(targetURL)
// 	return proxy
// }
