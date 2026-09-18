#!/usr/bin/env bash
# Usage: run.sh <path-to-click-checkout>
# Copies the hidden test into the repo's tests/ dir, runs only that file,
# exits nonzero on failure.
set -euo pipefail
REPO="${1:?usage: run.sh <repo-path>}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_NAME="test_hidden_parameter_source_timing.py"

cp "$HERE/$TEST_NAME" "$REPO/tests/$TEST_NAME"
trap 'rm -f "$REPO/tests/$TEST_NAME"' EXIT

cd "$REPO"
uv sync --locked --group tests --quiet
uv run pytest -v --tb=short -p no:cacheprovider "tests/$TEST_NAME"
