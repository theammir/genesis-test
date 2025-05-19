package main

import (
	"context"
	"log"

	"github.com/robfig/cron"
	"github.com/theammir/genesis-test/api"
	database "github.com/theammir/genesis-test/internal/db"
)

func SpawnHourlyWorker(ctx context.Context) *cron.Cron {
	scheduler := cron.New()

	var counter uint = 0
	scheduler.AddFunc("0 0 * * * *", func() {
		SendWeatherForecasts(ctx, counter)
		counter++
	})

	scheduler.Start()
	return scheduler
}

func SendWeatherForecasts(ctx context.Context, counter uint) {
	subs, err := database.FetchSubscribers(db, "hourly")
	if err != nil {
		log.Printf("Failed to fetch hourly subscribers: %v", err)
		return
	}
	if counter%24 == 0 {
		dailySubs, err := database.FetchSubscribers(db, "daily")
		if err != nil {
			log.Printf("Failed to fetch hourly subscribers: %v", err)
		} else {
			subs = append(subs, dailySubs...)
		}
	}
	weatherCache := make(map[string]api.Weather, len(subs))

	for _, sub := range subs {
		log.Printf("Searching %s in cache...", sub.City)
		weather, ok := weatherCache[sub.City]
		if !ok {
			log.Printf("...not found. Fetching from API.")
			weatherResp, err := weatherClient.GetCurrentWeather(ctx, sub.City)
			if err != nil {
				log.Printf("Error when getting weather: %v", err)
				continue
			}
			weather = api.FromWeatherResponse(weatherResp)
			weatherCache[sub.City] = weather
		}
		log.Printf("Sending %s weather forecast in %s to %s", sub.Frequency, sub.City, sub.Email)
		if err := mailClient.SendWeather(sub, weather, GetUnsubUrl(sub.Token)); err != nil {
			log.Printf("Couldn't send weather forecast: %v", err)
		}
	}
}
