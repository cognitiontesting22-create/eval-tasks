# Grading criteria: cli-2421

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `root mid leaf --help` still omits `--mid-persistent` (with its usage text) from the `GLOBAL OPTIONS:` section, or drops `--root-persistent` from it.
- Persistent flags from deeper lineages (root → a → b → c → leaf) are not all shown in the leaf's help.
- Ancestor flags marked `Local: true`, or hidden ancestor flags, appear in a descendant's help.
- The command's own flags are listed under `GLOBAL OPTIONS:` or duplicated between the two sections.
- Shadowing is not honoured: when root and `mid` both define `--shared`, `leaf`'s help must show only `mid`'s definition (and `mid`'s own help must show only its own definition, not root's); when `leaf` itself defines a flag with the same name as an ancestor's persistent flag, the ancestor's copy must not be listed. No flag name may appear twice in one help output.
- The heading is renamed (e.g. to `INHERITED OPTIONS:`) or the help layout otherwise changes for existing cases (root help, single-level children).
- Flag parsing behavior changes.
- No tests were added for the new behavior.
- The existing suite (`go test ./...`) is not green, or `go vet` / gofmt problems are introduced.
- Changes far outside the scope of the request.

## Nice to have (optional; do not block on these)

- Reusing the same lookup the parser uses to decide which definition of a name wins, so help and parsing can never disagree.
- A release note line describing the fix.

## Judgment calls

- The order in which inherited flags are listed (root first vs nearest ancestor first) is not specified; any consistent order is acceptable.
- How shadowing is detected (by name lookup, by set of names, etc.) is up to the implementer.
