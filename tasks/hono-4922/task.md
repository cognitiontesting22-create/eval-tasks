# `getCookie(c, name)` and `getCookie(c)[name]` return different values for duplicate cookie names

`getCookie(c, name)` and `getCookie(c)[name]` return different values when the request `Cookie` header contains the same name twice.

Minimal reproduction:

```ts
import { Hono } from 'hono'
import { getCookie } from 'hono/cookie'

const app = new Hono()

app.get('/probe', (c) => {
  const all = getCookie(c)              // bulk form
  const byKey = getCookie(c, 'a')       // key form
  return c.json({
    raw: c.req.raw.headers.get('Cookie'),
    'getCookie(c)[a]': all['a'],
    "getCookie(c, 'a')": byKey,
  })
})
```

```sh
curl http://localhost:3000/probe -H 'Cookie: a=first; a=last'
```

Actual response:

```json
{ "raw": "a=first; a=last", "getCookie(c)[a]": "last", "getCookie(c, 'a')": "first" }
```

The same `Cookie` header is interpreted as `last` by one call and `first` by the other.

## Expected behavior

Both call shapes should return the same value for the same `Cookie` header, and the rule should be **first-wins**: for `a=first; a=last`, both `getCookie(c)['a']` and `getCookie(c, 'a')` return `'first'`. This matches what Express (`jshttp/cookie`, "only assign once"), Go `net/http` `Request.Cookie`, Spring `WebUtils.getCookie` and unjs `cookie-es` do. The same applies to the low-level `parse(cookie)` / `parse(cookie, name)` helpers in `hono/utils/cookie`, and to `getSignedCookie` / `parseSigned` in both their bulk and named forms (the first occurrence is the one whose signature is checked and returned).

Cookies whose names happen to collide with `Object.prototype` members (e.g. `toString=foo; hasOwnProperty=bar; constructor=baz`) must still be parsed normally in both forms, including when duplicated (`toString=first; toString=last` gives `'first'`).

Everything else about cookie parsing (whitespace trimming, quoted values, percent-decoding, ignoring pairs without `=`, `__Secure-`/`__Host-` prefixes) must remain as it is.

Please include tests for the new behavior. The existing test suite must stay green:

```
bun install --frozen-lockfile && bun run test
```
