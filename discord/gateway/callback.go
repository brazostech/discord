package gateway

import (
	"context"
	"fmt"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

// Callback sends interaction responses to Discord. Requests carry the bot
// authorization header and address the interaction by id and token, as the
// interaction callback endpoint requires.
type Callback struct {
	client *discord.APIClient
}

func NewCallback(client *discord.APIClient) *Callback {
	return &Callback{client: client}
}

// Respond posts the response to /interactions/{id}/{token}/callback.
func (c *Callback) Respond(ctx context.Context, interaction interactions.Interaction, response interactions.InteractionResponse) error {
	endpoint := fmt.Sprintf("interactions/%s/%s/callback", interaction.ID, interaction.Token)

	if err := c.client.Execute(ctx, endpoint, discord.RequestOptions{
		Method: discord.MethodPost,
		Body:   response,
	}); err != nil {
		return fmt.Errorf("interaction callback: %w", err)
	}

	return nil
}
