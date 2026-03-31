package texttospeech

import (
	"reflect"
	"testing"
)

func TestResolveLocalReferenceAudioPath(t *testing.T) {
	t.Parallel()

	got, err := resolveLocalReferenceAudioPath("refs/custom.wav")
	if err != nil {
		t.Fatalf("resolveLocalReferenceAudioPath returned error: %v", err)
	}
	if got != "refs\\custom.wav" && got != "refs/custom.wav" {
		t.Fatalf("unexpected resolved path: %q", got)
	}
}

func TestResolveLocalReferenceAudioPathRejectsTraversal(t *testing.T) {
	t.Parallel()

	if _, err := resolveLocalReferenceAudioPath("../custom.wav"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

func TestParseIrodoriUploadResponseArray(t *testing.T) {
	t.Parallel()

	source := &irodoriReferenceAudioSource{
		Data:     []byte("audio"),
		FileName: "custom.wav",
		MimeType: "audio/wav",
	}

	fileData, err := parseIrodoriUploadResponse([]byte(`["/tmp/gradio/custom.wav"]`), source)
	if err != nil {
		t.Fatalf("parseIrodoriUploadResponse returned error: %v", err)
	}

	if fileData.Path != "/tmp/gradio/custom.wav" {
		t.Fatalf("unexpected uploaded path: %q", fileData.Path)
	}
	if fileData.OrigName != "custom.wav" {
		t.Fatalf("unexpected original name: %q", fileData.OrigName)
	}
	if fileData.MimeType != "audio/wav" {
		t.Fatalf("unexpected mime type: %q", fileData.MimeType)
	}
	if fileData.Meta.Type != "gradio.FileData" {
		t.Fatalf("unexpected meta type: %q", fileData.Meta.Type)
	}
}

func TestBuildDirectIrodoriReferenceAudio(t *testing.T) {
	t.Parallel()

	fileData := buildDirectIrodoriReferenceAudio("https://example.com/audio/custom.wav")

	if fileData.Path != "https://example.com/audio/custom.wav" {
		t.Fatalf("unexpected path: %q", fileData.Path)
	}
	if fileData.URL != "https://example.com/audio/custom.wav" {
		t.Fatalf("unexpected url: %q", fileData.URL)
	}
	if fileData.OrigName != "custom.wav" {
		t.Fatalf("unexpected orig_name: %q", fileData.OrigName)
	}
	if fileData.Meta.Type != "gradio.FileData" {
		t.Fatalf("unexpected meta type: %q", fileData.Meta.Type)
	}
}

func TestBuildIrodoriPayloadMatchesCurrentRunGenerationOrder(t *testing.T) {
	t.Parallel()

	refInput := map[string]any{
		"path": "/tmp/gradio/reference.wav",
	}

	payload := buildIrodoriPayload("Aratako/Irodori-TTS-500M-v2", "こんにちは", refInput)

	if got, want := len(payload), 23; got != want {
		t.Fatalf("unexpected payload length: got %d want %d", got, want)
	}
	if got := payload[5]; got != "こんにちは" {
		t.Fatalf("unexpected text position: %#v", got)
	}
	if got := payload[6]; !reflect.DeepEqual(got, refInput) {
		t.Fatalf("unexpected uploaded_audio position: %#v", got)
	}
	if got := payload[10]; got != "independent" {
		t.Fatalf("unexpected cfg_guidance_mode: %#v", got)
	}
	if got := payload[21]; got != "0.9" {
		t.Fatalf("unexpected speaker_kv_min_t_raw: %#v", got)
	}
	if got := payload[22]; got != "" {
		t.Fatalf("unexpected speaker_kv_max_layers_raw: %#v", got)
	}
}
