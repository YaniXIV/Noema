# PLANNING — Gemini 3 Hackathon (Noema)

Role: PLANNER

## Goal
Ship a compelling, hackathon‑ready version of Noema that demonstrates Gemini 3‑powered policy evaluation and verifiable compliance via zero‑knowledge proofs, with a polished demo and clear documentation.

## Assumptions
- Backend is Go + Gin, serving HTML/CSS/JS from `backend/web/`.
- Gemini 3 is used for reasoning over dataset content and producing constraint severities.
- Default model: Gemini 3 Pro (with optional fallback to Flash for cost/latency testing).
- ZK stack is gnark; full policy aggregation circuit is the target.
- Public demo must be accessible without login.
- Deadline: Feb 9, 2026 @ 6:00pm MST.

## Hackathon Requirements Checklist
- New application using Gemini 3.
- 200‑word Gemini 3 integration write‑up.
- Public demo link.
- Public code repo link (if not AI Studio app).
- ~3‑minute demo video.

## Product Scope (Hackathon MVP)
- Upload or paste dataset (JSON) + optional images (limit to a small count to control cost).
- Configure constraints (preset + custom), set allowed severity thresholds.
- Gemini 3 evaluates dataset against constraints and returns per‑constraint severity + rationale summary.
- Policy evaluation aggregates into PASS/FAIL + commitment.
- ZK proof generation (stub or real gnark circuit) produces `proof` + `public_inputs`.
- Public verification page shows proof and verification result without revealing dataset.

## System Interfaces (target)
- `POST /api/evaluate` returns:
  - `public_output`: `overall_pass`, `max_severity`, `policy_threshold`, `commitment`
  - `proof`: `system`, `curve`, `proof_b64`, `public_inputs_b64`
  - `verified`
- `GET /app/results/:id` renders results from server or persisted storage.
- `GET /verify/:id` public verification page (no login) or CLI verifier.

## Task Groups (parallelizable)

Group A — Product & Demo Narrative
- Agent: PLANNER
- Prompt: Craft the demo story, user flow, and “wow” moments. Define the 3‑minute demo script and which screens are shown.
- Dependencies: none

- Agent: WRITER
- Prompt: Draft the 200‑word Gemini 3 integration write‑up aligned to judging criteria. Emphasize Gemini 3 reasoning, multimodality, and latency.
- Dependencies: none

Group B — Backend + Gemini Integration
- Agent: BACKEND
- Prompt: Implement Gemini 3 evaluation pipeline in Go. Accept dataset JSON + limited images (cap count/size), run structured prompt, return per‑constraint severities and rationale summary. Add caching and graceful fallbacks; default to Gemini 3 Pro with env override to Flash.
- Dependencies: none

- Agent: BACKEND
- Prompt: Persist run results server‑side (e.g., `data/runs/<run_id>/result.json`) and update results page to read from server first, then localStorage fallback.
- Dependencies: none

Group C — ZK Proof Layer
- Agent: ZK
- Prompt: Define gnark circuit interface and public inputs for policy evaluation (severity aggregation + thresholding). Implement stub proof generator if full circuit is risky.
- Dependencies: none

- Agent: ZK
- Prompt: Implement verifier endpoint or CLI to validate proof with public inputs. Expose result to UI.
- Dependencies: SEQUENTIAL: depends on Group C task “Define gnark circuit interface and public inputs”.

Group D — Frontend UX
- Agent: FRONTEND
- Prompt: Update wizard UI to show streaming Gemini evaluation status, per‑constraint severity, and confidence. Add “Proof generated” section with copyable proof and public inputs.
- Dependencies: SEQUENTIAL: depends on Group B task “Implement Gemini 3 evaluation pipeline in Go”.

- Agent: FRONTEND
- Prompt: Add public verification page (`/verify/:id`) that displays proof, public inputs, and verification result.
- Dependencies: SEQUENTIAL: depends on Group C task “Implement verifier endpoint or CLI”.

Group E — Demo/Deployment
- Agent: DEVOPS
- Prompt: Provide a public demo deployment plan (Render/Fly/GCP), ensure no login required for verification page, configure env vars, and cap verifier listing to most recent N proofs.
- Dependencies: SEQUENTIAL: depends on Group B “Persist run results server‑side”.

- Agent: VIDEO
- Prompt: Produce demo video outline and capture list: dataset upload, constraints, Gemini evaluation, proof creation, public verify page.
- Dependencies: SEQUENTIAL: depends on Group D “Add public verification page”.

Group F — Documentation
- Agent: DOCS
- Prompt: Rewrite README to be hackathon‑ready, add architecture diagram and quickstart, include Gemini 3 usage, ZK flow, and demo link.
- Dependencies: SEQUENTIAL: depends on Group B “Implement Gemini 3 evaluation pipeline in Go”.

## Sequential Dependencies Summary
- ZK verifier depends on circuit interface definition.
- Frontend results enhancements depend on Gemini integration.
- Public verification page depends on verifier endpoint/CLI.
- Deployment depends on server‑side result persistence.
- Demo video depends on verification page.
- README refresh depends on Gemini integration.

## Risk Mitigations
- If gnark circuit is not ready, use a deterministic stub with the exact public input format to keep UI/UX and demo coherent.
- If Gemini 3 latency/cost is high, use short prompts, cap dataset size or sample rows, and cache outputs. Provide Flash fallback.
- Ensure public verify flow does not require login.

## Dataset Schema (proposed)
JSON top‑level:
- `items`: array of objects
- `items[].id`: string (required)
- `items[].text`: string (required)
- `items[].metadata`: object (optional)
- `items[].image_ref`: string (optional, matches uploaded image filename)

Images:
- Optional, capped small count (e.g., 5–10) and size (e.g., 5MB each).

Rationale:
- Simple to validate and prompt.
- Supports multimodal with explicit `image_ref` mapping.
