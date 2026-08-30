package interactions

import (
	"context"
)

// CommandTestHandler handles the /test command as a Chat Input
func CommandTestHandler(_ context.Context, _ InteractionPacket) (InteractionResponse, error) {
	return InteractionResponse{
		Type: ChannelMessageWithSourceInteractionResponseType,
		Data: &InteractionResponseData{
			Content: "test succeeded",
		},
	}, nil
}

// CommandBookHandler handles the /book command as a Chat Input
func CommandBookHandler(_ context.Context, _ InteractionPacket) (InteractionResponse, error) {
	return InteractionResponse{
		Type: ChannelMessageWithSourceInteractionResponseType,
		Data: &InteractionResponseData{
			Content: "book test succeeded",
		},
	}, nil
}
