package token

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/golang-jwt/jwt/v5"
)

// ExtractToken extracts JWT from the "Authorization: Bearer <token>" header
func ExtractToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}

var (
	jwks      *keyfunc.JWKS
	jwksOnce  sync.Once
	jwksError error
)

// InitJWKS initializes JWKS once at startup
func InitJWKS() {
	jwksOnce.Do(func() {
		awsRegion := os.Getenv("COG_REGION")
		userPoolId := os.Getenv("COG_USER_POOL_ID")
		jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", awsRegion, userPoolId)

		jwks, jwksError = keyfunc.Get(jwksURL, keyfunc.Options{RefreshInterval: time.Hour})
		if jwksError != nil {
			log.Println("InitJWKS: Failed to load JWKS: %v", jwksError)
		}
	})
}

// TokenAuth validates JWT using Cognito (if enabled) and then JWKS
func TokenAuth(ctx context.Context, tokenString string) (*jwt.MapClaims, error) {
	authMode := os.Getenv("TOKEN_AUTH")

	// Check with Cognito if enabled
	if authMode == "cognito" {
		if err := checkWithCognito(ctx, tokenString); err != nil {
			return nil, err
		}
	}

	// Validate JWT Signature using cached JWKS
	if jwks == nil {
		return nil, errors.New("JWKS not initialized")
	}

	token, err := jwt.Parse(tokenString, jwks.Keyfunc)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}

	return &claims, nil
}

// checkWithCognito verifies the token directly with AWS Cognito
func checkWithCognito(ctx context.Context, token string) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("COG_REGION")))
	if err != nil {
		log.Println("checkWithCognito: Failed to load AWS config:", err)
		return errors.New("internal server error")
	}

	client := cognitoidentityprovider.NewFromConfig(cfg)

	_, err = client.GetUser(ctx, &cognitoidentityprovider.GetUserInput{AccessToken: &token})
	if err != nil {
		return errors.New("token is invalid or expired")
	}

	return nil
}
