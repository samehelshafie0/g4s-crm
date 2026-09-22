import test from 'node:test'
import assert from 'node:assert/strict'
import { AxiosError } from 'axios'
const values = new Map()
globalThis.localStorage = { getItem: key => values.get(key) ?? null, setItem: (key, value) => values.set(key, value), removeItem: key => values.delete(key) }
let redirects = 0
globalThis.window = { location: { pathname: '/customers', replace: () => { redirects++ } } }
const { http, refreshHttp, tokenStorage } = await import('../src/services/http.ts')
function unauthorized(config) { throw new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, null, { status: 401, config, data: {}, headers: {}, statusText: 'Unauthorized' }) }
function ok(config, data) { return { status: 200, config, data, headers: {}, statusText: 'OK' } }
function reset() { values.clear(); redirects = 0; tokenStorage.setTokens('old-access', 'old-refresh') }

test('concurrent expired requests share refresh and retry with the new token', async () => {
 reset(); let refreshes = 0
 refreshHttp.defaults.adapter = async config => { refreshes++; await new Promise(resolve => setTimeout(resolve, 10)); return ok(config, { success: true, data: { accessToken: 'new-access', refreshToken: 'new-refresh' } }) }
 http.defaults.adapter = async config => config.headers.Authorization === 'Bearer new-access' ? ok(config, 'loaded') : unauthorized(config)
 const results = await Promise.all([http.get('/customers'), http.get('/quotes'), http.get('/auth/me')])
 assert.equal(refreshes, 1); assert.equal(results.length, 3); assert.equal(redirects, 0)
})
test('a retried 401 stops after one refresh', async () => {
 reset(); let refreshes = 0; let requests = 0
 refreshHttp.defaults.adapter = async config => { refreshes++; return ok(config, { success: true, data: { accessToken: 'new', refreshToken: 'new-refresh' } }) }
 http.defaults.adapter = async config => { requests++; return unauthorized(config) }
 await assert.rejects(http.get('/customers')); assert.equal(refreshes, 1); assert.equal(requests, 2); assert.equal(tokenStorage.getAccess(), null)
})
test('login rejection never refreshes an old session', async () => {
 reset(); let refreshes = 0
 refreshHttp.defaults.adapter = async config => { refreshes++; return unauthorized(config) }
 http.defaults.adapter = async config => unauthorized(config)
 await assert.rejects(http.post('/auth/login', {})); assert.equal(refreshes, 0); assert.equal(redirects, 0)
})
test('missing refresh does not leave the next login stuck', async () => {
 reset(); tokenStorage.clear(); http.defaults.adapter = async config => unauthorized(config)
 await assert.rejects(http.get('/customers'))
 tokenStorage.setTokens('old', 'restored-refresh')
 refreshHttp.defaults.adapter = async config => ok(config, { success: true, data: { accessToken: 'new', refreshToken: 'rotated' } })
 http.defaults.adapter = async config => config.headers.Authorization === 'Bearer new' ? ok(config, {}) : unauthorized(config)
 await http.get('/customers'); assert.equal(tokenStorage.getRefresh(), 'rotated')
})
test('logout during refresh cannot restore the expired session', async () => {
 reset()
 refreshHttp.defaults.adapter = async config => { tokenStorage.clear(); return ok(config, { success: true, data: { accessToken: 'new', refreshToken: 'rotated' } }) }
 http.defaults.adapter = async config => unauthorized(config)
 await assert.rejects(http.get('/customers')); assert.equal(tokenStorage.getAccess(), null)
})

test('a file upload is not sent with the instance JSON content type', async () => {
 // The shared instance defaults to JSON. Without stripping it for FormData the
 // multipart boundary is never generated and the server rejects the body.
 // The adapter sees the config after every request interceptor has run.
 const { http } = await import('../src/services/http.ts')
 const sentFor = async (body) => {
  let captured = null
  await http.post('/upload-probe', body, {
   adapter: async config => { captured = config; throw new Error('stop') },
  }).catch(() => {})
  assert.ok(captured, 'request was not built')
  return captured.headers['Content-Type']
 }

 // What the transport then chooses is its own business — in a browser it fills
 // in multipart with a boundary. The invariant here is that the instance's JSON
 // default never reaches a file upload, which is what breaks the boundary.
 const form = new FormData()
 form.append('file', new Blob(['x']), 'vendor.pdf')
 assert.notEqual(await sentFor(form), 'application/json', 'FormData must not carry the JSON content type')
 assert.equal(await sentFor({ a: 1 }), 'application/json', 'JSON bodies keep their content type')
})
