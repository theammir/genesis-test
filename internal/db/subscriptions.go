package db

import (
	"database/sql"
	"log"

	"github.com/theammir/genesis-test/api"
)

type Subscriber struct {
	Email     string
	City      string
	Frequency string
}

// Generates a unique confirmation token.
func GenerateToken(db *sql.DB) (string, error) {
	// TODO: Proper implementation
	var count string
	err := db.QueryRow("SELECT count(*) FROM confirmations;").Scan(&count)
	if err == sql.ErrNoRows {
		return "0", nil
	}
	if err != nil {
		return "", err
	}

	return count, nil
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

	token, err := GenerateToken(db)
	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(`
			INSERT INTO confirmations (email, token)
			VALUES ($1, $2);
		`, payload.Email, token); err != nil {
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
		SELECT (email, city, frequency)
		FROM subscriptions
		WHERE confirmed = true
		AND frequency = $1
	`, frequency)
	if err != nil {
		return nil, err
	}

	var subs []Subscriber
	for rows.Next() {
		var sub Subscriber
		if err := rows.Scan(&sub.Email, &sub.City, &sub.Frequency); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}

	return subs, nil
}
