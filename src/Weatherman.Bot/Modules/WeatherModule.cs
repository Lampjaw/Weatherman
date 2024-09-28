﻿using DarkSky.Models;
using Discord;
using Discord.Interactions;
using Discord.WebSocket;
using System.Text;
using TimeZoneNames;
using Weatherman.Bot.Models;
using Weatherman.Bot.Services;
using Weatherman.Bot.Utils;

namespace Weatherman.Bot.Modules
{
    [Group("weather", "Weather commands")]
    [EnabledInDm(true)]
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
            await DeferAsync();

            var weatherLocation = await ResolveUserLocationAsync(location, Context.User.Id);
            if (weatherLocation == null)
            {
                return;
            }

            var forecastData = await _weatherService.GetCurrentForecastAsync(weatherLocation.Coordinates);
            if (forecastData == null)
            {
                await ModifyOriginalResponseAsync(properties =>
                    properties.Content = "Failed to find a forecast for this location.");
                return;
            }

            var forecast = forecastData.Data;

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

                var tzCode = GetTimeZoneCode(forecastData.TimeZone);

                foreach (var alert in forecast.Alerts)
                {
                    var issueIdx = alert.Title.IndexOf("issued");
                    var alertTitle = issueIdx > 0 ? alert.Title.Substring(0, issueIdx).Trim() : alert.Title;

                    descriptionBuilder.AppendLine(
                        string.Format("[**{0}**]({1}) Until {2:dd MMM yy HH:mm} {3}", alertTitle, alert.Uri, alert.ExpirationDate, tzCode));
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

            await ModifyOriginalResponseAsync(properties => properties.Embed = embed);
        }

        [SlashCommand("hourly", "Get the hourly forecast.")]
        public async Task GetWeatherHourlyAsync(string location = null)
        {
            await ProcessGetWeatherHourly(0, Context.User.Id, location);
        }

        [ComponentInteraction("forecast_hourly_*_(*)", true)]
        public async Task GetWeatherHourly_ButtonAsync(int page, string encodedLocation)
        {
            var originalUserId = (Context.Interaction as SocketMessageComponent).Message.Interaction.User.Id;

            if (Context.Interaction.User.Id != originalUserId)
            {
                await RespondAsync();
                return;
            }

            var location = string.IsNullOrEmpty(encodedLocation) ? null : Uri.UnescapeDataString(encodedLocation);

            await ProcessGetWeatherHourly(page, originalUserId, location);
        }

        private async Task ProcessGetWeatherHourly(int page, ulong userId, string location = null)
        {
            await DeferAsync();

            var weatherLocation = await ResolveUserLocationAsync(location, userId);
            if (weatherLocation == null)
            {
                return;
            }

            var forecastData = await _weatherService.GetHourlyForecastAsync(weatherLocation.Coordinates);
            if (forecastData == null)
            {
                await ModifyOriginalResponseAsync(properties =>
                        properties.Content = "Failed to find a forecast for this location.");
                return;
            }

            var tz = GetTimeZoneCode(forecastData.TimeZone);

            var fieldBuilders = forecastData.Data
                .Skip(page * Constants.ForecastHoursPerPageLimit)
                .Take(Constants.ForecastHoursPerPageLimit)
                .Select(a =>
                {
                    var tzTime = TimeZoneInfo.ConvertTimeBySystemTimeZoneId(a.Date, forecastData.TimeZone);

                    var fieldName = string.Format(
                        "{0} - {1} {2}",
                        tzTime.ToString("h:mm tt"),
                        EmojiIconMap.Resolve(a.Icon),
                        a.Summary);

                    var fieldValue = string.Format(
                        "{0} | :droplet: {1:N0}% ({2:F2} in) | :dash: {3:N0} mph {4}",
                        ConvertToTempString(a.Temperature, weatherLocation),
                        a.PrecipitationProbability,
                        a.PrecipitationIntensity,
                        a.WindSpeed,
                        WindBearingConverter.ConvertToWindDirection(a.WindBearing),
                        a.Summary);

                    return new EmbedFieldBuilder()
                        .WithIsInline(false)
                        .WithName(fieldName)
                        .WithValue(fieldValue);
                });

            var embed = new EmbedBuilder()
                .WithAuthor(GetLocationString(weatherLocation))
                .WithTitle(Constants.TitleSeeMoreText)
                .WithUrl(string.Format(Constants.TitleSeeMoreUrlFormat, weatherLocation.Coordinates.Latitude, weatherLocation.Coordinates.Longitude))
                .WithColor(Constants.DefaultEmbedColor)
                .WithFields(fieldBuilders)
                .WithFooter(Constants.FooterPoweredByText)
                .Build();

            var encodedLocation = string.IsNullOrEmpty(location) ? null : Uri.EscapeDataString(location);

            var component = new ComponentBuilder()
                .WithButton(
                    customId: $"forecast_hourly_{page - 1}_({encodedLocation})",
                    emote: new Emoji("⬅️"),
                    disabled: page == 0)
                .WithButton(
                    customId: $"forecast_hourly_{page + 1}_({encodedLocation})",
                    emote: new Emoji("➡️"),
                    disabled: page >= Constants.MaxForecastHours / Constants.ForecastHoursPerPageLimit - 1)
                .Build();

            await ModifyOriginalResponseAsync(properties =>
            {
                properties.Embed = embed;
                properties.Components = component;
            });
        }


        [SlashCommand("week", "Get the weekly forecast.")]
        public async Task GetWeatherWeekAsync(string location = null)
        {
            await DeferAsync();

            var weatherLocation = await ResolveUserLocationAsync(location, Context.User.Id);
            if (weatherLocation == null)
            {
                return;
            }

            var forecastData = await _weatherService.GetWeeklyForecastAsync(weatherLocation.Coordinates);
            if (forecastData == null)
            {
                await ModifyOriginalResponseAsync(properties =>
                        properties.Content = "Failed to find a forecast for this location.");
                return;
            }

            var fieldBuilders = forecastData.Data
                .Take(Constants.MaxForecastDays)
                .Select(a =>
                {
                    return new EmbedFieldBuilder()
                        .WithIsInline(false)
                        .WithName(a.Date.ToString("dddd MMMM d"))
                        .WithValue(GetWeatherDailyString(a, weatherLocation));
                });

            var embed = new EmbedBuilder()
                .WithAuthor(GetLocationString(weatherLocation))
                .WithTitle(Constants.TitleSeeMoreText)
                .WithUrl(string.Format(Constants.TitleSeeMoreUrlFormat, weatherLocation.Coordinates.Latitude, weatherLocation.Coordinates.Longitude))
                .WithColor(new Color(Constants.DefaultEmbedColor))
                .WithFields(fieldBuilders)
                .WithFooter(Constants.FooterPoweredByText)
                .Build();

            await ModifyOriginalResponseAsync(properties => properties.Embed = embed);
        }

        private async Task<LocationDetails> ResolveUserLocationAsync(string location, ulong userId)
        {
            if (string.IsNullOrWhiteSpace(location))
            {
                var homeLocation = await _homeService.GetHomeAsync(userId);
                if (homeLocation == null || homeLocation.Coordinates == null)
                {
                    await ModifyOriginalResponseAsync(properties =>
                        properties.Content = "Please include a location or set a home. To set a home use `/home set <location>`.");
                    return null;
                }
                return homeLocation;
            }

            var weatherLocation = await _locationService.GetGeocodeForLocationStringAsync(location);
            if (weatherLocation == null || weatherLocation.Coordinates == null)
            {
                await ModifyOriginalResponseAsync(properties =>
                    properties.Content = "Failed to resolve this location.");
                return null;
            }
            return weatherLocation;
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

        private string GetWeatherDailyString(ForecastDay d, LocationDetails location)
        {
            return string.Format("{0} {1} / {2} - {3}",
                EmojiIconMap.Resolve(d.Icon),
                ConvertToTempString(d.High, location),
                ConvertToTempString(d.Low, location),
                d.Summary);
        }

        private string GetTimeZoneCode(string timezone)
        {
            try
            {
                var tzCode = TZNames.GetAbbreviationsForTimeZone(timezone, "en-US");
                return tzCode?.Generic;
            }
            catch(Exception)
            {
                return null;
            }
        }
    }
}
