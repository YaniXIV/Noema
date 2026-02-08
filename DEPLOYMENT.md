# Deployment guide (Render recommended)

This guide targets a public demo deployment with an open verification page and persistent storage for proofs.

## Render (recommended)

1. Create a **Web Service** from this repo.
2. Set **Root Directory** to `backend` (or keep root and use `cd backend` in the build command).
3. Build command:
   - `go build -o /tmp/noema ./cmd/server`
4. Start command:
   - `/tmp/noema`
5. Add environment variables:
   - `JUDGE_KEY`
   - `GEMINI_API_KEY`
   - `NOEMA_COOKIE_SECRET`
   - `NOEMA_SECURE_COOKIES=1`
   - `GEMINI_MODEL` (optional)
   - `NOEMA_RUNS_DIR=/var/data/runs`
   - `NOEMA_UPLOADS_DIR=/var/data/uploads`
   - `NOEMA_RUNS_MAX=50` (set `0` to disable pruning)
6. Add a **Persistent Disk** mounted at `/var/data`.
7. Health check path: `/ready` (or `/health`).

Notes:
- Render will set `PORT`; the server reads it automatically.
- Public routes (`/`, `/verify`, `/verify/:id`, `/api/verify`) are not gated by the judge key.

## Storage plan

- Run artifacts and uploads should live on a persistent disk (see `NOEMA_RUNS_DIR` and `NOEMA_UPLOADS_DIR`).
- The server prunes old run directories to keep the most recent `NOEMA_RUNS_MAX` runs. This caps the public verify list to a recent window.

## Domain + HTTPS

- Add a custom domain and HTTPS at the platform layer.
- Keep `NOEMA_SECURE_COOKIES=1` so cookies are marked `Secure` and only sent over HTTPS.
