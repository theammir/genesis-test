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

// Connect to the SMTP server, negotiate STARTTLS and authenticate.
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

func (c *Client) sendEmail(to string, message []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.Reset(); err != nil {
		newClient, err := NewClient(c.host, c.port, c.user, c.password, c.from)
		if err != nil {
			return err
		}
		c = newClient
	}

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
	if _, err := wc.Write(message); err != nil {
		return err
	}
	return wc.Close()
}

// Send a plain text email to a single subject. `message` must be CRLF formatted.
func (c *Client) sendEmailPlain(to, subject, message string) error {
	headers := []byte(
		"From: " + c.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n",
	)
	messageBytes := append(headers, []byte(
		"\r\n"+
			message+"\r\n",
	)...)

	return c.sendEmail(to, messageBytes)
}

// Send an HTML email to a single subject. `message` must be CRLF formatted.
func (c *Client) sendEmailHTML(to, subject, message string) error {
	headers := []byte(
		"From: " + c.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			`Content-Type: text/html; charset="utf-8"` + "\r\n",
	)
	messageBytes := append(headers, []byte(
		"\r\n"+
			message+"\r\n",
	)...)

	return c.sendEmail(to, messageBytes)
}

// Send a confirmation email to a user, with provided unsubscription URL
func (c *Client) SendConfirmation(payload api.SubscribePayload, unsubURL string) error {
	message, err := GetConfirmationHTML(payload, unsubURL)
	if err != nil {
		return err
	}

	return c.sendEmailHTML(payload.Email, "Weather subscription confirmation", message)
}

// Send a regular weather email to a user, with provided unsubscription URL
func (c *Client) SendWeather(sub database.Subscriber, weather api.Weather, unsubURL string) error {
	message, err := GetWeatherHTML(sub, weather, unsubURL)
	if err != nil {
		return err
	}
	return c.sendEmailHTML(sub.Email, fmt.Sprintf("Weather forecast in %s", sub.City), message)
}
