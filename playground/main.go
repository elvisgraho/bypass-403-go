package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Attempt struct {
	ID           int
	Time         time.Time
	Method       string
	Host         string
	Path         string
	RequestURI   string
	RemoteAddr   string
	StatusCode   int
	UserAgent    string
	BypassReason string
	Outcome      string
	Route        string
	Headers      map[string]string
}

type AttemptStore struct {
	mu       sync.RWMutex
	attempts []Attempt
	nextID   int
}

func (s *AttemptStore) Add(attempt Attempt) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	attempt.ID = s.nextID
	s.attempts = append(s.attempts, attempt)
}

func (s *AttemptStore) Snapshot() []Attempt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Attempt, len(s.attempts))
	for i := range s.attempts {
		out[i] = s.attempts[len(s.attempts)-1-i]
	}
	return out
}

func (s *AttemptStore) SinceID(sinceID int) []Attempt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.attempts) == 0 {
		return nil
	}

	startIndex := len(s.attempts)
	for i := len(s.attempts) - 1; i >= 0; i-- {
		if s.attempts[i].ID <= sinceID {
			startIndex = i + 1
			break
		}
		if i == 0 {
			startIndex = 0
		}
	}

	if startIndex >= len(s.attempts) {
		return nil
	}

	out := make([]Attempt, len(s.attempts)-startIndex)
	copy(out, s.attempts[startIndex:])
	return out
}

func (s *AttemptStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.attempts = nil
	s.nextID = 0
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type app struct {
	store *AttemptStore
}

func main() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := &app{store: &AttemptStore{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.baseHandler)
	mux.HandleFunc("/admin", app.adminHandler)
	mux.HandleFunc("/admin/.", app.bypassHandler)
	mux.HandleFunc("/admin/secret", app.adminHandler)
	mux.HandleFunc("/dashboard", app.dashboardHandler)
	mux.HandleFunc("/api/attempts", app.attemptsAPIHandler)
	mux.HandleFunc("/api/attempts/clear", app.clearAttemptsHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: app.loggingMiddleware(mux),
	}

	go func() {
		log.Printf("Playground server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("HTTP server error:", err)
		}
	}()

	<-sigCh
	log.Println("Shutting down playground server")
	os.Exit(0)
}

func (a *app) baseHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		a.errorHandler(w, r, http.StatusNotFound)
		return
	}

	fmt.Fprint(w, "Welcome! Visit /admin for the protected route or /dashboard to inspect traffic.")
}

func (a *app) adminHandler(w http.ResponseWriter, r *http.Request) {
	if bypassReason := getBypassReason(r); bypassReason != "" {
		if cookie := r.Header.Get("Cookie"); cookie != "" {
			log.Printf("User cookie is: %s", cookie)
		}

		data := getSensitiveData(r)
		fmt.Fprintf(w, "Welcome, admin! Bypass accepted via %s. Sensitive data: %s", bypassReason, data)
		return
	}

	if r.Header.Get("X-Forward-For") == "169.254.169.254" {
		http.Error(w, "You are unauthorized!", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("X-Forward-For") == "172.16.0.1" {
		http.Error(w, "You are not allowed!", http.StatusForbidden)
		return
	}

	http.Error(w, "Unauthorized", http.StatusForbidden)
}

func (a *app) bypassHandler(w http.ResponseWriter, r *http.Request) {
	data := getSensitiveData(r)
	fmt.Fprintf(w, "Welcome, admin! Path-specific bypass hit. Sensitive data: %s", data)
}

func (a *app) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTmpl.Execute(w, dashboardView{Attempts: a.store.Snapshot()}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *app) attemptsAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var attempts []Attempt
	sinceID := 0
	if rawSinceID := r.URL.Query().Get("since_id"); rawSinceID != "" {
		parsedSinceID, err := strconv.Atoi(rawSinceID)
		if err != nil {
			http.Error(w, "invalid since_id", http.StatusBadRequest)
			return
		}
		sinceID = parsedSinceID
	}

	if sinceID > 0 {
		attempts = a.store.SinceID(sinceID)
	} else {
		attempts = a.store.Snapshot()
	}

	if err := json.NewEncoder(w).Encode(attempts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *app) clearAttemptsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a.store.Clear()
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) errorHandler(w http.ResponseWriter, r *http.Request, status int) {
	w.WriteHeader(status)
	if status == http.StatusNotFound {
		fmt.Fprint(w, "404")
	}
}

func getSensitiveData(r *http.Request) string {
	return r.URL.Query().Get("data")
}

func getBypassReason(r *http.Request) string {
	switch {
	case r.Method == http.MethodPut:
		return "method-put"
	case r.Header.Get("Cluster-Client-IP") == "localhost":
		return "cluster-client-ip"
	case r.Header.Get("X-Forwarded-Port") == "8080":
		return "x-forwarded-port"
	case r.Header.Get("Accept") == "application/json":
		return "accept-json"
	case r.Header.Get("User-Agent") == "Mozilla/5.0 URL-Spider":
		return "legacy-user-agent"
	case r.Header.Get("X-Original-URL") == "/admin":
		return "x-original-url-admin"
	case r.Header.Get("X-Original-URL") == "/admin/secret":
		return "x-original-url-secret"
	default:
		return ""
	}
}

func (a *app) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newResponseRecorder(w)
		next.ServeHTTP(recorder, r)

		if r.URL.Path == "/dashboard" || r.URL.Path == "/api/attempts" || r.URL.Path == "/api/attempts/clear" {
			return
		}

		route := resolveRoute(r)
		bypassReason := resolveBypassReason(r, recorder.statusCode)
		outcome := resolveOutcome(r, recorder.statusCode, bypassReason)

		a.store.Add(Attempt{
			Time:         time.Now(),
			Method:       r.Method,
			Host:         r.Host,
			Path:         r.URL.Path,
			RequestURI:   r.RequestURI,
			RemoteAddr:   r.RemoteAddr,
			StatusCode:   recorder.statusCode,
			UserAgent:    r.UserAgent(),
			BypassReason: bypassReason,
			Outcome:      outcome,
			Route:        route,
			Headers:      collectInterestingHeaders(r),
		})
	})
}

func resolveRoute(r *http.Request) string {
	switch r.URL.Path {
	case "/admin/.":
		return "path-bypass"
	default:
		return r.URL.Path
	}
}

func resolveBypassReason(r *http.Request, statusCode int) string {
	if r.URL.Path == "/admin/." && statusCode >= 200 && statusCode < 400 {
		return "path-bypass"
	}

	if bypassReason := getBypassReason(r); bypassReason != "" {
		return bypassReason
	}

	if strings.HasPrefix(r.URL.Path, "/admin") && statusCode >= 200 && statusCode < 400 {
		return "allowed"
	}

	return ""
}

func resolveOutcome(r *http.Request, statusCode int, bypassReason string) string {
	if bypassReason != "" {
		return bypassReason
	}

	if statusCode >= 200 && statusCode < 400 {
		if strings.HasPrefix(r.URL.Path, "/admin") {
			return "success"
		}
		return "ok"
	}

	return "blocked"
}

func collectInterestingHeaders(r *http.Request) map[string]string {
	headers := map[string]string{}
	if r.Host != "" {
		headers["Host"] = r.Host
	}
	for _, key := range []string{
		"Accept",
		"Cluster-Client-IP",
		"Connection",
		"Cookie",
		"User-Agent",
		"X-Forward-For",
		"X-Forwarded-For",
		"X-Forwarded-Port",
		"X-Original-URL",
		"X-Rewrite-URL",
	} {
		if value := r.Header.Get(key); value != "" {
			headers[key] = value
		}
	}
	return headers
}
