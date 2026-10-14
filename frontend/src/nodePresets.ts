export type NodePreset = 'fast' | 'balanced' | 'anti'
// The default is a user-confirmed deployment policy, not a throughput benchmark.
export const DEFAULT_NODE_PRESET: NodePreset = 'anti'
export const NODE_PRESETS = [
  { id: 'fast', label: '极速', protocol: 'vless', transport: 'tcp', security: 'none', description: '无额外传输层加密；抗识别：低。吞吐未实测，不保证跑满线路。' },
  { id: 'balanced', label: '均衡', protocol: 'vless', transport: 'grpc', security: 'none', description: 'gRPC/HTTP2 封装；抗识别：中（相对裸 TCP，不等同 TLS）。性能影响未实测。' },
  { id: 'anti', label: '抗封锁', protocol: 'vless', transport: 'tcp', security: 'reality', description: 'TLS 外观与加密；抗识别：高（相对前两档）。性能影响未实测，不保证免封锁。' },
] as const
export function presetParameters(id: NodePreset) {
  const preset = NODE_PRESETS.find(p => p.id === id)!
  return { protocol: preset.protocol, transport: preset.transport, security: preset.security }
}
export function detectNodePreset(node: any): NodePreset | 'advanced' | 'socks5' {
  if (node.protocol === 'socks') return 'socks5'
  const preset = NODE_PRESETS.find(p => node.protocol === p.protocol && (node.transport || 'tcp') === p.transport && (node.security || 'none') === p.security)
  return preset?.id || 'advanced'
}
export function serverSpecAdvice(specs: any): string[] {
  if (!specs || typeof specs !== 'object') return []
  const hints: string[] = []
  const cores = specs.cpu_cores, memory = specs.memory_total_bytes
  if (Number.isInteger(cores) && cores > 0 && cores <= 2) hints.push('≤ 2 核：吞吐优先时建议极速档减少协议开销；CPU 可能成为瓶颈，需结合实际占用排查。墙内抗封锁需求仍可选 Reality。')
  if (Number.isInteger(cores) && cores >= 4) hints.push('≥ 4 核：可使用 Reality；核数不代表单连接吞吐，也不保证线路跑满。')
  if (typeof memory === 'number' && Number.isFinite(memory) && memory > 0 && memory < 1024 ** 3) hints.push('内存低于 1 GiB：注意内存余量和并发负载，避免 OOM。')
  return hints
}
