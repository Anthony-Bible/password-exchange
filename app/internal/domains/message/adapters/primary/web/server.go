package web

import (
	"html/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WebServer handles HTTP requests for the message service
type WebServer struct {
	messageHandler *MessageHandler
	messageService primary.MessageServicePort
	apiServer      *api.Server
	router         *gin.Engine
}

// NewWebServer creates a new web server
func NewWebServer(messageService primary.MessageServicePort) *WebServer {
	messageHandler := NewMessageHandler(messageService)
	apiServer := api.NewServer(messageService)
	
	router := gin.Default()
	
	// Create template functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
	}
	
	// Load HTML templates with custom functions
	router.SetFuncMap(funcMap)
	router.LoadHTMLGlob("/templates/*.html")
	router.Static("/assets", "/templates/assets")
	
	return &WebServer{
		messageHandler: messageHandler,
		messageService: messageService,
		apiServer:      apiServer,
		router:         router,
	}
}

// SetupRoutes configures the HTTP routes
func (s *WebServer) SetupRoutes() {
	// Setup API routes directly on the main router
	s.setupAPIRoutes()
	
	// Static pages
	s.router.GET("/", s.messageHandler.Home)
	s.router.GET("/about", s.messageHandler.About)
	s.router.GET("/confirmation", s.messageHandler.Confirmation)
	
	// Message operations
	s.router.POST("/", s.messageHandler.SubmitMessage)
	s.router.GET("/decrypt/:uuid/*key", s.messageHandler.DisplayDecrypted)
	s.router.POST("/decrypt/:uuid/*key", s.messageHandler.DecryptMessage)
	
	// 404 handler
	s.router.NoRoute(s.messageHandler.NotFound)
	
	log.Info().Msg("Web routes and API routes configured")
}

// setupAPIRoutes adds API routes to the main router
func (s *WebServer) setupAPIRoutes() {
	// Create API handler directly with the message service
	apiHandler := api.NewMessageAPIHandler(s.messageService)
	
	// Add API middleware
	apiGroup := s.router.Group("/api")
	apiGroup.Use(func(c *gin.Context) {
		// Add CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Correlation-ID")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	
	// Add correlation ID and error handling middleware
	apiGroup.Use(middleware.CorrelationID())
	apiGroup.Use(middleware.ErrorHandler())
	
	// API v1 routes
	v1 := apiGroup.Group("/v1")
	{
		// Message endpoints
		v1.POST("/messages", apiHandler.SubmitMessage)
		v1.GET("/messages/:id", apiHandler.GetMessageInfo)
		v1.POST("/messages/:id/decrypt", apiHandler.DecryptMessage)
		
		// Utility endpoints
		v1.GET("/health", apiHandler.HealthCheck)
		v1.GET("/info", apiHandler.APIInfo)
	}
	
	log.Info().Msg("API routes configured directly on main router")
}

// Start starts the web server
func (s *WebServer) Start() error {
	s.SetupRoutes()
	
	log.Info().Msg("Starting web server")
	return s.router.Run() // Default port :8080
}

// GetRouter returns the Gin router for testing
func (s *WebServer) GetRouter() *gin.Engine {
	return s.router
}