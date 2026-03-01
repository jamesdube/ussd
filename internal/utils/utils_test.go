package utils

import (
	"reflect"
	"testing"
)

// Test IsLongCode function
func TestIsLongCode(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected bool
	}{
		{
			name:     "Valid long code with numbers",
			message:  "*123*1*1*100#",
			expected: true,
		},
		{
			name:     "Valid long code with single digit",
			message:  "*1#",
			expected: true,
		},
		{
			name:     "Valid long code complex",
			message:  "*144*1*2*555*999#",
			expected: true,
		},
		{
			name:     "Valid long code balance check",
			message:  "*131#",
			expected: true,
		},
		{
			name:     "Invalid - no leading asterisk",
			message:  "123*1*1*100#",
			expected: false,
		},
		{
			name:     "Invalid - no trailing hash",
			message:  "*123*1*1*100",
			expected: false,
		},
		{
			name:     "Invalid - contains letters",
			message:  "*123*a*1*100#",
			expected: false,
		},
		{
			name:     "Invalid - regular number",
			message:  "123",
			expected: false,
		},
		{
			name:     "Invalid - empty string",
			message:  "",
			expected: false,
		},
		{
			name:     "Invalid - just asterisk and hash",
			message:  "*#",
			expected: false,
		},
		{
			name:     "Valid - multiple asterisks in sequence",
			message:  "*123**456#",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsLongCode(tt.message)
			if result != tt.expected {
				t.Errorf("IsLongCode(%q) = %v, expected %v", tt.message, result, tt.expected)
			}
		})
	}
}

// Test ParseLongCode function
func TestParseLongCode(t *testing.T) {
	tests := []struct {
		name     string
		longCode string
		expected []string
	}{
		{
			name:     "Parse balance check code",
			longCode: "*131#",
			expected: []string{"131"},
		},
		{
			name:     "Parse transfer code",
			longCode: "*123*1*1*100#",
			expected: []string{"123", "1", "1", "100"},
		},
		{
			name:     "Parse complex code",
			longCode: "*144*1*2*555*999#",
			expected: []string{"144", "1", "2", "555", "999"},
		},
		{
			name:     "Parse single digit code",
			longCode: "*1#",
			expected: []string{"1"},
		},
		{
			name:     "Invalid code returns nil",
			longCode: "123*1*1*100#",
			expected: nil,
		},
		{
			name:     "Empty segments in code",
			longCode: "*123**456#",
			expected: []string{"123", "", "456"},
		},
		{
			name:     "Just asterisk and hash",
			longCode: "*#",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLongCode(tt.longCode)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseLongCode(%q) = %v, expected %v", tt.longCode, result, tt.expected)
			}
		})
	}
}

// Test BuildLongCodeRoute function
func TestBuildLongCodeRoute(t *testing.T) {
	tests := []struct {
		name       string
		components []string
		expected   string
	}{
		{
			name:       "Build route from balance check",
			components: []string{"131"},
			expected:   "131",
		},
		{
			name:       "Build route from transfer components",
			components: []string{"123", "1", "1", "100"},
			expected:   "123.1.1.100",
		},
		{
			name:       "Build route from complex components",
			components: []string{"144", "1", "2", "555", "999"},
			expected:   "144.1.2.555.999",
		},
		{
			name:       "Build route from empty components",
			components: []string{},
			expected:   "",
		},
		{
			name:       "Build route with empty segments",
			components: []string{"123", "", "456"},
			expected:   "123..456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildLongCodeRoute(tt.components)
			if result != tt.expected {
				t.Errorf("BuildLongCodeRoute(%v) = %q, expected %q", tt.components, result, tt.expected)
			}
		})
	}
}

// Test StringToSlice function (existing)
func TestStringToSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Simple dot-separated string",
			input:    "1.2.3",
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "Single element",
			input:    "1",
			expected: []string{"1"},
		},
		{
			name:     "Empty string",
			input:    "",
			expected: []string{""},
		},
		{
			name:     "String with empty segments",
			input:    "1..3",
			expected: []string{"1", "", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("StringToSlice(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
