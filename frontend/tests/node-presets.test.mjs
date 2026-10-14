import test from 'node:test'
import assert from 'node:assert/strict'
import vm from 'node:vm'
import { build } from 'esbuild'
const result=await build({entryPoints:['src/nodePresets.ts'],bundle:true,platform:'node',format:'cjs',write:false})
const module={exports:{}}
vm.runInNewContext(result.outputFiles[0].text,{module,exports:module.exports})
const {NODE_PRESETS,DEFAULT_NODE_PRESET,presetParameters,detectNodePreset,serverSpecAdvice}=module.exports
const plain=value=>JSON.parse(JSON.stringify(value))
test('C24_6_three_preset_parameter_mappings',()=>{
 for(const [id,transport,security] of [['fast','tcp','none'],['balanced','grpc','none'],['anti','tcp','reality']]){
  assert.deepEqual(plain(presetParameters(id)),{protocol:'vless',transport,security})
  assert.equal(detectNodePreset({protocol:'vless',transport,security}),id)
 }
 assert.equal(DEFAULT_NODE_PRESET,'anti','human confirmed mainland anti-blocking default')
 assert.equal(detectNodePreset({protocol:'trojan',transport:'tcp',security:'reality'}),'advanced')
 assert.equal(detectNodePreset({protocol:'vless',transport:'ws',security:'none'}),'advanced')
 for(const p of NODE_PRESETS){ assert.match(p.description,/抗识别/);assert.match(p.description,/未实测/);assert.doesNotMatch(p.description,/\d+%/) }
})
test('C24_7_two_or_fewer_cores_suggest_fast_without_overriding_default',()=>{
 for(const cores of [1,2]) assert.match(serverSpecAdvice({cpu_cores:cores}).join(''),/极速档/)
 assert.equal(DEFAULT_NODE_PRESET,'anti')
 assert.match(serverSpecAdvice({cpu_cores:4}).join(''),/Reality/)
 assert.match(serverSpecAdvice({memory_total_bytes:961*1024**2}).join(''),/内存低于/)
 assert.deepEqual(plain(serverSpecAdvice({memory_total_bytes:1024**3})),[])
 for(const specs of [undefined,null,{}, {cpu_cores:0,memory_total_bytes:0}, {cpu_cores:-1,memory_total_bytes:-1}, {cpu_cores:'2'}]) assert.deepEqual(plain(serverSpecAdvice(specs)),[])
})
