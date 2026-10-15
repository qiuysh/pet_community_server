// server/util/validator.go
package utils

import (
	"regexp"
	"strings"
)

type ValidationRule struct {
	Required bool
	MinLen   int
	MaxLen   int
	Pattern  *regexp.Regexp
	Message  string
}

var emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func ValidateString(value string, rules ...ValidationRule) (bool, string) {
	for _, rule := range rules {
		if rule.Required && strings.TrimSpace(value) == "" {
			return false, rule.Message
		}

		if value != "" {
			if rule.MinLen > 0 && len(value) < rule.MinLen {
				return false, rule.Message
			}

			if rule.MaxLen > 0 && len(value) > rule.MaxLen {
				return false, rule.Message
			}

			if rule.Pattern != nil && !rule.Pattern.MatchString(value) {
				return false, rule.Message
			}
		}
	}
	return true, ""
}

func ValidateEmail(email string) bool {
	return emailPattern.MatchString(email)
}

func ValidatePhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

func ValidatePassword(password string) bool {
	return len(password) >= 6 && len(password) <= 50
}
