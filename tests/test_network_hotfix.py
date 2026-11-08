import json, pathlib, subprocess, shutil, os
root = pathlib.Path(__file__).resolve().parent.parent
jq = os.environ.get('JQ_BIN') or shutil.which('jq')
if not jq:
    raise SystemExit('jq is required')
flt = root / 'deploy/network-hotfix/xray-hotfix.jq'
base = {'inbounds': [
    {'tag': 'relay', 'protocol': 'dokodemo-door', 'sniffing': {'enabled': True}},
    {'tag': 'vision', 'protocol': 'vless', 'streamSettings': {'network': 'tcp', 'security': 'reality'}, 'settings': {'clients': [{'id': 'test-id', 'flow': ''}]}},
    {'tag': 'ws', 'protocol': 'vless', 'streamSettings': {'network': 'ws', 'security': 'reality'}, 'settings': {'clients': [{'id': 'test-id', 'flow': ''}]}},
    {'tag': 'unchanged', 'protocol': 'trojan', 'settings': {'clients': [{'password': 'fake-test-only'}]}}
], 'outbounds': [{'protocol': 'freedom'}], 'routing': {'rules': []}}
def run(cfg, tags):
    return subprocess.run([str(jq), '-e', '--argjson', 'tags', json.dumps(tags), '-f', 'xray-hotfix.jq'], cwd=flt.parent, input=json.dumps(cfg), text=True, capture_output=True)
r = run(base, ['vision'])
assert r.returncode == 0, r.stderr
out = json.loads(r.stdout)
assert 'sniffing' not in out['inbounds'][0]
assert out['inbounds'][1]['settings']['clients'][0]['flow'] == 'xtls-rprx-vision'
assert out['inbounds'][2:] == base['inbounds'][2:]
assert out['outbounds'] == base['outbounds'] and out['routing'] == base['routing']
assert json.loads(run(out, ['vision']).stdout) == out
assert json.loads(run(base, []).stdout)['inbounds'][1] == base['inbounds'][1]
for tags in [['ws'], ['missing'], 'not-array', [42]]:
    assert run(base, tags).returncode != 0, tags
bad = json.loads(json.dumps(base)); bad['inbounds'][1]['settings']['clients'][0]['flow'] = 'incompatible'
assert run(bad, ['vision']).returncode != 0
bad = json.loads(json.dumps(base)); bad['inbounds'].append(bad['inbounds'][1])
assert run(bad, ['vision']).returncode != 0
assert run({'inbounds': {}}, []).returncode != 0
print('PASS: transform, empty opt-in, eligibility, rejection, preservation, idempotence')
