package domain

import (
	"strings"
	"testing"
)

func compareArrayStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if strings.Compare(a[i], b[i]) != 0 {
			return false
		}
	}
	return true
}

func TestGenerateFizzbuzz(t *testing.T) {
	tests := []struct {
		name     string
		input    FizzbuzzParams
		expected []string
	}{
		{
			name: "example",
			input: FizzbuzzParams{
				Int1:  3,
				Int2:  5,
				Limit: 16,
				Str1:  "fizz",
				Str2:  "buzz",
			},
			expected: []string{"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz", "16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fizzbuzzed := GenerateFizzbuzz(&tt.input)
			if !compareArrayStrings(tt.expected, fizzbuzzed) {
				t.Fatalf("wrong output in test: %v, expected: %v, got: %v", tt.name, tt.expected, fizzbuzzed)
			}
		})
	}
}

func TestValidateFizzbuzzParams(t *testing.T) {
	tests := []struct {
		name         string
		int1         int
		int2         int
		limit        int
		str1         string
		str2         string
		expectsError bool
	}{
		{
			name:         "no error",
			int1:         3,
			int2:         5,
			limit:        16,
			str1:         "fizz",
			str2:         "buzz",
			expectsError: false,
		},
		{
			name:         "negative int",
			int1:         -3,
			int2:         5,
			limit:        16,
			str1:         "fizz",
			str2:         "buzz",
			expectsError: true,
		},
		{
			name:         "empty string",
			int1:         3,
			int2:         5,
			limit:        16,
			str1:         "",
			str2:         "buzz",
			expectsError: true,
		},
		{
			name:         "limit at max",
			int1:         3,
			int2:         5,
			limit:        100,
			str1:         "fizz",
			str2:         "buzz",
			expectsError: false,
		},
		{
			name:         "limit over max",
			int1:         3,
			int2:         5,
			limit:        101,
			str1:         "fizz",
			str2:         "buzz",
			expectsError: true,
		},
		{
			name:         "string over max length",
			int1:         3,
			int2:         5,
			limit:        16,
			str1:         "fizzfizzfiz",
			str2:         "buzz",
			expectsError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fizzbuzzParams := FizzbuzzParams{
				Int1:  tt.int1,
				Int2:  tt.int2,
				Limit: tt.limit,
				Str1:  tt.str1,
				Str2:  tt.str2,
			}

			err := fizzbuzzParams.Validate(Limits{MaxLimit: 100, MaxStrLength: 10})
			if tt.expectsError && err == nil {
				t.Fatalf("expects error in test: %v but got no error", tt.name)
			} else if !tt.expectsError && err != nil {
				t.Fatalf("expects no-error in test: %v, but got error: %v", tt.name, err)
			}
		})
	}
}
