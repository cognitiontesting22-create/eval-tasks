# Requester's original words

Issue: https://github.com/urfave/cli/issues/2250 (opened by @koct9i)
Fix PR: https://github.com/urfave/cli/pull/2363, merged 2026-06-14T13:21:27Z

## Issue title

Imbalance between Before and After calls at printing help

## Issue body (verbatim)

## My urfave/cli version is

_**3.6.2**_

## Checklist

- [x] Are you running the latest v3 release? The list of releases is [here](https://github.com/urfave/cli/releases).
- [x] Did you check the manual for your release? The v3 manual is [here](https://cli.urfave.org/v3/getting-started/)
- [x] Did you perform a search about this problem? Here's the [GitHub guide](https://help.github.com/en/github/managing-your-work-on-github/using-search-to-filter-issues-and-pull-requests) about searching.

## Dependency Management

- My project is using go modules.

## Describe the bug

For "help" action or "subcommand --help" "After" is called without calling "Before".

## To reproduce

Describe the steps or code required to reproduce the behavior

## Observed behavior

What did you see happen immediately after the reproduction steps
above?

## Expected behavior

Callback "After" is called iff "Before" was called in the past.

## Additional context

Add any other context about the problem here.

If the issue relates to a specific open source GitHub repo, please
link that repo here.

If you can reproduce this issue with a public CI system, please
link a failing build here.

## Want to fix this yourself?

We'd love to have more contributors on this project! If the fix for
this bug is easily explained and very small, feel free to create a
pull request for it.

## Run `go version` and paste its output here

```
# paste `go version` output in here
```

## Run `go env` and paste its output here

```
# paste `go env` output in here
```


## Issue comments (verbatim)

@dearchap:

@koct9i Thanks for raising this issue. Yes there seems to be a lot of inconsistency in the callbacks depending on if ```--help``` is used or ```help``` . I'll took a quick look and the fix doesnt seem to be that easy. Will think about it a bit. 
