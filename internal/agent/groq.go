package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GroqClient struct {
	APIKey string
	Model  string
}

func NewGroqClient(apiKey, model string) *GroqClient {
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &GroqClient{APIKey: apiKey, Model: model}
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqRequest struct {
	Model          string              `json:"model"`
	Messages       []groqMessage       `json:"messages"`
	Temperature    float64             `json:"temperature"`
	ResponseFormat *groqResponseFormat `json:"response_format,omitempty"`
}

type groqResponseFormat struct {
	Type string `json:"type"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GenerateAction sends the command to Groq and processes the JSON response
func (c *GroqClient) GenerateAction(ctx context.Context, userCommand string) (*AgentAction, error) {
	reqBody := groqRequest{
		Model: c.Model,
		Messages: []groqMessage{
			{Role: "system", Content: SystemPrompt},
			{Role: "user", Content: userCommand},
		},
		Temperature:    0.2,
		ResponseFormat: &groqResponseFormat{Type: "json_object"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling Groq API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return nil, fmt.Errorf("error decoding Groq response: %v", err)
	}

	if groqResp.Error != nil {
		return nil, fmt.Errorf("Groq API error: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return nil, fmt.Errorf("Groq returned no choices")
	}

	content := groqResp.Choices[0].Message.Content
	var action AgentAction
	if err := json.Unmarshal([]byte(content), &action); err != nil {
		return nil, fmt.Errorf("error parsing AI JSON response: %v\nRaw response: %s", err, content)
	}

	return &action, nil
}
