CREATE TABLE IF NOT EXISTS tracking_links (
  id BIGSERIAL PRIMARY KEY,
  token VARCHAR(64) UNIQUE NOT NULL,
  pixel_id VARCHAR(64) NOT NULL,
  label TEXT,
  redirect_url TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tracked_events (
  id BIGSERIAL PRIMARY KEY,
  token VARCHAR(64) NOT NULL,
  pixel_id VARCHAR(64) NOT NULL,
  uid TEXT,
  email TEXT,
  ip TEXT,
  user_agent TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tracked_events_token ON tracked_events(token);
CREATE INDEX IF NOT EXISTS idx_tracked_events_pixel_id ON tracked_events(pixel_id);
CREATE INDEX IF NOT EXISTS idx_tracked_events_created_at ON tracked_events(created_at);
