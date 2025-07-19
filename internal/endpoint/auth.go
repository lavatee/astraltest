package endpoint

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lavatee/astraltest/internal/model"
	"github.com/sirupsen/logrus"
)

func (e *Endpoint) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("Failed to register user (invalid request): %s", err.Error())
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 400, Text: "Invalid request format"},
		})
		return
	}
	login, err := e.services.Auth.Register(c.Request.Context(), req)
	if err != nil {
		logrus.Errorf("Failed to register user (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Response: gin.H{"login": login},
	})
}

func (e *Endpoint) Authenticate(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("Failed to authenticate user (invalid request): %s", err.Error())
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 400, Text: "Invalid request format"},
		})
		return
	}
	token, err := e.services.Auth.Authenticate(c.Request.Context(), req)
	if err != nil {
		logrus.Errorf("Failed to authenticate user (internal error): %s", err.Error())
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 401, Text: "Invalid credentials"},
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Response: gin.H{"token": token},
	})
}

func (e *Endpoint) Logout(c *gin.Context) {
	token := c.Param("token")
	if err := e.services.Auth.Logout(c.Request.Context(), token); err != nil {
		logrus.Errorf("Failed to logout user (internal error): %s", err.Error())
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error: model.ErrorInfo{Code: 500, Text: err.Error()},
		})
		return
	}
	c.JSON(http.StatusOK, model.Response{
		Response: gin.H{token: true},
	})
}
