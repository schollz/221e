package supercollider

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/schollz/collidertracker/internal/types"
)

//go:embed dx7.json
var dx7PatchesData []byte

var dx7Patches []string
var dx7PatchMap map[string]int

func init() {
	if err := json.Unmarshal(dx7PatchesData, &dx7Patches); err != nil {
		panic(fmt.Sprintf("Failed to load DX7 patches: %v", err))
	}

	dx7PatchMap = make(map[string]int, len(dx7Patches))
	for i, patch := range dx7Patches {
		dx7PatchMap[patch] = i
	}
}

func GetDX7PatchCount() int {
	return len(dx7Patches)
}

func GetDX7PatchName(index int) (string, error) {
	if index < 0 || index >= len(dx7Patches) {
		return "", fmt.Errorf("patch index %d out of range [0, %d]", index, len(dx7Patches)-1)
	}
	return dx7Patches[index], nil
}

func GetDX7PatchIndex(name string) (int, error) {
	if index, exists := dx7PatchMap[name]; exists {
		return index, nil
	}

	for patchName, index := range dx7PatchMap {
		if strings.EqualFold(patchName, name) {
			return index, nil
		}
	}

	return -1, fmt.Errorf("patch name %q not found", name)
}

func GetAllDX7PatchNames() []string {
	result := make([]string, len(dx7Patches))
	copy(result, dx7Patches)
	return result
}

func SetDX7PatchByName(settings *types.SoundMakerSettings, patchName string) error {
	if settings.Name != "DX7" {
		return fmt.Errorf("cannot set DX7 patch on non-DX7 SoundMaker: %s", settings.Name)
	}

	index, err := GetDX7PatchIndex(patchName)
	if err != nil {
		return err
	}

	settings.SetParameterValue("preset", float32(index))
	settings.PatchName = patchName
	return nil
}

func SetDX7PatchByIndex(settings *types.SoundMakerSettings, index int) error {
	if settings.Name != "DX7" {
		return fmt.Errorf("cannot set DX7 patch on non-DX7 SoundMaker: %s", settings.Name)
	}

	patchName, err := GetDX7PatchName(index)
	if err != nil {
		return err
	}

	settings.SetParameterValue("preset", float32(index))
	settings.PatchName = patchName
	return nil
}
