import * as THREE from 'three';
import { Game } from './game/scene.js';

// Game instance
let game = null;

// Canvas
const canvas = document.getElementById('game-canvas');

// UI Elements
const startScreen = document.getElementById('start-screen');
const gameOverScreen = document.getElementById('game-over-screen');
const leaderboardScreen = document.getElementById('leaderboard-screen');
const scoreValue = document.getElementById('score-value');
const levelValue = document.getElementById('level-value');
const healthFill = document.getElementById('health-fill');
const finalScore = document.getElementById('final-score');
const leaderboardList = document.getElementById('leaderboard-list');
const comboElement = document.getElementById('combo');
const comboValue = document.getElementById('combo-value');
const multiplierValue = document.getElementById('multiplier-value');

// Buttons
const startBtn = document.getElementById('start-btn');
const restartBtn = document.getElementById('restart-btn');
const saveScoreBtn = document.getElementById('save-score-btn');
const backBtn = document.getElementById('back-btn');

// Initialize game
function init() {
    console.log('🎮 Initializing Go Space Runner...');
    
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;

    game = new Game(canvas);
    game.init();
    
    console.log('✅ Game initialized');

    // Event listeners
    startBtn.addEventListener('click', startGame);
    restartBtn.addEventListener('click', restartGame);
    saveScoreBtn.addEventListener('click', saveScore);
    backBtn.addEventListener('click', showMainMenu);
    
    console.log('🔘 Start button:', startScreen.querySelector('#start-btn'));

    // Handle resize
    window.addEventListener('resize', onWindowResize);

    // Start render loop
    animate();
    
    console.log('🎬 Render loop started');
}

function startGame() {
    startScreen.classList.add('hidden');
    game.start();
}

function restartGame() {
    gameOverScreen.classList.add('hidden');
    game.restart();
}

function saveScore() {
    const score = game.getScore();
    const name = prompt('Enter your name:', 'Player');
    
    if (name) {
        fetch('http://localhost:8080/api/leaderboard', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ name, score }),
        })
        .then(response => response.json())
        .then(data => {
            alert('Score saved!');
            showLeaderboard();
        })
        .catch(error => {
            console.error('Error saving score:', error);
            alert('Failed to save score. Is the backend running?');
        });
    }
}

function showLeaderboard() {
    gameOverScreen.classList.add('hidden');
    leaderboardScreen.classList.remove('hidden');
    
    fetch('http://localhost:8080/api/leaderboard')
        .then(response => response.json())
        .then(data => {
            renderLeaderboard(data);
        })
        .catch(error => {
            console.error('Error loading leaderboard:', error);
            leaderboardList.innerHTML = '<p>Failed to load leaderboard</p>';
        });
}

function renderLeaderboard(scores) {
    leaderboardList.innerHTML = '';
    
    if (scores.length === 0) {
        leaderboardList.innerHTML = '<p>No scores yet. Be the first!</p>';
        return;
    }
    
    scores.forEach((entry, index) => {
        const item = document.createElement('div');
        item.className = 'leaderboard-item';
        item.innerHTML = `
            <span class="leaderboard-rank">#${index + 1}</span>
            <span class="leaderboard-name">${escapeHtml(entry.name)}</span>
            <span class="leaderboard-score">${entry.score}</span>
        `;
        leaderboardList.appendChild(item);
    });
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showMainMenu() {
    leaderboardScreen.classList.add('hidden');
    startScreen.classList.remove('hidden');
}

function onWindowResize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    game.onResize();
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
    if (game) {
        scoreValue.textContent = game.getScore();
        levelValue.textContent = game.getLevel();
        healthFill.style.width = `${game.getHealth()}%`;
        
        // Update combo
        const combo = game.getCombo();
        if (combo > 0) {
            comboElement.classList.remove('hidden');
            comboValue.textContent = combo;
        } else {
            comboElement.classList.add('hidden');
        }
        
        // Update multiplier
        multiplierValue.textContent = game.getMultiplier().toFixed(1);
        
        if (game.isGameOver()) {
            finalScore.textContent = game.getScore();
            gameOverScreen.classList.remove('hidden');
        }
    }
}

// Start the game
init();
