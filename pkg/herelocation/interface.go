package herelocation

func newGeoLocation(lat float64, long float64, country string, region string, city string) *GeoLocation {
	return &GeoLocation{
		Coordinates: Coordinates{
			Latitude:  lat,
			Longitude: long,
		},
		Country: country,
		Region:  region,
		City:    city,
	}
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type GeoLocation struct {
	Coordinates Coordinates
	Country     string
	Region      string
	City        string
}

type hereResponse struct {
	Response hereResponseModel `json:"Response"`
}

type hereResponseModel struct {
	View []hereResponseModelView `json:"View"`
}

type hereResponseModelView struct {
	Result []hereResponseViewResult `json:"Result"`
}

type hereResponseViewResult struct {
	Location hereResponseViewResultLocation `json:"Location"`
}

type hereResponseViewResultLocation struct {
	DisplayPosition hereResponseViewResultLocationDisplayPosition `json:"DisplayPosition"`
	Address         hereResponseViewResultLocationDisplayAddress  `json:"Address"`
}

type hereResponseViewResultLocationDisplayPosition struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

type hereResponseViewResultLocationDisplayAddress struct {
	Label          string              `json:"Label"`
	Country        string              `json:"Country"`
	State          string              `json:"State"`
	City           string              `json:"City"`
	District       string              `json:"District"`
	AdditionalData []map[string]string `json:"AdditionalData"`
}
