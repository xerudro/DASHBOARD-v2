package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xerudro/DASHBOARD-v2/internal/models"
	"github.com/xerudro/DASHBOARD-v2/web/templates/pages"
)

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		ServerHeader: "HostPanel-Preview",
		AppName:      "HostPanel Preview v1.0",
	})

	app.Use(logger.New())

	// Serve static files
	app.Static("/static", "./web/static")

	// Dashboard preview route
	app.Get("/", func(c *fiber.Ctx) error {
		// Mock user data
		user := &models.User{
			Name:  "Romeo Alexandru Neacsu",
			Email: "admin@hostpanel.com",
			Role:  "Admin",
		}

		// Mock dashboard data
		data := pages.DashboardData{
			VPSServers:     2,
			VPSActive:      2,
			Clients:        0,
			Domains:        1,
			ActiveServices: 0,
			SystemStatus: pages.SystemStatus{
				VPSServices: true,
				Database:    true,
				Security:    true,
			},
		}

		// Render the dashboard
		c.Set("Content-Type", "text/html; charset=utf-8")
		return pages.Dashboard(user, data).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		// Mock user data
		user := &models.User{
			Name:  "Romeo Alexandru Neacsu",
			Email: "admin@hostpanel.com",
			Role:  "Admin",
		}

		// Mock dashboard data
		data := pages.DashboardData{
			VPSServers:     2,
			VPSActive:      2,
			Clients:        0,
			Domains:        1,
			ActiveServices: 0,
			SystemStatus: pages.SystemStatus{
				VPSServices: true,
				Database:    true,
				Security:    true,
			},
		}

		// Render dashboard page
		c.Set("Content-Type", "text/html; charset=utf-8")
		return pages.Dashboard(user, data).Render(c.Context(), c.Response().BodyWriter())
	})

	app.Get("/servers", func(c *fiber.Ctx) error {
		// Mock user data
		user := &models.User{
			Name:  "Romeo Alexandru Neacsu",
			Email: "admin@hostpanel.com",
			Role:  "Admin",
		}

		// Mock servers data
		data := pages.ServersData{
			TotalServers:       0,
			OnlineServers:      0,
			OfflineServers:     0,
			MaintenanceServers: 0,
			Servers:            []pages.ServerItem{},
			ServerRoles: pages.ServerRolesData{
				WebServers:    0,
				Databases:     0,
				MailServers:   0,
				BackupServers: 0,
			},
		}

		// Render servers page
		c.Set("Content-Type", "text/html; charset=utf-8")
		return pages.Servers(user, data).Render(c.Context(), c.Response().BodyWriter())
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"port":   5501,
		})
	})

	// Start server on port 5502
	log.Println("🚀 Preview server starting on http://localhost:5502")
	log.Println("📱 Dashboard: http://localhost:5502")
	if err := app.Listen(":5502"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
