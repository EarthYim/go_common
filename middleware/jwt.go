package middleware

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

// JwtMiddleware validates and verifies JWT from Authorization header
func JwtMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Extract bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		base64Decoded, err := decodeBase64(cfg.Base64JwtPublicKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"msg": "cannot decode base64",
			})
		}

		// Parse public key
		publicKey, err := parsePublicKey(base64Decoded)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Parse and verify token
		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return publicKey, nil
		})

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check expiration
		if claims.ExpiresAt < time.Now().Unix() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Attach claims to context
		c.Set("claims", claims)
		c.Next()
	}
}

func parsePublicKey(key string) (*ecdsa.PublicKey, error) {
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
