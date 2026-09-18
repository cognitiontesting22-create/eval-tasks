# Requester's original words

Issue: https://github.com/honojs/hono/issues/5219 (opened by @jaoskrm)
Fix PR: https://github.com/honojs/hono/pull/5236 (by @usualoma), merged 2026-08-17T00:37:48Z

## Issue title

Suffix wildcard routes (/foo*) return 404 when a sub-app has static + :param routes with the same method

## Issue body (verbatim)

## Version

hono 4.12.29 and 4.13.1 (latest)

## Expected behavior

A suffix wildcard route like `/assets*` should match `/assets/app.js` regardless of what else is registered on the app.

## Actual behavior

When a mounted sub-app registers a static route and a `:param` route as siblings (same method), every suffix wildcard route (`/assets*`) on the parent app silently returns 404. Slash-star (`/assets/*`) is unaffected.

## Minimal reproduction

```ts
import { Hono } from 'hono'

// control: works without the sub-app
const control = new Hono()
control.get('/assets*', (c) => c.text('ok'))
console.log((await control.request('http://x/assets/app.js')).status) // 200

// bug: same wildcard, but sub-app has static + :param siblings
const app = new Hono()
const sub = new Hono()
sub.post('/items', (c) => c.text('items'))
sub.post('/:slug', (c) => c.text('slug'))
app.route('/api', sub)
app.get('/assets*', (c) => c.text('ok'))
console.log((await app.request('http://x/assets/app.js')).status) // 404
```

## What I tried

- Removing either `/items` or `/:slug` makes it work again
- Different methods (`get('/items')` + `post('/:slug')`) is fine
- `{slug}` instead of `:slug` also fixes it
- Confirmed on 4.12.29 and 4.13.1

## Related issues

#4975, #4623 — same router-fallback family, but those cover `/**` with param-vs-param conflicts; this is suffix-`*` with a static-vs-param conflict.


## Issue comments (verbatim)

(none)