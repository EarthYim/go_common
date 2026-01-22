package auth

import (
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
