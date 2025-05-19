package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/theammir/genesis-test/internal/cache"
)

const baseURL = "https://api.weatherapi.com/v1"

type CurrentWeatherResponse struct {
	Location struct {
		Latitude   float64 `json:"lat"`
		Longtitude float64 `json:"lon"`
		Name       string  `json:"name"`
		Region     string  `json:"region"`
		Country    string  `json:"country"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch uint64  `json:"last_updated_epoch"`
		TemperatureC     float32 `json:"temp_c"`
		TemperatureF     float32 `json:"temp_f"`
		Humidity         uint8   `json:"humidity"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
	} `json:"current"`
}

type Client struct {
	APIKey     string
	httpClient *http.Client
	ttlCache   *cache.TTLCache[string, *CurrentWeatherResponse]
}

func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey, httpClient: &http.Client{}, ttlCache: cache.NewTTLCache[string, *CurrentWeatherResponse](15 * time.Minute)}
}

func (c *Client) GetCurrentWeather(ctx context.Context, locationQuery string) (*CurrentWeatherResponse, error) {
	var locationLower = strings.ToLower(locationQuery)
	weather, ok := c.ttlCache.Get(locationLower)
	if !ok {
		var err error
		weather, err = c.fetchCurrentWeather(ctx, locationQuery)
		if err != nil {
			return nil, err
		}
		c.ttlCache.Set(locationLower, weather)
	} else {
		log.Printf("Reusing from cache for %s", locationQuery)
	}
	return weather, nil
}

func (c *Client) fetchCurrentWeather(ctx context.Context, locationQuery string) (*CurrentWeatherResponse, error) {
	endpoint := baseURL + "/current.json"
	requestURL, _ := url.Parse(endpoint)

	requestQuery := requestURL.Query()
	requestQuery.Set("key", c.APIKey)
	requestQuery.Set("q", locationQuery)
	requestURL.RawQuery = requestQuery.Encode()

	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-OK response status code %d", response.StatusCode)
	}

	var result CurrentWeatherResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("couldn't decode json response: %w", err)
	}

	return &result, nil
}
