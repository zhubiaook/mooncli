package tts

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/zhubiaook/moonai/internal/config"
)

const (
	defaultEndpoint   = "https://openspeech.bytedance.com/api/v3/tts/unidirectional"
	audioFormat       = "mp3"
	audioSampleRateHz = 24000
	DefaultVolume     = 1.0
	MaxVolume         = 10.0
)

type Client struct {
	cfg        config.TTSConfig
	httpClient *http.Client
}

func NewClient(cfg config.TTSConfig) (*Client, error) {
	volume, err := parseVolume(cfg.Volume)
	if err != nil {
		return nil, err
	}
	return NewClientWithVolume(cfg, volume)
}

func NewClientWithVolume(cfg config.TTSConfig, volume float64) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("VOLCENGINE_TTS_API_KEY not set in settings.json")
	}
	if cfg.ResourceID == "" {
		return nil, errors.New("VOLCENGINE_TTS_RESOURCE_ID not set in settings.json")
	}
	if cfg.VoiceType == "" {
		return nil, errors.New("VOLCENGINE_TTS_VOICE_TYPE not set in settings.json")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	}
	if err := ValidateVolume(volume); err != nil {
		return nil, err
	}
	cfg.Volume = strconv.FormatFloat(volume, 'f', -1, 64)
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) Speak(ctx context.Context, text string) error {
	return c.SpeakRepeat(ctx, text, 1)
}

func (c *Client) SpeakRepeat(ctx context.Context, text string, repeat int) error {
	return c.SpeakRepeatInterval(ctx, text, repeat, 0)
}

func (c *Client) SpeakRepeatInterval(ctx context.Context, text string, repeat int, interval time.Duration) error {
	audio, err := c.Synthesize(ctx, text)
	if err != nil {
		return err
	}
	return PlayMP3IntervalVolume(ctx, audio, repeat, interval, c.volume())
}

func (c *Client) Synthesize(ctx context.Context, text string) ([]byte, error) {
	body := synthesizeRequest{
		User: user{
			UID: "mooncli",
		},
		ReqParams: reqParams{
			Text:    text,
			Speaker: c.cfg.VoiceType,
			AudioParams: audioParams{
				Format:     audioFormat,
				SampleRate: audioSampleRateHz,
			},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal tts request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.Endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create tts request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.cfg.APIKey)
	req.Header.Set("X-Api-Resource-Id", c.cfg.ResourceID)
	req.Header.Set("X-Api-Request-Id", requestID())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call volcengine tts: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tts response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("volcengine tts returned %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}

	audio, err := DecodeAudio(raw)
	if err != nil {
		return nil, err
	}
	return audio, nil
}

func PlayMP3(ctx context.Context, audio []byte, repeat int) error {
	return PlayMP3Interval(ctx, audio, repeat, 0)
}

func PlayMP3Interval(ctx context.Context, audio []byte, repeat int, interval time.Duration) error {
	return PlayMP3IntervalVolume(ctx, audio, repeat, interval, DefaultVolume)
}

func PlayMP3IntervalVolume(ctx context.Context, audio []byte, repeat int, interval time.Duration, volume float64) error {
	return playMP3WithPlayerInterval(ctx, audio, repeat, func(path string) error {
		if _, err := exec.LookPath("afplay"); err != nil {
			return errors.New("afplay not found")
		}
		if err := exec.CommandContext(ctx, "afplay", afplayArgs(path, volume)...).Run(); err != nil {
			return fmt.Errorf("play audio with afplay: %w", err)
		}
		return nil
	}, interval, volume, time.Sleep)
}

func playMP3WithPlayer(ctx context.Context, audio []byte, repeat int, play func(string) error) error {
	return playMP3WithPlayerInterval(ctx, audio, repeat, play, 0, DefaultVolume, time.Sleep)
}

func playMP3WithPlayerInterval(ctx context.Context, audio []byte, repeat int, play func(string) error, interval time.Duration, volume float64, sleep func(time.Duration)) error {
	if repeat < 1 {
		return fmt.Errorf("repeat must be at least 1")
	}
	if interval < 0 {
		return fmt.Errorf("interval must be at least 0")
	}
	if err := ValidateVolume(volume); err != nil {
		return err
	}

	file, err := os.CreateTemp("", "mooncli-*.mp3")
	if err != nil {
		return fmt.Errorf("create temp audio file: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)

	if _, err := file.Write(audio); err != nil {
		file.Close()
		return fmt.Errorf("write temp audio file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temp audio file: %w", err)
	}

	for i := range repeat {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := play(path); err != nil {
			return err
		}
		if i < repeat-1 && interval > 0 {
			sleep(interval)
		}
	}
	return nil
}

func (c *Client) volume() float64 {
	volume, err := parseVolume(c.cfg.Volume)
	if err != nil {
		return DefaultVolume
	}
	return volume
}

func parseVolume(raw string) (float64, error) {
	if raw == "" {
		return DefaultVolume, nil
	}
	volume, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("VOLCENGINE_TTS_VOLUME must be a number between 0 and 10")
	}
	if err := ValidateVolume(volume); err != nil {
		return 0, err
	}
	return volume, nil
}

func ValidateVolume(volume float64) error {
	if volume < 0 || volume > MaxVolume {
		return fmt.Errorf("volume must be between 0 and 10")
	}
	return nil
}

func afplayArgs(path string, volume float64) []string {
	return []string{"-v", strconv.FormatFloat(volume, 'f', -1, 64), path}
}

func DecodeAudio(raw []byte) ([]byte, error) {
	var chunks [][]byte
	var apiErr error

	for _, payload := range splitPayloads(raw) {
		var frame responseFrame
		if err := json.Unmarshal(payload, &frame); err != nil {
			return nil, fmt.Errorf("parse tts response: %w", err)
		}
		if frame.Data != "" {
			chunk, err := base64.StdEncoding.DecodeString(frame.Data)
			if err != nil {
				return nil, fmt.Errorf("decode tts audio chunk: %w", err)
			}
			chunks = append(chunks, chunk)
			continue
		}
		if frame.Code != 0 && frame.Code != 20000000 {
			apiErr = fmt.Errorf("volcengine tts failed: code=%d message=%s", frame.Code, frame.Message)
		}
	}

	if len(chunks) == 0 {
		if apiErr != nil {
			return nil, apiErr
		}
		return nil, errors.New("volcengine tts response contained no audio")
	}
	return bytes.Join(chunks, nil), nil
}

func splitPayloads(raw []byte) [][]byte {
	if bytes.Contains(raw, []byte("data:")) {
		return splitSSEPayloads(raw)
	}
	return splitJSONPayloads(raw)
}

func splitSSEPayloads(raw []byte) [][]byte {
	var payloads [][]byte
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		payloads = append(payloads, []byte(payload))
	}
	return payloads
}

func splitJSONPayloads(raw []byte) [][]byte {
	var payloads [][]byte
	decoder := json.NewDecoder(bytes.NewReader(raw))
	for {
		var payload json.RawMessage
		if err := decoder.Decode(&payload); err != nil {
			break
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

func requestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("mooncli-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

type synthesizeRequest struct {
	User      user      `json:"user"`
	ReqParams reqParams `json:"req_params"`
}

type user struct {
	UID string `json:"uid"`
}

type reqParams struct {
	Text        string      `json:"text"`
	Speaker     string      `json:"speaker"`
	AudioParams audioParams `json:"audio_params"`
}

type audioParams struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
}

type responseFrame struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}
