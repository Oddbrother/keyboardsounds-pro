package rules

import (
	"path/filepath"
	"testing"
)

func TestDefaultProfilesPersist(t *testing.T) {
	dir := t.TempDir()

	if err := LoadRules(dir); err != nil {
		t.Fatalf("LoadRules() error = %v", err)
	}

	keyboard := "custom-keyboard"
	mouse := "custom-mouse"
	want := Profiles{Keyboard: &keyboard, Mouse: &mouse}
	if err := SetDefaultProfiles(want); err != nil {
		t.Fatalf("SetDefaultProfiles() error = %v", err)
	}

	if err := LoadRules(filepath.Clean(dir)); err != nil {
		t.Fatalf("LoadRules() after save error = %v", err)
	}

	got := GetDefaultProfiles()
	if got.Keyboard == nil || *got.Keyboard != keyboard {
		t.Fatalf("keyboard profile = %v, want %q", got.Keyboard, keyboard)
	}
	if got.Mouse == nil || *got.Mouse != mouse {
		t.Fatalf("mouse profile = %v, want %q", got.Mouse, mouse)
	}
}
