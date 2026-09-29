# speaking-cli

A terminal-based English speaking-practice app. Talk into your microphone, your speech is transcribed by an AI model, and an AI "teacher" responds with feedback and corrections — all inside a Bubble Tea TUI.

## Features

- **Voice-driven practice**: record speech from your microphone (16 kHz, mono, 16-bit WAV) and get it transcribed automatically
- **AI teacher personas** with different teaching styles:
  - **Strict Grammarian** — flags every mistake with corrected sentences and explanations
  - **Casual Conversationalist** — friendly chat, only major corrections
  - **Patient Beginner** — simple vocabulary, very supportive
  - **Vocabulary Builder** — suggests synonyms and introduces new words
- **Multi-segment messages** — append several recordings into one message before sending
- **Transcript editing** — fix the transcription before it's sent to the teacher
- **Conversation memory** — the teacher remembers the last 20 messages; history resets when you switch personas
- **Mid-session persona switching** via the `/style` command

## Requirements

- [Go](https://go.dev/) 1.26+
- A C compiler toolchain (the audio capture uses cgo via [malgo](https://github.com/gen2brain/malgo))
- A working microphone
- An [OpenRouter](https://openrouter.ai/) API key

## Setup

Set your OpenRouter API key:

```sh
export OPENROUTER_API_KEY=sk-or-...
```

(On Windows: `setx OPENROUTER_API_KEY "sk-or-..."` or set it per-session in PowerShell with `$env:OPENROUTER_API_KEY = "sk-or-..."`.)

## Build & Run

```sh
go build -o speaking .
./speaking
```

## Usage

1. Pick a teacher persona with ↑/↓ (or `j`/`k`) and press Enter.
2. Press Enter to start recording, Enter again to stop.
3. In the review screen:
   - `Enter` — send the transcription to the teacher
   - `a` — record another segment and append it to the same message
   - `e` — edit the transcript text before sending
   - `r` — re-record the segment
   - `c` — cancel while recording
4. Chat with the teacher. Commands:
   - `/style` — switch persona mid-session (clears conversation history)
   - `/quit` or `q` — exit

## Models

Models are configured in `main.go` via [OpenRouter](https://openrouter.ai/):

- **Transcription**: `mistralai/voxtral-small-24b-2507`
- **Teacher**: `openai/gpt-oss-120b`

## Project Structure

| File | Purpose |
|------|---------|
| `main.go` | Entry point: OpenRouter setup, models, main chat loop, `/style` and `/quit` commands |
| `tui_record.go` | Record/transcribe/review/edit Bubble Tea model |
| `tui_persona.go` | Persona picker and the four persona definitions |
| `record.go` | Microphone capture and WAV encoding |
| `conversation.go` | Message history (20-message cap) |
