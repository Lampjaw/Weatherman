package weatherplugin

import "math"

const hic1 float64 = -42.379
const hic2 float64 = 2.04901523
const hic3 float64 = 10.14333127
const hic4 float64 = -0.22475541
const hic5 float64 = -0.00683783
const hic6 float64 = -0.05481717
const hic7 float64 = 0.00122874
const hic8 float64 = 0.00085282
const hic9 float64 = -0.00000199

func calculateHeatIndex(temperature float64, humidity float64) float64 {
	t := temperature
	r := humidity

	heatIndex := 0.5 * (t + 61.0 + ((t - 68.0) * 1.2) + (r * 0.094))

	if heatIndex < 80 {
		return heatIndex
	}

	heatIndex =
		hic1 +
			hic2*t +
			hic3*r +
			hic4*t*r +
			hic5*t*t +
			hic6*r*r +
			hic7*t*t*r +
			hic8*t*r*r +
			hic9*t*t*r*r

	if r < 13 && t >= 80 && t <= 112 {
		return heatIndex - ((13-r)/4)*math.Sqrt((17-math.Abs(t-95))/17)
	}

	if r > 85 && t >= 80 && t <= 87 {
		return heatIndex + ((r-85)/10)*((87-t)/5)
	}

	return heatIndex
}
