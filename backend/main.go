package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pixelmapper/backend/handlers"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed db/schema.sql db/schema_postgres.sql
var schemaFS embed.FS

func main() {
	db, driver, err := openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db, driver); err != nil {
		log.Fatal(err)
	}

	app := handlers.NewApp(db, driver)
	mux := http.NewServeMux()
	app.Register(mux)
	mux.Handle("/", spaHandler())

	port := getenv("PORT", "8080")
	log.Printf("PixelMapper listening on :%s", port)
	if err := http.ListenAndServe(":"+port, withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

func openDB() (*sql.DB, string, error) {
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		db, err := sql.Open("pgx", databaseURL)
		if err != nil {
			return nil, "", err
		}
		return db, "postgres", db.Ping()
	}

	path := getenv("DATABASE_PATH", filepath.Join(".", "pixelmapper.db"))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return nil, "", err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, "", err
	}
	db.SetMaxOpenConns(1)
	return db, "sqlite", nil
}

func migrate(db *sql.DB, driver string) error {
	schemaPath := "db/schema.sql"
	if driver == "postgres" {
		schemaPath = "db/schema_postgres.sql"
	}
	schema, err := schemaFS.ReadFile(schemaPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(schema))
	return err
}

func spaHandler() http.Handler {
	dist := getenv("FRONTEND_DIST", "../frontend/dist")
	fileServer := http.FileServer(http.Dir(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/t/") || r.URL.Path == "/healthz" {
			http.NotFound(w, r)
			return
		}
		if _, err := os.Stat(filepath.Join(dist, filepath.Clean(r.URL.Path))); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}
		index, err := fs.ReadFile(os.DirFS(dist), "index.html")
		if err != nil {
			http.Error(w, "frontend build not found; run npm run build in frontend", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(index)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
