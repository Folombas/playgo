import * as THREE from 'three';

export class EffectManager {
    constructor(scene) {
        this.scene = scene;
        this.particles = [];
        this.createParticleMaterial();
    }

    createParticleMaterial() {
        this.particleMaterial = new THREE.PointsMaterial({
            color: 0xffffff,
            size: 0.3,
            transparent: true,
            opacity: 0.8,
            blending: THREE.AdditiveBlending,
            depthWrite: false
        });
    }

    createExplosion(position, color, count = 20) {
        const particles = [];
        const geometry = new THREE.BufferGeometry();
        
        for (let i = 0; i < count; i++) {
            const particle = {
                position: position.clone(),
                velocity: new THREE.Vector3(
                    (Math.random() - 0.5) * 10,
                    (Math.random() - 0.5) * 10,
                    (Math.random() - 0.5) * 10
                ),
                life: 1.0,
                decay: 0.5 + Math.random() * 1.0,
                color: new THREE.Color(color)
            };
            particles.push(particle);
        }

        this.particles.push(...particles);
    }

    createTrail(position, color) {
        const particle = {
            position: position.clone(),
            velocity: new THREE.Vector3(0, 0, 2),
            life: 0.5,
            decay: 2.0,
            color: new THREE.Color(color)
        };
        this.particles.push(particle);
    }

    update(deltaTime) {
        // Update particles
        this.particles = this.particles.filter(particle => {
            particle.position.add(particle.velocity.clone().multiplyScalar(deltaTime));
            particle.life -= particle.decay * deltaTime;
            return particle.life > 0;
        });

        // Create particle system for rendering
        if (this.particles.length > 0) {
            const positions = [];
            const colors = [];
            const sizes = [];

            for (const particle of this.particles) {
                positions.push(particle.position.x, particle.position.y, particle.position.z);
                colors.push(particle.color.r, particle.color.g, particle.color.b);
                sizes.push(particle.life * 2);
            }

            const geometry = new THREE.BufferGeometry();
            geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
            geometry.setAttribute('color', new THREE.Float32BufferAttribute(colors, 3));
            geometry.setAttribute('size', new THREE.Float32BufferAttribute(sizes, 1));

            const material = new THREE.PointsMaterial({
                size: 0.5,
                vertexColors: true,
                transparent: true,
                opacity: 0.8,
                blending: THREE.AdditiveBlending,
                depthWrite: false
            });

            // Remove old particle system
            if (this.particleSystem) {
                this.scene.remove(this.particleSystem);
                this.particleSystem.geometry.dispose();
            }

            this.particleSystem = new THREE.Points(geometry, material);
            this.scene.add(this.particleSystem);
        } else if (this.particleSystem) {
            this.scene.remove(this.particleSystem);
            this.particleSystem = null;
        }
    }

    createPowerUpEffect(position) {
        // Ring effect
        const ringGeometry = new THREE.TorusGeometry(1, 0.1, 8, 32);
        const ringMaterial = new THREE.MeshBasicMaterial({
            color: 0x00ff00,
            transparent: true,
            opacity: 0.8,
            side: THREE.DoubleSide
        });
        const ring = new THREE.Mesh(ringGeometry, ringMaterial);
        ring.position.copy(position);
        ring.rotation.x = Math.PI / 2;
        
        this.scene.add(ring);

        // Animate ring
        const animateRing = () => {
            ring.scale.multiplyScalar(1.1);
            ring.material.opacity -= 0.05;
            
            if (ring.material.opacity > 0) {
                requestAnimationFrame(animateRing);
            } else {
                this.scene.remove(ring);
                ring.geometry.dispose();
                ring.material.dispose();
            }
        };
        
        animateRing();
    }

    clear() {
        if (this.particleSystem) {
            this.scene.remove(this.particleSystem);
            this.particleSystem = null;
        }
        this.particles = [];
    }
}
