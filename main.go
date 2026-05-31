package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/openrouter"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	youStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("114")).
			Bold(true)

	teacherStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	teacherBodyStyle = lipgloss.NewStyle().
				PaddingLeft(2)

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

func main() {
	provider, err := openrouter.New(openrouter.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")))
	if err != nil {
		fmt.Fprintln(os.Stderr, errStyle.Render("Provider error: "+err.Error()))
		os.Exit(1)
	}

	ctx := context.Background()

	transcriptionModel, err := provider.LanguageModel(ctx, "xiaomi/mimo-v2.5")
	if err != nil {
		fmt.Fprintln(os.Stderr, errStyle.Render("Transcription model error: "+err.Error()))
		os.Exit(1)
	}

	teacherModel, err := provider.LanguageModel(ctx, "xiaomi/mimo-v2-flash")
	if err != nil {
		fmt.Fprintln(os.Stderr, errStyle.Render("Teacher model error: "+err.Error()))
		os.Exit(1)
	}

	transcriber := fantasy.NewAgent(transcriptionModel,
		fantasy.WithSystemPrompt("You are a transcription assistant. Output exactly what is said in the audio, word for word. Do not add commentary, interpretation, or formatting. Only output the raw transcribed text."),
	)

	// Initial persona selection
	persona := runPersonaSelect()
	teacher := fantasy.NewAgent(teacherModel,
		fantasy.WithSystemPrompt(persona.SystemPrompt),
	)

	conv := NewConversation()

	fmt.Println()
	fmt.Println(bannerStyle.Render("English Practice Session"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("Teacher: %s", persona.Name)))
	fmt.Println(infoStyle.Render("/style = switch persona · q = quit"))
	fmt.Println()

	for {
		// Run recording + transcription TUI
		recordModel := NewRecordModel(transcriber, ctx)
		m, err := tea.NewProgram(recordModel).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, errStyle.Render("TUI error: "+err.Error()))
			os.Exit(1)
		}

		result := m.(RecordModel)
		text, ok := result.Result()
		if !ok {
			fmt.Println("Goodbye!")
			return
		}

		transcribed := strings.TrimSpace(text)
		if transcribed == "" {
			continue
		}

		// Check for commands
		if strings.HasPrefix(strings.ToLower(transcribed), "/quit") {
			fmt.Println("Goodbye!")
			return
		}
		if strings.HasPrefix(strings.ToLower(transcribed), "/style") {
			persona = runPersonaSelect()
			teacher = fantasy.NewAgent(teacherModel,
				fantasy.WithSystemPrompt(persona.SystemPrompt),
			)
			conv.Clear()
			fmt.Println()
			fmt.Println(infoStyle.Render(fmt.Sprintf("Switched to: %s", persona.Name)))
			fmt.Println()
			continue
		}

		// Print user message
		fmt.Println(youStyle.Render("You: ") + transcribed)

		conv.AddUser(transcribed)

		// Generate teacher response
		fmt.Println(infoStyle.Render("  Thinking..."))
		teacherResult, err := teacher.Generate(ctx, fantasy.AgentCall{
			Messages: conv.FantasyMessages(),
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, errStyle.Render("  Teacher error: "+err.Error()))
			continue
		}

		response := strings.TrimSpace(teacherResult.Response.Content.Text())
		fmt.Println(teacherStyle.Render("Teacher:"))
		fmt.Println(teacherBodyStyle.Render(response))
		fmt.Println()

		conv.AddAssistant(response)
	}
}
