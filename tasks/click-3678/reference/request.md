# Requester's original words

Issue: https://github.com/pallets/click/issues/2819 (opened by @Rowlando13)
Fix PR: https://github.com/pallets/click/pull/3678 (by @kdeldycke), merged 2026-07-10T03:24:09Z

## Issue body (verbatim)

From the docs: 

The help parameter is implemented in Click in a very special manner. Unlike regular parameters it’s automatically added by Click for any command and it performs automatic conflict resolution. By default it’s called --help, but this can be changed. If a command itself implements a parameter with the same name, the default help parameter stops accepting it. There is a context setting that can be used to override the names of the help parameters called [help_option_names](https://click.palletsprojects.com/en/stable/api/#click.Context.help_option_names).

What is wrong: 
Help does not automatically resolve if there is an arg or kwarg called help. In the sample the file is called help.py

``` 
import click

# This works 
@click.command()
@click.argument('helps')
def this_works(helps):
    print(helps)

# > python -m help this
# this

# > python -m help --help
# Usage: help.py [OPTIONS] HELPS

# Options:
#   --help  Show this message and exit.

# This does not work
@click.command()
@click.argument('help')
def this_does_not_work(help):
    print(help)

# > python -m help this
# Usage: help.py [OPTIONS] HELP
# Try 'help.py --help' for help.

# Error: Invalid value for '--help': 'this' is not a valid boolean.

# > python -m help --help
# Usage: help.py [OPTIONS] HELP
# Try 'help.py --help' for help.

# Error: Missing argument 'HELP'.

# This does not work
@click.command()
@click.option('--help', default='this_2')
def this_does_not_work_also(help='this_2'):
    print(help)

# > python -m help 
# None

# > python -m help --help
# Error: Option '--help' requires an argument.

if __name__ == '__main__':
    #this_works()
    #this_does_not_work()
    this_does_not_work_also()

```
- Python version: 3.10.15
- Click version: 8.1.7


## Issue comments (verbatim)

@Veebaa:

I have been looking into this and have thought of two solutions for this issue:

1. Detect conflict early and warn the user about it:
- You could issue an error or a clear warning message when the user defines help as an argument, suggesting they rename it to avoid this conflict with --help.
- i.e. a safety net for users who would not be aware of this conflict.
``` 
if 'help' in command.params:
    raise click.UsageError("The parameter name 'help' conflicts with the default '--help' flag. Please choose a different name.")
```

2.  Handling the --help flag internally
- Modify the argument parsing flow to allow the help parameter to be used while still dealing with --help as a special case.


I would like to contribute to this issue. How can I help?

@Rowlando13:

Great! You verify that you can replicate error and write a test that captures this failure. 

@Veebaa:

I was able to replicate the errors. I ran 3 tests that capture the errors successfully. I have a created a PR. #2859

@kdeldycke:

This issue is partially covered by https://github.com/pallets/click/pull/3208 where I test the case of another option being reused with the `--help` name.

@kdeldycke:

Given https://github.com/pallets/click/pull/3676 , I guess the milestone change to 9.0.0 is just the blast radius of the recent `main` <-> `stable` dance. So I retarget this to 8.5.0 unless an explicit decision is made.

@kdeldycke:

@Rowlando13 I proposed a definitive fix at https://github.com/pallets/click/pull/3678, can you review and test it?
