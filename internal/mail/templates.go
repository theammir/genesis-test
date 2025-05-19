package mail

import (
	"bytes"
	"embed"
	"fmt"
	_template "html/template"

	"github.com/theammir/genesis-test/api"
	database "github.com/theammir/genesis-test/internal/db"
)

//go:embed templates/*.html
var templatesFS embed.FS
var Confirmation string = "templates/confirmation.html"
var Weather string = "templates/weather.html"

type ConfirmationFormat struct {
	City      string
	Frequency string
	UnsubURL  string
}

type WeatherFormat struct {
	Temperature float32
	Humidity    uint8
	Description string
	ConfirmationFormat
}

func formatTemplate(template string, data any) (string, error) {
	tmpl, err := _template.ParseFS(templatesFS, template)
	if err != nil {
		return "", fmt.Errorf("not found template %q: %w", template, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("couldn't execute template %q: %w", template, err)
	}

	return buf.String(), nil
}

func GetConfirmationHTML(payload api.SubscribePayload, unsubURL string) (string, error) {
	return formatTemplate(Confirmation, ConfirmationFormat{City: payload.City, Frequency: payload.Frequency, UnsubURL: unsubURL})
}

func GetWeatherHTML(sub database.Subscriber, weather api.Weather, unsubURL string) (string, error) {
	// does gofmt not care? anyway i'll fix it later
	return formatTemplate(Weather, WeatherFormat{Temperature: weather.Temperature, Humidity: weather.Humidity, Description: weather.Description, ConfirmationFormat: ConfirmationFormat{City: sub.City, Frequency: sub.Frequency, UnsubURL: unsubURL}})
}
