package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/lampjaw/discordgobot"
	commandplugin "github.com/lampjaw/weatherman/pkg/plugins/command"
	inviteplugin "github.com/lampjaw/weatherman/pkg/plugins/invite"
	statsplugin "github.com/lampjaw/weatherman/pkg/plugins/stats"
	weatherplugin "github.com/lampjaw/weatherman/pkg/plugins/weather"
)

// VERSION of Weatherman
const VERSION = "2.1.5"

func init() {
	token = os.Getenv("DiscordToken")
	clientID = os.Getenv("DiscordClientId")
	ownerUserID = os.Getenv("DiscordOwnerId")
	hereAppID = os.Getenv("HereAppId")
	hereAppCode = os.Getenv("HereAppCode")
	darkSkySecretKey = os.Getenv("DarkSkySecretKey")
	darkSkySecretKey = os.Getenv("DarkSkySecretKey")
	redisAddress = os.Getenv("RedisAddress")
}

var token string
var clientID string
var ownerUserID string
var hereAppID string
var hereAppCode string
var darkSkySecretKey string
var redisAddress string

func main() {
	if token == "" {
		fmt.Println("No token provided.")
		return
	}

	commandPlugin := commandplugin.New()

	config := &discordgobot.GobotConf{
		OwnerUserID: ownerUserID,
		ClientID:    clientID,
		CommandPrefixFunc: func(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message) string {
			channel, _ := client.Channel(message.Channel())
			prefix, err := commandPlugin.GetGuildPrefix(channel.GuildID)

			if err != nil || prefix == nil {
				return "?"
			}

			return *prefix
		},
	}

	bot, err := discordgobot.NewBot(token, config)

	if err != nil {
		log.Printf("Unable to create bot: %s", err)
		return
	}

	weatherConfig := weatherplugin.WeatherConfig{
		HereAppID:        hereAppID,
		HereAppCode:      hereAppCode,
		DarkSkySecretKey: darkSkySecretKey,
		RedisAddress:     redisAddress,
	}

	bot.RegisterPlugin(commandPlugin)
	bot.RegisterPlugin(inviteplugin.New())
	bot.RegisterPlugin(statsplugin.New(VERSION))
	bot.RegisterPlugin(weatherplugin.New(weatherConfig))

	bot.Open()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			bot.Client.Session.Close()
			break out
		}
	}
}
