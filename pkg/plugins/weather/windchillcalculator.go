package weatherplugin

import "math"

const wcc1 float64 = 35.74
const wcc2 float64 = 0.6215
const wcc3 float64 = 33.75
const wcc4 float64 = 0.4275

func calculateWindChill(temperature float64, windSpeed float64) float64 {
	ws := math.Pow(windSpeed, 0.16)
	return wcc1 + (wcc2 * temperature) - (wcc3 * ws) + (wcc4 * temperature * ws)
}
