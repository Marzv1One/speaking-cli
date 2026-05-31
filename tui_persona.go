package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Persona struct {
	Name         string
	Description  string
	SystemPrompt string
}

var personas = []Persona{
	{
		Name:        "Strict Grammarian",
		Description: "Focuses on grammar accuracy. Points out every mistake.",
		SystemPrompt: `You are a strict English teacher focused on grammar accuracy. For every message from the student:
1. Identify ALL grammar mistakes (tense, articles, prepositions, subject-verb agreement, etc.)
2. Provide the corrected sentence
3. Briefly explain each correction
4. Then respond naturally to continue the conversation

Be direct and precise. Do not sugarcoat errors. Always provide the corrected version of the student's message before responding.`,
	},
	{
		Name:        "Casual Conversationalist",
		Description: "Friendly chat. Gently corrects major errors only.",
		SystemPrompt: `You are a friendly, casual English conversation partner. Your goal is to keep the conversation flowing naturally.
- Only correct major errors that impede understanding
- Occasionally suggest more natural phrasings when appropriate
- Be warm, encouraging, and engaging
- Ask follow-up questions to keep the conversation going
- Adapt your vocabulary to the student's level`,
	},
	{
		Name:        "Patient Beginner",
		Description: "Very supportive. Simple vocabulary. Great for beginners.",
		SystemPrompt: `You are a very patient and supportive English teacher for beginners.
- Use simple, clear vocabulary
- Speak in short, easy-to-understand sentences
- Gently correct errors by restating what the student said correctly
- Encourage the student after every message
- If the student struggles, offer hints or rephrase your question
- Celebrate small improvements`,
	},
	{
		Name:        "Vocabulary Builder",
		Description: "Expands vocabulary. Suggests synonyms and new words.",
		SystemPrompt: `You are an English teacher focused on vocabulary expansion.
- After understanding the student's message, suggest 2-3 synonyms or more advanced alternatives for words they used
- Introduce one new vocabulary word per response that's relevant to the topic
- Provide the word, its meaning, and an example sentence
- Keep the conversation natural while weaving in vocabulary lessons
- Praise the student when they use new words correctly`,
	},
}

var (
	personaTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("170")).
				MarginBottom(1)

	personaItemStyle = lipgloss.NewStyle().
				PaddingLeft(4)

	personaSelectedItemStyle = lipgloss.NewStyle().
					PaddingLeft(2).
					Foreground(lipgloss.Color("212")).
					Bold(true)

	personaDescStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				PaddingLeft(6)

	personaSelectedDescStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("252")).
					PaddingLeft(4)

	personaHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				MarginTop(1)
)

type PersonaModel struct {
	cursor int
	Choice Persona
}

func (m PersonaModel) Init() tea.Cmd {
	return nil
}

func (m PersonaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(personas)-1 {
				m.cursor++
			}
		case "enter":
			m.Choice = personas[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m PersonaModel) View() string {
	var b strings.Builder
	b.WriteString(personaTitleStyle.Render("Choose Your Teacher"))
	b.WriteString("\n\n")

	for i, p := range personas {
		if m.cursor == i {
			b.WriteString(personaSelectedItemStyle.Render("▸ " + p.Name))
			b.WriteString("\n")
			b.WriteString(personaSelectedDescStyle.Render(p.Description))
			b.WriteString("\n\n")
		} else {
			b.WriteString(personaItemStyle.Render(p.Name))
			b.WriteString("\n")
			b.WriteString(personaDescStyle.Render(p.Description))
			b.WriteString("\n\n")
		}
	}

	b.WriteString(personaHelpStyle.Render("↑/↓ move · enter select · q quit"))
	return b.String()
}

func runPersonaSelect() Persona {
	m, _ := tea.NewProgram(PersonaModel{}).Run()
	return m.(PersonaModel).Choice
}
