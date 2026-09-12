package book

import "github.com/brazostech/discord/discord"

// Command registers the /book command, used to manage the club's Current Book.
var Command = discord.Command{
	Name:        "book",
	Description: "Book Club management command",
	Type:        discord.ChatInputCommandType,
	Options: []discord.ApplicationCommandOption{
		{
			Name:        "register",
			Description: "Register the book the club is reading",
			Type:        discord.SubCommandOptionType,
			Options: []discord.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "Book title",
					Type:        discord.StringOptionType,
					Required:    true,
					MinLength:   1,
					MaxLength:   2000,
				},
				{
					Name:        "url",
					Description: "Link to the book",
					Type:        discord.StringOptionType,
					MaxLength:   2000,
				},
			},
		},
		{
			Name:        "update-chapter",
			Description: "Update which chapter we're on",
			Type:        discord.SubCommandOptionType,
			Options: []discord.ApplicationCommandOption{
				{
					Name:        "chapter",
					Description: "Current chapter",
					Type:        discord.IntegerOptionType,
					Required:    true,
				},
			},
		},
	},
	IntegrationTypes: []discord.ApplicationIntegrationType{
		discord.GuildInstallIntegrationType,
	},
	Contexts: []discord.ApplicationInteractionContextType{
		discord.GuildInteractionContextType,
	},
}
