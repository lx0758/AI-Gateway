package handler

import (
	"fmt"
	"strconv"
	"strings"
)

// parseContextString parses a context window string into raw token count.
// Supports:
//   - Pure number: "128000" → 128000
//   - K/k suffix: "128k", "128K" → 131072 (128 × 1024)
//   - M/m suffix: "1m", "1M" → 1048576 (1 × 1024 × 1024)
//   - B/b suffix: "1b", "1B" → 1073741824 (1 × 1024 × 1024 × 1024)
func parseContextString(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(strings.TrimSpace(s))

	multiplier := 1
	numStr := s

	switch s[len(s)-1] {
	case 'k':
		multiplier = 1024
		numStr = s[:len(s)-1]
	case 'm':
		multiplier = 1024 * 1024
		numStr = s[:len(s)-1]
	case 'b':
		multiplier = 1024 * 1024 * 1024
		numStr = s[:len(s)-1]
	}

	num, err := strconv.Atoi(strings.TrimSpace(numStr))
	if err != nil {
		return 0
	}

	return num * multiplier
}

// formatContextNumber formats a raw token count as a human-readable string.
func formatContextNumber(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 1024 {
		return strconv.Itoa(n)
	}
	if n < 1024*1024 {
		k := float64(n) / 1024.0
		if k == float64(int(k)) {
			return fmt.Sprintf("%dK", int(k))
		}
		return fmt.Sprintf("%.1fK", k)
	}
	m := float64(n) / float64(1024*1024)
	if m == float64(int(m)) {
		return fmt.Sprintf("%dM", int(m))
	}
	return fmt.Sprintf("%.1fM", m)
}
