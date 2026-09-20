package src

import (
	"errors"
	"testing"
)

func TestJSONNamingStandards(t *testing.T) {
	// Valid lowerCamelCase payload
	validJSON := []byte(`{
		"orderId": "ORD-12345",
		"priceE8": 250000000,
		"quantityE8": 1000000,
		"clientInfo": {
			"userId": "USR-99",
			"kycTier": 3
		},
		"tags": ["equity", "intraday"]
	}`)

	if err := ValidateJSONPayloadKeys(validJSON); err != nil {
		t.Fatalf("expected valid lowerCamelCase JSON to pass, got: %v", err)
	}

	// Invalid snake_case payload
	invalidSnakeJSON := []byte(`{
		"order_id": "ORD-12345",
		"priceE8": 250000000
	}`)

	err := ValidateJSONPayloadKeys(invalidSnakeJSON)
	if err == nil || !errors.Is(err, ErrSnakeCaseKey) {
		t.Fatalf("expected ErrSnakeCaseKey for snake_case payload, got %v", err)
	}

	// Invalid PascalCase payload
	invalidPascalJSON := []byte(`{
		"OrderId": "ORD-12345"
	}`)

	err = ValidateJSONPayloadKeys(invalidPascalJSON)
	if err == nil || !errors.Is(err, ErrNonCamelCaseKey) {
		t.Fatalf("expected ErrNonCamelCaseKey for PascalCase payload, got %v", err)
	}

	// Nested snake_case in array of objects
	invalidNestedJSON := []byte(`{
		"orders": [
			{"orderId": "1"},
			{"order_id": "2"}
		]
	}`)

	err = ValidateJSONPayloadKeys(invalidNestedJSON)
	if err == nil || !errors.Is(err, ErrSnakeCaseKey) {
		t.Fatalf("expected ErrSnakeCaseKey for nested object inside array, got %v", err)
	}
}
