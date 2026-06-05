package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openRouterEndpoint = "https://openrouter.ai/api/v1/chat/completions"

// openRouterClient calls the OpenRouter chat completions API over plain net/http
// (no SDK, per project constraints). It is safe for concurrent use — all
// per-call state lives in Complete.
type openRouterClient struct {
	apiKey string
	model  string
	hc     *http.Client
}

// NewOpenRouterClient builds a Client backed by OpenRouter. If hc is nil, a
// client with a 30s timeout is used — http.DefaultClient has NO timeout, so a
// hung OpenRouter connection would block the request goroutine forever.
func NewOpenRouterClient(apiKey, model string, hc *http.Client) Client {
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &openRouterClient{apiKey: apiKey, model: model, hc: hc}
}

// --- request / response shapes ---

type openRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *openRouterClient) Complete(ctx context.Context, system string, messages []Message) (string, error) {
	// Prepend the system prompt, then the caller's thread. We don't mutate the
	// caller's slice — build a fresh one sized for system + thread.
	out := make([]Message, 0, len(messages)+1)
	out = append(out, Message{Role: "system", Content: system})
	out = append(out, messages...)

	reqBody := openRouterRequest{
		Model:    c.model,
		Messages: out,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w: %w", err, ErrLLMFailed)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterEndpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("building request: %w: %w", err, ErrLLMFailed)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling openrouter: %w: %w", err, ErrLLMFailed)
	}
	defer resp.Body.Close()

	// Cap the response read so a buggy or hostile upstream can't stream an
	// arbitrarily large body and exhaust memory. A chat-completions reply is a
	// few KB; 1MB is generous headroom.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("reading response: %w: %w", err, ErrLLMFailed)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("openrouter status %d: %s: %w", resp.StatusCode, strings.TrimSpace(string(body)), ErrLLMFailed)
	}

	var or openRouterResponse
	if err := json.Unmarshal(body, &or); err != nil {
		return "", fmt.Errorf("decoding openrouter envelope: %w: %w", err, ErrLLMFailed)
	}
	if len(or.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned no choices: %w", ErrLLMFailed)
	}

	return or.Choices[0].Message.Content, nil
}
