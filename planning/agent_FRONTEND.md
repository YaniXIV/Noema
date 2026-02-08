# Agent: FRONTEND

Goal: Redesign the UI to be professional and usable, keeping the dark/electric aesthetic but fixing layout, navigation, and interaction quality.

Read first:
- planning.md
- README.md
- backend/web/templates/*.html
- backend/web/static/styles.css
- backend/web/static/wizard.js
- backend/web/static/constraints.js

Testing requirement:
- Add minimal JS/UI checks if possible (or at least manual test steps).
- Run `go test ./...` only if you changed Go files (otherwise skip).
- If you can’t run tests, document why.

File ownership (allowed to edit):
- `backend/web/templates/index.html`
- `backend/web/templates/app.html`
- `backend/web/templates/app_new.html`
- `backend/web/templates/app_results.html`
- `backend/web/static/styles.css`
- `backend/web/static/wizard.js`
- `backend/web/static/results.js` (if exists; otherwise create only if required)

Do not edit:
- `backend/web/templates/verify*.html`
- `backend/web/static/verify.js`
- `backend/cmd/server/main.go`
- `backend/internal/*`
- `README.md`, `DEPLOYMENT.md`, `planning/*`

Design requirements
- Keep dark + electric accent, but make the layout feel intentional and professional.
- Navigation must be clear and logically grouped.
- Provide a public "Verify" tab in the main nav, visible on landing page and app pages.
- Judge key gating should only apply to upload/evaluate; verify page is public.
- Make constraints UI clearer with better grouping and controls.
- Improve button hierarchy and spacing.
- Avoid nested lists in documentation (but okay in UI).

Tasks (do in order):
1) IA + layout
- Redesign nav structure across all templates for clarity:
  - Public: Home, Verify
  - Auth: Dashboard, New Evaluation, Upload, Verify, Log out
- Fix card layout in dashboard and wizard so spacing feels balanced.
- Ensure responsive behavior.

2) Wizard UX
- Improve dataset step: clear file vs paste JSON choice.
- Add small preview or validation status for dataset.
- Make constraints list scannable: group by category, add subtle separators, better toggle + severity control.
- Improve custom constraint UX: inline add/remove, validation, disabled state when limit hit.

3) Results UX
- Show evaluation status, per-constraint summary, proof info, verify button.
- Add a small “copy proof” and “copy public inputs” control.

4) Styles
- Refine typography, spacing, button styles, and card alignment.
- Remove duplicated/unused stylesheet (if safe).
- Ensure accessibility: focus states, contrast.

Deliverables
- Updated templates and CSS.
- Updated JS only as needed for UI improvements.

Questions to resolve (ask PLANNER if blocked):
- Confirm if landing page should include a short product pitch or stay minimal.
