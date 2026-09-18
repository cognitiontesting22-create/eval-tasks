# Grading criteria: click-3818

## Merge question

Would you merge this? Judge it as a Click maintainer reviewing a bug fix for
"a second Ctrl-C during abort/error reporting escapes `Command.main()` as a
traceback". A different structure from the reference diff is fine as long as
the observable behavior below holds.

## Blockers (any one of these makes the output unacceptable)

- In standalone mode, a `KeyboardInterrupt` raised while Click is writing the
  `Aborted!` message (e.g. from `sys.stderr.isatty()` or `.write()`) still
  escapes `Command.main()` instead of ending in `SystemExit(1)`.
- The fix only covers the prompt/`Abort` path. An interrupt raised while a
  `ClickException` is being shown must still produce `SystemExit` with that
  exception's `exit_code`; an interrupt arriving on the `KeyboardInterrupt`/
  `EOFError`-from-command-body path (which writes a blank line before
  `Aborted!`) must still exit 1; an interrupt while a successful run is exiting
  must still exit 0.
- The late interrupt changes the exit code (e.g. everything becomes 1, or a
  `ClickException` with `exit_code=7` exits 1).
- Without a late interrupt, behavior regresses: `Aborted!` no longer printed to
  stderr, exit code not 1, `ClickException.show()` not called, `Exit` code not
  honored, EPIPE handling removed.
- `standalone_mode=False` semantics changed: the first interrupt must still
  reach the caller as `click.Abort` with the original exception as
  `__cause__`; a `ClickException` must still propagate; a later interrupt must
  propagate to the caller rather than being swallowed or converted to
  `SystemExit`.
- No tests added, or the existing suite (`uv run pytest`) no longer passes.
- Change far outside the scope of the request (e.g. rewriting prompt handling,
  changing the `Aborted!` text, altering `Context.exit` semantics).

## Nice to have (optional; the reference PR did these)

- A `CHANGES.md` entry.
- Documentation update describing what Click does for each way a command can
  end in standalone vs. non-standalone mode.
- Tests that cover several interrupt timings (during the blank line, during
  `Aborted!`, during `ClickException.show()`, during the final exit).
- Reduced nesting / a single teardown path in `main()` rather than per-handler
  try/except wrappers. This is a style preference, not a requirement.

## Judgment calls

- Whether the abort message is *lost* or *retried* on a late interrupt is up
  to the implementer; the prompt states losing it is acceptable.
- How the late interrupt is simulated in tests (fake stderr, patched
  `sys.exit`, real SIGINT) is up to the implementer.
