package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Class int

const (
	ClassNew Class = iota
	ClassReturning
	ClassFallback
)

type ClientIdentity struct {
	ID    string
	Class Class
}

type contextKey string

const IdentityKey contextKey = "client_identity"

func AnonymousSessionTracking() gin.HandlerFunc {
	return func(c *gin.Context) {
		ident, ok := detectIdentity(c)
		if !ok {
			ident = issueIdentity(c)
		}

		c.Set(string(IdentityKey), ident)
		c.Next()
	}
}

func detectIdentity(c *gin.Context) (ClientIdentity, bool) {
	cookie, err := c.Cookie("anon_id")
	if err != nil || cookie == "" {
		return ClientIdentity{}, false
	}
	return ClientIdentity{
		ID:    cookie,
		Class: ClassReturning,
	}, true
}

func issueIdentity(c *gin.Context) ClientIdentity {
	id := uuid.NewString()

	c.SetCookie(
		"anon_id",
		id,
		60*60,
		"/",
		"",
		true,
		true,
	)

	return ClientIdentity{
		ID:    id,
		Class: ClassNew,
	}
}
