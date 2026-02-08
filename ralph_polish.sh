#!/usr/bin/env bash
# Overnight-safe Codex polish loop:
# - No `resume` (avoids CLI option drift)
# - Codex never commits; bash commits
# - Retries on Codex crashes / arg weirdness
# - Enforces go test; auto-reverts on failure
# - Streams Codex output live AND logs it (tee) + preserves Codex exit code
set -u  # don't use -e; we want to survive errors

ROOT="$(pwd)"
BACKEND="$ROOT/backend"

BUDGET_SECONDS="${BUDGET_SECONDS:-28800}"   # default 8 hours
MAX_ITERS="${MAX_ITERS:-2000}"              # high cap so time is the real stop
SLEEP_ON_FAIL="${SLEEP_ON_FAIL:-10}"        # seconds
MAX_RETRIES="${MAX_RETRIES:-5}"             # per-iteration codex retries

LOGDIR="${LOGDIR:-$ROOT/planning/agent_reports/ralph_polish_runs}"
mkdir -p "$LOGDIR"

die() { echo "Error: $*" >&2; exit 1; }

[[ -d "$BACKEND" ]] || die "backend/ not found (run from repo root)"
command -v codex >/dev/null 2>&1 || die "codex not in PATH (npm i -g @openai/codex)"
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not a git repo"
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
[[ "$BRANCH" != "HEAD" ]] || die "detached HEAD; create a branch first"

# Prompt file avoids shell quoting/heredoc weirdness
PROMPT_FILE="$LOGDIR/prompt.txt"
cat > "$PROMPT_FILE" <<'PROMPT'
You are an autonomous coding agent polishing the Noema repository.

Hard rules:
- DO NOT run git commit, git reset, git clean, or anything that modifies git history.
- I will run git commits outside of you.
- Keep changes small: max 3 files per iteration.
- No cosmetic-only edits.
- No new dependencies unless absolutely necessary.
- keep all the same ui functionality, the end user should be able to do the same stuff but the experiance and design should be better.

Goal: continuously improve production quality. continuously improve front end quality and polish.

Iteration:
1) Run: cd backend && go test ./...
2) If failing: fix the smallest root cause (max 3 files), rerun tests until passing.
3) If passing: make ONE meaningful improvement (max 3 files), then rerun tests until passing.

Pick the highest-value improvement:
A) Reliability/correctness: validation, error handling, run persistence, verify robustness
B) Tests: add/strengthen tests for the change you made
C) upgrading the user interface so that it looks better.
D) Maintainability: reduce duplication with tiny refactors
E) UX: wizard/results/verify clarity (small, safe changes)
F) Docs: README/DEPLOYMENT polish
G) Performance: safe wins only

Finish with:
- What you changed (short)
- Tests run (must include go test ./...)
PROMPT

start_ts="$(date +%s)"
iter=0

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

run_tests() {
  (cd "$BACKEND" && go test ./...)
}

revert_worktree() {
  # Hard revert tracked changes + remove untracked files created by agent
  git reset --hard HEAD >/dev/null 2>&1 || true
  git clean -fd >/dev/null 2>&1 || true
}

commit_if_clean_tests() {
  local it="$1"
  local msg="$2"

  if git diff --quiet && git diff --cached --quiet; then
    log "iter $it: no changes to commit"
    return 0
  fi

  if ! run_tests; then
    log "iter $it: tests failing after agent run; reverting"
    revert_worktree
    return 1
  fi

  git add -A
  if git diff --cached --quiet; then
    log "iter $it: nothing staged"
    return 0
  fi

  git commit -m "ralph polish iter ${it}: ${msg}" >/dev/null 2>&1 \
    && log "iter $it: committed" \
    || { log "iter $it: commit failed; leaving changes staged"; return 1; }
}

log "Starting overnight-safe polish loop on branch: $BRANCH"
log "Budget: ${BUDGET_SECONDS}s, Max iters: ${MAX_ITERS}, Logs: $LOGDIR"

while true; do
  now="$(date +%s)"
  elapsed=$((now - start_ts))
  if (( elapsed >= BUDGET_SECONDS )); then
    log "Time budget reached (${BUDGET_SECONDS}s). Stopping."
    exit 0
  fi

  if (( iter >= MAX_ITERS )); then
    log "Iteration cap reached (${MAX_ITERS}). Stopping."
    exit 0
  fi

  iter=$((iter + 1))
  logfile="$(printf "%s/iter_%04d.txt" "$LOGDIR" "$iter")"
  log "iter $iter: starting (log: $logfile)"

  # Ensure we begin from a passing baseline; if not, revert to last commit.
  if ! run_tests >/dev/null 2>&1; then
    log "iter $iter: baseline tests failing; reverting to last commit"
    revert_worktree
    run_tests || { log "iter $iter: still failing after revert; stopping"; exit 1; }
  fi

  # Run Codex with retries so CLI hiccups don't stop the night
  attempt=1
  success=0
  while (( attempt <= MAX_RETRIES )); do
    log "iter $iter: codex attempt $attempt"

    # Stream output live AND capture it to logfile.
    # IMPORTANT: with pipes, $? is tee's status; use PIPESTATUS[0] for codex.
    (
      cd "$BACKEND"
      codex exec --full-auto "$(cat "$PROMPT_FILE")"
    ) 2>&1 | tee "$logfile"

    rc=${PIPESTATUS[0]}
    if (( rc == 0 )); then
      success=1
      break
    fi

    log "iter $iter: codex failed (rc=$rc). sleeping ${SLEEP_ON_FAIL}s and retrying"
    sleep "$SLEEP_ON_FAIL"
    attempt=$((attempt + 1))
  done

  if (( success == 0 )); then
    log "iter $iter: codex failed after ${MAX_RETRIES} retries; reverting and continuing"
    revert_worktree
    continue
  fi

  # Commit with enforced tests; if tests fail, revert and keep going
  commit_if_clean_tests "$iter" "autopolish" || true

  # Small cooldown to avoid thrashing
  sleep 1
done

