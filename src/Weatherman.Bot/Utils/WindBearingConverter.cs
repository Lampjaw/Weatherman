namespace Weatherman.Bot.Utils
{
    public static class WindBearingConverter
    {
        public static string ConvertToWindDirection(double? bearing)
        {
            if (bearing == null)
            {
                return "";
            }

            switch ((int)(bearing / 11.25))
            {
                case 0 or 31:
                    return "N";
                case 1 or 2 :
                    return "NNE";
                case 3 or 4:
                    return "NE";
                case 5 or 6:
                    return "ENE";
                case 7 or 8:
                    return "E";
                case 9 or 10:
                    return "ESE";
                case 11 or 12:
                    return "SE";
                case 13 or 14:
                    return "SSE";
                case 15 or 16:
                    return "S";
                case 17 or 18:
                    return "SSW";
                case 19 or 20:
                    return "SW";
                case 21 or 22:
                    return "WSW";
                case 23 or 24:
                    return "W";
                case 25 or 26:
                    return "WNW";
                case 27 or 28:
                    return "NW";
                case 29 or 30:
                    return "NNW";
                default:
                    return "";
            }
        }
    }
}
