# A parameter named `help` breaks parsing

The docs say the automatic help parameter "performs automatic conflict resolution": if a command implements a parameter with the same name, the default help stops accepting it. That is not what happens when the conflict is on the parameter *name* rather than the `--help` flag. In `help.py`:

```python
import click

# This works
@click.command()
@click.argument('helps')
def this_works(helps):
    print(helps)
# > python -m help this        -> this
# > python -m help --help      -> shows usage and "--help  Show this message and exit."

# This does not work
@click.command()
@click.argument('help')
def this_does_not_work(help):
    print(help)
# > python -m help this
# Usage: help.py [OPTIONS] HELP
# Error: Invalid value for '--help': 'this' is not a valid boolean.
# > python -m help --help
# Error: Missing argument 'HELP'.
```

Same with an option whose flag differs but whose destination name is `help`, e.g. `@click.option('--assist', 'help')` with `--assist value`: "Invalid value for '--help'". Same with a custom help flag: `context_settings={"help_option_names": ["--man"]}` plus `@click.option('--foo', 'man')` breaks alike.

## Wanted

A user parameter that merely shares its name with the help option must not interfere with it, and vice versa:

- `@click.argument('help')` receives its positional value normally, and `--help` still prints the help page and exits 0.
- `@click.option('--assist', 'help')` receives the value passed to `--assist`, and `--help` still prints help.
- The same holds for custom `help_option_names`.
- An option that reuses the help *flag* itself, such as `@click.option('--help', default='x')`, keeps today's behavior: it takes the flag over and the automatic help option is dropped for that command.

Please include tests for the new behavior. The existing test suite must stay green:

```
uv sync --locked --group tests && uv run pytest -v --tb=short
```
