# Imbalance between Before and After calls when a subcommand prints help

For `subcommand --help` (or `-h`), the parent command's `After` callback is called without `Before` ever having been called.

## To reproduce

```go
var log []string
app := &cli.Command{
	Name: "app",
	Before: func(ctx context.Context, _ *cli.Command) (context.Context, error) {
		log = append(log, "app.Before"); return ctx, nil
	},
	After: func(context.Context, *cli.Command) error {
		log = append(log, "app.After"); return nil
	},
	Commands: []*cli.Command{{
		Name:   "sub",
		Action: func(context.Context, *cli.Command) error { return nil },
	}},
}
_ = app.Run(context.Background(), []string{"app", "sub", "--help"})
fmt.Println(log)
```

## Observed behavior

```
[app.After]
```

Help for `sub` is printed and `app.After` runs though `app.Before` never ran. With a deeper tree (`app sub nested --help`) every ancestor's `After` fires (`[sub.After app.After]`). Same with `-h`, and with arguments after the flag (`app sub --help nested`).

## Expected behavior

Callback `After` is called iff `Before` was called in the past. So when a subcommand at any depth displays help via `--help`/`-h` (so `Before` never runs), none of its ancestors' `After` callbacks run either (the log above is empty), including for a command that defines `After` but no `Before`.

Everything else must stay as it is:

- A normal run keeps the current order: `app.Before sub.Before sub.Action sub.After app.After` (and the analogous order for nested commands).
- `After` still runs, paired with `Before`, when the `Action` returns an error.
- `app --help` at the root still runs no callbacks.
- The `help` command (`app help`, `app help sub`, `app sub help`) keeps its current behavior, where `Before`/`After` of the commands that actually ran stay paired.
- The help output itself is unchanged.

Please include tests for the new behavior. The existing test suite must stay green:

```
go test ./...
```
