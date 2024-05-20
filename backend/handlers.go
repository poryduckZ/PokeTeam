package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/patrickmn/go-cache"

	"github.com/poryduckZ/poketeam/backend/internal/models"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

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

func (app *App) FetchPokemon(db *sql.DB, name string) (*models.PokemonRes, error) {
	if pkm, found := app.cache.Get(name); found {
		app.infoLog.Printf("Retrieved from cache for %s", name)
		return pkm.(*models.PokemonRes), nil
	}

	pkm := &models.PokemonRes{}
	pokemon, err := pkm.Get(db, name)
	if err != nil {
		return nil, err
	}
	if pokemon != nil {
		app.infoLog.Printf("Retrieved from database for %s", name)
		app.cache.Set(name, pokemon, cache.DefaultExpiration)
		return pokemon, nil
	}

	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + name)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch pokemon data from PokeAPI: %s", resp.Status)
	}

	var pokemonFromAPI models.Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&pokemonFromAPI); err != nil {
		return nil, err
	}

	if err := pokemonFromAPI.Insert(db, &pokemonFromAPI); err != nil {
		return nil, err
	}

	app.cache.Set(name, &pokemonFromAPI, cache.DefaultExpiration)

	app.infoLog.Printf("Retrieved from PokeAPI for %s", name)

	pokemonResFromApi, err := pokemonFromAPI.MapPokemonToPokemonRes(&pokemonFromAPI, pkm)
	if err != nil {
		return nil, err
	}
	return pokemonResFromApi, nil
}
