package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/theammir/genesis-test/api"
)

type Subscriber struct {
	Email     string
	City      string
	Frequency string
	Token     string
}

// Delete all the confirmation tokens of a user.
// If `preserveToken` is specified, doesn't remove that token.
func deleteConfirmations(db *sql.DB, email string, preserveToken *string) error {
	// if tokens aren't unique, this isn't gonna work
	if preserveToken == nil {
		if _, err := db.Exec(`
			DELETE FROM confirmations
			WHERE email = $1;
		`, email); err != nil {
			return err
		}
	} else {
		if _, err := db.Exec(`
			DELETE FROM confirmations
			WHERE email = $1
			AND token != $2;
		`, email, *preserveToken); err != nil {
			return err
		}
	}
	return nil
}

// Adds the user to DB and returns a unique confirmation token
func SubscribeUser(db *sql.DB, payload *api.SubscribePayload) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}

	// FIX: Should I do nothing and let new subscriptions generate new
	// confirmation links for outdated requests, ignoring the new ones,
	// or UPDATE and have old confirmations work for newer requests?
	if _, err := tx.Exec(`
			INSERT INTO subscriptions (email, city, frequency)
			VALUES ($1, $2, $3)
			ON CONFLICT (email) DO NOTHING;
		`, payload.Email, payload.City, payload.Frequency); err != nil {
		tx.Rollback()
		return "", err
	}

	var token string
	if err := tx.QueryRow(`
			INSERT INTO confirmations (email)
			VALUES ($1)
			RETURNING token;
		`, payload.Email).Scan(&token); err != nil {
		tx.Rollback()
		return "", err
	}
	log.Printf("Subscribing new user: %s", payload.Email)
	return token, tx.Commit()
}

func ConfirmUser(db *sql.DB, token string) error {
	var email string
	if err := db.QueryRow(`
			UPDATE subscriptions AS s
			SET confirmed = true
			FROM confirmations AS c
			WHERE c.token = $1 AND s.email = c.email
			RETURNING s.email;
		`, token).Scan(&email); err != nil {
		return err
	}
	// delete all the confirmation tokens except for the one
	// that will be used for unsubscribing
	deleteConfirmations(db, email, &token)

	log.Printf("User confirmed: %s", email)
	return nil
}

func UnsubscribeUser(db *sql.DB, token string) error {
	var email string
	if err := db.QueryRow(`
			DELETE FROM subscriptions AS s
			USING confirmations AS c
			WHERE c.token = $1 AND s.email = c.email
			RETURNING s.email;
		`, token).Scan(&email); err != nil {
		return err
	}
	deleteConfirmations(db, email, nil)

	log.Printf("User unsubscribed: %s", email)
	return nil
}

func FetchSubscribers(db *sql.DB, frequency string) ([]Subscriber, error) {
	rows, err := db.Query(`
		SELECT s.email, s.city, s.frequency, c.token
		FROM subscriptions s
		JOIN confirmations c
		ON s.email = c.email
		WHERE s.confirmed = true
		AND s.frequency = $1;
	`, frequency)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("couldn't fetch subscribers: %w", err)
	}

	var subs []Subscriber
	for rows.Next() {
		var sub Subscriber
		if err := rows.Scan(&sub.Email, &sub.City, &sub.Frequency, &sub.Token); err != nil {
			return nil, fmt.Errorf("invalid subscriber row: %w", err)
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
