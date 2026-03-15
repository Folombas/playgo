// API модуль для взаимодействия с Go-сервером

const API_BASE_URL = 'http://localhost:3000/api';

// Генерация уникального ID игрока
function generatePlayerId() {
    const saved = localStorage.getItem('purpleLordPlayerId');
    if (saved) return saved;
    
    const newId = 'player_' + Math.random().toString(36).substr(2, 9);
    localStorage.setItem('purpleLordPlayerId', newId);
    return newId;
}

const playerId = generatePlayerId();

// Загрузка прогресса
export async function loadProgress() {
    try {
        const response = await fetch(`${API_BASE_URL}/progress?playerId=${playerId}`);
        if (!response.ok) throw new Error('Failed to load progress');
        return await response.json();
    } catch (error) {
        console.warn('Could not load progress, using default:', error);
        return {
            playerId: playerId,
            totalCrystals: 0,
            completedLevels: [],
            levelData: {},
            totalTime: 0
        };
    }
}

// Сохранение прогресса
export async function saveProgress(progress) {
    try {
        const response = await fetch(`${API_BASE_URL}/save`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                ...progress,
                playerId: playerId,
                lastSave: new Date().toISOString()
            })
        });
        if (!response.ok) throw new Error('Failed to save progress');
        return await response.json();
    } catch (error) {
        console.error('Could not save progress:', error);
        return null;
    }
}

// Получение таблицы лидеров
export async function getLeaderboard() {
    try {
        const response = await fetch(`${API_BASE_URL}/leaderboard`);
        if (!response.ok) throw new Error('Failed to load leaderboard');
        return await response.json();
    } catch (error) {
        console.warn('Could not load leaderboard:', error);
        return [];
    }
}

// Экспорт для использования в сценах Phaser
window.GameAPI = {
    loadProgress,
    saveProgress,
    getLeaderboard,
    getPlayerId: () => playerId
};
