# Requester's original words

Issue: https://github.com/urfave/cli/issues/2248 (opened by @TimSoethout)
Fix PR: https://github.com/urfave/cli/pull/2426 (by @fzlzjerry), merged 2026-09-14T14:29:02Z

## Issue title

Auto-completion with unfinished partial flags (`--XXX<TAB>` returns no completions

## Issue body (verbatim)

## My urfave/cli version is

v3.6.1

## Checklist

- [X] Are you running the latest v3 release? The list of releases is [here](https://github.com/urfave/cli/releases).
- [X] Did you check the manual for your release? The v3 manual is [here](https://cli.urfave.org/v3/getting-started/)
- [X] Did you perform a search about this problem? Here's the [GitHub guide](https://help.github.com/en/github/managing-your-work-on-github/using-search-to-filter-issues-and-pull-requests) about searching.

## Dependency Management

- My project is using go modules.

## Describe the bug

When passing in uncompleted flags (`--XXX<TAB>`), no auto-completion is available. 

## To reproduce

In a project with autocompletions enabled with command and arguments run:
```
$ command <TAB>
subcommand
$ command subcommand -<TAB>
--par1
--par2
--flag
$command subcommand --pa<TAB>
* empty
```
Underlying I see respectively begin passed to the binary:
- `--generate-shell-completion`, returns list of commands
- `- --generate-shell-completion` returns list of flags
- `--pa --generate-shell-completion`, but this one is empty

## Observed behavior

No autocompletions given

## Expected behavior

Correctly complete autocompletion of partial flags

## Additional context

I worked around this by using a bash wrapper around the `command` binary, which rewrites calls in this form `command --XXX  --generate-shell-completion` to `command -  --generate-shell-completion`, which seems to work for bash/zsh.
Even though the command returns all possible flags, it's completion suggestions are filtered to relevant completions only.

## Want to fix this yourself?

We'd love to have more contributors on this project! If the fix for
this bug is easily explained and very small, feel free to create a
pull request for it.

## Run `go version` and paste its output here

```
go version go1.25.3 darwin/arm64
```

## Run `go env` and paste its output here

```
Rather not, but available upon requests. Very simple default goenv setup. Go version 1.24.4
```


## Issue comments (verbatim)

@fzlzjerry:

On current main (`5081218`), `app sub --pa --generate-shell-completion` already returns the expected flags. There is still a failing case after a positional argument: `app sub value --pa --generate-shell-completion` returns the help-command suggestion instead of matching flags.

I'd like to take the remaining case. The default completer indexes the child's parsed arguments as though they still include the trailing completion marker. I'll keep the request format, custom completion handlers, and behavior after `--` unchanged, and add regression coverage for root, child, and nested commands.

