namespace Weatherman.Bot
{
    internal static class Constants
    {
        public const string PirateWeatherApi = "https://api.pirateweather.net/";
        public const string FooterPoweredByText = "Powered by Pirate Weather";
        public const string TitleSeeMoreText = "See more at merrysky.com";
        public const string TitleSeeMoreUrlFormat = "https://merrysky.net/forecast/{0},{1}";
        public const int MaxForecastDays = 7;
        public const int MaxForecastHours = 6;
        public const uint DefaultEmbedColor = 0x070707;
    }
}
