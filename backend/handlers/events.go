package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"pixelmapper/backend/models"
)

func (a *App) ListEvents(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT e.id, e.token, e.pixel_id, l.label, e.uid, e.email, e.ip, e.user_agent, e.created_at
		FROM tracked_events e
		LEFT JOIN tracking_links l ON l.token = e.token
		WHERE 1 = 1
	`
	args := make([]any, 0)

	if pixelID := strings.TrimSpace(r.URL.Query().Get("pixel_id")); pixelID != "" {
		query += " AND e.pixel_id = ?"
		args = append(args, pixelID)
	}
	if from := strings.TrimSpace(r.URL.Query().Get("from")); from != "" {
		query += " AND e.created_at >= ?"
		args = append(args, from)
	}
	if to := strings.TrimSpace(r.URL.Query().Get("to")); to != "" {
		query += " AND e.created_at <= ?"
		args = append(args, to+" 23:59:59")
	}
	query += " ORDER BY e.created_at DESC LIMIT 500"

	rows, err := a.DB.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not fetch events")
		return
	}
	defer rows.Close()

	events := make([]models.TrackedEvent, 0)
	for rows.Next() {
		var event models.TrackedEvent
		var label, uid, email, ip, userAgent sql.NullString
		if err := rows.Scan(&event.ID, &event.Token, &event.PixelID, &label, &uid, &email, &ip, &userAgent, &event.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "could not scan events")
			return
		}
		event.Label = optionalString(label)
		event.UID = optionalString(uid)
		event.Email = optionalString(email)
		event.IP = optionalString(ip)
		event.UserAgent = optionalString(userAgent)
		events = append(events, event)
	}

	writeJSON(w, http.StatusOK, events)
}
