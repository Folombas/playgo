import * as THREE from 'three';

export class Player {
    constructor(scene) {
        this.scene = scene;
        this.mesh = null;
        this.shieldMesh = null;
        this.velocity = new THREE.Vector3();
        this.speed = 15;
        this.boostSpeed = 30;
        this.hasShield = false;
        this.createSpaceship();
    }

    createSpaceship() {
        // Main body
        const bodyGeometry = new THREE.ConeGeometry(0.5, 2, 8);
        const bodyMaterial = new THREE.MeshStandardMaterial({
            color: 0x00ff87,
            metalness: 0.8,
            roughness: 0.2,
            emissive: 0x00ff87,
            emissiveIntensity: 0.3
        });
        this.mesh = new THREE.Mesh(bodyGeometry, bodyMaterial);
        this.mesh.rotation.x = Math.PI / 2;
        this.mesh.castShadow = true;
        this.mesh.receiveShadow = true;

        // Cockpit
        const cockpitGeometry = new THREE.SphereGeometry(0.3, 16, 16);
        const cockpitMaterial = new THREE.MeshStandardMaterial({
            color: 0x667eea,
            metalness: 0.9,
            roughness: 0.1,
            transparent: true,
            opacity: 0.7
        });
        const cockpit = new THREE.Mesh(cockpitGeometry, cockpitMaterial);
        cockpit.position.z = 0.3;
        this.mesh.add(cockpit);

        // Wings
        const wingGeometry = new THREE.BoxGeometry(2, 0.1, 0.5);
        const wingMaterial = new THREE.MeshStandardMaterial({
            color: 0x764ba2,
            metalness: 0.7,
            roughness: 0.3
        });
        const wings = new THREE.Mesh(wingGeometry, wingMaterial);
        wings.position.z = -0.5;
        this.mesh.add(wings);

        // Engine glow
        const engineGeometry = new THREE.SphereGeometry(0.2, 16, 16);
        const engineMaterial = new THREE.MeshBasicMaterial({
            color: 0x00ffff,
            transparent: true,
            opacity: 0.8
        });
        this.engine = new THREE.Mesh(engineGeometry, engineMaterial);
        this.engine.position.z = -1;
        this.mesh.add(this.engine);

        // Engine light
        this.engineLight = new THREE.PointLight(0x00ffff, 1, 5);
        this.engineLight.position.z = -1.5;
        this.mesh.add(this.engineLight);

        // Shield (hidden by default)
        const shieldGeometry = new THREE.SphereGeometry(1.2, 32, 32);
        const shieldMaterial = new THREE.MeshStandardMaterial({
            color: 0x0088ff,
            metalness: 0.9,
            roughness: 0.1,
            emissive: 0x0088ff,
            emissiveIntensity: 0.5,
            transparent: true,
            opacity: 0.3,
            wireframe: true,
            side: THREE.DoubleSide
        });
        this.shieldMesh = new THREE.Mesh(shieldGeometry, shieldMaterial);
        this.shieldMesh.visible = false;
        this.mesh.add(this.shieldMesh);

        this.scene.add(this.mesh);
        this.reset();
    }

    reset() {
        this.mesh.position.set(0, 0, 0);
        this.velocity.set(0, 0, 0);
        this.hasShield = false;
        this.shieldMesh.visible = false;
    }

    setShield(active) {
        this.hasShield = active;
        this.shieldMesh.visible = active;
    }

    update(deltaTime, keys) {
        // Movement
        let moveSpeed = this.speed;
        
        if (keys['Space']) {
            moveSpeed = this.boostSpeed;
            // Boost effect
            this.engine.scale.set(1.5, 1.5, 1.5);
            this.engineLight.intensity = 2;
        } else {
            this.engine.scale.set(1, 1, 1);
            this.engineLight.intensity = 1;
        }

        // Direction input
        const direction = new THREE.Vector3(0, 0, 0);
        
        if (keys['KeyW'] || keys['ArrowUp']) direction.z -= 1;
        if (keys['KeyS'] || keys['ArrowDown']) direction.z += 1;
        if (keys['KeyA'] || keys['ArrowLeft']) direction.x -= 1;
        if (keys['KeyD'] || keys['ArrowRight']) direction.x += 1;

        // Normalize diagonal movement
        if (direction.length() > 0) {
            direction.normalize();
        }

        // Apply movement
        this.mesh.position.x += direction.x * moveSpeed * deltaTime;
        this.mesh.position.z += direction.z * moveSpeed * deltaTime;

        // Tilt effect when moving
        this.mesh.rotation.z = -direction.x * 0.3;
        this.mesh.rotation.x = Math.PI / 2 + direction.z * 0.2;

        // Boundaries
        const boundary = 30;
        this.mesh.position.x = Math.max(-boundary, Math.min(boundary, this.mesh.position.x));
        this.mesh.position.z = Math.max(-boundary, Math.min(boundary, this.mesh.position.z));

        // Engine flicker effect
        this.engine.material.opacity = 0.6 + Math.sin(Date.now() * 0.02) * 0.2;
        
        // Shield animation
        if (this.hasShield) {
            this.shieldMesh.rotation.y += 0.02;
            this.shieldMesh.scale.setScalar(1 + Math.sin(Date.now() * 0.005) * 0.05);
        }
    }
}
