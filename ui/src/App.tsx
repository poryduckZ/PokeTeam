import { useState } from "react";
import "./App.css";
import { PokemonSearchBar } from "./components/PokemonSearchBar";
import { Pokemon } from "./types/pokemon";

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
                                <div className="font-medium text-gray-900 dark:text-gray-100">
                                    {pokemon.name}
                                </div>
                                <div className="text-sm text-gray-500 dark:text-gray-400">
                                    {pokemon.types
                                        .map((type) => type.type.name)
                                        .join(", ")}
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
                            <span>Ground, Water, Flying</span>
                        </div>
                    </div>
                    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
                        <h3 className="text-sm font-medium mb-2">
                            Not Very Effective
                        </h3>
                        <div className="flex items-center space-x-2">
                            <span>Grass, Electric, Steel</span>
                        </div>
                    </div>
                    <div className="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
                        <h3 className="text-sm font-medium mb-2">No Effect</h3>
                        <div className="flex items-center space-x-2">
                            <span>Ground</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default App;
