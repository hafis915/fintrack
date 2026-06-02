// Package ai is a thin wrapper around an OpenAI-compatible chat completions
// endpoint. We route through OpenRouter (see config.AIBaseURL); the same code
// also targets OpenAI / Together / Groq / any provider speaking that format.
package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	apiKey  string
	baseURL string
	model   string
	hc      *http.Client
}

func New(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		hc:      &http.Client{Timeout: 60 * time.Second},
	}
}

// Block is a piece of message content. For text-only messages set Text;
// for vision pass an image via NewImageBlock.
type Block struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

func NewTextBlock(text string) Block {
	return Block{Type: "text", Text: text}
}

func NewImageBlock(mimeType string, image []byte) Block {
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(image))
	return Block{Type: "image_url", ImageURL: &ImageURL{URL: dataURL}}
}

type Message struct {
	Role    string  `json:"role"`
	Content []Block `json:"content"`
}

type CompleteOptions struct {
	System      string
	Messages    []Message
	MaxTokens   int
	Temperature float64
	JSONOnly    bool
}

type completionRequest struct {
	Model          string         `json:"model"`
	Messages       []chatMessage  `json:"messages"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	Temperature    float64        `json:"temperature"`
	ResponseFormat *responseFmt   `json:"response_format,omitempty"`
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content any         `json:"content"` // either string or []Block depending on shape
}

type responseFmt struct {
	Type string `json:"type"`
}

type completionResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *apiError `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Complete sends the messages to the upstream and returns the raw assistant text.
// One retry on 5xx; auth failures (401/403) and 4xx are returned immediately.
func (c *Client) Complete(ctx context.Context, opts CompleteOptions) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("ai: empty AI_API_KEY")
	}

	msgs := make([]chatMessage, 0, len(opts.Messages)+1)
	if opts.System != "" {
		msgs = append(msgs, chatMessage{Role: "system", Content: opts.System})
	}
	for _, m := range opts.Messages {
		// Plain text shortcut keeps the wire format lean for non-vision calls.
		if len(m.Content) == 1 && m.Content[0].Type == "text" {
			msgs = append(msgs, chatMessage{Role: m.Role, Content: m.Content[0].Text})
		} else {
			msgs = append(msgs, chatMessage{Role: m.Role, Content: m.Content})
		}
	}

	body := completionRequest{
		Model:       c.model,
		Messages:    msgs,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	}
	if opts.JSONOnly {
		body.ResponseFormat = &responseFmt{Type: "json_object"}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("ai: marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		text, retriable, err := c.do(ctx, url, raw)
		if err == nil {
			return text, nil
		}
		lastErr = err
		if !retriable {
			return "", err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return "", lastErr
}

func (c *Client) do(ctx context.Context, url string, body []byte) (string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", false, fmt.Errorf("ai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	// OpenRouter recommends these for attribution/visibility
	req.Header.Set("HTTP-Referer", "https://github.com/hafis915/fintrack")
	req.Header.Set("X-Title", "Fintrack")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", true, fmt.Errorf("ai: http: %w", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		return "", true, fmt.Errorf("ai: upstream %d: %s", resp.StatusCode, string(rb))
	}
	if resp.StatusCode >= 400 {
		return "", false, fmt.Errorf("ai: upstream %d: %s", resp.StatusCode, string(rb))
	}

	var out completionResponse
	if err := json.Unmarshal(rb, &out); err != nil {
		return "", false, fmt.Errorf("ai: decode response: %w (body=%s)", err, truncate(string(rb), 200))
	}
	if out.Error != nil {
		return "", false, fmt.Errorf("ai: %s (%s)", out.Error.Message, out.Error.Code)
	}
	if len(out.Choices) == 0 {
		return "", false, errors.New("ai: empty choices")
	}
	return out.Choices[0].Message.Content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
