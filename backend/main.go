package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
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

type Pokemon struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Sprites   Sprite     `json:"sprites"`
	Abilities []Ability  `json:"abilities"`
	Types     []TypeInfo `json:"types"`
}

type Sprite struct {
	FrontDefault string `json:"front_default"`
}

type Ability struct {
	IsHidden bool `json:"is_hidden"`
	Slot     int  `json:"slot"`
	Ability  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"ability"`
}

type TypeInfo struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`
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

	router := httprouter.New()
	router.GET("/pokemon", app.GetPokemonHandler)

	// TODO: change url dynamically based on environment
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(router)

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

func (app *App) GetPokemonHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.URL.Query().Get("name")
	if name == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	pokemon, err := app.FetchPokemon(app.DB, name)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if pokemon == nil {
		app.notFound(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pokemon)
}

func (app *App) FetchPokemon(db *sql.DB, name string) (*Pokemon, error) {
	if x, found := app.cache.Get(name); found {
		return x.(*Pokemon), nil
	}

	pkm, err := fetchPokemonFromDatabase(db, name)
	if err != nil {
		return nil, err
	}
	if pkm != nil {
		app.cache.Set(name, pkm, cache.DefaultExpiration)
		return pkm, nil
	}

	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch pokemon data")
	}

	var pokemon Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemon); err != nil {
		return nil, err
	}

	if err := insertPokemonIntoDatabase(db, &pokemon); err != nil {
		return nil, err
	}

	app.cache.Set(name, &pokemon, cache.DefaultExpiration)

	return pkm, nil
}

func fetchPokemonFromDatabase(db *sql.DB, name string) (*Pokemon, error) {
	query := `
		SELECT p.id, p.name, p.sprite_url 
		FROM pokemon p 
		WHERE p.name = ?
	`
	row := db.QueryRow(query, name)

	var pokemon Pokemon
	var spriteURL string

	if err := row.Scan(&pokemon.ID, &pokemon.Name, &spriteURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	pokemon.Sprites.FrontDefault = spriteURL

	if err := fetchPokemonAbilities(db, &pokemon); err != nil {
		return nil, err
	}

	if err := fetchPokemonTypes(db, &pokemon); err != nil {
		return nil, err
	}

	return &pokemon, nil
}

func fetchPokemonAbilities(db *sql.DB, pokemon *Pokemon) error {
	query := `
		SELECT pa.is_hidden, pa.slot, a.name, a.pokeapi_url
		FROM pokemon_ability pa
		JOIN ability a ON pa.ability_id = a.id
		WHERE pa.pokemon_id = ?
		ORDER BY pa.slot
	`
	rows, err := db.Query(query, pokemon.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ability Ability
		if err := rows.Scan(&ability.IsHidden, &ability.Slot, &ability.Ability.Name, &ability.Ability.URL); err != nil {
			return err
		}
		pokemon.Abilities = append(pokemon.Abilities, ability)
	}

	return rows.Err()
}

func fetchPokemonTypes(db *sql.DB, pokemon *Pokemon) error {
	query := `
		SELECT pt.slot, t.name, t.pokeapi_url
		FROM pokemon_type pt
		JOIN type t ON pt.type_id = t.id
		WHERE pt.pokemon_id = ?
		ORDER BY pt.slot
	`
	rows, err := db.Query(query, pokemon.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var typeInfo TypeInfo
		if err := rows.Scan(&typeInfo.Slot, &typeInfo.Type.Name, &typeInfo.Type.URL); err != nil {
			return err
		}
		pokemon.Types = append(pokemon.Types, typeInfo)
	}

	return rows.Err()
}

func insertPokemonIntoDatabase(db *sql.DB, pokemon *Pokemon) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO pokemon (pokeapi_id, name, sprite_url) VALUES (?, ?, ?)",
		pokemon.ID, pokemon.Name, pokemon.Sprites.FrontDefault)
	if err != nil {
		return err
	}

	pokemonID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	if pokemonID == 0 {
		err := tx.QueryRow("SELECT id FROM pokemon WHERE pokeapi_id = ?", pokemon.ID).Scan(&pokemonID)
		if err != nil {
			return err
		}
	}

	for _, pkmAbility := range pokemon.Abilities {
		ability := pkmAbility.Ability
		_, err := tx.Exec("INSERT INTO ability (name, pokeapi_url) VALUES (?, ?)", ability.Name, ability.URL)
		if err != nil {
			return err
		}

		var abilityID int64
		err = tx.QueryRow("SELECT id FROM ability WHERE name = ?", ability.Name).Scan(&abilityID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO pokemon_ability (pokemon_id, ability_id, is_hidden, slot) VALUES (?, ?, ?, ?)",
			pokemonID, abilityID, pkmAbility.IsHidden, pkmAbility.Slot)
		if err != nil {
			return err
		}
	}

	for _, pkmType := range pokemon.Types {
		pokemonType := pkmType.Type
		_, err := tx.Exec("INSERT INTO type (name, pokeapi_url) VALUES (?, ?)", pokemonType.Name, pokemonType.URL)
		if err != nil {
			return err
		}

		var typeID int64
		err = tx.QueryRow("SELECT id FROM type WHERE name = ?", pokemonType.Name).Scan(&typeID)
		if err != nil {
			return err
		}

		_, err = tx.Exec("INSERT INTO pokemon_type (pokemon_id, type_id, slot) VALUES (?, ?, ?)",
			pokemonID, typeID, pkmType.Slot)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
