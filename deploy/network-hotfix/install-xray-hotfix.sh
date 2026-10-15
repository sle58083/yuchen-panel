#!/usr/bin/env bash
set -Eeuo pipefail
umask 077
[[ $EUID -eq 0 ]] || { echo 'Run with sudo bash' >&2; exit 1; }
tags='[]'; restart=0; config=/etc/yuchen-panel/xray/config.json
while (($#)); do
  case $1 in
    --vision-tags) tags=${2:?Missing JSON array}; shift 2 ;;
    --config) config=${2:?Missing absolute config path}; shift 2 ;;
    --restart) restart=1; shift ;;
    *) echo "Unknown option: $1" >&2; exit 2 ;;
  esac
done
for cmd in jq flock systemctl install mktemp; do
  command -v "$cmd" >/dev/null || { echo "Missing dependency: $cmd" >&2; exit 1; }
done
[[ -d /run/systemd/system && $config == /* && -r $config ]]
systemctl cat xray.service >/dev/null
user=$(systemctl show xray.service -p User --value)
[[ -z $user || $user == root ]] || { echo 'Non-root service: adapt runtime directory permissions first' >&2; exit 1; }
systemctl show xray.service -p ExecStart
echo 'This hook replaces ExecStart with one Xray run -config invocation.'
[[ $(systemctl show xray.service -p DynamicUser --value) != yes ]]
bin=$(command -v xray)
[[ $bin == /* ]]
base=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
hook_dir=/etc/yuchen-panel/hotfix
dropin=/etc/systemd/system/xray.service.d/zz-yuchen-hotfix.conf
[[ ! -e $dropin && ! -e $hook_dir ]] || { echo 'Hook already present; uninstall before replacing' >&2; exit 1; }
backup=$(mktemp -d /var/tmp/yuchen-hotfix-backup.XXXXXXXX)
systemctl cat xray.service > "$backup/service-before.txt"
cp -a -- "$config" "$backup/source-before.json"
install -d -m 0700 "$hook_dir"
install -m 0600 "$base/xray-hotfix.jq" "$hook_dir/filter.jq"
install -m 0700 "$base/xray-hotfix-start.sh" "$hook_dir/start.sh"
printf 'SOURCE_CONFIG=%q\nXRAY_BIN=%q\nVISION_TAGS=%q\n' "$config" "$bin" "$tags" > "$hook_dir/hook.env"
cat > "$backup/uninstall.sh" <<'UNDO'
#!/usr/bin/env bash
set -Eeuo pipefail
[[ $EUID -eq 0 ]]
rm -f -- /etc/systemd/system/xray.service.d/zz-yuchen-hotfix.conf
rm -f -- /etc/yuchen-panel/hotfix/filter.jq /etc/yuchen-panel/hotfix/start.sh /etc/yuchen-panel/hotfix/hook.env
rmdir /etc/yuchen-panel/hotfix
systemctl daemon-reload
systemctl restart xray.service
UNDO
chmod 0700 "$backup/uninstall.sh"
trap 'echo "Installation failed; recovery: bash $backup/uninstall.sh" >&2' ERR
"$hook_dir/start.sh" --check
install -d -m 0755 /etc/systemd/system/xray.service.d
cat > "$dropin" <<'UNIT'
[Service]
ExecStart=
ExecStart=/etc/yuchen-panel/hotfix/start.sh
UNIT
systemctl daemon-reload
echo "Uninstall and restart original service: bash $backup/uninstall.sh"
if ((restart)); then
  systemctl restart xray.service
  systemctl is-active --quiet xray.service
else
  echo 'Activate during maintenance: systemctl restart xray.service'
fi
