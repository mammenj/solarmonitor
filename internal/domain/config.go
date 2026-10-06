package domain

import (
	"os"
	"strings"
)

// loadEnv manually reads a .env file and injects variables into the environment
func LoadEnv(filepath string) error {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		// Clean up spacing and skip empty lines or comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on the first '=' sign only
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Optional: Remove surrounding quotes if present
		value = strings.Trim(value, `"'`)

		// Inject into the runtime environment
		os.Setenv(key, value)
	}
	return nil
}
