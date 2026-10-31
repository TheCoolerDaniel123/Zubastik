package ai

import (
	"context"
	"log"

	"shared/entity"
)

// StubAIProvider is a placeholder that logs the call and returns a canned response.
// Replace this with a real provider (e.g. OpenAI, Anthropic) by implementing the
// usecase.AIProvider interface and swapping it in service.go.
type StubAIProvider struct{}

func NewStubAIProvider() *StubAIProvider {
	return &StubAIProvider{}
}

func (p *StubAIProvider) Complete(ctx context.Context, systemPrompt string, history []entity.ChatMessage, userMessage string) (string, error) {
	log.Printf("[AI STUB] history_len=%d user_message=%q", len(history), userMessage)
	return "AI provider not connected yet. Please check back later.", nil
}
