# Grading criteria: cli-2363

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `app sub --help` or `app sub -h` still invokes the root's `After` (with `Before` never called).
- Deeper trees are not covered: `app sub nested --help` must not invoke `sub.After` or `app.After` (a fix that only suppresses the direct parent's `After` is wrong).
- `app sub --help nested` (help flag followed by more args) still fires an ancestor `After`.
- Regression in normal runs: `app sub` must still produce exactly `app.Before sub.Before sub.Action sub.After app.After`; nested runs analogously.
- `After` no longer runs when the `Action` returns an error (Before/After must stay paired).
- `app --help` starts running hooks, or the `help` command paths (`app help`, `app help sub`, `app sub help`) change their Before/After pairing or start running the command `Action`.
- Help output is no longer printed or changes.
- No tests were added for the new behavior.
- The existing suite (`go test ./...`) is not green, or `go vet` / gofmt problems are introduced.
- Changes far outside the scope of the request (e.g. reworking the hook lifecycle wholesale).

## Nice to have (optional; do not block on these)

- Also covering the "unknown flag plus `--help`" path (`app sub --bogus --help`), where help is shown after a parse error. The original fix does not cover it, so its absence is not a blocker.
- A release note line describing the fix.

## Judgment calls

- The signalling mechanism (context value, field on the command, return value, restructuring where `After` is deferred) is up to the implementer.
- Whether `After` should be skipped in *other* early-exit situations that never reach `Before` (e.g. version display, usage errors without help) is not part of this request; either behavior is acceptable as long as existing tests pass.
