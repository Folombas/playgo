import Phaser from 'phaser';
import { GameState, Upgrade, Quest } from './types';

export class GameScene extends Phaser.Scene {
  private gameState: GameState;
  private upgrades: Upgrade[];
  private quests: Quest[];
  
  // UI Elements
  private scoreText!: Phaser.GameObjects.Text;
  private incomeText!: Phaser.GameObjects.Text;
  private xpFill!: Phaser.GameObjects.Rectangle;
  private energyFill!: Phaser.GameObjects.Rectangle;
  private energyText!: Phaser.GameObjects.Text;
  private gopher!: Phaser.GameObjects.Image;
  private comboText!: Phaser.GameObjects.Text;
  private streakText!: Phaser.GameObjects.Text;
  private comboContainer!: Phaser.GameObjects.Container;
  
  // Visual elements
  private backgroundParticles: Phaser.GameObjects.Arc[] = [];
  
  // Timers
  private lastRegenTime: number = 0;
  private lastAutoTapTime: number = 0;
  private combo: number = 0;
  private comboTimer: number = 0;
  private maxCombo: number = 0;
  private clickStreak: number = 0;
  private streakTimer: number = 0;
  private criticalHitChance: number = 0.1;
  private lastCriticalHit: number = 0;
  
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
  onQuestCompleted: ((quest: Quest) => void) | null = null;

  constructor() {
    super({ key: 'GameScene' });
    this.gameState = {
      score: 0,
      energy: 100,
      maxEnergy: 100,
      energyRegen: 2,
      tapValue: 1,
      autoTapPerSec: 0,
      level: 1,
      xp: 0,
      xpToNextLevel: 50,
      totalTaps: 0,
      criticalHits: 0,
    };
    this.upgrades = [];
    this.quests = [
      { id: 'daily_taps', name: '100 тапов за день', progress: 0, target: 100, reward: 500, completed: false, type: 'daily' },
      { id: 'level_up', name: 'Повысь уровень 3 раза', progress: 0, target: 3, reward: 1000, completed: false, type: 'quest' },
      { id: 'combo_master', name: 'Комбо x25', progress: 0, target: 25, reward: 750, completed: false, type: 'quest' },
    ];
  }

  setGameState(state: GameState) { this.gameState = { ...state }; }
  setUpgrades(upgrades: Upgrade[]) { this.upgrades = upgrades.map(u => ({ ...u })); }

  preload(): void {
    const graphics = this.make.graphics({ x: 0, y: 0 });

    // Gopher
    graphics.fillStyle(0x00ADD8);
    graphics.fillCircle(50, 50, 45);
    graphics.fillStyle(0xFFFFFF);
    graphics.fillCircle(35, 40, 12);
    graphics.fillCircle(65, 40, 12);
    graphics.fillStyle(0x000000);
    graphics.fillCircle(35, 40, 5);
    graphics.fillCircle(65, 40, 5);
    graphics.fillStyle(0xFF1493);
    graphics.fillEllipse(50, 55, 20, 10);
    graphics.lineStyle(3, 0x00FFFF, 0.8);
    graphics.strokeCircle(50, 50, 48);
    graphics.generateTexture('gopher', 100, 100);
    graphics.clear();

    // Energy crystal
    graphics.fillStyle(0x00FF88);
    graphics.fillTriangle(25, 5, 15, 45, 35, 45);
    graphics.generateTexture('energy_crystal', 50, 50);
    graphics.clear();

    // Coin
    graphics.fillStyle(0xFFD700);
    graphics.fillCircle(25, 25, 20);
    graphics.fillStyle(0xFFA500);
    graphics.fillCircle(25, 25, 14);
    graphics.generateTexture('coin_glow', 50, 50);
    graphics.clear();

    // Lightning
    graphics.fillStyle(0xFF00FF);
    graphics.beginPath();
    graphics.moveTo(25, 5);
    graphics.lineTo(15, 25);
    graphics.lineTo(22, 25);
    graphics.lineTo(18, 45);
    graphics.lineTo(35, 20);
    graphics.lineTo(28, 20);
    graphics.closePath();
    graphics.fillPath();
    graphics.generateTexture('lightning', 50, 50);
    graphics.clear();

    // Particles
    graphics.fillStyle(0x00FFFF);
    graphics.fillCircle(4, 4, 4);
    graphics.generateTexture('particle_neon', 8, 8);
    graphics.clear();

    graphics.fillStyle(0xFF00FF);
    graphics.fillCircle(3, 3, 3);
    graphics.generateTexture('particle_pink', 6, 6);
    graphics.clear();

    // Upgrade icons
    const colors = ['#00FF88', '#00BFFF', '#DA70D6', '#FF1493', '#FFD700', '#00FFFF'];
    this.upgrades.forEach((upgrade, index) => {
      graphics.fillStyle(parseInt(colors[index].slice(1), 16));
      graphics.fillCircle(20, 20, 18);
      graphics.lineStyle(2, 0xFFFFFF, 0.6);
      graphics.strokeCircle(20, 20, 20);
      graphics.generateTexture(upgrade.id, 40, 40);
      graphics.clear();
    });
  }

  create(): void {
    // Initialize audio on first interaction
    const initAudio = () => {
      if (!this.audioCtx) {
        this.audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
      }
      if (this.audioCtx.state === 'suspended') {
        this.audioCtx.resume();
      }
    };

    const { width, height } = this.scale;

    // BACKGROUND
    const bgGraphics = this.add.graphics();
    bgGraphics.fillStyle(0x0a0a1a);
    bgGraphics.fillRect(0, 0, width, height);
    bgGraphics.lineStyle(1, 0x00FFFF, 0.15);
    for (let x = 0; x < width; x += 50) {
      bgGraphics.moveTo(x, 0);
      bgGraphics.lineTo(x, height);
    }
    for (let y = 0; y < height; y += 50) {
      bgGraphics.moveTo(0, y);
      bgGraphics.lineTo(width, y);
    }
    bgGraphics.strokePath();

    for (let i = 0; i < 20; i++) {
      const particle = this.add.circle(Math.random() * width, Math.random() * height, Math.random() * 3 + 1, Math.random() > 0.5 ? 0x00FFFF : 0xFF00FF, Math.random() * 0.5 + 0.2);
      this.backgroundParticles.push(particle);
    }

    // TOP BAR
    const topBarBg = this.add.rectangle(width / 2, 50, width - 20, 85, 0x1a1a3a, 0.8);
    topBarBg.setStrokeStyle(2, 0x00FFFF, 0.5);
    topBarBg.setOrigin(0.5);

    const levelText = this.add.text(70, 50, `${this.gameState.level}`, {
      fontFamily: 'Arial Black', fontSize: '22px', color: '#00FFFF', fontStyle: 'bold', stroke: '#000000', strokeThickness: 4,
    }).setOrigin(0.5);
    this.add.circle(70, 50, 32, 0x000000, 0.6).setStrokeStyle(3, 0x00FFFF);

    const xpBarBg = this.add.rectangle(70, 78, 80, 10, 0x000000, 0.8);
    xpBarBg.setStrokeStyle(1, 0x00FFFF, 0.3);
    this.xpFill = this.add.rectangle(70 - 38, 78, 0, 8, 0x00FFFF).setOrigin(0, 0.5);

    const coinIcon = this.add.image(width / 2 - 80, 50, 'coin_glow').setScale(0.6);
    this.scoreText = this.add.text(width / 2, 50, '0', {
      fontFamily: 'Arial Black', fontSize: '32px', color: '#FFD700', fontStyle: 'bold', stroke: '#000000', strokeThickness: 6,
    }).setOrigin(0.5);

    this.incomeText = this.add.text(width / 2, 78, '+0/сек', {
      fontFamily: 'Arial', fontSize: '14px', color: '#00FF88', fontStyle: 'bold', stroke: '#000000', strokeThickness: 3,
    }).setOrigin(0.5);

    // ENERGY BAR
    const energyBg = this.add.rectangle(width - 100, height - 40, 90, 25, 0x000000, 0.8);
    energyBg.setStrokeStyle(2, 0x00FF88, 0.6);
    this.energyFill = this.add.rectangle(width - 140, height - 40, 80, 25, 0x00FF88).setOrigin(0, 0.5);
    this.add.image(width - 160, height - 40, 'energy_crystal').setScale(0.5);
    this.energyText = this.add.text(width - 100, height - 40, '100', {
      fontFamily: 'Arial', fontSize: '14px', color: '#FFFFFF', fontStyle: 'bold',
    }).setOrigin(0.5);

    // TAP AREA
    const centerX = width / 2;
    const centerY = height / 2 + 20;

    const outerRing = this.add.circle(centerX, centerY, 150, 0x000000, 0);
    outerRing.setStrokeStyle(3, 0x00FFFF, 0.4);
    this.tweens.add({ targets: outerRing, rotation: Math.PI * 2, duration: 10000, repeat: -1, ease: 'Linear' });

    const glowBg = this.add.circle(centerX, centerY, 110, 0x00ADD8, 0.2);
    glowBg.setStrokeStyle(4, 0x00FFFF, 0.8);
    this.tweens.add({ targets: glowBg, scaleX: 1.05, scaleY: 1.05, duration: 1000, repeat: -1, yoyo: true, ease: 'Sine.easeInOut' });

    this.gopher = this.add.image(centerX, centerY, 'gopher').setScale(1.7);

    const hitArea = this.add.circle(centerX, centerY, 100);
    hitArea.setInteractive({ useHandCursor: true });
    hitArea.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
      initAudio();
      this.handleTap(pointer.x, pointer.y);
    });

    // COMBO DISPLAY
    this.comboContainer = this.add.container(width / 2, 130);
    this.comboText = this.add.text(0, 0, '', {
      fontFamily: 'Arial Black', fontSize: '28px', color: '#FF00FF', fontStyle: 'bold', stroke: '#000000', strokeThickness: 5,
    }).setOrigin(0.5);
    this.streakText = this.add.text(0, 30, '', {
      fontFamily: 'Arial', fontSize: '16px', color: '#00FFFF', fontStyle: 'bold', stroke: '#000000', strokeThickness: 3,
    }).setOrigin(0.5);
    this.comboContainer.add([this.comboText, this.streakText]);
    this.comboContainer.setVisible(false);

    // QUESTS
    const questsBg = this.add.rectangle(width / 2, height - 70, width - 20, 50, 0x000000, 0.7);
    questsBg.setStrokeStyle(2, 0xFF00FF, 0.4);
    questsBg.setOrigin(0.5);
    this.add.text(-width/2 + 15, height - 80, '📋 ЗАДАНИЯ', {
      fontFamily: 'Arial', fontSize: '12px', color: '#FF00FF', fontStyle: 'bold',
    }).setOrigin(0, 0);

    this.time.addEvent({ delay: 1000, callback: () => this.checkQuests(), loop: true });
    this.updateUI();
  }

  private playSound(frequency: number, duration: number, type: OscillatorType = 'sine'): void {
    if (!this.audioEnabled || !this.audioCtx) return;
    try {
      const oscillator = this.audioCtx.createOscillator();
      const gainNode = this.audioCtx.createGain();
      oscillator.connect(gainNode);
      gainNode.connect(this.audioCtx.destination);
      oscillator.frequency.value = frequency;
      oscillator.type = type;
      gainNode.gain.setValueAtTime(0.2, this.audioCtx.currentTime);
      gainNode.gain.exponentialRampToValueAtTime(0.01, this.audioCtx.currentTime + duration);
      oscillator.start(this.audioCtx.currentTime);
      oscillator.stop(this.audioCtx.currentTime + duration);
    } catch (e) {}
  }

  public toggleAudio(): void {
    this.audioEnabled = !this.audioEnabled;
    if (this.onToggleAudio) this.onToggleAudio(this.audioEnabled);
  }

  update(time: number): void {
    // Energy regen
    if (time - this.lastRegenTime > 1000 && this.gameState.energy < this.gameState.maxEnergy) {
      this.gameState.energy = Math.min(this.gameState.maxEnergy, this.gameState.energy + this.gameState.energyRegen);
      this.lastRegenTime = time;
      this.updateUI();
    }

    // Auto tap
    if (this.gameState.autoTapPerSec > 0 && time - this.lastAutoTapTime > 1000) {
      this.addScore(this.gameState.autoTapPerSec, false);
      this.lastAutoTapTime = time;
    }

    // Combo decay
    if (this.combo > 0) {
      this.comboTimer += 16;
      if (this.comboTimer > 2500) {
        this.combo = 0;
        this.comboTimer = 0;
        this.maxCombo = 0;
        this.comboContainer.setVisible(false);
      }
    }

    // Particles animation
    this.backgroundParticles.forEach((particle, i) => {
      particle.y += 0.5;
      particle.alpha = 0.3 + Math.sin(time / 500 + i) * 0.2;
      if (particle.y > this.scale.height) {
        particle.y = 0;
        particle.x = Math.random() * this.scale.width;
      }
    });
  }

  private handleTap(x: number, y: number): void {
    if (this.gameState.energy < 1) return;

    this.gameState.energy -= 1;
    this.gameState.totalTaps++;

    this.combo++;
    this.comboTimer = 0;
    if (this.combo > this.maxCombo) this.maxCombo = this.combo;

    this.clickStreak++;
    this.streakTimer = 0;

    const isCritical = Math.random() < this.criticalHitChance;
    if (isCritical) {
      this.lastCriticalHit = this.combo;
      this.gameState.criticalHits++;
    }

    if (this.combo >= 3) {
      this.comboContainer.setVisible(true);
      const multiplier = 1 + (this.combo - 1) * 0.1;
      this.comboText.setText(`x${multiplier.toFixed(1)} КОМБО 🔥`);
      this.streakText.setText(`Серия: ${this.clickStreak}`);
    }

    const comboMultiplier = 1 + (this.combo - 1) * 0.1;
    const streakBonus = Math.min(this.clickStreak * 0.05, 0.5);
    const criticalMultiplier = isCritical ? 2 : 1;
    const earnedScore = Math.floor(this.gameState.tapValue * comboMultiplier * (1 + streakBonus) * criticalMultiplier);

    this.addScore(earnedScore, true, isCritical);

    this.tweens.add({ targets: this.gopher, scaleX: 1.4, scaleY: 1.4, duration: 50, yoyo: true });

    this.playSound(isCritical ? 1200 : 800, 0.1);

    const textColor = isCritical ? '#FF00FF' : this.combo >= 10 ? '#FFD700' : '#00FFFF';
    const displayText = isCritical ? `⚡ x${earnedScore} КРИТ!` : `+${earnedScore}`;
    this.createFloatingText(x, y - 50, displayText, textColor, isCritical ? 32 : 24);

    this.createParticles(x, y, isCritical ? 'particle_pink' : 'particle_neon');

    if (isCritical) this.cameras.main.shake(100, 0.01);
  }

  private addScore(amount: number, isTap: boolean, isCritical: boolean = false): void {
    this.gameState.score += amount;
    if (isTap) this.gameState.xp += isCritical ? 2 : 1;

    if (this.gameState.xp >= this.gameState.xpToNextLevel) {
      this.gameState.level++;
      this.gameState.xp = 0;
      this.gameState.xpToNextLevel = Math.floor(this.gameState.xpToNextLevel * 1.3);
      this.gameState.tapValue += 2;

      this.createFloatingText(this.scale.width / 2, this.scale.height / 2 - 50, '⭐ УРОВЕНЬ ' + this.gameState.level + '! ⭐', '#00FF00', 36);
      this.playSound(523, 0.15);
      setTimeout(() => this.playSound(659, 0.15), 100);
      setTimeout(() => this.playSound(784, 0.3), 200);
    }

    this.updateUI();
    if (this.onScoreChange) this.onScoreChange(this.gameState.score);
    if (this.onLevelChange) this.onLevelChange(this.gameState.level, this.gameState.xp, this.gameState.xpToNextLevel);
  }

  public buyUpgrade(upgradeIndex: number): void {
    const upgrade = this.upgrades[upgradeIndex];
    if (!upgrade) return;
    const cost = Math.floor(upgrade.baseCost * Math.pow(1.15, upgrade.count));

    if (this.gameState.score >= cost) {
      this.gameState.score -= cost;
      upgrade.count++;
      this.gameState.autoTapPerSec += upgrade.income;
      this.createFloatingText(this.scale.width / 2, this.scale.height / 2 - 100, `+${upgrade.income}/сек`, '#00FF88', 28);
      this.playSound(440, 0.1, 'square');
      setTimeout(() => this.playSound(880, 0.15, 'square'), 80);
      this.updateUI();
      if (this.onScoreChange) this.onScoreChange(this.gameState.score);
      if (this.onIncomeChange) this.onIncomeChange(this.gameState.autoTapPerSec);
      if (this.onUpgradePurchased) this.onUpgradePurchased(upgrade.id);
    }
  }

  private checkQuests(): void {
    this.quests.forEach(quest => {
      if (quest.completed) return;
      if (quest.id === 'daily_taps') quest.progress = Math.min(this.gameState.totalTaps % 100, quest.target);
      else if (quest.id === 'level_up') quest.progress = Math.min(this.gameState.level - 1, quest.target);
      else if (quest.id === 'combo_master') quest.progress = Math.min(this.maxCombo, quest.target);

      if (quest.progress >= quest.target && !quest.completed) {
        quest.completed = true;
        this.gameState.score += quest.reward;
        if (this.onQuestCompleted) this.onQuestCompleted(quest);
        this.createFloatingText(this.scale.width / 2, 100, `✅ ${quest.name}! +${quest.reward}💰`, '#00FF00', 24);
      }
    });
  }

  private updateUI(): void {
    const { width } = this.scale;
    this.scoreText.setText(Math.floor(this.gameState.score).toLocaleString());
    this.incomeText.setText(`+${this.gameState.autoTapPerSec.toFixed(1)}/сек`);
    this.xpFill.width = (this.gameState.xp / this.gameState.xpToNextLevel) * 76;
    this.energyFill.width = (this.gameState.energy / this.gameState.maxEnergy) * 80;
    this.energyText.setText(`${Math.floor(this.gameState.energy)}`);
  }

  private createFloatingText(x: number, y: number, text: string, color: string = '#FFFFFF', size: number = 24): void {
    const floatText = this.add.text(x, y, text, {
      fontFamily: 'Arial Black', fontSize: `${size}px`, color, fontStyle: 'bold', stroke: '#000000', strokeThickness: 5,
    }).setOrigin(0.5);
    this.tweens.add({ targets: floatText, y: y - 100, alpha: 0, duration: 1000, onComplete: () => floatText.destroy() });
  }

  private createParticles(x: number, y: number, texture: string = 'particle_neon'): void {
    const particles = this.add.particles(x, y, texture, {
      speed: { min: 80, max: 200 }, angle: { min: 0, max: 360 }, scale: { start: 0.8, end: 0 }, alpha: { start: 1, end: 0 }, lifespan: 600, quantity: 8, blendMode: 'ADD',
    });
    this.time.delayedCall(600, () => particles.destroy());
  }

  getGameState(): GameState { return { ...this.gameState }; }
  getUpgrades(): Upgrade[] { return this.upgrades.map(u => ({ ...u })); }
  getQuests(): Quest[] { return this.quests.map(q => ({ ...q })); }
  isAudioEnabled(): boolean { return this.audioEnabled; }
}
