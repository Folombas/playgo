// Menu.js - Карта уровней (Древо технологий)

class Menu extends Phaser.Scene {
    constructor() {
        super({ key: 'MenuScene' });
    }
    
    init() {
        this.progress = {
            totalCrystals: 0,
            completedLevels: [],
            levelData: {}
        };
        this.selectedLevel = null;
    }
    
    async create() {
        // Загружаем прогресс
        if (window.GameAPI) {
            this.progress = await window.GameAPI.loadProgress();
        }
        
        // Фон
        this.createBackground();
        
        // Заголовок
        this.createTitle();
        
        // Древо технологий
        this.createTechTree();
        
        // Статистика
        this.createStats();
        
        // Таблица лидеров
        this.createLeaderboardButton();
        
        // Управление
        this.createControls();
    }
    
    createBackground() {
        // Градиентный фон
        const graphics = this.add.graphics();
        
        const gradient = graphics.generateTexture('menuBg', 800, 600);
        
        // Звёзды
        for (let i = 0; i < 100; i++) {
            const x = Math.random() * 800;
            const y = Math.random() * 600;
            const size = Math.random() * 2 + 1;
            const star = this.add.circle(x, y, size, 0xffffff, Math.random() * 0.5 + 0.3);
        }
        
        // Сетка
        graphics.lineStyle(1, 0x8a2be2, 0.2);
        for (let x = 0; x < 800; x += 40) {
            graphics.lineBetween(x, 0, x, 600);
        }
        for (let y = 0; y < 600; y += 40) {
            graphics.lineBetween(0, y, 800, y);
        }
    }
    
    createTitle() {
        // Заголовок
        const title = this.add.text(400, 50, '🟣 Purple Lord', {
            fontSize: '48px',
            fontFamily: 'Courier New',
            color: '#8a2be2',
            fontStyle: 'bold',
            stroke: '#000000',
            strokeThickness: 8
        });
        title.setOrigin(0.5);
        
        const subtitle = this.add.text(400, 90, 'Digital Odyssey', {
            fontSize: '24px',
            fontFamily: 'Courier New',
            color: '#00ffff',
            fontStyle: 'italic'
        });
        subtitle.setOrigin(0.5);
        
        // Анимация пульсации
        this.tweens.add({
            targets: title,
            scale: { from: 1, to: 1.05 },
            duration: 1500,
            yoyo: true,
            repeat: -1,
            ease: 'Sine.easeInOut'
        });
    }
    
    createTechTree() {
        // Древо технологий - узлы
        this.techNodes = [
            {
                id: 'web',
                name: 'Web Development',
                x: 400,
                y: 180,
                levels: ['web_1', 'web_2'],
                color: 0x00ff88,
                icon: '🌐',
                description: 'HTML, CSS, JavaScript'
            },
            {
                id: 'data',
                name: 'Data Science',
                x: 250,
                y: 300,
                levels: ['data_1'],
                color: 0x0088ff,
                icon: '📊',
                description: 'Python, ML, Analytics'
            },
            {
                id: 'cyber',
                name: 'Cybersecurity',
                x: 550,
                y: 300,
                levels: ['cyber_1'],
                color: 0xff4444,
                icon: '🔒',
                description: 'Security, Ethics, Protection'
            },
            {
                id: 'gamedev',
                name: 'Game Development',
                x: 320,
                y: 420,
                levels: ['gamedev_1'],
                color: 0xff8800,
                icon: '🎮',
                description: 'Unity, Unreal, Indie'
            },
            {
                id: 'ai',
                name: 'Artificial Intelligence',
                x: 480,
                y: 420,
                levels: ['ai_1'],
                color: 0xff00ff,
                icon: '🤖',
                description: 'Neural Networks, Deep Learning'
            }
        ];
        
        // Соединительные линии
        const graphics = this.add.graphics();
        graphics.lineStyle(3, 0x8a2be2, 0.5);
        
        // Центр к узлам
        this.techNodes.forEach(node => {
            graphics.lineBetween(400, 280, node.x, node.y);
        });
        
        // Создаём узлы
        this.techNodes.forEach(node => {
            this.createTechNode(node);
        });
        
        // Центральный узел (старт)
        const centerNode = this.add.circle(400, 280, 30, 0x8a2be2);
        centerNode.setStrokeStyle(3, 0xffffff);
        
        const centerIcon = this.add.text(400, 280, '🚀', { fontSize: '24px' });
        centerIcon.setOrigin(0.5);
    }
    
    createTechNode(node) {
        const container = this.add.container(node.x, node.y);

        // Круг узла (он же зона клика)
        const circle = this.add.circle(0, 0, 35, node.color, 0.3);
        circle.setStrokeStyle(3, node.color);
        circle.setInteractive({ useHandCursor: true });

        // Иконка
        const icon = this.add.text(0, 0, node.icon, { fontSize: '28px' });
        icon.setOrigin(0.5);
        icon.setDepth(10);

        // Название
        const name = this.add.text(0, 45, node.name, {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            align: 'center'
        });
        name.setOrigin(0.5);
        name.setDepth(10);

        // Проверка доступности
        const isUnlocked = this.progress.completedLevels.length > 0 || node.id === 'web';
        if (!isUnlocked) {
            circle.setAlpha(0.5);
            icon.setAlpha(0.5);
            name.setAlpha(0.5);
        }

        container.add([circle, icon, name]);

        // Функция клика
        const handleClick = () => {
            console.log('🔵 Клик по технологии:', node.name, 'isUnlocked:', isUnlocked);
            if (isUnlocked) {
                console.log('✅ Разблокировано, открываем панель');
                // Визуальный эффект клика
                this.tweens.add({
                    targets: circle,
                    scaleX: 0.9,
                    scaleY: 0.9,
                    duration: 100,
                    yoyo: true,
                    onComplete: () => {
                        console.log('🚀 Запуск selectTechnology для:', node.id);
                        this.selectTechnology(node);
                    }
                });
            } else {
                console.log('❌ Заблокировано');
            }
        };

        // Интерактивность только на круге
        circle.on('pointerover', () => {
            if (isUnlocked) {
                this.tweens.add({
                    targets: circle,
                    scaleX: 1.2,
                    scaleY: 1.2,
                    duration: 200
                });
                name.setStyle({ color: '#' + node.color.toString(16).padStart(6, '0') });
            }
        });

        circle.on('pointerout', () => {
            if (isUnlocked) {
                this.tweens.add({
                    targets: circle,
                    scaleX: 1,
                    scaleY: 1,
                    duration: 200
                });
                name.setStyle({ color: '#ffffff' });
            }
        });

        circle.on('pointerdown', handleClick);

        node.container = container;
    }
    
    selectTechnology(tech) {
        console.log('📋 selectTechnology:', tech.name);
        
        // Показываем уровни технологии
        if (this.levelPanel) {
            this.levelPanel.destroy();
        }

        this.levelPanel = this.add.container(400, 550);
        this.levelPanel.setDepth(100);

        // Фон панели
        const bg = this.add.rectangle(0, 0, 500, 120, 0x000000, 0.9);
        bg.setStrokeStyle(3, tech.color);
        bg.setOrigin(0.5);

        // Название
        const title = this.add.text(-240, -40, `${tech.icon} ${tech.name}`, {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            fontStyle: 'bold'
        });
        title.setOrigin(0, 0.5);

        // Описание
        const desc = this.add.text(-240, -15, tech.description, {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#888888'
        });
        desc.setOrigin(0, 0.5);

        console.log('🔘 Создаю кнопки для уровней:', tech.levels);
        
        // Кнопки уровней
        tech.levels.forEach((levelId, index) => {
            const isCompleted = this.progress.completedLevels.includes(levelId);
            const btnText = isCompleted ? `✅ ${levelId}` : `▶️ ${levelId}`;
            const btnColor = isCompleted ? '#00ff00' : '#ffffff';
            
            console.log(`  - Уровень ${levelId}: ${btnText}`);
            
            const btn = this.add.text(-240 + index * 130, 20, btnText, {
                fontSize: '16px',
                fontFamily: 'Courier New',
                color: btnColor,
                backgroundColor: '#333333',
                padding: { x: 15, y: 8 }
            });
            btn.setOrigin(0, 0.5);
            btn.setInteractive({ useHandCursor: true });

            btn.on('pointerover', () => {
                btn.setStyle({ backgroundColor: '#' + tech.color.toString(16).padStart(6, '0') });
            });

            btn.on('pointerout', () => {
                btn.setStyle({ backgroundColor: '#333333' });
            });

            btn.on('pointerdown', () => {
                console.log('🎮 Запуск уровня:', levelId);
                this.startLevel(levelId, tech);
            });

            this.levelPanel.add(btn);
        });

        // Кнопка назад
        const backBtn = this.add.text(240, 20, '❌ Назад', {
            fontSize: '16px',
            fontFamily: 'Courier New',
            color: '#ff4444',
            backgroundColor: '#333333',
            padding: { x: 15, y: 8 }
        });
        backBtn.setOrigin(0, 0.5);
        backBtn.setInteractive({ useHandCursor: true });
        backBtn.on('pointerdown', () => {
            this.levelPanel.destroy();
        });

        this.levelPanel.add([bg, title, desc, backBtn]);
        
        console.log('✅ Панель создана, элементов:', this.levelPanel.list.length);
    }
    
    startLevel(levelId, tech) {
        // Сохраняем выбранный уровень
        this.registry.set('currentLevel', levelId);
        this.registry.set('levelTech', tech.id);
        this.registry.set('levelName', tech.name);
        
        // Переход к уровню
        this.scene.start('LevelScene', {
            levelId: levelId,
            tech: tech.id,
            levelName: tech.name
        });
    }
    
    createStats() {
        // Статистика игрока
        const statsBg = this.add.rectangle(120, 550, 220, 100, 0x000000, 0.6);
        statsBg.setStrokeStyle(2, 0x8a2be2);
        
        const statsTitle = this.add.text(20, 510, '📊 Статистика', {
            fontSize: '16px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            fontStyle: 'bold'
        });
        
        const crystals = this.add.text(20, 535, `💎 ${this.progress.totalCrystals}`, {
            fontSize: '18px',
            fontFamily: 'Courier New',
            color: '#00ffff'
        });
        
        const completed = this.add.text(20, 560, `✅ Уровней: ${this.progress.completedLevels.length}`, {
            fontSize: '16px',
            fontFamily: 'Courier New',
            color: '#00ff00'
        });
        
        const totalTime = Math.floor(this.progress.totalTime || 0);
        const hours = Math.floor(totalTime / 3600);
        const minutes = Math.floor((totalTime % 3600) / 60);
        const timeText = this.add.text(20, 585, `⏱️ Время: ${hours}ч ${minutes}м`, {
            fontSize: '16px',
            fontFamily: 'Courier New',
            color: '#ff8800'
        });
    }
    
    createLeaderboardButton() {
        const btn = this.add.text(680, 550, '🏆 Топ', {
            fontSize: '18px',
            fontFamily: 'Courier New',
            color: '#ffff00',
            backgroundColor: '#333333',
            padding: { x: 20, y: 10 }
        });
        btn.setInteractive({ useHandCursor: true });
        
        btn.on('pointerover', () => {
            btn.setStyle({ backgroundColor: '#8a2be2' });
        });
        
        btn.on('pointerout', () => {
            btn.setStyle({ backgroundColor: '#333333' });
        });
        
        btn.on('pointerdown', () => {
            this.showLeaderboard();
        });
    }
    
    async showLeaderboard() {
        const leaderboard = await window.GameAPI.getLeaderboard();
        
        // Панель лидерборда
        const panel = this.add.container(400, 300);
        
        const bg = this.add.rectangle(0, 0, 400, 350, 0x000000, 0.9);
        bg.setStrokeStyle(3, 0xffd700);
        
        const title = this.add.text(0, -150, '🏆 Таблица лидеров', {
            fontSize: '24px',
            fontFamily: 'Courier New',
            color: '#ffd700',
            fontStyle: 'bold'
        });
        title.setOrigin(0.5);
        
        // Список игроков
        if (leaderboard.length === 0) {
            const empty = this.add.text(0, 0, 'Пока нет игроков', {
                fontSize: '18px',
                fontFamily: 'Courier New',
                color: '#888888'
            });
            empty.setOrigin(0.5);
            panel.add(empty);
        } else {
            leaderboard.forEach((entry, index) => {
                const medal = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : `#${index + 1}`;
                const text = this.add.text(0, -110 + index * 35, 
                    `${medal} ${entry.playerId}: ${entry.totalCrystals} 💎`, 
                    {
                        fontSize: '16px',
                        fontFamily: 'Courier New',
                        color: index < 3 ? '#ffd700' : '#ffffff'
                    }
                );
                text.setOrigin(0.5);
                panel.add(text);
            });
        }
        
        // Кнопка закрытия
        const closeBtn = this.add.text(0, 150, '❌ Закрыть', {
            fontSize: '18px',
            fontFamily: 'Courier New',
            color: '#ffffff',
            backgroundColor: '#333333',
            padding: { x: 15, y: 8 }
        });
        closeBtn.setOrigin(0.5);
        closeBtn.setInteractive({ useHandCursor: true });
        closeBtn.on('pointerdown', () => {
            panel.destroy();
        });
        
        panel.add([bg, title, closeBtn]);
        panel.setDepth(200);
    }
    
    createControls() {
        const controls = this.add.text(400, 580, 
            'Выберите технологию для начала уровня', 
            {
                fontSize: '16px',
                fontFamily: 'Courier New',
                color: '#888888',
                fontStyle: 'italic'
            }
        );
        controls.setOrigin(0.5);
    }
}

window.MenuScene = Menu;
