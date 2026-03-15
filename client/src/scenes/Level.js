// Level.js - Основная сцена уровня

class Level extends Phaser.Scene {
    constructor() {
        super({ key: 'LevelScene' });
    }
    
    init(data) {
        this.levelId = data.levelId || 'web_1';
        this.tech = data.tech || 'web';
        this.levelName = data.levelName || 'Web Development';
        this.crystals = 0;
        this.score = 0;
        this.totalCrystals = 0;
        this.crystalsCollected = 0;
        this.portalUnlocked = false;
        this.startTime = 0;
        this.spawnPoint = { x: 100, y: 400 };
    }
    
    create() {
        this.startTime = Date.now();
        
        // Настройки камеры
        this.cameras.main.setBounds(0, 0, 2000, 600);
        this.cameras.main.setBackgroundColor(this.getLevelColor());
        
        // Создаём уровень
        this.createLevel();
        
        // Игрок
        this.player = new Player(this, this.spawnPoint.x, this.spawnPoint.y);
        this.physics.add.collider(this.player, this.platforms);
        
        // Заклинания
        this.spells = this.physics.add.group({
            classType: Phaser.Physics.Arcade.Sprite,
            runChildUpdate: true
        });
        
        // Враги
        this.enemies = this.physics.add.group({
            classType: Enemy,
            runChildUpdate: true
        });
        this.createEnemies();
        
        // Кристаллы
        this.collectibles = this.physics.add.group({
            classType: Collectible,
            runChildUpdate: true
        });
        this.createCollectibles();
        
        // Портал
        this.createPortal();
        
        // Коллизии
        this.physics.add.collider(this.player, this.platforms);
        this.physics.add.collider(this.enemies, this.platforms);
        this.physics.add.collider(this.spells, this.platforms, this.hitWall, null, this);
        
        this.physics.add.overlap(this.player, this.collectibles, this.collectCrystal, null, this);
        this.physics.add.overlap(this.player, this.enemies, this.hitEnemy, null, this);
        this.physics.add.overlap(this.spells, this.enemies, this.hitEnemyWithSpell, null, this);
        this.physics.add.overlap(this.player, this.portal, this.enterPortal, null, this);
        
        // HUD
        this.hud = new HUD(this);
        this.hud.updateLevelInfo(this.levelName, 0, this.totalCrystals);
        
        // Управление
        this.cursors = this.input.keyboard.createCursorKeys();
        this.wasd = this.input.keyboard.addKeys({
            up: Phaser.Input.Keyboard.KeyCodes.W,
            left: Phaser.Input.Keyboard.KeyCodes.A,
            right: Phaser.Input.Keyboard.KeyCodes.D,
            f: Phaser.Input.Keyboard.KeyCodes.F,
            esc: Phaser.Input.Keyboard.KeyCodes.ESC
        });
        
        // Пауза
        this.isPaused = false;
        this.wasd.esc.on('down', () => this.togglePause());
        
        // Фоновая музыка (если есть)
        this.playBackgroundMusic();
    }
    
    getLevelColor() {
        const colors = {
            web: 0x1a2e1a,
            data: 0x0a1a2e,
            cyber: 0x2e0a0a,
            gamedev: 0x2e1a0a,
            ai: 0x2e0a2e
        };
        return colors[this.tech] || 0x1a1a2e;
    }
    
    createLevel() {
        // Платформы
        this.platforms = this.physics.add.staticGroup();
        
        // Генерация уровня в зависимости от технологии
        const levelData = this.getLevelData();
        
        // Земля
        for (let x = 0; x < 2000; x += 32) {
            // Пропуски для ям
            if (!(x > 600 && x < 750) && !(x > 1300 && x < 1450)) {
                const ground = this.platforms.create(x + 16, 584, 'ground');
                ground.setDisplaySize(32, 32);
                ground.refreshBody();
            }
        }
        
        // Платформы из уровня
        levelData.platforms.forEach(platform => {
            const p = this.platforms.create(platform.x, platform.y, 'platform');
            if (platform.width) {
                p.setDisplaySize(platform.width, 32);
                p.refreshBody();
            }
        });
        
        // Стены
        levelData.walls?.forEach(wall => {
            const w = this.platforms.create(wall.x, wall.y, 'wall');
            w.setDisplaySize(32, wall.height || 32);
            w.refreshBody();
        });
        
        // Декорации
        this.createDecorations(levelData);
    }
    
    getLevelData() {
        // Данные уровней
        const levels = {
            web_1: {
                platforms: [
                    { x: 300, y: 480, width: 150 },
                    { x: 500, y: 400, width: 100 },
                    { x: 700, y: 350, width: 120 },
                    { x: 900, y: 420, width: 100 },
                    { x: 1100, y: 350, width: 150 },
                    { x: 1400, y: 450, width: 100 },
                    { x: 1600, y: 380, width: 200 }
                ],
                walls: [
                    { x: 1984, y: 552, height: 32 }
                ],
                enemies: 5,
                crystals: 10,
                exitX: 1900,
                exitY: 520
            },
            web_2: {
                platforms: [
                    { x: 200, y: 500, width: 100 },
                    { x: 400, y: 450, width: 80 },
                    { x: 550, y: 380, width: 100 },
                    { x: 750, y: 320, width: 120 },
                    { x: 950, y: 400, width: 100 },
                    { x: 1150, y: 480, width: 80 },
                    { x: 1350, y: 350, width: 150 },
                    { x: 1600, y: 420, width: 100 },
                    { x: 1800, y: 500, width: 150 }
                ],
                enemies: 8,
                crystals: 15,
                exitX: 1900,
                exitY: 520
            },
            data_1: {
                platforms: [
                    { x: 250, y: 470, width: 120 },
                    { x: 450, y: 400, width: 100 },
                    { x: 650, y: 330, width: 140 },
                    { x: 900, y: 400, width: 100 },
                    { x: 1100, y: 470, width: 120 },
                    { x: 1350, y: 380, width: 150 },
                    { x: 1600, y: 450, width: 100 }
                ],
                enemies: 6,
                crystals: 12,
                exitX: 1900,
                exitY: 520
            },
            cyber_1: {
                platforms: [
                    { x: 300, y: 500, width: 100 },
                    { x: 500, y: 420, width: 80 },
                    { x: 700, y: 350, width: 100 },
                    { x: 900, y: 280, width: 120 },
                    { x: 1150, y: 350, width: 100 },
                    { x: 1400, y: 420, width: 150 },
                    { x: 1700, y: 500, width: 100 }
                ],
                walls: [
                    { x: 800, y: 520, height: 64 },
                    { x: 1200, y: 520, height: 64 }
                ],
                enemies: 7,
                crystals: 14,
                exitX: 1900,
                exitY: 520
            },
            gamedev_1: {
                platforms: [
                    { x: 200, y: 480, width: 100 },
                    { x: 400, y: 400, width: 100 },
                    { x: 600, y: 320, width: 100 },
                    { x: 800, y: 400, width: 100 },
                    { x: 1000, y: 480, width: 100 },
                    { x: 1200, y: 380, width: 150 },
                    { x: 1450, y: 300, width: 100 },
                    { x: 1650, y: 400, width: 100 }
                ],
                enemies: 9,
                crystals: 18,
                exitX: 1900,
                exitY: 520
            },
            ai_1: {
                platforms: [
                    { x: 250, y: 450, width: 120 },
                    { x: 500, y: 380, width: 100 },
                    { x: 750, y: 300, width: 140 },
                    { x: 1000, y: 380, width: 100 },
                    { x: 1250, y: 450, width: 120 },
                    { x: 1500, y: 350, width: 150 },
                    { x: 1750, y: 450, width: 100 }
                ],
                enemies: 10,
                crystals: 20,
                exitX: 1900,
                exitY: 520
            }
        };
        
        return levels[this.levelId] || levels.web_1;
    }
    
    createDecorations(levelData) {
        // Декоративные элементы в зависимости от технологии
        const graphics = this.add.graphics();
        
        // Сетка на фоне
        graphics.lineStyle(1, 0xffffff, 0.1);
        for (let x = 0; x < 2000; x += 100) {
            graphics.lineBetween(x, 0, x, 600);
        }
        for (let y = 0; y < 600; y += 100) {
            graphics.lineBetween(0, y, 2000, y);
        }
        
        // Тематические элементы
        if (this.tech === 'data') {
            // Графики
            for (let i = 0; i < 10; i++) {
                const x = 200 + i * 180;
                const height = 50 + Math.random() * 100;
                graphics.fillStyle(0x0088ff, 0.3);
                graphics.fillRect(x, 550 - height, 30, height);
            }
        } else if (this.tech === 'cyber') {
            // Бинарный код
            for (let i = 0; i < 20; i++) {
                const x = 100 + i * 100;
                const y = 100 + Math.random() * 300;
                const code = Math.random() > 0.5 ? '1' : '0';
                const text = this.add.text(x, y, code, {
                    fontSize: '20px',
                    fontFamily: 'Courier New',
                    color: '#00ff00',
                    alpha: 0.5
                });
            }
        } else if (this.tech === 'ai') {
            // Нейроны
            for (let i = 0; i < 15; i++) {
                const x = 200 + Math.random() * 1600;
                const y = 150 + Math.random() * 300;
                graphics.fillStyle(0xff00ff, 0.5);
                graphics.fillCircle(x, y, 8);
                
                // Связи
                for (let j = 0; j < 3; j++) {
                    const x2 = x + (Math.random() - 0.5) * 200;
                    const y2 = y + (Math.random() - 0.5) * 200;
                    graphics.lineStyle(2, 0xff00ff, 0.3);
                    graphics.lineBetween(x, y, x2, y2);
                }
            }
        }
    }
    
    createEnemies() {
        const levelData = this.getLevelData();
        const enemyTypes = ['bug', 'virus', 'error', 'firewall'];
        
        // Позиции врагов
        const positions = [
            { x: 400, y: 450 },
            { x: 600, y: 350 },
            { x: 800, y: 300 },
            { x: 1000, y: 400 },
            { x: 1200, y: 300 },
            { x: 1400, y: 400 },
            { x: 1600, y: 350 },
            { x: 1800, y: 500 }
        ];
        
        for (let i = 0; i < Math.min(levelData.enemies, positions.length); i++) {
            const pos = positions[i];
            const type = enemyTypes[Math.floor(Math.random() * enemyTypes.length)];
            const enemy = new Enemy(this, pos.x, pos.y, type);
            this.enemies.add(enemy);
        }
    }
    
    createCollectibles() {
        const levelData = this.getLevelData();
        this.totalCrystals = levelData.crystals;
        
        // Позиции кристаллов
        const positions = [
            { x: 300, y: 440 },
            { x: 350, y: 440 },
            { x: 520, y: 360 },
            { x: 560, y: 360 },
            { x: 750, y: 310 },
            { x: 800, y: 310 },
            { x: 920, y: 380 },
            { x: 1150, y: 310 },
            { x: 1200, y: 310 },
            { x: 1450, y: 410 },
            { x: 1650, y: 340 },
            { x: 1700, y: 340 },
            { x: 1750, y: 340 },
            { x: 1800, y: 460 },
            { x: 1850, y: 460 }
        ];
        
        const crystalTypes = ['crystal', 'crystal', 'crystal', 'blue_crystal', 'purple_crystal'];
        
        for (let i = 0; i < Math.min(levelData.crystals, positions.length); i++) {
            const pos = positions[i];
            const type = crystalTypes[Math.floor(Math.random() * crystalTypes.length)];
            const crystal = new Collectible(this, pos.x, pos.y, type);
            this.collectibles.add(crystal);
        }
        
        // Бонусы (здоровье, мана)
        const bonusPositions = [
            { x: 500, y: 520 },
            { x: 1000, y: 520 },
            { x: 1500, y: 520 }
        ];
        
        bonusPositions.forEach(pos => {
            const type = Math.random() > 0.5 ? 'health' : 'mana';
            const bonus = new Collectible(this, pos.x, pos.y, type);
            this.collectibles.add(bonus);
        });
    }
    
    createPortal() {
        // Портал (выход)
        const levelData = this.getLevelData();
        this.portal = this.physics.add.sprite(
            levelData.exitX || 1900,
            levelData.exitY || 520,
            'portal_0'
        );
        this.portal.setImmovable(true);
        this.portal.body.allowGravity = false;
        this.portal.setVisible(false);
        this.portal.body.enable = false;
        
        // Анимация портала
        this.portalFrame = 0;
        this.time.addEvent({
            delay: 150,
            callback: () => {
                this.portalFrame = (this.portalFrame + 1) % 8;
                this.portal.setTexture(`portal_${this.portalFrame}`);
            },
            loop: true
        });
    }
    
    playBackgroundMusic() {
        // Здесь можно добавить фоновую музыку
        // Для простоты пока без музыки
    }
    
    update(time, delta) {
        if (this.isPaused) return;
        
        // Обновление игрока
        const combinedCursors = {
            left: this.cursors.left.isDown || this.wasd.left.isDown,
            right: this.cursors.right.isDown || this.wasd.right.isDown,
            up: this.cursors.up.isDown || this.wasd.up.isDown,
            space: this.cursors.space.isDown,
            f: this.wasd.f.isDown
        };
        
        this.player.update(combinedCursors, time);
        
        // Обновление заклинаний
        this.spells.getChildren().forEach(spell => {
            if (spell.active) {
                spell.x += spell.direction * 400 * (delta / 1000);
                
                // Удаление за пределами
                if (Math.abs(spell.x - this.player.x) > 800) {
                    spell.disableBody(true, true);
                }
            }
        });
        
        // Проверка разблокировки портала
        if (!this.portalUnlocked && this.crystalsCollected >= this.totalCrystals) {
            this.unlockPortal();
        }
        
        // Проверка падения в пропасть
        if (this.player.y > 600) {
            this.player.die();
        }
    }
    
    hitWall(spell, wall) {
        spell.disableBody(true, true);
        
        // Эффект попадания
        const particles = this.add.particles(spell.x, spell.y, 'particle', {
            speed: { min: 50, max: 150 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.4, end: 0 },
            lifespan: 300,
            quantity: 8,
            tint: 0xff6600
        });
        
        setTimeout(() => particles.destroy(), 300);
    }
    
    collectCrystal(player, crystal) {
        if (!crystal.active) return;
        
        const value = crystal.collect(player);
        this.crystalsCollected++;
        this.hud.updateCrystals(this.crystalsCollected);
        this.hud.updateLevelInfo(this.levelName, this.crystalsCollected, this.totalCrystals);
    }
    
    hitEnemy(player, enemy) {
        if (!enemy.isAlive || player.isInvulnerable) return;
        
        // Урон от врага
        player.takeDamage(enemy.damage);
    }
    
    hitEnemyWithSpell(spell, enemy) {
        if (!enemy.isAlive) return;
        
        enemy.takeDamage(50, true);
        spell.disableBody(true, true);
    }
    
    unlockPortal() {
        this.portalUnlocked = true;
        this.portal.setVisible(true);
        this.portal.body.enable = true;
        
        // Эффект
        const particles = this.add.particles(this.portal.x, this.portal.y, 'particle', {
            speed: { min: 50, max: 200 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.5, end: 0 },
            lifespan: 600,
            quantity: 20,
            tint: 0x8a2be2
        });
        
        setTimeout(() => particles.destroy(), 600);
        
        this.hud.showPortalUnlocked();
    }
    
    enterPortal(player, portal) {
        if (!this.portalUnlocked) return;
        
        // Завершение уровня
        this.completeLevel();
    }
    
    completeLevel() {
        this.isPaused = true;
        
        const timeSpent = Math.floor((Date.now() - this.startTime) / 1000);
        
        // Сохранение прогресса
        const progressData = {
            totalCrystals: (this.registry.get('totalCrystals') || 0) + this.crystalsCollected,
            completedLevels: [...(this.registry.get('completedLevels') || []), this.levelId],
            levelData: {
                [this.levelId]: {
                    crystalsCollected: this.crystalsCollected,
                    completed: true,
                    bestTime: timeSpent
                }
            },
            totalTime: (this.registry.get('totalTime') || 0) + timeSpent
        };
        
        // Сохраняем в глобальный registry
        this.registry.set('totalCrystals', progressData.totalCrystals);
        this.registry.set('completedLevels', progressData.completedLevels);
        this.registry.set('totalTime', progressData.totalTime);
        
        // Сохраняем на сервер
        if (window.GameAPI) {
            window.GameAPI.saveProgress(progressData);
        }
        
        // Переход к сцене победы
        this.scene.start('VictoryScene', {
            levelId: this.levelId,
            levelName: this.levelName,
            crystals: this.crystalsCollected,
            totalCrystals: this.totalCrystals,
            time: timeSpent,
            score: this.score
        });
    }
    
    togglePause() {
        this.isPaused = !this.isPaused;
        
        if (this.isPaused) {
            // Показываем меню паузы
            const pauseBg = this.add.rectangle(
                this.cameras.main.scrollX + 400,
                300,
                400,
                300,
                0x000000,
                0.8
            );
            pauseBg.setStrokeStyle(3, 0x8a2be2);
            pauseBg.setScrollFactor(0);
            pauseBg.setDepth(300);
            
            const pauseText = this.add.text(
                this.cameras.main.scrollX + 400,
                250,
                '⏸️ PAUSE',
                {
                    fontSize: '36px',
                    fontFamily: 'Courier New',
                    color: '#ffffff',
                    fontStyle: 'bold'
                }
            );
            pauseText.setOrigin(0.5);
            pauseText.setScrollFactor(0);
            pauseText.setDepth(300);
            
            const resumeText = this.add.text(
                this.cameras.main.scrollX + 400,
                320,
                'Нажмите ESC для продолжения',
                {
                    fontSize: '18px',
                    fontFamily: 'Courier New',
                    color: '#888888'
                }
            );
            resumeText.setOrigin(0.5);
            resumeText.setScrollFactor(0);
            resumeText.setDepth(300);
            
            this.pauseObjects = [pauseBg, pauseText, resumeText];
        } else {
            // Скрываем меню паузы
            if (this.pauseObjects) {
                this.pauseObjects.forEach(obj => obj.destroy());
                this.pauseObjects = null;
            }
        }
    }
}

window.LevelScene = Level;
