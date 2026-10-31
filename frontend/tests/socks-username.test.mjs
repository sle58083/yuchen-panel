import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'
import { webcrypto } from 'node:crypto'
import ts from 'typescript'

// Compile the actual form functions, without mounting the UI or issuing requests.
const source = fs.readFileSync('src/views/Nodes.vue', 'utf8')
const script = source.match(/<script setup lang="ts">([\s\S]*?)<\/script>/)[1]
const ast = ts.createSourceFile('Nodes.ts', script, ts.ScriptTarget.Latest, true, ts.ScriptKind.TS)
const names = new Set(['randomPassword', 'randomRecommendedPort', 'applySocks5', 'setMode', 'editNode'])
const selected = ast.statements.filter(statement =>
  (ts.isFunctionDeclaration(statement) && names.has(statement.name?.text)) ||
  (ts.isVariableStatement(statement) && statement.declarationList.declarations.some(d => d.name.getText(ast) === 'emptyForm'))
)
assert.equal(selected.length, 6, 'all six production form declarations must be exercised')
const compiled = ts.transpileModule(selected.map(s => s.getText(ast)).join('\n') +
  '\nglobalThis.subject = {emptyForm, setMode, applySocks5, editNode}',
  { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText

function harness() {
  const context = {
    crypto: webcrypto,
    DEFAULT_NODE_PRESET: 'anti',
    presetParameters: () => ({ protocol: 'vless', transport: 'tcp', security: 'reality' }),
    detectMode: () => 'socks5',
    form: { value: {} }, mode: { value: 'anti' }, showEditor: { value: false },
    editingId: { value: '' }, config: { value: '' }, window: { scrollTo() {} },
  }
  vm.runInNewContext(compiled, context)
  context.form.value = context.subject.emptyForm()
  return context
}

test('SOCKS5 mode generates a 12 character username and a 22 character password', async () => {
  const c = harness()
  assert.equal(c.form.value.socks_username, '')
  await c.subject.setMode('socks5')
  assert.equal(c.form.value.protocol, 'socks')
  assert.match(c.form.value.socks_username, /^[ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789]{12}$/)
  assert.equal(c.form.value.socks_password.length, 22)
})

test('SOCKS5 editing preserves yuchen and manual usernames', () => {
  for (const username of ['yuchen', 'custom-account']) {
    const c = harness()
    c.subject.editNode({ id: 'old', protocol: 'socks', socks_username: username, socks_password: 'old-password' })
    c.subject.applySocks5(false) // same preservation path used by the protocol watcher
    assert.equal(c.form.value.socks_username, username)
    assert.equal(c.form.value.socks_password, 'old-password')
  }
})

test('SOCKS5 empty legacy username is generated and explicit regeneration replaces it', () => {
  const c = harness()
  c.subject.editNode({ id: 'old', protocol: 'socks', socks_username: '', socks_password: 'old-password' })
  assert.equal(c.form.value.socks_username.length, 12)
  c.form.value.socks_username = 'manual-account'
  c.subject.applySocks5(true)
  assert.equal(c.form.value.socks_username.length, 12)
  assert.notEqual(c.form.value.socks_username, 'manual-account')
  assert.equal(c.form.value.socks_password.length, 22)
  assert.match(source, /@click="form.socks_username = randomPassword\(12\)">重新生成用户名/)
  assert.match(source, /v-model="form.socks_username" placeholder="自动生成或手动填写"/)
})

test('new gRPC nodes do not use the legacy fixed service name default', () => {
  assert.match(source, /form\.value\.transport === 'grpc'.*randomPassword\(12\)/)
  assert.match(source, /transport === 'grpc' && !form\.value\.path/)
})
