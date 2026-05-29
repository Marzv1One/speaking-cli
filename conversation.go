package main

import (
	"strings"

	"charm.land/fantasy"
)

const maxHistoryMessages = 20

type Message struct {
	Role    string
	Content string
}

type Conversation struct {
	Messages []Message
}

func NewConversation() *Conversation {
	return &Conversation{}
}

func (c *Conversation) AddUser(text string) {
	c.Messages = append(c.Messages, Message{Role: "user", Content: text})
	c.trim()
}

func (c *Conversation) AddAssistant(text string) {
	c.Messages = append(c.Messages, Message{Role: "assistant", Content: text})
	c.trim()
}

func (c *Conversation) Clear() {
	c.Messages = nil
}

func (c *Conversation) trim() {
	if len(c.Messages) > maxHistoryMessages {
		c.Messages = c.Messages[len(c.Messages)-maxHistoryMessages:]
	}
}

// FantasyMessages converts the conversation history into fantasy.Message values
// suitable for passing to fantasy.AgentCall.Messages.
func (c *Conversation) FantasyMessages() []fantasy.Message {
	msgs := make([]fantasy.Message, 0, len(c.Messages))
	for _, m := range c.Messages {
		switch m.Role {
		case "user":
			msgs = append(msgs, fantasy.NewUserMessage(m.Content))
		case "assistant":
			msgs = append(msgs, fantasy.Message{
				Role:    fantasy.MessageRoleAssistant,
				Content: []fantasy.MessagePart{fantasy.TextPart{Text: m.Content}},
			})
		}
	}
	return msgs
}

func (c *Conversation) FormatHistory() string {
	var b strings.Builder
	for _, m := range c.Messages {
		switch m.Role {
		case "user":
			b.WriteString("Student: ")
		case "assistant":
			b.WriteString("Teacher: ")
		}
		b.WriteString(m.Content)
		b.WriteString("\n")
	}
	return b.String()
}
