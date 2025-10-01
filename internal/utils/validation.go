package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// UsernameRegex validates Linux usernames (alphanumeric, underscore, hyphen, 1-32 chars)
	UsernameRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

	// PathRegex validates basic file paths (no special shell characters)
	PathRegex = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

	// IQNRegex validates iSCSI Qualified Names
	IQNRegex = regexp.MustCompile(`^iqn\.\d{4}-\d{2}\.[a-z0-9.-]+:[a-z0-9._-]+$`)
)

// ValidateUsername checks if a username is valid for Linux systems
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) > 32 {
		return fmt.Errorf("username too long (max 32 characters)")
	}

	if !UsernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username format: must start with lowercase letter or underscore, contain only lowercase letters, digits, underscores, and hyphens")
	}

	// Reject reserved usernames
	reserved := []string{"root", "bin", "daemon", "sys", "adm", "nobody", "sshd"}
	for _, r := range reserved {
		if username == r {
			return fmt.Errorf("username '%s' is reserved", username)
		}
	}

	return nil
}

// ValidatePassword checks password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password too long (max 128 characters)")
	}

	// Check for at least one letter and one digit
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasLetter || !hasDigit {
		return fmt.Errorf("password must contain at least one letter and one digit")
	}

	return nil
}

// ValidatePath checks if a path is safe (no command injection risk)
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check for dangerous characters
	dangerous := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r"}
	for _, char := range dangerous {
		if strings.Contains(path, char) {
			return fmt.Errorf("path contains forbidden character: %s", char)
		}
	}

	// Prevent directory traversal
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal (..) not allowed")
	}

	return nil
}

// ValidateIQN validates iSCSI Qualified Name format
func ValidateIQN(iqn string) error {
	if iqn == "" {
		return fmt.Errorf("IQN cannot be empty")
	}

	if !strings.HasPrefix(iqn, "iqn.") {
		return fmt.Errorf("IQN must start with 'iqn.'")
	}

	if !IQNRegex.MatchString(iqn) {
		return fmt.Errorf("invalid IQN format")
	}

	return nil
}

// SanitizeString removes potentially dangerous characters from input
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters
	cleaned := strings.Builder{}
	for _, r := range input {
		if r >= 32 || r == '\t' || r == '\n' {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}

// ValidatePageParams validates pagination parameters
func ValidatePageParams(page, pageSize int) (int, int, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		return 0, 0, fmt.Errorf("page size too large (max 100)")
	}

	return page, pageSize, nil
}

// ParsePositiveInt parses a string to a positive integer
func ParsePositiveInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %w", err)
	}

	if i < 0 {
		return 0, fmt.Errorf("must be non-negative")
	}

	return i, nil
}
