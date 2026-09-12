package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/brazostech/discord/discord"
)

func TestDispatcherResolvesSubcommandPathAndParsesOptions(t *testing.T) {
	dispatcher := NewDispatcher()

	var got []discord.InteractionOption
	dispatcher.Subscribe("book register", func(_ context.Context, invocation Invocation) (InteractionResponse, error) {
		got = invocation.Options
		return InteractionResponse{
			Type: ChannelMessageWithSourceInteractionResponseType,
			Data: &InteractionResponseData{Content: "handled"},
		}, nil
	})

	raw := json.RawMessage(`{
		"id": "1",
		"name": "book",
		"type": 1,
		"options": [
			{
				"name": "register",
				"type": 1,
				"options": [
					{"name": "name", "type": 3, "value": "Dune"}
				]
			}
		]
	}`)

	res, err := dispatcher.Handle(t.Context(), Interaction{
		Type:    ApplicationCommandInteractionType,
		GuildID: "server-1",
		Data:    raw,
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(got) != 1 || got[0].Name != "name" {
		t.Fatalf("handler options = %+v, want the name option", got)
	}
	name, ok := got[0].StringValue()
	if !ok || name != "Dune" {
		t.Fatalf("option value = %q, %v; want Dune", name, ok)
	}

	if res.Type != ChannelMessageWithSourceInteractionResponseType {
		t.Fatalf("response type = %d", res.Type)
	}
	if res.Data == nil || res.Data.Content != "handled" {
		t.Fatalf("response data = %+v", res.Data)
	}
}

func TestDispatcherAnswersPing(t *testing.T) {
	res, err := NewDispatcher().Handle(t.Context(), Interaction{Type: PingInteractionType})
	if err != nil {
		t.Fatalf("handle ping: %v", err)
	}
	if res.Type != PongInteractionResponseType {
		t.Fatalf("response type = %d, want pong", res.Type)
	}
}

func TestDispatcherUnknownCommand(t *testing.T) {
	dispatcher := NewDispatcher()
	dispatcher.Subscribe("book register", func(context.Context, Invocation) (InteractionResponse, error) {
		return InteractionResponse{}, nil
	})

	raw := json.RawMessage(`{"id":"1","name":"book","type":1,"options":[{"name":"update-chapter","type":1}]}`)

	_, err := dispatcher.Handle(t.Context(), Interaction{
		Type: ApplicationCommandInteractionType,
		Data: raw,
	})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrHandlerNotFound)
	}
}

func TestDispatcherUnsupportedInteractionType(t *testing.T) {
	_, err := NewDispatcher().Handle(t.Context(), Interaction{Type: MessageComponentInteractionType})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrHandlerNotFound)
	}
}
