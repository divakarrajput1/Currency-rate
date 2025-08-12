package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"exchange-rate-service/internal/cache"
	"exchange-rate-service/internal/external"
	"exchange-rate-service/internal/handlers"
	"exchange-rate-service/internal/services"
)

func main() {
	log.Println("Starting Exchange Rate Service...")

	cacheService := cache.NewMemoryCache(1 * time.Hour) // 1 hour TTL
	apiClient := external.NewExchangeRateClient()
	rateFetcher := services.NewRateFetcher(apiClient, cacheService)
	exchangeService := services.NewExchangeService(cacheService, rateFetcher, apiClient)
	handler := handlers.NewExchangeHandler(exchangeService)

	rateFetcher.Start()

	router := setupRouter(handler)

	setupGracefulShutdown(rateFetcher)

	port := getPort()
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(handler *handlers.ExchangeHandler) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(corsMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		v1.POST("/convert", handler.ConvertCurrency)
		v1.GET("/convert", handler.ConvertCurrencyQuery)

		// Rate endpoints
		v1.GET("/rates/latest", handler.GetLatestRate)
		v1.POST("/rates/historical", handler.GetHistoricalRates)
		v1.GET("/rates/historical", handler.GetHistoricalRatesQuery)

		v1.GET("/currencies", handler.GetSupportedCurrencies)
		v1.GET("/health", handler.GetHealth)
		v1.GET("/stats/cache", handler.GetCacheStats)
	}

	router.GET("/health", handler.GetHealth)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Exchange Rate Service",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func setupGracefulShutdown(rateFetcher *services.RateFetcher) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		rateFetcher.Stop()
		os.Exit(0)
	}()
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}
