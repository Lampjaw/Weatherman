using DarkSky.Models;

namespace Weatherman.Bot.Models
{
    public class ForecastDay
    {
        public DateTimeOffset Date { get; set; }
        public double High { get; set; }
        public double Low { get; set; }
        public string Summary { get; set; }
        public Icon Icon { get; set; }
    }
}
