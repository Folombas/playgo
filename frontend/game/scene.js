import * as THREE from 'three';

export class Game {
    constructor(canvas) {
        this.canvas = canvas;
        this.scene = null;
        this.camera = null;
        this.renderer = null;
        this.player = null;
        this.enemies = [];
        this.particles = [];
        this.clock = new THREE.Clock();
        this.score = 0;
        this.level = 1;
        this.health = 100;
        this.shield = false;
        this.combo = 0;
        this.scoreMultiplier = 1;
        this.isRunning = false;
        this.isGameOver = false;
        this.lastSpawn = 0;
        this.lastBonus = 0;
        this.spawnInterval = 1500;
        this.bonusInterval = 8000;
    }

    init() {
        // Scene with dark background
        this.scene = new THREE.Scene();
        this.scene.background = new THREE.Color(0x000011);
        this.scene.fog = new THREE.Fog(0x000011, 20, 100);

        // Camera
        this.camera = new THREE.PerspectiveCamera(
            75,
            this.canvas.width / this.canvas.height,
            0.1,
            100
        );
        this.camera.position.set(0, 3, 8);
        this.camera.lookAt(0, 0, 0);

        // Renderer - оптимизированный
        this.renderer = new THREE.WebGLRenderer({
            canvas: this.canvas,
            antialias: false, // Отключаем для производительности
            powerPreference: 'high-performance'
        });
        this.renderer.setSize(this.canvas.width, this.canvas.height);
        this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

        // Lighting - упрощённое
        const ambientLight = new THREE.AmbientLight(0x404040, 0.6);
        this.scene.add(ambientLight);

        const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
        dirLight.position.set(5, 10, 5);
        this.scene.add(dirLight);

        // Звёзды - оптимизированные
        this.createStarfield(2000);

        // Игрок
        this.player = this.createPlayer();
        this.scene.add(this.player);

        // Input
        this.keys = {};
        window.addEventListener('keydown', (e) => {
            this.keys[e.code] = true;
            if (e.code === 'KeyM') {
                // Toggle music placeholder
            }
        });
        window.addEventListener('keyup', (e) => this.keys[e.code] = false);
        window.addEventListener('resize', () => this.onResize());
    }

    createStarfield(count) {
        const geometry = new THREE.BufferGeometry();
        const positions = new Float32Array(count * 3);
        
        for (let i = 0; i < count * 3; i++) {
            positions[i] = (Math.random() - 0.5) * 200;
        }
        
        geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
        
        const material = new THREE.PointsMaterial({
            color: 0xffffff,
            size: 0.15,
            transparent: true,
            opacity: 0.8
        });
        
        this.starfield = new THREE.Points(geometry, material);
        this.scene.add(this.starfield);
    }

    createPlayer() {
        const group = new THREE.Group();

        // Корпус - простой конус
        const bodyGeo = new THREE.ConeGeometry(0.4, 1.2, 8);
        const bodyMat = new THREE.MeshStandardMaterial({
            color: 0x00ff87,
            metalness: 0.6,
            roughness: 0.4
        });
        const body = new THREE.Mesh(bodyGeo, bodyMat);
        body.rotation.x = Math.PI / 2;
        group.add(body);

        // Кокпит
        const cockpitGeo = new THREE.SphereGeometry(0.2, 8, 8);
        const cockpitMat = new THREE.MeshStandardMaterial({
            color: 0x667eea,
            metalness: 0.8,
            roughness: 0.2
        });
        const cockpit = new THREE.Mesh(cockpitGeo, cockpitMat);
        cockpit.position.z = 0.2;
        group.add(cockpit);

        // Крылья
        const wingGeo = new THREE.BoxGeometry(1.2, 0.1, 0.3);
        const wingMat = new THREE.MeshStandardMaterial({ color: 0x764ba2 });
        const wings = new THREE.Mesh(wingGeo, wingMat);
        wings.position.z = -0.3;
        group.add(wings);

        // Двигатель - свет
        this.engineLight = new THREE.PointLight(0x00ffff, 0.5, 3);
        this.engineLight.position.z = -0.6;
        group.add(this.engineLight);

        // Щит (скрыт)
        const shieldGeo = new THREE.SphereGeometry(0.9, 8, 8);
        const shieldMat = new THREE.MeshBasicMaterial({
            color: 0x0088ff,
            transparent: true,
            opacity: 0.3,
            wireframe: true
        });
        this.shieldMesh = new THREE.Mesh(shieldGeo, shieldMat);
        this.shieldMesh.visible = false;
        group.add(this.shieldMesh);

        group.position.set(0, 0, 0);
        return group;
    }

    spawnEnemy() {
        const type = Math.random();
        let mesh;

        if (type < 0.5) {
            // Астероид - простой
            const geo = new THREE.DodecahedronGeometry(0.5, 0);
            const mat = new THREE.MeshStandardMaterial({
                color: 0x8b4513,
                flatShading: true
            });
            mesh = new THREE.Mesh(geo, mat);
        } else if (type < 0.8) {
            // Враг
            const geo = new THREE.ConeGeometry(0.3, 1, 6);
            const mat = new THREE.MeshStandardMaterial({
                color: 0xff0000,
                metalness: 0.5
            });
            mesh = new THREE.Mesh(geo, mat);
            mesh.rotation.x = Math.PI;
        } else {
            // Кристалл
            const geo = new THREE.OctahedronGeometry(0.4, 0);
            const mat = new THREE.MeshStandardMaterial({
                color: 0x00ffff,
                emissive: 0x00ffff,
                emissiveIntensity: 0.3
            });
            mesh = new THREE.Mesh(geo, mat);
        }

        // Позиция спавна
        const angle = Math.random() * Math.PI * 2;
        const dist = 15 + Math.random() * 10;
        mesh.position.set(
            Math.cos(angle) * dist,
            (Math.random() - 0.5) * 3,
            this.player.position.z - 25
        );

        mesh.userData = {
            type: type < 0.8 ? 'enemy' : 'bonus',
            rotSpeed: {
                x: (Math.random() - 0.5) * 2,
                y: (Math.random() - 0.5) * 2
            },
            speed: 8 + this.level * 0.5
        };

        this.scene.add(mesh);
        this.enemies.push(mesh);
    }

    createExplosion(position, color) {
        const count = 10;
        const geo = new THREE.BufferGeometry();
        const positions = new Float32Array(count * 3);
        const velocities = [];

        for (let i = 0; i < count; i++) {
            positions[i * 3] = position.x;
            positions[i * 3 + 1] = position.y;
            positions[i * 3 + 2] = position.z;
            velocities.push({
                x: (Math.random() - 0.5) * 8,
                y: (Math.random() - 0.5) * 8,
                z: (Math.random() - 0.5) * 8,
                life: 0.5 + Math.random() * 0.5
            });
        }

        geo.setAttribute('position', new THREE.BufferAttribute(positions, 3));

        const mat = new THREE.PointsMaterial({
            color: color,
            size: 0.3,
            transparent: true
        });

        const points = new THREE.Points(geo, mat);
        points.userData = { velocities, maxLife: 1 };
        this.scene.add(points);
        this.particles.push(points);
    }

    start() {
        this.isRunning = true;
        this.isGameOver = false;
        this.score = 0;
        this.level = 1;
        this.health = 100;
        this.shield = false;
        this.combo = 0;
        this.scoreMultiplier = 1;
        this.player.position.set(0, 0, 0);
        this.player.visible = true;
        this.shieldMesh.visible = false;

        // Очистка врагов
        this.enemies.forEach(e => this.scene.remove(e));
        this.enemies = [];

        // Очистка частиц
        this.particles.forEach(p => this.scene.remove(p));
        this.particles = [];

        this.clock.start();
    }

    restart() {
        this.start();
    }

    update() {
        const delta = Math.min(this.clock.getDelta(), 0.1); // Ограничение delta
        const now = Date.now();

        // Игрок
        this.updatePlayer(delta);

        // Спавн врагов
        if (now - this.lastSpawn > this.spawnInterval / (1 + this.level * 0.1)) {
            this.spawnEnemy();
            this.lastSpawn = now;
        }

        // Спавн бонусов
        if (now - this.lastBonus > this.bonusInterval) {
            this.spawnBonus();
            this.lastBonus = now;
        }

        // Враги
        this.updateEnemies(delta);

        // Частицы
        this.updateParticles(delta);

        // Звёзды
        this.starfield.rotation.y += 0.0002;

        // Level up
        this.level = Math.floor(this.score / 100) + 1;
        this.scoreMultiplier = 1 + Math.floor(this.combo / 5) * 0.5;
    }

    updatePlayer(delta) {
        const speed = 12;
        const boost = this.keys['Space'] ? 2 : 1;

        let dx = 0, dz = 0;
        if (this.keys['KeyW'] || this.keys['ArrowUp']) dz = -1;
        if (this.keys['KeyS'] || this.keys['ArrowDown']) dz = 1;
        if (this.keys['KeyA'] || this.keys['ArrowLeft']) dx = -1;
        if (this.keys['KeyD'] || this.keys['ArrowRight']) dx = 1;

        if (dx !== 0 || dz !== 0) {
            const len = Math.sqrt(dx * dx + dz * dz);
            dx /= len;
            dz /= len;
        }

        this.player.position.x += dx * speed * boost * delta;
        this.player.position.z += dz * speed * boost * delta;

        // Границы
        const bound = 20;
        this.player.position.x = Math.max(-bound, Math.min(bound, this.player.position.x));
        this.player.position.z = Math.max(-bound, Math.min(bound, this.player.position.z));

        // Наклон
        this.player.rotation.z = -dx * 0.3;
        this.player.rotation.x = Math.PI / 2 + dz * 0.15;

        // Двигатель
        this.engineLight.intensity = 0.5 + Math.sin(now * 0.02) * 0.2;

        // Камера
        this.camera.position.x = this.player.position.x * 0.3;
        this.camera.position.z = this.player.position.z + 8;
        this.camera.lookAt(this.player.position.x * 0.1, 0, 0);
    }

    updateEnemies(delta) {
        const playerPos = this.player.position;

        for (let i = this.enemies.length - 1; i >= 0; i--) {
            const enemy = this.enemies[i];
            const data = enemy.userData;

            // Движение к игроку
            const dir = new THREE.Vector3()
                .subVectors(playerPos, enemy.position)
                .normalize();
            enemy.position.add(dir.multiplyScalar(data.speed * delta));

            // Вращение
            enemy.rotation.x += data.rotSpeed.x * delta;
            enemy.rotation.y += data.rotSpeed.y * delta;

            // Удаление если далеко
            if (enemy.position.z > playerPos.z + 5) {
                this.scene.remove(enemy);
                this.enemies.splice(i, 1);
                continue;
            }

            // Коллизия
            const dist = enemy.position.distanceTo(playerPos);
            if (dist < 1.2) {
                if (data.type === 'bonus') {
                    this.collectBonus(enemy, i);
                } else {
                    this.hitEnemy(enemy, i);
                }
            }
        }
    }

    collectBonus(enemy, index) {
        this.createExplosion(enemy.position, 0x00ff00);
        this.score += 50;
        this.scene.remove(enemy);
        this.enemies.splice(index, 1);
    }

    hitEnemy(enemy, index) {
        if (this.shield) {
            this.shield = false;
            this.shieldMesh.visible = false;
            this.createExplosion(enemy.position, 0x0088ff);
        } else {
            this.health -= 20;
            this.combo = 0;
            this.createExplosion(enemy.position, 0xff0000);
            if (this.health <= 0) {
                this.gameOver();
            }
        }
        this.scene.remove(enemy);
        this.enemies.splice(index, 1);
    }

    spawnBonus() {
        const geo = new THREE.SphereGeometry(0.3, 8, 8);
        const mat = new THREE.MeshStandardMaterial({
            color: 0x00ff00,
            emissive: 0x00ff00,
            emissiveIntensity: 0.5
        });
        const mesh = new THREE.Mesh(geo, mat);

        const angle = Math.random() * Math.PI * 2;
        const dist = 15 + Math.random() * 10;
        mesh.position.set(
            Math.cos(angle) * dist,
            (Math.random() - 0.5) * 3,
            this.player.position.z - 30
        );

        mesh.userData = { type: 'bonus', rotSpeed: { x: 1, y: 1 }, speed: 6 };
        this.scene.add(mesh);
        this.enemies.push(mesh);
    }

    updateParticles(delta) {
        for (let i = this.particles.length - 1; i >= 0; i--) {
            const p = this.particles[i];
            const positions = p.geometry.attributes.position.array;
            const vels = p.userData.velocities;

            for (let j = 0; j < vels.length; j++) {
                positions[j * 3] += vels[j].x * delta;
                positions[j * 3 + 1] += vels[j].y * delta;
                positions[j * 3 + 2] += vels[j].z * delta;
                vels[j].life -= delta;
            }

            p.geometry.attributes.position.needsUpdate = true;
            p.material.opacity *= 0.95;

            if (p.material.opacity < 0.05) {
                this.scene.remove(p);
                this.particles.splice(i, 1);
            }
        }
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
        this.player.visible = false;
        this.createExplosion(this.player.position, 0x00ff87, 30);
    }

    getScore() { return this.score; }
    getLevel() { return this.level; }
    getHealth() { return this.health; }
    getCombo() { return this.combo; }
    getMultiplier() { return this.scoreMultiplier; }
    isGameOver() { return this.isGameOver; }
}
