package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GenerateUUID generates a new UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// GetEnv gets an environment variable with a fallback value
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// GetEnvAsInt gets an environment variable as integer with a fallback value
func GetEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// GetEnvAsBool gets an environment variable as boolean with a fallback value
func GetEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// FormatTime formats time to RFC3339 string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseTime parses RFC3339 time string
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, timeStr)
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes and other potentially dangerous characters
	return strings.Map(func(r rune) rune {
		if r == 0 || r == '\r' || r == '\n' || r == '\t' {
			return -1
		}
		return r
	}, input)
}

// TruncateString truncates a string to specified length
func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength]
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Unique removes duplicate strings from slice
func Unique(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// ToSnakeCase converts camelCase to snake_case
func ToSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r-'A'+'a')
	}
	return string(result)
}

// ToCamelCase converts snake_case to camelCase
func ToCamelCase(str string) string {
	parts := strings.Split(str, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// MaskString masks sensitive information (e.g., credit card numbers, secrets)
func MaskString(input string, visibleStart, visibleEnd int) string {
	if len(input) <= visibleStart+visibleEnd {
		return input
	}

	masked := strings.Repeat("*", len(input)-visibleStart-visibleEnd)
	return input[:visibleStart] + masked + input[len(input)-visibleEnd:]
}

// IsValidJSON checks if a string is valid JSON (basic check)
func IsValidJSON(str string) bool {
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// BuildURL builds a URL from base URL and path components
func BuildURL(baseURL string, paths ...string) string {
	result := strings.TrimRight(baseURL, "/")

	for _, path := range paths {
		path = strings.Trim(path, "/")
		if path != "" {
			result += "/" + path
		}
	}

	return result
}

// Retry executes a function with retry logic
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("failed after %d attempts: %w", attempts, err)
}

// WithTimeout executes a function with a timeout
func WithTimeout(timeout time.Duration, fn func() error) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}
