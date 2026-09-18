# `get_parameter_source()` returns `None` during type conversion and eager callbacks

On current main (8.4.0), `ctx.get_parameter_source()` returns `None` inside `ParamType.convert()` and inside eager option callbacks. 8.3.3 returned the real source.

Repro:

```python
import click

class Source(click.ParamType):
    name = "source"
    def convert(self, value, param, ctx):
        return {'value': value, 'source': ctx.get_parameter_source(param.name)}

@click.command
@click.option('--default', type=Source(), default='/tmp/file')
@click.option('--nodefault', type=Source())
def main(default, nodefault):
    print("default:", default)
    print("nodefault:", nodefault)
```

```console
$ python source.py
default: {'value': '/tmp/file', 'source': None}
nodefault: None
$ python source.py --default cli --nodefault cli
default: {'value': 'cli', 'source': None}
nodefault: {'value': 'cli', 'source': None}
```

Expected (8.3.3 behavior):

```console
$ python source.py
default: {'value': '/tmp/file', 'source': <ParameterSource.DEFAULT: 5>}
nodefault: None
$ python source.py --default cli --nodefault cli
default: {'value': 'cli', 'source': <ParameterSource.COMMANDLINE: 2>}
nodefault: {'value': 'cli', 'source': <ParameterSource.COMMANDLINE: 2>}
```

This breaks tools whose eager callback checks whether a flag was explicitly given (Flask's `--debug` callback skips overriding `FLASK_DEBUG=1` when the source is `DEFAULT`/`DEFAULT_MAP`); that guard never fires now.

## Wanted

- `get_parameter_source(name)` must return the correct source (`COMMANDLINE`, `ENVIRONMENT`, `DEFAULT_MAP`, `DEFAULT`, `PROMPT`) while the parameter's type conversion and its callback (eager or not) run, not only after parsing completes.
- Behavior after parsing must be unchanged. Feature-switch groups (several options sharing one destination via `flag_value`) must keep the source of the option that won the slot. Example: `--without-xyz` (`flag_value=False`) and `--with-xyz` (`flag_value=True, default=True`) sharing `enable_xyz`; invoked with `--without-xyz`, the final source is `COMMANDLINE` even if a `default_map` also supplies `enable_xyz`; invoked with nothing, value `True` with source `DEFAULT`; if `--with-xyz` has an `envvar` that is set and nothing is passed, the source is `ENVIRONMENT` (also when a `default_map` supplies `enable_xyz`).

Please include tests for the new behavior. The existing test suite must stay green:

```
uv sync --locked --group tests && uv run pytest -v --tb=short
```
