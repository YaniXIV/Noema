#!/usr/bin/env bash
set -euo pipefail

ROOT="$(pwd)"
BACKEND="$ROOT/backend"

# Defaults: 4 hours, 300 iters
BUDGET_SECONDS="${BUDGET_SECONDS:-14400}"
MAX_ITERS="${MAX_ITERS:-300}"

LOGDIR="${LOGDIR:-$ROOT/planning/agent_reports/ralph_polish_runs}"
mkdir -p "$LOGDIR"

# Sanity checks
if [[ ! -d "$BACKEND" ]]; then
  echo "Error: backend/ not found. Run from repo root."
  exit 1
fi

if ! command -v codex >/dev/null 2>&1; then
  echo "Error: codex not found in PATH. Install with: npm i -g @openai/codex"
  exit 1
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Error: not inside a git repo."
  exit 1
fi

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" == "HEAD" ]]; then
  echo "Error: detached HEAD. Create a branch first (e.g. git checkout -b polish/auto)."
  exit 1
fi

START_TS="$(date +%s)"
ITER=1

# Build the prompt WITHOUT fragile heredoc-in-$(...)
INITIAL_PROMPT=""
IFS= read -r -d '' INITIAL_PROMPT <<'PROMPT' || true
You are an autonomous coding agent polishing the Noema repository.

Hard rule: do NOT run git commit. Do NOT attempt to write to .git.
I will commit changes externally.

Each execution performs EXACTLY ONE iteration:

1) Run tests:
   cd backend && go test ./...

2) If tests fail:
   Fix the root cause with the smallest patch (max 3 files), then re-run tests until they pass.

3) If tests pass:
   Make ONE meaningful improvement, highest value first:
   A) Wire Gemini into /api/evaluate when eval_output is absent (keep deterministic stub fallback when GEMINI_API_KEY unset).
   B) Reliability: validation, error handling, storage/index updates (data/runs/index.json).
   C) Tests: add or strengthen tests for the change.
   D) Maintainability: small refactors that reduce duplication without behavioral change.
   E) UX polish: wizard/results/verify feedback (small, safe tweaks).
   F) Docs: README/DEPLOYMENT clarity.
   G) Performance: safe improvements backed by measurement or obvious wins.

Constraints:
- Max 3 files modified per iteration.
- No new dependencies unless absolutely necessary.
- No big refactors or style-only edits.
- Never leave the repo failing tests.

Finish after the iteration with:
- What you changed
- Tests you ran (must include go test ./...)
- Any follow-up you recommend (optional)
PROMPT

commit_if_dirty() {
  local iter="$1"
  local msg="$2"

  # Nothing changed?
  if git diff --quiet && git diff --cached --quiet; then
    echo "== No git changes to commit for iter $iter =="
    return 0
  fi

  git add -A

  # Nothing staged?
  if git diff --cached --quiet; then
    echo "== No staged changes to commit for iter $iter =="
    return 0
  fi

  git commit -m "ralph polish iter ${iter}: ${msg}" >/dev/null
  echo "== Committed iter $iter =="
}

echo "== Starting Ralph polish loop on branch: $BRANCH =="

# Iteration 1
(
  cd "$BACKEND"
  codex exec \
    --full-auto \
    --sandbox workspace-write \
    "$INITIAL_PROMPT"
) | tee "$LOGDIR/iter_0001.txt"

commit_if_dirty 1 "autopolish"

# Subsequent iterations
while true; do
  NOW="$(date +%s)"
  ELAPSED=$((NOW - START_TS))

  if (( ELAPSED >= BUDGET_SECONDS )); then
    echo "== Time budget reached (${BUDGET_SECONDS}s). Stopping. =="
    exit 0
  fi

  if (( ITER >= MAX_ITERS )); then
    echo "== Iteration cap reached (${MAX_ITERS}). Stopping. =="
    exit 0
  fi

  ITER=$((ITER + 1))
  LOGFILE="$(printf "%s/iter_%04d.txt" "$LOGDIR" "$ITER")"

  (
    cd "$BACKEND"
    # resume doesn't take -C; we cd instead
    codex exec resume --last \
      --full-auto \
      --sandbox workspace-write \
      "Next iteration (${ITER}). Do NOT git commit. Run go test ./..., make ONE improvement, stop."
  ) | tee "$LOGFILE"

  commit_if_dirty "$ITER" "autopolish"
done

