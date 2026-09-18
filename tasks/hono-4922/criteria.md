# Grading criteria: hono-4922

## Merge question

Would you merge this? Answer yes only if none of the blockers below apply.

## Blockers (any one of these makes the output unacceptable)

- `getCookie(c)['a']` and `getCookie(c, 'a')` still disagree for `Cookie: a=first; a=last`, or either returns `'last'` (the required rule is first-wins).
- First-wins does not hold with other cookies in between (`session=legit; other=x; session=evil` must give `legit` in both forms) or for more than two duplicates.
- The low-level `parse(cookie)` / `parse(cookie, name)` from `hono/utils/cookie` do not follow the same first-wins rule, or the named form starts returning keys other than the requested one.
- Signed cookies diverge: `getSignedCookie` / `parseSigned` bulk and named forms disagree, or they do not return the first occurrence (a first occurrence with a bad signature yields `false`; it must not fall through to a later duplicate).
- Cookies whose names collide with `Object.prototype` members (`toString`, `hasOwnProperty`, `constructor`) are dropped or return the prototype member instead of the cookie value in either form, or first-wins does not hold for them.
- Regression in existing parsing behavior: whitespace trimming, quoted values, percent-decoding, skipping pairs without `=`, or `__Secure-`/`__Host-` prefix handling.
- The fix picks last-wins (even consistently) — the request explicitly asks for first-wins.
- No tests were added for the new behavior.
- The existing suite (`bun install --frozen-lockfile && bun run test`) is not green (this includes the `tsc -p tsconfig.spec.json` step).
- Changes far outside the scope of the request.

## Nice to have (optional; do not block on these)

- Returning prototype-less objects from the bulk forms so that `getCookie(c).toString` is `undefined` when no such cookie exists.
- Tests at both the utility (`parse`/`parseSigned`) and helper (`getCookie`/`getSignedCookie`) level.
- A short note in docs/JSDoc stating the first-wins rule.

## Judgment calls

- Whether the named-lookup fast path (early return) is kept is an implementation detail.
- Behavior for a duplicate whose first occurrence has an *invalid value* (fails the cookie-value grammar) is not specified by the request; either skipping it in favor of the next valid occurrence or treating the name as absent is acceptable.
