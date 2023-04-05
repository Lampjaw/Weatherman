using System.Drawing;

namespace Weatherman.Bot
{
    public class BotConfiguration
    {
        public string DiscordToken { get; set; }
        public string DiscordClientId { get; set; }
        public ulong? GuildId { get; set; }
        public string HereApiKey { get; set; }
        public string PirateWeatherKey { get; set; }
        public string RedisAddress { get; set; }
    }
}
