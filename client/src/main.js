// main.js - Точка входа игры

// Конфигурация Phaser
const config = {
    type: Phaser.AUTO,
    width: 800,
    height: 600,
    parent: 'game-container',
    backgroundColor: '#1a0a2e',
    physics: {
        default: 'arcade',
        arcade: {
            gravity: { y: 800 },
            debug: false
        }
    },
    scene: [
        BootScene,
        MenuScene,
        LevelScene,
        VictoryScene
    ],
    render: {
        pixelArt: true,
        antialias: false
    }
};

// Создание игры
const game = new Phaser.Game(config);

// Глобальные переменные для отладки
window.gameInstance = game;

// Обработчик ошибок
window.addEventListener('error', (e) => {
    console.error('Game Error:', e.error);
});

// Обработчик загрузки
window.addEventListener('load', () => {
    console.log('🟣 Purple Lord: Digital Odyssey loaded');
    console.log('Controls: WASD/Arrows to move, SPACE to jump, F to cast spell, ESC to pause');
});

// Service Worker для офлайн-работы (опционально)
if ('serviceWorker' in navigator) {
    // Можно добавить позже для PWA
}

// Сохранение при закрытии страницы
window.addEventListener('beforeunload', () => {
    // Автосохранение прогресса
    if (window.GameAPI && game.registry) {
        const progress = {
            totalCrystals: game.registry.get('totalCrystals') || 0,
            completedLevels: game.registry.get('completedLevels') || [],
            levelData: {},
            totalTime: game.registry.get('totalTime') || 0
        };
        window.GameAPI.saveProgress(progress);
    }
});
