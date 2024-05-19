import { typeEffectiveness } from "@/constants/typeEffectiveness";

// TODO: move all these to backend
export type Effectiveness = {
    [type: string]: number;
};

export function calcuateTypeCoverage(types: string[]) {
    console.log(types);
    const strengths: Effectiveness[] = [];
    const weaknesses: Effectiveness[] = [];
    const neutrals: Effectiveness[] = [];
    const noEffects: Effectiveness[] = [];

    const effectiveness: Effectiveness = {
        Normal: 1,
        Fighting: 1,
        Flying: 1,
        Poison: 1,
        Ground: 1,
        Rock: 1,
        Bug: 1,
        Ghost: 1,
        Steel: 1,
        Fire: 1,
        Water: 1,
        Grass: 1,
        Electric: 1,
        Psychic: 1,
        Ice: 1,
        Dragon: 1,
        Dark: 1,
        Fairy: 1,
    };

    for (const type of types) {
        const relationships = typeEffectiveness[type];
        for (const [key, value] of Object.entries(relationships)) {
            if (value === 0) {
                effectiveness[key] *= 0;
            } else if (value === 0.5) {
                effectiveness[key] *= 0.5;
            } else if (value === 2) {
                effectiveness[key] *= 2;
            }
        }
    }

    console.log(effectiveness);

    for (const [key, value] of Object.entries(effectiveness)) {
        if (value === 0) {
            noEffects.push({ [key]: value });
        } else if (value === 1) {
            neutrals.push({ [key]: value });
        } else if (value < 1) {
            strengths.push({ [key]: value });
        } else if (value > 1) {
            weaknesses.push({ [key]: value });
        }
    }

    console.log(strengths);

    return { strengths, weaknesses, neutrals, noEffects };
}
