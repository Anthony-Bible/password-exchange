package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRedactingSlogLogger_RedactsKeyQueryParam(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := gin.New()
	router.Use(RedactingLogger(&buf))
	router.GET("/api/v1/files/:fileID", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/abc123?key=SUPERSECRETKEY==", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logged := buf.String()
	if strings.Contains(logged, "SUPERSECRETKEY") {
		t.Errorf("log output must not contain the raw key value; got: %s", logged)
	}
	if !strings.Contains(logged, "/api/v1/files/abc123") {
		t.Errorf("log output should contain the path; got: %s", logged)
	}
}

func TestRedactingSlogLogger_PreservesOtherQueryParams(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := gin.New()
	router.Use(RedactingLogger(&buf))
	router.GET("/api/v1/files/:fileID", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/abc123?foo=bar&key=SECRET==&baz=qux", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logged := buf.String()
	// The key value must be redacted but the other params must remain visible.
	if strings.Contains(logged, "SECRET") {
		t.Errorf("log output must not contain the raw key value; got: %s", logged)
	}
	if !strings.Contains(logged, "foo=bar") {
		t.Errorf("log output should retain non-sensitive query params; got: %s", logged)
	}
	if !strings.Contains(logged, "baz=qux") {
		t.Errorf("log output should retain non-sensitive query params; got: %s", logged)
	}
}

func TestRedactingSlogLogger_NoKeyParam_Unmodified(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	router := gin.New()
	router.Use(RedactingLogger(&buf))
	router.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	logged := buf.String()
	if !strings.Contains(logged, "/healthz") {
		t.Errorf("log output should contain path; got: %s", logged)
	}
}
