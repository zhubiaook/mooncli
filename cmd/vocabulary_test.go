package cmd

import "testing"

func TestValidatePronunciationRepeatAcceptsOneToOneHundred(t *testing.T) {
	for repeat := 1; repeat <= 100; repeat++ {
		if err := validatePronunciationRepeat(repeat); err != nil {
			t.Fatalf("validatePronunciationRepeat(%d) returned error: %v", repeat, err)
		}
	}
}

func TestValidatePronunciationRepeatRejectsOutOfRangeValues(t *testing.T) {
	for _, repeat := range []int{0, -1, 101} {
		if err := validatePronunciationRepeat(repeat); err == nil {
			t.Fatalf("validatePronunciationRepeat(%d) returned nil error", repeat)
		}
	}
}

func TestValidateVocabularyOptionsAllowsInvalidRepeatWhenSpeechDisabled(t *testing.T) {
	if err := validateVocabularyOptions(true, 0); err != nil {
		t.Fatalf("validateVocabularyOptions(true, 0) returned error: %v", err)
	}
}
