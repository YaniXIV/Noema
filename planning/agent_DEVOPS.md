# Agent: DEVOPS

Goal: Provide deployment plan for hosted Gin app with public verification access.

Read first:
- planning.md
- README.md

Testing requirement:
- Add or update any deployment checks if applicable.
- Run `go test ./...` only if you changed Go files (otherwise skip).
- If you can’t run tests, document why.

File ownership (allowed to edit):
- `DEPLOYMENT.md`
- `backend/.env.example`
- `backend/cmd/server/main.go` (only for PORT/health endpoints)
- `backend/internal/config/*` (only for env var wiring)

Do not edit:
- `backend/internal/evaluate/*`
- `backend/internal/zk/*`
- `backend/internal/verify/*`
- `backend/web/*`
- `README.md` (DOCS agent owns)
- `planning/*`

Tasks (do in order):
1) Hosting plan
- Recommend a deployment target (Render/Fly/GCP) suitable for a Go/Gin app.
- Provide steps to deploy with env vars:
  - JUDGE_KEY
  - GEMINI_API_KEY
  - NOEMA_COOKIE_SECRET
  - NOEMA_SECURE_COOKIES
  - optional GEMINI_MODEL
- Ensure public routes (`/`, `/verify`) are accessible without login.

2) Storage plan
- Decide where run artifacts live in hosted environment.
- Provide guidance for retaining recent N proofs for public verify list.

3) Basic ops
- Add health checks.
- Add guidance on domain + HTTPS (Secure cookies).

Deliverables
- A short deployment guide in markdown.
- Any needed config changes.

Questions to resolve (ask PLANNER if blocked):
- Preferred hosting provider.
