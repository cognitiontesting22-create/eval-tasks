# A second Ctrl-C while Click is reporting an abort crashes the program

When a user presses Ctrl-C during `click.prompt()`, this may cause an unhandled exception. We see it intermittently in CI on a slow Ubuntu arm runner, so it is a timing race. A single interrupt prints `Aborted!` and exits with code 1. But if a second `KeyboardInterrupt` arrives while Click is still writing that message (the traceback ends inside the `echo()` writing `Aborted!` to stderr), it escapes `Command.main()` as a traceback and the process dies with Python's default interrupt exit instead of Click's.

Repro sketch: a command with `@click.option("--name", prompt="Name")` in standalone mode; the user interrupts the prompt and a second interrupt lands during the abort message. In a test, simulate it by making the `stderr` object raise `KeyboardInterrupt` from `isatty()` or `write()`.

## Expected

Click should handle `KeyboardInterrupt` gracefully regardless of timing. In standalone mode a late interrupt arriving while `Command.main()` reports the outcome or exits must never escape as a traceback, nor change the exit code already decided:

- an abort (from a prompt, or a `KeyboardInterrupt`/`EOFError` raised in the command body) still exits 1;
- a `ClickException` still exits with its own `exit_code`, even if the interrupt arrives while its message is shown;
- a successful run still exits 0.

Losing the message is acceptable; losing the exit code is not. Without a late interrupt, behavior is unchanged: `Aborted!` on stderr, exit code 1.

With `standalone_mode=False`, keep today's contract: the first interrupt reaches the caller as `click.Abort` (original exception as `__cause__`); a later interrupt propagates to the caller rather than being swallowed or turned into an exit.

Please include tests for the new behavior. The existing test suite must stay green:

```
uv sync --locked --group tests && uv run pytest -v --tb=short
```
