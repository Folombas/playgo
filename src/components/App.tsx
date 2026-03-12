import React, { useEffect, useRef, useState, useCallback } from 'react';
import Phaser from 'phaser';
import { GameScene } from '../game/GameScene';
import { GameState, Upgrade, initialGameState, upgrades as initialUpgrades, getUpgradeCost } from '../game/types';

const App: React.FC = () => {
  const gameRef = useRef<Phaser.Game | null>(null);
  const sceneRef = useRef<GameScene | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  
  const [gameState, setGameState] = useState<GameState>(initialGameState);
  const [upgrades, setUpgrades] = useState<Upgrade[]>(initialUpgrades);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [audioEnabled, setAudioEnabled] = useState(true);
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkMobile = () => {
      const mobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)
        || window.innerWidth <= 768;
      setIsMobile(mobile);
    };
    
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

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

      setAudioEnabled(scene.isAudioEnabled());
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

      {/* Stats Overlay */}
      <div className="stats-overlay">
        <div>📊 Уровень: {gameState.level}</div>
        <div>💪 Сила тапа: {gameState.tapValue}</div>
        <div>🤖 Авто: {gameState.autoTapPerSec.toFixed(1)}/сек</div>
      </div>

      {/* Upgrade Button */}
      <button
        onClick={togglePanel}
        className="upgrade-toggle-btn"
      >
        {isPanelOpen ? '✕ Закрыть' : '⬆ Улучшения'}
      </button>

      {/* Upgrade Panel */}
      <div className={`upgrade-panel ${isPanelOpen ? 'open' : ''}`}>
        <div className="upgrade-panel-header">
          <h2>🚀 УЛУЧШЕНИЯ</h2>
          <button
            className="close-panel-btn"
            onClick={togglePanel}
          >
            ✕
          </button>
        </div>
        
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
                  if (e.key === 'Enter' || e.key === ' ' && canAfford) {
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
                  <div className="upgrade-cost">💰 {cost.toLocaleString()} | +{upgrade.income}/сек</div>
                </div>
                <div className="upgrade-count">ур. {upgrade.count}</div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default App;
