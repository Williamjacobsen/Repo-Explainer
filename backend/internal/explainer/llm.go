package explainer
// Smallest "send a prompt → get a reply" using the Ollama Go client.
// Model: qwen2.5-coder:0.5b-instruct
//
// Run from repo root:
//   go run ./cmd/llmtest "Explain goroutines in one sentence."
//
// If OLLAMA_HOST isn't set, it defaults to http://localhost:11434.

import (
	"bytes"
	"context"
	"fmt"

	api "github.com/ollama/ollama/api"
)

func Llm(prompt string) {
	ctx := context.Background()

	// Create a client that talks to OLLAMA_HOST (or localhost:11434 by default).
	client, err := api.ClientFromEnvironment()
	if err != nil {
		panic(err)
	}

	// Minimal chat request: one user message to the tiny model.
	req := &api.ChatRequest{
		Model: "qwen2.5-coder:0.5b-instruct",
		Messages: []api.Message{
			{Role: "user", Content: prompt},
		},
	}

	// Stream response chunks into a buffer, then print.
	var out bytes.Buffer
	if err := client.Chat(ctx, req, func(cr api.ChatResponse) error {
		out.WriteString(cr.Message.Content)
		return nil
	}); err != nil {
		panic(err)
	}

	fmt.Println(out.String())
}
