namespace Weatherman.Bot.Utils
{
    internal static class WindChillCalculator
    {
        private const double wcc1 = 35.74;
        private const double wcc2 = 0.6215;
        private const double wcc3 = 33.75;
        private const double wcc4 = 0.4275;

        public static double Calculate(double temperature, double windSpeed)
        {
            var ws = Math.Pow(windSpeed, 0.16);
            return wcc1 + (wcc2 * temperature) - (wcc3 * ws) + (wcc4 * temperature * ws);
        }
    }
}
