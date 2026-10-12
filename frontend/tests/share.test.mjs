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
 const client={uuid:'f9b77af1-736a-4f5b-8bc5-273ba9b552ea',password:'p@ss w#rd&test=1'}
 const ipv6={...node,protocol:'trojan',host:'2001:db8::1',name:'测试 节点 #1',security:'tls'}
 const parsed=new URL(share.buildTrojanLink(ipv6,client))
 assert.equal(parsed.hostname,'[2001:db8::1]')
 assert.equal(decodeURIComponent(parsed.username),client.password)
 assert.equal(decodeURIComponent(parsed.hash.slice(1)),ipv6.name)
 assert.equal(new URL(share.buildVlessLink({...ipv6,protocol:'vless'},client)).hostname,'[2001:db8::1]')
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
