package auth

import (
	"common/config"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type authHandler struct {
	cfg config.Config
}

func NewAuthHandler(cfg config.Config) *authHandler {
	return &authHandler{
		cfg: cfg,
	}
}

func (h *authHandler) AdminLoginHandler(c *gin.Context) {
	// Create claims with expiration
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
		Audience:  "admin-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	base64decoded, err := decodeBase64(h.cfg.Base64JwtPrivateKey)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Parse private key
	privateKey, err := parsePrivateKey(base64decoded)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Sign token
	ss, err := token.SignedString(privateKey)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": ss,
	})
}

func parsePrivateKey(key string) (*ecdsa.PrivateKey, error) {
	// Handle escaped newlines from env variables
	key = strings.ReplaceAll(key, "\\n", "\n")

	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	return privateKey, nil
}

func parsePublicKey(key string) (*ecdsa.PublicKey, error) {
	// Handle escaped newlines from env variables
	key = strings.ReplaceAll(key, "\\n", "\n")

	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	publicKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an ECDSA key")
	}

	return publicKey, nil
}

func decodeBase64(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// // JwtMiddleware validates and verifies JWT from Authorization header
// func JwtMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Extract token from Authorization header
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "missing authorization header",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Extract bearer token
// 		parts := strings.Split(authHeader, " ")
// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "invalid authorization header format",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		tokenString := parts[1]

// 		// Parse public key
// 		publicKey, err := parsePublicKey(cfg.JwtPublicKey)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "failed to parse public key",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Parse and verify token
// 		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
// 			// Verify signing method
// 			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return publicKey, nil
// 		})

// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "invalid token",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		if !token.Valid {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "token is not valid",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Extract claims
// 		claims, ok := token.Claims.(*jwt.StandardClaims)
// 		if !ok {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "invalid token claims",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Check expiration
// 		if claims.ExpiresAt < time.Now().Unix() {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": "token has expired",
// 			})
// 			c.Abort()
// 			return
// 		}

// 		// Attach claims to context
// 		c.Set("claims", claims)

// 		c.Next()
// 	}
// }
