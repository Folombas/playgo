// HUD.js - Интерфейс (здоровье, мана, кристаллы, счёт)

class HUD extends Phaser.GameObjects.Container {
    constructor(scene) {
        super(scene, 0, 0);
        
        this.scene = scene;
        this.scene.add.existing(this);
        this.setScrollFactor(0);
        this.setDepth(1000);
        
        // Размеры
        this.width = scene.scale.width;
        this.height = 60;
        
        // Создаём элементы интерфейса
        this.createHealthBar();
        this.createManaBar();
        this.createCrystalCounter();
        this.createScoreCounter();
        this.createLevelInfo();
        
        // Начальные значения
        this.updateHealth(100);
        this.updateMana(100);
        this.updateCrystals(0);
        this.updateScore(0);
    }
    
    createHealthBar() {
        // Фон полосы здоровья
        this.healthBg = this.scene.add.rectangle(20, 20, 200, 20, 0x333333);
        this.healthBg.setOrigin(0, 0.5);
        this.healthBg.setScrollFactor(0);
        
        // Полоса здоровья
        this.healthBar = this.scene.add.rectangle(20, 20, 200, 20, 0x00ff00);
        this.healthBar.setOrigin(0, 0.5);
        this.healthBar.setScrollFactor(0);
        
        // Текст
        this.healthText = this.scene.add.text(20, 10, 'HP', {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#ffffff'
        });
        this.healthText.setOrigin(0, 0);
        this.healthText.setScrollFactor(0);
        
        this.healthValue = this.scene.add.text(220, 10, '100/100', {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#ffffff'
        });
        this.healthValue.setOrigin(1, 0);
        this.healthValue.setScrollFactor(0);
    }
    
    createManaBar() {
        // Фон полосы маны
        this.manaBg = this.scene.add.rectangle(20, 45, 200, 20, 0x333333);
        this.manaBg.setOrigin(0, 0.5);
        this.manaBg.setScrollFactor(0);
        
        // Полоса маны
        this.manaBar = this.scene.add.rectangle(20, 45, 200, 20, 0x0088ff);
        this.manaBar.setOrigin(0, 0.5);
        this.manaBar.setScrollFactor(0);
        
        // Текст
        this.manaText = this.scene.add.text(20, 35, 'MP', {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#ffffff'
        });
        this.manaText.setOrigin(0, 0);
        this.manaText.setScrollFactor(0);
        
        this.manaValue = this.scene.add.text(220, 35, '100/100', {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#ffffff'
        });
        this.manaValue.setOrigin(1, 0);
        this.manaValue.setScrollFactor(0);
    }
    
    createCrystalCounter() {
        // Иконка кристалла
        this.crystalIcon = this.scene.add.text(this.width - 180, 20, '💎', {
            fontSize: '24px'
        });
        this.crystalIcon.setOrigin(0.5);
        this.crystalIcon.setScrollFactor(0);
        
        // Счётчик кристаллов
        this.crystalCount = this.scene.add.text(this.width - 140, 20, '0', {
            fontSize: '28px',
            fontFamily: 'Courier New',
            color: '#00ffff',
            fontStyle: 'bold'
        });
        this.crystalCount.setOrigin(0.5, 0.5);
        this.crystalCount.setScrollFactor(0);
    }
    
    createScoreCounter() {
        // Счёт
        this.scoreText = this.scene.add.text(this.width - 80, 20, '0', {
            fontSize: '28px',
            fontFamily: 'Courier New',
            color: '#ffff00',
            fontStyle: 'bold'
        });
        this.scoreText.setOrigin(0.5, 0.5);
        this.scoreText.setScrollFactor(0);
        
        // Подпись
        this.scoreLabel = this.scene.add.text(this.width - 80, 45, 'SCORE', {
            fontSize: '12px',
            fontFamily: 'Courier New',
            color: '#888888'
        });
        this.scoreLabel.setOrigin(0.5, 0.5);
        this.scoreLabel.setScrollFactor(0);
    }
    
    createLevelInfo() {
        // Информация об уровне
        this.levelInfo = this.scene.add.text(this.width / 2, 20, 'Level 1', {
            fontSize: '20px',
            fontFamily: 'Courier New',
            color: '#8a2be2',
            fontStyle: 'bold'
        });
        this.levelInfo.setOrigin(0.5, 0.5);
        this.levelInfo.setScrollFactor(0);
        
        // Прогресс кристаллов
        this.crystalProgress = this.scene.add.text(this.width / 2, 40, '💎 0/0', {
            fontSize: '14px',
            fontFamily: 'Courier New',
            color: '#00ffff'
        });
        this.crystalProgress.setOrigin(0.5, 0.5);
        this.crystalProgress.setScrollFactor(0);
    }
    
    updateHealth(value) {
        const maxHealth = 100;
        const percent = value / maxHealth;
        
        // Обновляем ширину полосы
        this.healthBar.width = Math.max(0, 200 * percent);
        
        // Обновляем цвет в зависимости от здоровья
        if (percent > 0.6) {
            this.healthBar.setFillStyle(0x00ff00);
        } else if (percent > 0.3) {
            this.healthBar.setFillStyle(0xffff00);
        } else {
            this.healthBar.setFillStyle(0xff0000);
        }
        
        // Текст
        this.healthValue.setText(`${Math.floor(value)}/${maxHealth}`);
    }
    
    updateMana(value) {
        const maxMana = 100;
        const percent = value / maxMana;
        
        this.manaBar.width = Math.max(0, 200 * percent);
        this.manaValue.setText(`${Math.floor(value)}/${maxMana}`);
    }
    
    updateCrystals(count) {
        this.crystalCount.setText(count.toString());
    }
    
    updateScore(score) {
        this.scoreText.setText(score.toString());
    }
    
    updateLevelInfo(levelName, crystalsCollected, totalCrystals) {
        this.levelInfo.setText(levelName);
        this.crystalProgress.setText(`💎 ${crystalsCollected}/${totalCrystals}`);
    }
    
    showMessage(text, duration = 2000) {
        const message = this.scene.add.text(
            this.width / 2,
            this.height / 2,
            text,
            {
                fontSize: '32px',
                fontFamily: 'Courier New',
                color: '#ffffff',
                fontStyle: 'bold',
                stroke: '#000000',
                strokeThickness: 6
            }
        );
        message.setOrigin(0.5);
        message.setScrollFactor(0);
        message.setDepth(2000);
        
        // Анимация появления
        this.scene.tweens.add({
            targets: message,
            alpha: 0,
            delay: duration - 500,
            duration: 500,
            onComplete: () => message.destroy()
        });
    }
    
    showPortalUnlocked() {
        this.showMessage('🌀 PORTAL UNLOCKED!', 2000);
    }
    
    showLevelComplete() {
        this.showMessage('✅ LEVEL COMPLETE!', 2500);
    }
}

// Экспорт
window.HUD = HUD;
