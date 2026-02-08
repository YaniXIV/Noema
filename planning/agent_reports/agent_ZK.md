# Agent Report: ZK

## Summary
- Restored `backend/internal/evaluate/handler.go` to the original stub behavior to comply with read-only constraint.
- Added a ZK stub verifier with deterministic public input checking and integrated it into `/api/verify`.
- Implemented legacy fallback verification for existing `stub_proof_<runID>` / `stub_inputs_<runID>` payloads.
- Kept deterministic stub ZK pipeline, public input format, commitment helper, and unit tests in `backend/internal/zk/*`.
- Ensured server wiring uses `evaluate.Handler(config.RunsDir(), config.RunsMax())` so builds pass.

## Files Touched
- backend/internal/evaluate/handler.go
- backend/internal/verify/handler.go
- backend/internal/zk/zk.go
- backend/internal/zk/zk_test.go
- backend/internal/zk/FORMAT.md
- backend/cmd/server/main.go

## Commands Run
- `ls`
- `ls planning`
- `cat planning/agent_ZK.md`
- `cat planning.md`
- `sed -n '1,200p' backend/internal/evaluate/spec.go`
- `sed -n '1,200p' backend/internal/evaluate/handler.go`
- `rg --files backend/internal`
- `sed -n '1,200p' backend/internal/verify/handler.go`
- `sed -n '1,200p' backend/internal/crypto/crypto.go`
- `sed -n '1,200p' backend/internal/evaluate/parse.go`
- `sed -n '1,200p' backend/internal/evaluate/storage.go`
- `sed -n '1,200p' backend/internal/gemini/gemini.go`
- `rg -n "gemini" backend/internal/evaluate`
- `cat backend/go.mod`
- `rg -n "verify" backend -g'*.go'`
- `sed -n '1,140p' backend/cmd/server/main.go`
- `rg -n "policy_threshold|allowed_max_severity|max_severity" backend web -g'*.*'`
- `sed -n '1,220p' backend/internal/evaluate/output.go`
- `sed -n '1,220p' backend/web/static/wizard.js`
- `mkdir -p backend/internal/zk`
- `gofmt -w backend/internal/evaluate/handler.go backend/internal/verify/handler.go backend/internal/zk/zk.go backend/internal/zk/zk_test.go`
- `sed -n '1,220p' backend/internal/evaluate/handler.go`
- `sed -n '220,260p' backend/internal/evaluate/handler.go`
- `sed -n '1,120p' backend/internal/verify/handler.go`
- `go test ./...`
- `mkdir -p planning/agent_reports`
- `rg -n "evaluate.Handler" backend/cmd/server/main.go`
- `sed -n '80,110p' backend/cmd/server/main.go`

## Key Decisions
- Chose a stable string-based public inputs encoding (`noema_public_inputs_v1|pt=...|ms=...|op=...|c=...`) to keep stub and future circuit aligned.
- Used SHA-256 commitment helper for deterministic commitments in the stub pipeline.
- Added legacy verifier fallback to avoid breaking existing runs that produced `stub_proof_<runID>` format.

## Tests
- `go test ./...` (passed)

## TODO / Follow-ups
- Implement full gnark circuit and replace stub proof/verification with real proof system.
- Decide final commitment hash function (Poseidon vs SHA-256) once gnark integration is confirmed.
- Determine whether per-constraint outputs should be private or exposed as public inputs in the final circuit.

## Blockers / Questions
- Need PLANNER decision on commitment hash choice for the final circuit.
