using DarkSky.Models;
using Discord;
using Discord.Interactions;
using System.Globalization;
using System.Text;
using TimeZoneNames;
using Weatherman.Bot.Models;
using Weatherman.Bot.Services;
using Weatherman.Bot.Utils;

namespace Weatherman.Bot.Modules
{
    [Group("weather", "Weather commands")]
    [EnabledInDm(false)]
    public class WeatherModule : InteractionModuleBase<SocketInteractionContext>
    {
        private readonly LocationService _locationService;
        private readonly WeatherService _weatherService;
        private readonly HomeService _homeService;

        public WeatherModule(LocationService locationService, WeatherService weatherService, HomeService homeService)
        {
            _locationService = locationService;
            _weatherService = weatherService;
            _homeService = homeService;
        }

        [SlashCommand("now", "Get the current forecast.")]
        public async Task GetWeatherNowAsync(string location = null)
        {
            LocationDetails weatherLocation;

            if (string.IsNullOrWhiteSpace(location))
            {
                weatherLocation = await _homeService.GetHomeAsync(Context.User.Id);
                if (weatherLocation == null)
                {
                    await RespondAsync("Please include a location or set a home. To set a home use `/home set <location>`.");
                    return;
                }
            }
            else
            {
                weatherLocation = await _locationService.GetGeocodeForLocationStringAsync(location);
                if (weatherLocation == null)
                {
                    await RespondAsync("Failed to resolve this location.");
                    return;
                }
            }

            var forecast = await _weatherService.GetCurrentForecastAsync(weatherLocation.Coordinates);
            if (forecast == null)
            {
                await RespondAsync("Failed to find a forecast for this location.");
                return;
            }

            var descriptionBuilder = new StringBuilder();

            descriptionBuilder.Append(
                string.Format("{0} Currently {1} and {2} with a high of {3} and a low of {4}.",
                    EmojiIconMap.Resolve(forecast.Icon),
                    ConvertToTempString(forecast.Temperature, weatherLocation),
                    forecast.Condition,
                    ConvertToTempString(forecast.ForecastHigh, weatherLocation),
                    ConvertToTempString(forecast.ForecastLow, weatherLocation)));

            if (forecast.Alerts != null && forecast.Alerts.Any())
            {
                descriptionBuilder.AppendLine();

                var tzName = TZNames.GetAbbreviationsForTimeZone(forecast.TimeZone, CultureInfo.CurrentCulture.Name);

                foreach (var alert in forecast.Alerts)
                {
                    descriptionBuilder.AppendLine(
                        string.Format("[**{0}**]({1}) Until {2:dd ddd yy HH:mm} {3}", alert.Title, alert.Uri, alert.ExpirationDate, tzName));
                }
            }

            var fieldBuilders = new List<EmbedFieldBuilder>();

            if (forecast.PrecipitationProbability >= 0.05)
            {
                var precipAccumulation = 0.0;

                if (forecast.SnowAccumulation > 0 && forecast.PrecipitationType == PrecipitationType.Snow)
                {
                    precipAccumulation = forecast.SnowAccumulation;
                }
                else if (forecast.PrecipitationIntensity > 0)
                {
                    precipAccumulation = forecast.PrecipitationIntensity * 24;
                }

                if (precipAccumulation > 0.1)
                {
                    fieldBuilders.Add(
                        new EmbedFieldBuilder()
                            .WithIsInline(true)
                            .WithName("Precipitation")
                            .WithValue(
                                string.Format("There is a {0:P0} chance of {1} with an estimated accumulation of {2:F1} inches",
                                    forecast.PrecipitationProbability, forecast.PrecipitationType.ToString().ToLower(), precipAccumulation)));
                }
            }

            fieldBuilders.AddRange(new[] {
                new EmbedFieldBuilder()
                    .WithIsInline(true)
                    .WithName("Wind")
                    .WithValue(string.Format("{0:F1} MpH with gusts up to {1:F1} MpH", forecast.WindSpeed, forecast.WindGust)),
                new EmbedFieldBuilder()
                    .WithIsInline(true)
                    .WithName("Humidity")
                    .WithValue(string.Format("{0:N0}%", forecast.Humidity))
            });

            if (forecast.Temperature >= 80 && forecast.Humidity >= 40)
            {
                fieldBuilders.Add(
                    new EmbedFieldBuilder()
                        .WithIsInline(true)
                        .WithName("Heat Index")
                        .WithValue(ConvertToTempString(forecast.HeatIndex, weatherLocation)));
            }

            if (forecast.Temperature <= 50 && forecast.WindGust >= 3)
            {
                fieldBuilders.Add(
                    new EmbedFieldBuilder()
                        .WithIsInline(true)
                        .WithName("Wind Chill")
                        .WithValue(ConvertToTempString(forecast.WindChill, weatherLocation)));
            }

            if (forecast.UVIndex > 0)
            {
                fieldBuilders.Add(
                    new EmbedFieldBuilder()
                        .WithIsInline(true)
                        .WithName("UV Index")
                        .WithValue(string.Format("({0}) {1}", forecast.UVIndex, GetUvIndexString(forecast.UVIndex))));
            }

            var embed = new EmbedBuilder()
                .WithAuthor(GetLocationString(weatherLocation))
                .WithTitle(Constants.TitleSeeMoreText)
                .WithUrl(string.Format(Constants.TitleSeeMoreUrlFormat, weatherLocation.Coordinates.Latitude, weatherLocation.Coordinates.Longitude))
                .WithColor(Constants.DefaultEmbedColor)
                .WithDescription(descriptionBuilder.ToString())
                .WithFields(fieldBuilders)
                .WithFooter(Constants.FooterPoweredByText)
                .Build();

            await RespondAsync(embed: embed);
        }

        [SlashCommand("week", "Get the weekly forecast.")]
        public async Task GetWeatherWeekAsync(string location = null)
        {
            LocationDetails weatherLocation;

            if (string.IsNullOrWhiteSpace(location))
            {
                weatherLocation = await _homeService.GetHomeAsync(Context.User.Id);
                if (weatherLocation == null)
                {
                    await RespondAsync("Please include a location or set a home. To set a home use `/home set <location>`.");
                    return;
                }
            }
            else
            {
                weatherLocation = await _locationService.GetGeocodeForLocationStringAsync(location);
                if (weatherLocation == null)
                {
                    await RespondAsync("Failed to resolve this location.");
                    return;
                }
            }

            var weatherSummaries = await _weatherService.GetWeeklyForecastAsync(weatherLocation.Coordinates);
            if (weatherSummaries == null)
            {
                await RespondAsync("Failed to find a forecast for this location.");
                return;
            }

            var fieldBuilders = weatherSummaries
                .Take(Constants.MaxForecastDays)
                .Select(a =>
                {
                    return new EmbedFieldBuilder()
                        .WithIsInline(false)
                        .WithName(a.Date.ToString("dddd MMMM d"))
                        .WithValue(GetWeatherSummaryString(a, weatherLocation));
                });

            var embed = new EmbedBuilder()
                .WithAuthor(GetLocationString(weatherLocation))
                .WithTitle(Constants.TitleSeeMoreText)
                .WithUrl(string.Format(Constants.TitleSeeMoreUrlFormat, weatherLocation.Coordinates.Latitude, weatherLocation.Coordinates.Longitude))
                .WithColor(Constants.DefaultEmbedColor)
                .WithFields(fieldBuilders)
                .WithFooter(Constants.FooterPoweredByText)
                .Build();

            await RespondAsync(embed: embed);
        }

        private string GetUvIndexString(int uvIndex)
        {
            switch (uvIndex)
            {
                case < 3:
                    return "Low";
                case < 6:
                    return "Moderate";
                case < 8:
                    return "High";
                case < 11:
                    return "Very High";
                case >= 11:
                    return "Extreme";
            };
        }

        private string ConvertToTempString(double temperature, LocationDetails location)
        {
            var tempCelsius = ConvertToCelsius(temperature);

            if (location.Country == "United States" || location.Country == "USA")
            {
                return string.Format("{0:N0} °F ({1:N0} °C)", temperature, tempCelsius);
            }

            return string.Format("{0:N0} °C ({1:N0} °F)", tempCelsius, temperature);
        }

        private double ConvertToCelsius(double temperature)
        {
            return (temperature - 32.0) / 1.8;
        }

        private string GetLocationString(LocationDetails location)
        {
            var sb = new StringBuilder();

            if (!string.IsNullOrEmpty(location.City))
            {
                sb.Append($"{location.City}, ");
            }

            if (!string.IsNullOrEmpty(location.Region))
            {
                sb.Append($"{location.Region} - ");
            }

            sb.Append(location.Country);

            return sb.ToString();
        }

        private string GetWeatherSummaryString(WeatherSummary d, LocationDetails location)
        {
            return string.Format("{1} {2} / {3} - {4}",
                d.Date,
                EmojiIconMap.Resolve(d.Icon),
                ConvertToTempString(d.High, location),
                ConvertToTempString(d.Low, location),
                d.Summary);
        }
    }
}
