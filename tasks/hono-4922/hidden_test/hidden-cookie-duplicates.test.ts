import { Hono } from './hono'
import { getCookie, getSignedCookie } from './helper/cookie'
import { parse, parseSigned, serializeSigned } from './utils/cookie'

const secret = 'hidden-test-secret'

const probe = (header: string) => {
  const app = new Hono()
  app.get('/probe', (c) =>
    c.json({
      bulk: getCookie(c),
      a: getCookie(c, 'a'),
      session: getCookie(c, 'session'),
      toString: getCookie(c, 'toString'),
    })
  )
  return app.request('http://x/probe', { headers: { Cookie: header } })
}

const signedValue = async (name: string, value: string) => {
  // serializeSigned returns "name=value.signature"; keep only the value part
  const serialized = await serializeSigned(name, value, secret)
  return serialized.slice(name.length + 1)
}

describe('duplicate cookie names: getCookie helper', () => {
  it('bulk and key forms agree and return the first value', async () => {
    const res = await probe('a=first; a=last')
    const json = await res.json()
    expect(json.a).toBe('first')
    expect(json.bulk.a).toBe('first')
  })

  it('first-wins with other cookies in between', async () => {
    const res = await probe('session=legit; other=x; session=evil')
    const json = await res.json()
    expect(json.session).toBe('legit')
    expect(json.bulk.session).toBe('legit')
    expect(json.bulk.other).toBe('x')
  })

  it('first-wins across three duplicates', async () => {
    const res = await probe('a=1; a=2; a=3')
    const json = await res.json()
    expect(json.a).toBe('1')
    expect(json.bulk.a).toBe('1')
  })

  it('handles cookie names that collide with Object.prototype members', async () => {
    const res = await probe('toString=foo; hasOwnProperty=bar; constructor=baz; a=x')
    const json = await res.json()
    expect(json.toString).toBe('foo')
    expect(json.bulk.toString).toBe('foo')
    expect(json.bulk.hasOwnProperty).toBe('bar')
    expect(json.bulk.constructor).toBe('baz')
    expect(json.a).toBe('x')
  })

  it('first-wins for duplicated prototype-colliding names', async () => {
    const res = await probe('toString=first; toString=last')
    const json = await res.json()
    expect(json.toString).toBe('first')
    expect(json.bulk.toString).toBe('first')
  })

  it('does not alter single-cookie behavior', async () => {
    const res = await probe('a=only; b=%20spaced%20; c="quoted"')
    const json = await res.json()
    expect(json.a).toBe('only')
    expect(json.bulk).toEqual({ a: 'only', b: ' spaced ', c: 'quoted' })
  })
})

describe('duplicate cookie names: parse utility', () => {
  it('bulk parse is first-wins', () => {
    expect(parse('a=first; a=last')['a']).toBe('first')
    expect(parse('x=1; a=first; y=2; a=last')).toEqual({ x: '1', a: 'first', y: '2' })
  })

  it('named parse is first-wins and agrees with bulk parse', () => {
    const header = 'session=legit; other=x; session=evil'
    expect(parse(header, 'session')['session']).toBe('legit')
    expect(parse(header, 'session')['session']).toBe(parse(header)['session'])
  })

  it('named parse still only returns the requested key', () => {
    const result = parse('a=1; b=2; a=3', 'a')
    expect(result['a']).toBe('1')
    expect(result['b']).toBeUndefined()
  })

  it('parses prototype-colliding names in both forms', () => {
    expect(parse('toString=foo; constructor=baz')['toString']).toBe('foo')
    expect(parse('toString=foo; constructor=baz')['constructor']).toBe('baz')
    expect(parse('toString=foo', 'toString')['toString']).toBe('foo')
    expect(parse('toString=first; toString=last')['toString']).toBe('first')
    expect(parse('toString=first; toString=last', 'toString')['toString']).toBe('first')
  })

  it('percent-decodes the first value', () => {
    expect(parse('a=hello%20world; a=other')['a']).toBe('hello world')
    expect(parse('a=hello%20world; a=other', 'a')['a']).toBe('hello world')
  })
})

describe('duplicate cookie names: signed cookies', () => {
  it('parseSigned is first-wins in bulk and named forms', async () => {
    const first = await signedValue('yummy', 'choco')
    const second = await signedValue('yummy', 'strawberry')
    const header = `yummy=${first}; yummy=${second}`

    expect((await parseSigned(header, secret))['yummy']).toBe('choco')
    expect((await parseSigned(header, secret, 'yummy'))['yummy']).toBe('choco')
  })

  it('a tampered first occurrence yields false rather than falling through to the second', async () => {
    const good = await signedValue('yummy', 'strawberry')
    const header = `yummy=tampered.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=; yummy=${good}`

    expect((await parseSigned(header, secret))['yummy']).toBe(false)
    expect((await parseSigned(header, secret, 'yummy'))['yummy']).toBe(false)
  })

  it('getSignedCookie helper agrees between bulk and key forms', async () => {
    const first = await signedValue('yummy', 'choco')
    const second = await signedValue('yummy', 'strawberry')
    const app = new Hono()
    app.get('/s', async (c) => {
      const all = await getSignedCookie(c, secret)
      const byKey = await getSignedCookie(c, secret, 'yummy')
      return c.json({ all: all['yummy'], byKey })
    })
    const res = await app.request('http://x/s', {
      headers: { Cookie: `yummy=${first}; yummy=${second}` },
    })
    const json = await res.json()
    expect(json.byKey).toBe('choco')
    expect(json.all).toBe('choco')
  })

  it('parses signed cookies with prototype-colliding names', async () => {
    const v = await signedValue('toString', 'choco')
    expect((await parseSigned(`toString=${v}`, secret))['toString']).toBe('choco')
    expect((await parseSigned(`toString=${v}`, secret, 'toString'))['toString']).toBe('choco')
  })
})
