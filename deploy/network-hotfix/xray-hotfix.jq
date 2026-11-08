if type != "object" or (.inbounds | type) != "array" then
  error("Expected an Xray object with inbounds array")
else . end
| . as $original
| def eligible:
    .protocol == "vless"
    and .streamSettings.security == "reality"
    and ((.streamSettings.network // "tcp") | . == "tcp" or . == "raw");
  if ($tags | type) != "array" or any($tags[]; type != "string") then
    error("vision-tags must be a JSON array of strings")
  elif any($tags[]; . as $t |
    ([$original.inbounds[] | select(.tag == $t and eligible)] | length) != 1) then
    error("Each selected tag must identify exactly one VLESS Reality TCP/RAW inbound")
  else . end
| .inbounds |= map(
    if .protocol == "dokodemo-door" then del(.sniffing) else . end
    | if .tag as $t | $tags | index($t) != null then
        if (.settings.clients | type) != "array" or
           any(.settings.clients[]; type != "object" or
             (.id | type) != "string" or
             ((.flow // "") != "" and .flow != "xtls-rprx-vision")) then
          error("Selected inbound has missing clients/IDs or an incompatible existing flow")
        else .settings.clients |= map(.flow = "xtls-rprx-vision") end
      else . end)
