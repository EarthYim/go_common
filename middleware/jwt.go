package middleware

import (
	"common/config"

	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	ClaimsKey string = "claims"
)

// JwtMiddleware validates and verifies JWT from Authorization header
func JwtMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		token := ValidateJwt(c, authHeader, cfg)

		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.ExpiresAt < time.Now().Unix() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(AppContextKey, AppContext{
			Tier: GlobalTier,
			UID:  claims.Subject,
		})

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

func JwtOptionalMiddleware(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// case no jwt -> throttled still needs ip, user-agent, x-device-fingerprint
		if authHeader == "" {
			parts := []string{c.ClientIP()}
			if ua := c.GetHeader("User-Agent"); ua != "" {
				parts = append(parts, ua)
			}

			if deviceFP := c.GetHeader("X-Device-Fingerprint"); deviceFP != "" {
				parts = append(parts, deviceFP)
			}

			fp := GetDeviceFingerprint(c)

			c.Set(AppContextKey, AppContext{
				Tier: ThrottledTier,
				UID:  fp,
			})
			c.Next()
		}

		token := ValidateJwt(c, authHeader, cfg)

		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if claims.ExpiresAt < time.Now().Unix() {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set(ClaimsKey, claims)
		c.Set(AppContextKey, AppContext{
			Tier: NormalTier,
			UID:  claims.Subject,
		})

		c.Next()
	}
}

func GetDeviceFingerprint(c *gin.Context) string {
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	deviceFP := c.GetHeader("X-Device-Fingerprint")

	if ua == "" || deviceFP == "" {
		c.AbortWithStatus(http.StatusForbidden)
		return ""
	}

	combined := fmt.Sprintf("%s|%s|%s", ip, ua, deviceFP)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

func GetClaims(c *gin.Context) (*jwt.StandardClaims, error) {
	val, ok := c.Get(ClaimsKey)
	if !ok {
		return &jwt.StandardClaims{}, errors.New("key is empty")
	}

	claims, ok := val.(*jwt.StandardClaims)
	if !ok {
		return &jwt.StandardClaims{}, errors.New("failed type assertion")
	}

	return claims, nil
}

func ValidateJwt(c *gin.Context, authHeader string, cfg config.Config) *jwt.Token {
	// Extract bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}

	tokenString := parts[1]

	base64Decoded, err := decodeBase64(cfg.Base64JwtPublicKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"msg": "cannot decode base64",
		})
		return nil
	}

	// Parse public key
	publicKey, err := parsePublicKey(base64Decoded)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return nil
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
		return nil
	}

	if !token.Valid {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil
	}

	return token
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
