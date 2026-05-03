package agent

import "context"

// LLMClient is the interface that all AI providers must implement
type LLMClient interface {
	GenerateAction(ctx context.Context, userCommand string) (*AgentAction, error)
}
