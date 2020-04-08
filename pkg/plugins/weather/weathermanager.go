package weatherplugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"weatherman/pkg/herelocation"

	forecast "github.com/mlbright/darksky/v2"
)

type weatherManager struct {
	repository         *repository
	herelocationClient *herelocation.HereLocationClient
	cacheManager       *cacheManager
	darkskyKey         string
}

func newWeatherManager(config WeatherConfig) *weatherManager {
	manager := &weatherManager{
		repository:         newRepository(),
		herelocationClient: herelocation.NewClient(config.HereAppID, config.HereAppCode),
		cacheManager:       newCacheManager(config),
		darkskyKey:         config.DarkSkySecretKey,
	}

	manager.repository.initRepository()

	return manager
}

func (l *weatherManager) setUserHomeLocation(userID string, locationQuery string) error {
	geoLocation, err := l.getLocation(locationQuery)

	locationBytes, err := json.Marshal(geoLocation)

	if err != nil {
		log.Printf("Failed to marshal geolocation for user '%s': %s", userID, err)
		return errors.New("Failed to prepare this home location.")
	}

	err = l.repository.updateUserHomeLocation(userID, string(locationBytes))

	if err != nil {
		log.Printf("Failed to update home location for user '%s': %s", userID, err)
		return errors.New("Failed to save this home location.")
	}

	return nil
}

func (l *weatherManager) deleteUserHomeLocation(userID string) error {
	err := l.repository.deleteUser(userID)
	if err != nil {
		log.Printf("Failed to delete user '%s': %s", userID, err)
		return errors.New("Failed to delete user data. Please try again later.")
	}

	return nil
}

func (l *weatherManager) getCurrentWeatherByLocation(userID string, locationQuery string) (*CurrentWeather, *herelocation.GeoLocation, error) {
	geoLocation, err := l.resolveLocationForUser(userID, locationQuery)

	if err != nil {
		return nil, nil, err
	}

	weatherResult, err := l.getCurrentWeather(geoLocation)

	if err != nil {
		log.Printf("Failed to get current weather for '%f' lat '%f' long: %s", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude, err)
		return nil, nil, errors.New("Failed to resolve weather data for this location.")
	}

	return weatherResult, geoLocation, nil
}

func (l *weatherManager) getForecastWeatherByLocation(userID string, locationQuery string) ([]*WeatherDay, *herelocation.GeoLocation, error) {
	geoLocation, err := l.resolveLocationForUser(userID, locationQuery)

	if err != nil {
		return nil, nil, err
	}

	weatherResult, err := l.getForecastWeather(geoLocation)

	if err != nil {
		log.Printf("Failed to get forecast weather for '%f' lat '%f' long: %s", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude, err)
		return nil, nil, errors.New("Failed to resolve weather data for this location.")
	}

	return weatherResult, geoLocation, nil
}

func (l *weatherManager) getStoredUserLocation(userID string) (*herelocation.GeoLocation, error) {
	userProfile, err := l.repository.getUserProfile(userID)

	if err != nil {
		log.Printf("Failed to get user profile '%s': %s", userID, err)
		return nil, errors.New("There was an issue retrieving your location profile.")
	}

	var geoLocation *herelocation.GeoLocation

	if userProfile == nil || userProfile.HomeLocation == nil {
		return nil, nil
	}

	err = json.Unmarshal([]byte(*userProfile.HomeLocation), &geoLocation)

	if err != nil {
		log.Printf("Failed to resolve location profile '%s': %s", userID, err)
		return nil, errors.New("There was an issue resolving your location profile.")
	}

	return geoLocation, nil
}

func (l *weatherManager) getLocation(locationQuery string) (*herelocation.GeoLocation, error) {
	geoLocation := l.cacheManager.getLocationResult(locationQuery)

	if geoLocation != nil {
		return geoLocation, nil
	}

	geoLocation, err := l.herelocationClient.GetLocationByText(locationQuery)

	if err != nil {
		log.Printf("Failed to resolve location '%s': %s", locationQuery, err)
		return nil, errors.New("Failed to resolve this location.")
	}

	if geoLocation == nil {
		return nil, errors.New("Failed to find this location.")
	}

	go l.cacheManager.setLocationResult(locationQuery, geoLocation)

	return geoLocation, nil
}

func (l *weatherManager) resolveLocationForUser(userID string, locationQuery string) (*herelocation.GeoLocation, error) {
	if locationQuery == "" {
		geoLocation, err := l.getStoredUserLocation(userID)

		if err != nil {
			return nil, err
		}

		if geoLocation == nil {
			return nil, errors.New("Please include a location or set a home. To set a home use `sethome <location>`.")
		}

		return geoLocation, nil
	}

	geoLocation, err := l.getLocation(locationQuery)

	if err != nil {
		return nil, err
	}

	return geoLocation, nil
}

func (l *weatherManager) getCurrentWeather(geoLocation *herelocation.GeoLocation) (*CurrentWeather, error) {
	weatherResult, err := l.getCurrentForecast(geoLocation)

	if err != nil {
		return nil, err
	}

	currentWeather := convertCurrentDarkSkyResponse(weatherResult)
	return currentWeather, nil
}

func (l *weatherManager) getForecastWeather(geoLocation *herelocation.GeoLocation) ([]*WeatherDay, error) {
	weatherResult, err := l.getCurrentForecast(geoLocation)

	if err != nil {
		return nil, err
	}

	forecastWeather := convertForecastDarkSkyResponse(weatherResult)
	return forecastWeather, nil
}

func (l *weatherManager) getCurrentForecast(geoLocation *herelocation.GeoLocation) (*forecast.Forecast, error) {
	weatherResult := l.cacheManager.getWeatherResult(geoLocation)

	if weatherResult != nil {
		return weatherResult, nil
	}

	sLat := fmt.Sprintf("%f", geoLocation.Coordinates.Latitude)
	sLong := fmt.Sprintf("%f", geoLocation.Coordinates.Longitude)
	weatherResult, err := forecast.Get(l.darkskyKey, sLat, sLong, "now", forecast.US, forecast.English)

	if err == nil {
		go l.cacheManager.setWeatherResult(geoLocation, weatherResult)
	}

	return weatherResult, err
}

func convertCurrentDarkSkyResponse(resp *forecast.Forecast) *CurrentWeather {
	alerts := convertDarkSkyAlerts(resp, resp.Timezone)

	currentDay := resp.Daily.Data[0]

	temp := resp.Currently.Temperature
	humidity := resp.Currently.Humidity * 100
	windSpeed := resp.Currently.WindSpeed
	heatIndex := calculateHeatIndex(temp, humidity)
	windChill := calculateWindChill(temp, windSpeed)

	return &CurrentWeather{
		Condition:                 resp.Currently.Summary,
		Temperature:               temp,
		Humidity:                  humidity,
		WindChill:                 windChill,
		WindSpeed:                 windSpeed,
		WindGust:                  currentDay.WindGust,
		ForecastHigh:              currentDay.TemperatureMax,
		ForecastLow:               currentDay.TemperatureMin,
		HeatIndex:                 heatIndex,
		Icon:                      resp.Currently.Icon,
		UVIndex:                   currentDay.UVIndex,
		PrecipitationProbability:  currentDay.PrecipProbability * 100,
		PrecipitationType:         currentDay.PrecipType,
		PrecipitationIntensity:    currentDay.PrecipIntensity,
		PrecipitationIntensityMax: currentDay.PrecipIntensityMax,
		SnowAccumulation:          currentDay.PrecipAccumulation,
		Alerts:                    alerts,
	}
}

func convertForecastDarkSkyResponse(resp *forecast.Forecast) []*WeatherDay {
	result := make([]*WeatherDay, 0)

	locale := getTimeLocale(resp.Timezone)

	for _, day := range resp.Daily.Data {
		date := time.Unix(day.Time, 0).In(locale)

		weatherDay := &WeatherDay{
			Date: date,
			High: day.TemperatureMax,
			Low:  day.TemperatureMin,
			Text: day.Summary,
			Icon: day.Icon,
		}
		result = append(result, weatherDay)
	}

	return result
}

func convertDarkSkyAlerts(forecast *forecast.Forecast, tz string) []CurrentWeatherAlert {
	alerts := forecast.Alerts

	sort.Slice(alerts, func(i int, j int) bool {
		return alerts[i].Expires < alerts[j].Expires
	})

	locale := getTimeLocale(tz)

	currentAlerts := make([]CurrentWeatherAlert, 0)

alertLoop:
	for i, alert := range alerts {
		for j := i + 1; j < len(alerts); j++ {
			if alert.URI == alerts[j].URI && alert.Expires < alerts[j].Expires {
				continue alertLoop
			}
		}

		issuedDate := time.Unix(alert.Time, 0).In(locale)
		expirationDate := time.Unix(int64(alert.Expires), 0).In(locale)

		currentAlerts = append(currentAlerts, CurrentWeatherAlert{
			IssuedDate:     issuedDate,
			ExpirationDate: expirationDate,
			Title:          alert.Title,
			Description:    alert.Description,
			URI:            alert.URI,
		})
	}

	return currentAlerts
}

func getTimeLocale(tz string) *time.Location {
	timeLocale, err := time.LoadLocation(tz)

	if err != nil {
		log.Printf("Error reading timezone: %s", err)
		return time.UTC
	}

	return timeLocale
}
