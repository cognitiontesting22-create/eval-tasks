# Grading criteria: click-3484

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `ctx.get_parameter_source(param.name)` still returns `None` inside `ParamType.convert()` for a value that came from the command line, an env var, a `default_map`, or the declared default.
- `get_parameter_source()` still returns `None` inside an option callback (eager or not) at the time the callback runs.
- The "was this flag explicitly given?" pattern does not work: an eager callback that bails out when the source is `DEFAULT`/`DEFAULT_MAP` still overwrites an environment setting when the flag was not passed.
- Feature-switch groups regress: the final source (after parsing) for a shared destination is not the winning option's source. Specifically, an option that lost arbitration (e.g. it was only supplied by `default_map` or its declared default) must not leave its own source behind when another option in the group won from the command line or environment, regardless of declaration order.
- The value arbitration for feature-switch groups changes (which option wins, what value is exposed) compared to BASE.
- No tests were added for the new behavior.
- The existing suite (`uv sync --locked --group tests && uv run pytest -v --tb=short`) is not green.
- Changes far outside the scope of the request (unrelated refactors, public API changes).

## Nice to have (optional; do not block on these)

- A changelog entry.
- A docs note that writing `ctx.params[...]` directly from a callback bypasses source tracking.
- Test coverage of the env-var winner / default-map loser combination in both declaration orders.

## Judgment calls

- Exactly where the source is recorded during the parameter pipeline is the implementer's choice; grade only the observable results from `get_parameter_source()` during conversion, during callbacks, and after parsing.
- Behavior of `get_parameter_source()` for a name that was written into `ctx.params` directly by another callback (bypassing normal handling) is undefined by the request; do not penalize either `None` or the provisional source.
