package substrate

import "path"

// ZoneMatch checks whether a zone name matches a glob pattern.
// Supports:
//   - Exact match: "backend/auth" matches "backend/auth"
//   - Single wildcard: "backend/*" matches "backend/auth" but not "backend/auth/tokens"
//   - Recursive wildcard: "backend/**" matches "backend/auth" and "backend/auth/tokens"
func ZoneMatch(pattern, zone string) bool {
	if pattern == "" || zone == "" {
		return false
	}
	if pattern == zone {
		return true
	}
	if pattern == "**" {
		return true
	}

	// Handle ** recursive wildcard
	if len(pattern) >= 3 && pattern[len(pattern)-3:] == "/**" {
		prefix := pattern[:len(pattern)-3]
		if zone == prefix {
			return true
		}
		if len(zone) > len(prefix) && zone[:len(prefix)+1] == prefix+"/" {
			return true
		}
		return false
	}

	// Fall back to path.Match for single-segment wildcards
	matched, _ := path.Match(pattern, zone)
	return matched
}
