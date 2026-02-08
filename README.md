# Noema

Verifiable dataset compliance using Gemini 3 reasoning and zero-knowledge proofs.

## Demo
- Public demo: TODO
- 3‑minute video: TODO

## Architecture
```mermaid
graph TD
  A[Dataset + optional images] --> B[Gemini 3 policy evaluation]
  B --> C[Constraint severities + rationale]
  C --> D[Policy aggregation]
  D --> E[ZK proof generation (gnark)]
  E --> F[Public verification]
```

## Quickstart
From `backend/`:
```bash
cd backend
cp .env.example .env
# Edit .env: set JUDGE_KEY, optionally GEMINI_API_KEY and NOEMA_COOKIE_SECRET
go run ./cmd/server
```
Server listens on `:8080`. Run from `backend/` so `web/templates` and `web/static` are found.

### Env vars
| Variable | Required | Description |
|----------|----------|-------------|
| `JUDGE_KEY` | Yes (for gating) | Secret used to allow API and UI access. |
| `NOEMA_COOKIE_SECRET` | Recommended | Secret for signing session cookies. If unset, a dev default is used (and a warning is logged). |
| `GEMINI_API_KEY` | Optional | Enables Gemini 3 evaluation. |
| `NOEMA_SECURE_COOKIES` or `HTTPS=1` | Optional | Set to use `Secure` on cookies (e.g. behind HTTPS). |
| `NOEMA_UPLOADS_DIR` | Optional | Directory for `/upload` files (default `data/uploads`). |
| `NOEMA_RUNS_DIR` | Optional | Directory for `/api/evaluate` runs (default `data/runs`). |
| `NOEMA_RUNS_MAX` | Optional | Max run artifacts to retain (default `50`, set `0` to disable pruning). |

## API endpoints
Public:
- `GET /health` → `{"status":"ok"}`
- `GET /ready` → `{"status":"ok"}`
- `POST /api/verify` → stub verifier (returns verification result)
- `GET /verify` and `GET /verify/:id` → public verify page

Gated by judge key (header, query, or session cookie):
- `GET /ping` → `{"message":"pong"}`
- `POST /api/evaluate` → evaluates dataset, returns policy output + proof bundle
- `GET /app`, `GET /app/new`, `GET /app/results/:id`, `GET /upload`, `POST /upload`

### Example: call `/ping` with judge key
```bash
curl -H "X-Judge-Key: YOUR_JUDGE_KEY" http://localhost:8080/ping
```

## Gemini 3 integration
Noema uses Gemini 3 (default: Pro; optional Flash for latency/cost) to evaluate datasets against a structured set of governance constraints. The model produces per‑constraint severity scores and a short rationale summary, which are then aggregated into a policy decision and converted into a ZK witness.

## ZK proof pipeline
1. Gemini 3 returns per‑constraint severities.
2. Policy aggregation computes overall pass/fail and max severity.
3. A gnark circuit (or stub with the same public inputs) produces a proof and public inputs.
4. The public verify page validates the proof without exposing the dataset.

## Submission checklist
- Gemini 3 integration write‑up (200 words): TODO
- Public demo link: TODO
- Code repo link: TODO
- 3‑minute demo video link: TODO
