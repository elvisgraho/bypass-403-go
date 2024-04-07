package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func baseHandler(w http.ResponseWriter, r *http.Request) {
	// todo, secondary context / traverse mock
	if r.URL.Path != "/" {
		errorHandler(w, r, http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Welcome! Please visit /admin for more.")
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	// Check for bypass conditions
	if isBypassed(r) {
		// Bypass condition is met, allow access

		// Check if the Cookie header is set
		cookie := r.Header.Get("Cookie")
		if cookie != "" {
			// Cookie header is set, print the cookie value
			log.Printf("User cookie is: %s", cookie)
		}

		// Cookie header is not set, print the sensitive data
		data := getSensitiveData(r)
		fmt.Fprintf(w, "Welcome, admin! Here's your sensitive data: %s", data)
		return
	}

	if r.Header.Get("X-Forward-For") == "169.254.169.254" {
		// handle different error lengths
		http.Error(w, "You are unauthorized!", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("X-Forward-For") == "172.16.0.1" {
		// handle different error lengths
		http.Error(w, "You are not allowed!", http.StatusForbidden)
		return
	}

	http.Error(w, "Unauthorized", http.StatusForbidden)
}

func bypassHandler(w http.ResponseWriter, r *http.Request) {
	data := getSensitiveData(r)
	fmt.Fprintf(w, "Welcome, admin! Here's your sensitive data: %s", data)
}

func errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "404")
	}
}

func getSensitiveData(r *http.Request) string {
	// Querying a database or external service.
	// For simplicity, let's assume the data is specified in a query parameter "data".
	// Note: This is an insecure direct object reference vulnerability.
	data := r.URL.Query().Get("data")
	return data
}

// Check for bypass conditions
func isBypassed(r *http.Request) bool {
	// mock bypasses to test the implementation

	if r.Method == http.MethodPut {
		// Bypass if method is not GET
		return true
	}

	if r.Header.Get("Cluster-Client-IP") == "localhost" {
		// Bypass if Cluster-Client-IP header points to localhost
		return true
	}

	if r.Header.Get("X-Forwarded-Port") == "8080" {
		// Bypass if X-Forwarded-Port header points to 8080
		return true
	}

	if r.Header.Get("Accept") == "application/json" {
		// Bypass Accept
		return true
	}

	if r.Header.Get("User-Agent") == "Mozilla/5.0 URL-Spider" {
		// Agent
		return true
	}

	if r.Header.Get("X-Original-URL") == "/admin" {
		// Agent
		return true
	}

	if r.Header.Get("X-Original-URL") == "/admin/secret" {
		// Agent
		return true
	}

	return false
}

func main() {
	// Create a channel to listen for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Dynamically get the port from the Docker daemon
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not provided by Docker
	}

	// Start the HTTP server
	go func() {
		http.HandleFunc("/", baseHandler)
		http.HandleFunc("/admin", adminHandler)
		http.HandleFunc("/admin/.", bypassHandler)
		http.HandleFunc("/admin/secret", adminHandler)
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	fmt.Printf("Server is running on port %s\n", port)

	// Wait for OS signal
	<-sigCh

	// Gracefully shut down the HTTP server
	fmt.Println("Shutting down...")
	os.Exit(0)
}
