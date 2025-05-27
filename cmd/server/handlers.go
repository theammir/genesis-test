package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/theammir/genesis-test/api"
	database "github.com/theammir/genesis-test/internal/db"
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

	log.Printf("Sending confirmation email to %s (token `%s`)", payload.Email, newToken)
	if err := mailClient.SendConfirmation(payload, GetUnsubUrl(newToken)); err != nil {
		log.Printf("Couldn't send confirmation email: %v", err)
	}

	c.JSON(
		200,
		api.TextResponse{Code: 200, Message: "Subscription successful. Confirmation email sent."},
	)
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
