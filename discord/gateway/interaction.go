package gateway

import (
	"context"
	"encoding/json"
	"log"

	"github.com/brazostech/discord/discord/interactions"
)

func (c *Client) handleInteraction(ctx context.Context, data json.RawMessage) {
	var interaction interactions.Interaction
	if err := json.Unmarshal(data, &interaction); err != nil {
		log.Printf("discord gateway: decode interaction: %v", err)
		return
	}

	response, err := c.dispatcher.Handle(ctx, interaction)
	if err != nil {
		log.Printf("discord gateway: handle interaction %s: %v", interaction.ID, err)
		return
	}

	if err := c.responder.Respond(ctx, interaction, response); err != nil {
		log.Printf("discord gateway: respond to interaction %s: %v", interaction.ID, err)
	}
}
