package main

import (
	"strings"
)

func normalizeTargets(targets []string) []string {
	seen := make(map[string]struct{}, len(targets))
	cleaned := make([]string, 0, len(targets))

	for _, target := range targets {
		target = strings.TrimSpace(target)
		target = strings.TrimPrefix(target, "http://")
		target = strings.TrimPrefix(target, "https://")
		target = strings.Trim(target, "/")
		if slashIdx := strings.Index(target, "/"); slashIdx >= 0 {
			target = target[:slashIdx]
		}
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		cleaned = append(cleaned, target)
	}

	return cleaned
}

func parseTargetsText(content string) []string {
	content = strings.TrimPrefix(content, "\ufeff")
	parts := strings.FieldsFunc(content, func(r rune) bool {
		switch r {
		case '\n', '\r', '\t', ' ', ',', ';', '，', '；':
			return true
		default:
			return false
		}
	})

	for idx, part := range parts {
		parts[idx] = strings.Trim(part, `"'`)
	}

	return normalizeTargets(parts)
}
