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

export const initialGameState: GameState = {
  score: 0,
  energy: 100,
  maxEnergy: 100,
  energyRegen: 1,
  tapValue: 1,
  autoTapPerSec: 0,
  level: 1,
  xp: 0,
  xpToNextLevel: 100,
};

export const upgrades: Upgrade[] = [
  { id: 'vars', name: 'Variables', baseCost: 15, income: 0.5, count: 0, icon: '📦', color: '#4CAF50' },
  { id: 'functions', name: 'Functions', baseCost: 100, income: 2, count: 0, icon: '⚡', color: '#2196F3' },
  { id: 'structs', name: 'Structs', baseCost: 500, income: 8, count: 0, icon: '🏗️', color: '#9C27B0' },
  { id: 'interfaces', name: 'Interfaces', baseCost: 2000, income: 25, count: 0, icon: '🔌', color: '#FF5722' },
  { id: 'goroutines', name: 'Goroutines', baseCost: 10000, income: 100, count: 0, icon: '🔄', color: '#FFC107' },
  { id: 'channels', name: 'Channels', baseCost: 50000, income: 400, count: 0, icon: '📡', color: '#00BCD4' },
];

export const goFacts: string[] = [
  'Go was created at Google in 2007',
  'Go compiles to native machine code',
  'Go has garbage collection built-in',
  'Goroutines are lightweight threads',
  'Go uses channels for communication',
  'Go interfaces are implicitly satisfied',
  'Go has no classes or inheritance',
  'Go errors are values',
  'Go fmt formats your code automatically',
  'Go supports concurrent programming',
  'Go has a simple syntax',
  'Go is great for microservices',
];

export const getUpgradeCost = (upgrade: Upgrade): number => {
  return Math.floor(upgrade.baseCost * Math.pow(1.15, upgrade.count));
};
