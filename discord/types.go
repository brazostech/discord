package discord

import "encoding/json"

type ApplicationCommandType int

const (
	ChatInputCommandType         ApplicationCommandType = 1
	UserCommandType              ApplicationCommandType = 2
	MessageCommandType           ApplicationCommandType = 3
	PrimaryEntryPointCommandType ApplicationCommandType = 4
)

type ApplicationIntegrationType int

const (
	GuildInstallIntegrationType ApplicationIntegrationType = iota
	UserInstallIntegrationType
)

type ApplicationInteractionContextType int

const (
	GuildInteractionContextType ApplicationInteractionContextType = iota
	BotDMInteractionContextType
	PrivateChannelInteractionContextType
)

type ApplicationCommandOptionType int

const (
	SubCommandOptionType ApplicationCommandOptionType = iota + 1
	SubCommandGroupOptionType
	StringOptionType
	IntegerOptionType
	BooleanOptionType
	UserOptionType
	ChannelOptionType
	RoleOptionType
	MentionableOptionType
	NumberOptionType
	AttachmentOptionType
)

type ApplicationCommandOption struct {
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Type        ApplicationCommandOptionType `json:"type"`
	Options     []ApplicationCommandOption   `json:"options,omitempty"`
	Choices     []ApplicationCommandChoice   `json:"choices,omitempty"`
	Required    bool                         `json:"required,omitempty"`
	MinLength   int                          `json:"min_length,omitempty"`
	MaxLength   int                          `json:"max_length,omitempty"`
}

type ApplicationCommandChoice struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
type Command struct {
	Name             string                              `json:"name"`
	Description      string                              `json:"description"`
	Type             ApplicationCommandType              `json:"type"`
	Options          []ApplicationCommandOption          `json:"options,omitempty"`
	IntegrationTypes []ApplicationIntegrationType        `json:"integration_types,omitempty"`
	Contexts         []ApplicationInteractionContextType `json:"contexts,omitempty"`
}

type ApplicationCommandData struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     ApplicationCommandType `json:"type"`
	Options  []InteractionOption    `json:"options,omitempty"`
	TargetID string                 `json:"target_id,omitempty"`
}

// InteractionOption is an option submitted with an application command
// interaction. Unlike ApplicationCommandOption, it carries the user's value.
type InteractionOption struct {
	Name    string                       `json:"name"`
	Type    ApplicationCommandOptionType `json:"type"`
	Value   json.RawMessage              `json:"value,omitempty"`
	Options []InteractionOption          `json:"options,omitempty"`
}

// Subcommand returns the invoked subcommand's name, if the command uses subcommands.
func (d ApplicationCommandData) Subcommand() (string, bool) {
	for _, opt := range d.Options {
		if opt.Type == SubCommandOptionType {
			return opt.Name, true
		}
	}

	return "", false
}

// SubcommandOptions returns the options passed to the invoked subcommand, or the
// command's own options when it has no subcommands.
func (d ApplicationCommandData) SubcommandOptions() []InteractionOption {
	for _, opt := range d.Options {
		if opt.Type == SubCommandOptionType {
			return opt.Options
		}
	}

	return d.Options
}

// StringValue decodes the option's value as a string.
func (o InteractionOption) StringValue() (string, bool) {
	var value string
	if err := json.Unmarshal(o.Value, &value); err != nil {
		return "", false
	}

	return value, true
}

// IntValue decodes the option's value as an integer.
func (o InteractionOption) IntValue() (int, bool) {
	var value int
	if err := json.Unmarshal(o.Value, &value); err != nil {
		return 0, false
	}

	return value, true
}
