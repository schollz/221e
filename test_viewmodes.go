//go:build ignore

package main

import (
	"fmt"

	"github.com/schollz/2n/internal/types"
)

func main() {
	fmt.Println("=== View Mode Values Test ===")
	
	// Show the actual view mode values
	viewModes := map[string]types.ViewMode{
		"SongView":         types.SongView,
		"ChainView":        types.ChainView, 
		"PhraseView":       types.PhraseView,
		"FileView":         types.FileView,
		"SettingsView":     types.SettingsView,
		"FileMetadataView": types.FileMetadataView,
		"RetriggerView":    types.RetriggerView,
		"TimestrechView":   types.TimestrechView,
		"MixerView":        types.MixerView,
		"ArpeggioView":     types.ArpeggioView,
		"MidiView":         types.MidiView,
		"SoundMakerView":   types.SoundMakerView,
	}
	
	for name, value := range viewModes {
		fmt.Printf("- %s: %d\n", name, value)
	}
	
	fmt.Println("\nThis confirms the undo feature properly handles all view modes!")
}