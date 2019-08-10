package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/lampjaw/discordgobot"
	inviteplugin "github.com/lampjaw/weatherman/pkg/plugins/invite"
	statsplugin "github.com/lampjaw/weatherman/pkg/plugins/stats"
	weatherplugin "github.com/lampjaw/weatherman/pkg/plugins/weather"
)

// VERSION of Weatherman
const VERSION = "2.0.0"

func init() {
	token = os.Getenv("DiscordToken")
	clientID = os.Getenv("DiscordClientId")
	ownerUserID = os.Getenv("DiscordOwnerId")
	hereAppID = os.Getenv("HereAppId")
	hereAppCode = os.Getenv("HereAppCode")
	darkSkySecretKey = os.Getenv("DarkSkySecretKey")
}

var token string
var clientID string
var ownerUserID string
var hereAppID string
var hereAppCode string
var darkSkySecretKey string

func main() {
	if token == "" {
		fmt.Println("No token provided.")
		return
	}

	config := &discordgobot.GobotConf{
		OwnerUserID:       ownerUserID,
		ClientID:          clientID,
		CommandPrefixFunc: getCommandPrefix,
	}

	bot, err := discordgobot.NewBot(token, config)

	if err != nil {
		fmt.Sprintln("Unable to create bot: %s", err)
		return
	}

	bot.RegisterPlugin(inviteplugin.New())
	bot.RegisterPlugin(statsplugin.New())

	weatherConfig := weatherplugin.WeatherConfig{
		HereAppID:        hereAppID,
		HereAppCode:      hereAppCode,
		DarkSkySecretKey: darkSkySecretKey,
	}
	bot.RegisterPlugin(weatherplugin.New(weatherConfig))

	bot.Open()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			break out
		}
	}
}

func getCommandPrefix(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message) string {
	return "?"
}
