// Package llm provides the language layer for the financial planner chat. The
// package exposes a tiny Client interface so callers (the planner handler)
// depend on behaviour, not on the HTTP transport. The LLM is ONLY a language
// layer: it interprets the user's intent and narrates trade-offs. It never
// invents budget numbers — the actual money math is re-balanced deterministically
// by app code. The system prompt the caller passes instructs the model to emit
// STRICT JSON, and Complete returns that raw assistant text unparsed.
//
// A deterministic stub implementation (NewStubClient) is provided for tests and
// for local dev without an OpenRouter API key. It performs regex NLU on the last
// user message and never touches the network.
package llm

import (
	"context"
	"errors"
)

// Message is a single turn in a chat thread. Role is "system", "user", or
// "assistant". The frontend holds the thread and sends it on every turn — the
// Client itself is stateless.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client turns a system prompt + a message thread into raw assistant text.
// Implementations must respect ctx cancellation and return a wrapped
// ErrLLMFailed on any failure so callers can branch with errors.Is. The returned
// string is the assistant's raw content — parsing (the caller's system prompt
// asks for STRICT JSON) is the caller's responsibility.
type Client interface {
	Complete(ctx context.Context, system string, messages []Message) (string, error)
}

// ErrLLMFailed is the sentinel wrapped by every Client failure path (network
// error, non-2xx response, unparseable envelope, empty choices, etc.). Callers
// use errors.Is(err, llm.ErrLLMFailed) rather than matching on transport detail.
var ErrLLMFailed = errors.New("llm completion failed")
