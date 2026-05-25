package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type App struct {
	DB     *sql.DB
	Driver string
}

func NewApp(db *sql.DB, driver string) *App {
	return &App{DB: db, Driver: driver}
}

func (a *App) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", a.Health)
	mux.HandleFunc("POST /api/links", a.CreateLink)
	mux.HandleFunc("GET /api/links", a.ListLinks)
	mux.HandleFunc("GET /api/events", a.ListEvents)
	mux.HandleFunc("GET /api/profiles", a.ListProfiles)
	mux.HandleFunc("POST /api/profiles/{id}/sync", a.SyncProfile)
	mux.HandleFunc("POST /api/profiles/sync_all", a.SyncAllProfiles)
	mux.HandleFunc("GET /t/{token}", a.Track)
}

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	return nil
}

func baseURL(r *http.Request) string {
	if configured := strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/"); configured != "" {
		return configured
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	return scheme + "://" + host
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"CF-Connecting-IP", "Fly-Client-IP", "X-Real-IP"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			return value
		}
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func optionalString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func (a *App) bind(index int) string {
	if a.Driver == "postgres" {
		return "$" + strconv.Itoa(index)
	}
	return "?"
}

func validatePixelID(pixelID string) error {
	pixelID = strings.TrimSpace(pixelID)
	if pixelID == "" {
		return errors.New("pixel_id is required")
	}
	if len(pixelID) > 64 {
		return errors.New("pixel_id must be 64 characters or fewer")
	}
	for _, ch := range pixelID {
		if ch < '0' || ch > '9' {
			return errors.New("pixel_id must contain digits only")
		}
	}
	return nil
}
