# Agent Report: BACKEND

## Summary
- Replaced the `/api/evaluate` handler with a Gemini-driven pipeline including dataset validation, sampling, caching, and server-side persistence.
- Added strict JSON response schema and parsing/validation for Gemini outputs (per-constraint severity + rationale + max severity + optional confidence).
- Implemented prompt construction using preset constraints plus custom constraints and attached image metadata for multimodal context.
- Added local caching keyed by dataset+spec+model+prompt version and persisted `gemini_output.json` for each run.
- Added dataset schema parsing (`items[].id`, `items[].text`) and configurable sampling limit via `NOEMA_SAMPLE_ITEMS`.
- Added runs index file `data/runs/index.json` with a configurable cap (`NOEMA_RUNS_INDEX_LIMIT`).
- Added Gemini client streaming support, structured request config, and usage metadata extraction.
- Fixed missing `encoding/json` import in `backend/internal/evaluate/parse.go` after refactor.

## Files Touched
- `backend/internal/gemini/gemini.go`
- `backend/internal/evaluate/handler.go`
- `backend/internal/evaluate/parse.go`
- `backend/internal/evaluate/storage.go`
- `backend/internal/config/config.go`
- `backend/cmd/server/main.go`
- `backend/internal/evaluate/dataset.go`
- `backend/internal/evaluate/images.go`
- `backend/internal/evaluate/output.go`
- `backend/internal/evaluate/cache.go`
- `backend/internal/evaluate/policy.go`
- `backend/internal/evaluate/schema.go`
- `backend/internal/evaluate/prompt.go`
- `backend/internal/evaluate/commitment.go`

## Commands Run
- `ls`
- `ls planning`
- `cat planning/agent_BACKEND.md`
- `cat planning.md`
- `cat README.md`
- `sed -n '1,220p' backend/internal/evaluate/handler.go`
- `sed -n '1,240p' backend/internal/gemini/gemini.go`
- `sed -n '1,220p' backend/internal/config/config.go`
- `rg -n "evaluate" backend/internal -g'*.go'`
- `sed -n '1,240p' backend/internal/evaluate/parse.go`
- `sed -n '1,240p' backend/internal/evaluate/spec.go`
- `sed -n '1,260p' backend/internal/evaluate/storage.go`
- `sed -n '1,200p' backend/internal/config/limits.go`
- `rg -n "api/evaluate|Evaluate" backend -g'*.go'`
- `sed -n '1,220p' backend/cmd/server/main.go`
- `rg -n "evaluate" backend/web -g'*.js'`
- `sed -n '200,320p' backend/web/static/wizard.js`
- `sed -n '1,220p' backend/internal/web/results.go`
- `sed -n '1,240p' backend/web/static/results.js`
- `rg -n "runs" backend/internal -g'*.go'`
- `rg -n "genai" -S backend go.sum go.mod`
- `sed -n '1,200p' backend/go.mod`
- `go env GOPATH`
- `rg -n "GenerateContentStream" /Users/yani/go/pkg/mod -g'*.go'`
- `sed -n '1,140p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/examples/models/generate_content/text_stream.go`
- `sed -n '1,180p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/examples/models/generate_content/image_modality_stream.go`
- `rg -n "InlineData" /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0 -g'*.go' | head`
- `sed -n '1110,1185p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/types.go`
- `rg -n "Usage" /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/types.go | head`
- `sed -n '2620,2695p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/types.go`
- `rg -n "constraints" backend/web/static/wizard.js`
- `sed -n '1,160p' backend/web/static/wizard.js`
- `rg -n "PRESET_CONSTRAINTS|constraints" backend/web/templates -S`
- `sed -n '1,200p' backend/web/static/constraints.js`
- `rg -n "func Text" /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0 -g'*.go' | head`
- `sed -n '1,120p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/models_helpers.go`
- `rg -n "ResponseSchema|ResponseMime" /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/types.go`
- `sed -n '2100,2195p' /Users/yani/go/pkg/mod/google.golang.org/genai@v1.44.0/types.go`
- `ls backend/internal/zk`
- `ls backend/internal`
- `ls -la backend/internal/zk`
- `ls backend/internal/crypto`
- `sed -n '1,200p' backend/internal/crypto/crypto.go`
- `gofmt -w backend/internal/evaluate/*.go backend/internal/gemini/gemini.go backend/internal/config/config.go`
- `rg -n "evaluate.Handler" backend/cmd/server/main.go`
- `sed -n '80,110p' backend/cmd/server/main.go`
- `rg -n "RunsMax" backend/internal/config`
- `sed -n '60,110p' backend/internal/config/config.go`
- `gofmt -w backend/cmd/server/main.go`
- `cd backend && go test ./...`
- `rg -n "policyThresholdFromSpec|pruneRuns|hashFileSHA256" backend/internal/evaluate`
- `gofmt -w backend/internal/evaluate/handler.go`
- `rg --files -g 'AGENTS.md'`
- `cat planning/agent_BACKEND.md`
- `cd backend && go test ./...`
- `cd backend && go build -o /tmp/noema-server ./cmd/server`

## Key Decisions
- Adopted JSON Schema enforcement using `response_mime_type=application/json` and `response_json_schema` to harden Gemini outputs.
- Chose SSE for streaming (`event: status|chunk|done`) with a `?stream=1` toggle so the existing JSON-only flow stays intact.
- Defined constraint policy evaluation as “fail if any constraint exceeds its allowed max severity,” and `policy_threshold` as the strictest allowed max.
- Used cache key hashing of dataset+spec+model+prompt version+sample size to reuse Gemini outputs and keep commitment deterministic.

## Tests
- `cd backend && go test ./...` (passed)
- `cd backend && go build -o /tmp/noema-server ./cmd/server` (passed)

## TODO / Follow-ups
- Decide if streaming format should be SSE vs chunked JSON (UI will need to parse the SSE events).
- Consider adding a server-side results fetch endpoint for `/app/results/:id` to reduce reliance on localStorage.
- Confirm ZK input contract; adjust `public_output` mapping if the circuit expects different semantics.

## Blockers / Questions
- Need PLANNER confirmation on the final Gemini JSON schema if it must match a specific ZK input shape.
- Need PLANNER decision on streaming protocol expectations for the frontend.
