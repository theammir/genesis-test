package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/goloop/env"
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

func GetUnsubUrl(token string) string {
	url := "http://" + cfg.Host
	if cfg.Port != "80" && cfg.Port != "443" {
		url += ":" + cfg.Port
	}
	url += "/confirm/" + token
	return url
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

	scheduler := SpawnHourlyWorker(context.Background())
	defer scheduler.Stop()

	router := gin.Default()

	router.GET("/weather", weatherHandler)
	router.POST("/subscribe", subscribeHandler)
	router.GET("/confirm/:token", confirmHandler)
	router.GET("/unsubscribe/:token", unsubscribeHandler)

	router.Run(cfg.Host + ":" + cfg.Port)
}
