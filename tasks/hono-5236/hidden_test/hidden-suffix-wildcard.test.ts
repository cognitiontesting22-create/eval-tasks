import { Hono } from './hono'
import { TrieRouter } from './router/trie-router'

const status = async (app: Hono, path: string) => (await app.request('http://x' + path)).status
const body = async (app: Hono, path: string) => {
  const res = await app.request('http://x' + path)
  return `${res.status}:${await res.text()}`
}

describe('suffix wildcard routes with the default router', () => {
  it('matches after a sub-app registers static + :param siblings (issue repro)', async () => {
    const app = new Hono()
    const sub = new Hono()
    sub.post('/items', (c) => c.text('items'))
    sub.post('/:slug', (c) => c.text('slug'))
    app.route('/api', sub)
    app.get('/assets*', (c) => c.text('ok'))

    expect(await body(app, '/assets/app.js')).toBe('200:ok')
    expect(await body(app, '/assets')).toBe('200:ok')
    expect(await body(app, '/assets-v2')).toBe('200:ok')
    expect(await status(app, '/asset')).toBe(404)

    const items = await app.request('http://x/api/items', { method: 'POST' })
    expect(await items.text()).toBe('items')
    const slug = await app.request('http://x/api/other', { method: 'POST' })
    expect(await slug.text()).toBe('slug')
  })

  it('still works without the sub-app (control)', async () => {
    const app = new Hono()
    app.get('/assets*', (c) => c.text('ok'))
    expect(await body(app, '/assets/app.js')).toBe('200:ok')
  })

  it('matches when the sub-app conflict is registered after the wildcard', async () => {
    const app = new Hono()
    app.get('/static*', (c) => c.text('static'))
    const sub = new Hono()
    sub.put('/fixed', (c) => c.text('fixed'))
    sub.put('/:name', (c) => c.text('name'))
    app.route('/v1', sub)

    expect(await body(app, '/static/css/site.css')).toBe('200:static')
  })
})

describe('suffix wildcard routes with TrieRouter', () => {
  const make = () => new Hono({ router: new TrieRouter() })

  it('matches the prefix itself, a longer segment and deeper paths', async () => {
    const app = make()
    app.get('/assets*', (c) => c.text('assets'))
    expect(await body(app, '/assets')).toBe('200:assets')
    expect(await body(app, '/assets-v2')).toBe('200:assets')
    expect(await body(app, '/assets/app.js')).toBe('200:assets')
    expect(await body(app, '/assets/deep/er/file.js')).toBe('200:assets')
  })

  it('does not match a shorter prefix or an unrelated path', async () => {
    const app = make()
    app.get('/assets*', (c) => c.text('assets'))
    expect(await status(app, '/asset')).toBe(404)
    expect(await status(app, '/ass')).toBe(404)
    expect(await status(app, '/other/assets')).toBe(404)
  })

  it('treats regular expression characters in the prefix literally', async () => {
    const app = make()
    app.get('/file.+*', (c) => c.text('file'))
    expect(await body(app, '/file.+js')).toBe('200:file')
    expect(await body(app, '/file.+')).toBe('200:file')
    expect(await status(app, '/fileZZjs')).toBe(404)
    expect(await status(app, '/filejs')).toBe(404)
  })

  it('composes with path parameters', async () => {
    const app = make()
    app.get('/users/:id/avatar*', (c) => c.text(`avatar ${c.req.param('id')}`))
    expect(await body(app, '/users/42/avatar.png')).toBe('200:avatar 42')
    expect(await body(app, '/users/42/avatar')).toBe('200:avatar 42')
    expect(await status(app, '/users/42/photo.png')).toBe(404)
  })

  it('is only a wildcard on the terminal segment; a middle * remains a full segment wildcard', async () => {
    const app = make()
    app.get('/a/*/c', (c) => c.text('mid'))
    expect(await body(app, '/a/b/c')).toBe('200:mid')
    expect(await status(app, '/a/b/d')).toBe(404)
  })

  it('registers the pattern when a longer route with the same segment already exists', async () => {
    const app = make()
    app.get('/docs*/x', (c) => c.text('literal'))
    app.get('/docs*', (c) => c.text('docs'))
    expect(await body(app, '/docs/a/b')).toBe('200:docs')
    expect(await body(app, '/docs-old')).toBe('200:docs')
  })

  it('respects the HTTP method', async () => {
    const app = make()
    app.get('/assets*', (c) => c.text('assets'))
    const res = await app.request('http://x/assets/app.js', { method: 'POST' })
    expect(res.status).toBe(404)
  })

  it('keeps ordinary wildcards and params working', async () => {
    const app = make()
    app.get('/foo/*', (c) => c.text('foo'))
    app.get('/bar/:id', (c) => c.text(`bar ${c.req.param('id')}`))
    app.get('/*', (c) => c.text('root'))
    expect(await body(app, '/anything')).toBe('200:root')
    expect(await body(app, '/foo/x/y')).toBe('200:foo')
    expect(await body(app, '/bar/7')).toBe('200:bar 7')
  })

  it('exposes the match through the router API directly', () => {
    const router = new TrieRouter<string>()
    router.add('GET', '/assets*', 'assets')
    router.add('GET', '/users/:id/avatar*', 'avatar')

    expect(router.match('GET', '/assets/app.js')[0].map(([h]) => h)).toEqual(['assets'])
    expect(router.match('GET', '/asset')[0]).toEqual([])
    const [[handler, params]] = router.match('GET', '/users/9/avatar.png')[0]
    expect(handler).toBe('avatar')
    expect(params).toEqual({ id: '9' })
  })
})
