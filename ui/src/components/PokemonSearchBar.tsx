import { useEffect, useState } from "react";
import { Input } from "./ui/input";
import { SearchIcon } from "lucide-react";
import { useDebounce } from "@/hooks/useDebounce";
import { getPokemonByName } from "@/api/pokemon";
import { Pokemon } from "@/types/pokemon";
import { Button } from "./ui/button";
import { useToast } from "./ui/use-toast";

interface PokemonSearchBarProps {
    setPokemons: React.Dispatch<React.SetStateAction<Pokemon[]>>;
}

export function PokemonSearchBar({ setPokemons }: PokemonSearchBarProps) {
    const [input, setInput] = useState("");
    const { toast } = useToast();

    async function handleSearchPokemon() {
        if (input) {
            try {
                const pokemon = await getPokemonByName(input);
                setPokemons((prevPokemons) => [...prevPokemons, pokemon]);
            } catch (error) {
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
