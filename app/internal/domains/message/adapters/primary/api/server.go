package api

import (
	"time"

	_ "github.com/Anthony-Bible/password-exchange/app/docs" // Import generated docs
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/middleware"
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/ports/primary"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Server represents the API server
type Server struct {
	handler           *MessageAPIHandler
	router            *gin.Engine
	metricsRegistry   *prometheus.Registry
	prometheusMetrics *middleware.PrometheusMetrics
}

// NewServer creates a new API server with the given message service
func NewServer(messageService primary.MessageServicePort) *Server {
	handler := NewMessageAPIHandler(messageService)

	// Initialize Prometheus metrics
	metricsRegistry := prometheus.NewRegistry()
	prometheusMetrics := middleware.NewPrometheusMetrics(metricsRegistry)

	router := setupRouter(handler, prometheusMetrics, metricsRegistry)

	return &Server{
		handler:           handler,
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
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
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
			"Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Correlation-ID",
		)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

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
