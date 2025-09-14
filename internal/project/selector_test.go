package project

import (
	"testing"
)

// Test to verify the RunProjectSelector function signature and basic functionality
func TestRunProjectSelectorSignature(t *testing.T) {
	// This test ensures our changes don't break the function signature
	// and that it compiles correctly

	// We can't easily test the UI interaction without mocking,
	// but we can test that the function exists and has the right signature
	_ = RunProjectSelector // This line will fail to compile if the function doesn't exist

	// Test that we can call it (though it will block waiting for user input)
	// In a real test environment, we'd mock the tea program
}

func TestProjectNameInputCreation(t *testing.T) {
	// Test that we can create a ProjectNameInput
	input := NewProjectNameInput()
	
	if input == nil {
		t.Fatal("NewProjectNameInput returned nil")
	}
	
	if input.projectName != "" {
		t.Errorf("Expected empty project name, got %q", input.projectName)
	}
	
	if input.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", input.cursor)
	}
	
	if input.done {
		t.Error("Expected done to be false initially")
	}
	
	if input.cancelled {
		t.Error("Expected cancelled to be false initially")
	}
}