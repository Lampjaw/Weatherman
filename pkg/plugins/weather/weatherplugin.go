package weatherplugin

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lampjaw/discordgobot"
	"github.com/lampjaw/weatherman/pkg/darksky"
	"github.com/lampjaw/weatherman/pkg/herelocation"
)

type weatherPlugin struct {
	discordgobot.Plugin
	herelocationClient *herelocation.HereLocationClient
	darkSkyClient      *darksky.DarkSkyClient
}

func New(config WeatherConfig) discordgobot.IPlugin {
	return &weatherPlugin{
		herelocationClient: herelocation.NewClient(config.HereAppID, config.HereAppCode),
		darkSkyClient:      darksky.NewClient(config.DarkSkySecretKey),
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
					Alias:    "LocationText",
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
					Alias:    "LocationText",
				},
			},
			Description: "Get the forecasted weather for a location",
			Callback:    p.runForecastWeatherCommand,
		},
	}
}

func (p *weatherPlugin) Name() string {
	return "Weather"
}

func (p *weatherPlugin) runCurrentWeatherCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	location := args["LocationText"]

	geoLocation, err := p.herelocationClient.GetLocationByTextAsync(location)

	if err != nil {
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("Failed to find location '%s'", location))
		p.Unlock()
		return
	}

	darkSkyResponse, err := p.darkSkyClient.GetCurrentWeather(geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude)

	if err != nil {
		fmt.Println(err)
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("Failed to get weather for location '%s'", location))
		p.Unlock()
		return
	}

	weather := convertCurrentDarkSkyResponse(darkSkyResponse)

	description := fmt.Sprintf("Currently %s and %s with a high of %s and a low of %s.",
		convertToTempString(weather.Temperature), weather.Condition, convertToTempString(weather.ForecastHigh), convertToTempString(weather.ForecastLow))

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: buildLocationString(geoLocation),
		},
		Color:       0x070707,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Wind Speed",
				Value:  fmt.Sprintf("%0.1f MpH", weather.WindSpeed),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Wind Chill",
				Value:  convertToTempString(weather.WindChill),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Humidity",
				Value:  fmt.Sprintf("%d%%", int32(weather.Humidity*100)),
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Heat Index",
				Value:  convertToTempString(weather.HeatIndex),
				Inline: true,
			},
		},
	}

	p.RLock()
	client.SendEmbedMessage(message.Channel(), embed)
	p.RUnlock()
}

func (p *weatherPlugin) runForecastWeatherCommand(bot *discordgobot.Gobot, client *discordgobot.DiscordClient, message discordgobot.Message, args map[string]string, trigger string) {
	location := args["LocationText"]

	geoLocation, err := p.herelocationClient.GetLocationByTextAsync(location)

	if err != nil {
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("Failed to find location '%s'", location))
		p.Unlock()
		return
	}

	darkSkyResponse, err := p.darkSkyClient.GetCurrentWeather(geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude)

	if err != nil {
		fmt.Println(err)
		p.Lock()
		client.SendMessage(message.Channel(), fmt.Sprintf("Failed to get weather for location '%s'", location))
		p.Unlock()
		return
	}

	weatherDays := convertForecastDarkSkyResponse(darkSkyResponse)

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

func createWeatherDay(d WeatherDay) string {
	var temperatureHigh = convertToTempString(d.High)
	var temperatureLow = convertToTempString(d.Low)
	return fmt.Sprintf("%s: %s / %s - %s", d.Day, temperatureHigh, temperatureLow, d.Text)
}

func convertCurrentDarkSkyResponse(resp *darksky.DarkSkyResponse) *CurrentWeather {
	currentDay := resp.Daily.Data[0]

	temp := resp.Currently.Temperature
	humidity := resp.Currently.Humidity
	windSpeed := resp.Currently.WindSpeed
	heatIndex := calculateHeatIndex(temp, humidity)
	windChill := calculateWindChill(temp, windSpeed)

	return &CurrentWeather{
		Condition:    resp.Currently.Summary,
		Temperature:  temp,
		Humidity:     humidity,
		WindChill:    windChill,
		WindSpeed:    windSpeed,
		ForecastHigh: currentDay.TemperatureHigh,
		ForecastLow:  currentDay.TemperatureLow,
		HeatIndex:    heatIndex,
		Icon:         currentDay.Icon,
	}
}

func convertForecastDarkSkyResponse(resp *darksky.DarkSkyResponse) []WeatherDay {
	result := make([]WeatherDay, 0)

	for _, day := range resp.Daily.Data {
		date := time.Unix(day.Time, 0)
		weatherDay := WeatherDay{
			Date: date.Format("01/02/06"),
			Day:  date.Format("Mon"),
			High: day.TemperatureHigh,
			Low:  day.TemperatureLow,
			Text: day.Summary,
			Icon: day.Icon,
		}
		result = append(result, weatherDay)
	}

	return result
}
