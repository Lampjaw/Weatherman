using DarkSky.Services;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Weatherman.Bot
{
    internal class DarkSkyJsonSerializerService : IJsonSerializerService
    {
        JsonSerializerSettings _jsonSettings = new JsonSerializerSettings();

        public async Task<T> DeserializeJsonAsync<T>(Task<string> json)
        {
            try
            {
                var jsonString = await json;
                jsonString = FixJsonTypeValues(jsonString);

                return (jsonString != null)
                    ? JsonConvert.DeserializeObject<T>(jsonString, _jsonSettings)
                    : default;
            }
            catch (JsonReaderException e)
            {
                throw new FormatException("Json Parsing Error", e);
            }
        }

        private string FixJsonTypeValues(string json)
        {
            var jsonToken = JToken.Parse(json);

            var jobj = (JObject)jsonToken.SelectToken("currently");
            if (jobj != null)
            {
                jobj.Property("windBearing").Remove();
                jobj["uvIndex"] = (int)(jobj["uvIndex"].Value<double>());
            }

            var dailyObj = jsonToken.SelectTokens("daily.data[*]");
            if (dailyObj != null)
            {
                dailyObj.Cast<JObject>().ToList().ForEach(a =>
                {
                    a.Property("windBearing").Remove();
                    a["uvIndex"] = (int)(a["uvIndex"].Value<double>());
                });
            }

            return jsonToken.ToString();
        }
    }
}
