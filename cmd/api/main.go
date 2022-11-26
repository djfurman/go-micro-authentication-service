package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/djfurman/go-micro-authentication-service/data"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const webPort = "80"
const retryDBConnAttempts = 10
const retryDBConnBackoff = 2 // backoff timer in seconds
var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting authentication service")

	// connect to the DB
	conn := connectToDB()
	if conn == nil {
		log.Panic("Cannot connect to Postgres")
	}

	// setup config
	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func connectToDB() *sql.DB {
	// ! todo identify how to handle a missing DSN
	dsn := os.Getenv("DSN")

	for {

		connection, err := openDB(dsn)
		if err != nil {
			log.Printf("Attempt %d, Postgres not yet ready...", counts)
			counts++
		} else {
			log.Println("Successfully Connected to Postgres!")
			return connection
		}

		if counts > retryDBConnAttempts {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for two seconds ...")
		time.Sleep(retryDBConnBackoff * time.Second)
		continue
	}
}
