#!/usr/bin/env bash
# Usage: run.sh <path-to-urfave-cli-checkout>
# Copies the hidden test into the module root (external test package), runs
# only that file's tests, exits nonzero on failure. Requires go 1.22+ on PATH.
set -euo pipefail
REPO="${1:?usage: run.sh <repo-path>}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_NAME="hidden_help_hook_balance_test.go"

cp "$HERE/$TEST_NAME" "$REPO/$TEST_NAME"
trap 'rm -f "$REPO/$TEST_NAME"' EXIT

cd "$REPO"
go test -count=1 -run 'Test(SubcommandHelpFlag|NestedHelpFlag|HelpFlagBeforeTrailingArgs|RootHelpFlagRunsNoHooks|Normal(Subcommand|Nested)RunKeepsHookOrder|HelpCommandStaysBalanced|AfterStillRunsWhenActionFails|AfterRunsWithoutBeforeDefined)' -v .
