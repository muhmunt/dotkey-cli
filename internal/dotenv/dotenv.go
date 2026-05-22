package dotenv

import (
	"fmt"
	"os"
	"strings"
)

// Parse reads a .env file and returns a map of key → value.
// Handles comments (#), blank lines, and quoted values.
func Parse(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseString(string(data))
}

// ParseString parses KEY=VALUE lines from a raw string.
func ParseString(content string) (map[string]string, error) {
	result := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := parts[1]

		// strip surrounding quotes
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') ||
				(val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		result[key] = val
	}
	return result, nil
}

// Serialize converts a map back to KEY=VALUE line format (sorted alphabetically).
func Serialize(vars map[string]string) string {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	// sort
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, vars[k]))
	}
	return sb.String()
}

// Write writes KEY=VALUE content to a file.
func Write(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

// ToLines converts a map to KEY=VALUE newline-joined string (for API import).
func ToLines(vars map[string]string) string {
	return Serialize(vars)
}
