package api

import (
	"errors"

	"github.com/theammir/genesis-test/api/weather"
)

type Weather struct {
	Temperature float32 `json:"temperature"`
	Humidity    uint8   `json:"humidity"`
	Description string  `json:"description"`
}

func FromWeatherResponse(w *weather.CurrentWeatherResponse) Weather {
	return Weather{
		Temperature: w.Current.TemperatureC,
		Humidity:    w.Current.Humidity,
		Description: w.Current.Condition.Text,
	}
}

type Frequency uint8

const (
	Hourly Frequency = iota
	Daily
)

func (f Frequency) String() string {
	switch f {
	case Hourly:
		return "hourly"
	case Daily:
		return "daily"
	default:
		panic("impossible frequency value")
	}
}

func FrequencyFromString(s string) (Frequency, error) {
	switch s {
	case "hourly":
		return Hourly, nil
	case "daily":
		return Daily, nil
	default:
		return 0, errors.New("invalid frequency")
	}
}

type WeatherPayload struct {
	City string `form:"city" binding:"required"`
}

type SubscribePayload struct {
	Email     string `form:"email" binding:"required,email"`
	City      string `form:"city" binding:"required"`
	Frequency string `form:"frequency" binding:"required,oneof=hourly daily"`
}

type ConfirmPayload struct {
	Token string
}

type UnsubscribePayload struct {
	Token string
}

type TextResponse struct {
	Code    uint16 `json:"code"`
	Message string `json:"message"`
}
