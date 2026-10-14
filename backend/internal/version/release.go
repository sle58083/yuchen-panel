package version

import "strings"

// ParseReleaseID splits a release identifier into comparable Version and Channel.
func ParseReleaseID(raw string) (string, string, bool) {
	fields := strings.SplitN(strings.TrimSpace(raw), "-", 2)
	if fields[0] == "" {
		return "", "", false
	}
	channel := ""
	if len(fields) == 2 {
		channel = strings.TrimSpace(fields[1])
	}
	return fields[0], channel, true
}
