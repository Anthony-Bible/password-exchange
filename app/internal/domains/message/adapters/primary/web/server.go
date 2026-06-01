package web

import (
	"html/template"
	"net/http"
	"os"

	_ "github.com/Anthony-Bible/password-exchange/app/docs" // Import generated docs
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/Anthony-Bible/password-exchange/app/internal/shared/logging"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// WebServer handles HTTP requests for the message service
type WebServer struct {
	messageHandler    *MessageHandler
	messageService    primary.MessageServicePort
	encryptionService secondary.EncryptionServicePort
	storageService    secondary.StorageServicePort
	fileHandler       *api.FileAPIHandler
	router            *gin.Engine
}

// NewWebServer creates a new web server. The encryption and storage ports are
// required so the root-level /readyz probe can verify downstream gRPC services.
func NewWebServer(
	messageService primary.MessageServicePort,
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
) *WebServer {
	messageHandler := NewMessageHandler(messageService)

	// Use gin.New() + RedactingLogger instead of gin.Default() so that the
	// ?key= query parameter (AES-256 file-decryption key) is never written to
	// access logs. gin.Default() wraps gin.Logger() which logs the full URL
	// including all query params. gin.Recovery() is added to retain panic
	// recovery behaviour from the original Default.
	router := gin.New()
	router.Use(middleware.RedactingLogger(os.Stdout))
	router.Use(gin.Recovery())

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
		messageHandler:    messageHandler,
		messageService:    messageService,
		encryptionService: encryptionService,
		storageService:    storageService,
		router:            router,
	}
}

// WithFileHandler registers the optional file API handler used for chunked uploads.
func (s *WebServer) WithFileHandler(fileHandler *api.FileAPIHandler) *WebServer {
	s.fileHandler = fileHandler
	return s
}

// SetupRoutes configures the HTTP routes
func (s *WebServer) SetupRoutes() {
	// Setup API routes directly on the main router
	s.setupAPIRoutes()

	// Static pages
	s.router.GET("/", s.messageHandler.Home)
	s.router.GET("/about", s.messageHandler.About)
	s.router.GET("/confirmation", s.messageHandler.Confirmation)

	// File download page — serves the download UI for a shared file link.
	// The decryption key is carried in the URL fragment (#key=...) so it
	// never reaches the server; the page JS reads it and calls the file API.
	s.router.GET("/files/:fileID", s.messageHandler.FileDownload)

	// Agent discovery (sitemap, robots, API catalog)
	s.router.GET("/robots.txt", s.messageHandler.RobotsTxt)
	s.router.GET("/sitemap.xml", s.messageHandler.SitemapXML)
	s.router.GET("/.well-known/api-catalog", s.messageHandler.APICatalog)

	// Message operations
	s.router.POST("/", s.messageHandler.SubmitMessage)
	s.router.GET("/decrypt/:uuid", s.messageHandler.DisplayDecrypted)
	s.router.GET("/decrypt/:uuid/*key", s.messageHandler.DisplayDecrypted)
	s.router.POST("/decrypt/:uuid", s.messageHandler.DecryptMessage)
	s.router.POST("/decrypt/:uuid/*key", s.messageHandler.DecryptMessage)

	// 404 handler
	s.router.NoRoute(s.messageHandler.NotFound)

	logging.Info().Msg("Web routes and API routes configured")
}

// setupAPIRoutes adds API routes to the main router
func (s *WebServer) setupAPIRoutes() {
	// Create API handler directly with the message service plus the two
	// secondary ports that /readyz needs to probe.
	apiHandler := api.NewMessageAPIHandler(s.messageService, s.encryptionService, s.storageService)

	// Process-alive and dependency-aware probes live on the root, not under
	// /api/v1 — k8s probe paths shouldn't be versioned alongside the public
	// REST surface. They're attached to s.router (not the apiGroup) so they
	// stay reachable at the documented /livez and /readyz paths. They are
	// deliberately NOT rate limited: the kubelet probes both endpoints from a
	// single node IP every few seconds, which blows past any per-IP limit and
	// turns a 429 into a failed readiness probe that pulls a healthy pod out of
	// rotation.
	s.router.GET("/livez", apiHandler.Livez)
	s.router.GET("/readyz", apiHandler.Readyz)

	// Add API middleware
	apiGroup := s.router.Group("/api")
	apiGroup.Use(func(c *gin.Context) {
		// Add CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header(
			"Access-Control-Allow-Headers",
			"Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Correlation-ID, X-File-Key",
		)

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

		if s.fileHandler != nil {
			files := v1.Group("/files")
			files.POST("/initiate", middleware.FileInitiateRateLimit(), s.fileHandler.InitiateUpload)
			files.POST("/:fileID/chunks", middleware.FileUploadRateLimit(), s.fileHandler.UploadChunk)
			files.GET("/:fileID", middleware.MessageAccessRateLimit(), s.fileHandler.DownloadFile)
		}

		// Utility endpoints
		v1.GET("/health", apiHandler.HealthCheck)
		v1.GET("/info", apiHandler.APIInfo)

		// Documentation endpoints — gin-swagger v1.6+ requires a specific filename in
		// the URL (index.html, doc.json, etc.); a bare /docs/ doesn't match its regex
		// and returns 404, so redirect the root docs URL to index.html.
		v1.GET("/docs", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/index.html") })
		v1.GET("/docs/", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/api/v1/docs/index.html") })
		v1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	logging.Info().Msg("API routes configured directly on main router")
}

// Start starts the web server
func (s *WebServer) Start() error {
	s.SetupRoutes()

	logging.Info().Msg("Starting web server")
	return s.router.Run() // Default port :8080
}

// GetRouter returns the Gin router for testing
func (s *WebServer) GetRouter() *gin.Engine {
	return s.router
}
