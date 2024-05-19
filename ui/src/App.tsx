import { useState } from "react";
import "./App.css";
import { PokemonSearchBar } from "./components/PokemonSearchBar";
import { Pokemon } from "./types/pokemon";
import { TypeChip } from "./components/TypeChip";

function App() {
    const [selectedPokemons, setSelectedPokemons] = useState<Pokemon[]>([]);
    return (
        <div className="w-full max-w-6xl mx-auto py-8">
            <PokemonSearchBar setPokemons={setSelectedPokemons} />
            <div className="mt-4 bg-white border border-gray-200 rounded-lg shadow-sm dark:bg-gray-900 dark:border-gray-800">
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 p-4">
                    {selectedPokemons?.map((pokemon) => (
                        <div
                            key={pokemon.id}
                            className="flex items-center bg-gray-100 dark:bg-gray-800 rounded-lg p-3 cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                        >
                            <img
                                alt={pokemon.name}
                                className="rounded-full"
                                height={40}
                                src={pokemon.sprites.front_default}
                                style={{
                                    aspectRatio: "40/40",
                                    objectFit: "cover",
                                }}
                                width={40}
                            />
                            <div className="ml-3">
                                <div className="font-medium text-gray-900 dark:text-gray-100 text-lg">
                                    {pokemon.name.charAt(0).toUpperCase() +
                                        pokemon.name.slice(1)}
                                </div>
                                <div className="text-sm text-gray-500 dark:text-gray-400 mt-2">
                                    {pokemon.types.map((type) => (
                                        <TypeChip type={type.type.name} />
                                    ))}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
            <div className="mt-8 bg-white border border-gray-200 rounded-lg shadow-sm dark:bg-gray-900 dark:border-gray-800 p-4">
                <h2 className="text-lg font-medium mb-4">Team Effectiveness</h2>
                <div className="grid grid-cols-3 gap-4">
                    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
                        <h3 className="text-sm font-medium mb-2">
                            Super Effective
                        </h3>
                        <div className="flex items-center space-x-2">
                            <span>TODO</span>
                        </div>
                    </div>
                    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
                        <h3 className="text-sm font-medium mb-2">
                            Not Very Effective
                        </h3>
                        <div className="flex items-center space-x-2">
                            <span>TODO</span>
                        </div>
                    </div>
                    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
                        <h3 className="text-sm font-medium mb-2">No Effect</h3>
                        <div className="flex items-center space-x-2">
                            <span>TODO</span>
                        </div>
                    </div>
                </div>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-6 max-w-6xl mx-auto py-4">
                {selectedPokemons.map((pokemon) => (
                    <div
                        key={pokemon.id}
                        className="bg-white dark:bg-gray-950 rounded-lg shadow-md overflow-hidden"
                    >
                        <div className="p-6">
                            <img
                                alt={pokemon.name}
                                className="rounded-full"
                                height={40}
                                src={pokemon.sprites.front_default}
                                style={{
                                    aspectRatio: "40/40",
                                    objectFit: "cover",
                                }}
                                width={40}
                            />
                            <h3 className="text-lg font-semibold">
                                {pokemon.name.charAt(0).toUpperCase() +
                                    pokemon.name.slice(1)}
                            </h3>
                            <div className="flex items-center gap-2 mt-2">
                                {pokemon.types.map((type) => (
                                    <TypeChip type={type.type.name} />
                                ))}
                            </div>
                            <div className="mt-4">
                                <p className="text-gray-500 dark:text-gray-400 text-sm">
                                    Resistances:
                                </p>
                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                    {pokemon.typeStrengths.map((strength) => {
                                        return (
                                            <TypeChip
                                                type={Object.keys(strength)[0]}
                                                effectiveness={
                                                    Object.values(strength)[0]
                                                }
                                            />
                                        );
                                    })}
                                </div>
                            </div>
                            <div className="mt-4">
                                <p className="text-gray-500 dark:text-gray-400 text-sm">
                                    Weaknesses:
                                </p>
                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                    {pokemon.typeWeaknesses.map((weakness) => {
                                        return (
                                            <TypeChip
                                                type={Object.keys(weakness)[0]}
                                                effectiveness={
                                                    Object.values(weakness)[0]
                                                }
                                            />
                                        );
                                    })}
                                </div>
                            </div>
                            <div className="mt-4">
                                <p className="text-gray-500 dark:text-gray-400 text-sm">
                                    Neutrals:
                                </p>
                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                    {pokemon.typeNeutrals.map((neutral) => {
                                        return (
                                            <TypeChip
                                                type={Object.keys(neutral)[0]}
                                                effectiveness={
                                                    Object.values(neutral)[0]
                                                }
                                            />
                                        );
                                    })}
                                </div>
                            </div>
                            <div className="mt-4">
                                <p className="text-gray-500 dark:text-gray-400 text-sm">
                                    No effect:
                                </p>
                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                    {pokemon.typeNoEffects.length > 0 ? (
                                        pokemon.typeNoEffects.map(
                                            (noEffect) => {
                                                return (
                                                    <TypeChip
                                                        type={
                                                            Object.keys(
                                                                noEffect
                                                            )[0]
                                                        }
                                                    />
                                                );
                                            }
                                        )
                                    ) : (
                                        <span>None</span>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
}

export default App;
