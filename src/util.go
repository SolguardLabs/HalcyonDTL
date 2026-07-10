package main

import (
	"encoding/json"
	"sort"
)

func sortStrings(values []string) {
	sort.Strings(values)
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func cloneStrings(values []string) []string {
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func stableJSON(value any) []byte {
	out, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return out
}

func boolToInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func statusRank(status string) int {
	switch status {
	case "open", "active":
		return 0
	case "closing", "restricted":
		return 1
	case "closed", "liquidated":
		return 2
	default:
		return 3
	}
}

func nonEmpty(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
