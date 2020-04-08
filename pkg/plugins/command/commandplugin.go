package commandplugin

import (
	"log"

	"github.com/lampjaw/discordgobot"
)

type commandPlugin struct {
	discordgobot.Plugin
	repository *repository
}

func New() *commandPlugin {
	plugin := &commandPlugin{
		repository: newRepository(),
	}

	plugin.repository.initRepository()

	return plugin
}

func (p *commandPlugin) Commands() []*discordgobot.CommandDefinition {
	return []*discordgobot.CommandDefinition{
		&discordgobot.CommandDefinition{
			CommandID: "command-setprefix",
			Triggers: []string{
				"setprefix",
			},
			PermissionLevel: discordgobot.PERMISSION_ADMIN,
			ExposureLevel:   discordgobot.EXPOSURE_PUBLIC,
			Arguments: []discordgobot.CommandDefinitionArgument{
				discordgobot.CommandDefinitionArgument{
					Optional: false,
					Pattern:  "\\S+",
					Alias:    "prefix",
				},
			},
			Description: "Set the command prefix for this server",
			Callback:    p.runSetPrefixCommand,
		},
	}
}

func (p *commandPlugin) Name() string {
	return "Command"
}

func (p *commandPlugin) runSetPrefixCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, payload discordgobot.CommandPayload) {
	prefix := payload.Arguments["prefix"]

	channel, _ := client.Channel(payload.Message.Channel())

	err := p.repository.updateGuildPrefix(channel.GuildID, payload.Message.UserID(), prefix)

	p.Lock()

	if err != nil {
		client.SendMessage(payload.Message.Channel(), "Failed to set new prefix.")
	} else {
		client.SendMessage(payload.Message.Channel(), "Prefix set!")
	}

	p.Unlock()
}

func (p *commandPlugin) GetGuildPrefix(guildID string) (*string, error) {
	guildProfile, err := p.repository.getGuildProfile(guildID)

	if err != nil {
		log.Printf("Failed to get guild profile for '%s': %s", guildID, err)
		return nil, err
	}

	if guildProfile == nil {
		return nil, nil
	}

	return guildProfile.Prefix, nil
}
