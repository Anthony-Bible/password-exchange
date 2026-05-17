package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"

	"github.com/gin-gonic/gin"
)

// baseURL derives the absolute base URL ("scheme://host") for the incoming request,
// honoring X-Forwarded-Proto when present so handlers behind TLS-terminating proxies
// emit correct https links.
func baseURL(c *gin.Context) string {
	scheme := "http"
	if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if c.Request.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

// RobotsTxt serves /robots.txt advertising the sitemap and disallowing single-use secret URLs.
func (h *MessageHandler) RobotsTxt(c *gin.Context) {
	body := "User-agent: *\n" +
		"Allow: /\n" +
		"Disallow: /decrypt/\n" +
		"Disallow: /confirmation\n" +
		"\n" +
		"Sitemap: " + baseURL(c) + "/sitemap.xml\n"
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(body))
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// SitemapXML serves the public sitemap. Decrypt and confirmation URLs are intentionally omitted.
func (h *MessageHandler) SitemapXML(c *gin.Context) {
	base := baseURL(c)
	set := sitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs: []sitemapURL{
			{Loc: base + "/"},
			{Loc: base + "/about"},
			{Loc: base + "/api/v1/docs/"},
		},
	}
	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	body := append([]byte(xml.Header), out...)
	c.Data(http.StatusOK, "application/xml; charset=utf-8", body)
}

type linksetHref struct {
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

type linksetEntry struct {
	Anchor      string        `json:"anchor"`
	ServiceDesc []linksetHref `json:"service-desc,omitempty"`
	ServiceDoc  []linksetHref `json:"service-doc,omitempty"`
	Status      []linksetHref `json:"status,omitempty"`
}

type linksetDocument struct {
	Linkset []linksetEntry `json:"linkset"`
}

// APICatalog serves /.well-known/api-catalog per RFC 9727, pointing agents at the Swagger spec.
func (h *MessageHandler) APICatalog(c *gin.Context) {
	base := baseURL(c)
	doc := linksetDocument{
		Linkset: []linksetEntry{
			{
				Anchor: base + "/api/v1",
				ServiceDesc: []linksetHref{
					{Href: base + "/api/v1/docs/doc.json", Type: "application/json"},
				},
				ServiceDoc: []linksetHref{
					{Href: base + "/api/v1/docs/", Type: "text/html"},
				},
				Status: []linksetHref{
					{Href: base + "/api/v1/health"},
				},
			},
		},
	}
	body, err := json.Marshal(doc)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Data(http.StatusOK, "application/linkset+json", body)
}
