package models

import (
	"database/sql"
	"fmt"
)

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

func (pkm *Pokemon) Get(db *sql.DB, name string) (*Pokemon, error) {
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
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	pokemon.Sprites.FrontDefault = spriteURL

	if err := pkm.getPokemonAbilities(db, &pokemon); err != nil {
		return nil, fmt.Errorf("failed to get abilities: %w", err)
	}

	if err := pkm.getPokemonTypes(db, &pokemon); err != nil {
		return nil, fmt.Errorf("failed to get types: %w", err)
	}

	return &pokemon, nil
}

func (pkm *Pokemon) getPokemonAbilities(db *sql.DB, pokemon *Pokemon) error {
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

func (pkm *Pokemon) getPokemonTypes(db *sql.DB, pokemon *Pokemon) error {
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

func (pkm *Pokemon) Insert(db *sql.DB, pokemon *Pokemon) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("INSERT INTO pokemon (pokeapi_id, name, sprite_url) VALUES (?, ?, ?)",
		pokemon.ID, pokemon.Name, pokemon.Sprites.FrontDefault)
	if err != nil {
		return fmt.Errorf("failed to insert pokemon: %w", err)
	}

	pokemonID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	if pokemonID == 0 {
		err := tx.QueryRow("SELECT id FROM pokemon WHERE pokeapi_id = ?", pokemon.ID).Scan(&pokemonID)
		if err != nil {
			return fmt.Errorf("failed to get existing pokemon id: %w", err)
		}
	}

	for _, pkmAbility := range pokemon.Abilities {
		ability := pkmAbility.Ability
		_, err := tx.Exec("INSERT INTO ability (name, pokeapi_url) VALUES (?, ?)", ability.Name, ability.URL)
		if err != nil {
			return fmt.Errorf("failed to insert ability: %w", err)
		}

		var abilityID int64
		err = tx.QueryRow("SELECT id FROM ability WHERE name = ?", ability.Name).Scan(&abilityID)
		if err != nil {
			return fmt.Errorf("failed to get ability id: %w", err)
		}

		_, err = tx.Exec("INSERT INTO pokemon_ability (pokemon_id, ability_id, is_hidden, slot) VALUES (?, ?, ?, ?)",
			pokemonID, abilityID, pkmAbility.IsHidden, pkmAbility.Slot)
		if err != nil {
			return fmt.Errorf("failed to insert pokemon ability: %w", err)
		}
	}

	for _, pkmType := range pokemon.Types {
		pokemonType := pkmType.Type
		_, err := tx.Exec("INSERT INTO type (name, pokeapi_url) VALUES (?, ?)", pokemonType.Name, pokemonType.URL)
		if err != nil {
			return fmt.Errorf("failed to insert type: %w", err)
		}

		var typeID int64
		err = tx.QueryRow("SELECT id FROM type WHERE name = ?", pokemonType.Name).Scan(&typeID)
		if err != nil {
			return fmt.Errorf("failed to get type id: %w", err)
		}

		_, err = tx.Exec("INSERT INTO pokemon_type (pokemon_id, type_id, slot) VALUES (?, ?, ?)",
			pokemonID, typeID, pkmType.Slot)
		if err != nil {
			return fmt.Errorf("failed to insert pokemon type: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
