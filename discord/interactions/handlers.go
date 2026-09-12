package interactions

import (
	"context"
)

// CommandTestHandler handles the /test command as a Chat Input
func CommandTestHandler(_ context.Context, _ Invocation) (InteractionResponse, error) {
	return InteractionResponse{
		Type: ChannelMessageWithSourceInteractionResponseType,
		Data: &InteractionResponseData{
			Content: "test succeeded",
		},
	}, nil
}
