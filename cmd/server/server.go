package main

import (
	"github.com/gin-gonic/gin"
	"github.com/theammir/genesis-test/api"
)

func weatherHandler(c *gin.Context) {
	var payload api.WeatherPayload
	if err := c.ShouldBind(&payload); err != nil {
		c.JSON(400, api.TextResponse{Code: 400, Message: "Invalid request"})
		return
	}

	c.JSON(200, api.TextResponse{Code: 200, Message: payload.City})
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

func main() {
	router := gin.Default()

	router.GET("/weather", weatherHandler)
	router.POST("/subscribe", subscribeHandler)
	router.GET("/confirm/:token", confirmHandler)
	router.GET("/unsubscribe/:token", unsubscribeHandler)

	router.Run(":8080")
}
