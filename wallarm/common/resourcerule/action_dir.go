package resourcerule

import (
	"strings"

	wallarm "github.com/wallarm/wallarm-go"
)

const (
	actionDirMaxLen  = 64
	actionDirHashLen = 8
	// maxPrefixLen = 64 - 1 (underscore) - 8 (hash) = 55
	actionDirMaxPrefixLen = actionDirMaxLen - 1 - actionDirHashLen
)

// ActionDirName computes a filesystem-safe directory name for an action.
//
// Format: {instance}_{domain}_{path}_{hash8}
//   - Empty conditions → "_default" (no hash)
//   - Path-only → "_" prefix: "_api_v1_users_e3a1ef0f"
//   - Instance-based → numeric prefix: "13_example.com_api_c522d1d1"
//   - Domain-based → alpha prefix: "example.com_api_a3f2e1b7"
//
// Path transform: "/" → "_", "*" → ".", "**" → ".."
// Max 64 chars. Prefix truncated at last "_" boundary if needed.
func ActionDirName(conditions []wallarm.ActionDetails) string {
	if len(conditions) == 0 {
		return defaultActionDir
	}

	rev := ReverseMapActions(conditions)
	hash := ConditionsHash(conditions)[:actionDirHashLen]

	prefix := buildDirPrefix(rev.Instance, rev.Domain, rev.Path)
	if prefix == "" {
		// Only non-path/domain/instance conditions (e.g., method-only).
		// Use hash only with underscore prefix.
		return "_" + hash
	}

	// Truncate prefix to fit within max length.
	if len(prefix) > actionDirMaxPrefixLen {
		prefix = truncateAtBoundary(prefix, actionDirMaxPrefixLen)
	}

	return prefix + "_" + hash
}

// buildDirPrefix constructs the human-readable prefix from instance, domain, path.
func buildDirPrefix(instance, domain, path string) string {
	var parts []string

	if instance != "" {
		parts = append(parts, instance)
	}

	if domain != "" {
		parts = append(parts, domain)
	}

	pathPart := sanitizePath(path)
	if pathPart != "" {
		parts = append(parts, pathPart)
	}

	result := strings.Join(parts, "_")

	// Path-only (no instance, no domain) → underscore prefix.
	if instance == "" && domain == "" && result != "" {
		result = "_" + result
	}

	return result
}

// sanitizePath converts a URL path to a filesystem-safe directory name component.
//
//	"/" → "root"
//	"/**/*.*" → "" (global, handled by caller)
//	"/api/v1/users" → "api_v1_users"
//	"/api/**/*.*" → "api_.._._."
//	"*" → ".", "**" → ".."
func sanitizePath(path string) string {
	if path == "" || path == pathGlobalWildcard {
		return ""
	}
	if path == "/" {
		return "root"
	}

	// Split into segments and sanitize each individually to avoid
	// false wildcard matches (e.g., "*.*" must become "._." not "...").
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	sanitized := make([]string, len(segments))
	for i, seg := range segments {
		sanitized[i] = sanitizeSegment(seg)
	}

	return strings.Join(sanitized, "_")
}

// sanitizeSegment converts a single path segment to filesystem-safe form.
//
//	"**" → ".."
//	"*" → "."
//	"*.*" → "._." (dot becomes "_" separator, wildcards become ".")
//	"action.ext" → "action.ext" (literal dots preserved when no wildcards)
func sanitizeSegment(seg string) string {
	if seg == "**" {
		return ".."
	}

	// If segment contains wildcards, split at dots and rejoin with "_"
	// to keep the name/ext structure visible: *.*  → ._. not ...
	if strings.Contains(seg, "*") {
		parts := strings.Split(seg, ".")
		for i, p := range parts {
			parts[i] = strings.ReplaceAll(p, "*", ".")
		}
		return strings.Join(parts, "_")
	}

	return seg
}

// truncateAtBoundary truncates s to at most maxLen, cutting at the last "_" boundary.
func truncateAtBoundary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	truncated := s[:maxLen]
	lastUnderscore := strings.LastIndex(truncated, "_")
	if lastUnderscore > 0 {
		return truncated[:lastUnderscore]
	}
	return truncated
}
