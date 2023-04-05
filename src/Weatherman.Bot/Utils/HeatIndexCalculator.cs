namespace Weatherman.Bot.Utils
{
    internal static class HeatIndexCalculator
    {
        private const double hic1 = -42.379;
        private const double hic2 = 2.04901523;
        private const double hic3 = 10.14333127;
        private const double hic4 = -0.22475541;
        private const double hic5 = -0.00683783;
        private const double hic6 = -0.05481717;
        private const double hic7 = 0.00122874;
        private const double hic8 = 0.00085282;
        private const double hic9 = -0.00000199;

        public static double Calculate(double temperature, double humidity)
        {
            var t = temperature;
            var r = humidity;

            var heatIndex = 0.5 * (t + 61.0 + ((t - 68.0) * 1.2) + (r * 0.094));

            if (heatIndex < 80)
            {
                return heatIndex;
            }

            heatIndex =
                hic1 +
                hic2 * t +
                hic3 * r +
                hic4 * t * r +
                hic5 * t * t +
                hic6 * r * r +
                hic7 * t * t * r +
                hic8 * t * r * r +
                hic9 * t * t * r * r;

            if (r < 13 && t >= 80 && t <= 112)
            {
                return heatIndex - ((13 - r) / 4) * Math.Sqrt((17 - Math.Abs(t - 95)) / 17);
            }

            if (r > 85 && t >= 80 && t <= 87)
            {
                return heatIndex + ((r - 85) / 10) * ((87 - t) / 5);
            }

            return heatIndex;
        }
    }
}