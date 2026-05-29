package main

import "strings"

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
}

func (c *Conversation) AddAssistant(text string) {
	c.Messages = append(c.Messages, Message{Role: "assistant", Content: text})
}

func (c *Conversation) Clear() {
	c.Messages = nil
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
