import { Pokemon } from "@/types/pokemon";
import { api } from "..";
import { Ability } from "@/types/ability";
import { Type } from "@/types/type";

export async function fetchPokemonByName(name: string) {
    // const searchParams = new URLSearchParams({ name });
    // const pokemon: Pokemon = await api.get("pokemon/", { searchParams }).json();
    // return pokemon;

    const response = await fetch(`http://localhost:8080/pokemon?name=${name}`);

    if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
    }

    const pokemon = await response.json();
    return pokemon;
}
