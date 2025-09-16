package types

import (
	"encoding/json"
	"testing"
)

func TestDuckingSettingsThreshField(t *testing.T) {
	// Test that Thresh field is properly included in DuckingSettings
	ds := DuckingSettings{
		Type:    1,
		Bus:     2,
		Attack:  0.1,
		Release: 0.2,
		Depth:   0.5,
		Thresh:  0.03,
	}
	
	// Test JSON serialization includes Thresh
	jsonData, err := json.Marshal(ds)
	if err != nil {
		t.Fatalf("JSON marshal error: %v", err)
	}
	
	jsonStr := string(jsonData)
	if jsonStr == "" {
		t.Fatalf("JSON serialization failed")
	}
	
	// Check that "thresh" is in the JSON
	if !containsSubstring(jsonStr, "thresh") {
		t.Errorf("JSON does not contain 'thresh' field: %s", jsonStr)
	}
	
	// Test JSON deserialization preserves Thresh
	var ds2 DuckingSettings
	err = json.Unmarshal(jsonData, &ds2)
	if err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	
	if ds2.Thresh != 0.03 {
		t.Errorf("Thresh field not preserved: expected 0.03, got %f", ds2.Thresh)
	}
}

func TestDuckingSettingsRowThresh(t *testing.T) {
	// Test that DuckingSettingsRowThresh constant is defined correctly
	if int(DuckingSettingsRowThresh) != 5 {
		t.Errorf("DuckingSettingsRowThresh should be 5, got %d", DuckingSettingsRowThresh)
	}
	
	// Test that all rows are in correct order
	if int(DuckingSettingsRowType) != 0 {
		t.Errorf("DuckingSettingsRowType should be 0, got %d", DuckingSettingsRowType)
	}
	if int(DuckingSettingsRowBus) != 1 {
		t.Errorf("DuckingSettingsRowBus should be 1, got %d", DuckingSettingsRowBus)
	}
	if int(DuckingSettingsRowAttack) != 2 {
		t.Errorf("DuckingSettingsRowAttack should be 2, got %d", DuckingSettingsRowAttack)
	}
	if int(DuckingSettingsRowRelease) != 3 {
		t.Errorf("DuckingSettingsRowRelease should be 3, got %d", DuckingSettingsRowRelease)
	}
	if int(DuckingSettingsRowDepth) != 4 {
		t.Errorf("DuckingSettingsRowDepth should be 4, got %d", DuckingSettingsRowDepth)
	}
	if int(DuckingSettingsRowThresh) != 5 {
		t.Errorf("DuckingSettingsRowThresh should be 5, got %d", DuckingSettingsRowThresh)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s) - len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}