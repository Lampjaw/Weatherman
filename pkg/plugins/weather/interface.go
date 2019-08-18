package weatherplugin

import (
	"github.com/lampjaw/weatherman/pkg/darksky"
	"github.com/lampjaw/weatherman/pkg/herelocation"
)

type WeatherConfig struct {
	HereAppID        string
	HereAppCode      string
	DarkSkySecretKey string
	RedisAddress     string
	RedisPassword    string
	RedisDatabase    int
}

type CurrentWeather struct {
	Temperature               float64                  `json:"Temperature"`
	Condition                 string                   `json:"Condition"`
	Humidity                  float64                  `json:"Humidity"`
	WindChill                 float64                  `json:"WindChill"`
	WindSpeed                 float64                  `json:"WindSpeed"`
	WindGust                  float64                  `json:"WindGust"`
	ForecastHigh              float64                  `json:"ForecastHigh"`
	ForecastLow               float64                  `json:"ForecastLow"`
	HeatIndex                 float64                  `json:"HeatIndex"`
	Icon                      string                   `json:"Icon"`
	UVIndex                   float64                  `json:"UVIndex"`
	PrecipitationProbability  float64                  `json:"PrecipitationProbability"`
	PrecipitationType         string                   `json:"PrecipitationType"`
	PrecipitationIntensity    float64                  `json:"PrecipitationIntensity"`
	PrecipitationIntensityMax float64                  `json:"PrecipitationIntensityMax"`
	SnowAccumulation          float64                  `json:"SnowAccumulation"`
	Alerts                    []darksky.DarkSkyAlert   `json:"Alerts"`
	Location                  herelocation.GeoLocation `json:"Location"`
}

type WeatherDay struct {
	Date string  `json:"Date"`
	Day  string  `json:"Day"`
	High float64 `json:"High"`
	Low  float64 `json:"Low"`
	Text string  `json:"Text"`
	Icon string  `json:"Icon"`
}
