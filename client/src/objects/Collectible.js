// Collectible.js - Собираемые предметы (Кристаллы, Бонусы)

class Collectible extends Phaser.Physics.Arcade.Sprite {
    constructor(scene, x, y, type = 'crystal') {
        super(scene, x, y, type);
        
        this.scene.add.existing(this);
        this.scene.physics.add.existing(this);
        
        // Тип предмета
        this.type = type;
        
        // Характеристики по типам
        this.setupByType();
        
        // Анимация парения
        this.floatOffset = Math.random() * Math.PI * 2;
        this.floatSpeed = 0.05;
        this.floatAmount = 5;
        
        // Свечение
        this.createGlow();
    }
    
    setupByType() {
        switch (this.type) {
            case 'crystal':
                this.value = 10;
                this.setTint(0x00ffff);
                this.scale = 0.8;
                break;
            case 'blue_crystal':
                this.value = 20;
                this.setTint(0x0088ff);
                this.scale = 1;
                break;
            case 'purple_crystal':
                this.value = 50;
                this.setTint(0x8a2be2);
                this.scale = 1.2;
                break;
            case 'health':
                this.value = 25;
                this.setTint(0x00ff00);
                this.scale = 0.9;
                break;
            case 'mana':
                this.value = 30;
                this.setTint(0x0088ff);
                this.scale = 0.9;
                break;
            default:
                this.value = 10;
                this.setTint(0x00ffff);
        }
        
        this.setImmovable(true);
        this.body.allowGravity = false;
    }
    
    createGlow() {
        // Создаём спрайт свечения позади
        this.glow = this.scene.add.circle(this.x, this.y, 15, this.getTint(), 0.3);
        this.glow.setDepth(this.depth - 1);
    }
    
    update(time) {
        // Парение вверх-вниз
        const floatY = Math.sin(time * this.floatSpeed + this.floatOffset) * this.floatAmount;
        this.y = this.originalY || this.y;
        this.y += floatY;
        
        // Обновление позиции свечения
        if (this.glow) {
            this.glow.x = this.x;
            this.glow.y = this.y;
        }
        
        // Вращение (если есть)
        this.angle += 0.5;
    }
    
    collect(player) {
        // Эффект сбора
        this.createCollectEffect();
        
        // Применяем эффект
        this.applyEffect(player);
        
        // Скрываем предмет
        this.disableBody(true, true);
        if (this.glow) {
            this.glow.destroy();
        }
        
        // Звук
        if (this.scene.soundCollect) {
            this.scene.soundCollect.play();
        }
        
        return this.value;
    }
    
    applyEffect(player) {
        switch (this.type) {
            case 'crystal':
            case 'blue_crystal':
            case 'purple_crystal':
                // Добавляем очки/кристаллы
                if (this.scene.crystals !== undefined) {
                    this.scene.crystals += this.value;
                    if (this.scene.hud) {
                        this.scene.hud.updateCrystals(this.scene.crystals);
                    }
                }
                break;
                
            case 'health':
                // Лечение
                if (player && player.health) {
                    player.heal(this.value);
                }
                break;
                
            case 'mana':
                // Восстановление маны
                if (player && player.mana) {
                    player.mana = Math.min(player.mana + this.value, player.maxMana);
                }
                break;
        }
    }
    
    createCollectEffect() {
        // Частицы при сборе
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 50, max: 150 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.5, end: 0 },
            lifespan: 400,
            quantity: 10,
            tint: this.getTint()
        });
        
        // Анимация "всасывания" к игроку
        const tween = this.scene.tweens.createCounter({
            from: 0,
            to: 1,
            duration: 300,
            onUpdate: (value) => {
                // Можно добавить анимацию полёта к игроку
            },
            onComplete: () => {
                particles.destroy();
            }
        });
    }
    
    enable(x, y) {
        this.enableBody(true, x, y, true, true);
        this.originalY = y;
        if (this.glow) {
            this.glow.setPosition(x, y);
            this.glow.setVisible(true);
        }
    }
    
    disable() {
        this.disableBody(true, true);
        if (this.glow) {
            this.glow.setVisible(false);
        }
    }
}

// Экспорт
window.Collectible = Collectible;
