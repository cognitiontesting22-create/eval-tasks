# Requester's original words

Issue: https://github.com/urfave/cli/issues/2420 (opened by @0x0013)
Fix PR: https://github.com/urfave/cli/pull/2421, merged 2026-09-04T14:05:37Z

## Issue title

Persistent flags defined on non-root command, do not show up in help output of leaf command

## Issue body (verbatim)

## My urfave/cli version is

`v3.10.0` (also verified in `v3.11.0`)

## Checklist

- [X] Are you running the latest v3 release? The list of releases is [here](https://github.com/urfave/cli/releases).
- [X] Did you check the manual for your release? The v3 manual is [here](https://cli.urfave.org/v3/getting-started/)
- [X] Did you perform a search about this problem? Here's the [GitHub guide](https://help.github.com/en/github/managing-your-work-on-github/using-search-to-filter-issues-and-pull-requests) about searching.

## Dependency Management

<!--
  Delete any of the following that do not apply:
-->

- My project is using go modules.

## Describe the bug

When a flag is defined as persistent (`Local == false`), this flag is not displayed in the help output of a child command that inherits it. Only the persistent flags defined on root command are displayed under `GLOBAL OPTIONS:`. Note that the flag gets parsed as expected, just isn't displayed in the help output.

## To reproduce

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name: "root",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "root-persistent", Usage: "persistent flag on root", Local: false},
		},
		Commands: []*cli.Command{
			{
				Name: "mid",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "mid-persistent", Usage: "persistent flag on mid", Local: false},
					&cli.StringFlag{Name: "mid-local", Usage: "local flag on mid", Local: true},
				},
				Commands: []*cli.Command{
					{
						Name:  "leaf",
						Flags: []cli.Flag{&cli.StringFlag{Name: "leaf-local", Usage: "local flag on leaf"}},
						Action: func(_ context.Context, c *cli.Command) error {
							fmt.Println("leaf action; mid-persistent =", c.String("mid-persistent"))
							return nil
						},
					},
				},
			},
		},
	}

	fmt.Println("=========== root mid leaf --help ===========")
	_ = cmd.Run(context.Background(), []string{"root", "mid", "leaf", "--help"})
	fmt.Println("=========== root mid --help ===========")
	_ = cmd.Run(context.Background(), []string{"root", "mid", "--help"})
	fmt.Println("=========== parse check: root mid leaf --mid-persistent=X ===========")
	if err := cmd.Run(context.Background(), []string{"root", "mid", "leaf", "--mid-persistent", "X"}); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
}
```

## Observed behavior

Persistent flags defined on non-root commands are absent in the help of a child command inheriting said flag.

Example output where `--mid-persistent` flag is declared on `mid` subcommand:

```
NAME:
   root mid leaf

USAGE:
   root mid leaf [options]

OPTIONS:
   --leaf-local string  local flag on leaf
   --help, -h           show help

GLOBAL OPTIONS:
   --root-persistent string  persistent flag on root
```

## Expected behavior

All non-hidden inherited flags are displayed in help output of the commands inheriting them.

## Additional context

I assume this is a bug, not intended behavior.

The root cause seems to be in the `VisiblePersistentFlags()` method, which only loops over `cmd.root.Flags()` and [does not traverse the whole command chain](https://github.com/urfave/cli/blob/0ed457cf64de75990ce1779f0b9632796f1a4962/command.go#L309-L315).

This is in contrast to `ParseFlags()` method which [walks the full command lineage](https://github.com/urfave/cli/blob/0ed457cf64de75990ce1779f0b9632796f1a4962/command_parse.go#L33).

## Want to fix this yourself?

I'd be happy to make a contribution. The obvious solution would be to replicate similar loop as that in `ParseFlags()`. However, that would populate the persistent interim-command flags under the existing "GLOBAL OPTIONS:" header, which might not be entirely accurate for flags not defined on root command.

I think it would be more accurate to rename the heading to "INHERITED OPTIONS:", but would like confirmation whether this would be acceptable, and whether this is truly a bug and not intended behavior.



## Issue comments (verbatim)

(none)