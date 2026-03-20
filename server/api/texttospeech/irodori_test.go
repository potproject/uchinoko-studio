package texttospeech

import "testing"

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
