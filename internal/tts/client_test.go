package tts

import (
	"encoding/base64"
	"testing"
)

func TestDecodeAudioFromChunkedJSON(t *testing.T) {
	first := base64.StdEncoding.EncodeToString([]byte("abc"))
	second := base64.StdEncoding.EncodeToString([]byte("def"))
	raw := []byte(`{"code":0,"data":"` + first + `"}` + "\n" + `{"code":0,"data":"` + second + `"}` + "\n" + `{"code":20000000,"message":"ok"}`)

	audio, err := DecodeAudio(raw)
	if err != nil {
		t.Fatalf("DecodeAudio returned error: %v", err)
	}
	if string(audio) != "abcdef" {
		t.Fatalf("audio = %q, want %q", string(audio), "abcdef")
	}
}

func TestDecodeAudioFromSSE(t *testing.T) {
	chunk := base64.StdEncoding.EncodeToString([]byte("abc"))
	raw := []byte("event: 352\n" + `data: {"code":0,"data":"` + chunk + `"}` + "\n\n" + "event: 152\n" + `data: {"code":20000000,"message":"ok"}` + "\n\n")

	audio, err := DecodeAudio(raw)
	if err != nil {
		t.Fatalf("DecodeAudio returned error: %v", err)
	}
	if string(audio) != "abc" {
		t.Fatalf("audio = %q, want %q", string(audio), "abc")
	}
}

func TestDecodeAudioReportsAPIErrorWithoutAudio(t *testing.T) {
	_, err := DecodeAudio([]byte(`{"code":55000000,"message":"resource ID is mismatched"}`))
	if err == nil {
		t.Fatal("DecodeAudio returned nil error")
	}
}

func TestPlayMP3RepeatsAudio(t *testing.T) {
	calls := 0

	err := playMP3WithPlayer(t.Context(), []byte("audio"), 3, func(path string) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("playMP3WithPlayer returned error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
}

func TestPlayMP3RejectsInvalidRepeat(t *testing.T) {
	err := playMP3WithPlayer(t.Context(), []byte("audio"), 0, func(path string) error {
		t.Fatal("player should not be called")
		return nil
	})
	if err == nil {
		t.Fatal("playMP3WithPlayer returned nil error")
	}
}
