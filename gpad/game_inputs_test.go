package gpad

import (
	"testing"
)

// TestTouchDpadMovement verifies that if the internal touch state is active,
// the public directional APIs return the correct boolean states.
func TestTouchDpadMovement(t *testing.T) {
	// 1. Arrange: Explicitly manipulate internal, unexported package variables
	// This simulates what UpdateTouch() would calculate from a swipe gesture.
	dpadTouch.up = true
	dpadTouch.down = false
	dpadTouch.left = false
	dpadTouch.right = false

	// 2. Act & Assert: Verify the public API responds correctly
	if !MoveUp() {
		t.Errorf("MoveUp() should be true when dpadTouch.up is enabled")
	}
	if MoveDown() {
		t.Errorf("MoveDown() should be false when dpadTouch.down is disabled")
	}

	// 3. Clean up / Reset state for subsequent tests
	dpadTouch.up = false
}

// TestTouchButtonB verifies that a touch-driven B button trigger executes properly.
func TestTouchButtonB(t *testing.T) {
	// Arrange
	bButtonTouch = true

	// Act & Assert
	if !PressB() {
		t.Errorf("PressB() should be true when bButtonTouch is active")
	}

	// Clean up
	bButtonTouch = false
}

func TestAbs(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Negative integer", -5, 5},
		{"Positive integer", 10, 10},
		{"Zero value", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := abs(tt.input)
			if result != tt.expected {
				t.Errorf("abs(%d) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
