CREATE TABLE pokemon (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    pokeapi_id INT UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    sprite_url VARCHAR(255) NOT NULL
);

CREATE TABLE ability (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    pokeapi_url VARCHAR(255) NOT NULL
);

CREATE TABLE type (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL UNIQUE,
    pokeapi_url VARCHAR(255) NOT NULL
);

CREATE TABLE damage_relations (
    type_id INTEGER,
    relation_type VARCHAR(255),
    related_type_name VARCHAR(255),
    FOREIGN KEY (type_id) REFERENCES type(id)
);

CREATE TABLE pokemon_ability (
    pokemon_id INT NOT NULL,
    ability_id INT NOT NULL,
    is_hidden BOOLEAN NOT NULL,
    slot INT NOT NULL,
    FOREIGN KEY (pokemon_id) REFERENCES pokemon (id),
    FOREIGN KEY (ability_id) REFERENCES ability (id),
    PRIMARY KEY (pokemon_id, ability_id, slot)
);

CREATE TABLE pokemon_type (
    pokemon_id INT NOT NULL,
    type_id INT NOT NULL,
    slot INT NOT NULL,
    FOREIGN KEY (pokemon_id) REFERENCES pokemon (id),
    FOREIGN KEY (type_id) REFERENCES type (id),
    PRIMARY KEY (pokemon_id, type_id)
);
