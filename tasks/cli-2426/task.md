# Auto-completion with unfinished partial flags (`--XXX<TAB>`) returns no completions after a positional argument

When passing in uncompleted flags (`--XXX<TAB>`), no auto-completion is available in some situations.

## To reproduce

In a project with shell completion enabled (`EnableShellCompletion: true`) and a subcommand `subcommand` that has flags `--par1`, `--par2`, `--flag` and takes positional arguments:

```
$ command <TAB>
subcommand
$ command subcommand -<TAB>
--par1
--par2
--flag
$ command subcommand --pa<TAB>
--par1
--par2
$ command subcommand value --pa<TAB>
* empty / suggests subcommands instead of flags
```

Underlying, the shell passes the completion marker to the binary: `command subcommand value --pa --generate-shell-completion`. Without a positional argument (`command subcommand --pa --generate-shell-completion`) the matching flags are printed. As soon as there is one or more positional arguments before the partial flag, the child command prints its subcommand suggestions (e.g. `help`) instead of the matching flags. The same thing happens for nested commands (`command sub nested value --pa<TAB>`).

## Expected behavior

Correctly complete partial flags regardless of how many positional arguments precede them, at any command depth:

- `command sub value --pa<TAB>` and `command sub first second --pa<TAB>` list the `sub` flags starting with `pa`, one per line, exactly as the no-positional case does today (hidden flags excluded; a single `-` lists all visible flags).
- Same for nested commands (`command sub nested value --pa<TAB>` lists `nested`'s matching flags).
- Root-level completion (`command --pa<TAB>`, `command value --pa<TAB>`) keeps working as today.
- Completion requests after `--` still produce no suggestions, a custom `ShellComplete` handler still receives the arguments (`value --pa`) unchanged, and with completion disabled an unknown `--pa` flag is still an error.
- Completion never runs the command's `Action`.

Please include tests for the new behavior. The existing test suite must stay green:

```
go test ./...
```
