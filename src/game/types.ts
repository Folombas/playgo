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
  { id: 'vars', name: 'Переменные', baseCost: 15, income: 0.5, count: 0, icon: '📦', color: '#4CAF50' },
  { id: 'functions', name: 'Функции', baseCost: 100, income: 2, count: 0, icon: '⚡', color: '#2196F3' },
  { id: 'structs', name: 'Структуры', baseCost: 500, income: 8, count: 0, icon: '🏗️', color: '#9C27B0' },
  { id: 'interfaces', name: 'Интерфейсы', baseCost: 2000, income: 25, count: 0, icon: '🔌', color: '#FF5722' },
  { id: 'goroutines', name: 'Горутины', baseCost: 10000, income: 100, count: 0, icon: '🔄', color: '#FFC107' },
  { id: 'channels', name: 'Каналы', baseCost: 50000, income: 400, count: 0, icon: '📡', color: '#00BCD4' },
];

export const goFacts: string[] = [
  'Go был создан в Google в 2007 году',
  'Go компилируется в нативный машинный код',
  'Go имеет встроенную сборку мусора',
  'Горутины — это лёгкие потоки',
  'Go использует каналы для коммуникации',
  'Интерфейсы в Go удовлетворяются неявно',
  'В Go нет классов или наследования',
  'Ошибки в Go — это значения',
  'Go fmt автоматически форматирует код',
  'Go поддерживает конкурентное программирование',
  'У Go простой и понятный синтаксис',
  'Go отлично подходит для микросервисов',
  'Go имеет мощную стандартную библиотеку',
  'В Go есть указатели, но нет арифметики указателей',
  'Go поддерживает обобщения (generics) с версии 1.18',
];

export const getUpgradeCost = (upgrade: Upgrade): number => {
  return Math.floor(upgrade.baseCost * Math.pow(1.15, upgrade.count));
};
