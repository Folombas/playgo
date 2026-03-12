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
  const [isLoaded, setIsLoaded] = useState(false);

  // Detect mobile device
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

    gameRef.current.events.once('ready', () => {
      const scene = gameRef.current?.scene.getScene('GameScene') as GameScene;
      if (scene) {
        sceneRef.current = scene;
        scene.setGameState(gameState);
        scene.setUpgrades(upgrades);
        setIsLoaded(true);

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

        scene.onToggleAudio = (enabled: boolean) => {
          setAudioEnabled(enabled);
        };

        setAudioEnabled(scene.isAudioEnabled());
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
      {/* Game canvas will be injected here by Phaser */}
      
      {/* Audio Toggle Button */}
      <button
        onClick={toggleAudio}
        title={audioEnabled ? 'Mute' : 'Unmute'}
        style={{
          position: 'absolute',
          top: 20,
          right: 20,
          background: audioEnabled ? 'rgba(0, 173, 216, 0.5)' : 'rgba(100, 100, 100, 0.5)',
          border: 'none',
          borderRadius: '50%',
          width: 44,
          height: 44,
          cursor: 'pointer',
          fontSize: 20,
          zIndex: 100,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          transition: 'all 0.2s',
          color: '#fff',
        }}
      >
        {audioEnabled ? '🔊' : '🔇'}
      </button>

      {/* Upgrade Button */}
      <button
        onClick={togglePanel}
        style={{
          position: 'absolute',
          bottom: isMobile ? 15 : 20,
          right: isMobile ? 15 : 20,
          background: 'linear-gradient(135deg, #00ADD8 0%, #0097B5 100%)',
          color: 'white',
          border: 'none',
          padding: '16px 28px',
          borderRadius: '12px',
          cursor: 'pointer',
          fontSize: '16px',
          fontWeight: 'bold',
          boxShadow: '0 4px 20px rgba(0, 173, 216, 0.5)',
          transition: 'all 0.2s',
          zIndex: 100,
        }}
      >
        {isPanelOpen ? '✕ Close' : '⬆ Upgrades'}
      </button>

      {/* Upgrade Panel */}
      <div className={`upgrade-panel ${isPanelOpen ? 'open' : ''}`}>
        <h2>UPGRADES</h2>
        
        {upgrades.map((upgrade, index) => {
          const cost = getUpgradeCost(upgrade);
          const canAfford = gameState.score >= cost;

          return (
            <div
              key={upgrade.id}
              className="upgrade-item"
              onClick={() => handleBuyUpgrade(index)}
              style={{ opacity: canAfford ? 1 : 0.5 }}
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
      <div className="stats-overlay">
        <div>Level: {gameState.level}</div>
        <div>Tap Power: {gameState.tapValue}</div>
        <div>Auto: {gameState.autoTapPerSec.toFixed(1)}/sec</div>
      </div>
    </div>
  );
};

export default App;
