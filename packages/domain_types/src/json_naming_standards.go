package src

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrNonCamelCaseKey = errors.New("json_standards: property keys must follow strict lowerCamelCase")
	ErrSnakeCaseKey    = errors.New("json_standards: snake_case keys are forbidden in REST API JSON payloads")
)

// IsStrictLowerCamelCase verifies that a JSON field key conforms to lowerCamelCase.
// 1. First character must be lowercase ASCII letter.
// 2. Contains no underscores or hyphens.
// 3. Allows alphanumeric characters.
func IsStrictLowerCamelCase(key string) bool {
	if len(key) == 0 {
		return false
	}

	runes := []rune(key)
	if !unicode.IsLower(runes[0]) || !unicode.IsLetter(runes[0]) {
		return false
	}

	for _, r := range runes {
		if r == '_' || r == '-' {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ValidateJSONPayloadKeys recursively checks that all keys in a JSON byte payload adhere to lowerCamelCase.
func ValidateJSONPayloadKeys(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var root interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid json payload: %w", err)
	}

	return validateNodeKeys(root, "")
}

func validateNodeKeys(node interface{}, currentPath string) error {
	switch v := node.(type) {
	case map[string]interface{}:
		for k, child := range v {
			fieldPath := k
			if currentPath != "" {
				fieldPath = currentPath + "." + k
			}

			if strings.Contains(k, "_") {
				return fmt.Errorf("%w: field '%s' uses snake_case", ErrSnakeCaseKey, fieldPath)
			}
			if !IsStrictLowerCamelCase(k) {
				return fmt.Errorf("%w: field '%s' is not lowerCamelCase", ErrNonCamelCaseKey, fieldPath)
			}

			if err := validateNodeKeys(child, fieldPath); err != nil {
				return err
			}
		}
	case []interface{}:
		for i, item := range v {
			indexPath := fmt.Sprintf("%s[%d]", currentPath, i)
			if err := validateNodeKeys(item, indexPath); err != nil {
				return err
			}
		}
	}
	return nil
}
