# Persistent flags defined on non-root commands do not show up in help output of descendant commands

When a flag is persistent (`Local == false`, the default), it is not displayed in the help of a child command that inherits it. Only root's persistent flags appear under `GLOBAL OPTIONS:`. The flag is parsed as expected, it just isn't displayed.

## To reproduce

```go
cmd := &cli.Command{
	Name: "root",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "root-persistent", Usage: "persistent flag on root", Local: false},
	},
	Commands: []*cli.Command{{
		Name: "mid",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "mid-persistent", Usage: "persistent flag on mid", Local: false},
			&cli.StringFlag{Name: "mid-local", Usage: "local flag on mid", Local: true},
		},
		Commands: []*cli.Command{{
			Name:  "leaf",
			Flags: []cli.Flag{&cli.StringFlag{Name: "leaf-local", Usage: "local flag on leaf"}},
		}},
	}},
}
_ = cmd.Run(context.Background(), []string{"root", "mid", "leaf", "--help"})
```

## Observed behavior

```
OPTIONS:
   --leaf-local string  local flag on leaf
   --help, -h           show help

GLOBAL OPTIONS:
   --root-persistent string  persistent flag on root
```

`--mid-persistent` is absent even though `root mid leaf --mid-persistent X` parses.

## Expected behavior

All non-hidden persistent flags inherited from *any* ancestor (root, intermediate commands, at any depth) are displayed in the help output of the commands inheriting them, under the existing `GLOBAL OPTIONS:` heading (keep the heading as is). Specifically:

- Local (`Local: true`) and hidden ancestor flags stay out of it.
- The command's own flags stay under `OPTIONS:`, never duplicated under `GLOBAL OPTIONS:`.
- If a nearer command (including the command itself) defines a flag with the same name as an ancestor's persistent flag, only the nearest definition is shown; the shadowed ancestor definition must not appear, so no name is listed twice.
- Root help and flag parsing are unchanged.

Please include tests for the new behavior. The existing test suite must stay green:

```
go test ./...
```
