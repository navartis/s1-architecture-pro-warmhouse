package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ProxyHandler struct {
	DeviceServiceURL    string
	TelemetryServiceURL string
	client              *http.Client
}

func NewProxyHandler(deviceServiceURL, telemetryServiceURL string) *ProxyHandler {
	return &ProxyHandler{
		DeviceServiceURL:    strings.TrimRight(deviceServiceURL, "/"),
		TelemetryServiceURL: strings.TrimRight(telemetryServiceURL, "/"),
		client:              &http.Client{},
	}
}

func (h *ProxyHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/devices", func(c *gin.Context) { h.forward(c, h.DeviceServiceURL) })
	router.POST("/devices", func(c *gin.Context) { h.forward(c, h.DeviceServiceURL) })
	router.GET("/devices/:id", func(c *gin.Context) { h.forward(c, h.DeviceServiceURL) })
	router.PATCH("/devices/:id", func(c *gin.Context) { h.forward(c, h.DeviceServiceURL) })

	router.GET("/telemetry", func(c *gin.Context) { h.forward(c, h.TelemetryServiceURL) })
	router.POST("/telemetry", func(c *gin.Context) { h.forward(c, h.TelemetryServiceURL) })
}

func (h *ProxyHandler) forward(c *gin.Context, target string) {
	path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
	url := target + path
	if c.Request.URL.RawQuery != "" {
		url += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, url, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": "proxy_error", "message": err.Error()})
		return
	}

	if contentType := c.Request.Header.Get("Content-Type"); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"code": "proxy_error", "message": err.Error()})
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}

	c.DataFromReader(resp.StatusCode, resp.ContentLength, contentType, resp.Body, nil)
}
