package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/patrickmn/go-cache"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type App struct {
	DB    *sql.DB
	cache *cache.Cache
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

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	dbURL := os.Getenv("TURSO_DATABASE_URL")
	authToken := os.Getenv("TURSO_AUTH_TOKEN")

	if dbURL == "" || authToken == "" {
		fmt.Println("DATABASE_NAME and AUTH_TOKEN must be set in the .env file")
		return
	}

	url := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)

	db, err := sql.Open("libsql", url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", url, err)
		os.Exit(1)
	}
	defer db.Close()

	pokemonCache := cache.New(5*time.Minute, 10*time.Minute)

	app := &App{DB: db, cache: pokemonCache}

	router := httprouter.New()
	router.GET("/pokemon", app.GetPokemonHandler)
	http.ListenAndServe(":8080", router)
}

func (app *App) GetPokemonHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	pokemon, err := app.FetchPokemon(app.DB, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	return &pokemon, nil
}

func fetchPokemonFromDatabase(db *sql.DB, name string) (*Pokemon, error) {
	rows, err := db.Query("SELECT * FROM pokemon WHERE name = ?", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var pokemon Pokemon
	if err := rows.Scan(&pokemon.ID, &pokemon.Name); err != nil {
		return nil, err
	}

	return &pokemon, nil
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
