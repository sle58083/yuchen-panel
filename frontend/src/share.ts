import qrcode from 'qrcode-generator'

export type SubscriptionFormat = 'auto' | 'clash' | 'singbox' | 'v2rayn' | 'raw'
export const subscriptionFormatOptions: Array<{value: SubscriptionFormat;label: string;tip: string}> = [
  {value:'auto',label:'自动识别',tip:'根据客户端 User-Agent 选择格式；识别失败使用通用 base64。'},
  {value:'clash',label:'Clash / Mihomo',tip:'Clash Verge Rev、Mihomo Party、Stash：使用 YAML。'},
  {value:'singbox',label:'sing-box ≥1.12',tip:'sing-box、SFA/SFI，以及支持现代 sing-box JSON 的 NekoBox 构建。≥1.13 请勿手工编辑导出的 DNS 段。'},
  {value:'v2rayn',label:'通用 v2rayN',tip:'v2rayN、v2rayNG、Shadowrocket：使用 base64 链接列表。'},
]

export type ShareFormat = 'v2rayn' | 'shadowrocket' | 'clash' | 'singbox'

export const shareFormatOptions: Array<{ value: ShareFormat; label: string; tip: string }> = [
  { value: 'v2rayn', label: 'V2rayN', tip: '生成 vless:// / vmess:// / trojan:// 单节点链接，优先用于扫码导入。' },
  { value: 'shadowrocket', label: 'Shadowrocket', tip: 'iPhone 小火箭，使用同一条单节点链接。' },
  { value: 'clash', label: 'Clash Meta', tip: '生成 Clash Meta / Mihomo YAML。Clash Verge 建议优先使用专用订阅链接或下载 yaml 文件。' },
  { value: 'singbox', label: 'sing-box', tip: '生成 sing-box ≥1.12 JSON 配置。≥1.13 请勿手工编辑导出的 DNS 段。' },
]


export function buildRelayVirtualNode(relay: any): any {
  return {
    id: relay.id,
    name: relay.name || `中转线路-${relay.relay_port}`,
    protocol: 'vless',
    host: relay.relay_host,
    port: relay.relay_port,
    transport: 'tcp',
    security: 'reality',
    sni: relay.relay_sni || 'www.intel.com',
    fingerprint: relay.relay_fingerprint || 'chrome',
    reality_public_key: relay.relay_reality_public_key || '',
    reality_short_id: relay.relay_reality_short_id || '',
    reality_spider_x: relay.relay_reality_spider_x || '/',
    is_relay_route: true,
    exit_label: relay.landing_mode === 'manual_socks5' ? `${relay.manual_socks_host}:${relay.manual_socks_port}` : relay.landing_node_id,
  }
}

export function clientCanUseRelay(client: any, relay: any): boolean {
  const ids = Array.isArray(client?.relay_route_ids) ? client.relay_route_ids : []
  return ids.includes(relay.id)
}

export function relayNodesForClient(client: any, relays: any[]): any[] {
  return relays.filter(r => r && r.enabled !== false && String(r.route_mode || '') === 'socks5_route' && clientCanUseRelay(client, r)).map(buildRelayVirtualNode)
}

export function qrImageUrl(text: string, size = 560): string {
  const data = String(text || '').trim()
  if (!data) return ''
  const targetSize = Math.max(260, Math.min(Number(size) || 560, 900))
  const qr = qrcode(0, 'M')
  qr.addData(data)
  qr.make()
  const margin = 8
  const moduleCount = qr.getModuleCount() || 41
  const cellSize = Math.max(3, Math.floor(targetSize / (moduleCount + margin * 2)))
  return qr.createDataURL(cellSize, margin)
}

function appBasePath(): string {
  const base = import.meta.env.BASE_URL || '/'
  return base.endsWith('/') ? base : `${base}/`
}

export function buildSubscriptionUrl(client: any, format: SubscriptionFormat = 'auto', vpn = false): string {
  // These values are path/query components, not userinfo; encodeURIComponent is correct here.
  const base = `${location.origin}${appBasePath()}sub/${encodeURIComponent(client.subscribe_token)}`
  return format === 'auto' ? base : `${base}?format=${encodeURIComponent(format)}${format === 'singbox' && vpn ? '&profile=tun' : ''}`
}

export function buildClashSubscriptionUrl(client: any): string {
  return `${buildSubscriptionUrl(client)}?format=clash`
}

export function buildClashDownloadUrl(client: any): string {
  return `${buildSubscriptionUrl(client)}?format=clash&download=1`
}

export function buildShortNodeUrl(client: any, node: any): string {
  return `${location.origin}${appBasePath()}s/${client.subscribe_token}/${node.id}`
}

export function clientCanUseNode(client: any, node: any): boolean {
  const ids = Array.isArray(client?.node_ids) ? client.node_ids : []
  if (client?.enabled === false) return false
  const expiry = String(client?.expire_at || '')
  const permanent = /^0001-01-01T00:00:00(?:\.0+)?(?:Z|\+00:00)$/.test(expiry)
  if (expiry && !permanent) {
    const timestamp = new Date(expiry).getTime()
    if (!Number.isFinite(timestamp) || timestamp <= Date.now()) return false
  }
  return ids.includes(node.id)
}

// 客户可分享的入站协议：SOCKS5 / dokodemo-door 属于内部链路，不生成客户订阅。
const SHAREABLE_PROTOCOLS = ['vless', 'vmess', 'trojan']

export function isClientShareNode(node: any): boolean {
  return SHAREABLE_PROTOCOLS.includes(String(node?.protocol || '').toLowerCase())
}

export function nodesForClient(client: any, nodes: any[]): any[] {
  return nodes.filter(n => n && n.enabled !== false && isClientShareNode(n) && clientCanUseNode(client, n))
}

export function clientsForNode(node: any, clients: any[]): any[] {
  return clients.filter(c => c && c.enabled !== false && clientCanUseNode(c, node))
}

function valueOr(v: any, fallback: string): string {
  const s = String(v || '').trim()
  return s || fallback
}

function nodeHost(node: any): string {
  return valueOr(node.host || node.address || node.ip, '127.0.0.1').replace(/^\[|\]$/g, '')
}

function uriHost(host: string): string {
  return host.includes(':') ? `[${host.replace(/%/g, '%25')}]` : host
}

function uniqueProxyNames(nodes: any[]): any[] {
  const used = new Set(['PROXY', 'direct', 'DIRECT'])
  return nodes.map(node => {
    const base = safeName(node)
    let name = base
    for (let i = 2; used.has(name); i++) name = `${base} #${i}`
    used.add(name)
    return {...node, name}
  })
}

function nodeName(node: any): string {
  return valueOr(node.name || node.remark, `${nodeHost(node)}:${Number(node.port || 443)}`)
}

function shortNodeId(node: any): string {
  const raw = String(node?.id || node?.host || 'node').replace(/[^A-Za-z0-9]/g, '').toLowerCase()
  return (raw.slice(-4) || 'node').padStart(4, '0')
}

function safeName(node: any, client?: any): string {
  const cn = client?.username ? `-${client.username}` : ''
  return `${nodeName(node)} #${shortNodeId(node)}${cn}`.replace(/[\r\n]+/g, ' ').trim()
}

// Userinfo must use RFC3986 percent encoding; URLSearchParams is for query values.
function encodeUserInfo(v: string): string {
  return encodeURIComponent(v).replace(/[!'()*]/g, c => '%' + c.charCodeAt(0).toString(16).toUpperCase())
}

function isReality(node: any): boolean {
  return String(node.security || '').toLowerCase() === 'reality'
}

function transportOf(node: any): string {
  return valueOr(node.transport, 'tcp').toLowerCase()
}

function utf8ToBase64(s: string): string {
  // Unicode 安全的 base64：vmess JSON 里可能包含中文节点名。
  const bytes = new TextEncoder().encode(s)
  let bin = ''
  bytes.forEach(b => { bin += String.fromCharCode(b) })
  return btoa(bin)
}

export function buildVlessLink(node: any, client: any, format: 'v2rayn' | 'shadowrocket' = 'v2rayn'): string {
  const host = uriHost(nodeHost(node))
  const port = Number(node.port || 443)
  const transport = transportOf(node)
  const security = valueOr(node.security, 'none').toLowerCase()
  const q = new URLSearchParams()
  q.set('encryption', 'none')
  q.set('security', security)
  if (node.sni) q.set('sni', node.sni)
  if (security === 'reality') {
    // V0.7.5.9.1: keep QR/import URI aligned with server-side Xray client flow.
    // Do not force flow=xtls-rprx-vision unless the server actually configures client.flow.
    q.set('fp', valueOr(node.fingerprint, 'chrome'))
    if (node.reality_public_key) q.set('pbk', node.reality_public_key)
    if (node.reality_short_id) q.set('sid', node.reality_short_id)
    q.set('spx', valueOr(node.reality_spider_x, '/'))
  }
  q.set('type', transport)
  if (transport === 'ws') q.set('path', valueOr(node.path, '/yuchen'))
  if (transport === 'grpc') q.set('serviceName', valueOr(String(node.path || '').replace(/^\/+|\/+$/g, ''), 'yuchen'))
  const label = encodeURIComponent(safeName(node))
  return `vless://${client.uuid}@${host}:${port}?${q.toString()}#${label}`
}

// buildVmessLink 生成 vmess:// 链接（v2rayN base64 JSON 格式），与后端订阅导出一致。
export function buildVmessLink(node: any, client: any): string {
  const transport = transportOf(node)
  const security = valueOr(node.security, 'none').toLowerCase()
  const obj: any = {
    v: '2',
    ps: nodeName(node),
    add: nodeHost(node),
    port: String(Number(node.port || 443)),
    id: String(client.uuid || ''),
    aid: '0',
    scy: 'auto',
    net: transport,
    type: 'none',
    host: '',
    path: '',
    tls: '',
  }
  if (transport === 'ws') obj.path = valueOr(node.path, '/yuchen')
  if (transport === 'grpc') obj.path = valueOr(String(node.path || '').replace(/^\/+|\/+$/g, ''), 'yuchen')
  if (security === 'tls' || security === 'reality') {
    obj.tls = 'tls'
    obj.sni = String(node.sni || '')
    obj.fp = valueOr(node.fingerprint, 'chrome')
  }
  if (security === 'reality') {
    obj.pbk = String(node.reality_public_key || '')
    obj.sid = String(node.reality_short_id || '')
    obj.spx = valueOr(node.reality_spider_x, '/')
  }
  return `vmess://${utf8ToBase64(JSON.stringify(obj))}`
}

// buildTrojanLink 生成 trojan:// 链接，密码与后端 Xray clients 保持一致。
export function buildTrojanLink(node: any, client: any): string {
  const host = uriHost(nodeHost(node))
  const port = Number(node.port || 443)
  const transport = transportOf(node)
  const security = valueOr(node.security, 'none').toLowerCase()
  const q = new URLSearchParams()
  if (security === 'reality') {
    q.set('security', 'reality')
    q.set('fp', valueOr(node.fingerprint, 'chrome'))
    if (node.reality_public_key) q.set('pbk', node.reality_public_key)
    if (node.reality_short_id) q.set('sid', node.reality_short_id)
    q.set('spx', valueOr(node.reality_spider_x, '/'))
  } else {
    q.set('security', 'tls')
  }
  if (node.sni) q.set('sni', node.sni)
  q.set('type', transport)
  if (transport === 'ws') q.set('path', valueOr(node.path, '/yuchen'))
  if (transport === 'grpc') q.set('serviceName', valueOr(String(node.path || '').replace(/^\/+|\/+$/g, ''), 'yuchen'))
  const label = encodeURIComponent(safeName(node))
  const password = encodeUserInfo(String(client.password || client.uuid || ''))
  return `trojan://${password}@${host}:${port}?${q.toString()}#${label}`
}

// buildNodeLink 按入站协议生成对应客户端分享链接。
export function buildNodeLink(node: any, client: any, format: 'v2rayn' | 'shadowrocket' = 'v2rayn'): string {
  if (!isClientShareNode(node)) throw new Error('该协议已停用或不支持分享')
  const proto = String(node?.protocol || '').toLowerCase()
  if (proto === 'vmess') return buildVmessLink(node, client)
  if (proto === 'trojan') return buildTrojanLink(node, client)
  return buildVlessLink(node, client, format)
}

function yamlQuote(v: any): string {
  return JSON.stringify(String(v ?? ''))
}

function clashProxyTail(node: any, security: string, transport: string, lines: string[]) {
  if (security === 'reality' || security === 'tls') {
    lines.push(`    tls: true`)
    if (node.sni) lines.push(`    servername: ${yamlQuote(node.sni)}`)
    lines.push(`    client-fingerprint: ${yamlQuote(valueOr(node.fingerprint, 'chrome'))}`)
    lines.push('    skip-cert-verify: false')
  } else {
    lines.push(`    tls: false`)
  }
  if (security === 'reality') {
    lines.push(`    reality-opts:`)
    lines.push(`      public-key: ${yamlQuote(node.reality_public_key || '')}`)
    if (node.reality_short_id) lines.push(`      short-id: ${yamlQuote(node.reality_short_id || '')}`)
  }
  if (transport === 'ws') {
    lines.push(`    ws-opts:`)
    lines.push(`      path: ${yamlQuote(valueOr(node.path, '/yuchen'))}`)
  }
  if (transport === 'grpc') {
    lines.push(`    grpc-opts:`)
    lines.push(`      grpc-service-name: ${yamlQuote(valueOr(String(node.path || '').replace(/^\/+|\/+$/g, ''), 'yuchen'))}`)
  }
}

function clashProxy(node: any, client: any): string {
  const proto = String(node?.protocol || '').toLowerCase()
  const transport = transportOf(node)
  const security = valueOr(node.security, 'none').toLowerCase()
  const lines: string[] = []
  lines.push(`  - name: ${yamlQuote(safeName(node))}`)
  lines.push(`    server: ${yamlQuote(nodeHost(node))}`)
  lines.push(`    port: ${Number(node.port || 443)}`)
  if (proto === 'vmess') {
    lines.push(`    type: vmess`)
    lines.push(`    uuid: ${yamlQuote(client.uuid)}`)
    lines.push(`    alterId: 0`)
    lines.push(`    cipher: auto`)
    lines.push(`    network: ${yamlQuote(transport)}`)
    lines.push(`    udp: true`)
    clashProxyTail(node, security, transport, lines)
    return lines.join('\n')
  }
  if (proto === 'trojan') {
    lines.push(`    type: trojan`)
    lines.push(`    password: ${yamlQuote(String(client.password || client.uuid || ''))}`)
    lines.push(`    network: ${yamlQuote(transport)}`)
    lines.push(`    udp: true`)
    // Trojan 协议本身要求 TLS；reality 时输出 reality-opts。
    clashProxyTail(node, security === 'reality' ? 'reality' : 'tls', transport, lines)
    return lines.join('\n')
  }
  lines.push(`    type: vless`)
  lines.push(`    uuid: ${yamlQuote(client.uuid)}`)
  lines.push(`    network: ${yamlQuote(transport)}`)
  lines.push(`    udp: true`)
  clashProxyTail(node, security, transport, lines)
  return lines.join('\n')
}

export function buildClashMetaConfig(nodes: any[], client: any, policy: any = {}): string {
  const enabled = uniqueProxyNames(nodes.filter(n => n && n.enabled !== false && isClientShareNode(n)))
  const names = enabled.map(n => safeName(n))
  const proxyList = enabled.map(n => clashProxy(n, client)).join('\n')
  const groupProxies = names.length ? [...names.map(n => `      - ${yamlQuote(n)}`), '      - DIRECT'].join('\n') : '      - DIRECT'
  const nameservers = ['    - 1.1.1.1', '    - 8.8.8.8']
  if (policy?.clash_include_quad9) nameservers.push('    - 9.9.9.9')
  return [
    'mixed-port: 7890',
    'allow-lan: false',
    'mode: rule',
    'log-level: info',
    'ipv6: false',
    'global-client-fingerprint: chrome',
    'unified-delay: true',
    'tcp-concurrent: true',
    'dns:',
    '  enable: true',
    '  ipv6: false',
    '  enhanced-mode: fake-ip',
    '  nameserver:',
    ...nameservers,
    'proxies:',
    proxyList || '  []',
    'proxy-groups:',
    '  - name: PROXY',
    '    type: select',
    '    proxies:',
    groupProxies,
    'rules:',
    '  - MATCH,PROXY',
    '',
  ].join('\n')
}

function singBoxTLS(node: any, forceTLS = false): any | undefined {
  const security = valueOr(node.security, 'none').toLowerCase()
  const needTLS = forceTLS || security === 'tls' || security === 'reality'
  if (!needTLS) return undefined
  const tls: any = {
    enabled: true,
    server_name: valueOr(node.sni, nodeHost(node)),
    utls: { enabled: true, fingerprint: valueOr(node.fingerprint, 'chrome') },
  }
  if (security === 'reality') {
    tls.reality = {
      enabled: true,
      public_key: node.reality_public_key || '',
      short_id: node.reality_short_id || '',
    }
  }
  return tls
}

function singBoxTransport(node: any): any | undefined {
  const transport = transportOf(node)
  if (transport === 'ws') {
    return { type: 'ws', path: valueOr(node.path, '/yuchen') }
  }
  if (transport === 'grpc') {
    return { type: 'grpc', service_name: valueOr(String(node.path || '').replace(/^\/+|\/+$/g, ''), 'yuchen') }
  }
  return undefined
}

function singBoxOutbound(node: any, client: any): any {
  const proto = String(node?.protocol || '').toLowerCase()
  if (proto === 'vmess') {
    const outbound: any = {
      type: 'vmess',
      tag: safeName(node),
      server: nodeHost(node),
      server_port: Number(node.port || 443),
      uuid: String(client.uuid || ''),
      security: 'auto',
      packet_encoding: 'xudp',
    }
    const tls = singBoxTLS(node)
    if (tls) outbound.tls = tls
    const transport = singBoxTransport(node)
    if (transport) outbound.transport = transport
    return outbound
  }
  if (proto === 'trojan') {
    const outbound: any = {
      type: 'trojan',
      tag: safeName(node),
      server: nodeHost(node),
      server_port: Number(node.port || 443),
      password: String(client.password || client.uuid || ''),
    }
    const tls = singBoxTLS(node, true)
    if (tls) outbound.tls = tls
    const transport = singBoxTransport(node)
    if (transport) outbound.transport = transport
    return outbound
  }
  const outbound: any = {
    type: 'vless',
    tag: safeName(node),
    server: nodeHost(node),
    server_port: Number(node.port || 443),
    uuid: client.uuid,
    packet_encoding: 'xudp',
  }
  const tls = singBoxTLS(node)
  if (tls) outbound.tls = tls
  const transport = singBoxTransport(node)
  if (transport) outbound.transport = transport
  return outbound
}

export function buildSingBoxConfig(nodes: any[], client: any, policy: any = {}): string {
  const enabled = uniqueProxyNames(nodes.filter(n => n && n.enabled !== false && isClientShareNode(n)))
  const outbounds = enabled.map(n => singBoxOutbound(n, client))
  const selector = {
    type: 'selector',
    tag: 'PROXY',
    outbounds: [...outbounds.map(o => o.tag), 'direct'],
    default: outbounds[0]?.tag || 'direct',
  }
  const config: any = {
    log: { level: 'info' },
    dns: {
      servers: [
        { type: 'udp', tag: 'cloudflare', server: '1.1.1.1' },
        { type: 'udp', tag: 'google', server: '8.8.8.8' },
        ...(policy?.sing_box_include_quad9 ? [{ type: 'udp', tag: 'quad9', server: '9.9.9.9' }] : []),
      ],
      final: 'cloudflare',
      strategy: 'ipv4_only',
    },
    inbounds: [
      { type: 'mixed', tag: 'mixed-in', listen: '127.0.0.1', listen_port: 2080 },
    ],
    outbounds: [selector, ...outbounds, { type: 'direct', tag: 'direct' }],
    route: { final: 'PROXY', default_domain_resolver: 'cloudflare' },
  }
  if (policy?.client_tun) {
    config.inbounds.unshift({type:'tun',tag:'tun-in',address:['172.19.0.1/30','fdfe:dcba:9876::1/126'],mtu:1400,auto_route:true,strict_route:true,stack:'system'})
    config.route.auto_detect_interface = true
    config.route.rules = [{port:53,action:'hijack-dns'}]
  }
  return JSON.stringify(config, null, 2)
}

export function buildClientShare(node: any, client: any, format: ShareFormat, policy: any = {}): string {
  if (!isClientShareNode(node)) throw new Error('该协议已停用或不支持分享')
  if (format === 'shadowrocket') return buildNodeLink(node, client, 'shadowrocket')
  if (format === 'clash') return buildClashMetaConfig([node], client, policy)
  if (format === 'singbox') return buildSingBoxConfig([node], client, policy)
  return buildNodeLink(node, client, 'v2rayn')
}

export function buildClientMultiShare(nodes: any[], client: any, format: ShareFormat, policy: any = {}): string {
  const clientNodes = nodes.filter(isClientShareNode)
  if (format === 'clash') return buildClashMetaConfig(clientNodes, client, policy)
  if (format === 'singbox') return buildSingBoxConfig(clientNodes, client, policy)
  return clientNodes.map(n => buildNodeLink(n, client, format === 'shadowrocket' ? 'shadowrocket' : 'v2rayn')).join('\n')
}

export function isQrShareFormat(format: ShareFormat): boolean {
  return format === 'v2rayn' || format === 'shadowrocket'
}

export function shareFormatLabel(format: ShareFormat): string {
  return shareFormatOptions.find(x => x.value === format)?.label || format
}
