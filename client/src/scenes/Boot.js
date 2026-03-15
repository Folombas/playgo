// Boot.js - Сцена загрузки

class Boot extends Phaser.Scene {
    constructor() {
        super({ key: 'BootScene' });
    }
    
    preload() {
        // Показываем прогресс загрузки
        const progressBar = this.add.graphics();
        const progressBox = this.add.graphics();
        progressBox.fillStyle(0x222222, 0.8);
        progressBox.fillRect(
            this.cameras.main.width / 2 - 160,
            this.cameras.main.height / 2 - 25,
            320,
            50
        );
        
        const loadingText = this.add.text(
            this.cameras.main.width / 2,
            this.cameras.main.height / 2 - 50,
            'Loading Assets...',
            {
                fontSize: '24px',
                fontFamily: 'Courier New',
                color: '#ffffff'
            }
        );
        loadingText.setOrigin(0.5);
        
        // Обработчики прогресса
        this.load.on('progress', (value) => {
            progressBar.clear();
            progressBar.fillStyle(0x8a2be2, 1);
            progressBar.fillRect(
                this.cameras.main.width / 2 - 150,
                this.cameras.main.height / 2 - 15,
                300 * value,
                30
            );
        });
        
        this.load.on('complete', () => {
            progressBar.destroy();
            progressBox.destroy();
            loadingText.destroy();
        });
        
        // === ЗАГРУЗКА АССЕТОВ ===
        
        // Спрайт игрока (программная генерация)
        this.createPlayerSprite();
        
        // Спрайты врагов
        this.createEnemySprites();
        
        // Кристаллы
        this.createCrystalSprites();
        
        // Частицы
        this.createParticleSprite();
        
        // Тайлсеты (программная генерация)
        this.createTilesets();
        
        // Звуки (синтезированные)
        this.createSounds();
    }
    
    createPlayerSprite() {
        // Создаём спрайт игрока программно
        const graphics = this.make.graphics({ x: 0, y: 0, add: false });
        
        // Тело (фиолетовое)
        graphics.fillStyle(0x8a2be2);
        graphics.fillCircle(16, 16, 14);
        
        // Глаза
        graphics.fillStyle(0xffffff);
        graphics.fillCircle(12, 14, 4);
        graphics.fillCircle(20, 14, 4);
        
        // Зрачки
        graphics.fillStyle(0x000000);
        graphics.fillCircle(12, 14, 2);
        graphics.fillCircle(20, 14, 2);
        
        // Мантия
        graphics.fillStyle(0x6a1b9a);
        graphics.fillRect(4, 24, 24, 8);
        
        graphics.generateTexture('player', 32, 32);
        graphics.clear();
        
        // Спрайт для заклинания (фаербол)
        graphics.fillStyle(0xff6600);
        graphics.fillCircle(8, 8, 7);
        graphics.fillStyle(0xffcc00);
        graphics.fillCircle(8, 8, 4);
        graphics.generateTexture('spell', 16, 16);
        graphics.clear();
    }
    
    createEnemySprites() {
        const graphics = this.make.graphics({ x: 0, y: 0, add: false });
        
        // Bug (красный)
        graphics.fillStyle(0xff4444);
        graphics.fillCircle(16, 16, 14);
        graphics.fillStyle(0x000000);
        graphics.fillRect(8, 12, 4, 4);
        graphics.fillRect(20, 12, 4, 4);
        graphics.lineStyle(2, 0x000000);
        graphics.strokeRect(6, 8, 20, 16);
        graphics.generateTexture('bug', 32, 32);
        graphics.clear();
        
        // Virus (зелёный)
        graphics.fillStyle(0x44ff44);
        graphics.fillCircle(16, 16, 12);
        for (let i = 0; i < 8; i++) {
            const angle = (i / 8) * Math.PI * 2;
            const x = 16 + Math.cos(angle) * 18;
            const y = 16 + Math.sin(angle) * 18;
            graphics.fillCircle(x, y, 4);
        }
        graphics.generateTexture('virus', 32, 32);
        graphics.clear();
        
        // Error (фиолетовый)
        graphics.fillStyle(0xff44ff);
        graphics.fillRect(4, 4, 24, 24);
        graphics.fillStyle(0x000000);
        graphics.fillRect(8, 10, 16, 3);
        graphics.fillRect(8, 15, 12, 3);
        graphics.generateTexture('error', 32, 32);
        graphics.clear();
        
        // Firewall (оранжевый)
        graphics.fillStyle(0xff8800);
        graphics.fillRect(4, 8, 24, 20);
        graphics.lineStyle(2, 0xffcc00);
        for (let i = 0; i < 5; i++) {
            graphics.moveTo(6 + i * 5, 8);
            graphics.lineTo(4 + i * 5, 2);
        }
        graphics.strokePath();
        graphics.generateTexture('firewall', 32, 32);
        graphics.clear();
    }
    
    createCrystalSprites() {
        const graphics = this.make.graphics({ x: 0, y: 0, add: false });
        
        // Обычный кристалл (голубой)
        graphics.fillStyle(0x00ffff);
        graphics.beginPath();
        graphics.moveTo(16, 2);
        graphics.lineTo(24, 12);
        graphics.lineTo(20, 28);
        graphics.lineTo(12, 28);
        graphics.lineTo(8, 12);
        graphics.closePath();
        graphics.fillPath();
        graphics.fillStyle(0xffffff, 0.6);
        graphics.fillCircle(14, 10, 3);
        graphics.generateTexture('crystal', 32, 32);
        graphics.clear();
        
        // Синий кристалл
        graphics.fillStyle(0x0088ff);
        graphics.beginPath();
        graphics.moveTo(16, 0);
        graphics.lineTo(26, 14);
        graphics.lineTo(22, 30);
        graphics.lineTo(10, 30);
        graphics.lineTo(6, 14);
        graphics.closePath();
        graphics.fillPath();
        graphics.generateTexture('blue_crystal', 32, 32);
        graphics.clear();
        
        // Фиолетовый кристалл (редкий)
        graphics.fillStyle(0x8a2be2);
        graphics.beginPath();
        graphics.moveTo(16, 0);
        graphics.lineTo(28, 16);
        graphics.lineTo(24, 32);
        graphics.lineTo(8, 32);
        graphics.lineTo(4, 16);
        graphics.closePath();
        graphics.fillPath();
        graphics.fillStyle(0xffffff, 0.5);
        graphics.fillCircle(16, 12, 4);
        graphics.generateTexture('purple_crystal', 32, 32);
        graphics.clear();
        
        // Здоровье (зелёное)
        graphics.fillStyle(0x00ff00);
        graphics.fillCircle(16, 16, 12);
        graphics.fillStyle(0xffffff);
        graphics.fillRect(14, 8, 4, 16);
        graphics.fillRect(8, 14, 16, 4);
        graphics.generateTexture('health', 32, 32);
        graphics.clear();
        
        // Мана (синяя)
        graphics.fillStyle(0x0088ff);
        graphics.fillCircle(16, 16, 12);
        graphics.fillStyle(0xffffff);
        graphics.fillCircle(16, 12, 4);
        graphics.fillCircle(12, 20, 3);
        graphics.fillCircle(20, 20, 3);
        graphics.generateTexture('mana', 32, 32);
        graphics.clear();
    }
    
    createParticleSprite() {
        const graphics = this.make.graphics({ x: 0, y: 0, add: false });
        graphics.fillStyle(0xffffff);
        graphics.fillCircle(4, 4, 4);
        graphics.generateTexture('particle', 8, 8);
    }
    
    createTilesets() {
        const graphics = this.make.graphics({ x: 0, y: 0, add: false });
        
        // Платформа (Web Dev стиль)
        graphics.fillStyle(0x2d4a22);
        graphics.fillRect(0, 0, 32, 32);
        graphics.lineStyle(2, 0x3d5a32);
        graphics.strokeRect(0, 0, 32, 32);
        graphics.generateTexture('platform', 32, 32);
        graphics.clear();
        
        // Земля
        graphics.fillStyle(0x1a1a2e);
        graphics.fillRect(0, 0, 32, 32);
        graphics.fillStyle(0x8a2be2);
        for (let i = 0; i < 10; i++) {
            graphics.fillCircle(
                Math.random() * 32,
                Math.random() * 32,
                1 + Math.random() * 2
            );
        }
        graphics.generateTexture('ground', 32, 32);
        graphics.clear();
        
        // Стена
        graphics.fillStyle(0x16213e);
        graphics.fillRect(0, 0, 32, 32);
        graphics.lineStyle(1, 0x0f3460);
        graphics.strokeRect(2, 2, 28, 28);
        graphics.generateTexture('wall', 32, 32);
        graphics.clear();
        
        // Портал
        for (let frame = 0; frame < 8; frame++) {
            graphics.clear();
            graphics.fillStyle(0x000000, 0.3);
            graphics.fillCircle(16, 16, 16);
            graphics.lineStyle(3 - frame * 0.3, 0x8a2be2, 0.8);
            graphics.strokeCircle(16, 16, 14 - frame);
            graphics.generateTexture(`portal_${frame}`, 32, 32);
        }
        graphics.clear();
        
        // Выход (разные стили для уровней)
        // Web Dev
        graphics.fillStyle(0x00ff88);
        graphics.fillRect(4, 0, 24, 32);
        graphics.fillStyle(0x00ff88, 0.5);
        graphics.fillRect(8, 4, 16, 28);
        graphics.generateTexture('exit_web', 32, 32);
        graphics.clear();
        
        // Data Science
        graphics.fillStyle(0x0088ff);
        graphics.fillCircle(16, 16, 14);
        graphics.fillStyle(0xffffff, 0.5);
        graphics.fillCircle(16, 16, 10);
        graphics.generateTexture('exit_data', 32, 32);
        graphics.clear();
        
        // Cybersecurity
        graphics.fillStyle(0xff4444);
        graphics.fillRect(4, 4, 24, 24);
        graphics.lineStyle(2, 0xff8888);
        graphics.strokeRect(4, 4, 24, 24);
        graphics.generateTexture('exit_cyber', 32, 32);
        graphics.clear();
        
        // GameDev
        graphics.fillStyle(0xff8800);
        graphics.fillTriangle(16, 2, 30, 28, 2, 28);
        graphics.fillStyle(0xffcc00);
        graphics.fillTriangle(16, 6, 26, 24, 6, 24);
        graphics.generateTexture('exit_gamedev', 32, 32);
        graphics.clear();
        
        // AI
        graphics.fillStyle(0xff00ff);
        graphics.fillCircle(16, 16, 14);
        graphics.lineStyle(2, 0xffffff);
        for (let i = 0; i < 8; i++) {
            const angle = (i / 8) * Math.PI * 2;
            const x1 = 16 + Math.cos(angle) * 6;
            const y1 = 16 + Math.sin(angle) * 6;
            const x2 = 16 + Math.cos(angle) * 14;
            const y2 = 16 + Math.sin(angle) * 14;
            graphics.lineBetween(x1, y1, x2, y2);
        }
        graphics.generateTexture('exit_ai', 32, 32);
        graphics.clear();
    }
    
    createSounds() {
        // Создаём звуковые эффекты программно (Web Audio API будет использоваться в игре)
        // Для простоты помечаем что звуки доступны
        this.soundAvailable = true;
    }
    
    create() {
        // Переходим к меню
        this.scene.start('MenuScene');
    }
}

window.BootScene = Boot;
