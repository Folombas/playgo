import * as THREE from 'three';

export class EffectManager {
    constructor(scene) {
        this.scene = scene;
        this.particles = [];
        this.comets = [];
        this.nebulae = [];
        this.createParticleMaterial();
        this.createNebula();
    }

    createNebula() {
        // Create cosmic nebula clouds
        const nebulaGeometry = new THREE.PlaneGeometry(100, 100);
        const nebulaMaterial = new THREE.MeshBasicMaterial({
            color: 0x667eea,
            transparent: true,
            opacity: 0.05,
            side: THREE.DoubleSide,
            depthWrite: false,
            blending: THREE.AdditiveBlending
        });
        
        for (let i = 0; i < 3; i++) {
            const nebula = new THREE.Mesh(nebulaGeometry, nebulaMaterial.clone());
            nebula.position.set(
                (Math.random() - 0.5) * 50,
                (Math.random() - 0.5) * 50,
                -50 - Math.random() * 30
            );
            nebula.rotation.z = Math.random() * Math.PI;
            this.scene.add(nebula);
            this.nebulae.push({
                mesh: nebula,
                speed: 0.001 + Math.random() * 0.002,
                opacity: 0.03 + Math.random() * 0.04
            });
        }
    }

    spawnComet() {
        const geometry = new THREE.SphereGeometry(0.3, 8, 8);
        const material = new THREE.MeshBasicMaterial({
            color: 0xffffff,
            transparent: true,
            opacity: 0.8
        });
        
        const comet = new THREE.Mesh(geometry, material);
        
        // Random start position
        const angle = Math.random() * Math.PI * 2;
        const distance = 60;
        comet.position.set(
            Math.cos(angle) * distance,
            (Math.random() - 0.5) * 30,
            Math.sin(angle) * distance - 30
        );
        
        // Trail
        const trailGeometry = new THREE.BufferGeometry();
        const trailPositions = [];
        for (let i = 0; i < 20; i++) {
            trailPositions.push(0, 0, i * 0.5);
        }
        trailGeometry.setAttribute('position', new THREE.Float32BufferAttribute(trailPositions, 3));
        const trailMaterial = new THREE.LineBasicMaterial({
            color: 0x00ffff,
            transparent: true,
            opacity: 0.5
        });
        const trail = new THREE.Line(trailGeometry, trailMaterial);
        trail.rotation.x = Math.PI / 2;
        comet.add(trail);
        
        this.scene.add(comet);
        
        this.comets.push({
            mesh: comet,
            velocity: new THREE.Vector3(
                (Math.random() - 0.5) * 10,
                (Math.random() - 0.5) * 5,
                30 + Math.random() * 20
            ),
            life: 5
        });
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

        // Update comets (spawn randomly)
        if (Math.random() < 0.01) {
            this.spawnComet();
        }
        
        this.comets = this.comets.filter(comet => {
            comet.mesh.position.add(comet.velocity.clone().multiplyScalar(deltaTime));
            comet.life -= deltaTime;
            
            // Update trail
            if (comet.mesh.children[0]) {
                comet.mesh.children[0].material.opacity = comet.life / 5 * 0.5;
            }
            
            return comet.life > 0;
        });
        
        // Remove dead comets
        for (const comet of this.comets) {
            if (comet.life <= 0) {
                this.scene.remove(comet.mesh);
            }
        }
        
        // Animate nebulae
        for (const nebula of this.nebulae) {
            nebula.mesh.rotation.z += nebula.speed;
            nebula.mesh.material.opacity = nebula.opacity + Math.sin(Date.now() * 0.001) * 0.02;
        }

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
        
        // Clear comets
        for (const comet of this.comets) {
            this.scene.remove(comet.mesh);
        }
        this.comets = [];
    }
}
