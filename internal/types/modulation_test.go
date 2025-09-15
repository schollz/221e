package types

import (
	"testing"
)

// Test that modulation column is correctly positioned
func TestInstrumentModulationColumn(t *testing.T) {
	// Test that MO column is at position 3
	if int(InstrumentColMO) != 3 {
		t.Errorf("Expected InstrumentColMO to be at position 3, got %d", int(InstrumentColMO))
	}
	
	// Test that C column moved to position 4
	if int(InstrumentColC) != 4 {
		t.Errorf("Expected InstrumentColC to be at position 4, got %d", int(InstrumentColC))
	}
	
	// Test that columns after are correctly shifted
	if int(InstrumentColVE) != 7 {
		t.Errorf("Expected InstrumentColVE to be at position 7, got %d", int(InstrumentColVE))
	}
	
	// Test that final column is at position 20
	if int(InstrumentColSO) != 20 {
		t.Errorf("Expected InstrumentColSO to be at position 20, got %d", int(InstrumentColSO))
	}
}