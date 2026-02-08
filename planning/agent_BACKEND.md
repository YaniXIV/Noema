# Agent: BACKEND

Goal: Implement Gemini 3 Pro evaluation pipeline in Go, streaming responses, caching, and server-side persistence.

Read first:
- planning.md
- README.md
- backend/internal/evaluate/handler.go
- backend/internal/gemini/gemini.go
- backend/internal/config/config.go

Server build owner:
- You are responsible for keeping the server buildable after your changes.
- Run `go test ./...` and `go build ./cmd/server` before reporting done.
- If tests fail for reasons outside your control, document the exact error.

File ownership (allowed to edit):
- `backend/internal/evaluate/*`
- `backend/internal/gemini/*`
- `backend/internal/config/*`
- `backend/cmd/server/main.go` (only for wiring routes/config)

Do not edit:
- `backend/internal/zk/*`
- `backend/internal/verify/*`
- `backend/web/*`
- `README.md`, `DEPLOYMENT.md`, `planning/*`

Tasks (do in order):
1) Gemini client integration
- Update `backend/internal/gemini/gemini.go` to support structured evaluation calls.
- Default model: Gemini 3 Pro, allow override via `GEMINI_MODEL` (fallback to Flash for cost/latency tests).
- Support streaming responses (server-sent events or chunked JSON) for evaluation output.
- Input should include dataset JSON + optional images (multimodal).
- Ensure `GEMINI_API_KEY` is required; return a clean error if missing.

2) Evaluation prompt + schema
- Define a stable system prompt and JSON response schema for:
  - Per-constraint severity (0/1/2)
  - Short rationale summary per constraint
  - Overall max severity
  - Optional confidence (0-1)
- Enforce schema validation on responses (strict JSON parse). If invalid, return error.

3) Server-side persistence
- Persist each run in `backend/data/runs/<run_id>/`:
  - `dataset.json`
  - `spec.json`
  - `gemini_output.json`
  - `result.json` (the API response)
- Implement a simple index file `backend/data/runs/index.json` with most recent N runs for public verifier listing.
- Make sure this index caps to N (configurable via env, default 50).

4) API behavior updates
- `POST /api/evaluate` should:
  - Validate dataset schema from `planning.md`.
  - Call Gemini; stream progress to client (if feasible).
  - Produce deterministic, structured output for ZK input.
  - Write files in run folder.
  - Return response consistent with existing `EvaluateResponse`.

5) Robustness
- Add dataset sampling for large files (e.g., take first N items, N configurable).
- Log timing and basic metrics (latency, token usage if available).
- Add cache by hashing dataset + spec to reuse previous Gemini output.

Deliverables
- Updated Go code with unit-safe behavior.
- Clear errors surfaced to frontend.
- No breaking of existing routes.

Questions to resolve (ask PLANNER if blocked):
- Final Gemini JSON schema shape if conflict with ZK input.
- SSE vs chunked JSON for streaming UI.
