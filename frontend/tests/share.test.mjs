import test from 'node:test'
import assert from 'node:assert/strict'
import vm from 'node:vm'
import fs from 'node:fs'
import path from 'node:path'
import { build } from 'esbuild'

const result = await build({entryPoints:['src/share.ts'],bundle:true,platform:'node',format:'cjs',write:false,define:{'import.meta.env':'{"BASE_URL":"/"}'}})
const module = {exports:{}}
vm.runInNewContext(result.outputFiles[0].text,{module,exports:module.exports,URLSearchParams,btoa,TextEncoder,location:{origin:'http://localhost'}})
const share = module.exports
const node = {id:'node',protocol:'vless',enabled:true,host:'example.com',port:23456,security:'none',transport:'tcp'}

test('special password and IPv6 URI round-trip preserves endpoint and name',()=>{
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea',password:`&=+,\'" /中文😀:@`}
 const ipv6={...node,protocol:'trojan',host:'2001:db8::1',name:'测试 节点 #1',security:'tls'}
 const parsed=new URL(share.buildTrojanLink(ipv6,client))
 assert.equal(parsed.hostname,'[2001:db8::1]')
 assert.equal(decodeURIComponent(parsed.username),client.password)
 assert.equal(decodeURIComponent(parsed.hash.slice(1)),ipv6.name)
 assert.equal(new URL(share.buildVlessLink({...ipv6,protocol:'vless'},client)).hostname,'[2001:db8::1]')
})

test('display names use numeric duplicate suffixes without short ids',()=>{
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea'}
 const nodes=Array.from({length:12},(_,i)=>({...node,id:`node-${i.toString().padStart(2,'0')}`,name:'重复名称'}))
 const config=JSON.parse(share.buildSingBoxConfig(nodes,client))
 const tags=config.outbounds.filter(o=>o.type!=='selector'&&o.type!=='direct').map(o=>o.tag)
 assert.equal(new Set(tags).size,12)
 assert.deepEqual(tags,nodes.map((_,i)=>i===0?'重复名称':`重复名称 ${i+1}`))
 const yaml=share.buildClashMetaConfig(nodes,client)
 const names=[...yaml.matchAll(/^  - name: "([^"]+)"$/gm)].map(m=>m[1])
 assert.equal(new Set(names).size,12)
 assert.deepEqual(names,tags)
})

test('display names do not vary with client username',()=>{
 const sample={...node,id:'node-de00',name:'重复名称'}
 const first=JSON.parse(share.buildSingBoxConfig([sample],{uuid:'a',username:'alice'}))
 const second=JSON.parse(share.buildSingBoxConfig([sample],{uuid:'b',username:'bob'}))
 const tag=config=>config.outbounds.find(out=>out.type==='vless')?.tag
 assert.equal(tag(first), '重复名称')
 assert.equal(tag(first), tag(second))
 assert.equal(tag(first).includes('alice'), false)
 assert.equal(tag(first).includes('bob'), false)
})

test('clean original names and numeric duplicates agree across all four export paths',()=>{
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea',password:'fixture-password'}
 for(const [input,want] of [
  [['美国','日本','韩国'],['美国','日本','韩国']],
  [['美国','美国','美国'],['美国','美国 2','美国 3']],
  [[' 美国 ','','   '],['美国','example.com:23456','example.com:23456 2']],
 ]) {
  const nodes=input.map((name,i)=>({...node,id:`node-${i}`,name}))
  const snapshot=JSON.stringify(nodes)
  const tags=JSON.parse(share.buildSingBoxConfig(nodes,client)).outbounds.filter(o=>o.type==='vless').map(o=>o.tag)
  const names=[...share.buildClashMetaConfig(nodes,client).matchAll(/^  - name: "([^"]+)"$/gm)].map(m=>m[1])
  const fragments=share.buildClientMultiShare(nodes,client,'v2rayn').split('\n').map(link=>decodeURIComponent(new URL(link).hash.slice(1)))
  const ps=share.buildClientMultiShare(nodes.map(n=>({...n,protocol:'vmess'})),client,'v2rayn').split('\n').map(link=>JSON.parse(Buffer.from(link.slice(8),'base64').toString()).ps)
  const trojan=share.buildClientMultiShare(nodes.map(n=>({...n,protocol:'trojan'})),client,'shadowrocket').split('\n').map(link=>decodeURIComponent(new URL(link).hash.slice(1)))
  for(const values of [tags,names,fragments,ps,trojan]) {
   assert.deepEqual(values,want)
   assert.equal(new Set(values).size,want.length)
   assert.equal(values.some(name=>name.includes('#')),false)
  }
  assert.equal(JSON.stringify(nodes),snapshot)
  console.log('TASK11 exported names:',JSON.stringify({tags,names,fragments,ps}))
 }
})

test('unnamed node uses host and actual port rather than remark or ID',()=>{
 const sample={...node,id:'node-de00',name:'',remark:'not a display name',port:0}
 const client={uuid:'fixture'}
 assert.equal(decodeURIComponent(new URL(share.buildVlessLink(sample,client)).hash.slice(1)),'example.com:0')
 assert.equal(JSON.parse(Buffer.from(share.buildVmessLink({...sample,protocol:'vmess'},client).slice(8),'base64').toString()).ps,'example.com:0')
})

test('VMess network and default WS/gRPC transport parameters survive export',()=>{
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea'}
 for(const transport of ['ws','grpc']) {
  const sample={...node,protocol:'vmess',transport}
  const yaml=share.buildClashMetaConfig([sample],client)
  assert.equal(yaml.includes(`network: "${transport}"`),true)
  assert.equal(yaml.includes(transport==='grpc'?'grpc-service-name: "yuchen"':'path: "/yuchen"'),true)
  const obj=JSON.parse(Buffer.from(share.buildVmessLink(sample,client).slice(8),'base64').toString())
  assert.equal(obj.net,transport)
  assert.equal(obj.path,transport==='grpc'?'yuchen':'/yuchen')
  const uri=new URL(share.buildVlessLink({...sample,protocol:'vless'},client))
  assert.equal(uri.searchParams.get(transport==='grpc'?'serviceName':'path'),obj.path)
 }
})

test('sing-box DNS is modern, selector choices valid, tags unique even when empty',()=>{
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea'}
 for(const nodes of [[],[{...node,name:'direct'},{...node,id:'other',name:'direct'}]]) {
  const config=JSON.parse(share.buildSingBoxConfig(nodes,client,{sing_box_include_quad9:true}))
  for(const server of config.dns.servers) {
   assert.equal(server.type,'udp');assert.equal('address' in server,false);assert.equal(typeof server.server,'string')
  }
  assert.equal(config.dns.servers.length,3)
  const tags=config.outbounds.map(out=>out.tag)
  assert.equal(new Set(tags).size,tags.length)
  const selector=config.outbounds.find(out=>out.type==='selector')
  assert.equal(selector.outbounds.includes(selector.default),true)
  for(const tag of selector.outbounds)assert.equal(tags.includes(tag),true)
  assert.equal(config.route.default_domain_resolver,'cloudflare')
 }
})

test('subscription format URLs are explicit and token path is encoded',()=>{
 const client={subscribe_token:'fixture token'}
 assert.equal(share.buildSubscriptionUrl(client),'http://localhost/sub/fixture%20token')
 assert.equal(share.buildSubscriptionUrl(client,'singbox'),'http://localhost/sub/fixture%20token?format=singbox')
 assert.equal(share.buildSubscriptionUrl(client,'singbox',true),'http://localhost/sub/fixture%20token?format=singbox&profile=tun')
})

test('optional VPN profile includes IPv4/IPv6 TUN, DNS hijack and loop prevention',()=>{
 const config=JSON.parse(share.buildSingBoxConfig([node],{uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea'},{client_tun:true}))
 const tun=config.inbounds.find(inbound=>inbound.type==='tun')
 assert.equal(tun.mtu,1400)
 assert.equal(tun.auto_route,true)
 assert.equal(tun.address.length,2)
 assert.equal(config.route.auto_detect_interface,true)
 assert.equal(config.route.rules[0].action,'hijack-dns')
})

test('empty, disabled, expired and fixed-exit clients have no ordinary node grant',()=>{
 for(const client of [
  {enabled:true,node_ids:[]},
  {enabled:true,node_ids:[],relay_route_ids:['relay']},
  {enabled:false,node_ids:['node']},
  {enabled:true,node_ids:['node'],expire_at:'2000-01-01T00:00:00Z'},
 ]) assert.equal(share.clientCanUseNode(client,node),false)
 assert.equal(share.clientCanUseNode({enabled:true,node_ids:['node']},node),true)
 assert.equal(share.clientCanUseNode({enabled:true,node_ids:['node'],expire_at:'0001-01-01T00:00:00Z'},node),true)
 assert.equal(share.clientCanUseNode({enabled:true,node_ids:['node'],expire_at:'invalid'},node),false)
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

// Optional fixture export for real config-only core checks. Uses documentation
// addresses from the Go fixture; never starts a proxy or contacts a node.
if(process.env.YC_CLIENT_FIXTURES_DIR) {
 const dir=process.env.YC_CLIENT_FIXTURES_DIR
 const db=JSON.parse(fs.readFileSync(path.join(dir,'fixture-db.json'),'utf8'))
 const client=Object.values(db.clients)[0]
 const nodes=share.nodesForClient(client,Object.values(db.nodes))
 for(const [file,tun] of [['frontend-singbox.json',false],['frontend-singbox-tun.json',true]]) {
  fs.writeFileSync(path.join(dir,file),share.buildSingBoxConfig(nodes,client,{client_tun:tun}),{mode:0o600})
 }
}
