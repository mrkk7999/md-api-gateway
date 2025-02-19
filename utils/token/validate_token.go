package token

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

// ExtractToken extracts JWT from the "Authorization: Bearer <token>" header
func ExtractToken(authHeader string) string {
	log.Println("ExtractToken: Received Authorization header:", authHeader)

	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		log.Println("ExtractToken: Extracted token successfully")

		return parts[1]
	}
	log.Println("ExtractToken: Failed to extract token")

	return ""
}

// ValidateToken validates JWT using AWS Cognito JWKS
func ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	log.Println("ValidateToken: Starting token validation")

	awsRegion := os.Getenv("COG_REGION")
	userPoolId := os.Getenv("COG_USER_POOL_ID")
	jwksURL := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v/.well-known/jwks.json", awsRegion, userPoolId)

	// jwksURL := fmt.Sprintf("https://cognito-idp.eu-north-1.amazonaws.com/eu-north-1_OIlvRiSLJ/.well-known/jwks.json")

	// Fetch JWKS (JSON Web Key Set)
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{RefreshInterval: time.Hour})
	if err != nil {
		log.Printf("ValidateToken: Failed to fetch JWKS: %v", err)

		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)

	if err != nil || !token.Valid {
		log.Printf("ValidateToken: Token parsing failed: %v", err)

		return nil, errors.New("invalid token")
	}
	log.Println("ValidateToken: Token is valid")

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("unable to extract claims")
	}
	log.Println("ValidateToken: Successfully extracted claims", claims)

	return &claims, nil
}
