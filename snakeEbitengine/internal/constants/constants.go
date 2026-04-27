package constants

// Размеры игрового поля и тайла
const (
    TileSize      = 32
    GridW         = 20 // ширина в тайлах, можно вынести в настройки
    GridH         = 15 // высота в тайлах
    InitialLength = 5
    MaxHealth     = 3
)

// Количества и таймеры
const (
    GhostFrames      = 11
    GiftCount        = 6
    CoinFrames       = 4
    IceSpawnChance   = 0.3
    RoachFramesX     = 32
    RoachFramesY     = 32
    RoachCols        = 4
    RoachRows        = 5
    VikingCols       = 5
    VikingRows       = 2
)

// Скорости при разных сложностях
const (
    SpeedEasy   = 6
    SpeedNormal = 9
    SpeedHard   = 12
)
