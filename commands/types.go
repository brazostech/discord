package commands

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
	ID       string                     `json:"id"`
	Name     string                     `json:"name"`
	Type     ApplicationCommandType     `json:"type"`
	Options  []ApplicationCommandOption `json:"options,omitempty"`
	TargetID string                     `json:"target_id,omitempty"`
}
