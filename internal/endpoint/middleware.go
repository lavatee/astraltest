package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lavatee/astraltest/internal/model"
	"github.com/sirupsen/logrus"
)

type BodyWithToken struct {
	Token string `json:"token" binding:"required"`
}

func (e *Endpoint) Middleware(c *gin.Context) {
	var token string
	var req BodyWithToken
	var metaData model.DocumentMeta
	meta := c.PostForm("meta")
	if err := json.Unmarshal([]byte(meta), &metaData); err == nil {
		token = metaData.Token
	}
	if err := c.ShouldBindJSON(&req); err == nil {
		token = req.Token
	}
	if c.Query("token") != "" {
		token = c.Query("token")
	}
	if token == "" {
		logrus.Error("Middleware error: token required")
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
			Error: model.ErrorInfo{
				Text: "Authorization token required",
				Code: 401,
			},
		})
		return
	}
	valid, err := e.services.Auth.ValidateToken(c.Request.Context(), token)
	if err != nil || !valid {
		logrus.Error("Middleware error: token is invalid")
		c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{
			Error: model.ErrorInfo{
				Text: "Invalid token",
				Code: 401,
			},
		})
		return
	}
	c.Set("token", token)
	c.Next()
}
