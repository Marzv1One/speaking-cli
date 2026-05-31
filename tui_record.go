package main

import (
	"context"

	"charm.land/fantasy"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type recordState int

const (
	recordIdle recordState = iota
	recordActive
	recordTranscribing
	recordReview
)

var (
	recordStatusStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("212"))

	recordTextStyle = lipgloss.NewStyle().
			PaddingLeft(4).
			Foreground(lipgloss.Color("252"))

	recordHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	recordErrorStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("196"))
)

type recordDoneMsg struct{ data []byte }

type transcriptionDoneMsg struct {
	text string
	err  error
}

type RecordModel struct {
	state       recordState
	audio       []byte
	transcribed string
	spinner     spinner.Model
	transcriber fantasy.Agent
	ctx         context.Context
	stopCh      chan struct{}
	width       int
}

func NewRecordModel(transcriber fantasy.Agent, ctx context.Context) RecordModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	return RecordModel{
		state:       recordIdle,
		spinner:     sp,
		transcriber: transcriber,
		ctx:         ctx,
		width:       80,
	}
}

func (m RecordModel) Result() (string, bool) {
	if m.state == recordReview {
		return m.transcribed, true
	}
	return "", false
}

func (m RecordModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m RecordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case recordIdle:
			return m.handleIdleKeys(msg)
		case recordActive:
			return m.handleActiveKeys(msg)
		case recordTranscribing:
			return m.handleTranscribingKeys(msg)
		case recordReview:
			return m.handleReviewKeys(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case recordDoneMsg:
		m.audio = msg.data
		m.state = recordTranscribing
		return m, m.transcribeAudio()

	case transcriptionDoneMsg:
		if msg.err != nil {
			m.transcribed = ""
			m.state = recordIdle
			return m, nil
		}
		m.transcribed = msg.text
		m.state = recordReview
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m RecordModel) handleIdleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		m.state = recordActive
		m.stopCh = make(chan struct{})
		return m, m.startRecording()
	}
	return m, nil
}

func (m RecordModel) handleActiveKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		cancelMicCapture()
		return m, tea.Quit
	case "enter":
		rec, err := stopMicCapture()
		if err != nil {
			return m, nil
		}
		m.audio = rec.WAVData
		m.state = recordTranscribing
		return m, m.transcribeAudio()
	case "c":
		cancelMicCapture()
		m.state = recordIdle
		return m, nil
	}
	return m, nil
}

func (m RecordModel) handleTranscribingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}

func (m RecordModel) handleReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		return m, tea.Quit
	case "r":
		m.state = recordIdle
		m.audio = nil
		m.transcribed = ""
		return m, nil
	}
	return m, nil
}

func (m RecordModel) startRecording() tea.Cmd {
	return func() tea.Msg {
		if err := startMicCapture(); err != nil {
			return nil
		}
		return nil
	}
}

func (m RecordModel) transcribeAudio() tea.Cmd {
	return func() tea.Msg {
		result, err := m.transcriber.Generate(m.ctx, fantasy.AgentCall{
			Prompt: "Transcribe this audio verbatim.",
			Files: []fantasy.FilePart{
				{MediaType: "audio/wav", Data: m.audio},
			},
		})
		if err != nil {
			return transcriptionDoneMsg{err: err}
		}
		return transcriptionDoneMsg{text: result.Response.Content.Text()}
	}
}

func (m RecordModel) View() string {
	switch m.state {
	case recordIdle:
		return recordStatusStyle.Render("● Ready") + "\n" +
			recordHelpStyle.Render("enter record · q quit")

	case recordActive:
		return recordStatusStyle.Render("● Recording...") + "\n" +
			recordHelpStyle.Render("enter stop · c cancel")

	case recordTranscribing:
		return recordStatusStyle.Render("● Transcribing ") + m.spinner.View() + "\n" +
			recordHelpStyle.Render("ctrl+c quit")

	case recordReview:
		textStyle := recordTextStyle.Width(m.width)
		return recordStatusStyle.Render("● You said:") + "\n" +
			textStyle.Render(m.transcribed) + "\n" +
			recordHelpStyle.Render("enter send · r re-record · q quit")

	default:
		return ""
	}
}
