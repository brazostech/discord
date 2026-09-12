package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

func TestCallbackPostsInteractionResponse(t *testing.T) {
	type capturedRequest struct {
		method string
		path   string
		auth   string
		body   []byte
	}

	requests := make(chan capturedRequest, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		requests <- capturedRequest{
			method: r.Method,
			path:   r.URL.Path,
			auth:   r.Header.Get("Authorization"),
			body:   body,
		}
	}))
	defer server.Close()

	callback := NewCallback(discord.NewAPIClient("test-token", discord.WithBaseURL(server.URL)))

	err := callback.Respond(t.Context(),
		interactions.Interaction{ID: "42", Token: "interaction-token"},
		interactions.InteractionResponse{
			Type: interactions.ChannelMessageWithSourceInteractionResponseType,
			Data: &interactions.InteractionResponseData{Content: "hello"},
		},
	)
	if err != nil {
		t.Fatalf("respond: %v", err)
	}

	got := <-requests

	if got.method != http.MethodPost {
		t.Fatalf("method = %s, want POST", got.method)
	}
	if got.path != "/interactions/42/interaction-token/callback" {
		t.Fatalf("path = %s", got.path)
	}
	if got.auth != "Bot test-token" {
		t.Fatalf("authorization = %q, want the bot token", got.auth)
	}

	var payload interactions.InteractionResponse
	if err := json.Unmarshal(got.body, &payload); err != nil {
		t.Fatalf("decode callback body %q: %v", got.body, err)
	}
	if payload.Type != interactions.ChannelMessageWithSourceInteractionResponseType {
		t.Fatalf("body type = %d", payload.Type)
	}
	if payload.Data == nil || payload.Data.Content != "hello" {
		t.Fatalf("body data = %+v", payload.Data)
	}
}

func TestCallbackReportsDiscordError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
	}))
	defer server.Close()

	callback := NewCallback(discord.NewAPIClient("test-token", discord.WithBaseURL(server.URL)))

	err := callback.Respond(t.Context(),
		interactions.Interaction{ID: "42", Token: "interaction-token"},
		interactions.InteractionResponse{Type: interactions.PongInteractionResponseType},
	)
	if err == nil {
		t.Fatal("expected an error for a non-2xx callback response")
	}
}
