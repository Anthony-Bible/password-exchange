package api

import (
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	_ "github.com/Anthony-Bible/password-exchange/app/docs" // Import generated docs
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/secondary"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server represents the API server
type Server struct {
	handler           *MessageAPIHandler
	fileHandler       *FileAPIHandler
	router            *gin.Engine
	metricsRegistry   *prometheus.Registry
	prometheusMetrics *middleware.PrometheusMetrics
}

// NewServer creates a new API server with the given message service plus the
// encryption and storage ports the /readyz probe needs to check downstream
// dependency health.
func NewServer(
	messageService primary.MessageServicePort,
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
) *Server {
	return newServer(messageService, encryptionService, storageService, nil)
}

// NewServerWithFileHandler creates a new API server and optionally registers the
// chunked file upload endpoints when a file handler is provided.
func NewServerWithFileHandler(
	messageService primary.MessageServicePort,
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
	fileHandler *FileAPIHandler,
) *Server {
	return newServer(messageService, encryptionService, storageService, fileHandler)
}

func newServer(
	messageService primary.MessageServicePort,
	encryptionService secondary.EncryptionServicePort,
	storageService secondary.StorageServicePort,
	fileHandler *FileAPIHandler,
) *Server {
	handler := NewMessageAPIHandler(messageService, encryptionService, storageService)

	metricsRegistry := prometheus.NewRegistry()
	prometheusMetrics := middleware.NewPrometheusMetrics(metricsRegistry)
	router := setupRouter(handler, prometheusMetrics, metricsRegistry, fileHandler)

	return &Server{
		handler:           handler,
		fileHandler:       fileHandler,
		router:            router,
		metricsRegistry:   metricsRegistry,
		prometheusMetrics: prometheusMetrics,
	}
}

// GetRouter returns the configured Gin router
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// setupRouter configures the API routes and middleware
func setupRouter(
	handler *MessageAPIHandler,
	prometheusMetrics *middleware.PrometheusMetrics,
	metricsRegistry *prometheus.Registry,
	fileHandler *FileAPIHandler,
) *gin.Engine {
	router := gin.New()
	// Use a custom recovery that re-panics http.ErrAbortHandler so net/http's
	// serve() goroutine can cleanly close the TCP connection when streaming fails
	// mid-response. gin.Recovery() would otherwise swallow the sentinel and call
	// AbortWithStatus(500) on an already-committed 200 response.
	router.Use(gin.CustomRecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err any) {
		if err == http.ErrAbortHandler {
			panic(err)
		}
		fmt.Fprintf(gin.DefaultErrorWriter, "panic recovered: %v\n%s\n", err, debug.Stack())
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// Global middleware — use a redacting logger so the ?key= query parameter
	// (AES-256 decryption key for file downloads) is never written to access logs.
	router.Use(middleware.RedactingLogger(os.Stdout))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CorrelationID())
	router.Use(middleware.PrometheusMiddleware(prometheusMetrics)) // Add Prometheus metrics collection
	router.Use(middleware.CustomRateLimitErrorHandler())
	router.Use(middleware.ValidationMiddleware())
	router.Use(middleware.RequestTimeoutMiddleware(30 * time.Second))

	// CORS middleware - allow all origins for now
	router.Use(func(c *gin.Context) {
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

	// Process-alive (/livez) and dependency-aware (/readyz) probes live on the
	// root, not under /api/v1, so k8s probes don't get versioned along with
	// the public REST surface. They are deliberately NOT rate limited: the
	// kubelet probes both endpoints from a single node IP every few seconds,
	// which blows past any per-IP limit and turns a 429 into a failed readiness
	// probe that pulls a healthy pod out of rotation.
	router.GET("/livez", handler.Livez)
	router.GET("/readyz", handler.Readyz)

	// API routes with rate limiting
	v1 := router.Group("/api/v1")
	{
		// Message endpoints with specific rate limits
		messages := v1.Group("/messages")
		{
			messages.POST("", middleware.MessageSubmissionRateLimit(), handler.SubmitMessage)
			messages.GET("/:id", middleware.MessageAccessRateLimit(), handler.GetMessageInfo)
			messages.POST("/:id/decrypt", middleware.MessageDecryptRateLimit(), handler.DecryptMessage)
		}

		if fileHandler != nil {
			files := v1.Group("/files")
			files.POST("/initiate", middleware.FileInitiateRateLimit(), fileHandler.InitiateUpload)
			files.POST("/:fileID/chunks", middleware.FileUploadRateLimit(), fileHandler.UploadChunk)
			files.GET("/:fileID", middleware.MessageAccessRateLimit(), fileHandler.DownloadFile)
		}

		// Utility endpoints with lenient rate limits
		v1.GET("/health", middleware.HealthCheckRateLimit(), handler.HealthCheck)
		v1.GET("/info", middleware.MessageAccessRateLimit(), handler.APIInfo)

		// Documentation endpoints with lenient rate limits
		v1.GET("/docs/*any", middleware.HealthCheckRateLimit(), ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Metrics endpoint (outside rate limiting to avoid interfering with monitoring)
	router.GET("/metrics", middleware.PrometheusHandler(metricsRegistry))

	return router
}
