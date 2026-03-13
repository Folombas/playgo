export class AudioManager {
    constructor() {
        this.audioContext = null;
        this.musicOscillators = [];
        this.isMusicPlaying = false;
        this.musicGain = null;
        this.masterGain = null;
        this.musicEnabled = true;
    }

    init() {
        if (!this.audioContext) {
            this.audioContext = new (window.AudioContext || window.webkitAudioContext)();
            this.masterGain = this.audioContext.createGain();
            this.masterGain.connect(this.audioContext.destination);
            this.masterGain.gain.value = 0.3;
        }
    }

    playMusic() {
        if (!this.musicEnabled) return;
        
        this.init();
        this.stopMusic();
        
        this.isMusicPlaying = true;
        
        // Create ambient space music
        this.createAmbientDrone();
        this.createArpeggio();
    }

    createAmbientDrone() {
        const droneFreqs = [65.41, 98.00, 130.81]; // C2, G2, C3
        
        droneFreqs.forEach((freq, index) => {
            const oscillator = this.audioContext.createOscillator();
            const gain = this.audioContext.createGain();
            
            oscillator.type = 'sine';
            oscillator.frequency.value = freq;
            
            gain.gain.value = 0.1 / (index + 1);
            
            oscillator.connect(gain);
            gain.connect(this.masterGain);
            
            oscillator.start();
            this.musicOscillators.push({ oscillator, gain, type: 'drone' });
        });
    }

    createArpeggio() {
        const notes = [261.63, 329.63, 392.00, 523.25]; // C4, E4, G4, C5
        let noteIndex = 0;
        
        const playNote = () => {
            if (!this.isMusicPlaying) return;
            
            const oscillator = this.audioContext.createOscillator();
            const gain = this.audioContext.createGain();
            
            oscillator.type = 'triangle';
            oscillator.frequency.value = notes[noteIndex % notes.length];
            
            gain.gain.setValueAtTime(0.05, this.audioContext.currentTime);
            gain.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + 1.5);
            
            oscillator.connect(gain);
            gain.connect(this.masterGain);
            
            oscillator.start();
            oscillator.stop(this.audioContext.currentTime + 1.5);
            
            noteIndex++;
            setTimeout(playNote, 250);
        };
        
        playNote();
    }

    stopMusic() {
        this.isMusicPlaying = false;
        
        this.musicOscillators.forEach(({ oscillator, gain }) => {
            gain.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + 0.5);
            oscillator.stop(this.audioContext.currentTime + 0.5);
        });
        
        this.musicOscillators = [];
    }

    toggleMusic() {
        this.musicEnabled = !this.musicEnabled;
        
        if (this.musicEnabled && this.isMusicPlaying) {
            this.playMusic();
        } else if (!this.musicEnabled) {
            this.stopMusic();
        }
        
        return this.musicEnabled;
    }

    playSound(frequency, type = 'sine', duration = 0.1) {
        this.init();
        
        const oscillator = this.audioContext.createOscillator();
        const gain = this.audioContext.createGain();
        
        oscillator.type = type;
        oscillator.frequency.value = frequency;
        
        gain.gain.setValueAtTime(0.3, this.audioContext.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + duration);
        
        oscillator.connect(gain);
        gain.connect(this.masterGain);
        
        oscillator.start();
        oscillator.stop(this.audioContext.currentTime + duration);
    }

    playCollectSound() {
        this.playSound(880, 'sine', 0.1);
        setTimeout(() => this.playSound(1174.66, 'sine', 0.1), 50);
    }

    playHitSound() {
        this.playSound(150, 'sawtooth', 0.3);
    }

    playExplosionSound() {
        this.init();
        
        const bufferSize = this.audioContext.sampleRate * 0.5;
        const buffer = this.audioContext.createBuffer(1, bufferSize, this.audioContext.sampleRate);
        const data = buffer.getChannelData(0);
        
        for (let i = 0; i < bufferSize; i++) {
            data[i] = Math.random() * 2 - 1;
        }
        
        const noise = this.audioContext.createBufferSource();
        noise.buffer = buffer;
        
        const gain = this.audioContext.createGain();
        gain.gain.setValueAtTime(0.3, this.audioContext.currentTime);
        gain.gain.exponentialRampToValueAtTime(0.001, this.audioContext.currentTime + 0.5);
        
        const filter = this.audioContext.createBiquadFilter();
        filter.type = 'lowpass';
        filter.frequency.value = 1000;
        
        noise.connect(filter);
        filter.connect(gain);
        gain.connect(this.masterGain);
        
        noise.start();
    }
}
