package auth

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type AuthClient struct {
	client *auth.Client
}

// NewFirebaseAuthClient initialises Firebase Admin with service account file and returns auth client.
func NewFirebaseAuthClient(ctx context.Context, projectID string, credentialsFile string) (*AuthClient, error) {
	opts := []option.ClientOption{}
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	cfg := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, cfg, opts...)
	if err != nil {
		return nil, err
	}
	c, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	return &AuthClient{client: c}, nil
}

// AuthMiddleware verifies "Authorization: Bearer <Firebase ID token>" and sets claims in context.
func AuthMiddleware(ac *AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearer(c.Request.Header.Get("Authorization"))
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tok, err := ac.client.VerifyIDToken(c, tokenString)
		if err != nil {
			log.Printf("verify id token error: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// Set useful fields in context
		c.Set("firebase_uid", tok.UID)
		c.Set("firebase_claims", tok.Claims)
		c.Next()
	}
}

func extractBearer(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// Helpers to retrieve from gin.Context
func MustUID(c *gin.Context) string {
	if v, ok := c.Get("firebase_uid"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetClaims(c *gin.Context) map[string]interface{} {
	if v, ok := c.Get("firebase_claims"); ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return map[string]interface{}{}
}

// For unit tests / local mock
var ErrNoAuthHeader = errors.New("no auth header")
