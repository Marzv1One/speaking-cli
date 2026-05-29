package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openrouter"
)

func main() {
	provider, err := openrouter.New(openrouter.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Provider error:", err)
		os.Exit(1)
	}

	ctx := context.Background()

	transcriptionModel, err := provider.LanguageModel(ctx, "xiaomi/mimo-v2.5")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Transcription model error:", err)
		os.Exit(1)
	}

	teacherModel, err := provider.LanguageModel(ctx, "xiaomi/mimo-v2-flash")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Teacher model error:", err)
		os.Exit(1)
	}

	transcriber := fantasy.NewAgent(transcriptionModel,
		fantasy.WithSystemPrompt("You are a transcription assistant. Output exactly what is said in the audio, word for word. Do not add commentary, interpretation, or formatting. Only output the raw transcribed text."),
	)

	persona := selectPersona()
	teacher := fantasy.NewAgent(teacherModel,
		fantasy.WithSystemPrompt(persona.SystemPrompt),
	)

	conv := NewConversation()

	fmt.Println("=== English Practice Session ===")
	fmt.Println("Commands: /style = switch persona, /quit = exit\n")

	for {
		rec, err := recordFromMic()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Recording error:", err)
			continue
		}

		// Transcribe
		fmt.Print("Transcribing...")
		result, err := transcriber.Generate(ctx, fantasy.AgentCall{
			Prompt: "Transcribe this audio verbatim.",
			Files: []fantasy.FilePart{
				{MediaType: "audio/wav", Data: rec.WAVData},
			},
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nTranscription error:", err)
			continue
		}

		transcribed := strings.TrimSpace(result.Response.Content.Text())
		fmt.Printf("\rYou said: %s\n", transcribed)

		// Check for commands
		if strings.HasPrefix(strings.ToLower(transcribed), "/quit") {
			fmt.Println("Goodbye!")
			return
		}
		if strings.HasPrefix(strings.ToLower(transcribed), "/style") {
			persona = selectPersona()
			teacher = fantasy.NewAgent(teacherModel,
				fantasy.WithSystemPrompt(persona.SystemPrompt),
			)
			conv.Clear()
			continue
		}

		conv.AddUser(transcribed)

		// Build teacher prompt with conversation history
		teacherPrompt := conv.FormatHistory() + "\n\nRespond as the English teacher."

		fmt.Print("Thinking...")
		teacherResult, err := teacher.Generate(ctx, fantasy.AgentCall{
			Prompt: teacherPrompt,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "\nTeacher error:", err)
			continue
		}

		response := strings.TrimSpace(teacherResult.Response.Content.Text())
		fmt.Printf("\rTeacher: %s\n\n", response)

		conv.AddAssistant(response)
	}
}
