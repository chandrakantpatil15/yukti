package security

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// PasswordRequirement represents a single password requirement
type PasswordRequirement struct {
	Met     bool   `json:"met"`
	Message string `json:"message"`
}

// PasswordValidation represents the complete password validation result
type PasswordValidation struct {
	IsValid      bool                  `json:"is_valid"`
	Score        int                   `json:"score"`
	Requirements []PasswordRequirement `json:"requirements"`
}

// ValidatePassword validates password against enterprise security standards
func ValidatePassword(password string) *PasswordValidation {
	requirements := []PasswordRequirement{
		{
			Met:     len(password) >= 12,
			Message: "At least 12 characters long",
		},
		{
			Met:     hasLowercase(password),
			Message: "Contains lowercase letters",
		},
		{
			Met:     hasUppercase(password),
			Message: "Contains uppercase letters",
		},
		{
			Met:     hasDigit(password),
			Message: "Contains numbers",
		},
		{
			Met:     hasSpecialChar(password),
			Message: "Contains special characters (!@#$%^&*)",
		},
		{
			Met:     !hasConsecutiveChars(password),
			Message: "No more than 2 consecutive identical characters",
		},
		{
			Met:     !isCommonPassword(password),
			Message: "Not a commonly used password",
		},
		{
			Met:     !hasWeakPattern(password),
			Message: "Does not start with common weak patterns",
		},
	}

	metCount := 0
	for _, req := range requirements {
		if req.Met {
			metCount++
		}
	}

	score := (metCount * 100) / len(requirements)
	isValid := metCount == len(requirements)

	return &PasswordValidation{
		IsValid:      isValid,
		Score:        score,
		Requirements: requirements,
	}
}

// ValidatePasswordStrict returns error if password doesn't meet requirements
func ValidatePasswordStrict(password string) error {
	validation := ValidatePassword(password)
	if !validation.IsValid {
		return errors.New("password does not meet security requirements")
	}
	return nil
}

// Helper functions
func hasLowercase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasUppercase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}

func hasConsecutiveChars(s string) bool {
	if len(s) < 3 {
		return false
	}
	
	for i := 0; i < len(s)-2; i++ {
		if s[i] == s[i+1] && s[i+1] == s[i+2] {
			return true
		}
	}
	return false
}

func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456789", "12345678", "1234567890", "qwerty123",
		"password123", "admin123", "welcome123", "letmein123", "monkey123",
		"dragon123", "sunshine123", "master123", "shadow123", "football123",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if strings.Contains(lowerPassword, common) {
			return true
		}
	}
	return false
}

func hasWeakPattern(password string) bool {
	weakPatterns := []string{
		"^password", "^123456", "^qwerty", "^admin", "^user",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, pattern := range weakPatterns {
		matched, _ := regexp.MatchString(pattern, lowerPassword)
		if matched {
			return true
		}
	}
	return false
}