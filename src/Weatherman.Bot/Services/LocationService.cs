using Geo.Here.Abstractions;
using Geo.Here.Models.Parameters;
using Geo.Here.Models.Responses;
using Microsoft.Extensions.Logging;
using Weatherman.Bot.Cache;
using Weatherman.Bot.Models;

namespace Weatherman.Bot.Services
{
    public class LocationService
    {
        private readonly IHereGeocoding _hereGeocoding;
        private readonly ICache _cache;
        private readonly ILogger<LocationService> _logger;

        private readonly TimeSpan _geocodeCacheExpiration = TimeSpan.FromHours(1);

        public LocationService(IHereGeocoding hereGeocoding, ICache cache, ILogger<LocationService> logger)
        {
            _hereGeocoding = hereGeocoding;
            _cache = cache;
            _logger = logger;
        }

        public async Task<LocationDetails> GetGeocodeForLocationStringAsync(string locationQuery)
        {
            var cacheKey = $"geocode-{locationQuery}";

            var details = await _cache.GetAsync<LocationDetails>(cacheKey);
            if (details != null)
            {
                return details;
            }

            _logger.LogInformation("Fetching location for '{0}'", locationQuery);

            var geocodeResponse = await _hereGeocoding.GeocodingAsync(new GeocodeParameters { Query = locationQuery });

            var location = geocodeResponse.Items.OrderByDescending(a => a, new GeocodeComparer()).ToList().First();

            details = new LocationDetails
            {
                Latitude = location.Position.Latitude,
                Longitude = location.Position.Longitude,
                Country = location.Address.CountryName,
                Region = location.Address.State,
                City = location.Address.City
            };

            await _cache.SetAsync(cacheKey, details, _geocodeCacheExpiration);

            return details;
        }

        private class GeocodeComparer : IComparer<GeocodeLocation>
        {
            public int Compare(GeocodeLocation x, GeocodeLocation y)
            {
                if (x.Scoring.QueryScore == y.Scoring.QueryScore && x.Address.CountryCode != y.Address.CountryCode && x.Address.CountryCode == "USA")
                {
                    return 1;
                }

                return x.Scoring.QueryScore.CompareTo(y.Scoring.QueryScore);
            }
        }
    }
}