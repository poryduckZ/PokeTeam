import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "./components/ui/card";
import { Input } from "./components/ui/input";
import { fetchPokemonByName } from "./api/pokemon";
import { Pokemon } from "./types/pokemon";

function App() {
    const [pokemon, setPokemon] = useState<Pokemon | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [pokemonName, setPokemonName] = useState("wailord");

    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            try {
                const pokemonData = await fetchPokemonByName(pokemonName);
                setPokemon(pokemonData);
            } catch (err) {
                setError((err as Error).message || "An unknown error occurred");
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [pokemonName]);

    const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setPokemonName(event.target.value);
    };

    return (
        <Card>
            <CardHeader>
                <CardTitle>Search a Pokemon</CardTitle>
            </CardHeader>
            <CardContent>
                <Input
                    type="text"
                    value={pokemonName}
                    onChange={handleInputChange}
                    placeholder="Wailord"
                />
                <div>
                    {loading ? (
                        <h1>Loading...</h1>
                    ) : error ? (
                        <h1>Error: {error}</h1>
                    ) : pokemon ? (
                        <div>
                            <h1>{pokemon.name}</h1>
                            <img
                                src={pokemon.sprites.front_default}
                                alt={pokemon.name}
                            />
                            <h2>Abilities</h2>
                            <ul>
                                {pokemon.abilities.map((ability, index) => (
                                    <li key={index}>
                                        {ability.slot}: {ability.ability.name}{" "}
                                        {ability.is_hidden && "(Hidden)"}
                                    </li>
                                ))}
                            </ul>
                            <h2>Types</h2>
                            <ul>
                                {pokemon.types.map((type, index) => (
                                    <li key={index}>
                                        {type.slot}: {type.type.name}
                                    </li>
                                ))}
                            </ul>
                        </div>
                    ) : (
                        <h1>Pokemon not found</h1>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}

export default App;
