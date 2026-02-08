# Noema Project Summary

**Overview**
Noema is a Go/Gin web app that evaluates dataset compliance against governance constraints, produces a policy decision, and generates a verifiable proof bundle. The UI is server-rendered HTML with client-side JS for wizards and localStorage persistence. The current proof pipeline is a deterministic stub that can later be swapped with a real ZK system. The Gemini client exists but is not currently wired into `/api/evaluate` in the server handler.

**Repository Layout**
- `backend/` holds the Go server, internal packages, templates, and static assets.
- `planning/` and `planning/agent_reports/` capture design intent and agent activity.
- `README.md` and `DEPLOYMENT.md` describe product framing and deployment.

**Backend Architecture (Go/Gin)**
1. Server entrypoint and routes
- `backend/cmd/server/main.go` configures Gin, loads templates and static files, wires public and gated routes, and optionally runs a Gemini test if `GEMINI_TEST=1`.
- Public routes: `/`, `/verify`, `/verify/:id`, `/health`, `/ready`, `/api/verify`.
- Cookie-gated routes: `/app`, `/app/new`, `/app/results/:id`, `/upload`, `/api/evaluate`.
- Judge key gated API: `/ping`.

2. Auth and session handling
- `backend/internal/auth/middleware.go` implements `JudgeKey` (header/query) and `CookieAuth` (signed cookie) middleware.
- `backend/internal/session/session.go` signs and verifies cookies using HMAC-SHA256 with `NOEMA_COOKIE_SECRET`.
- `backend/internal/web/auth.go` renders the login form and sets the session cookie on successful login.

3. Configuration and limits
- `backend/internal/config/config.go` loads `.env`, resolves env vars for judge key, uploads, runs dir, sampling, and run limits.
- `backend/internal/config/limits.go` defines max dataset/image sizes and multipart memory limits.

4. Evaluation pipeline
- `backend/internal/evaluate/handler.go` handles `POST /api/evaluate` and is the primary pipeline stage.
- The handler parses multipart form fields for `spec`, `dataset`, optional `images`, and optional `eval_output`.
- Spec validation occurs in `backend/internal/evaluate/parse.go` and `backend/internal/evaluate/spec.go`.
- Dataset schema validation occurs in `backend/internal/evaluate/dataset.go` and is enforced by `parseUploads`.
- Run artifacts are saved under `data/runs/<run_id>/` by `backend/internal/evaluate/storage.go`.
- Policy aggregation uses `backend/internal/evaluate/policy.go` and checks per-constraint severities against allowed thresholds.
- Proof generation uses `backend/internal/zk/zk.go` (deterministic stub). Public outputs include `overall_pass`, `max_severity`, `policy_threshold`, and a commitment hash.
- If `eval_output` is not provided, the handler uses a stub output with severity 0 for all enabled constraints.

5. Gemini integration (present but not wired)
- `backend/internal/gemini/gemini.go` provides Gemini client wrappers and structured JSON schema support.
- `backend/internal/evaluate/prompt.go`, `schema.go`, and `output.go` define the prompt, response schema, and JSON validation for Gemini outputs.
- This is ready for integration but not connected in the current handler.

6. Verification pipeline
- `backend/internal/verify/handler.go` exposes `/api/verify` to validate proofs. It supports the stub proof format and a legacy stub fallback.
- `backend/internal/zk/FORMAT.md` documents the public inputs encoding for the stub pipeline.

**Frontend Architecture (Templates + JS + CSS)**
1. Templates
- `backend/web/templates/index.html`: judge key login and link to public verify page.
- `backend/web/templates/app.html`: dashboard with recent runs and demo buttons.
- `backend/web/templates/app_new.html`: multi-step evaluation wizard.
- `backend/web/templates/app_results.html`: private results view.
- `backend/web/templates/verify.html`: public verification list.
- `backend/web/templates/verify_results.html`: public results view.
- `backend/web/templates/upload.html`: file upload form.
- `backend/web/templates/partial_header_*.html` and `partial_orb.html` define shared layout components.

2. Client scripts
- `backend/web/static/wizard.js` controls the evaluation wizard, builds the `spec`, posts multipart form data, and stores runs in localStorage.
- `backend/web/static/constraints.js` defines preset constraints and severity level descriptions for UI rendering.
- `backend/web/static/results.js` renders results from localStorage and labels severities as Limited/Moderate/Severe.
- `backend/web/static/verify.js` renders recent runs, verifies proofs via `/api/verify`, and provides copy actions.
- `backend/web/static/app.js` renders recent runs on the dashboard and handles demo button alerts.
- `backend/web/static/upload.js` provides the upload drag-and-drop UX.

3. Styling
- `backend/web/static/styles.css` defines the dark/electric design system, layouts, button styles, wizard, results views, and verify list presentation.

**Data Model and Schema**
1. Evaluation spec
- `backend/internal/evaluate/spec.go` defines `Spec` with `schema_version`, `evaluation_name`, `policy.reveal`, `constraints`, and `custom_constraints`.
- Each constraint has `id`, `enabled`, and `allowed_max_severity` (0..2).

2. Dataset schema
- `backend/internal/evaluate/dataset.go` defines `Dataset` with `items[]`, where each item must include `id` and `text`.
- Optional fields include `metadata` and `image_ref`.

3. Evaluation output
- `backend/internal/evaluate/output.go` defines `EvalOutput` with `schema_version`, per-constraint severity and rationale, and `max_severity`.
- Validation ensures matching enabled constraint IDs, allowed severities, and strict JSON shape.

**Proof System (Stub ZK)**
- `backend/internal/zk/zk.go` implements a deterministic stub proof using SHA-256 over public inputs.
- Public inputs are encoded as a string with fields for policy threshold, max severity, overall pass, and commitment.
- `backend/internal/zk/zk_test.go` validates roundtrip generation and input validation.

**Storage and Persistence**
- Run artifacts are stored under `data/runs/<run_id>/` by `backend/internal/evaluate/storage.go`.
- The storage layer includes helper functions for saving JSON and updating an `index.json` list, although index updates are not currently called in the evaluate handler.
- The UI relies on localStorage for recent runs and result retrieval.

**Deployment and Environment**
- `DEPLOYMENT.md` documents Render-based deployment with persistent storage and environment variables.
- `README.md` provides quickstart and API endpoints.
- `backend/go.mod` and `backend/go.sum` define the module and dependency lockfile.

**API Endpoints**
1. Public
- `GET /health` and `GET /ready` for health checks.
- `GET /` landing page.
- `GET /verify` and `GET /verify/:id` public verification UI.
- `POST /api/verify` proof verification endpoint.

2. Cookie-gated
- `GET /app`, `/app/new`, `/app/results/:id` for the evaluation UI.
- `GET /upload` and `POST /upload` for file uploads.
- `POST /api/evaluate` for evaluation, proof generation, and run persistence.

3. Judge-key gated
- `GET /ping`.

**Current Evaluation Flow (End-to-End)**
1. User logs in with judge key on `/` and receives a signed session cookie.
2. User runs the wizard at `/app/new` and submits a dataset with a spec.
3. The client posts `spec`, `dataset`, optional `images`, and optional `eval_output` to `/api/evaluate`.
4. The server validates the dataset, spec, and eval output (or stubs it), computes policy results, generates a stub proof, and returns the evaluation response.
5. The client stores the response in localStorage and navigates to the results view.
6. The verify page uses localStorage entries to list runs and can re-verify proof bundles via `/api/verify`.

**Testing**
- Unit and integration tests live under `backend/internal/evaluate` and `backend/internal/zk`.
- Evaluate tests cover optional eval output parsing, policy aggregation, handler integration, dataset validation, and image uploads.

**Notable Gaps and TODOs**
1. Gemini integration exists but is not wired into `/api/evaluate`.
2. The run index (`data/runs/index.json`) is not currently updated by the handler.
3. The public verify list is driven by localStorage rather than server-side data.

**Files Reference**
- Backend core: `backend/cmd/server/main.go`, `backend/internal/evaluate/*`, `backend/internal/auth/*`, `backend/internal/session/*`, `backend/internal/verify/*`, `backend/internal/zk/*`.
- Frontend: `backend/web/templates/*`, `backend/web/static/*`.
- Docs and planning: `README.md`, `DEPLOYMENT.md`, `planning.md`, `planning/agent_*.md`, `planning/agent_reports/*.md`.

