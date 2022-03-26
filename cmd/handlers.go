package main

import (
	"context"
	"encoding/json"
	"net/http"
)

// wrap is a middleware that wraps HTTP handlers and injects the "app" context.
func wrap(app *App, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "app", app)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// resp is used to send uniform response structure.
type resp struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// sendResponse sends a JSON envelope to the HTTP response.
func sendResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	out, err := json.Marshal(resp{Status: "success", Data: data})

	if err != nil {
		sendErrorResponse(w, "Internal Server Error.", http.StatusInternalServerError, nil)
		return
	}

	w.Write(out)
}

// sendErrorResponse sends a JSON error envelope to the HTTP response.
func sendErrorResponse(w http.ResponseWriter, message string, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	resp := resp{Status: "error",
		Message: message,
		Data:    data}
	out, _ := json.Marshal(resp)

	w.Write(out)
}

// Index page.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, "welcome to ebs-exporter!")
}

// Health check.
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, "pong")
}

// Export prometheus metrics.
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	var (
		app = r.Context().Value("app").(*App)
	)
	app.metrics.FlushMetrics(w)
}
