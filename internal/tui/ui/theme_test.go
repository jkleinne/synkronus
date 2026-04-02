package ui

import "testing"

func TestGradientTextEmpty(t *testing.T) {
	result := GradientText("", gradientColors)
	if result != "" {
		t.Errorf("expected empty string for empty input, got %q", result)
	}
}

func TestGradientTextSingleChar(t *testing.T) {
	result := GradientText("X", gradientColors)
	if result == "" {
		t.Error("expected non-empty string for single character")
	}
}

func TestGradientTextLength(t *testing.T) {
	input := "SYNKRONUS"
	result := GradientText(input, gradientColors)
	for _, r := range input {
		found := false
		for _, c := range result {
			if c == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("character %q not found in gradient output", string(r))
		}
	}
}

func TestContentWidthClamp(t *testing.T) {
	tests := []struct {
		termWidth int
		expected  int
	}{
		{50, 40},
		{100, 90},
		{110, 90},
		{20, 10},
	}
	for _, tt := range tests {
		got := ContentWidth(tt.termWidth)
		if got != tt.expected {
			t.Errorf("ContentWidth(%d) = %d, expected %d", tt.termWidth, got, tt.expected)
		}
	}
}
