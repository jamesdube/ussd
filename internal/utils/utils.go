package utils

import (
	"regexp"
	"strings"
)

// Constants for USSD response messages and XML formatting.
const MenuInvalidSelection = "Invalid menu option"
const MenuNoMoreOptions = "Invalid menu option"

const Header = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

// StringToSlice splits a dot-separated route string into segments.
func StringToSlice(s string) []string {
	return strings.Split(s, ".")
}

// IsLongCode checks if a message is a USSD long code (e.g., *123*1*1*100#).
func IsLongCode(message string) bool {
	longCodePattern := `^\*[\d*]+#$`
	matched, _ := regexp.MatchString(longCodePattern, message)
	return matched
}

// ParseLongCode parses a USSD long code into its components.
// Example: "*123*1*1*100#" -> ["123", "1", "1", "100"]
func ParseLongCode(longCode string) []string {
	if !IsLongCode(longCode) {
		return nil
	}
	cleaned := strings.TrimPrefix(longCode, "*")
	cleaned = strings.TrimSuffix(cleaned, "#")
	if cleaned == "" {
		return []string{}
	}
	return strings.Split(cleaned, "*")
}

// BuildLongCodeRoute creates a route pattern from long code components.
// Example: ["123", "1", "1", "100"] -> "123.1.1.100"
func BuildLongCodeRoute(components []string) string {
	return strings.Join(components, ".")
}
