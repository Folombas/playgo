// Player.js - Класс игрока (Purple Lord)

class Player extends Phaser.Physics.Arcade.Sprite {
    constructor(scene, x, y) {
        super(scene, x, y, 'player');
        
        this.scene.add.existing(this);
        this.scene.physics.add.existing(this);
        
        // Настройки физики
        this.setCollideWorldBounds(true);
        this.setBounce(0.1);
        this.setDragX(800);
        
        // Характеристики
        this.health = 100;
        this.maxHealth = 100;
        this.mana = 100;
        this.maxMana = 100;
        this.speed = 200;
        this.jumpForce = -450;
        this.isGrounded = false;
        this.isInvulnerable = false;
        
        // Направления
        this.facingRight = true;
        
        // Анимации
        this.createAnimations();
        
        // Кулдаун заклинания
        this.spellCooldown = 0;
        this.spellCooldownTime = 1000; // 1 секунда
    }
    
    createAnimations() {
        const config = this.scene;
        
        // Анимация ходьбы (если есть спрайты)
        config.anims.create({
            key: 'walk',
            frames: config.anims.generateFrameNumbers('player', { start: 0, end: 3 }),
            frameRate: 10,
            repeat: -1
        }, true);
        
        // Анимация бега
        config.anims.create({
            key: 'run',
            frames: config.anims.generateFrameNumbers('player', { start: 4, end: 7 }),
            frameRate: 12,
            repeat: -1
        }, true);
        
        // Анимация прыжка
        config.anims.create({
            key: 'jump',
            frames: config.anims.generateFrameNumbers('player', { start: 8, end: 9 }),
            frameRate: 8,
            repeat: 0
        }, true);
        
        // Анимация каста заклинания
        config.anims.create({
            key: 'cast',
            frames: config.anims.generateFrameNumbers('player', { start: 10, end: 12 }),
            frameRate: 15,
            repeat: 0
        }, true);
    }
    
    update(cursors, time) {
        // Движение влево/вправо
        if (cursors.left.isDown) {
            this.setVelocityX(-this.speed);
            this.facingRight = false;
            this.setFlipX(true);
            
            if (this.isGrounded) {
                this.play('run', true);
            }
        } else if (cursors.right.isDown) {
            this.setVelocityX(this.speed);
            this.facingRight = true;
            this.setFlipX(false);
            
            if (this.isGrounded) {
                this.play('run', true);
            }
        } else {
            this.setVelocityX(0);
            
            if (this.isGrounded) {
                this.setFrame(0);
            }
        }
        
        // Прыжок
        if (cursors.up.isDown || cursors.space.isDown) {
            if (this.isGrounded) {
                this.setVelocityY(this.jumpForce);
                this.isGrounded = false;
                this.play('jump', true);
                
                // Эффект прыжка
                this.createJumpParticles();
            }
        }
        
        // Заклинание
        if (cursors.f.isDown && time > this.spellCooldown) {
            this.castSpell();
            this.spellCooldown = time + this.spellCooldownTime;
        }
        
        // Проверка земли
        this.isGrounded = this.body.blocked.down || 
                         (this.body.touching.down && this.body.velocity.y >= 0);
        
        // Восстановление маны
        if (this.mana < this.maxMana) {
            this.mana += 0.1;
        }
        
        // Обновление HUD
        if (this.scene.hud) {
            this.scene.hud.updateHealth(this.health);
            this.scene.hud.updateMana(this.mana);
        }
    }
    
    castSpell() {
        if (this.mana < 20) return;
        
        this.mana -= 20;
        this.play('cast', true);
        
        // Создаём фаербол
        const spell = this.scene.spells.get();
        if (spell) {
            spell.enable(this.x, this.y, this.facingRight);
        }
        
        // Звук заклинания
        if (this.scene.soundSpell) {
            this.scene.soundSpell.play();
        }
    }
    
    createJumpParticles() {
        const particles = this.scene.add.particles(this.x, this.y + 16, 'particle', {
            speed: { min: 50, max: 100 },
            angle: { min: 180, max: 270 },
            scale: { start: 0.5, end: 0 },
            lifespan: 300,
            quantity: 5,
            tint: 0x8a2be2
        });
        
        setTimeout(() => particles.destroy(), 300);
    }
    
    takeDamage(amount) {
        if (this.isInvulnerable) return;
        
        this.health -= amount;
        
        // Эффект получения урона
        this.scene.cameras.main.shake(100, 0.01);
        this.createHitParticles();
        
        // Временная неуязвимость
        this.isInvulnerable = true;
        this.setTint(0xff0000);
        
        setTimeout(() => {
            this.isInvulnerable = false;
            this.clearTint();
        }, 1000);
        
        // Проверка смерти
        if (this.health <= 0) {
            this.die();
        }
    }
    
    createHitParticles() {
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 50, max: 150 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.5, end: 0 },
            lifespan: 400,
            quantity: 10,
            tint: 0xff0000
        });
        
        setTimeout(() => particles.destroy(), 400);
    }
    
    heal(amount) {
        this.health = Math.min(this.health + amount, this.maxHealth);
        
        // Эффект лечения
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 30, max: 80 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.4, end: 0 },
            lifespan: 500,
            quantity: 8,
            tint: 0x00ff00
        });
        
        setTimeout(() => particles.destroy(), 500);
    }
    
    die() {
        // Эффект смерти
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 50, max: 200 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.6, end: 0 },
            lifespan: 600,
            quantity: 20,
            tint: 0x8a2be2
        });
        
        setTimeout(() => particles.destroy(), 600);
        
        // Респаун
        setTimeout(() => {
            this.health = this.maxHealth;
            this.mana = this.maxMana;
            this.setPosition(this.scene.spawnPoint.x, this.scene.spawnPoint.y);
        }, 1500);
    }
    
    collectCrystal(crystal) {
        // Эффект сбора
        const particles = this.scene.add.particles(crystal.x, crystal.y, 'particle', {
            speed: { min: 30, max: 100 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.4, end: 0 },
            lifespan: 400,
            quantity: 8,
            tint: 0x00ffff
        });
        
        setTimeout(() => particles.destroy(), 400);
        
        // Звук
        if (this.scene.soundCollect) {
            this.scene.soundCollect.play();
        }
    }
}

// Экспорт для использования в сценах
window.Player = Player;
