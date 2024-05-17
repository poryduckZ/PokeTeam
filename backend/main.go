package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
	"github.com/rs/cors"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type App struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	DB       *sql.DB
	cache    *cache.Cache
}

func main() {

	addr := flag.String("addr", ":8080", "HTTP network address")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		errorLog.Println("Error loading .env file:", err)
		return
	}

	dbURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbURL == "" || authToken == "" {
		errorLog.Println("Missing required environment variables. Please check .env file for database URL and auth token.")
		return
	}

	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", url)
	if err != nil {
		errorLog.Output(2, fmt.Sprintf("failed to open db %s: %s", url, err))
		os.Exit(1)
	}
	defer db.Close()

	infoLog.Println("Successfully connected to database")

	pokemonCache := cache.New(5*time.Minute, 10*time.Minute)

	app := &App{
		errorLog: errorLog,
		infoLog:  infoLog,
		DB:       db,
		cache:    pokemonCache,
	}

	// TODO: change url dynamically based on environment
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(app.routes())

	srv := &http.Server{
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
