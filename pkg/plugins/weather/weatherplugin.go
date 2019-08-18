package weatherplugin

import (
	"fmt"
	"time"

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
		convertToTempString(weather.Temperature, geoLocation),
		weather.Condition,
		convertToTempString(weather.ForecastHigh, geoLocation),
		convertToTempString(weather.ForecastLow, geoLocation))

	if weather.Alerts != nil && len(weather.Alerts) > 0 {
		for _, alert := range weather.Alerts {
			expiration := time.Unix(alert.Expires, 0).Format("02 Jan 06 15:04 MST")
			description += fmt.Sprintf("\n[**%s**](%s) Until %s", alert.Title, alert.Uri, expiration)
		}
	}

	fields := make([]*discordgo.MessageEmbedField, 0)

	if weather.PrecipitationProbability >= 5 {
		var precipAccumulation float64
		if weather.SnowAccumulation > 0 && weather.PrecipitationType == "snow" {
			precipAccumulation = weather.SnowAccumulation
		} else if weather.PrecipitationIntensity > 0 {
			precipAccumulation = weather.PrecipitationIntensity * 24
		}

		precipMsg := fmt.Sprintf("There is a %d%% chance of %s with an estimated accumulation of %0.1f inches",
			int32(weather.PrecipitationProbability), weather.PrecipitationType, precipAccumulation)

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Precipitation",
			Value:  precipMsg,
			Inline: true,
		})
	}

	fields = append(fields,
		&discordgo.MessageEmbedField{
			Name:   "Wind",
			Value:  fmt.Sprintf("%0.1f MpH with gusts up to %0.1f MpH", weather.WindSpeed, weather.WindGust),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Humidity",
			Value:  fmt.Sprintf("%d%%", int32(weather.Humidity)),
			Inline: true,
		})

	if weather.Temperature >= 80 && weather.Humidity >= 40 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Heat Index",
			Value:  convertToTempString(weather.HeatIndex, geoLocation),
			Inline: true,
		})
	}

	if weather.Temperature <= 50 && weather.WindSpeed >= 3 {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Wind Chill",
			Value:  convertToTempString(weather.WindChill, geoLocation),
			Inline: true,
		})
	}

	if weather.UVIndex > 0 {
		var indexMsg string
		switch {
		case weather.UVIndex >= 0.0 && weather.UVIndex <= 2.9:
			indexMsg = "Low"
		case weather.UVIndex >= 3.0 && weather.UVIndex <= 5.9:
			indexMsg = "Moderate"
		case weather.UVIndex >= 6.0 && weather.UVIndex <= 7.9:
			indexMsg = "High"
		case weather.UVIndex >= 8.0 && weather.UVIndex <= 10.9:
			indexMsg = "Very High"
		case weather.UVIndex >= 11:
			indexMsg = "Extreme"
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "UV Index",
			Value:  fmt.Sprintf("(%0.1f) %s", weather.UVIndex, indexMsg),
			Inline: true,
		})
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: buildLocationString(geoLocation),
		},
		Title:       "See more at darksky.com",
		URL:         fmt.Sprintf("https://darksky.net/forecast/%0.4f,%0.4f", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude),
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
			Value:  createWeatherDay(weatherDays[i], geoLocation),
			Inline: false,
		}
		messageFields = append(messageFields, field)
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: buildLocationString(geoLocation),
		},
		Title:  "See more at darksky.com",
		URL:    fmt.Sprintf("https://darksky.net/forecast/%0.4f,%0.4f", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude),
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

func convertToTempString(temp float64, geoLocation *herelocation.GeoLocation) string {
	var tempCelsius = convertToCelsius(temp)

	if geoLocation.Country == "USA" {
		return fmt.Sprintf("%d °F (%d °C)", int32(temp), int32(tempCelsius))
	}

	return fmt.Sprintf("%d °C (%d °F)", int32(tempCelsius), int32(temp))
}

func convertToCelsius(temp float64) float64 {
	return (temp - 32.0) / 1.8
}

func createWeatherDay(d *WeatherDay, geoLocation *herelocation.GeoLocation) string {
	var temperatureHigh = convertToTempString(d.High, geoLocation)
	var temperatureLow = convertToTempString(d.Low, geoLocation)
	return fmt.Sprintf("%s: %s %s / %s - %s", d.Day, iconToEmojiMap[d.Icon], temperatureHigh, temperatureLow, d.Text)
}
