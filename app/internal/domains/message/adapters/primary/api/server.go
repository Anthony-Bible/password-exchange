package api

import (
	"time"

	_ "github.com/Anthony-Bible/password-exchange/app/docs" // Import generated docs
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server represents the API server
type Server struct {
	handler *MessageAPIHandler
	router  *gin.Engine
}

// NewServer creates a new API server with the given message service
func NewServer(messageService primary.MessageServicePort) *Server {
	handler := NewMessageAPIHandler(messageService)
	router := setupRouter(handler)
	
	return &Server{
		handler: handler,
		router:  router,
	}
}

// GetRouter returns the configured Gin router
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// setupRouter configures the API routes and middleware
func setupRouter(handler *MessageAPIHandler) *gin.Engine {
	router := gin.New()
	
	// Global middleware
	router.Use(gin.Logger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CorrelationID())
	router.Use(middleware.ValidationMiddleware())
	router.Use(middleware.RequestTimeoutMiddleware(30 * time.Second))
	
	// CORS middleware - allow all origins for now
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Correlation-ID")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		// Message endpoints
		v1.POST("/messages", handler.SubmitMessage)
		v1.GET("/messages/:id", handler.GetMessageInfo)
		v1.POST("/messages/:id/decrypt", handler.DecryptMessage)
		
		// Utility endpoints
		v1.GET("/health", handler.HealthCheck)
		v1.GET("/info", handler.APIInfo)
		
		// Documentation endpoints
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	
	return router
}