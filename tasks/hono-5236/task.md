# Suffix wildcard routes (`/foo*`) return 404 when a sub-app has static + `:param` routes with the same method

## Expected behavior

A suffix wildcard route like `/assets*` should match `/assets/app.js` regardless of what else is registered on the app.

## Actual behavior

When a mounted sub-app registers a static route and a `:param` route as siblings (same method), every suffix wildcard route (`/assets*`) on the parent app silently returns 404. Slash-star (`/assets/*`) is unaffected.

## Minimal reproduction

```ts
import { Hono } from 'hono'

// without the sub-app the same wildcard returns 200
const app = new Hono()
const sub = new Hono()
sub.post('/items', (c) => c.text('items'))
sub.post('/:slug', (c) => c.text('slug'))
app.route('/api', sub)
app.get('/assets*', (c) => c.text('ok'))
console.log((await app.request('http://x/assets/app.js')).status) // 404
```

Removing either `/items` or `/:slug` makes it work again; different methods (`get('/items')` + `post('/:slug')`) is fine; `{slug}` instead of `:slug` also fixes it.

## Wanted

Every router an app can end up using (including the one Hono falls back to when the faster ones cannot handle a route set) must support a trailing `*` glued to the last path segment, with the meaning the other routers already give it:

- `/assets*` matches `/assets`, `/assets-v2` and `/assets/app.js`, but not `/asset`.
- The text before the `*` is matched literally, so `/file.+*` matches `/file.+js` but not `/fileZZjs`.
- It composes with parameters, e.g. `/users/:id/avatar*` matches `/users/42/avatar.png` with `id === '42'`.
- Registering `/assets*` still works when a longer route such as `/assets*/x` was registered first.
- Existing behavior of `/*`, `/foo/*` and `:param` routes is unchanged, and the sub-app repro above returns 200 `ok`.

Please include tests for the new behavior. The existing test suite must stay green:

```
bun install --frozen-lockfile && bun run test
```
