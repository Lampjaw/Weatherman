using DarkSky.Models;
using DarkSky.Services;
using Microsoft.Extensions.Logging;
using Weatherman.Bot.Cache;
using Weatherman.Bot.Models;
using Weatherman.Bot.Utils;

namespace Weatherman.Bot.Services
{
    public class WeatherService
    {
        private readonly DarkSkyService _darkSky;
        private readonly ICache _cache;
        private readonly ILogger<WeatherService> _logger;

        private readonly TimeSpan _forecastCacheExpiration = TimeSpan.FromMinutes(10);

        public WeatherService(DarkSkyService darkSkyService, ICache cache, ILogger<WeatherService> logger)
        {
            _darkSky = darkSkyService;
            _cache = cache;
            _logger = logger;
        }

        public async Task<WeatherForecast> GetCurrentForecastAsync(Coordinates coordinates)
        {
            var forecast = await GetForecastAsync(coordinates.Latitude, coordinates.Longitude);
            if (forecast == null)
            {
                return null;
            }

            var alerts = ConvertAlerts(forecast.Alerts);

            var temp = forecast.Currently.Temperature.GetValueOrDefault();
            var humidity = forecast.Currently.Humidity.GetValueOrDefault() * 100;
            var windSpeed = forecast.Currently.WindSpeed.GetValueOrDefault();
            var heatIndex = HeatIndexCalculator.Calculate(temp, humidity);
            var windChill = WindChillCalculator.Calculate(temp, windSpeed);

            var currentDay = forecast.Daily.Data[0];

            return new WeatherForecast
            {
                TimeZone = forecast.TimeZone,
                Condition = forecast.Currently.Summary,
                Temperature = temp,
                Humidity = humidity,
                WindChill = windChill,
                WindSpeed = windSpeed,
                WindGust = currentDay.WindGust.GetValueOrDefault(),
                ForecastHigh = currentDay.TemperatureHigh.GetValueOrDefault(),
                ForecastLow = currentDay.TemperatureLow.GetValueOrDefault(),
                HeatIndex = heatIndex,
                Icon = forecast.Currently.Icon,
                UVIndex = currentDay.UvIndex.GetValueOrDefault(),
                PrecipitationProbability = currentDay.PrecipProbability.GetValueOrDefault(),
                PrecipitationType = currentDay.PrecipType,
                PrecipitationIntensity = currentDay.PrecipIntensity.GetValueOrDefault(),
                PrecipitationIntensityMax = currentDay.PrecipIntensityMax,
                SnowAccumulation = currentDay.PrecipAccumulation.GetValueOrDefault(),
                Alerts = alerts
            };
        }

        public async Task<IEnumerable<WeatherSummary>> GetWeeklyForecastAsync(Coordinates coordinates)
        {
            var forecast = await GetForecastAsync(coordinates.Latitude, coordinates.Longitude);
            if (forecast == null)
            {
                return null;
            }

            return forecast.Daily.Data.Select(a =>
            {
                return new WeatherSummary
                {
                    TimeZone = forecast.TimeZone,
                    Date = a.DateTime,
                    High = a.TemperatureHigh.GetValueOrDefault(),
                    Low = a.TemperatureLow.GetValueOrDefault(),
                    Icon = a.Icon,
                    Summary = a.Summary
                };
            });
        }

        private async Task<Forecast> GetForecastAsync(double latitude, double longitude)
        {
            var cacheKey = $"forecastv2-{latitude}-{longitude}";

            var forecast = await _cache.GetAsync<Forecast>(cacheKey);
            if (forecast != null)
            {
                return forecast;
            }

            _logger.LogInformation("Fetching forecast for {0}, {1}", latitude, longitude);

            var result = await _darkSky.GetForecast(latitude, longitude, new OptionalParameters
            {
                MeasurementUnits = "us",
                DataBlocksToExclude = new() { ExclusionBlocks.Hourly, ExclusionBlocks.Minutely, ExclusionBlocks.Flags }
            });

            if (result?.IsSuccessStatus != true)
            {
                _logger.LogError("Failed to fetch forecast for '{0}, {1}': {2}.", latitude, longitude, result?.ResponseReasonPhrase);
                return null;
            }

            if (result.Response != null)
            {
                await _cache.SetAsync(cacheKey, result.Response, _forecastCacheExpiration);

                return result.Response;
            }

            return null;
        }

        private IEnumerable<WeatherAlert> ConvertAlerts(IEnumerable<Alert> alerts)
        {
            if (alerts == null)
            {
                return new List<WeatherAlert>();
            }

            return alerts
                .OrderBy(a => a.ExpiresDateTime)
                .Where(a => !alerts.Any(b => a.Uri == b.Uri && a.ExpiresDateTime < b.ExpiresDateTime))
                .Select(a =>
                {
                    return new WeatherAlert
                    {
                        IssuedDate = a.DateTime,
                        ExpirationDate = a.ExpiresDateTime,
                        Title = a.Title,
                        Description = a.Description,
                        Uri = a.Uri
                    };
                });
        }
    }
}
