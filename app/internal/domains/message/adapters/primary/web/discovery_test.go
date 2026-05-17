package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func newDiscoveryEngine() (*gin.Engine, *MessageHandler) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockMessageService)
	handler := NewMessageHandler(mockService)
	engine := gin.New()
	engine.SetHTMLTemplate(createMockTemplate())
	engine.GET("/", handler.Home)
	engine.GET("/robots.txt", handler.RobotsTxt)
	engine.GET("/sitemap.xml", handler.SitemapXML)
	engine.GET("/.well-known/api-catalog", handler.APICatalog)
	return engine, handler
}

func TestRobotsTxt(t *testing.T) {
	engine, _ := newDiscoveryEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.Host = "example.test"
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "text/plain"))
	body := w.Body.String()
	assert.Contains(t, body, "Sitemap: http://example.test/sitemap.xml")
	assert.Contains(t, body, "Disallow: /decrypt/")
	assert.Contains(t, body, "Disallow: /confirmation")
	assert.Contains(t, body, "User-agent: *")
}

func TestRobotsTxt_RespectsForwardedProto(t *testing.T) {
	engine, _ := newDiscoveryEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/robots.txt", nil)
	req.Host = "password.exchange"
	req.Header.Set("X-Forwarded-Proto", "https")
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Sitemap: https://password.exchange/sitemap.xml")
}

func TestSitemapXML(t *testing.T) {
	engine, _ := newDiscoveryEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/sitemap.xml", nil)
	req.Host = "example.test"
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, strings.HasPrefix(w.Header().Get("Content-Type"), "application/xml"))
	body := w.Body.String()
	assert.Contains(t, body, `<?xml`)
	assert.Contains(t, body, `xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"`)
	assert.Contains(t, body, "<loc>http://example.test/</loc>")
	assert.Contains(t, body, "<loc>http://example.test/about</loc>")
	assert.Contains(t, body, "<loc>http://example.test/api/v1/docs/</loc>")
	assert.NotContains(t, body, "/decrypt/")
}

func TestAPICatalog(t *testing.T) {
	engine, _ := newDiscoveryEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/.well-known/api-catalog", nil)
	req.Host = "example.test"
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/linkset+json", w.Header().Get("Content-Type"))

	var doc linksetDocument
	err := json.Unmarshal(w.Body.Bytes(), &doc)
	assert.NoError(t, err)
	assert.Len(t, doc.Linkset, 1)

	entry := doc.Linkset[0]
	assert.Equal(t, "http://example.test/api/v1", entry.Anchor)
	assert.Len(t, entry.ServiceDesc, 1)
	assert.Equal(t, "http://example.test/api/v1/docs/doc.json", entry.ServiceDesc[0].Href)
	assert.Equal(t, "application/json", entry.ServiceDesc[0].Type)
	assert.Len(t, entry.ServiceDoc, 1)
	assert.Equal(t, "http://example.test/api/v1/docs/", entry.ServiceDoc[0].Href)
	assert.Len(t, entry.Status, 1)
	assert.Equal(t, "http://example.test/api/v1/health", entry.Status[0].Href)
}

func TestHome_LinkHeaders(t *testing.T) {
	engine, _ := newDiscoveryEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Host = "example.test"
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	links := w.Header().Values("Link")
	assert.Len(t, links, 2)

	joined := strings.Join(links, " | ")
	assert.Contains(t, joined, `</.well-known/api-catalog>; rel="api-catalog"`)
	assert.Contains(t, joined, `</api/v1/docs/>; rel="service-doc"`)
}
