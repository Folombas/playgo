import React, { useEffect, useRef, useState, useCallback } from 'react';
import Phaser from 'phaser';
import { GameScene } from '../game/GameScene';
import { GameState, Upgrade, initialGameState, upgrades as initialUpgrades, getUpgradeCost } from '../game/types';

const App: React.FC = () => {
  const gameRef = useRef<Phaser.Game | null>(null);
  const sceneRef = useRef<GameScene | null>(null);
  
  const [gameState, setGameState] = useState<GameState>(initialGameState);
  const [upgrades, setUpgrades] = useState<Upgrade[]>(initialUpgrades);
  const [isPanelOpen, setIsPanelOpen] = useState(false);

  const initGame = useCallback(() => {
    if (gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 400,
      height: 700,
      parent: 'game-canvas',
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
      },
    };

    gameRef.current = new Phaser.Game(config);

    gameRef.current.events.once('ready', () => {
      const scene = gameRef.current?.scene.getScene('GameScene') as GameScene;
      if (scene) {
        sceneRef.current = scene;
        scene.setGameState(gameState);
        scene.setUpgrades(upgrades);

        // Setup callbacks
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
      }
    });
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

  // Sync upgrades with scene
  useEffect(() => {
    if (sceneRef.current) {
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

  return (
    <div className="game-container">
      <div id="game-canvas" className="game-canvas"></div>
      
      {/* Upgrade Button */}
      <button
        onClick={togglePanel}
        style={{
          position: 'absolute',
          bottom: 20,
          right: 20,
          background: '#00ADD8',
          color: 'white',
          border: 'none',
          padding: '15px 25px',
          borderRadius: '10px',
          cursor: 'pointer',
          fontSize: '16px',
          fontWeight: 'bold',
          boxShadow: '0 4px 15px rgba(0, 173, 216, 0.4)',
          transition: 'all 0.2s',
        }}
      >
        {isPanelOpen ? '✕ Close' : '⬆ Upgrades'}
      </button>

      {/* Upgrade Panel */}
      <div className={`upgrade-panel ${isPanelOpen ? 'open' : ''}`}>
        <h2 style={{ color: '#00ADD8', marginBottom: '20px', textAlign: 'center' }}>
          UPGRADES
        </h2>
        
        {upgrades.map((upgrade, index) => {
          const cost = getUpgradeCost(upgrade);
          const canAfford = gameState.score >= cost;

          return (
            <div
              key={upgrade.id}
              className="upgrade-item"
              onClick={() => handleBuyUpgrade(index)}
              style={{ opacity: canAfford ? 1 : 0.5 }}
            >
              <div
                className="upgrade-icon"
                style={{ background: upgrade.color }}
              >
                {upgrade.icon}
              </div>
              <div className="upgrade-info">
                <div className="upgrade-name">{upgrade.name}</div>
                <div className="upgrade-cost">💰 {cost} | +{upgrade.income}/sec</div>
              </div>
              <div className="upgrade-count">x{upgrade.count}</div>
            </div>
          );
        })}

        <button
          className="close-panel-btn"
          onClick={togglePanel}
        >
          CLOSE
        </button>
      </div>

      {/* Stats Overlay */}
      <div
        style={{
          position: 'absolute',
          top: 20,
          left: 20,
          color: '#888',
          fontSize: '12px',
        }}
      >
        <div>Level: {gameState.level}</div>
        <div>Tap Power: {gameState.tapValue}</div>
        <div>Auto: {gameState.autoTapPerSec.toFixed(1)}/sec</div>
      </div>
    </div>
  );
};

export default App;
