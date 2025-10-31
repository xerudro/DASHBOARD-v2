package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"version":   "v2.0.0",
		}
		json.NewEncoder(w).Encode(response)
	})

	// Auth login endpoint
	http.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
			return
		}

		var loginData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&loginData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid request body",
			})
			return
		}

		email, _ := loginData["email"].(string)
		password, _ := loginData["password"].(string)

		// Input validation
		if email == "" || password == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Email and password are required",
			})
			return
		}

		// Check for malicious patterns (basic XSS/SQL injection protection)
		if strings.Contains(email, "<script") || strings.Contains(password, "<script") ||
		   strings.Contains(email, "' OR '") || strings.Contains(email, "'; DROP") {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Invalid input detected",
			})
			return
		}

		// Mock successful response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Login successful",
			"token":   "mock-jwt-token",
		})
	})

	// Dashboard endpoint (protected)
	http.HandleFunc("/api/v1/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
			return
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Dashboard data",
			"stats": map[string]int{
				"servers": 5,
				"sites":   12,
				"users":   3,
			},
		})
	})

	// Catch-all 404
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Not found"})
			return
		}
		
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>VIP Hosting Panel v2</title>
</head>
<body>
    <h1>🚀 VIP Hosting Panel v2 - Security Test Server</h1>
    <p>✅ Server is running and ready for security testing!</p>
    <ul>
        <li><a href="/health">Health Check</a></li>
        <li><a href="/api/v1/dashboard">Dashboard API</a> (requires auth)</li>
    </ul>
</body>
</html>`)
	})

	fmt.Println("🚀 Starting VIP Hosting Panel v2 Security Test Server")
	fmt.Println("📊 Health check: http://localhost:8080/health")
	fmt.Println("🔒 Ready for security testing!")
	fmt.Println("🌐 Server running on: http://localhost:8080")
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}