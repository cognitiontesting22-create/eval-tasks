# Grading criteria: click-3678

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `@click.argument('help')` still fails: the positional value is not delivered to the callback, or `--help` no longer prints the help page and exits 0.
- An option whose destination name is `help` but whose flag differs (e.g. `@click.option('--assist', 'help')`) does not receive its value, or `--help` no longer works for that command.
- The fix only handles `--help` literally: with `context_settings={"help_option_names": ["--man"]}` a parameter named `man` still collides.
- Regression: an option that reuses the help flag itself (`@click.option('--help', ...)`) no longer takes the flag over, or a plain command's `--help` output changes.
- The submission "fixes" the problem by raising an error / refusing parameters named `help` instead of making them work.
- No tests were added for the new behavior.
- The existing suite (`uv sync --locked --group tests && uv run pytest -v --tb=short`) is not green.
- Changes far outside the scope of the request (unrelated refactors, changed public signatures, altered unrelated output).

## Nice to have (optional; do not block on these)

- A changelog entry and a docs note describing how the help option coexists with same-named parameters.
- A warning or clear error when a user explicitly defines a parameter whose *storage name* is one Click reserves internally (if the chosen implementation reserves one).
- Coverage of the `help` name being used in a group subcommand as well as a plain command.

## Judgment calls

- How the help value is stored internally (renamed destination, separate parsing path, etc.) is up to the implementer; grade only observable behavior.
- Whether the help option's own storage name should be exposed/documented is optional.
