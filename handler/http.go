package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// writeJSON answers with v encoded as JSON and the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// decodeJSON reads the request body into v. On failure it answers 400 and
// reports false, so the caller only has to return.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	return true
}

// pathID reads a positive integer from the URL (e.g. {id} in /courses/{id}).
// On failure it answers 400 with msg and reports false.
func pathID(w http.ResponseWriter, r *http.Request, name, msg string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue(name))
	if err != nil || id <= 0 {
		http.Error(w, msg, http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// internalError logs the real cause and answers 500 without leaking it.
func internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// CORSMiddleware lets the frontend (served from another origin in dev) call
// the API, and answers preflight requests.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
