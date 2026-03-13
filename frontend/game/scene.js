import * as THREE from 'three';
import { Player } from './player.js';
import { EnemyManager } from './enemies.js';
import { EffectManager } from './effects.js';
import { AudioManager } from './audio.js';

export class Game {
    constructor(canvas) {
        this.canvas = canvas;
        this.scene = null;
        this.camera = null;
        this.renderer = null;
        this.player = null;
        this.enemyManager = null;
        this.effectManager = null;
        this.audioManager = null;
        this.clock = new THREE.Clock();
        this.score = 0;
        this.level = 1;
        this.health = 100;
        this.shield = false;
        this.combo = 0;
        this.maxCombo = 0;
        this.scoreMultiplier = 1;
        this.isRunning = false;
        this.isGameOver = false;
    }

    init() {
        // Scene
        this.scene = new THREE.Scene();
        this.scene.fog = new THREE.FogExp2(0x000011, 0.02);

        // Camera
        this.camera = new THREE.PerspectiveCamera(
            75,
            this.canvas.width / this.canvas.height,
            0.1,
            1000
        );
        this.camera.position.set(0, 5, 10);
        this.camera.lookAt(0, 0, 0);

        // Renderer
        this.renderer = new THREE.WebGLRenderer({ 
            canvas: this.canvas,
            antialias: true 
        });
        this.renderer.setSize(this.canvas.width, this.canvas.height);
        this.renderer.setPixelRatio(window.devicePixelRatio);
        this.renderer.shadowMap.enabled = true;
        this.renderer.shadowMap.type = THREE.PCFSoftShadowMap;

        // Lighting
        this.setupLighting();

        // Starfield background
        this.createStarfield();

        // Game objects
        this.player = new Player(this.scene);
        this.enemyManager = new EnemyManager(this.scene);
        this.effectManager = new EffectManager(this.scene);
        this.audioManager = new AudioManager();

        // Input
        this.setupInput();
    }

    setupLighting() {
        // Ambient light
        const ambientLight = new THREE.AmbientLight(0x404040, 0.5);
        this.scene.add(ambientLight);

        // Directional light (sun)
        const directionalLight = new THREE.DirectionalLight(0xffffff, 1);
        directionalLight.position.set(10, 20, 10);
        directionalLight.castShadow = true;
        directionalLight.shadow.mapSize.width = 2048;
        directionalLight.shadow.mapSize.height = 2048;
        this.scene.add(directionalLight);

        // Point lights for effects
        const pointLight1 = new THREE.PointLight(0x667eea, 1, 50);
        pointLight1.position.set(-5, 3, -5);
        this.scene.add(pointLight1);

        const pointLight2 = new THREE.PointLight(0x764ba2, 1, 50);
        pointLight2.position.set(5, 3, -5);
        this.scene.add(pointLight2);
    }

    createStarfield() {
        const starsGeometry = new THREE.BufferGeometry();
        const starsMaterial = new THREE.PointsMaterial({
            color: 0xffffff,
            size: 0.1,
            transparent: true,
            opacity: 0.8
        });

        const starsVertices = [];
        for (let i = 0; i < 10000; i++) {
            const x = (Math.random() - 0.5) * 2000;
            const y = (Math.random() - 0.5) * 2000;
            const z = (Math.random() - 0.5) * 2000;
            starsVertices.push(x, y, z);
        }

        starsGeometry.setAttribute(
            'position',
            new THREE.Float32BufferAttribute(starsVertices, 3)
        );

        this.starfield = new THREE.Points(starsGeometry, starsMaterial);
        this.scene.add(this.starfield);
    }

    setupInput() {
        this.keys = {};
        window.addEventListener('keydown', (e) => {
            this.keys[e.code] = true;
            
            // Toggle music
            if (e.code === 'KeyM') {
                this.audioManager.toggleMusic();
            }
        });
        window.addEventListener('keyup', (e) => {
            this.keys[e.code] = false;
        });
    }

    start() {
        this.isRunning = true;
        this.isGameOver = false;
        this.score = 0;
        this.level = 1;
        this.health = 100;
        this.shield = false;
        this.combo = 0;
        this.maxCombo = 0;
        this.scoreMultiplier = 1;
        this.clock.start();
        this.player.reset();
        this.enemyManager.reset();
        this.audioManager.playMusic();
    }

    restart() {
        this.start();
    }

    update() {
        const deltaTime = this.clock.getDelta();

        // Update player
        this.player.update(deltaTime, this.keys);

        // Update camera follow
        this.camera.position.x = this.player.mesh.position.x * 0.3;
        this.camera.position.z = this.player.mesh.position.z + 10;
        this.camera.lookAt(this.player.mesh.position.x * 0.1, 0, 0);

        // Update enemies
        const playerPos = this.player.mesh.position;
        const newEnemies = this.enemyManager.update(deltaTime, playerPos, this.level);
        
        // Check collisions
        for (const enemy of newEnemies) {
            if (this.checkCollision(this.player.mesh, enemy.mesh)) {
                if (enemy.isBonus) {
                    // Collect bonus
                    this.collectBonus(enemy);
                } else {
                    // Hit enemy
                    if (this.shield) {
                        this.shield = false;
                        this.player.setShield(false);
                        this.effectManager.createExplosion(enemy.mesh.position, 0x0088ff);
                        enemy.destroy();
                    } else {
                        this.health -= 20;
                        this.combo = 0;
                        this.scoreMultiplier = 1;
                        this.effectManager.createExplosion(enemy.mesh.position, 0xff0000);
                        enemy.destroy();
                        
                        if (this.health <= 0) {
                            this.health = 0;
                            this.gameOver();
                        }
                    }
                }
            }
        }

        // Remove dead enemies and add score
        const destroyedCount = this.enemyManager.removeDead();
        if (destroyedCount > 0) {
            this.combo += destroyedCount;
            if (this.combo > this.maxCombo) this.maxCombo = this.combo;
            // Combo multiplier
            this.scoreMultiplier = 1 + Math.floor(this.combo / 5) * 0.5;
        }
        this.score += Math.floor(destroyedCount * 10 * this.scoreMultiplier);

        // Level up every 100 points
        this.level = Math.floor(this.score / 100) + 1;

        // Update effects
        this.effectManager.update(deltaTime);

        // Update starfield
        this.starfield.rotation.y += 0.0001;
    }

    checkCollision(obj1, obj2) {
        const distance = obj1.position.distanceTo(obj2.position);
        return distance < 1.5;
    }

    collectBonus(bonus) {
        const pos = bonus.mesh.position;
        
        switch(bonus.type) {
            case 'health':
                this.health = Math.min(100, this.health + 25);
                this.effectManager.createExplosion(pos, 0x00ff00, 15);
                this.audioManager.playCollectSound();
                break;
            case 'shield':
                this.shield = true;
                this.player.setShield(true);
                this.effectManager.createExplosion(pos, 0x0088ff, 20);
                this.audioManager.playCollectSound();
                break;
            case 'multiplier':
                this.score += 50;
                this.effectManager.createExplosion(pos, 0xffd700, 25);
                this.audioManager.playCollectSound();
                break;
        }
        
        bonus.destroy();
    }

    render() {
        this.renderer.render(this.scene, this.camera);
    }

    onResize() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
        this.camera.aspect = this.canvas.width / this.canvas.height;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(this.canvas.width, this.canvas.height);
    }

    gameOver() {
        this.isRunning = false;
        this.isGameOver = true;
        this.audioManager.stopMusic();
        this.effectManager.createExplosion(this.player.mesh.position, 0x00ff87, 50);
    }

    getScore() {
        return this.score;
    }

    getLevel() {
        return this.level;
    }

    getHealth() {
        return this.health;
    }

    getCombo() {
        return this.combo;
    }

    getMultiplier() {
        return this.scoreMultiplier;
    }

    isGameOver() {
        return this.isGameOver;
    }
}
