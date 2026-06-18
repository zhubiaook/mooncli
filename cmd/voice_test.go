package cmd

import "testing"

func TestValidateVoiceOptionsAcceptsRepeatAndInterval(t *testing.T) {
	if err := validateVoiceOptions(5, 0); err != nil {
		t.Fatalf("validateVoiceOptions(5, 0) returned error: %v", err)
	}
}

func TestValidateVoiceOptionsRejectsInvalidRepeat(t *testing.T) {
	if err := validateVoiceOptions(11, 0); err == nil {
		t.Fatal("validateVoiceOptions(11, 0) returned nil error")
	}
}

func TestValidateVoiceOptionsRejectsNegativeInterval(t *testing.T) {
	if err := validateVoiceOptions(5, -1); err == nil {
		t.Fatal("validateVoiceOptions(5, -1) returned nil error")
	}
}
