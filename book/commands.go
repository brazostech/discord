package book

import (
	"github.com/brazostech/discord/discord"
)

// BookCommand registers the /book command, used to manage book club books
var BookCommand = discord.Command{
	Name:        "book",
	Description: "Book Club management command",
	Type:        discord.ChatInputCommandType,
	Options: []discord.ApplicationCommandOption{
		{
			Name:        "register",
			Description: "Choose which book to register",
			Type:        discord.StringOptionType,
			Required:    false,
			MinLength:   1,
			MaxLength:   2000,
		},
		{
			Name:        "update_chapter",
			Description: "Update which chapter we're on",
			Type:        discord.IntegerOptionType,
			Required:    false,
		},
	},
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.GuildInstallIntegrationType,
		discord.UserInstallIntegrationType,
	},
	Contexts: []discord.ApplicationInteractionContextType{
		discord.GuildInteractionContextType,
		discord.BotDMInteractionContextType,
		discord.PrivateChannelInteractionContextType,
	},
}
