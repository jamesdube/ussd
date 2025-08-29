package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
	
)

func StringToSlice(s string) []string {
	return strings.Split(s, ".")
}

func Convert(b []byte, i interface{}) error {
	err := json.Unmarshal(b, &i)
	if err != nil {
		fmt.Println("Can;t unmarshal the byte array")
		return err
	}
	return nil
}

// IsLongCode checks if a message is a USSD long code (e.g., *123*1*1*100#)
func IsLongCode(message string) bool {
	// USSD long codes start with * and end with #, with * separators
	longCodePattern := `^\*[\d*]+#$`
	matched, _ := regexp.MatchString(longCodePattern, message)
	return matched
}

// ParseLongCode parses a USSD long code into its components
// Example: "*123*1*1*100#" -> ["123", "1", "1", "100"]
func ParseLongCode(longCode string) []string {
	if !IsLongCode(longCode) {
		return nil
	}

	// Remove the leading * and trailing #
	cleaned := strings.TrimPrefix(longCode, "*")
	cleaned = strings.TrimSuffix(cleaned, "#")

	// Split by * to get the components
	if cleaned == "" {
		return []string{}
	}

	return strings.Split(cleaned, "*")
}

// BuildLongCodeRoute creates a route pattern from long code components
// Example: ["123", "1", "1", "100"] -> "123.1.1.100"
func BuildLongCodeRoute(components []string) string {
	return strings.Join(components, ".")
}
