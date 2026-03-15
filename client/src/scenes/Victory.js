// Victory.js - Сцена завершения уровня

class Victory extends Phaser.Scene {
    constructor() {
        super({ key: 'VictoryScene' });
    }
    
    init(data) {
        this.levelId = data.levelId;
        this.levelName = data.levelName;
        this.crystals = data.crystals;
        this.totalCrystals = data.totalCrystals;
        this.time = data.time;
        this.score = data.score;
    }
    
    create() {
        // Фон
        this.cameras.main.setBackgroundColor(0x0a0a1a);
        
        // Создаём элементы
        this.createBackground();
        this.createVictoryPanel();
        this.createStats();
        this.createButtons();
        this.createConfetti();
    }
    
    createBackground() {
        // Звёзды
        for (let i = 0; i < 150; i++) {
            const x = Math.random() * 800;
            const y = Math.random() * 600;
            const size = Math.random() * 2 + 1;
            this.add.circle(x, y, size, 0xffffff, Math.random() * 0.5 + 0.3);
        }
        
        // Сетка
        const graphics = this.add.graphics();
        graphics.lineStyle(1, 0x8a2be2, 0.2);
        for (let x = 0; x < 800; x += 40) {
            graphics.lineBetween(x, 0, x, 600);
        }
        for (let y = 0; y < 600; y += 40) {
            graphics.lineBetween(0, y, 800, y);
        }
    }
    
    createVictoryPanel() {
        // Панель победы
        const panel = this.add.container(400, 250);
        
        const bg = this.add.rectangle(0, 0, 500, 280, 0x000000, 0.9);
        bg.setStrokeStyle(4, 0x00ff88);
        
        // Заголовок
        const title = this.add.text(0, -100, '✅ LEVEL COMPLETE!', {
            fontSize: '32px',
            fontFamily: 'Courier New',
            color: '#00ff88',
            fontStyle: 'bold'
        });
        title.setOrigin(0.5);
        
        // Название уровня
        const levelName = this.add.text(0, -50, this.levelName, {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#8a2be2',
            fontStyle: 'bold'
        });
        levelName.setOrigin(0.5);
        
        // Иконка технологии
        const icons = {
            web: '🌐',
            data: '📊',
            cyber: '🔒',
            gamedev: '🎮',
            ai: '🤖'
        };
        const icon = this.add.text(0, -100, icons[this.levelId.split('_')[0]] || '🌟', {
            fontSize: '48px'
        });
        icon.setOrigin(0, 0.5);
        icon.setX(200);
        
        panel.add([bg, title, levelName, icon]);
    }
    
    createStats() {
        // Панель статистики
        const statsBg = this.add.rectangle(400, 420, 500, 120, 0x000000, 0.7);
        statsBg.setStrokeStyle(2, 0x8a2be2);
        
        // Кристаллы
        const crystalsText = this.add.text(150, 390, `💎 Собрано: ${this.crystals}/${this.totalCrystals}`, {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#00ffff'
        });
        crystalsText.setOrigin(0.5);
        
        // Время
        const minutes = Math.floor(this.time / 60);
        const seconds = this.time % 60;
        const timeText = this.add.text(400, 390, `⏱️ Время: ${minutes}:${seconds.toString().padStart(2, '0')}`, {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#ff8800'
        });
        timeText.setOrigin(0.5);
        
        // Счёт
        const scoreText = this.add.text(650, 390, `🏆 Счёт: ${this.score}`, {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#ffff00'
        });
        scoreText.setOrigin(0.5);
        
        // Рейтинг
        let rating = '⭐';
        if (this.crystals === this.totalCrystals && this.time < 60) {
            rating = '⭐⭐⭐';
        } else if (this.crystals >= this.totalCrystals * 0.8) {
            rating = '⭐⭐';
        }
        
        const ratingText = this.add.text(400, 430, rating, {
            fontSize: '36px'
        });
        ratingText.setOrigin(0.5);
    }
    
    createButtons() {
        // Кнопка "Продолжить"
        const continueBtn = this.add.text(280, 520, '▶️ Продолжить', {
            fontSize: '22px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            backgroundColor: '#00aa44',
            padding: { x: 25, y: 12 }
        });
        continueBtn.setOrigin(0.5);
        continueBtn.setInteractive({ useHandCursor: true });
        
        continueBtn.on('pointerover', () => {
            continueBtn.setStyle({ backgroundColor: '#00cc55' });
        });
        
        continueBtn.on('pointerout', () => {
            continueBtn.setStyle({ backgroundColor: '#00aa44' });
        });
        
        continueBtn.on('pointerdown', () => {
            this.returnToMenu();
        });
        
        // Кнопка "Выбрать другую технологию"
        const menuBtn = this.add.text(520, 520, '🗺️ Карта технологий', {
            fontSize: '22px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            backgroundColor: '#8a2be2',
            padding: { x: 25, y: 12 }
        });
        menuBtn.setOrigin(0.5);
        menuBtn.setInteractive({ useHandCursor: true });
        
        menuBtn.on('pointerover', () => {
            menuBtn.setStyle({ backgroundColor: '#9d3bf0' });
        });
        
        menuBtn.on('pointerout', () => {
            menuBtn.setStyle({ backgroundColor: '#8a2be2' });
        });
        
        menuBtn.on('pointerdown', () => {
            this.returnToMenu();
        });
    }
    
    createConfetti() {
        // Конфетти
        const colors = [0x00ff88, 0x00ffff, 0xff8800, 0xff00ff, 0xffff00];
        
        for (let i = 0; i < 5; i++) {
            const x = Math.random() * 800;
            const particles = this.add.particles(x, -20, 'particle', {
                speed: { min: 100, max: 200 },
                angle: { min: 60, max: 120 },
                scale: { start: 0.6, end: 0 },
                lifespan: { min: 2000, max: 3000 },
                quantity: 2,
                tint: colors[i % colors.length],
                gravityY: 200
            });
            
            // Запуск с задержкой
            this.time.delayedCall(i * 300, () => {
                particles.emitParticleAt(x, -20);
            });
        }
    }
    
    returnToMenu() {
        // Возврат к меню
        this.scene.start('MenuScene');
    }
}

window.VictoryScene = Victory;
