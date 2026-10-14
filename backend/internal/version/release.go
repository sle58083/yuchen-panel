package version

import "strings"

// ParseReleaseID preserves SemVer prereleases; -stable is legacy channel metadata.
func ParseReleaseID(raw string) (string, string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", "", false
	}
	if strings.HasSuffix(value, "-stable") {
		return strings.TrimSuffix(value, "-stable"), "stable", true
	}
	parts := strings.SplitN(value, "-", 2)
	if len(parts) == 1 {
		return value, "stable", true
	}
	if parts[1] == "" {
		return "", "", false
	}
	return value, strings.SplitN(parts[1], ".", 2)[0], true
}
