//go:build simple

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ServerHeader: "VIP-Hosting-Panel",
		AppName:      "VIP Hosting Panel v2",
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,https://localhost:3000",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "v2.0.0",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (mock for testing)
	api.Post("/auth/login", func(c *fiber.Ctx) error {
		var loginData map[string]interface{}
		if err := c.BodyParser(&loginData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
			})
		}

		email, emailOk := loginData["email"].(string)
		password, passwordOk := loginData["password"].(string)

		// Basic validation
		if !emailOk || !passwordOk || email == "" || password == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Email and password are required",
			})
		}

		// Mock authentication
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Login successful",
			"token":   "mock-jwt-token",
		})
	})

	api.Post("/auth/register", func(c *fiber.Ctx) error {
		var registerData map[string]interface{}
		if err := c.BodyParser(&registerData); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request body",
			})
		}

		email, emailOk := registerData["email"].(string)
		password, passwordOk := registerData["password"].(string)

		// Basic validation
		if !emailOk || !passwordOk || email == "" || password == "" {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Email and password are required",
			})
		}

		// Mock registration
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Registration successful",
		})
	})

	// Protected routes (mock)
	api.Get("/dashboard", func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		
		return c.JSON(fiber.Map{
			"message": "Dashboard data",
			"stats": fiber.Map{
				"servers": 5,
				"sites":   12,
				"users":   3,
			},
		})
	})

	api.Get("/servers", func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		
		return c.JSON(fiber.Map{
			"servers": []fiber.Map{
				{"id": 1, "name": "Server 1", "status": "running"},
				{"id": 2, "name": "Server 2", "status": "stopped"},
			},
		})
	})

	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error": "Not found",
		})
	})

	// Start server
	fmt.Println("🚀 Starting VIP Hosting Panel v2 on :8080")
	fmt.Println("📊 Health check: http://localhost:8080/health")
	fmt.Println("🔒 Security tests ready to run!")
	
	log.Fatal(app.Listen(":8080"))
}