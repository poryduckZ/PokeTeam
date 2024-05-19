import { typeStyles } from "../constants/typeStyles";

interface TypeChipProps {
    type: string;
    effectiveness?: number;
}

export function TypeChip({ type, effectiveness }: TypeChipProps) {
    return (
        <span
            className={`${typeStyles[type].backgroundColor} text-white font-medium py-1 px-3 rounded-full  mr-1 text-sm`}
        >
            {type.charAt(0).toUpperCase() + type.slice(1)}{" "}
            {effectiveness && `${effectiveness}x`}
        </span>
    );
}
