import { Game } from './game/scene.js';

let game = null;
const canvas = document.getElementById('game-canvas');

// UI Elements
const startScreen = document.getElementById('start-screen');
const gameOverScreen = document.getElementById('game-over-screen');
const scoreValue = document.getElementById('score-value');
const healthFill = document.getElementById('health-fill');
const finalScore = document.getElementById('final-score');
const comboElement = document.getElementById('combo');
const comboValue = document.getElementById('combo-value');
const multiplierValue = document.getElementById('multiplier-value');

// Buttons
document.getElementById('start-btn').addEventListener('click', () => {
    startScreen.classList.add('hidden');
    game.start();
});

document.getElementById('restart-btn').addEventListener('click', () => {
    gameOverScreen.classList.add('hidden');
    game.restart();
});

document.getElementById('save-score-btn').addEventListener('click', saveScore);
document.getElementById('back-btn').addEventListener('click', () => {
    document.getElementById('leaderboard-screen').classList.add('hidden');
    startScreen.classList.remove('hidden');
});

function init() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    game = new Game(canvas);
    game.init();

    window.addEventListener('resize', () => game.onResize());
    animate();
}

function saveScore() {
    const score = game.getScore();
    const name = prompt('Enter your name:', 'Player');
    if (name) {
        fetch('http://localhost:8080/api/leaderboard', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, score }),
        })
        .then(r => r.json())
        .then(() => alert('Score saved!'))
        .catch(() => alert('Failed to save. Is backend running?'));
    }
}

function animate() {
    requestAnimationFrame(animate);

    if (game && game.isRunning) {
        game.update();
    }

    if (game) {
        game.render();
        updateHUD();
    }
}

function updateHUD() {
    if (!game) return;

    scoreValue.textContent = game.getScore();
    healthFill.style.width = `${game.getHealth()}%`;

    const combo = game.getCombo();
    if (combo > 0) {
        comboElement.classList.remove('hidden');
        comboValue.textContent = combo;
    } else {
        comboElement.classList.add('hidden');
    }

    multiplierValue.textContent = game.getMultiplier().toFixed(1);

    if (game.isGameOver()) {
        finalScore.textContent = game.getScore();
        gameOverScreen.classList.remove('hidden');
    }
}

init();
