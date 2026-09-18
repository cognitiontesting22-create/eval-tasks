# Grading criteria: cli-2426

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `app sub value --pa --generate-shell-completion` (one positional before the partial flag, in a child command) still prints subcommand suggestions or nothing instead of the child's flags whose names start with `pa`.
- Two or more positionals (`app sub first second --pa ...`) or nested commands (`app sub nested value --pa ...`) are not handled.
- Regression at the root level: `app --pa --generate-shell-completion` or `app value --pa --generate-shell-completion` no longer lists the root flags (a common wrong fix that always reads the final argument breaks this).
- Regression in the no-positional child case (`app sub --pa ...`).
- Output format changes: each suggestion must remain on its own line as `--name` (or `--name:usage` when the flag has usage text); hidden flags must not be suggested; a lone `-` must still list all visible flags.
- Completion runs the command's `Action`, or completion requests after `--` start producing output.
- A custom `ShellComplete` handler no longer receives the raw arguments (`value --pa`) or is bypassed.
- With `EnableShellCompletion` false, an unknown `--pa` flag stops being an error.
- No tests were added for the new behavior.
- The existing suite (`go test ./...`) is not green, or `go vet` / gofmt problems are introduced.
- Changes far outside the scope of the request.

## Nice to have (optional; do not block on these)

- Table-driven regression tests covering root, child and nested commands with zero, one and two positionals, plus hidden flags, custom completer, disabled completion and `--`.
- A release note line describing the fix.

## Judgment calls

- Whether the fix works by locating the argument being completed differently, by normalizing the argument list first, or by another mechanism is up to the implementer.
- Ordering of the suggested flags is not specified; any order is acceptable as long as the set is right.
