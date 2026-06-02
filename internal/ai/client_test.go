package ai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hafis915/fintrack/internal/ai"
)

func TestComplete_TextOnly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chat/completions", r.URL.Path)
		require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		require.NoError(t, json.Unmarshal(body, &got))
		require.Equal(t, "anthropic/claude-haiku-4.5", got["model"])
		// Text-only messages should send content as plain strings, not arrays
		msgs := got["messages"].([]any)
		first := msgs[0].(map[string]any)
		_, isString := first["content"].(string)
		require.True(t, isString, "text-only content should be a string, got %T", first["content"])
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"hello"}}]}`))
	}))
	defer srv.Close()

	c := ai.New("test-key", srv.URL, "anthropic/claude-haiku-4.5")
	out, err := c.Complete(context.Background(), ai.CompleteOptions{
		System: "be brief",
		Messages: []ai.Message{
			{Role: "user", Content: []ai.Block{ai.NewTextBlock("hi")}},
		},
		MaxTokens: 50,
	})
	require.NoError(t, err)
	require.Equal(t, "hello", out)
}

func TestComplete_VisionContentIsArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Vision content should serialize as an array of blocks
		require.True(t, strings.Contains(string(body), `"type":"image_url"`))
		require.True(t, strings.Contains(string(body), `"data:image/jpeg;base64,`))
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"amount\":50000}"}}]}`))
	}))
	defer srv.Close()

	c := ai.New("k", srv.URL, "anthropic/claude-haiku-4.5")
	out, err := c.Complete(context.Background(), ai.CompleteOptions{
		Messages: []ai.Message{
			{Role: "user", Content: []ai.Block{
				ai.NewTextBlock("read this struk"),
				ai.NewImageBlock("image/jpeg", []byte{0xff, 0xd8, 0xff}),
			}},
		},
		JSONOnly: true,
	})
	require.NoError(t, err)
	require.Equal(t, `{"amount":50000}`, out)
}

func TestComplete_4xxIsNotRetried(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		http.Error(w, `{"error":{"message":"bad key","code":"invalid_api_key"}}`, 401)
	}))
	defer srv.Close()

	c := ai.New("k", srv.URL, "m")
	_, err := c.Complete(context.Background(), ai.CompleteOptions{Messages: []ai.Message{{Role: "user", Content: []ai.Block{ai.NewTextBlock("x")}}}})
	require.Error(t, err)
	require.Equal(t, 1, calls, "401 should not retry")
}

func TestComplete_5xxRetriesOnce(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			http.Error(w, "upstream burp", 502)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"recovered"}}]}`))
	}))
	defer srv.Close()

	c := ai.New("k", srv.URL, "m")
	out, err := c.Complete(context.Background(), ai.CompleteOptions{Messages: []ai.Message{{Role: "user", Content: []ai.Block{ai.NewTextBlock("x")}}}})
	require.NoError(t, err)
	require.Equal(t, "recovered", out)
	require.Equal(t, 2, calls)
}
