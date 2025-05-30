package monitor

import "testing"

func TestDefaultDecoder(t *testing.T) {
	inputStr := "1234567890abcdef1234567890abff00"
	input := []byte(inputStr)

	decoder := DefaultDecoder()

	result, err := decoder(input)
	if err != nil {
		t.Errorf("Error decoding: %v", err)
	}

	if result.Temperature() != -10.0 {
		t.Errorf("Expected -10.0, but got %.2f", result.Temperature())
	}
}
