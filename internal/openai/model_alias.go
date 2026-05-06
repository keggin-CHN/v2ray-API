package openai

import "strings"

func CanonicalModelName(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "gpt5.4":
		return "gpt-5.4"
	case "gpt5.5":
		return "gpt-5.5"
	case "gpt5.3codex":
		return "gpt-5.3-codex"
	default:
		return name
	}
}
