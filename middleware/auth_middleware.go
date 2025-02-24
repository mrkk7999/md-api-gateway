package middleware

import (
	"encoding/json"
	"md-api-gateway/caches"
	"md-api-gateway/config"
	"md-api-gateway/utils/token"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrorResponse represents a structured error message
type ErrorResponse struct {
	Message string `json:"message"`
}

// AuthMiddleware enforces role-based access control
func AuthMiddleware(next http.Handler, cache caches.Cache) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := token.ExtractToken(r.Header.Get("Authorization"))
		if tokenString == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if token is cached
		cachedRoles, err := cache.GetHash(r.Context(), tokenString, "roles")
		var userRoles []string

		if err != nil {
			// Token not found in cache, validate it
			claims, err := token.TokenAuth(r.Context(), tokenString)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Calculate remaining TTL based on token's expiration time
			expirationTime := time.Unix(int64((*claims)["exp"].(float64)), 0)
			remainingTTL := time.Until(expirationTime)

			// Extract required claims
			userRoles = getUserRoles(claims)

			// Cache the token claims and roles with remaining TTL
			claimsJSON, _ := json.Marshal(claims)
			cache.SetHash(r.Context(), tokenString, map[string]interface{}{
				"claims": string(claimsJSON),
				"roles":  strings.Join(userRoles, ","),
			})
			cache.SetTTL(r.Context(), tokenString, remainingTTL)
		} else {
			// Token found in cache, unmarshal roles
			userRoles = strings.Split(cachedRoles, ",")
		}

		if !isAuthorized(r.URL.Path, userRoles) {
			writeErrorResponse(w, http.StatusForbidden, "Forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getUserRoles(claims *jwt.MapClaims) []string {
	if groups, ok := (*claims)["cognito:groups"].([]interface{}); ok {
		var groupStrings []string
		for _, group := range groups {
			groupStrings = append(groupStrings, group.(string))
		}
		return groupStrings
	}
	return []string{}
}

func isAuthorized(path string, userRoles []string) bool {
	for _, service := range config.AuthSettings.Services {
		for route, roles := range service.Routes {
			if matchRoute(route, path) {
				return hasRole(userRoles, roles)
			}
		}
	}
	return false
}

func matchRoute(pattern, path string) bool {

	// Convert "{id}" to a regex pattern matching UUIDs or general IDs
	patternRegex := regexp.MustCompile(`\{[^/]+\}`)
	regexPattern := patternRegex.ReplaceAllString(pattern, `[^/]+`)

	// Ensure full match
	matched, err := regexp.MatchString("^"+regexPattern+"$", path)
	if err != nil {
		return false
	}

	return matched
}

func hasRole(userRoles, allowedRoles []string) bool {
	for _, userRole := range userRoles {
		if contains(allowedRoles, userRole) {
			return true
		}
	}
	return false
}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}
