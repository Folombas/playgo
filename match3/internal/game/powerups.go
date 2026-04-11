package game

// PowerUpType определяет тип бонуса
type PowerUpType int

const (
	PowerUpNone PowerUpType = iota
	PowerUpHammer      // Молоток - уничтожает один выбранный камень
	PowerUpShuffle     // Перемешивает доску
	PowerUpHint        // Показывает 3 возможных хода
)

// PowerUp представляет бонус
type PowerUp struct {
	Type     PowerUpType
	Quantity int // Количество
}

// PowerUpSystem управляет бонусами игрока
type PowerUpSystem struct {
	powerUps map[PowerUpType]*PowerUp
}

// NewPowerUpSystem создаёт систему бонусов
func NewPowerUpSystem() *PowerUpSystem {
	ps := &PowerUpSystem{
		powerUps: make(map[PowerUpType]*PowerUp),
	}
	
	// Инициализация пустых бонусов
	ps.powerUps[PowerUpHammer] = &PowerUp{Type: PowerUpHammer, Quantity: 0}
	ps.powerUps[PowerUpShuffle] = &PowerUp{Type: PowerUpShuffle, Quantity: 0}
	ps.powerUps[PowerUpHint] = &PowerUp{Type: PowerUpHint, Quantity: 0}
	
	return ps
}

// AddPowerUp добавляет бонус
func (ps *PowerUpSystem) AddPowerUp(typ PowerUpType, amount int) {
	if pu, exists := ps.powerUps[typ]; exists {
		pu.Quantity += amount
	}
}

// UsePowerUp использует бонус
func (ps *PowerUpSystem) UsePowerUp(typ PowerUpType) bool {
	if pu, exists := ps.powerUps[typ]; exists && pu.Quantity > 0 {
		pu.Quantity--
		return true
	}
	return false
}

// GetQuantity возвращает количество бонуса
func (ps *PowerUpSystem) GetQuantity(typ PowerUpType) int {
	if pu, exists := ps.powerUps[typ]; exists {
		return pu.Quantity
	}
	return 0
}

// HasPowerUp проверяет, есть ли бонус
func (ps *PowerUpSystem) HasPowerUp(typ PowerUpType) bool {
	return ps.GetQuantity(typ) > 0
}

// AwardPowerUpsForCombo награждает бонусами за комбо
func (ps *PowerUpSystem) AwardPowerUpsForCombo(combo int) []PowerUpType {
	awarded := make([]PowerUpType, 0)
	
	if combo >= 5 && combo < 10 {
		// Комбо x5 - даёт молоток
		ps.AddPowerUp(PowerUpHammer, 1)
		awarded = append(awarded, PowerUpHammer)
	} else if combo >= 10 && combo < 15 {
		// Комбо x10 - даёт перемешивание
		ps.AddPowerUp(PowerUpShuffle, 1)
		awarded = append(awarded, PowerUpShuffle)
	} else if combo >= 15 {
		// Комбо x15+ - даёт подсказку
		ps.AddPowerUp(PowerUpHint, 1)
		awarded = append(awarded, PowerUpHint)
	}
	
	return awarded
}
