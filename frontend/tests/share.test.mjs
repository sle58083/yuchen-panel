import test from 'node:test'
import assert from 'node:assert/strict'
import vm from 'node:vm'
import { build } from 'esbuild'

const result = await build({entryPoints:['src/share.ts'],bundle:true,platform:'node',format:'cjs',write:false,define:{'import.meta.env':'{"BASE_URL":"/"}'}})
const module = {exports:{}}
vm.runInNewContext(result.outputFiles[0].text,{module,exports:module.exports,URLSearchParams,btoa,TextEncoder,location:{origin:'http://localhost'}})
const share = module.exports
const node = {id:'node',protocol:'vless',enabled:true,host:'example.com',port:23456,security:'none',transport:'tcp'}

test('empty, disabled, expired and fixed-exit clients have no ordinary node grant',()=>{
 for(const client of [
  {enabled:true,node_ids:[]},
  {enabled:true,node_ids:[],relay_route_ids:['relay']},
  {enabled:false,node_ids:['node']},
  {enabled:true,node_ids:['node'],expire_at:'2000-01-01T00:00:00Z'},
 ]) assert.equal(share.clientCanUseNode(client,node),false)
 assert.equal(share.clientCanUseNode({enabled:true,node_ids:['node']},node),true)
})

test('retired protocols cannot be shared in any format',()=>{
 const retired={...node,protocol:'shadowsocks'}
 assert.equal(share.isClientShareNode(retired),false)
 for(const format of ['v2rayn','shadowrocket','clash','singbox']) {
  assert.throws(()=>share.buildClientShare(retired,{uuid:'test'},format))
  const config=share.buildClientMultiShare([retired],{enabled:true,node_ids:['node'],uuid:'test'},format)
  assert.equal(/shadowsocks|(?:^|\n)ss:\/\//i.test(config),false)
 }
})
