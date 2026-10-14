import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'
import vm from 'node:vm'

// Exercise the actual timer/lifecycle block in Clients.vue without a DOM.
const source=fs.readFileSync('src/views/Clients.vue','utf8')
const block=source.slice(source.indexOf('let refreshTimer:'),source.indexOf('</script>')).replace(': ReturnType<typeof setInterval> | undefined','')

test('Traffic polling uses 30 seconds, pauses during editing, rejects late responses, and cleans up',async()=>{
 let interval,unmount,cleared=0,calls=0,loads=0,resolve
 const context={showAdvancedClientForm:{value:false},showFixedClientModal:{value:false},clients:{value:['original']},load:()=>{loads++},onMounted:fn=>fn(),onBeforeUnmount:fn=>{unmount=fn},setInterval:(fn,ms)=>{assert.equal(ms,30000);interval=fn;return 7},clearInterval:id=>{assert.equal(id,7);cleared++},api:()=>{calls++;return new Promise(r=>{resolve=r})}}
 vm.runInNewContext(block,context)
 assert.equal(loads,1)
 context.showAdvancedClientForm.value=true;await interval();assert.equal(calls,0)
 context.showAdvancedClientForm.value=false;context.showFixedClientModal.value=true;await interval();assert.equal(calls,0)
 context.showFixedClientModal.value=false
 const pending=interval();assert.equal(calls,1)
 await interval();assert.equal(calls,1) // no overlapping calls
 context.showAdvancedClientForm.value=true;resolve(['new']);await pending
 assert.deepEqual(context.clients.value,['original'])
 context.showAdvancedClientForm.value=false
 const second=interval();resolve(['fresh']);await second;assert.deepEqual(context.clients.value,['fresh'])
 const late=interval();unmount();resolve(['after unmount']);await late
 assert.deepEqual(context.clients.value,['fresh']);assert.equal(cleared,1)
})

test('Traffic display uses byte precision and marks zero limit unlimited',()=>{
 const text=source.slice(source.indexOf('function trafficText('),source.indexOf('function trafficText(')+source.slice(source.indexOf('function trafficText(')).indexOf('\n}')+2).replace('c:any','c')
 const context={fmtTime:v=>v||'长期可用'}
 vm.runInNewContext(text+'; globalThis.show=trafficText',context)
 assert.match(context.show({traffic_used_bytes:95226307,traffic_limit_gb:100,billing_cycle_start:'start',expire_at:'end'}),/本周期用量 (?!0 \/).+100 GB（start — end）/)
 assert.match(context.show({traffic_used_bytes:100,traffic_limit_gb:0}),/不限流量/)
})

test('Traffic form retains explicit zero and rejects invalid limits',()=>{
 const start=source.indexOf('function parseTrafficLimit(')
 const text=source.slice(start,source.indexOf('\n}',start)+2).replace('value:any','value')
 const context={}
 vm.runInNewContext(text+'; globalThis.parse=parseTrafficLimit',context)
 assert.equal(context.parse(0),0)
 assert.equal(context.parse('0'),0)
 assert.equal(context.parse(''),100)
 for(const value of [-1,NaN,1.5,'invalid']) assert.throws(()=>context.parse(value))
 assert.match(source,/traffic_limit_gb:parseTrafficLimit\(directForm.value.traffic_limit_gb\)/)
 assert.match(source,/traffic_limit_gb:parseTrafficLimit\(fixedForm.value.traffic_limit_gb\)/)
})
