package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/brazostech/discord/discord"
	"github.com/brazostech/discord/discord/interactions"
)

const maxErrorBodySize = 1 << 12

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

	resp, err := c.client.Request(ctx, endpoint, discord.RequestOptions{
		Method: discord.MethodPost,
		Body:   response,
	})
	if err != nil {
		return fmt.Errorf("request interaction callback: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return fmt.Errorf("interaction callback: discord returned %s: %s", resp.Status, body)
	}

	return nil
}
