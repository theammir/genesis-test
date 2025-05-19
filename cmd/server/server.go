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
	var url string
	if cfg.TLSCertPath != "" {
		url += "https://"
	} else {
		url += "http://"
	}
	url += cfg.Host
	if cfg.Port != "80" && cfg.Port != "443" {
		url += ":" + cfg.Port
	}
	url += "/confirm/" + token
	return url
}

type EnvConfig struct {
	Host            string `env:"HOST" def:"0.0.0.0"`
	Port            string `env:"PORT" def:"8080"`
	TLSCertPath     string `env:"TLS_CERT_PATH"`
	TLSKeyPath      string `env:"TLS_KEY_PATH"`
	TrustedPlatform string `env:"TRUSTED_PLATFORM" def:""`

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

func initGlobalHandles() {
	if err := env.Unmarshal("", &cfg); err != nil {
		log.Fatal("Environment variables missing!!")
	}

	db = database.Get(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	weatherClient = weather.NewClient(cfg.APIKey)

	log.Println("Initializing SMTP client...")
	var err error
	mailClient, err = mail.NewClient(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)
	if err != nil {
		log.Printf("Couldn't create an SMTP client: %v", err)
	}
}

func setTrustedPlatform(router *gin.Engine) {
	if cfg.TrustedPlatform == "" {
		router.SetTrustedProxies(nil)
	} else {
		switch cfg.TrustedPlatform {
		case "GOOGLE":
			router.TrustedPlatform = gin.PlatformGoogleAppEngine
		case "CLOUDFLARE":
			router.TrustedPlatform = gin.PlatformCloudflare
		case "FLYIO":
			router.TrustedPlatform = gin.PlatformFlyIO
		default:
			router.TrustedPlatform = cfg.TrustedPlatform
		}
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetOutput(os.Stdout)

	initGlobalHandles()

	database.MigrateUp(db)
	defer mailClient.Close()

	scheduler := SpawnHourlyWorker(context.Background())
	defer scheduler.Stop()

	router := gin.Default()
	setTrustedPlatform(router)

	router.GET("/weather", weatherHandler)
	router.POST("/subscribe", subscribeHandler)
	router.GET("/confirm/:token", confirmHandler)
	router.GET("/unsubscribe/:token", unsubscribeHandler)

	if cfg.TLSCertPath != "" {
		router.RunTLS(cfg.Host+":"+cfg.Port, cfg.TLSCertPath, cfg.TLSKeyPath)
	} else {
		router.Run(cfg.Host + ":" + cfg.Port)
	}
}
