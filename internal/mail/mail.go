package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"sync"

	"github.com/theammir/genesis-test/api"
	database "github.com/theammir/genesis-test/internal/db"
)

type Client struct {
	host     string
	port     string
	user     string
	password string
	from     string
	mu       sync.Mutex
	*smtp.Client
}

func NewClient(host, port, user, password, from string) (*Client, error) {
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, fmt.Errorf("dial error: %w", err)
	}
	smtpC, err := smtp.NewClient(conn, host)
	if err != nil {
		return nil, fmt.Errorf("new client error: %w", err)
	}

	if ok, _ := smtpC.Extension("STARTTLS"); !ok {
		return nil, fmt.Errorf("server does not support STARTTLS")
	}
	if err := smtpC.StartTLS(&tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
	}); err != nil {
		return nil, fmt.Errorf("starttls error: %w", err)
	}

	auth := smtp.PlainAuth("", user, password, host)
	if err := smtpC.Auth(auth); err != nil {
		return nil, fmt.Errorf("auth error: %w", err)
	}

	return &Client{host, port, user, password, from, sync.Mutex{}, smtpC}, nil
}

// Send an email to a single subject. `message` must be CRLF formatted.
func (c *Client) sendEmail(to, subject, message string) error {
	messageBytes := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			message + "\r\n",
	)

	c.mu.Lock()
	defer c.mu.Unlock()

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

func (c *Client) SendConfirmation(payload api.SubscribePayload, unsubUrl string) error {
	message := fmt.Sprintf("Confirm your subscription for %s forecast updates in %s: %s",
		payload.Frequency, payload.City, unsubUrl)
	return c.sendEmail(payload.Email, "Weather subscription confirmation", message)
}

func (c *Client) SendWeather(sub database.Subscriber, weather api.Weather, unsubUrl string) error {
	message := fmt.Sprintf(
		"Your %s forecast:\r\n"+
			"It is %f °C in %s; %s; %d%% of humidity.\r\n\r\n"+
			"Unsubscribe: %s",
		sub.Frequency, weather.Temperature, sub.City, weather.Description, weather.Humidity, unsubUrl)
	return c.sendEmail(sub.Email, fmt.Sprintf("Weather forecast in %s", sub.City), message)
}
