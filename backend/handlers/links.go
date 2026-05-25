package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"pixelmapper/backend/models"
)

func (a *App) CreateLink(w http.ResponseWriter, r *http.Request) {
	var req models.CreateLinkRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	req.PixelID = strings.TrimSpace(req.PixelID)
	req.Label = strings.TrimSpace(req.Label)
	req.RedirectURL = strings.TrimSpace(req.RedirectURL)

	if err := validatePixelID(req.PixelID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.RedirectURL != "" {
		parsed, err := url.ParseRequestURI(req.RedirectURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			writeError(w, http.StatusBadRequest, "redirect_url must be a valid absolute URL")
			return
		}
	}

	token, err := a.uniqueToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate token")
		return
	}

	_, err = a.DB.Exec(
		`INSERT INTO tracking_links (token, pixel_id, label, redirect_url) VALUES (`+a.bind(1)+`, `+a.bind(2)+`, `+a.bind(3)+`, `+a.bind(4)+`)`,
		token, req.PixelID, nullable(req.Label), nullable(req.RedirectURL),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create link")
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateLinkResponse{
		Token:       token,
		TrackingURL: baseURL(r) + "/t/" + token,
	})
}

func (a *App) ListLinks(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(`
		SELECT l.id, l.token, l.pixel_id, l.label, l.redirect_url, l.created_at, COALESCE(ec.event_count, 0) AS event_count
		FROM tracking_links l
		LEFT JOIN (
			SELECT token, COUNT(*) AS event_count
			FROM tracked_events
			GROUP BY token
		) ec ON ec.token = l.token
		ORDER BY l.created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch links")
		return
	}
	defer rows.Close()

	links := make([]models.TrackingLink, 0)
	for rows.Next() {
		var link models.TrackingLink
		var label, redirectURL sql.NullString
		if err := rows.Scan(&link.ID, &link.Token, &link.PixelID, &label, &redirectURL, &link.CreatedAt, &link.EventCount); err != nil {
			writeError(w, http.StatusInternalServerError, "could not scan links")
			return
		}
		link.Label = optionalString(label)
		link.RedirectURL = optionalString(redirectURL)
		link.TrackingURL = baseURL(r) + "/t/" + link.Token
		links = append(links, link)
	}

	writeJSON(w, http.StatusOK, links)
}

func (a *App) uniqueToken() (string, error) {
	for range 8 {
		token, err := randomToken()
		if err != nil {
			return "", err
		}
		var exists int
		err = a.DB.QueryRow(`SELECT 1 FROM tracking_links WHERE token = `+a.bind(1), token).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return token, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", errors.New("could not generate unique token")
}

func randomToken() (string, error) {
	buf := make([]byte, 9)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(buf), "="), nil
}

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
