package weatherplugin

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/lampjaw/discordgobot"
	"github.com/lampjaw/weatherman/pkg/herelocation"
)

type weatherPlugin struct {
	discordgobot.Plugin
	manager *weatherManager
}

func New(config WeatherConfig) *weatherPlugin {
	return &weatherPlugin{
		manager: newWeatherManager(config),
	}
}

func (p *weatherPlugin) Commands() []discordgobot.CommandDefinition {
	return []discordgobot.CommandDefinition{
		discordgobot.CommandDefinition{
			CommandID: "weather-current",
			Triggers: []string{
				"w",
			},
			Arguments: []discordgobot.CommandDefinitionArgument{
				discordgobot.CommandDefinitionArgument{
					Optional: true,
					Pattern:  ".*",
					Alias:    "location",
				},
			},
			Description: "Get the current weather for a location",
			Callback:    p.runCurrentWeatherCommand,
		},
		discordgobot.CommandDefinition{
			CommandID: "weather-forecast",
			Triggers: []string{
				"wf",
			},
			Arguments: []discordgobot.CommandDefinitionArgument{
				discordgobot.CommandDefinitionArgument{
					Optional: true,
					Pattern:  ".*",
					Alias:    "location",
				},
			},
			Description: "Get the forecasted weather for a location",
			Callback:    p.runForecastWeatherCommand,
		},
		discordgobot.CommandDefinition{
			CommandID: "weather-sethome",
			Triggers: []string{
				"sethome",
			},
			Arguments: []discordgobot.CommandDefinitionArgument{
				discordgobot.CommandDefinitionArgument{
					Optional: false,
					Pattern:  ".*",
					Alias:    "location",
				},
			},
			Description: "Set a location to remember as your home",
			Callback:    p.runSetHomeCommand,
		},
	}
}

func (p *weatherPlugin) Name() string {
	return "Weather"
}

func (p *weatherPlugin) runCurrentWeatherCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	location := args["location"]

	weather, geoLocation, err := p.manager.getCurrentWeatherByLocation(message.UserID(), location)

	if err != nil {
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.Unlock()
		return
	}

	description := fmt.Sprintf("%s Currently %s and %s with a high of %s and a low of %s.", iconToEmojiMap[weather.Icon],
		convertToTempString(weather.Temperature), weather.Condition, convertToTempString(weather.ForecastHigh), convertToTempString(weather.ForecastLow))

	fields := []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name:   "Wind Speed",
			Value:  fmt.Sprintf("%0.1f MpH", weather.WindSpeed),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Humidity",
			Value:  fmt.Sprintf("%d%%", int32(weather.Humidity)),
			Inline: true,
		},
	}

	if weather.Temperature >= 80 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Heat Index",
			Value:  convertToTempString(weather.HeatIndex),
			Inline: true,
		})
	}

	if weather.Temperature <= 50 && weather.WindSpeed >= 3 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Wind Chill",
			Value:  convertToTempString(weather.WindChill),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: buildLocationString(geoLocation),
		},
		Color:       0x070707,
		Description: description,
		Fields:      fields,
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *weatherPlugin) runForecastWeatherCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	location := args["location"]

	weatherDays, geoLocation, err := p.manager.getForecastWeatherByLocation(message.UserID(), location)

	if err != nil {
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
		p.Unlock()
		return
	}

	var messageFields []*discordgo.MessageEmbedField

	for i := 0; i < 5; i++ {
		var field = &discordgo.MessageEmbedField{
			Name:   weatherDays[i].Date,
			Value:  createWeatherDay(weatherDays[i]),
			Inline: false,
		}
		messageFields = append(messageFields, field)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: buildLocationString(geoLocation),
		},
		Color:  0x070707,
		Fields: messageFields,
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *weatherPlugin) runSetHomeCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	location := args["location"]

	err := p.manager.setUserHomeLocation(message.UserID(), location)

	p.Lock()

	if err != nil {
		client.SendMessage(message.Channel(), fmt.Sprintf("%s", err))
	} else {
		client.SendMessage(message.Channel(), "Home set!")
	}

	p.Unlock()
}

func buildLocationString(location *herelocation.GeoLocation) string {
	cityPart := ""
	regionPart := ""

	if location.City != "" {
		cityPart = fmt.Sprintf("%s, ", location.City)
	}

	if location.Region != "" {
		regionPart = fmt.Sprintf("%s - ", location.Region)
	}

	return fmt.Sprintf("%s%s%s", cityPart, regionPart, location.Country)
}

func convertToTempString(temp float64) string {
	var tempCelsius = convertToCelsius(temp)
	return fmt.Sprintf("%d °F (%d °C)", int32(temp), int32(tempCelsius))
}

func convertToCelsius(temp float64) float64 {
	return (temp - 32.0) / 1.8
}

func createWeatherDay(d *WeatherDay) string {
	var temperatureHigh = convertToTempString(d.High)
	var temperatureLow = convertToTempString(d.Low)
	return fmt.Sprintf("%s: %s %s / %s - %s", d.Day, iconToEmojiMap[d.Icon], temperatureHigh, temperatureLow, d.Text)
}
