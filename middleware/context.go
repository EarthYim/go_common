package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Tier int

const (
	NormalTier Tier = iota
	ThrottledTier
	GlobalTier
	AuthTier
)

const (
	AppContextKey string = "app_context"
)

type AppContext struct {
	Tier Tier
	UID  string
	JWT  bool
}

func GetAppContext(c *gin.Context) (AppContext, error) {
	val, ok := c.Get(AppContextKey)
	if !ok {
		return AppContext{}, errors.New("ket is empty")
	}

	appContext, ok := val.(AppContext)
	if !ok {
		return AppContext{}, errors.New("failed type assertion")
	}

	return appContext, nil
}
