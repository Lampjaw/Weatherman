package weatherplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/lampjaw/weatherman/pkg/herelocation"
)

const locationCachePrefix string = "location"
const currrentWeatherCachePrefix string = "currentWeather"
const forecastWeatherCachePrefix string = "forecastWeather"

type cacheManager struct {
	redisClient *redis.Client
}

func newCacheManager(config WeatherConfig) *cacheManager {
	redisOptions := &redis.Options{
		Addr:     config.RedisAddress,
		Password: config.RedisPassword,
		DB:       config.RedisDatabase,
	}

	return &cacheManager{
		redisClient: redis.NewClient(redisOptions),
	}
}

func (m *cacheManager) setLocationResult(queryText string, obj *herelocation.GeoLocation) {
	setStore(m, locationCachePrefix, strings.ToLower(queryText), obj, 30*time.Minute)
}

func (m *cacheManager) setCurrentWeatherResult(geolocation *herelocation.GeoLocation, obj *CurrentWeather) {
	cacheStoreWeather(m, currrentWeatherCachePrefix, geolocation, obj)
}

func (m *cacheManager) setForecastWeatherResult(geolocation *herelocation.GeoLocation, obj []*WeatherDay) {
	cacheStoreWeather(m, forecastWeatherCachePrefix, geolocation, obj)
}

func (m *cacheManager) getLocationResult(queryText string) *herelocation.GeoLocation {
	var result *herelocation.GeoLocation

	fetchStore(m, locationCachePrefix, strings.ToLower(queryText), &result)

	return result
}

func (m *cacheManager) getCurrentWeatherResult(geolocation *herelocation.GeoLocation) *CurrentWeather {
	var result *CurrentWeather

	cacheGetWeather(m, currrentWeatherCachePrefix, geolocation, &result)

	return result
}

func (m *cacheManager) getForecastWeatherResult(geolocation *herelocation.GeoLocation) []*WeatherDay {
	var result []*WeatherDay

	cacheGetWeather(m, forecastWeatherCachePrefix, geolocation, &result)

	return result
}

func getWeatherCacheKey(geolocation *herelocation.GeoLocation) string {
	return fmt.Sprintf("%f,%f", geolocation.Coordinates.Latitude, geolocation.Coordinates.Longitude)
}

func cacheStoreWeather(m *cacheManager, keyPrefix string, geolocation *herelocation.GeoLocation, obj interface{}) {
	cacheKey := getWeatherCacheKey(geolocation)
	setStore(m, keyPrefix, cacheKey, obj, 10*time.Minute)
}

func cacheGetWeather(m *cacheManager, keyPrefix string, geolocation *herelocation.GeoLocation, obj interface{}) error {
	cacheKey := getWeatherCacheKey(geolocation)
	return fetchStore(m, keyPrefix, cacheKey, &obj)
}

func setStore(m *cacheManager, keyPrefix string, key string, obj interface{}, expiration time.Duration) {
	cacheKey := fmt.Sprintf("%s:%s", keyPrefix, key)

	result, err := json.Marshal(obj)
	if err != nil {
		log.Printf("Failed to marshal data for cache '%s': %s", cacheKey, err)
		return
	}

	_, err = m.redisClient.Set(cacheKey, string(result), expiration).Result()
	if err != nil {
		log.Printf("Failed to set cache for '%s': %s", cacheKey, err)
		return
	}
}

func fetchStore(m *cacheManager, keyPrefix string, key string, storeResult interface{}) error {
	cacheKey := fmt.Sprintf("%s:%s", keyPrefix, key)

	cacheResult, err := m.redisClient.Get(cacheKey).Bytes()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		log.Printf("Failed to retrieve from cache '%s': %s", cacheKey, err)
		return err
	}

	err = json.Unmarshal(cacheResult, &storeResult)
	if err != nil {
		log.Printf("Failed to unmarshal from cache '%s': %s", cacheKey, err)
		return err
	}

	return nil
}
