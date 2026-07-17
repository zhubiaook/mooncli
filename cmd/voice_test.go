package cmd

import "testing"

func TestValidateVoiceOptionsAcceptsRepeatAndInterval(t *testing.T) {
	if err := validateVoiceOptions(5, 0, 1, false); err != nil {
		t.Fatalf("validateVoiceOptions(5, 0, 1, false) returned error: %v", err)
	}
}

func TestValidateVoiceOptionsRejectsInvalidRepeat(t *testing.T) {
	if err := validateVoiceOptions(101, 0, 1, false); err == nil {
		t.Fatal("validateVoiceOptions(101, 0, 1, false) returned nil error")
	}
}

func TestValidateVoiceOptionsRejectsNegativeInterval(t *testing.T) {
	if err := validateVoiceOptions(5, -1, 1, false); err == nil {
		t.Fatal("validateVoiceOptions(5, -1, 1, false) returned nil error")
	}
}

func TestValidateVoiceOptionsRejectsInvalidVolumeOverride(t *testing.T) {
	if err := validateVoiceOptions(5, 0, 11, true); err == nil {
		t.Fatal("validateVoiceOptions(5, 0, 11, true) returned nil error")
	}
}
