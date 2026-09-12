package book

import (
	"context"
	"github.com/brazostech/discord/discord/interactions"
	"github.com/brazostech/discord/ports"
)

type BookService struct {
	repository ports.Repository[Book, string]
}

func NewBookService(repository ports.Repository[Book, string]) *BookService {
	bs := BookService{
		repository: repository,
	}

	return &bs
}

// CommandBookHandler handles the /book command as a Chat Input
// TODO:
//   - use context
//   - test the thing
func (bs *BookService) CommandBookHandler(ctx context.Context, _ interactions.InteractionPacket) (interactions.InteractionResponse, error) {
	select {
	case <-ctx.Done():
		return interactions.InteractionResponse{
			Type: interactions.ChannelMessageWithSourceInteractionResponseType,
			Data: &interactions.InteractionResponseData{
				Content: "book test failed",
			},
		}, ctx.Err()
	default:
	}
	return interactions.InteractionResponse{
		Type: interactions.ChannelMessageWithSourceInteractionResponseType,
		Data: &interactions.InteractionResponseData{
			Content: "book test succeeded",
		},
	}, nil
}
