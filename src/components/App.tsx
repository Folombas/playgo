import React, { useEffect, useRef, useState, useCallback } from 'react';
import Phaser from 'phaser';
import { GameScene } from '../game/GameScene';
import { GameState, Upgrade, Quest, initialGameState, upgrades as initialUpgrades, getUpgradeCost } from '../game/types';

const App: React.FC = () => {
  const gameRef = useRef<Phaser.Game | null>(null);
  const sceneRef = useRef<GameScene | null>(null);
  const containerRef = useRef<HTMLDivElement | null>(null);
  
  const [gameState, setGameState] = useState<GameState>(initialGameState);
  const [upgrades, setUpgrades] = useState<Upgrade[]>(initialUpgrades);
  const [quests, setQuests] = useState<Quest[]>([]);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [audioEnabled, setAudioEnabled] = useState(true);
  const [activeTab, setActiveTab] = useState<'upgrades' | 'quests' | 'stats'>('upgrades');

  const initGame = useCallback(() => {
    if (!containerRef.current || gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 400,
      height: 700,
      parent: containerRef.current,
      backgroundColor: '#0a0a1a',
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

      scene.onQuestCompleted = (quest: Quest) => {
        setQuests(prev => prev.map(q => q.id === quest.id ? quest : q));
      };

      setAudioEnabled(scene.isAudioEnabled());
      setQuests(scene.getQuests());
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

  const totalProgress = quests.reduce((acc, q) => acc + (q.completed ? 1 : 0), 0);

  return (
    <div className="game-container" ref={containerRef}>
      {/* Audio Toggle */}
      <button onClick={toggleAudio} className="game-btn audio-btn" title={audioEnabled ? 'Выключить звук' : 'Включить звук'}>
        {audioEnabled ? '🔊' : '🔇'}
      </button>

      {/* Menu Button */}
      <button onClick={togglePanel} className="game-btn menu-btn">
        {isPanelOpen ? '✕' : '☰'}
      </button>

      {/* Side Panel */}
      <div className={`side-panel ${isPanelOpen ? 'open' : ''}`}>
        {/* Panel Header */}
        <div className="panel-header">
          <div className="panel-tabs">
            <button className={`panel-tab ${activeTab === 'upgrades' ? 'active' : ''}`} onClick={() => setActiveTab('upgrades')}>
              🚀
            </button>
            <button className={`panel-tab ${activeTab === 'quests' ? 'active' : ''}`} onClick={() => setActiveTab('quests')}>
              📋
            </button>
            <button className={`panel-tab ${activeTab === 'stats' ? 'active' : ''}`} onClick={() => setActiveTab('stats')}>
              📊
            </button>
          </div>
          <button className="panel-close" onClick={togglePanel}>✕</button>
        </div>

        {/* Panel Content */}
        <div className="panel-content">
          {/* UPGRADES TAB */}
          {activeTab === 'upgrades' && (
            <div className="upgrades-container">
              <h2 className="section-title">⚡ УЛУЧШЕНИЯ</h2>
              {upgrades.map((upgrade, index) => {
                const cost = getUpgradeCost(upgrade);
                const canAfford = gameState.score >= cost;
                return (
                  <button
                    key={upgrade.id}
                    className={`game-card upgrade-card ${!canAfford ? 'disabled' : ''}`}
                    onClick={() => canAfford && handleBuyUpgrade(index)}
                    disabled={!canAfford}
                  >
                    <div className="card-icon" style={{ background: upgrade.color }}>{upgrade.icon}</div>
                    <div className="card-info">
                      <div className="card-name">{upgrade.name}</div>
                      <div className="card-cost">💰 {cost.toLocaleString()}</div>
                      <div className="card-bonus">+{upgrade.income}/сек</div>
                    </div>
                    <div className="card-level">ур.{upgrade.count}</div>
                  </button>
                );
              })}
            </div>
          )}

          {/* QUESTS TAB */}
          {activeTab === 'quests' && (
            <div className="quests-container">
              <h2 className="section-title">📋 ЗАДАНИЯ</h2>
              <div className="progress-summary">
                Выполнено: {totalProgress}/{quests.length}
              </div>
              {quests.map(quest => {
                const progress = Math.min((quest.progress / quest.target) * 100, 100);
                return (
                  <div key={quest.id} className={`game-card quest-card ${quest.completed ? 'completed' : ''}`}>
                    <div className="quest-info">
                      <div className="quest-name">{quest.name}</div>
                      <div className="quest-progress">{quest.progress} / {quest.target}</div>
                    </div>
                    <div className="quest-bar">
                      <div className="quest-fill" style={{ width: `${progress}%` }}></div>
                    </div>
                    <div className="quest-reward">💰 {quest.reward}</div>
                  </div>
                );
              })}
            </div>
          )}

          {/* STATS TAB */}
          {activeTab === 'stats' && (
            <div className="stats-container">
              <h2 className="section-title">📊 СТАТИСТИКА</h2>
              <div className="stat-grid">
                <div className="stat-card">
                  <div className="stat-icon">👆</div>
                  <div className="stat-value">{gameState.totalTaps.toLocaleString()}</div>
                  <div className="stat-label">Всего тапов</div>
                </div>
                <div className="stat-card">
                  <div className="stat-icon">⚡</div>
                  <div className="stat-value">{gameState.criticalHits.toLocaleString()}</div>
                  <div className="stat-label">Критов</div>
                </div>
                <div className="stat-card">
                  <div className="stat-icon">💰</div>
                  <div className="stat-value">{gameState.score.toLocaleString()}</div>
                  <div className="stat-label">Монеты</div>
                </div>
                <div className="stat-card">
                  <div className="stat-icon">🔥</div>
                  <div className="stat-value">{gameState.autoTapPerSec.toFixed(1)}</div>
                  <div className="stat-label">Доход/сек</div>
                </div>
                <div className="stat-card">
                  <div className="stat-icon">⭐</div>
                  <div className="stat-value">{gameState.level}</div>
                  <div className="stat-label">Уровень</div>
                </div>
                <div className="stat-card">
                  <div className="stat-icon">💪</div>
                  <div className="stat-value">{gameState.tapValue}</div>
                  <div className="stat-label">Сила тапа</div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default App;
