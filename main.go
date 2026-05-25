package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openrouter"
	"github.com/gen2brain/malgo"
)

func main() {
	// --- Record audio from mic ---
	audioData, sampleRate, err := recordFromMic()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Recording failed:", err)
		os.Exit(1)
	}

	// Encode raw PCM as WAV in memory
	wavData := encodeWAV(audioData, sampleRate, 1, 16)

	// --- Send to LLM ---
	provider, err := openrouter.New(openrouter.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Provider error:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	model, err := provider.LanguageModel(ctx, "xiaomi/mimo-v2.5")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Model error:", err)
		os.Exit(1)
	}

	agent := fantasy.NewAgent(model,
		fantasy.WithSystemPrompt("You are a transcription assistant. Output exactly what is said in the audio, word for word. Do not add commentary, interpretation, or formatting. Only output the raw transcribed text."),
	)

	result, err := agent.Generate(ctx, fantasy.AgentCall{
		Prompt: "Transcribe this audio verbatim.",
		Files: []fantasy.FilePart{
			{
				MediaType: "audio/wav",
				Data:      wavData,
			},
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Generation error:", err)
		os.Exit(1)
	}
	fmt.Println(result.Response.Content.Text())
}

func recordFromMic() (pcm []byte, sampleRate uint32, err error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("init context: %w", err)
	}
	defer func() {
		_ = ctx.Uninit()
		ctx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000

	var captured []byte

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: func(output, input []byte, framecount uint32) {
			captured = append(captured, input...)
		},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("init device: %w", err)
	}

	fmt.Println("Press Enter to start recording...")
	fmt.Scanln()

	if err := device.Start(); err != nil {
		return nil, 0, fmt.Errorf("start device: %w", err)
	}

	fmt.Println("Recording... Press Enter to stop.")
	fmt.Scanln()

	device.Uninit()

	return captured, deviceConfig.SampleRate, nil
}

func encodeWAV(pcm []byte, sampleRate uint32, channels uint16, bitsPerSample uint16) []byte {
	var buf bytes.Buffer

	dataSize := uint32(len(pcm))
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	blockAlign := channels * bitsPerSample / 8

	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(&buf, binary.LittleEndian, channels)
	binary.Write(&buf, binary.LittleEndian, sampleRate)
	binary.Write(&buf, binary.LittleEndian, byteRate)
	binary.Write(&buf, binary.LittleEndian, blockAlign)
	binary.Write(&buf, binary.LittleEndian, bitsPerSample)

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, dataSize)
	buf.Write(pcm)

	return buf.Bytes()
}
