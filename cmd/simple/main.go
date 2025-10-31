package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Create Fiber app with security configuration
	app := fiber.New(fiber.Config{
		ServerHeader: "VIP-Panel",
		AppName:      "VIP Hosting Panel v2.0",
	})

	// Security middleware
	app.Use(helmet.New())
	app.Use(limiter.New(limiter.Config{
		Max:               1000,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	}))
	app.Use(logger.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"version": "2.0.0",
		})
	})

	// Get port from environment or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}