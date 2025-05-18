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
	database "github.com/theammir/genesis-test/internal/db"
	"github.com/theammir/genesis-test/internal/mail"
)

var (
	cfg           EnvConfig
	db            *sql.DB
	weatherClient *weather.Client
	mailClient    *mail.Client
)

func weatherHandler(c *gin.Context) {
	var payload api.WeatherPayload
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid request"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	weather, err := weatherClient.GetCurrentWeather(ctx, payload.City)
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

	newToken, err := database.SubscribeUser(db, &payload)
	if err != nil {
		c.JSON(409, api.TextResponse{Code: 409, Message: "Email already subscribed"})
		return
	}

	url := "http://" + cfg.Host
	if cfg.Port != "80" && cfg.Port != "443" {
		url += ":" + cfg.Port
	}
	url += "/confirm/" + newToken
	log.Printf("Sending confirmation email to %s (token `%s`)", payload.Email, newToken)
	mailClient.SendConfirmation(payload, url)

	c.JSON(200, api.TextResponse{Code: 200, Message: "Subscription successful. Confirmation email sent."})
}

func confirmHandler(c *gin.Context) {
	payload := api.ConfirmPayload{Token: c.Param("token")}
	if payload.Token == "" {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid token"})
		return
	}

	err := database.ConfirmUser(db, payload.Token)
	if err != nil {
		c.JSON(404, api.TextResponse{Code: 404, Message: "Token not found"})
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

	err := database.UnsubscribeUser(db, payload.Token)
	if err != nil {
		c.JSON(404, api.TextResponse{Code: 404, Message: "Token not found"})
		return
	}

	c.JSON(200, api.TextResponse{Code: 200, Message: "Unsubscribed successfully"})
}

type EnvConfig struct {
	Host string `env:"HOST" def:"0.0.0.0"`
	Port string `env:"PORT" def:"8080"`

	APIKey string `env:"WEATHERAPI_KEY"`

	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DB"`
	DBHost     string `env:"POSTGRES_HOST" def:"db"`
	DBPort     string `env:"POSTGRES_PORT" def:"5432"`

	SMTPHost string `env:"SMTP_HOST"`
	SMTPPort string `env:"SMTP_PORT" def:"587"`
	SMTPUser string `env:"SMTP_USER"`
	SMTPPass string `env:"SMTP_PASS"`
	SMTPFrom string `env:"SMTP_FROM"`
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	if err := env.Unmarshal("", &cfg); err != nil {
		log.Fatal("Environment variables missing!!")
	}

	db = database.Get(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	database.MigrateUp(db)

	weatherClient = weather.NewClient(cfg.APIKey)
	log.Println("Initializing SMTP client...")
	newMailClient, err := mail.NewClient(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	if err != nil {
		log.Fatalf("Couldn't create an SMTP client: %v", err)
	}
	mailClient = newMailClient
	defer mailClient.Close()

	router := gin.Default()

	router.GET("/weather", weatherHandler)
	router.POST("/subscribe", subscribeHandler)
	router.GET("/confirm/:token", confirmHandler)
	router.GET("/unsubscribe/:token", unsubscribeHandler)

	router.Run(cfg.Host + ":" + cfg.Port)
}
