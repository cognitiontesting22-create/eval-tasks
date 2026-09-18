# Grading criteria: hono-5236

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- The issue repro still returns 404: a parent app with `app.get('/assets*')` plus a mounted sub-app that has a static route and a `:param` route for the same method does not serve `/assets/app.js`.
- The fallback router (the one Hono uses when the faster routers reject a route set — currently `TrieRouter`) does not implement suffix wildcards with the same semantics as the other routers. Checked via `new Hono({ router: new TrieRouter() })` and via `TrieRouter#match` directly:
  - `/assets*` must match `/assets`, `/assets-v2`, `/assets/app.js` (and deeper), and must not match `/asset`.
  - The prefix is literal: `/file.+*` matches `/file.+js`, not `/fileZZjs`.
  - Combined with params: `/users/:id/avatar*` matches `/users/42/avatar.png` with `id === '42'`.
  - Registering `/assets*` after `/assets*/x` still makes `/assets*` match.
  - Method is still respected (a GET-only route does not answer POST).
- Regression in ordinary routing on the fallback router: `/*`, `/foo/*`, `:param`, or middle-segment `*` change behavior.
- The fix only special-cases the literal path from the issue, or "fixes" it by expanding `/foo*` into `/foo` + `/foo/*` (which misses `/foo-bar`).
- Fixing it by making the fast routers *not* fall back (so the failing combination throws or becomes unsupported) instead of making the fallback correct.
- No tests were added for the new behavior.
- The existing suite (`bun install --frozen-lockfile && bun run test`) is not green, including lint/typecheck steps that `bun run test` performs (`tsc -p tsconfig.spec.json`).
- Changes far outside the scope of the request (unrelated refactors, behavior changes in other routers or middleware).

## Nice to have (optional; do not block on these)

- Tests at both the router level (`TrieRouter`/`Node`) and the `Hono` app level.
- Keeping bundle-size growth minimal or offsetting it.
- Keeping the change confined to the fallback router rather than touching the shared routing utilities.

## Judgment calls

- Whether the change lands in the router class, its node/trie implementation, or shared path utilities is up to the implementer; only behavior is graded.
- Precedence between `/docs*/x` and `/docs*` for a path like `/docs-old/x` differs between existing routers and is not specified by the request; do not penalize either ordering.
- Removing dead code or an unused constructor in the touched router is acceptable scope, not a blocker.
