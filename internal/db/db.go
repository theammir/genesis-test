package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func GetConnString(host, port, user, password, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

func Get(host, port, user, password, dbname string) *sql.DB {
	log.Printf("Connectng to DB(%s) at %s:%s", dbname, host, port)

	connString := GetConnString(host, port, user, password, dbname)
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Couldn't open DB connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}
	return db
}

func MigrateUp(db *sql.DB) {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	log.Println("Migrating database...")
	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
