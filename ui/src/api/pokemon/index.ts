import { Pokemon } from "@/types/pokemon";
import { api } from "..";

export async function getPokemonByName(name: string) {
    const searchParams = new URLSearchParams({ name });
    const pokemon: Pokemon = await api.get("pokemon", { searchParams }).json();
    return pokemon;
}
