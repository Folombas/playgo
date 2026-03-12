import Phaser from 'phaser';
import { GameState, Upgrade, goFacts, Achievement } from './types';

export class GameScene extends Phaser.Scene {
  private gameState: GameState;
  private upgrades: Upgrade[];
  private achievements: Achievement[];
  
  // UI Elements - Top Bar
  private levelText!: Phaser.GameObjects.Text;
  private levelBadge!: Phaser.GameObjects.Container;
  private xpBar!: Phaser.GameObjects.Rectangle;
  private xpFill!: Phaser.GameObjects.Rectangle;
  private scoreText!: Phaser.GameObjects.Text;
  private scoreContainer!: Phaser.GameObjects.Container;
  private incomeText!: Phaser.GameObjects.Text;
  
  // UI Elements - Energy Bar
  private energyBar!: Phaser.GameObjects.Container;
  private energyFill!: Phaser.GameObjects.Rectangle;
  private energyText!: Phaser.GameObjects.Text;
  
  // Game objects
  private tapButton!: Phaser.GameObjects.Container;
  private gopherImage!: Phaser.GameObjects.Image;
  private comboText!: Phaser.GameObjects.Text;
  private comboContainer!: Phaser.GameObjects.Container;
  private floatingTexts: Phaser.GameObjects.Text[] = [];
  
  // Timers
  private lastRegenTime: number = 0;
  private lastAutoTapTime: number = 0;
  private currentFactIndex: number = 0;
  private combo: number = 0;
  private comboTimer: number = 0;
  private maxCombo: number = 0;
  
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
  onAchievementUnlocked: ((achievement: Achievement) => void) | null = null;

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
    this.achievements = [
      { id: 'first_blood', name: 'Первый тап', description: 'Сделай первый тап', icon: '🎯', unlocked: false },
      { id: 'combo_10', name: 'Комбо мастер', description: 'Достигни комбо x10', icon: '🔥', unlocked: false },
      { id: 'level_5', name: 'Новичок', description: 'Достигни 5 уровня', icon: '⭐', unlocked: false },
      { id: 'rich', name: 'Богач', description: 'Накопи 1000 монет', icon: '💰', unlocked: false },
    ];
  }

  setGameState(state: GameState) {
    this.gameState = { ...state };
  }

  setUpgrades(upgrades: Upgrade[]) {
    this.upgrades = upgrades.map(u => ({ ...u }));
  }

  preload(): void {
    const graphics = this.make.graphics({ x: 0, y: 0 });

    // Go gopher mascot - bigger and more detailed
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
    // Add shine
    graphics.fillStyle(0x66D9E8);
    graphics.fillCircle(35, 35, 8);
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

    // Coin icon
    graphics.fillStyle(0xFFD700);
    graphics.fillCircle(25, 25, 20);
    graphics.fillStyle(0xFFA500);
    graphics.fillCircle(25, 25, 15);
    graphics.fillStyle(0xFFD700);
    graphics.fillCircle(25, 25, 10);
    graphics.generateTexture('coin', 50, 50);
    graphics.clear();

    // Star icon
    graphics.fillStyle(0xFFD700);
    // Draw star manually
    graphics.fillTriangle(25, 5, 20, 20, 30, 20);
    graphics.fillTriangle(25, 5, 25, 15, 35, 18);
    graphics.fillTriangle(25, 5, 15, 18, 25, 15);
    graphics.fillTriangle(25, 25, 15, 25, 25, 35);
    graphics.fillTriangle(25, 25, 35, 25, 25, 35);
    graphics.generateTexture('star', 50, 50);
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

    // === BACKGROUND ===
    const bgGraphics = this.add.graphics();
    bgGraphics.fillStyle(0x1a1a2e);
    bgGraphics.fillRect(0, 0, width, height);

    // Animated grid pattern
    bgGraphics.lineStyle(1, 0x2a2a4e, 0.3);
    for (let x = 0; x < width; x += 40) {
      bgGraphics.moveTo(x, 0);
      bgGraphics.lineTo(x, height);
    }
    for (let y = 0; y < height; y += 40) {
      bgGraphics.moveTo(0, y);
      bgGraphics.lineTo(width, y);
    }
    bgGraphics.strokePath();

    // === TOP BAR (Level, XP, Score, Income) ===
    const topBarY = 35;
    
    // Top bar background
    const topBarBg = this.add.graphics();
    topBarBg.fillStyle(0x16213e, 0.9);
    topBarBg.fillRoundedRect(10, 10, width - 20, 90, 15);
    topBarBg.lineStyle(2, 0x00ADD8, 0.5);
    topBarBg.strokeRoundedRect(10, 10, width - 20, 90, 15);

    // Level badge (left side)
    this.levelBadge = this.add.container(50, topBarY);
    const levelBg = this.add.circle(0, 0, 28, 0x00ADD8);
    const levelBorder = this.add.circle(0, 0, 28, 0x00ADD8).setStrokeStyle(3, 0x00FFFF);
    this.levelText = this.add.text(0, 0, '1', {
      fontFamily: 'Arial',
      fontSize: '20px',
      color: '#FFFFFF',
      fontStyle: 'bold',
    }).setOrigin(0.5);
    this.levelBadge.add([levelBg, levelBorder, this.levelText]);

    // XP bar (below level)
    this.add.text(50, topBarY + 38, 'Опыт', {
      fontFamily: 'Arial',
      fontSize: '10px',
      color: '#888888',
    }).setOrigin(0.5);
    this.xpBar = this.add.rectangle(50, topBarY + 52, 70, 8, 0x2a2a4e).setOrigin(0.5);
    this.xpFill = this.add.rectangle(50 - 32, topBarY + 52, 0, 6, 0x00ADD8).setOrigin(0, 0.5);

    // Score (center)
    this.scoreContainer = this.add.container(width / 2, topBarY);
    const coinIcon = this.add.image(-70, 0, 'coin').setScale(0.5);
    this.scoreText = this.add.text(0, 0, '0', {
      fontFamily: 'Arial',
      fontSize: '28px',
      color: '#FFD700',
      fontStyle: 'bold',
    }).setOrigin(0.5);
    this.scoreContainer.add([coinIcon, this.scoreText]);

    // Income (below score)
    this.incomeText = this.add.text(width / 2, topBarY + 35, '+0/сек', {
      fontFamily: 'Arial',
      fontSize: '14px',
      color: '#4CAF50',
      fontStyle: 'bold',
    }).setOrigin(0.5);

    // === ENERGY BAR (below top bar) ===
    this.energyBar = this.add.container(width / 2, 125);
    const energyBg = this.add.rectangle(0, 0, width - 40, 28, 0x2a2a4e).setOrigin(0.5);
    energyBg.setStrokeStyle(2, 0xFFD700, 0.5);
    this.energyFill = this.add.rectangle(-(width/2 - 20), 0, width - 40, 28, 0xFFD700).setOrigin(0, 0.5);
    const energyIcon = this.add.image(-(width/2 - 60), 0, 'energy').setScale(0.4);
    this.energyText = this.add.text(0, 0, '100/100', {
      fontFamily: 'Arial',
      fontSize: '16px',
      color: '#FFFFFF',
      fontStyle: 'bold',
    }).setOrigin(0.5);
    this.energyBar.add([energyBg, this.energyFill, energyIcon, this.energyText]);

    // === MAIN TAP AREA (center) ===
    const centerX = width / 2;
    const centerY = height / 2 + 30;

    // Outer glow ring
    const outerGlow = this.add.circle(centerX, centerY, 140, 0x00ADD8, 0.2);
    
    // Inner glow ring
    const innerGlow = this.add.circle(centerX, centerY, 110, 0x00ADD8, 0.3);

    // Tap button container
    this.tapButton = this.add.container(centerX, centerY);
    
    // Background circle
    const tapBg = this.add.circle(0, 0, 100, 0x00ADD8, 0.3);
    tapBg.setStrokeStyle(4, 0x00FFFF, 0.8);
    
    // Gopher image
    this.gopherImage = this.add.image(0, 0, 'gopher').setScale(1.6);
    
    // Hit area (invisible but interactive)
    const hitArea = this.add.circle(0, 0, 95);
    hitArea.setInteractive({ useHandCursor: true });
    hitArea.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
      initAudio();
      this.handleTap(pointer.x, pointer.y);
    });
    hitArea.on('pointerup', () => {
      this.tweens.add({
        targets: this.gopherImage,
        scaleX: 1.6,
        scaleY: 1.6,
        duration: 100,
      });
    });

    this.tapButton.add([tapBg, this.gopherImage, hitArea]);

    // === COMBO DISPLAY (above tap button) ===
    this.comboContainer = this.add.container(centerX, centerY - 140);
    this.comboText = this.add.text(0, 0, '', {
      fontFamily: 'Arial',
      fontSize: '24px',
      color: '#FF6B6B',
      fontStyle: 'bold',
      stroke: '#000000',
      strokeThickness: 4,
    }).setOrigin(0.5);
    this.comboContainer.add(this.comboText);
    this.comboContainer.setVisible(false);

    // === FACT TEXT (bottom) ===
    this.factText = this.add.text(centerX, height - 35, this.getNewFact(), {
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

    // Check achievements periodically
    this.time.addEvent({
      delay: 1000,
      callback: () => {
        this.checkAchievements();
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
    } catch (e) {}
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

    // Combo decay
    if (this.combo > 0) {
      this.comboTimer += 16; // ~16ms per frame
      if (this.comboTimer > 2000) { // 2 seconds to maintain combo
        this.combo = 0;
        this.comboTimer = 0;
        this.maxCombo = 0;
        this.comboContainer.setVisible(false);
      }
    }

    // Update floating texts
    this.floatingTexts = this.floatingTexts.filter((text) => {
      if (text.alpha <= 0) {
        text.destroy();
        return false;
      }
      text.y -= 2;
      text.alpha -= 0.02;
      return true;
    });

    // Animate glow
    const pulse = 0.2 + Math.sin(time / 200) * 0.1;
    // Note: Can't directly animate graphics in update without recreation
  }

  private handleTap(x: number, y: number): void {
    if (this.gameState.energy < 1) return;

    this.gameState.energy -= 1;
    
    // Combo system
    this.combo++;
    this.comboTimer = 0;
    if (this.combo > this.maxCombo) {
      this.maxCombo = this.combo;
    }
    
    // Show combo
    if (this.combo >= 5) {
      this.comboContainer.setVisible(true);
      this.comboText.setText(`🔥 x${this.combo} КОМБО!`);
      this.comboText.setScale(1);
      this.tweens.add({
        targets: this.comboText,
        scaleX: 1.3,
        scaleY: 1.3,
        duration: 100,
        yoyo: true,
      });
    }

    // Calculate score with combo bonus
    const comboMultiplier = 1 + (this.combo - 1) * 0.1; // +10% per combo level
    const earnedScore = Math.floor(this.gameState.tapValue * comboMultiplier);
    
    this.addScore(earnedScore, true);

    // Animation
    this.tweens.add({
      targets: this.gopherImage,
      scaleX: 1.3,
      scaleY: 1.3,
      duration: 50,
      yoyo: true,
    });

    // Sound effect
    this.playSound(800 + this.combo * 50, 0.1);

    // Floating text with combo color
    const textColor = this.combo >= 10 ? '#FF00FF' : this.combo >= 5 ? '#FF6B6B' : '#FFFFFF';
    const displayText = this.combo >= 5 ? `+${earnedScore} 🔥` : `+${earnedScore}`;
    this.createFloatingText(x, y - 50, displayText, textColor, 24);

    // Particles
    this.createParticles(x, y);

    this.updateUI();
    if (this.onEnergyChange) {
      this.onEnergyChange(this.gameState.energy, this.gameState.maxEnergy);
    }
  }

  private addScore(amount: number, isTap: boolean): void {
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

      this.createFloatingText(this.scale.width / 2, this.scale.height / 2 - 50, '⭐ УРОВЕНЬ ' + this.gameState.level + '! ⭐', '#00FF00', 32);
      
      // Level up sound
      this.playSound(523, 0.15);
      setTimeout(() => this.playSound(659, 0.15), 100);
      setTimeout(() => this.playSound(784, 0.3), 200);
      
      // Celebration particles
      for (let i = 0; i < 3; i++) {
        setTimeout(() => {
          this.createParticles(this.scale.width / 2 + (Math.random() - 0.5) * 100, this.scale.height / 2);
        }, i * 200);
      }
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

      this.createFloatingText(this.scale.width / 2, this.scale.height / 2 - 100, `+${upgrade.income}/сек`, '#4CAF50', 24);
      
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

  private checkAchievements(): void {
    // First tap
    if (this.gameState.score >= 1 && !this.achievements[0].unlocked) {
      this.achievements[0].unlocked = true;
      this.unlockAchievement(this.achievements[0]);
    }
    
    // Combo 10
    if (this.maxCombo >= 10 && !this.achievements[1].unlocked) {
      this.achievements[1].unlocked = true;
      this.unlockAchievement(this.achievements[1]);
    }
    
    // Level 5
    if (this.gameState.level >= 5 && !this.achievements[2].unlocked) {
      this.achievements[2].unlocked = true;
      this.unlockAchievement(this.achievements[2]);
    }
    
    // Rich 1000
    if (this.gameState.score >= 1000 && !this.achievements[3].unlocked) {
      this.achievements[3].unlocked = true;
      this.unlockAchievement(this.achievements[3]);
    }
  }

  private unlockAchievement(achievement: Achievement): void {
    if (this.onAchievementUnlocked) {
      this.onAchievementUnlocked(achievement);
    }
    
    // Show achievement notification
    const { width } = this.scale;
    const notifyContainer = this.add.container(width / 2, 100);
    
    const notifyBg = this.add.rectangle(0, 0, 280, 70, 0x16213e, 0.95);
    notifyBg.setStrokeStyle(2, 0xFFD700);
    notifyBg.setOrigin(0.5);
    
    const icon = this.add.text(-100, 0, achievement.icon, {
      fontSize: '36px',
    }).setOrigin(0.5);
    
    const title = this.add.text(20, -15, '🏆 ДОСТИЖЕНИЕ!', {
      fontFamily: 'Arial',
      fontSize: '14px',
      color: '#FFD700',
      fontStyle: 'bold',
    }).setOrigin(0, 0.5);
    
    const name = this.add.text(20, 10, achievement.name, {
      fontFamily: 'Arial',
      fontSize: '16px',
      color: '#FFFFFF',
      fontStyle: 'bold',
    }).setOrigin(0, 0.5);

    notifyContainer.add([notifyBg, icon, title, name]);
    
    // Animate in
    notifyContainer.setY(-100);
    this.tweens.add({
      targets: notifyContainer,
      y: 100,
      duration: 500,
      ease: 'Back.out',
    });
    
    // Remove after 3 seconds
    this.time.delayedCall(3000, () => {
      this.tweens.add({
        targets: notifyContainer,
        y: -100,
        duration: 300,
        onComplete: () => notifyContainer.destroy(),
      });
    });
  }

  private updateUI(): void {
    const { width } = this.scale;

    this.scoreText.setText(Math.floor(this.gameState.score).toLocaleString());
    this.levelText.setText(this.gameState.level.toString());
    this.xpFill.width = (this.gameState.xp / this.gameState.xpToNextLevel) * 64;
    this.energyFill.width = (this.gameState.energy / this.gameState.maxEnergy) * (width - 40);
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
      strokeThickness: 4,
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

  private factText!: Phaser.GameObjects.Text;
  
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

  getAchievements(): Achievement[] {
    return this.achievements.map(a => ({ ...a }));
  }

  isAudioEnabled(): boolean {
    return this.audioEnabled;
  }
}
