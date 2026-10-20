package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := getEnv("DATABASE_URL", "postgres://postgres:postgres@device-db:5432/devices")

	ctx := context.Background()
	store, err := NewStore(ctx, dsn)
	if err != nil {
		log.Fatalf("device-service: %v", err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		if err := store.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/devices", func(c *gin.Context) {
		devices, err := store.List(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, devices)
	})

	router.GET("/devices/:id", func(c *gin.Context) {
		device, err := store.Get(c.Request.Context(), c.Param("id"))
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "device_not_found", "message": "Устройство не найдено"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, device)
	})

	router.POST("/devices", func(c *gin.Context) {
		var input DeviceCreate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": err.Error()})
			return
		}

		device, err := store.Create(c.Request.Context(), input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, device)
	})

	router.PATCH("/devices/:id", func(c *gin.Context) {
		var input DeviceUpdate
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": err.Error()})
			return
		}

		device, err := store.UpdateStatus(c.Request.Context(), c.Param("id"), input.Status)
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": "device_not_found", "message": "Устройство не найдено"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, device)
	})

	server := &http.Server{
		Addr:              ":" + getEnv("PORT", "8082"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("device-service listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("device-service: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
