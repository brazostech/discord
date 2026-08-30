package commands

// SimpleTestCommand registers the most basic Ping/Pong interaction
var SimpleTestCommand = Command{
	Name:        "test",
	Description: "test command",
	Type:        ChatInputCommandType,
	IntegrationTypes: []ApplicationIntegrationType{
		GuildInstallIntegrationType,
		UserInstallIntegrationType,
	},
	Contexts: []ApplicationInteractionContextType{
		GuildInteractionContextType,
		BotDMInteractionContextType,
		PrivateChannelInteractionContextType,
	},
}

var BookCommand = Command{
	Name:        "book",
	Description: "Book Club management command",
	Type:        ChatInputCommandType,
	Options: []ApplicationCommandOption{
		{
			Name:        "register",
			Description: "Choose which book to register",
			Type:        StringOptionType,
			Required:    false,
			MinLength:   1,
			MaxLength:   2000,
		},
		{
			Name:        "update_chapter",
			Description: "Update which chapter we're on",
			Type:        IntegerOptionType,
			Required:    false,
		},
	},
	IntegrationTypes: []ApplicationIntegrationType{
		GuildInstallIntegrationType,
		UserInstallIntegrationType,
	},
	Contexts: []ApplicationInteractionContextType{
		GuildInteractionContextType,
		BotDMInteractionContextType,
		PrivateChannelInteractionContextType,
	},
}
