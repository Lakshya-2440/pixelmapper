package models

import "time"

type TrackingLink struct {
	ID          int64     `json:"id"`
	Token       string    `json:"token"`
	PixelID     string    `json:"pixel_id"`
	Label       string    `json:"label"`
	RedirectURL string    `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
	EventCount  int64     `json:"event_count,omitempty"`
	TrackingURL string    `json:"tracking_url,omitempty"`
}

type TrackedEvent struct {
	ID        int64     `json:"id"`
	Token     string    `json:"token"`
	PixelID   string    `json:"pixel_id"`
	Label     string    `json:"label"`
	UID       string    `json:"uid"`
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateLinkRequest struct {
	PixelID     string `json:"pixel_id"`
	Label       string `json:"label"`
	RedirectURL string `json:"redirect_url"`
}

type CreateLinkResponse struct {
	Token       string `json:"token"`
	TrackingURL string `json:"tracking_url"`
}
