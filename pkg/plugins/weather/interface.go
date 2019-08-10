package weatherplugin

import "github.com/lampjaw/weatherman/pkg/herelocation"

type WeatherConfig struct {
	HereAppID        string
	HereAppCode      string
	DarkSkySecretKey string
}

type CurrentWeather struct {
	Temperature  float64
	Condition    string
	Humidity     float64
	WindChill    float64
	WindSpeed    float64
	ForecastHigh float64
	ForecastLow  float64
	HeatIndex    float64
	Icon         string
	Location     herelocation.GeoLocation
}

type WeatherDay struct {
	Date string
	Day  string
	High float64
	Low  float64
	Text string
	Icon string
}
