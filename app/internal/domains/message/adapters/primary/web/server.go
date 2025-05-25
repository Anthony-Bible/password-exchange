package web

import (
	"html/template"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// WebServer handles HTTP requests for the message service
type WebServer struct {
	messageHandler *MessageHandler
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
	// Get the configured API router and mount its handlers
	apiRouter := s.apiServer.GetRouter()
	
	// Mount the entire API router under /api prefix
	s.router.Any("/api/*path", func(c *gin.Context) {
		// Update the path to remove the /api prefix for the API router
		c.Request.URL.Path = c.Param("path")
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}
		apiRouter.ServeHTTP(c.Writer, c.Request)
	})
	
	log.Info().Msg("API routes mounted on main router")
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