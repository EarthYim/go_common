package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	JwtAdminPrivateKey string = "place_holder_private"
	JwtPublicKey       string = "place_public_private"
)

func AdminLoginHandler(c *gin.Context) {
	// Create claims with expiration
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		IssuedAt:  time.Now().Unix(),
		Audience:  "admin-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	// Parse private key
	// privateKey, err := parsePrivateKey(JwtAdminPrivateKey)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{
	// 		"error": "failed to parse private key",
	// 	})
	// 	return
	// }

	// Sign token
	ss, err := token.SignedString(JwtAdminPrivateKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to sign token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": ss,
	})
}

func parsePrivateKey(key string) (*ecdsa.PrivateKey, error) {
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
