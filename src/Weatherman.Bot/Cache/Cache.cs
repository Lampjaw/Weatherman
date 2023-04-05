using Microsoft.Extensions.Options;
using Newtonsoft.Json;

namespace Weatherman.Bot.Cache
{
    public class Cache : ICache
    {
        private readonly CacheConnector _connector;
        private readonly IOptions<CacheOptions> _options;

        public Cache(CacheConnector connector, IOptions<CacheOptions> options)
        {
            _connector = connector;
            _options = options;
        }

        public async Task SetAsync(string key, object value, TimeSpan expires)
        {
            var db = await _connector.ConnectAsync();

            try
            {
                var sValue = JsonConvert.SerializeObject(value);
                await db.StringSetAsync(KeyFormatter(key), sValue, expires);
            }
            catch (Exception)
            {
                // ignored
            }
        }

        public async Task<T> GetAsync<T>(string key)
        {
            var db = await _connector.ConnectAsync();

            try
            {
                var value = await db.StringGetAsync(KeyFormatter(key));

                return JsonConvert.DeserializeObject<T>(value);
            }
            catch (Exception)
            {
                return default;
            }
        }

        public async Task RemoveAsync(string key)
        {
            var db = await _connector.ConnectAsync();

            try
            {
                await db.KeyDeleteAsync(KeyFormatter(key));
            }
            catch (Exception)
            {
                // ignored
            }
        }

        private string KeyFormatter(string key)
        {
            if (!string.IsNullOrWhiteSpace(_options.Value.KeyPrefix))
            {
                return $"{_options.Value.KeyPrefix}_{key}";
            }

            return key;
        }
    }
}
