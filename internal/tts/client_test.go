package tts

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/zhubiaook/moonai/internal/config"
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

func TestNewClientRejectsInvalidConfiguredVolume(t *testing.T) {
	cfg := validTTSConfig()
	cfg.Volume = "11"

	if _, err := NewClient(cfg); err == nil {
		t.Fatal("NewClient returned nil error")
	}
}

func TestNewClientWithVolumeOverridesConfiguredVolume(t *testing.T) {
	cfg := validTTSConfig()
	cfg.Volume = "bad"

	client, err := NewClientWithVolume(cfg, 2)
	if err != nil {
		t.Fatalf("NewClientWithVolume returned error: %v", err)
	}
	if client.volume() != 2 {
		t.Fatalf("volume = %v, want 2", client.volume())
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

func TestPlayMP3WaitsBetweenRepeats(t *testing.T) {
	calls := 0
	var sleeps []time.Duration

	err := playMP3WithPlayerInterval(t.Context(), []byte("audio"), 3, func(path string) error {
		calls++
		return nil
	}, time.Second, DefaultVolume, func(d time.Duration) {
		sleeps = append(sleeps, d)
	})
	if err != nil {
		t.Fatalf("playMP3WithPlayerInterval returned error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("sleeps = %d, want 2", len(sleeps))
	}
	for _, sleep := range sleeps {
		if sleep != time.Second {
			t.Fatalf("sleep = %v, want %v", sleep, time.Second)
		}
	}
}

func TestPlayMP3AllowsZeroInterval(t *testing.T) {
	calls := 0
	sleeps := 0

	err := playMP3WithPlayerInterval(t.Context(), []byte("audio"), 2, func(path string) error {
		calls++
		return nil
	}, 0, DefaultVolume, func(time.Duration) {
		sleeps++
	})
	if err != nil {
		t.Fatalf("playMP3WithPlayerInterval returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if sleeps != 0 {
		t.Fatalf("sleeps = %d, want 0", sleeps)
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

func TestPlayMP3RejectsNegativeInterval(t *testing.T) {
	err := playMP3WithPlayerInterval(t.Context(), []byte("audio"), 1, func(path string) error {
		t.Fatal("player should not be called")
		return nil
	}, -time.Second, DefaultVolume, func(time.Duration) {
		t.Fatal("sleep should not be called")
	})
	if err == nil {
		t.Fatal("playMP3WithPlayerInterval returned nil error")
	}
}

func TestPlayMP3RejectsInvalidVolume(t *testing.T) {
	err := playMP3WithPlayerInterval(t.Context(), []byte("audio"), 1, func(path string) error {
		t.Fatal("player should not be called")
		return nil
	}, 0, 11, func(time.Duration) {
		t.Fatal("sleep should not be called")
	})
	if err == nil {
		t.Fatal("playMP3WithPlayerInterval returned nil error")
	}
}

func TestAfplayArgsIncludeVolume(t *testing.T) {
	args := afplayArgs("/tmp/audio.mp3", 2.5)
	want := []string{"-v", "2.5", "/tmp/audio.mp3"}

	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for i := range args {
		if args[i] != want[i] {
			t.Fatalf("args = %v, want %v", args, want)
		}
	}
}

func validTTSConfig() config.TTSConfig {
	return config.TTSConfig{
		APIKey:     "key",
		ResourceID: "resource",
		VoiceType:  "voice",
	}
}
