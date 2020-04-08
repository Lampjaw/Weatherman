package weatherplugin

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"weatherman/pkg/herelocation"

	"github.com/go-redis/redis"
	forecast "github.com/mlbright/darksky/v2"
)

const locationCachePrefix string = "location"
const weatherCachePrefix string = "weather"

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

func (m *cacheManager) setWeatherResult(geolocation *herelocation.GeoLocation, obj *forecast.Forecast) {
	cacheStoreWeather(m, weatherCachePrefix, geolocation, obj)
}

func (m *cacheManager) getLocationResult(queryText string) *herelocation.GeoLocation {
	var result *herelocation.GeoLocation

	fetchStore(m, locationCachePrefix, strings.ToLower(queryText), &result)

	return result
}

func (m *cacheManager) getWeatherResult(geolocation *herelocation.GeoLocation) *forecast.Forecast {
	var result *forecast.Forecast

	cacheGetWeather(m, weatherCachePrefix, geolocation, &result)

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
