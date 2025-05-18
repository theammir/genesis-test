package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/theammir/genesis-test/api"
)

type Client struct {
	host     string
	port     string
	user     string
	password string
	from     string
	smtp.Client
}

func NewClient(host, port, user, password, from string) (*Client, error) {
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, fmt.Errorf("couldn't dial tcp: %w", err)
	}
	smtpC, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, fmt.Errorf("couldn't create SMTP connection: %w", err)
	}

	if ok, _ := smtpC.Extension("STARTTLS"); ok {
		tlsConf := &tls.Config{ServerName: host, InsecureSkipVerify: true}
		if err := smtpC.StartTLS(tlsConf); err != nil {
			return nil, fmt.Errorf("error negotiating STARTTLS: %w", err)
		}
	}

	auth := smtp.PlainAuth("", user, password, host)
	if err := smtpC.Auth(auth); err != nil {
		return nil, fmt.Errorf("couldn't Auth SMTP connection: %w", err)
	}

	return &Client{host, port, user, password, from, *smtpC}, nil
}

// Send an email to a single subject. `message` must be CRLF formatted.
func (c *Client) sendEmail(to, subject, message string) error {
	messageBytes := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			message + "\r\n",
	)
	if err := c.Mail(c.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(messageBytes); err != nil {
		return err
	}
	return wc.Close()
}

// TODO: On the server, make a goroutine that invokes these periodically, fetching recipients
// from DB.

func (c *Client) SendConfirmation(payload api.SubscribePayload, url string) error {
	message := fmt.Sprintf("Confirm your subscription for %s forecast updates in %s: %s",
		payload.Frequency, payload.City, url)
	return c.sendEmail(payload.Email, "Weather subscription confirmation", message)
}

func (c *Client) SendWeather(payload api.SubscribePayload, weather api.Weather, unsubUrl string) error {
	message := fmt.Sprintf(
		"Your %s forecast:\r\n"+
			"It is %f °C in %s; %s; %d%% of humidity.\r\n\r\n"+
			"Unsubscribe: %s",
		payload.Frequency, weather.Temperature, payload.City, weather.Description, weather.Humidity, unsubUrl)
	return c.sendEmail(payload.Email, fmt.Sprintf("Weather forecast in %s", payload.City), message)
}
