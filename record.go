package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/gen2brain/malgo"
)

type Recording struct {
	WAVData    []byte
	SampleRate uint32
}

func recordFromMic() (*Recording, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, fmt.Errorf("init context: %w", err)
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
		return nil, fmt.Errorf("init device: %w", err)
	}

	fmt.Print("Press Enter to start recording...")
	fmt.Scanln()

	if err := device.Start(); err != nil {
		return nil, fmt.Errorf("start device: %w", err)
	}

	fmt.Println("Recording... Press Enter to stop.")
	fmt.Scanln()

	device.Uninit()

	wav := encodeWAV(captured, deviceConfig.SampleRate, 1, 16)

	return &Recording{
		WAVData:    wav,
		SampleRate: deviceConfig.SampleRate,
	}, nil
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
