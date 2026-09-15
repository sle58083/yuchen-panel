// SPDX-License-Identifier: AGPL-3.0-only
package xray

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"yuchen-panel/backend/internal/model"
)

// TrojanClientPassword 返回客户在 Trojan 入站中使用的密码。
// 优先使用客户保存的 Password；旧客户没有 Password 时从客户 ID 确定性派生，
// 保证 Xray 配置生成、订阅导出与 API 返回的密码完全一致且在多次同步间保持稳定。
func TrojanClientPassword(c model.Client) string {
	if pw := strings.TrimSpace(c.Password); pw != "" {
		return pw
	}
	sum := sha256.Sum256([]byte("trojan-" + c.ID))
	return hex.EncodeToString(sum[:])[:16]
}

// SSCredentials 返回 Shadowsocks 入站的 method 和 password。
// Xray 的 SS 入站是单用户凭据（settings.method + settings.password），
// 因此凭据保存在入站上而不是每个客户上。
func SSCredentials(n model.Node) (string, string) {
	method := strings.TrimSpace(n.SSMethod)
	if method == "" {
		method = "aes-256-gcm"
	}
	password := strings.TrimSpace(n.SSPassword)
	if password == "" {
		sum := sha256.Sum256([]byte("ss-" + n.ID))
		password = hex.EncodeToString(sum[:])[:16]
	}
	return method, password
}
