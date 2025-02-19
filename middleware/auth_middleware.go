package middleware

import (
	"md-api-gateway/config"
	"md-api-gateway/utils/token"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware enforces role-based access control
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := token.ExtractToken(r.Header.Get("Authorization"))
		if tokenString == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims, err := token.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userRoles := getUserRoles(claims)

		if !isAuthorized(r.URL.Path, userRoles) {
			http.Error(w, "Forbidden", http.StatusForbidden)
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
	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "/*"))
	}
	return pattern == path
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

// func AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		authHeader := r.Header.Get("Authorization")
// 		if authHeader == "" {
// 			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
// 			return
// 		}

// 		tokenString := token.ExtractToken(authHeader)
// 		if tokenString == "" {
// 			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
// 			return
// 		}

// 		claims, err := token.ValidateToken(tokenString)
// 		if err != nil {
// 			http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
// 			return
// 		}
// 		fmt.Println(claims)

// 		next.ServeHTTP(w, r)
// 	})
// }
