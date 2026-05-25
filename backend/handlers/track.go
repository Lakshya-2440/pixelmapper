package handlers

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
)

type trackPageData struct {
	PixelID     string
	Token       string
	Label       string
	RedirectURL string
}

func (a *App) Track(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		http.NotFound(w, r)
		return
	}

	var pixelID string
	var label, redirectURL sql.NullString
	err := a.DB.QueryRow(`SELECT pixel_id, label, redirect_url FROM tracking_links WHERE token = `+a.bind(1), token).Scan(&pixelID, &label, &redirectURL)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load tracking link")
		return
	}

	_, err = a.DB.Exec(
		`INSERT INTO tracked_events (token, pixel_id, uid, email, ip, user_agent) VALUES (`+a.bind(1)+`, `+a.bind(2)+`, `+a.bind(3)+`, `+a.bind(4)+`, `+a.bind(5)+`, `+a.bind(6)+`)`,
		token,
		pixelID,
		nullable(r.URL.Query().Get("uid")),
		nullable(r.URL.Query().Get("email")),
		clientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not record tracking event")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = trackTemplate.Execute(w, trackPageData{
		PixelID:     pixelID,
		Token:       token,
		Label:       optionalString(label),
		RedirectURL: optionalString(redirectURL),
	})
}

var trackTemplate = template.Must(template.New("track").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>PixelMapper</title>
  <script>
    !function(f,b,e,v,n,t,s)
    {if(f.fbq)return;n=f.fbq=function(){n.callMethod?
    n.callMethod.apply(n,arguments):n.queue.push(arguments)};
    if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
    n.queue=[];t=b.createElement(e);t.async=!0;
    t.src=v;s=b.getElementsByTagName(e)[0];
    s.parentNode.insertBefore(t,s)}(window, document,'script',
    'https://connect.facebook.net/en_US/fbevents.js');
    fbq('init', '{{ .PixelID }}');
    fbq('track', 'PageView');
  </script>
  <noscript>
    <img height="1" width="1" style="display:none" alt=""
      src="https://www.facebook.com/tr?id={{ .PixelID }}&ev=PageView&noscript=1" />
  </noscript>
  {{ if .RedirectURL }}
  <meta http-equiv="refresh" content="2;url={{ .RedirectURL }}">
  {{ end }}
  <style>
    :root { color-scheme: light; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; }
    body { min-height: 100vh; margin: 0; display: grid; place-items: center; background: #f6f3ee; color: #18201f; }
    main { width: min(520px, calc(100vw - 32px)); text-align: center; }
    .mark { width: 52px; height: 52px; margin: 0 auto 18px; border-radius: 14px; display: grid; place-items: center; background: #132d2a; color: #f6f3ee; font-weight: 800; }
    h1 { margin: 0 0 10px; font-size: clamp(24px, 6vw, 36px); letter-spacing: 0; }
    p { margin: 0; color: #5f6764; line-height: 1.55; }
  </style>
</head>
<body>
  <main>
    <div class="mark">PM</div>
    <h1>{{ if .Label }}{{ .Label }}{{ else }}Visit recorded{{ end }}</h1>
    <p>{{ if .RedirectURL }}Taking you to destination...{{ else }}Meta Pixel PageView fired for this tracking link.{{ end }}</p>
  </main>
</body>
</html>`))
