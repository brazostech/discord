package book

import (
	"github.com/brazostech/discord/commands"
)

// BookCommand registers the /book command, used to manage book club books
var BookCommand = commands.Command{
	Name:        "book",
	Description: "Book Club management command",
	Type:        commands.ChatInputCommandType,
	Options: []commands.ApplicationCommandOption{
		{
			Name:        "register",
			Description: "Choose which book to register",
			Type:        commands.StringOptionType,
			Required:    false,
			MinLength:   1,
			MaxLength:   2000,
		},
		{
			Name:        "update_chapter",
			Description: "Update which chapter we're on",
			Type:        commands.IntegerOptionType,
			Required:    false,
		},
	},
	IntegrationTypes: []commands.ApplicationIntegrationType{
		commands.GuildInstallIntegrationType,
		commands.UserInstallIntegrationType,
	},
	Contexts: []commands.ApplicationInteractionContextType{
		commands.GuildInteractionContextType,
		commands.BotDMInteractionContextType,
		commands.PrivateChannelInteractionContextType,
	},
}
