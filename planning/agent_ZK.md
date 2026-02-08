# Agent: ZK

Goal: Implement stub-first ZK layer, then full gnark policy aggregation circuit and verifier, aligned to evaluation outputs.

Read first:
- planning.md
- README.md
- backend/internal/evaluate/spec.go
- backend/internal/evaluate/handler.go

Testing requirement:
- Add or update unit tests for your changes.
- Run `go test ./...` before reporting done.
- If tests fail for reasons outside your control, document the exact error.

File ownership (allowed to edit):
- `backend/internal/zk/*`
- `backend/internal/verify/*`
- `backend/cmd/server/main.go` (only to wire verify routes)

Do not edit:
- `backend/internal/evaluate/*` (read-only, do not modify)
- `backend/internal/gemini/*`
- `backend/web/*`
- `README.md`, `DEPLOYMENT.md`, `planning/*`

Tasks (do in order):
1) Stub-first proof pipeline
- Implement a deterministic stub proof generator that matches the final public input format.
- Ensure `GenerateProof` returns placeholder `proof_b64` and `public_inputs_b64` but is stable across identical inputs.
- Ensure `VerifyProof` validates the stub format and returns true/false deterministically.

2) Define public inputs & proof format
- Public inputs should include:
  - policy threshold
  - max severity
  - overall pass/fail (1/0)
  - commitment (hash of dataset/spec + gemini output)
- Decide commitment hash function (Poseidon or SHA-compatible in gnark).

3) Circuit design (full policy aggregator)
- Inputs: per-constraint severities (0/1/2), enabled flags, allowed_max_severity.
- Compute max severity of enabled constraints.
- Compute pass/fail: max severity <= policy threshold.
- Prove that commitment matches private inputs (dataset/spec hash + gemini output hash).

4) Implementation
- Create new package: `backend/internal/zk` (or similar).
- Build gnark circuit and proof generation helpers.
- Provide `GenerateProof` and `VerifyProof` functions with clear API.

5) Verifier endpoint integration
- Add verifier endpoint or helper for public verification.
- Ensure verification result is exposed to frontend (public verify tab).

6) Testing
- Add minimal unit test for circuit constraints and verification.

Deliverables
- Circuit + proof generation + verification integrated into backend.
- Clear public input format documented for frontend.

Questions to resolve (ask PLANNER if blocked):
- Exact commitment format and hash choice.
- Whether to include per-constraint outputs as private or public inputs.
