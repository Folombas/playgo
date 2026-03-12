import Phaser from 'phaser';
import { GameState, Upgrade, goFacts } from './types';

export class GameScene extends Phaser.Scene {
  private gameState: GameState;
  private upgrades: Upgrade[];
  
  // UI Elements
  private energyBar!: Phaser.GameObjects.Rectangle;
  private energyFill!: Phaser.GameObjects.Rectangle;
  private scoreText!: Phaser.GameObjects.Text;
  private levelText!: Phaser.GameObjects.Text;
  private xpBar!: Phaser.GameObjects.Rectangle;
  private incomeText!: Phaser.GameObjects.Text;
  private energyText!: Phaser.GameObjects.Text;
  private factText!: Phaser.GameObjects.Text;
  
  // Game objects
  private tapButton!: Phaser.GameObjects.Image;
  private floatingTexts: Phaser.GameObjects.Text[] = [];
  private lastRegenTime: number = 0;
  private lastAutoTapTime: number = 0;
  private currentFactIndex: number = 0;
  
  // Audio
  private audioEnabled: boolean = true;
  private audioCtx: AudioContext | null = null;

  // Callbacks
  onScoreChange: ((score: number) => void) | null = null;
  onEnergyChange: ((energy: number, maxEnergy: number) => void) | null = null;
  onLevelChange: ((level: number, xp: number, xpToNext: number) => void) | null = null;
  onIncomeChange: ((income: number) => void) | null = null;
  onUpgradePurchased: ((upgradeId: string) => void) | null = null;
  onToggleAudio: ((enabled: boolean) => void) | null = null;

  constructor() {
    super({ key: 'GameScene' });
    this.gameState = {
      score: 0,
      energy: 100,
      maxEnergy: 100,
      energyRegen: 1,
      tapValue: 1,
      autoTapPerSec: 0,
      level: 1,
      xp: 0,
      xpToNextLevel: 100,
    };
    this.upgrades = [];
  }

  setGameState(state: GameState) {
    this.gameState = { ...state };
  }

  setUpgrades(upgrades: Upgrade[]) {
    this.upgrades = upgrades.map(u => ({ ...u }));
  }

  preload(): void {
    const graphics = this.make.graphics({ x: 0, y: 0 });

    // Go gopher mascot
    graphics.fillStyle(0x00ADD8);
    graphics.fillCircle(50, 50, 45);
    graphics.fillStyle(0xFFFFFF);
    graphics.fillCircle(35, 40, 12);
    graphics.fillCircle(65, 40, 12);
    graphics.fillStyle(0x000000);
    graphics.fillCircle(35, 40, 5);
    graphics.fillCircle(65, 40, 5);
    graphics.fillStyle(0xFF69B4);
    graphics.fillEllipse(50, 55, 20, 10);
    graphics.generateTexture('gopher', 100, 100);
    graphics.clear();

    // Energy icon
    graphics.fillStyle(0xFFD700);
    graphics.beginPath();
    graphics.moveTo(50, 10);
    graphics.lineTo(30, 40);
    graphics.lineTo(50, 40);
    graphics.lineTo(45, 70);
    graphics.lineTo(70, 35);
    graphics.lineTo(50, 35);
    graphics.closePath();
    graphics.fillPath();
    graphics.generateTexture('energy', 80, 80);
    graphics.clear();

    // Particle
    graphics.fillStyle(0x00ADD8);
    graphics.fillCircle(5, 5, 5);
    graphics.generateTexture('particle', 10, 10);
    graphics.clear();

    // Upgrade icons
    const colors = [0x4CAF50, 0x2196F3, 0x9C27B0, 0xFF5722, 0xFFC107, 0x00BCD4];
    this.upgrades.forEach((upgrade, index) => {
      graphics.fillStyle(colors[index]);
      graphics.fillCircle(20, 20, 18);
      graphics.fillStyle(0xFFFFFF);
      graphics.fillCircle(20, 20, 12);
      graphics.generateTexture(upgrade.id, 40, 40);
      graphics.clear();
    });

    graphics.clear();
  }

  create(): void {
    const initAudio = () => {
      if (!this.audioCtx) {
        this.audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
      }
      if (this.audioCtx.state === 'suspended') {
        this.audioCtx.resume();
      }
    };

    const { width, height } = this.scale;

    // Background
    const bgGraphics = this.add.graphics();
    bgGraphics.fillStyle(0x1a1a2e);
    bgGraphics.fillRect(0, 0, width, height);

    // Grid pattern
    bgGraphics.lineStyle(1, 0x2a2a4e, 0.3);
    for (let x = 0; x < width; x += 30) {
      bgGraphics.moveTo(x, 0);
      bgGraphics.lineTo(x, height);
    }
    for (let y = 0; y < height; y += 30) {
      bgGraphics.moveTo(0, y);
      bgGraphics.lineTo(width, y);
    }
    bgGraphics.strokePath();

    // Header background
    const headerBg = this.add.graphics();
    headerBg.fillStyle(0x16213e, 0.8);
    headerBg.fillRoundedRect(15, 15, width - 30, 130, 15);

    // Level text
    this.levelText = this.add.text(30, 28, 'Уровень 1', {
      fontFamily: 'Arial',
      fontSize: '18px',
      color: '#00ADD8',
      fontStyle: 'bold',
    });

    // XP bar background
    this.add.rectangle(width / 2, 58, width - 60, 12, 0x2a2a4e).setOrigin(0.5);
    this.xpBar = this.add.rectangle(30, 58, 0, 12, 0x00ADD8).setOrigin(0, 0.5);

    // Score text
    this.scoreText = this.add.text(width / 2, 85, '0 Гоферокоинов', {
      fontFamily: 'Arial',
      fontSize: '24px',
      color: '#FFD700',
      fontStyle: 'bold',
    }).setOrigin(0.5);

    // Income text
    this.incomeText = this.add.text(width / 2, 115, '+0/сек', {
      fontFamily: 'Arial',
      fontSize: '15px',
      color: '#4CAF50',
    }).setOrigin(0.5);

    // Energy bar
    this.add.text(30, 165, '⚡ Энергия', {
      fontFamily: 'Arial',
      fontSize: '16px',
      color: '#FFD700',
    });

    this.energyBar = this.add.rectangle(30, 190, width - 60, 20, 0x2a2a4e).setOrigin(0, 0.5);
    this.energyFill = this.add.rectangle(30, 190, width - 60, 20, 0xFFD700).setOrigin(0, 0.5);

    // Tap area with Gopher
    const centerX = width / 2;
    const centerY = height / 2;

    // Glow effect
    this.add.circle(centerX, centerY, 120, 0x00ADD8, 0.3);

    // Gopher button
    this.tapButton = this.add.image(centerX, centerY, 'gopher');
    this.tapButton.setScale(1.5);
    this.tapButton.setInteractive({ useHandCursor: true });

    // Hit area
    const hitArea = this.add.circle(centerX, centerY, 100);
    hitArea.setInteractive({ useHandCursor: true });
    
    hitArea.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
      initAudio();
      this.handleTap(pointer.x, pointer.y);
    });

    // Energy text
    this.energyText = this.add.text(centerX, centerY + 130, '100/100', {
      fontFamily: 'Arial',
      fontSize: '20px',
      color: '#FFD700',
      fontStyle: 'bold',
    }).setOrigin(0.5);

    // Fact text
    this.factText = this.add.text(centerX, height - 30, this.getNewFact(), {
      fontFamily: 'Arial',
      fontSize: '12px',
      color: '#888888',
      fontStyle: 'italic',
      wordWrap: { width: width - 40 }
    }).setOrigin(0.5);

    // Update fact every 10 seconds
    this.time.addEvent({
      delay: 10000,
      callback: () => {
        this.factText.setText(this.getNewFact());
      },
      loop: true,
    });

    this.updateUI();
  }

  private playSound(frequency: number, duration: number, type: OscillatorType = 'sine') {
    if (!this.audioEnabled || !this.audioCtx) return;
    
    try {
      const oscillator = this.audioCtx.createOscillator();
      const gainNode = this.audioCtx.createGain();
      
      oscillator.connect(gainNode);
      gainNode.connect(this.audioCtx.destination);
      
      oscillator.frequency.value = frequency;
      oscillator.type = type;
      
      gainNode.gain.setValueAtTime(0.3, this.audioCtx.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioCtx.currentTime + duration);
      
      oscillator.start(this.audioCtx.currentTime);
      oscillator.stop(this.audioCtx.currentTime + duration);
    } catch (e) {
      // Ignore audio errors
    }
  }

  public toggleAudio(): void {
    this.audioEnabled = !this.audioEnabled;
    if (this.onToggleAudio) {
      this.onToggleAudio(this.audioEnabled);
    }
  }

  update(time: number): void {
    // Energy regeneration
    if (time - this.lastRegenTime > 1000 && this.gameState.energy < this.gameState.maxEnergy) {
      this.gameState.energy = Math.min(this.gameState.maxEnergy, this.gameState.energy + this.gameState.energyRegen);
      this.lastRegenTime = time;
      this.updateUI();
      if (this.onEnergyChange) {
        this.onEnergyChange(this.gameState.energy, this.gameState.maxEnergy);
      }
    }

    // Auto tap from upgrades
    if (this.gameState.autoTapPerSec > 0 && time - this.lastAutoTapTime > 1000) {
      this.addScore(this.gameState.autoTapPerSec, false);
      this.lastAutoTapTime = time;
    }

    // Update floating texts
    this.floatingTexts = this.floatingTexts.filter((text) => {
      if (text.alpha <= 0) {
        text.destroy();
        return false;
      }
      text.y -= 1;
      text.alpha -= 0.02;
      return true;
    });
  }

  private handleTap(x: number, y: number): void {
    if (this.gameState.energy < 1) return;

    this.gameState.energy -= 1;
    this.addScore(this.gameState.tapValue, true);

    // Animation
    this.tweens.add({
      targets: this.tapButton,
      scaleX: 1.3,
      scaleY: 1.3,
      duration: 50,
      yoyo: true,
    });

    // Sound effect
    this.playSound(800, 0.1);

    // Floating text
    this.createFloatingText(x, y - 50, `+${this.gameState.tapValue}`);

    // Particles
    this.createParticles(x, y);

    this.updateUI();
    if (this.onEnergyChange) {
      this.onEnergyChange(this.gameState.energy, this.gameState.maxEnergy);
    }
  }

  private addScore(amount: number, isTap: boolean): void {
    const oldLevel = this.gameState.level;
    this.gameState.score += amount;
    if (isTap) {
      this.gameState.xp += 1;
    }

    // Level up check
    if (this.gameState.xp >= this.gameState.xpToNextLevel) {
      this.gameState.level++;
      this.gameState.xp = 0;
      this.gameState.xpToNextLevel = Math.floor(this.gameState.xpToNextLevel * 1.5);
      this.gameState.tapValue++;

      this.createFloatingText(this.scale.width / 2, this.scale.height / 2, 'УРОВЕНЬ ПОВЫШЕН!', '#00FF00', 28);
      
      // Level up sound
      this.playSound(523, 0.15);
      setTimeout(() => this.playSound(659, 0.15), 100);
      setTimeout(() => this.playSound(784, 0.3), 200);
    }

    this.updateUI();
    if (this.onScoreChange) {
      this.onScoreChange(this.gameState.score);
    }
    if (this.onLevelChange) {
      this.onLevelChange(this.gameState.level, this.gameState.xp, this.gameState.xpToNextLevel);
    }
  }

  public buyUpgrade(upgradeIndex: number): void {
    const upgrade = this.upgrades[upgradeIndex];
    if (!upgrade) return;

    const cost = Math.floor(upgrade.baseCost * Math.pow(1.15, upgrade.count));

    if (this.gameState.score >= cost) {
      this.gameState.score -= cost;
      upgrade.count++;
      this.gameState.autoTapPerSec += upgrade.income;

      this.createFloatingText(this.scale.width / 2, this.scale.height / 2 - 100, `+${upgrade.income}/сек`, '#4CAF50');
      
      // Upgrade sound
      this.playSound(440, 0.1, 'square');
      setTimeout(() => this.playSound(880, 0.15, 'square'), 80);
      
      this.updateUI();
      if (this.onScoreChange) {
        this.onScoreChange(this.gameState.score);
      }
      if (this.onIncomeChange) {
        this.onIncomeChange(this.gameState.autoTapPerSec);
      }
      if (this.onUpgradePurchased) {
        this.onUpgradePurchased(upgrade.id);
      }
    }
  }

  private updateUI(): void {
    const { width } = this.scale;

    this.scoreText.setText(`${Math.floor(this.gameState.score)} Гоферокоинов`);
    this.levelText.setText(`Уровень ${this.gameState.level}`);
    this.xpBar.width = (this.gameState.xp / this.gameState.xpToNextLevel) * (width - 60);
    this.energyFill.width = (this.gameState.energy / this.gameState.maxEnergy) * (width - 60);
    this.energyText.setText(`${Math.floor(this.gameState.energy)}/${this.gameState.maxEnergy}`);
    this.incomeText.setText(`+${this.gameState.autoTapPerSec.toFixed(1)}/сек`);
  }

  private createFloatingText(x: number, y: number, text: string, color: string = '#FFFFFF', size: number = 20): void {
    const floatText = this.add.text(x, y, text, {
      fontFamily: 'Arial',
      fontSize: `${size}px`,
      color,
      fontStyle: 'bold',
      stroke: '#000000',
      strokeThickness: 3,
    }).setOrigin(0.5);

    this.floatingTexts.push(floatText);
  }

  private createParticles(x: number, y: number): void {
    const particles = this.add.particles(x, y, 'particle', {
      speed: { min: 50, max: 150 },
      angle: { min: 0, max: 360 },
      scale: { start: 0.5, end: 0 },
      alpha: { start: 1, end: 0 },
      lifespan: 500,
      quantity: 5,
      blendMode: 'ADD',
    });

    this.time.delayedCall(500, () => particles.destroy());
  }

  private getNewFact(): string {
    const newIndex = (this.currentFactIndex + 1) % goFacts.length;
    this.currentFactIndex = newIndex;
    return goFacts[newIndex];
  }

  getGameState(): GameState {
    return { ...this.gameState };
  }

  getUpgrades(): Upgrade[] {
    return this.upgrades.map(u => ({ ...u }));
  }

  isAudioEnabled(): boolean {
    return this.audioEnabled;
  }
}
