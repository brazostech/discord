package interactions

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/brazostech/discord/discord"
)

func TestCommandRouterDispatchesBySubcommandPath(t *testing.T) {
	router := NewCommandRouter()

	var got []discord.InteractionOption
	router.Subscribe("book register", func(_ context.Context, packet InteractionPacket) (InteractionResponse, error) {
		got = packet.Data.SubcommandOptions()
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

	packet := InteractionPacket{
		Interaction: Interaction{
			Type:    ApplicationCommandInteractionType,
			GuildID: "server-1",
			Data:    raw,
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/interactions", nil)

	if err := router.Route(t.Context(), recorder, request, packet); err != nil {
		t.Fatalf("route: %v", err)
	}

	if len(got) != 1 || got[0].Name != "name" {
		t.Fatalf("handler options = %+v, want the name option", got)
	}
	name, ok := got[0].StringValue()
	if !ok || name != "Dune" {
		t.Fatalf("option value = %q, %v; want Dune", name, ok)
	}

	if !strings.Contains(recorder.Body.String(), "handled") {
		t.Fatalf("response body = %q", recorder.Body.String())
	}
}

func TestCommandRouterUnknownPath(t *testing.T) {
	router := NewCommandRouter()
	router.Subscribe("book register", func(context.Context, InteractionPacket) (InteractionResponse, error) {
		return InteractionResponse{}, nil
	})

	raw := json.RawMessage(`{"id":"1","name":"book","type":1,"options":[{"name":"update-chapter","type":1}]}`)

	packet := InteractionPacket{
		Interaction: Interaction{
			Type: ApplicationCommandInteractionType,
			Data: raw,
		},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/interactions", nil)

	err := router.Route(t.Context(), recorder, request, packet)
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("error = %v, want %v", err, ErrHandlerNotFound)
	}
}
