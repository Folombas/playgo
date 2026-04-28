package levels

import "snake/internal/types"

type Level struct {
    Width     int
    Height    int
    Obstacles []types.Vec // coordinates of wall cells
}

// LoadDefault returns a simple level with no obstacles.
func LoadDefault() *Level {
    return &Level{
        Width:  types.GridW,
        Height: types.GridH,
        // No obstacles for now; can be extended later.
    }
}
