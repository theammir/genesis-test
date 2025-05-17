package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goloop/env"
	"github.com/theammir/genesis-test/api"
	"github.com/theammir/genesis-test/api/weather"
	database "github.com/theammir/genesis-test/internal"
)

var db *sql.DB
var client weather.Client

func weatherHandler(c *gin.Context) {
	var payload api.WeatherPayload
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	weather, err := client.GetCurrentWeather(ctx, payload.City)
	if err != nil {
		log.Printf(`GetCurrentWeather("%s") failed: %v`, payload.City, err)
		c.JSON(400, api.TextResponse{Code: 400, Message: "Something went wrong"})
		return
	}

	c.JSON(200, api.FromWeatherResponse(weather))
}

func subscribeHandler(c *gin.Context) {
	var payload api.SubscribePayload
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid input"})
		return
	}

	c.JSON(200, payload)
}

func confirmHandler(c *gin.Context) {
	payload := api.ConfirmPayload{Token: c.Param("token")}
	if payload.Token == "" {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid token"})
		return
	}

	c.JSON(200, api.TextResponse{Code: 200, Message: "Subscription confirmed successfully"})
}

func unsubscribeHandler(c *gin.Context) {
	payload := api.UnsubscribePayload{Token: c.Param("token")}
	if payload.Token == "" {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid token"})
		return
	}

	c.JSON(200, api.TextResponse{Code: 200, Message: "Unsubscribed successfully"})
}

type EnvConfig struct {
	Host       string `env:"HOST" def:"0.0.0.0"`
	Port       string `env:"PORT" def:"8080"`
	APIKey     string `env:"WEATHERAPI_KEY"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DB"`
	DBHost     string `env:"POSTGRES_HOST" def:"db"`
	DBPort     string `env:"POSTGRES_PORT" def:"5432"`
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	var cfg EnvConfig
	if err := env.Unmarshal("", &cfg); err != nil {
		log.Fatal("Environment variables missing!!")
		return
	}

	db = database.Get(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	database.MigrateUp(db)

	client = *weather.NewClient(cfg.APIKey)

	router := gin.Default()

	router.GET("/weather", weatherHandler)
	router.POST("/subscribe", subscribeHandler)
	router.GET("/confirm/:token", confirmHandler)
	router.GET("/unsubscribe/:token", unsubscribeHandler)

	router.Run(cfg.Host + ":" + cfg.Port)
}
