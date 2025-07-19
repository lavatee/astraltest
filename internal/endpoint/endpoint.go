package endpoint

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lavatee/astraltest/internal/service"
)

type Endpoint struct {
	services *service.Service
}

func NewEndpoint(services *service.Service) *Endpoint {
	return &Endpoint{
		services: services,
	}
}

func (e *Endpoint) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "X-Auth-Token, Content-Type, Origin, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Creditionals", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
	})
	public := router.Group("/api")
	{
		public.POST("/register", e.Register)
		public.POST("/auth", e.Authenticate)
	}
	protected := router.Group("/api", e.Middleware)
	{
		protected.POST("/docs", e.UploadDocument)
		protected.GET("/docs", e.GetDocuments)
		protected.GET("/docs/:id", e.GetDocument)
		protected.DELETE("/docs/:id", e.DeleteDocument)
		protected.DELETE("/auth/:token", e.Logout)
	}
	return router
}
