import * as THREE from 'three';

export class EnemyManager {
    constructor(scene) {
        this.scene = scene;
        this.enemies = [];
        this.spawnTimer = 0;
        this.spawnInterval = 2;
        this.createEnemyTypes();
    }

    createEnemyTypes() {
        // Asteroid type
        this.asteroidGeometry = new THREE.DodecahedronGeometry(0.8, 0);
        this.asteroidMaterial = new THREE.MeshStandardMaterial({
            color: 0x8b4513,
            metalness: 0.3,
            roughness: 0.9,
            flatShading: true
        });

        // Enemy ship type
        this.enemyShipGeometry = new THREE.ConeGeometry(0.4, 1.5, 6);
        this.enemyShipMaterial = new THREE.MeshStandardMaterial({
            color: 0xff0000,
            metalness: 0.7,
            roughness: 0.3,
            emissive: 0xff0000,
            emissiveIntensity: 0.4
        });

        // Crystal type
        this.crystalGeometry = new THREE.OctahedronGeometry(0.6, 0);
        this.crystalMaterial = new THREE.MeshStandardMaterial({
            color: 0x00ffff,
            metalness: 0.9,
            roughness: 0.1,
            emissive: 0x00ffff,
            emissiveIntensity: 0.5,
            transparent: true,
            opacity: 0.8
        });
    }

    spawnEnemy(playerZ) {
        const types = ['asteroid', 'asteroid', 'enemyShip', 'crystal'];
        const type = types[Math.floor(Math.random() * types.length)];
        
        let geometry, material, mesh;
        
        switch(type) {
            case 'asteroid':
                geometry = this.asteroidGeometry;
                material = this.asteroidMaterial;
                break;
            case 'enemyShip':
                geometry = this.enemyShipGeometry;
                material = this.enemyShipMaterial;
                break;
            case 'crystal':
                geometry = this.crystalGeometry;
                material = this.crystalMaterial;
                break;
        }

        mesh = new THREE.Mesh(geometry, material);
        
        // Random position ahead of player
        const angle = Math.random() * Math.PI * 2;
        const radius = 20 + Math.random() * 10;
        mesh.position.x = Math.cos(angle) * radius;
        mesh.position.z = playerZ - 30 - Math.random() * 20;
        mesh.position.y = (Math.random() - 0.5) * 5;
        
        // Random rotation
        mesh.rotation.x = Math.random() * Math.PI;
        mesh.rotation.y = Math.random() * Math.PI;
        
        mesh.castShadow = true;
        mesh.receiveShadow = true;

        // Store enemy data
        const enemy = {
            mesh: mesh,
            type: type,
            speed: 5 + Math.random() * 5 + (this.enemies.length * 0.5),
            rotationSpeed: new THREE.Vector3(
                (Math.random() - 0.5) * 2,
                (Math.random() - 0.5) * 2,
                (Math.random() - 0.5) * 2
            ),
            health: type === 'crystal' ? 1 : (type === 'enemyShip' ? 2 : 3)
        };

        this.scene.add(mesh);
        this.enemies.push(enemy);
    }

    update(deltaTime, playerPosition, level) {
        const newEnemies = [];
        this.spawnTimer += deltaTime;
        
        // Spawn interval decreases with level
        const currentSpawnInterval = Math.max(0.5, this.spawnInterval - (level - 1) * 0.2);
        
        if (this.spawnTimer >= currentSpawnInterval) {
            this.spawnEnemy(playerPosition.z);
            this.spawnTimer = 0;
        }

        // Update enemies
        for (const enemy of this.enemies) {
            // Move towards player
            const direction = new THREE.Vector3()
                .subVectors(playerPosition, enemy.mesh.position)
                .normalize();
            
            enemy.mesh.position.add(direction.multiplyScalar(enemy.speed * deltaTime));
            
            // Rotate
            enemy.mesh.rotation.x += enemy.rotationSpeed.x * deltaTime;
            enemy.mesh.rotation.y += enemy.rotationSpeed.y * deltaTime;
            enemy.mesh.rotation.z += enemy.rotationSpeed.z * deltaTime;

            // Remove if too far behind
            if (enemy.mesh.position.z > playerPosition.z + 10) {
                enemy.destroyed = true;
            }

            newEnemies.push(enemy);
        }

        return newEnemies;
    }

    removeDead() {
        let count = 0;
        this.enemies = this.enemies.filter(enemy => {
            if (enemy.destroyed) {
                this.scene.remove(enemy.mesh);
                enemy.mesh.geometry.dispose();
                count++;
                return false;
            }
            return true;
        });
        return count;
    }

    reset() {
        // Remove all enemies
        for (const enemy of this.enemies) {
            this.scene.remove(enemy.mesh);
        }
        this.enemies = [];
        this.spawnTimer = 0;
    }
}
