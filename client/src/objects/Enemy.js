// Enemy.js - Класс врага (Баги, Вирусы, Ошибки)

class Enemy extends Phaser.Physics.Arcade.Sprite {
    constructor(scene, x, y, type = 'bug') {
        super(scene, x, y, type);
        
        this.scene.add.existing(this);
        this.scene.physics.add.existing(this);
        
        // Тип врага
        this.type = type;
        
        // Характеристики по типам
        this.setupByType();
        
        // Состояние
        this.isAlive = true;
        this.isHit = false;
        this.moveDirection = 1;
        this.moveSpeed = 50;
        
        // Таймер движения
        this.lastMoveChange = 0;
        this.moveChangeInterval = 2000;
    }
    
    setupByType() {
        switch (this.type) {
            case 'bug':
                this.health = 30;
                this.damage = 10;
                this.speed = 40;
                this.score = 10;
                this.setTint(0xff4444);
                break;
            case 'virus':
                this.health = 50;
                this.damage = 20;
                this.speed = 60;
                this.score = 20;
                this.setTint(0x44ff44);
                break;
            case 'error':
                this.health = 80;
                this.damage = 30;
                this.speed = 30;
                this.score = 30;
                this.setTint(0xff44ff);
                break;
            case 'firewall':
                this.health = 100;
                this.damage = 25;
                this.speed = 20;
                this.score = 40;
                this.setTint(0xff8800);
                break;
            default:
                this.health = 30;
                this.damage = 10;
                this.speed = 40;
                this.score = 10;
        }
        
        this.maxHealth = this.health;
        this.setCollideWorldBounds(true);
        this.setBounce(0.2);
    }
    
    update(time) {
        if (!this.isAlive) return;
        
        // Простое патрулирование
        if (time > this.lastMoveChange + this.moveChangeInterval) {
            this.changeDirection();
            this.lastMoveChange = time;
        }
        
        // Движение
        this.setVelocityX(this.moveSpeed * this.moveDirection);
        
        // Проверка краёв платформы
        this.checkPlatformEdges();
        
        // Анимация
        if (this.body.velocity.x !== 0) {
            this.setFlipX(this.moveDirection < 0);
        }
    }
    
    changeDirection() {
        this.moveDirection *= -1;
    }
    
    checkPlatformEdges() {
        // Разворот на краях платформы (упрощённо)
        if (this.body.blocked.left) {
            this.moveDirection = 1;
        } else if (this.body.blocked.right) {
            this.moveDirection = -1;
        }
    }
    
    takeDamage(amount, fromSpell = false) {
        if (!this.isAlive || this.isHit) return;
        
        this.health -= amount;
        this.isHit = true;
        
        // Отталкивание
        const knockback = fromSpell ? 200 : 100;
        this.setVelocityY(-100);
        this.setVelocityX(-this.moveDirection * knockback);
        
        // Эффект попадания
        this.createHitParticles();
        
        // Сброс состояния попадания
        setTimeout(() => {
            this.isHit = false;
        }, 300);
        
        // Проверка смерти
        if (this.health <= 0) {
            this.die();
        }
    }
    
    createHitParticles() {
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 50, max: 150 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.4, end: 0 },
            lifespan: 300,
            quantity: 6,
            tint: this.getTint()
        });
        
        setTimeout(() => particles.destroy(), 300);
    }
    
    die() {
        this.isAlive = false;
        this.disableBody(true, false);
        
        // Эффект смерти
        const particles = this.scene.add.particles(this.x, this.y, 'particle', {
            speed: { min: 50, max: 200 },
            angle: { min: 0, max: 360 },
            scale: { start: 0.5, end: 0 },
            lifespan: 500,
            quantity: 15,
            tint: this.getTint()
        });
        
        setTimeout(() => particles.destroy(), 500);
        
        // Звук
        if (this.scene.soundEnemyDie) {
            this.scene.soundEnemyDie.play();
        }
        
        // Добавляем очки
        if (this.scene.score !== undefined) {
            this.scene.score += this.score;
            if (this.scene.hud) {
                this.scene.hud.updateScore(this.scene.score);
            }
        }
    }
    
    respawn(x, y) {
        this.isAlive = true;
        this.health = this.maxHealth;
        this.enableBody(true, x, y, true, true);
        this.clearTint();
        this.setupByType();
    }
}

// Экспорт
window.Enemy = Enemy;
