# Requester's original words

Issue: https://github.com/honojs/hono/issues/4916 (opened by @j-smz)
Fix PR: https://github.com/honojs/hono/pull/4922 (by @usualoma), merged 2026-05-16T09:17:36Z

## Issue title

bug: getCookie(c, name) and getCookie(c)[name] return different values for duplicate cookie names

## Issue body (verbatim)

### What version of Hono are you using?

4.12.18

### What runtime/platform is your app running on? (with version if possible)

Bun

### What steps can reproduce the bug?

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

export default app
```

Request:

```sh
curl http://localhost:3000/probe -H 'Cookie: a=first; a=last'
```

Actual response:

```json
{
  "raw": "a=first; a=last",
  "getCookie(c)[a]": "last",
  "getCookie(c, 'a')": "first"
}
```

The same `Cookie` header is interpreted as `last` by one call and `first` by the other.

Root cause is in `src/utils/cookie.ts` `parse(cookie, name?)`:

- When `name` is provided, the function `break`s out of the loop on the first match → first-wins.
- When `name` is omitted, the loop runs to completion and later pairs overwrite earlier ones in the returned object → last-wins.

Both forms are exposed via `getCookie(c, name)` and `getCookie(c)` in `src/helper/cookie/index.ts`, so both code paths are reachable from a normal application.


### What is the expected behavior?

Both call shapes should return the same value for the same `Cookie` header.

I prefer both last-wins (matches Express's `cookie`, Node `http.cookies`, and most browsers' DOM-side `document.cookie` enumeration order). Implementation: remove the early `break` in `parse(cookie, name)` so the loop always processes every pair.

### What do you see instead?

_No response_

### Additional information

_No response_

## Issue comments (verbatim)

@usualoma:

Hi @j-smz,
Thank you for creating this issue!

I think “first-wins” is better than “last-wins.”

A survey:

## What major frameworks actually do

### First-wins

**Express (via `jshttp/cookie`, the de-facto Node.js parser)** — explicit `// only assign once` in the source:
```js
// from https://github.com/jshttp/cookie/blob/master/src/index.ts
const key = valueSlice(str, index, eqIdx);

// only assign once
if (obj[key] === undefined) {
  obj[key] = dec(valueSlice(str, eqIdx + 1, endIdx));
}
```

**Go `net/http`** — `Request.Cookie(name)` returns the first match by short-circuiting (https://github.com/golang/go/blob/master/src/net/http/request.go, around line 531):
```go
// Cookie returns the named cookie provided in the request or
// [ErrNoCookie] if not found.
// If multiple cookies match the given name, only one cookie will
// be returned.
func (r *Request) Cookie(name string) (*Cookie, error) {
    if name == "" {
        return nil, ErrNoCookie
    }
    for _, c := range readCookies(r.Header, name) {
        return c, nil  // first match wins
    }
    return nil, ErrNoCookie
}
```

**Spring (`WebUtils.getCookie`)** — JavaDoc literally says "Retrieve the first cookie with the given name":
```java
public static @Nullable Cookie getCookie(HttpServletRequest request, String name) {
    Assert.notNull(request, "Request must not be null");
    Cookie[] cookies = request.getCookies();
    if (cookies != null) {
        for (Cookie cookie : cookies) {
            if (name.equals(cookie.getName())) {
                return cookie;  // first match wins
            }
        }
    }
    return null;
}
```

**unjs/`cookie-es`** (used by Nuxt, h3, Nitro) — JSDoc states the rule explicitly:
```ts
/**
 * Parse a `Cookie` header string into an object.
 * ...
 * First occurrence wins for duplicate names unless `allowMultiple` is set.
 */
```
https://github.com/unjs/cookie-es/blob/main/src/cookie/parse.ts#L15-L24

### Last-wins

There are also several frameworks that have been labeled as "Last-wins."

**Deno `std/http` `getCookies`** — plain object overwrite in a loop, no guard ([http/cookie.ts](https://github.com/denoland/std/blob/main/http/cookie.ts#L251)):
```ts
export function getCookies(headers: Headers): Record<string, string> {
  const cookie = headers.get("Cookie");
  const out: Record<string, string> = Object.create(null);
  if (cookie !== null) {
    const c = cookie.split(";");
    for (const kv of c) {
      const [cookieKey, ...cookieVal] = kv.split("=");
      // ...
      const key = cookieKey!.trim();
      out[key] = cookieVal.join("=");  // last write wins
    }
  }
  return out;
}
```

**Django `parse_cookie`** — same pattern, last-wins via dict assignment ([django/http/cookie.py](https://github.com/django/django/blob/main/django/http/cookie.py)):
```python
def parse_cookie(cookie):
    cookiedict = {}
    for chunk in cookie.split(";"):
        if "=" in chunk:
            key, val = chunk.split("=", 1)
        else:
            key, val = "", chunk
        key, val = key.strip(), val.strip()
        if key or val:
            cookiedict[key] = cookies._unquote(val)  # last write wins
    return cookiedict
```
Worth noting: Django ticket [#33212](https://code.djangoproject.com/ticket/33212) attempted to switch this to first-wins. In that discussion the original author of the parser wrote:
> "Django has historically always used the _last_ cookie if there are multiple, but I could see changing that to use the first one instead, as **most other server software seem to look at the _first_ cookie when there are multiple with the same name**."

The ticket was closed `wontfix` purely for backwards-compatibility reasons, not because last-wins was considered correct.

**Rust `cookie` crate `CookieJar`** (used by axum-extra, actix-web, Rocket) — uses a `HashSet` keyed by cookie name; the seeding method calls `HashSet::replace`, so adding a duplicate replaces the existing entry ([cookie-rs/src/jar.rs](https://github.com/rwf2/cookie-rs/blob/master/src/jar.rs)):
```rust
pub struct CookieJar {
    original_cookies: HashSet<DeltaCookie>,
    delta_cookies: HashSet<DeltaCookie>,
}

/// Adds an "original" `cookie` to this jar. If an original cookie with the
/// same name already exists, it is replaced with `cookie`. [...]
pub fn add_original<C: Into<Cookie<'static>>>(&mut self, cookie: C) {
    self.original_cookies.replace(DeltaCookie::added(cookie.into()));
}
```
The author of this crate (the same author as `cookie-rs`) later published a separate [`biscotti`](https://lpalmieri.com/posts/biscotti-http-cookies-in-rust/) crate precisely because `CookieJar` cannot represent duplicate cookie names — an acknowledgment that the last-wins behavior here is a limitation, not a deliberate convention.


## Why first-wins is the right default

Beyond the popularity argument:

1. **RFC 6265 §5.4** says user agents SHOULD send cookies "with longer paths listed before cookies with shorter paths." So the **first** cookie on the wire is the most specific one — typically the user's actual session cookie, not an inherited one from a parent domain.

2. **Security**: last-wins is the parsing direction exploited by cookie tossing and cookie injection attacks. An attacker who can plant a cookie at a parent domain or sibling path produces a `Cookie:` header where the legitimate, more-specific cookie comes first and the attacker's overrides come later. Last-wins picks the attacker's value. This is exactly why the Django ticket above exists and why PortSwigger's cookie manipulation guidance flags last-wins backends.

@usualoma:

@j-smz 

I couldn't verify this part from the sources I checked:

> matches Express's cookie, Node http.cookies, and most browsers' DOM-side document.cookie enumeration order

At least `cookie.parse()` used by Express-style stacks seems to be first-wins, and `document.cookie` feels like a different layer from server-side `Cookie` header parsing.

Could you share the code references for these comparisons? I may be missing something.


@yusukebe:

@j-smz @usualoma 

I also agree with the @usualoma opinion that "first-wins,"  and we can accept #4922
