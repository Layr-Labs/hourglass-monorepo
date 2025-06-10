package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Hostname  string            `json:"hostname"`
	Env       map[string]string `json:"env,omitempty"`
}

type InfoResponse struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Port        int    `json:"port"`
}

var startTime = time.Now()

func main() {
	port := 8080
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		}
	}

	hostname, _ := os.Hostname()

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		env := make(map[string]string)
		for _, e := range os.Environ() {
			if len(e) > 0 && e[0] != '_' { // Skip internal vars
				parts := strings.SplitN(e, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					value := parts[1]
					// Truncate value for brevity
					if len(value) > 50 {
						value = value[:50] + "..."
					}
					env[key] = value
				}
			}
		}

		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Uptime:    time.Since(startTime).String(),
			Hostname:  hostname,
			Env:       env,
		}

		_ = json.NewEncoder(w).Encode(response)
	})

	// Info endpoint
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := InfoResponse{
			Name:        "ponos-test-container",
			Version:     "1.0.0",
			Description: "Test container for ponos containerManager integration tests",
			Port:        port,
		}

		_ = json.NewEncoder(w).Encode(response)
	})

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := map[string]interface{}{
			"message":   "ponos test container is running",
			"timestamp": time.Now(),
			"uptime":    time.Since(startTime).String(),
			"hostname":  hostname,
			"endpoints": []string{"/", "/health", "/info"},
		}

		_ = json.NewEncoder(w).Encode(response)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Test container starting on %s", addr)
	log.Printf("Endpoints available: /, /health, /info")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
