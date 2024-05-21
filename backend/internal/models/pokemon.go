package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Pokemon struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Sprite    Sprite    `json:"sprites"`
	Abilities []Ability `json:"abilities"`
	Types     []Type    `json:"types"`
}

type PokemonRes struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Sprite    Sprite       `json:"sprites"`
	Abilities []AbilityRes `json:"abilities"`
	Types     []TypeRes    `json:"types"`
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

type AbilityRes struct {
	Name string `json:"name"`
}

type Type struct {
	Slot int `json:"slot"`
	Type struct {
		Name            string          `json:"name"`
		URL             string          `json:"url"`
		DamageRelations DamageRelations `json:"damage_relations"`
	} `json:"type"`
}

type DamageRelations struct {
	DoubleDamageFrom []RelationType `json:"double_damage_from"`
	DoubleDamageTo   []RelationType `json:"double_damage_to"`
	HalfDamageFrom   []RelationType `json:"half_damage_from"`
	HalfDamageTo     []RelationType `json:"half_damage_to"`
	NoDamageFrom     []RelationType `json:"no_damage_from"`
	NoDamageTo       []RelationType `json:"no_damage_to"`
}

type RelationType struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type TypeRes struct {
	Name string `json:"name"`
}

func (pkm *PokemonRes) Get(db *sql.DB, name string) (*PokemonRes, error) {
	query := `
		SELECT p.id, p.name, p.sprite_url 
		FROM pokemon p 
		WHERE p.name = ?
	`
	row := db.QueryRow(query, name)

	var pokemon PokemonRes
	var spriteURL string

	if err := row.Scan(&pokemon.ID, &pokemon.Name, &spriteURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	pokemon.Sprite.FrontDefault = spriteURL

	if err := pkm.getPokemonAbilities(db, &pokemon); err != nil {
		return nil, fmt.Errorf("failed to get abilities: %w", err)
	}

	if err := pkm.getPokemonTypes(db, &pokemon); err != nil {
		return nil, fmt.Errorf("failed to get types: %w", err)
	}

	return &pokemon, nil
}

func (pkm *PokemonRes) getPokemonAbilities(db *sql.DB, pokemon *PokemonRes) error {
	query := `
		SELECT a.name
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
		var ability AbilityRes
		if err := rows.Scan(&ability.Name); err != nil {
			return err
		}
		pokemon.Abilities = append(pokemon.Abilities, ability)
	}

	return rows.Err()
}

func (pkm *PokemonRes) getPokemonTypes(db *sql.DB, pokemon *PokemonRes) error {
	query := `
		SELECT t.name
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
		var typeInfo TypeRes
		if err := rows.Scan(&typeInfo.Name); err != nil {
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
		pokemon.ID, pokemon.Name, pokemon.Sprite.FrontDefault)
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
		var abilityID int64
		err := tx.QueryRow("SELECT id FROM ability WHERE name = ?", pkmAbility.Ability.Name).Scan(&abilityID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get ability id: %w", err)
		}

		if abilityID == 0 {
			res, err := tx.Exec("INSERT INTO ability (name, pokeapi_url) VALUES (?, ?)", pkmAbility.Ability.Name, pkmAbility.Ability.URL)
			if err != nil {
				return fmt.Errorf("failed to insert ability: %w", err)
			}
			abilityID, err = res.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert id for ability: %w", err)
			}
		}

		_, err = tx.Exec("INSERT INTO pokemon_ability (pokemon_id, ability_id, is_hidden, slot) VALUES (?, ?, ?, ?)",
			pokemonID, abilityID, pkmAbility.IsHidden, pkmAbility.Slot)
		if err != nil {
			return fmt.Errorf("failed to insert pokemon ability: %w", err)
		}
	}

	for _, pkmType := range pokemon.Types {
		var typeID int64
		err := tx.QueryRow("SELECT id FROM type WHERE name = ?", pkmType.Type.Name).Scan(&typeID)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get type id: %w", err)
		}

		if typeID == 0 {
			typeID, err = strconv.ParseInt(pkmType.Type.URL[strings.LastIndex(pkmType.Type.URL, "/")+1:], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse type ID from URL: %w", err)
			}

			fetchedType, err := fetchTypeDetails(typeID)
			if err != nil {
				return fmt.Errorf("error fetching type details: %w", err)
			}

			res, err := tx.Exec("INSERT INTO type (name, pokeapi_url) VALUES (?, ?)", fetchedType.Type.Name, fetchedType.Type.URL)
			if err != nil {
				return fmt.Errorf("failed to insert type: %w", err)
			}
			typeID, err = res.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert id for type: %w", err)
			}
		}

		_, err = tx.Exec("INSERT INTO pokemon_type (pokemon_id, type_id, slot) VALUES (?, ?, ?)", pokemonID, typeID, pkmType.Slot)
		if err != nil {
			return fmt.Errorf("failed to insert pokemon type: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func fetchTypeDetails(typeID int64) (*Type, error) {
	resp, err := http.Get(fmt.Sprintf("https://pokeapi.co/api/v2/type/%d", typeID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch type from API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response from API: %s", resp.Status)
	}

	var pkmType Type
	err = json.NewDecoder(resp.Body).Decode(&pkmType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &pkmType, nil
}

func (pkm *Pokemon) MapPokemonToPokemonRes(pokemon *Pokemon, pokemonRes *PokemonRes) (*PokemonRes, error) {
	pokemonRes.ID = pokemon.ID
	pokemonRes.Name = pokemon.Name
	pokemonRes.Sprite.FrontDefault = pokemon.Sprite.FrontDefault

	for _, ability := range pokemon.Abilities {
		pokemonRes.Abilities = append(pokemonRes.Abilities, AbilityRes{
			Name: ability.Ability.Name,
		})
	}

	for _, pkmType := range pokemon.Types {
		pokemonRes.Types = append(pokemonRes.Types, TypeRes{
			Name: pkmType.Type.Name,
		})
	}

	return pokemonRes, nil
}
