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
  totalTaps: number;
  criticalHits: number;
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

export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  unlocked: boolean;
}

export interface Quest {
  id: string;
  name: string;
  progress: number;
  target: number;
  reward: number;
  completed: boolean;
  type: 'daily' | 'quest';
}

export interface DailyBonus {
  day: number;
  reward: number;
  claimed: boolean;
}

export const initialGameState: GameState = {
  score: 0,
  energy: 100,
  maxEnergy: 100,
  energyRegen: 2,
  tapValue: 1,
  autoTapPerSec: 0,
  level: 1,
  xp: 0,
  xpToNextLevel: 50,
  totalTaps: 0,
  criticalHits: 0,
};

export const upgrades: Upgrade[] = [
  { id: 'vars', name: 'Переменные', baseCost: 15, income: 0.5, count: 0, icon: '📦', color: '#00FF88' },
  { id: 'functions', name: 'Функции', baseCost: 100, income: 2, count: 0, icon: '⚡', color: '#00BFFF' },
  { id: 'structs', name: 'Структуры', baseCost: 500, income: 8, count: 0, icon: '🏗️', color: '#DA70D6' },
  { id: 'interfaces', name: 'Интерфейсы', baseCost: 2000, income: 25, count: 0, icon: '🔌', color: '#FF1493' },
  { id: 'goroutines', name: 'Горутины', baseCost: 10000, income: 100, count: 0, icon: '🔄', color: '#FFD700' },
  { id: 'channels', name: 'Каналы', baseCost: 50000, income: 400, count: 0, icon: '📡', color: '#00FFFF' },
];

export const achievements: Achievement[] = [
  { id: 'first_blood', name: 'Первый тап', description: 'Сделай первый клик', icon: '🎯', unlocked: false },
  { id: 'combo_10', name: 'Комбо мастер', description: 'Достигни комбо x10', icon: '🔥', unlocked: false },
  { id: 'combo_25', name: 'Легенда комбо', description: 'Достигни комбо x25', icon: '💎', unlocked: false },
  { id: 'level_5', name: 'Новичок', description: 'Достигни 5 уровня', icon: '⭐', unlocked: false },
  { id: 'level_10', name: 'Опытный', description: 'Достигни 10 уровня', icon: '🏅', unlocked: false },
  { id: 'level_25', name: 'Мастер', description: 'Достигни 25 уровня', icon: '👑', unlocked: false },
  { id: 'rich', name: 'Богач', description: 'Накопи 1,000 монет', icon: '💰', unlocked: false },
  { id: 'millionaire', name: 'Миллионер', description: 'Накопи 1,000,000 монет', icon: '💎', unlocked: false },
  { id: 'critical_100', name: 'Критикан', description: 'Поймай 100 критов', icon: '⚡', unlocked: false },
  { id: 'taps_1000', name: 'Тап-машина', description: 'Сделай 1000 тапов', icon: '👆', unlocked: false },
];

export const dailyBonuses: DailyBonus[] = [
  { day: 1, reward: 100, claimed: false },
  { day: 2, reward: 200, claimed: false },
  { day: 3, reward: 350, claimed: false },
  { day: 4, reward: 500, claimed: false },
  { day: 5, reward: 750, claimed: false },
  { day: 6, reward: 1000, claimed: false },
  { day: 7, reward: 2500, claimed: false },
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
