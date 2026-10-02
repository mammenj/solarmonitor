package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// loadEnv manually reads a .env file and injects variables into the environment
func loadEnv(filepath string) error {
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

func main2() {
	// Load the file manually
	if err := loadEnv(".env"); err != nil {
		log.Println("Warning: Could not load .env file:", err)
	}

	// Read your variable normally
	dbUser := os.Getenv("DB_USER")
	fmt.Println("Database User:", dbUser)
}
