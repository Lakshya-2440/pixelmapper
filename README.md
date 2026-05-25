# PixelMapper

Meta Pixel user tracking and ad targeting MVP.

## What ships

- Go backend with SQLite persistence
- React/Vite dashboard
- Tracking link creation
- `GET /t/{token}` logs visit signals and serves Meta Pixel PageView HTML
- Dashboard filters by Pixel ID and date range
- Copyable tracking links
- Docker and Render blueprint

## Local run

```bash
cd backend
go mod tidy
go run .
```

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:5173`.

## Single server production run

```bash
cd frontend
npm install
npm run build

cd ../backend
go mod tidy
FRONTEND_DIST=../frontend/dist DATABASE_PATH=./pixelmapper.db go run .
```

Open `http://localhost:8080`.

## API

`POST /api/links`

```json
{
  "pixel_id": "123456789012345",
  "label": "Summer Campaign",
  "redirect_url": "https://mybrand.com"
}
```

`GET /api/links`

`GET /api/events?pixel_id=123456789012345&from=2026-05-25&to=2026-05-25`

`GET /t/{token}?uid=user_42&email=hashed@example`

## Deploy on Render

1. Push this repo to GitHub.
2. Create Render Blueprint from `render.yaml`.
3. Set `PUBLIC_BASE_URL` to final Render URL after first deploy if generated tracking URLs need canonical domain.

## Notes

- SQLite is used for MVP simplicity. Render disk keeps data persistent.
- Meta Pixel fires client-side from tracking page via standard `fbq('init', pixel_id)` and `fbq('track', 'PageView')`.
- No auth is included per MVP scope. Add admin auth before real customer usage.
