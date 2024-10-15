package generator

import (
	"testing"
)

func TestGenerateRandomCode(t *testing.T) {
	tests := []struct {
		length    int
		expectLen int
	}{
		{5, 5},
		{10, 10},
		{15, 15},
		{0, 0},
	}

	for _, tt := range tests {
		t.Run("Length "+string(rune(tt.length)), func(t *testing.T) {
			code := GenerateRandomCode(tt.length)

			if len(code) != tt.expectLen {
				t.Errorf("expected length %d, got %d", tt.expectLen, len(code))
			}

			for _, char := range code {
				if !isValidChar(char) {
					t.Errorf("generated code contains invalid character: %c", char)
				}
			}
		})
	}
}

func isValidChar(c rune) bool {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for _, validChar := range charset {
		if c == validChar {
			return true
		}
	}
	return false
}
