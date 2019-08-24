package herelocation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
)

const HERE_URL string = "https://geocoder.api.here.com/6.2/geocode.json"

type HereLocationClient struct {
	appID   string
	appCode string
}

func NewClient(appID string, appCode string) *HereLocationClient {
	return &HereLocationClient{
		appID:   appID,
		appCode: appCode,
	}
}

//GetLocationByText Resolves a location from a search string
func (c *HereLocationClient) GetLocationByText(location string) (*GeoLocation, error) {
	hereResponse, err := getHereResponse(c.appID, c.appCode, location)

	if err != nil {
		return nil, err
	}

	foundLocation, err := getBestRelevantLocation(hereResponse)

	if err != nil {
		return nil, err
	}

	return &GeoLocation{
		Coordinates: Coordinates{
			Latitude:  foundLocation.DisplayPosition.Latitude,
			Longitude: foundLocation.DisplayPosition.Longitude,
		},
		Country: getCountryName(foundLocation),
		Region:  foundLocation.Address.State,
		City:    foundLocation.Address.City,
	}, nil
}

func getHereResponse(appID string, appCode string, locationText string) (*hereResponse, error) {
	uri := fmt.Sprintf("%s?app_id=%s&app_code=%s&searchtext=%s", HERE_URL, appID, appCode, url.QueryEscape(locationText))

	resp, err := http.Get(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to get %s: %s", uri, err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Request to get %s returned unexpected status code %v", uri, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var hereResponse hereResponse
	err = json.Unmarshal(body, &hereResponse)

	if err != nil {
		return nil, fmt.Errorf("Unable to unmarshal HERE data: %s", err)
	}

	return &hereResponse, nil
}

func getBestRelevantLocation(data *hereResponse) (*locationType, error) {
	var foundPlace *locationType

	if len(data.Response.View) > 0 && len(data.Response.View[0].Result) > 0 {
		results := data.Response.View[0].Result

		sort.Slice(results, func(i int, j int) bool {
			r1 := results[i]
			r2 := results[j]

			return r1.Relevance > r2.Relevance || (r1.Relevance == r1.Relevance && r1.Location.Address.Country != r2.Location.Address.Country && r1.Location.Address.Country == "USA")
		})

		foundPlace = &results[0].Location
	}

	if foundPlace == nil {
		return nil, errors.New("Location not found")
	}

	return foundPlace, nil
}

func getCountryName(location *locationType) string {
	if location.Address.AdditionalData != nil && len(location.Address.AdditionalData) > 0 {
		for _, kvp := range location.Address.AdditionalData {
			if kvp.Key == "CountryName" {
				return kvp.Value
			}
		}
	}

	return location.Address.Country
}
