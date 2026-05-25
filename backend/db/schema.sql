CREATE TABLE IF NOT EXISTS tracking_links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token TEXT UNIQUE NOT NULL,
  pixel_id TEXT NOT NULL,
  label TEXT,
  redirect_url TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tracked_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  token TEXT NOT NULL,
  pixel_id TEXT NOT NULL,
  uid TEXT,
  email TEXT,
  ip TEXT,
  user_agent TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tracked_events_token ON tracked_events(token);
CREATE INDEX IF NOT EXISTS idx_tracked_events_pixel_id ON tracked_events(pixel_id);
CREATE INDEX IF NOT EXISTS idx_tracked_events_created_at ON tracked_events(created_at);
