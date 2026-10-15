package utils

import (
	"regexp"
	"testing"
)

func TestValidateString(t *testing.T) {
	pattern := regexp.MustCompile(`^\d+$`)

	tests := []struct {
		name   string
		value  string
		rules  []ValidationRule
		wantOK bool
	}{
		{
			name:   "required pass",
			value:  "hello",
			rules:  []ValidationRule{{Required: true, Message: "required"}},
			wantOK: true,
		},
		{
			name:   "required fail",
			value:  "",
			rules:  []ValidationRule{{Required: true, Message: "required"}},
			wantOK: false,
		},
		{
			name:   "whitespace only required fail",
			value:  "   ",
			rules:  []ValidationRule{{Required: true, Message: "required"}},
			wantOK: false,
		},
		{
			name:   "min length fail",
			value:  "hi",
			rules:  []ValidationRule{{MinLen: 3, Message: "too short"}},
			wantOK: false,
		},
		{
			name:   "max length fail",
			value:  "hello world",
			rules:  []ValidationRule{{MaxLen: 5, Message: "too long"}},
			wantOK: false,
		},
		{
			name:   "pattern fail",
			value:  "abc",
			rules:  []ValidationRule{{Pattern: pattern, Message: "not numeric"}},
			wantOK: false,
		},
		{
			name:   "pattern pass",
			value:  "123",
			rules:  []ValidationRule{{Pattern: pattern, Message: "not numeric"}},
			wantOK: true,
		},
		{
			name:   "empty value skips optional rules",
			value:  "",
			rules:  []ValidationRule{{MinLen: 3, Message: "too short"}},
			wantOK: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, _ := ValidateString(tc.value, tc.rules...)
			if ok != tc.wantOK {
				t.Errorf("ValidateString(%q) = %v, want %v", tc.value, ok, tc.wantOK)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid", "test@example.com", true},
		{"missing at", "testexample.com", false},
		{"missing domain", "test@", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidateEmail(tc.email); got != tc.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tc.email, got, tc.want)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{"valid", "13800138000", true},
		{"invalid prefix", "12800138000", false},
		{"too short", "1380013800", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidatePhone(tc.phone); got != tc.want {
				t.Errorf("ValidatePhone(%q) = %v, want %v", tc.phone, got, tc.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"valid", "password123", true},
		{"too short", "12345", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidatePassword(tc.password); got != tc.want {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tc.password, got, tc.want)
			}
		})
	}
}
