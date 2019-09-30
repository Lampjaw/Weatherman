package weatherplugin

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/lampjaw/weatherman/pkg/darksky"
	"github.com/lampjaw/weatherman/pkg/herelocation"
)

type weatherManager struct {
	repository         *repository
	herelocationClient *herelocation.HereLocationClient
	darkSkyClient      *darksky.DarkSkyClient
	cacheManager       *cacheManager
}

func newWeatherManager(config WeatherConfig) *weatherManager {
	manager := &weatherManager{
		repository:         newRepository(),
		herelocationClient: herelocation.NewClient(config.HereAppID, config.HereAppCode),
		darkSkyClient:      darksky.NewClient(config.DarkSkySecretKey),
		cacheManager:       newCacheManager(config),
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

func (l *weatherManager) getCurrentWeatherByLocation(userID string, locationQuery string) (*CurrentWeather, *herelocation.GeoLocation, error) {
	geoLocation, err := l.resolveLocationForUser(userID, locationQuery)

	if err != nil {
		return nil, nil, err
	}

	currentWeather := l.cacheManager.getCurrentWeatherResult(geoLocation)

	if currentWeather != nil {
		return currentWeather, geoLocation, nil
	}

	weatherResult, err := l.darkSkyClient.GetCurrentWeather(geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude)

	if err != nil {
		log.Printf("Failed to get current weather for '%f' lat '%f' long: %s", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude, err)
		return nil, nil, errors.New("Failed to resolve weather data for this location.")
	}

	currentWeather = convertCurrentDarkSkyResponse(weatherResult)

	go l.cacheManager.setCurrentWeatherResult(geoLocation, currentWeather)

	return currentWeather, geoLocation, nil
}

func (l *weatherManager) getForecastWeatherByLocation(userID string, locationQuery string) ([]*WeatherDay, *herelocation.GeoLocation, error) {
	geoLocation, err := l.resolveLocationForUser(userID, locationQuery)

	if err != nil {
		return nil, nil, err
	}

	forecastWeather := l.cacheManager.getForecastWeatherResult(geoLocation)

	if forecastWeather != nil {
		return forecastWeather, geoLocation, nil
	}

	weatherResult, err := l.darkSkyClient.GetForecastWeather(geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude)

	if err != nil {
		log.Printf("Failed to get forecast weather for '%f' lat '%f' long: %s", geoLocation.Coordinates.Latitude, geoLocation.Coordinates.Longitude, err)
		return nil, nil, errors.New("Failed to resolve weather data for this location.")
	}

	forecastWeather = convertForecastDarkSkyResponse(weatherResult)

	go l.cacheManager.setForecastWeatherResult(geoLocation, forecastWeather)

	return forecastWeather, geoLocation, nil
}

func (l *weatherManager) getStoredUserLocation(userID string) (*herelocation.GeoLocation, error) {
	userProfile, err := l.repository.getUserProfile(userID)

	if err != nil {
		log.Printf("Failed to get user profile '%s': %s", userID, err)
		return nil, errors.New("There was an issue retrieving your location profile.")
	}

	var geoLocation *herelocation.GeoLocation

	if userProfile == nil || (userProfile.HomeLocation == nil && userProfile.LastLocation == nil) {
		return nil, nil
	} else if userProfile.HomeLocation != nil {
		err = json.Unmarshal([]byte(*userProfile.HomeLocation), &geoLocation)
	} else if userProfile.LastLocation != nil {
		err = json.Unmarshal([]byte(*userProfile.LastLocation), &geoLocation)
	}

	if err != nil {
		log.Printf("Failed to resolve location profile '%s': %s", userID, err)
		return nil, errors.New("There was an issue resolving your location profile.")
	}

	return geoLocation, nil
}

func (l *weatherManager) updateUserLastLocation(userID string, geoLocation *herelocation.GeoLocation) error {
	locationBytes, err := json.Marshal(geoLocation)

	if err != nil {
		log.Printf("Failed to marshal geolocation for user '%s': %s", userID, err)
		return err
	}

	err = l.repository.updateUserLastLocation(userID, string(locationBytes))

	if err != nil {
		log.Printf("Failed to update last location for user '%s': %s", userID, err)
	}

	return err
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
			return nil, errors.New("No home or previous search history found. Please use `sethome <location>` to set your home.")
		}

		return geoLocation, nil
	}

	geoLocation, err := l.getLocation(locationQuery)

	if err != nil {
		return nil, err
	}

	go l.updateUserLastLocation(userID, geoLocation)

	return geoLocation, nil
}

func convertCurrentDarkSkyResponse(resp *darksky.DarkSkyResponse) *CurrentWeather {
	alerts := convertDarkSkyAlerts(resp.Alerts, resp.Timezone)

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
		ForecastHigh:              currentDay.TemperatureHigh,
		ForecastLow:               currentDay.TemperatureLow,
		HeatIndex:                 heatIndex,
		Icon:                      resp.Currently.Icon,
		UVIndex:                   currentDay.UVIndex,
		PrecipitationProbability:  currentDay.PrecipitationProbability * 100,
		PrecipitationType:         currentDay.PrecipitationType,
		PrecipitationIntensity:    currentDay.PrecipitationIntensity,
		PrecipitationIntensityMax: currentDay.PrecipitationIntensityMax,
		SnowAccumulation:          currentDay.SnowAccumulation,
		Alerts:                    alerts,
	}
}

func convertForecastDarkSkyResponse(resp *darksky.DarkSkyResponse) []*WeatherDay {
	result := make([]*WeatherDay, 0)

	locale := getTimeLocale(resp.Timezone)

	localeNow := time.Now().In(locale)

	for _, day := range resp.Daily.Data {
		date := time.Unix(day.Time, 0).In(locale)

		// if localeNow.Day() > date.Day() || (localeNow.Day() < date.Day() && localeNow.Month() > date.Month()) {
		// 	continue
		//}

		weatherDay := &WeatherDay{
			Date: date,
			High: day.TemperatureHigh,
			Low:  day.TemperatureLow,
			Text: day.Summary,
			Icon: day.Icon,
		}
		result = append(result, weatherDay)
	}

	return result
}

func convertDarkSkyAlerts(alerts []darksky.DarkSkyAlert, tz string) []CurrentWeatherAlert {
	sort.Slice(alerts, func(i int, j int) bool {
		return alerts[i].Expires < alerts[j].Expires
	})

	locale := getTimeLocale(tz)

	currentAlerts := make([]CurrentWeatherAlert, 0)

alertLoop:
	for i, alert := range alerts {
		for j := i + 1; j < len(alerts); j++ {
			if alert.Uri == alerts[j].Uri && alert.Expires < alerts[j].Expires {
				continue alertLoop
			}
		}

		issuedDate := time.Unix(alert.Time, 0).In(locale)
		expirationDate := time.Unix(alert.Expires, 0).In(locale)

		currentAlerts = append(currentAlerts, CurrentWeatherAlert{
			IssuedDate:     issuedDate,
			ExpirationDate: expirationDate,
			Title:          alert.Title,
			Description:    alert.Description,
			URI:            alert.Uri,
		})
	}

	return currentAlerts
}

func getTimeLocale(tz string) *time.Location {
	timeLocale, err := time.LoadLocation(tz)

	if err != nil {
		log.Println(err)
		return time.UTC
	}

	return timeLocale
}
