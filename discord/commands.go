package discord

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
