import { useState } from "react";
import { Input } from "./ui/input";
import { getPokemonByName } from "@/api/pokemon";
import { Pokemon } from "@/types/pokemon";
import { Button } from "./ui/button";
import { useToast } from "./ui/use-toast";
import { calcuateTypeCoverage } from "@/lib/typeCalculator";

interface PokemonSearchBarProps {
    setPokemons: React.Dispatch<React.SetStateAction<Pokemon[]>>;
}

export function PokemonSearchBar({ setPokemons }: PokemonSearchBarProps) {
    const [input, setInput] = useState("");
    const { toast } = useToast();

    // TODO: clean up this function
    async function handleSearchPokemon() {
        if (input) {
            try {
                const pokemon = await getPokemonByName(input);
                const parsedTypes = pokemon.types;
                for (const type of parsedTypes) {
                    type.type.name =
                        type.type.name.charAt(0).toUpperCase() +
                        type.type.name.slice(1);
                }
                pokemon.types = parsedTypes;
                const coverages = calcuateTypeCoverage(
                    pokemon.types.map((type) => type.type.name)
                );
                pokemon.typeStrengths = coverages.strengths;
                pokemon.typeWeaknesses = coverages.weaknesses;
                pokemon.typeNeutrals = coverages.neutrals;
                pokemon.typeNoEffects = coverages.noEffects;
                setPokemons((prevPokemons) => [...prevPokemons, pokemon]);
            } catch (error) {
                console.log(error);
                toast({
                    variant: "destructive",
                    title: "Uh oh! Something went wrong.",
                    description: "There was a problem with your request.",
                });
            }
        }
    }

    return (
        <div className="flex items-center space-x-2">
            <Input
                placeholder="Search for Pokemon"
                type="search"
                value={input}
                onChange={(e) => setInput(e.target.value)}
            />
            <Button onClick={handleSearchPokemon}>Search</Button>
        </div>
    );
}
