# Requester's original words

Issue: https://github.com/pallets/click/issues/3458 (opened by @lemon24)
Fix PR: https://github.com/pallets/click/pull/3484 (by @yurishevtsov), merged 2026-05-21T15:20:36Z

## Issue title

`get_parameter_source()` returns `None` in 8.4.0

## Issue body (verbatim)

Starting with 8.4, get_parameter_source() returns None.

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

if __name__ == '__main__':
    main()
```

Output:

```console
$ pip install click==8.4.0 -q
$ python source.py                  
default: {'value': '/tmp/file', 'source': None}
nodefault: None
$ python source.py --default cli --nodefault cli
default: {'value': 'cli', 'source': None}
nodefault: {'value': 'cli', 'source': None}

$ pip install click==8.3.3 -q                   
$ python source.py                              
default: {'value': '/tmp/file', 'source': <ParameterSource.DEFAULT: 5>}
nodefault: None
$ python source.py --default cli --nodefault cli
default: {'value': 'cli', 'source': <ParameterSource.COMMANDLINE: 2>}
nodefault: {'value': 'cli', 'source': <ParameterSource.COMMANDLINE: 2>}
```

Environment:

- Python version: 3.14.4 (seems to happen on all versions, on both macOS and Linux)
- Click version: 8.4.0


## Issue comments (verbatim)

@kdeldycke:

Can confirm the issue. My Click Extra tests detected this regression on `main` before the 8.4.0 release. @Rowlando13 next time give me 1 or 2 days before releasing once all PRs are merged so I can catch the last remaining bugs! 😅

@kdeldycke:

A fix has been proposed in https://github.com/pallets/click/pull/3484 . @lemon24 can you test it?

@lemon24:

Thank you @yurishevtsov, @kdeldycke, all good on my side!

```console
$ pip install 'git+https://github.com/pallets/click.git@refs/pull/3483/head' -q
$ python -m reader --db db.sqlite list feeds | head -n1
reader:archived
$ ./run.sh test tests/test_cli.py tests/test_config_utils.py --runslow -q --disable-warnings
................                                                          [100%]
16 passed, 2 warnings in 0.94s
```
