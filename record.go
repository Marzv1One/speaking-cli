package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/gen2brain/malgo"
)

type Recording struct {
	WAVData    []byte
	SampleRate uint32
}

var (
	mu          sync.Mutex
	malgoCtx    *malgo.AllocatedContext
	captureDev  *malgo.Device
	capturedPCM []byte
)

// startMicCapture begins recording from the microphone in the background.
// Call stopMicCapture to stop and retrieve the WAV data.
func startMicCapture() error {
	mu.Lock()
	defer mu.Unlock()

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return fmt.Errorf("init context: %w", err)
	}
	malgoCtx = ctx

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000

	capturedPCM = nil

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: func(output, input []byte, framecount uint32) {
			mu.Lock()
			capturedPCM = append(capturedPCM, input...)
			mu.Unlock()
		},
	})
	if err != nil {
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("init device: %w", err)
	}

	captureDev = device

	if err := device.Start(); err != nil {
		device.Uninit()
		ctx.Uninit()
		ctx.Free()
		return fmt.Errorf("start device: %w", err)
	}

	return nil
}

// stopMicCapture stops recording and returns the captured audio as a WAV.
func stopMicCapture() (*Recording, error) {
	mu.Lock()
	defer mu.Unlock()

	if captureDev == nil {
		return nil, fmt.Errorf("not recording")
	}

	captureDev.Uninit()
	captureDev = nil

	if malgoCtx != nil {
		malgoCtx.Uninit()
		malgoCtx.Free()
		malgoCtx = nil
	}

	wav := encodeWAV(capturedPCM, 16000, 1, 16)

	rec := &Recording{
		WAVData:    wav,
		SampleRate: 16000,
	}

	capturedPCM = nil
	return rec, nil
}

// cancelMicCapture stops recording without returning audio data.
func cancelMicCapture() {
	mu.Lock()
	defer mu.Unlock()

	if captureDev != nil {
		captureDev.Uninit()
		captureDev = nil
	}
	if malgoCtx != nil {
		malgoCtx.Uninit()
		malgoCtx.Free()
		malgoCtx = nil
	}
	capturedPCM = nil
}

func encodeWAV(pcm []byte, sampleRate uint32, channels uint16, bitsPerSample uint16) []byte {
	var buf bytes.Buffer

	dataSize := uint32(len(pcm))
	byteRate := sampleRate * uint32(channels) * uint32(bitsPerSample) / 8
	blockAlign := channels * bitsPerSample / 8

	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))
	binary.Write(&buf, binary.LittleEndian, uint16(1))
	binary.Write(&buf, binary.LittleEndian, channels)
	binary.Write(&buf, binary.LittleEndian, sampleRate)
	binary.Write(&buf, binary.LittleEndian, byteRate)
	binary.Write(&buf, binary.LittleEndian, blockAlign)
	binary.Write(&buf, binary.LittleEndian, bitsPerSample)

	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, dataSize)
	buf.Write(pcm)

	return buf.Bytes()
}
