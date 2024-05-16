import { Ability } from "./ability";
import { Sprite } from "./sprite";
import { Type } from "./type";

export type Pokemon = {
    id: number;
    name: string;
    sprites: Sprite;
    abilities: Ability[];
    types: Type[];
};
