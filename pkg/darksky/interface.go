package darksky

type DarkSkyResponse struct {
	Timezone  string           `json:"timezone"`
	Currently DarkSkyCurrently `json:"currently"`
	Daily     DarkSkyDaily     `json:"daily"`
	Alerts    []DarkSkyAlert   `json:"alerts"`
}

type DarkSkyCurrently struct {
	Time                int64   `json:"time"`
	Summary             string  `json:"summary"`
	Icon                string  `json:"icon"`
	Temperature         float64 `json:"temperature"`
	ApparentTemperature float64 `json:"apparentTemperature"`
	Humidity            float64 `json:"humidity"`
	WindSpeed           float64 `json:"windSpeed"`
}

type DarkSkyDaily struct {
	Data []DarkSkyDailyData
}

type DarkSkyDailyData struct {
	Time            int64   `json:"time"`
	Summary         string  `json:"summary"`
	Icon            string  `json:"icon"`
	TemperatureLow  float64 `json:"temperatureLow"`
	TemperatureHigh float64 `json:"temperatureHigh"`
}

type DarkSkyAlert struct {
	Title       string `json:"title"`
	Time        int64  `json:"time"`
	Expires     int64  `json:"expires"`
	Description string `json:"description"`
	Uri         string `json:"uri"`
}
