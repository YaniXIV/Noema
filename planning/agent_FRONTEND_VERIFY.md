# Agent: FRONTEND-VERIFY

Goal: Implement public verification tab/page and proof listing UX.

Read first:
- planning.md
- README.md
- backend/web/templates/*.html
- backend/web/static/styles.css
- backend/web/static/wizard.js

Testing requirement:
- Add minimal JS/UI checks if possible (or at least manual test steps).
- Run `go test ./...` only if you changed Go files (otherwise skip).
- If you can’t run tests, document why.

File ownership (allowed to edit):
- `backend/web/templates/verify.html`
- `backend/web/templates/verify_results.html` (or `backend/web/templates/verify_result.html` if preferred)
- `backend/web/static/verify.js`
- `backend/web/static/styles.css` (only sections scoped to verify page; do not refactor global styles)

Do not edit:
- `backend/web/templates/index.html`
- `backend/web/templates/app*.html`
- `backend/web/static/wizard.js`
- `backend/cmd/server/main.go`
- `backend/internal/*`
- `README.md`, `DEPLOYMENT.md`, `planning/*`

Tasks (do in order):
1) Public Verify page
- Create template `/verify` (public, no judge key required).
- Show list of most recent N proofs with:
  - run id, status, timestamp
  - verify action button
  - quick result indicator (check/failed)
- Add filter/search by run id.

2) Verification flow
- Add a “Verify” button per entry that calls backend verify endpoint and updates UI inline.
- Add a “Verify all recent” action with progress indication (optional).

3) Results linking
- Allow clicking a run to open its results page (public safe view).
- Provide “copy proof” and “copy public inputs” per run.

4) Styling
- Ensure Verify page matches improved global design.
- Make list scannable and compact.

Deliverables
- New template and JS for verify page.
- Minimal backend hooks (if needed, coordinate with BACKEND/ZK).

Questions to resolve (ask PLANNER if blocked):
- Whether public verify should show any non-public data (default: no).
