package handlers

import (
    "bytes"
    "crypto/sha256"
    "database/sql"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "pixelmapper/backend/models"
)

func (a *App) ListProfiles(w http.ResponseWriter, r *http.Request) {
    rows, err := a.DB.Query(`SELECT id, uid, email, created_at, last_seen, last_synced FROM user_profiles ORDER BY last_seen DESC LIMIT 500`)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "could not fetch profiles")
        return
    }
    defer rows.Close()

    profiles := make([]models.UserProfile, 0)
    for rows.Next() {
        var p models.UserProfile
        var lastSynced sql.NullTime
        if err := rows.Scan(&p.ID, &p.UID, &p.Email, &p.CreatedAt, &p.LastSeen, &lastSynced); err != nil {
            writeError(w, http.StatusInternalServerError, "could not scan profiles")
            return
        }
        if lastSynced.Valid {
            p.LastSynced = lastSynced.Time
        }
        profiles = append(profiles, p)
    }

    writeJSON(w, http.StatusOK, profiles)
}

func (a *App) SyncProfile(w http.ResponseWriter, r *http.Request) {
    // parse id from path
    id := r.PathValue("id")
    if id == "" {
        writeError(w, http.StatusBadRequest, "missing profile id")
        return
    }

    var p models.UserProfile
    row := a.DB.QueryRow(`SELECT id, uid, email FROM user_profiles WHERE id = `+a.bind(1), id)
    if err := row.Scan(&p.ID, &p.UID, &p.Email); err != nil {
        if err == sql.ErrNoRows {
            writeError(w, http.StatusNotFound, "profile not found")
            return
        }
        writeError(w, http.StatusInternalServerError, "could not load profile")
        return
    }

    ok, err := a.sendToFacebook(p)
    if err != nil {
        writeError(w, http.StatusInternalServerError, fmt.Sprintf("sync failed: %v", err))
        return
    }
    if ok {
        _, _ = a.DB.Exec(`UPDATE user_profiles SET last_synced = CURRENT_TIMESTAMP WHERE id = `+a.bind(1), p.ID)
    }
    writeJSON(w, http.StatusOK, map[string]any{"synced": ok})
}

func (a *App) SyncAllProfiles(w http.ResponseWriter, r *http.Request) {
    rows, err := a.DB.Query(`SELECT id, uid, email FROM user_profiles ORDER BY last_seen DESC LIMIT 500`)
    if err != nil {
        writeError(w, http.StatusInternalServerError, "could not fetch profiles")
        return
    }
    defer rows.Close()

    results := make(map[int64]bool)
    for rows.Next() {
        var p models.UserProfile
        if err := rows.Scan(&p.ID, &p.UID, &p.Email); err != nil {
            continue
        }
        ok, _ := a.sendToFacebook(p)
        if ok {
            _, _ = a.DB.Exec(`UPDATE user_profiles SET last_synced = CURRENT_TIMESTAMP WHERE id = `+a.bind(1), p.ID)
        }
        results[p.ID] = ok
    }
    writeJSON(w, http.StatusOK, results)
}

func (a *App) sendToFacebook(p models.UserProfile) (bool, error) {
    token := strings.TrimSpace(os.Getenv("FACEBOOK_ACCESS_TOKEN"))
    pixel := strings.TrimSpace(os.Getenv("FACEBOOK_PIXEL_ID"))
    if token == "" || pixel == "" {
        return false, fmt.Errorf("FACEBOOK_ACCESS_TOKEN or FACEBOOK_PIXEL_ID not configured")
    }

    // prepare user_data with hashed email if available
    userData := map[string]any{}
    if p.Email != "" {
        normalized := strings.TrimSpace(strings.ToLower(p.Email))
        h := sha256.Sum256([]byte(normalized))
        userData["em"] = hex.EncodeToString(h[:])
    }
    if p.UID != "" {
        userData["external_id"] = p.UID
    }

    if len(userData) == 0 {
        return false, fmt.Errorf("no identifiable user data to send")
    }

    event := map[string]any{
        "event_name":  "PageView",
        "event_time":  time.Now().Unix(),
        "user_data":   userData,
        "action_source": "website",
    }
    body := map[string]any{"data": []any{event}}

    b, _ := json.Marshal(body)
    url := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/events?access_token=%s", pixel, token)
    req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return true, nil
    }
    return false, fmt.Errorf("facebook API status %d", resp.StatusCode)
}
