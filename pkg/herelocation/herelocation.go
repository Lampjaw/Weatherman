package herelocation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func (c *HereLocationClient) GetLocationByTextAsync(location string) (*GeoLocation, error) {
	uri := fmt.Sprintf("%s?app_id=%s&app_code=%s&searchtext=%s", HERE_URL, c.appID, c.appCode, url.QueryEscape(location))

	resp, err := http.Get(uri)

	if err != nil {
		log.Printf("Failed to get %s: %s", uri, err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Request to get %s returned unexpected status code %v", uri, resp.StatusCode)
		log.Printf(msg)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var hereResponse hereResponse
	err = json.Unmarshal(body, &hereResponse)

	if err != nil {
		msg := fmt.Sprintf("Unable to unmarshal Here data: %s", err)
		log.Printf(msg)
		return nil, errors.New(msg)
	}

	var foundPlace *hereResponseViewResultLocation

	if len(hereResponse.Response.View) > 0 && len(hereResponse.Response.View[0].Result) > 0 {
		foundPlace = &hereResponse.Response.View[0].Result[0].Location
	}

	if foundPlace == nil {
		return nil, errors.New("Location not found")
	}

	countryName := foundPlace.Address.Country

	fmt.Printf("%+v", foundPlace.Address.AdditionalData)

	//countryName := foundPlace.Address.AdditionalData["CountryName"]
	//if countryName == "" {
	//countryName = foundPlace.Address.Country
	//}

	return newGeoLocation(foundPlace.DisplayPosition.Latitude, foundPlace.DisplayPosition.Longitude, countryName, foundPlace.Address.State, foundPlace.Address.City), nil
}
