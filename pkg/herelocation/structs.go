package herelocation

type Coordinates struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

type GeoLocation struct {
	Coordinates Coordinates `json:"Coordinates"`
	Country     string      `json:"Country"`
	Region      string      `json:"Region"`
	City        string      `json:"City"`
}

type hereResponse struct {
	Response searchResponseType `json:"Response"`
}

type searchResponseType struct {
	View []searchResultsViewType `json:"View"`
}

type searchResultsViewType struct {
	Result []searchResultType `json:"Result"`
}

type searchResultType struct {
	Relevance float32 `json:"Relevance"`
	//MatchLevel can be country, state, county, city, district, street, intersection, houseNumber, postalCode, landmark
	MatchLevel   string                   `json:"MatchLevel"`
	MatchQuality locationMatchQualityType `json:"MatchQuality"`
	Location     locationType             `json:"Location"`
}

type locationMatchQualityType struct {
	Country     float32 `json:"Country"`
	State       float32 `json:"State"`
	County      float32 `json:"County"`
	City        float32 `json:"City"`
	District    float32 `json:"District"`
	Subdistrict float32 `json:"Subdistrict"`
	Street      float32 `json:"Street"`
	HouseNumber float32 `json:"HouseNumber"`
	PostalCode  float32 `json:"PostalCode"`
	Building    float32 `json:"Building"`
}

type locationType struct {
	DisplayPosition displayPositionType `json:"DisplayPosition"`
	Address         addressType         `json:"Address"`
}

type displayPositionType struct {
	Latitude  float64 `json:"Latitude"`
	Longitude float64 `json:"Longitude"`
}

type addressType struct {
	Label          string                   `json:"Label"`
	Country        string                   `json:"Country"`
	State          string                   `json:"State"`
	City           string                   `json:"City"`
	District       string                   `json:"District"`
	AdditionalData []additionalDataKeyValue `json:"AdditionalData"`
}

type additionalDataKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
