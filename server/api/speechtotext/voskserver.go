package speechtotext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gorilla/websocket"
)

func resample(data []byte, sampleRate uint32) ([]byte, error) {
	// Read the WAV file
	reader := bytes.NewReader(data)
	decoder := wav.NewDecoder(reader)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file")
	}

	// Decode the PCM data
	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to decode WAV file: %v", err)
	}

	// Downsample the PCM data
	downsampledBuffer := downsampleBuffer(buffer, int(sampleRate), int(decoder.SampleRate))

	// Create a temporary file to write the downsampled data
	outputFile, err := os.CreateTemp("", "output*.wav")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(outputFile.Name())

	// Encode the downsampled buffer to WAV
	encoder := wav.NewEncoder(outputFile, int(sampleRate), int(decoder.BitDepth), int(decoder.NumChans), 1)
	err = encoder.Write(downsampledBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to encode WAV file: %v", err)
	}
	encoder.Close()

	// Read the encoded data back into a byte slice
	resampledData, err := os.ReadFile(outputFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read temp file: %v", err)
	}

	return resampledData, nil
}

func downsampleBuffer(buffer *audio.IntBuffer, targetRate, originalRate int) *audio.IntBuffer {
	if targetRate >= originalRate {
		return buffer
	}

	ratio := float64(originalRate) / float64(targetRate)
	newLength := int(float64(len(buffer.Data)) / ratio)
	downsampledData := make([]int, newLength)

	for i := 0; i < newLength; i++ {
		downsampledData[i] = buffer.Data[int(float64(i)*ratio)]
	}

	return &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  targetRate,
			NumChannels: buffer.Format.NumChannels,
		},
		Data:           downsampledData,
		SourceBitDepth: buffer.SourceBitDepth,
	}
}

func chunkBytes(data []byte, size int) [][]byte {
	var chunks [][]byte
	for size < len(data) {
		data, chunks = data[size:], append(chunks, data[0:size:size])
	}
	chunks = append(chunks, data)
	return chunks
}

// 16bit mono PCM Only
func VoskServer(endpoint string, fileData []byte, extention string, language string) (string, error) {
	if extention != "wav" {
		return "", fmt.Errorf("only wav files are supported")
	}

	fileData, err := resample(fileData, 16000)
	if err != nil {
		return "", err
	}

	var m struct {
		Text string `json:"text"`
	}
	u := url.URL{Scheme: "ws", Host: endpoint, Path: ""}

	// Opening websocket connection
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return "", err
	}
	defer c.Close()

	err = c.WriteMessage(websocket.TextMessage, []byte("{\"config\" : {\"sample_rate\" : 16000}}"))
	if err != nil {
		return "", err
	}
	for _, chunk := range chunkBytes(fileData, 8000) {
		err = c.WriteMessage(websocket.BinaryMessage, chunk)
		if err != nil {
			return "", err
		}
		_, _, err = c.ReadMessage()
		if err != nil {
			return "", err
		}
	}

	err = c.WriteMessage(websocket.TextMessage, []byte("{\"eof\" : 1}"))
	if err != nil {
		return "", err
	}

	_, msg, err := c.ReadMessage()
	if err != nil {
		return "", err
	}

	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// Unmarshalling received message
	err = json.Unmarshal(msg, &m)
	if err != nil {
		return "", err
	}
	if language == "ja-JP" {
		return strings.ReplaceAll(m.Text, " ", ""), nil
	}
	return m.Text, nil
}
