package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AgentAction represents the JSON structure we expect from the AI
type AgentAction struct {
	Thought string `json:"thought"`
	Code    string `json:"code"`
	IsFinal bool   `json:"is_final"`
}

type OllamaClient struct {
	BaseURL string
	Model   string
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{BaseURL: baseURL, Model: model}
}

type ollamaRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Format  string         `json:"format"`
	Options *ollamaOptions `json:"options,omitempty"`
}

type ollamaOptions struct {
	NumPredict int `json:"num_predict"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// GenerateAction sends the command to Ollama and processes the JSON response
func (c *OllamaClient) GenerateAction(ctx context.Context, userCommand string) (*AgentAction, error) {
	prompt := fmt.Sprintf("%s\n\nUser command: %s", SystemPrompt, userCommand)

	reqBody := ollamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
		Options: &ollamaOptions{
			NumPredict: 8192,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Ollama API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error (status %d): %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("error decoding Ollama response: %v", err)
	}

	var action AgentAction
	if err := json.Unmarshal([]byte(ollamaResp.Response), &action); err != nil {
		return nil, fmt.Errorf("error parsing AI JSON response: %v\nRaw response: %s", err, ollamaResp.Response)
	}

	return &action, nil
}
