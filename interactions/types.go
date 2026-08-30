package interactions

import (
	"encoding/json"
)

type InteractionType int

const (
	PingInteractionType InteractionType = iota + 1
	ApplicationCommandInteractionType
	MessageComponentInteractionType
	ApplicationCommandAutocompleteInteractionType
	ModalSubmitInteractionType
)

type Interaction struct {
	ID             string          `json:"id"`
	ApplicationID  string          `json:"application_id"`
	Type           InteractionType `json:"type"`
	GuildID        string          `json:"guild_id,omitempty"`
	ChannelID      string          `json:"channel_id,omitempty"`
	Version        int             `json:"version"`
	Data           json.RawMessage `json:"data,omitempty"`
	AppPermissions string          `json:"app_permissions,omitempty"`
	Token          string          `json:"token"`
}

type InteractionResponseType int

const (
	PongInteractionResponseType                                 InteractionResponseType = 1
	ChannelMessageWithSourceInteractionResponseType             InteractionResponseType = 4
	DeferredChannelMessageWithSourceInteractionResponseType     InteractionResponseType = 5
	DeferredUpdateMessageInteractionResponseType                InteractionResponseType = 6
	UpdateMessageInteractionResponseType                        InteractionResponseType = 7
	ApplicationCommandAutocompleteResultInteractionResponseType InteractionResponseType = 8
	ModalInteractionResponseType                                InteractionResponseType = 9
	PremiumRequiredInteractionResponseType                      InteractionResponseType = 10
	LaunchActivityInteractionResponseType                       InteractionResponseType = 12
)

type InteractionResponse struct {
	Type InteractionResponseType  `json:"type"`
	Data *InteractionResponseData `json:"data,omitempty"`
}

type InteractionResponseData struct {
	Content string `json:"content,omitempty"`
	Flags   int    `json:"flags,omitempty"`
}
