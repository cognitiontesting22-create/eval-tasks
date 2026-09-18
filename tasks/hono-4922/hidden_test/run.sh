#!/usr/bin/env bash
# Usage: run.sh <path-to-hono-checkout>
# Copies the hidden test into src/, runs only that file with vitest, exits
# nonzero on failure. Requires bun (1.2.19) and node (24.x) on PATH.
set -euo pipefail
REPO="${1:?usage: run.sh <repo-path>}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_NAME="hidden-cookie-duplicates.test.ts"

cp "$HERE/$TEST_NAME" "$REPO/src/$TEST_NAME"
trap 'rm -f "$REPO/src/$TEST_NAME"' EXIT

cd "$REPO"
bun install --frozen-lockfile --silent
bunx vitest --run --project main "src/$TEST_NAME"
