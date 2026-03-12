export interface GameState {
    score: number;
    energy: number;
    maxEnergy: number;
    energyRegen: number;
    tapValue: number;
    autoTapPerSec: number;
    level: number;
    xp: number;
    xpToNextLevel: number;
}
export interface Upgrade {
    id: string;
    name: string;
    baseCost: number;
    income: number;
    count: number;
    icon: string;
    color: string;
}
export declare const initialGameState: GameState;
export declare const upgrades: Upgrade[];
export declare const goFacts: string[];
export declare const getUpgradeCost: (upgrade: Upgrade) => number;
//# sourceMappingURL=types.d.ts.map