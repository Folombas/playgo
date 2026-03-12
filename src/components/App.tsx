import React, { useEffect, useRef, useState, useCallback } from 'react';
import Phaser from 'phaser';
import { GameScene } from '../game/GameScene';
import { GameState, Upgrade, Achievement, initialGameState, upgrades as initialUpgrades, getUpgradeCost } from '../game/types';

const App: React.FC = () => {
  const gameRef = useRef<Phaser.Game | null>(null);
  const sceneRef = useRef<GameScene | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  
  const [gameState, setGameState] = useState<GameState>(initialGameState);
  const [upgrades, setUpgrades] = useState<Upgrade[]>(initialUpgrades);
  const [achievements, setAchievements] = useState<Achievement[]>([]);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [audioEnabled, setAudioEnabled] = useState(true);
  const [activeTab, setActiveTab] = useState<'upgrades' | 'achievements'>('upgrades');

  const initGame = useCallback(() => {
    if (!containerRef.current || gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 400,
      height: 700,
      parent: containerRef.current,
      backgroundColor: '#1a1a2e',
      physics: {
        default: 'arcade',
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: GameScene,
      scale: {
        mode: Phaser.Scale.FIT,
        autoCenter: Phaser.Scale.CENTER_BOTH,
        width: 400,
        height: 700,
      },
    };

    gameRef.current = new Phaser.Game(config);

    const scene = gameRef.current.scene.getScene('GameScene') as GameScene;
    if (scene) {
      sceneRef.current = scene;
      scene.setGameState(gameState);
      scene.setUpgrades(upgrades);

      scene.onScoreChange = (score: number) => {
        setGameState(prev => ({ ...prev, score }));
      };

      scene.onEnergyChange = (energy: number, maxEnergy: number) => {
        setGameState(prev => ({ ...prev, energy, maxEnergy }));
      };

      scene.onLevelChange = (level: number, xp: number, xpToNext: number) => {
        setGameState(prev => ({ ...prev, level, xp, xpToNextLevel: xpToNext }));
      };

      scene.onIncomeChange = (income: number) => {
        setGameState(prev => ({ ...prev, autoTapPerSec: income }));
      };

      scene.onUpgradePurchased = (upgradeId: string) => {
        setUpgrades(prev => prev.map(u => {
          if (u.id === upgradeId) {
            return { ...u, count: u.count + 1 };
          }
          return u;
        }));
      };

      scene.onToggleAudio = (enabled: boolean) => {
        setAudioEnabled(enabled);
      };

      scene.onAchievementUnlocked = (achievement: Achievement) => {
        setAchievements(prev => {
          const exists = prev.find(a => a.id === achievement.id);
          if (exists) return prev;
          return [...prev, achievement];
        });
      };

      setAudioEnabled(scene.isAudioEnabled());
      setAchievements(scene.getAchievements());
    }
  }, []);

  useEffect(() => {
    initGame();

    return () => {
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
      }
    };
  }, [initGame]);

  useEffect(() => {
    if (sceneRef.current && upgrades) {
      sceneRef.current.setUpgrades(upgrades);
    }
  }, [upgrades]);

  const handleBuyUpgrade = (index: number) => {
    const upgrade = upgrades[index];
    const cost = getUpgradeCost(upgrade);

    if (gameState.score >= cost && sceneRef.current) {
      sceneRef.current.buyUpgrade(index);
    }
  };

  const togglePanel = () => {
    setIsPanelOpen(!isPanelOpen);
  };

  const toggleAudio = () => {
    if (sceneRef.current) {
      sceneRef.current.toggleAudio();
    }
  };

  const unlockedCount = achievements.filter(a => a.unlocked).length;

  return (
    <div className="game-container" ref={containerRef}>
      {/* Audio Toggle Button */}
      <button
        onClick={toggleAudio}
        title={audioEnabled ? 'Выключить звук' : 'Включить звук'}
        className="audio-btn"
      >
        {audioEnabled ? '🔊' : '🔇'}
      </button>

      {/* Upgrade Button */}
      <button
        onClick={togglePanel}
        className="upgrade-toggle-btn"
      >
        {isPanelOpen ? '✕' : '⬆'}
      </button>

      {/* Side Panel */}
      <div className={`side-panel ${isPanelOpen ? 'open' : ''}`}>
        {/* Panel Header with Tabs */}
        <div className="panel-header">
          <button
            className={`panel-tab ${activeTab === 'upgrades' ? 'active' : ''}`}
            onClick={() => setActiveTab('upgrades')}
          >
            🚀 Улучшения
          </button>
          <button
            className={`panel-tab ${activeTab === 'achievements' ? 'active' : ''}`}
            onClick={() => setActiveTab('achievements')}
          >
            🏆 ({unlockedCount}/{achievements.length})
          </button>
          <button
            className="panel-close-btn"
            onClick={togglePanel}
          >
            ✕
          </button>
        </div>

        {/* Panel Content */}
        <div className="panel-content">
          {activeTab === 'upgrades' && (
            <div className="upgrade-list">
              {upgrades.map((upgrade, index) => {
                const cost = getUpgradeCost(upgrade);
                const canAfford = gameState.score >= cost;

                return (
                  <div
                    key={upgrade.id}
                    className={`upgrade-item ${!canAfford ? 'disabled' : ''}`}
                    onClick={() => canAfford && handleBuyUpgrade(index)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        handleBuyUpgrade(index);
                      }
                    }}
                  >
                    <div
                      className="upgrade-icon"
                      style={{ background: upgrade.color }}
                    >
                      {upgrade.icon}
                    </div>
                    <div className="upgrade-info">
                      <div className="upgrade-name">{upgrade.name}</div>
                      <div className="upgrade-cost">💰 {cost.toLocaleString()}</div>
                      <div className="upgrade-income">+{upgrade.income}/сек</div>
                    </div>
                    <div className="upgrade-count">ур. {upgrade.count}</div>
                  </div>
                );
              })}
            </div>
          )}

          {activeTab === 'achievements' && (
            <div className="achievements-list">
              {achievements.map((achievement) => (
                <div
                  key={achievement.id}
                  className={`achievement-item ${achievement.unlocked ? 'unlocked' : 'locked'}`}
                >
                  <div className="achievement-icon">
                    {achievement.unlocked ? achievement.icon : '🔒'}
                  </div>
                  <div className="achievement-info">
                    <div className="achievement-name">{achievement.name}</div>
                    <div className="achievement-description">{achievement.description}</div>
                  </div>
                  {achievement.unlocked && (
                    <div className="achievement-check">✅</div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default App;
