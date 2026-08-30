package book

import (
	"context"
	"github.com/brazostech/discord/interactions"
)

// CommandBookHandler handles the /book command as a Chat Input
func CommandBookHandler(_ context.Context, _ interactions.InteractionPacket) (interactions.InteractionResponse, error) {
	return interactions.InteractionResponse{
		Type: interactions.ChannelMessageWithSourceInteractionResponseType,
		Data: &interactions.InteractionResponseData{
			Content: "book test succeeded",
		},
	}, nil
}
