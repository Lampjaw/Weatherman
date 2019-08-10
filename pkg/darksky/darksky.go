package darksky

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const WEATHER_URL string = "https://api.darksky.net"

type DarkSkyClient struct {
	secretKey string
}

func NewClient(secretKey string) *DarkSkyClient {
	return &DarkSkyClient{
		secretKey: secretKey,
	}
}

func (c *DarkSkyClient) GetCurrentWeather(latitude float64, longitude float64) (*DarkSkyResponse, error) {
	uri := fmt.Sprintf("%s/forecast/%s/%f,%f", WEATHER_URL, c.secretKey, latitude, longitude)

	return handleDarkSkyRequest(uri)
}

func (c *DarkSkyClient) GetForecastWeather(latitude float64, longitude float64) (*DarkSkyResponse, error) {
	uri := fmt.Sprintf("%s/forecast/%s/%f,%f?exclude=currently,flags,hourly,minutely", WEATHER_URL, c.secretKey, latitude, longitude)

	return handleDarkSkyRequest(uri)
}

func handleDarkSkyRequest(uri string) (*DarkSkyResponse, error) {
	resp, err := http.Get(uri)

	if err != nil {
		log.Printf("DarkSky request failed: %s", err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Request to get %s returned unexpected status code %v", uri, resp.StatusCode)
		log.Printf(msg)
		return nil, errors.New(msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var darkSkyResponse DarkSkyResponse
	err = json.Unmarshal(body, &darkSkyResponse)

	return &darkSkyResponse, err
}
