package game

type TileType int
const (
    Red TileType = iota
    Blue
    Green
    Yellow
    Purple
)

type Tile struct {
    Type  TileType
    State string
}

type Board [8][8]Tile

type Game struct {
    Board     Board
    Score     int
    Selected  *[2]int
    Combo     int
    MovesLeft int
    GameState string
    Level     int
    HintTimer int
    HintTarget [2]int
}

func NewGame() *Game {
    return &Game{
        Score:     0,
        MovesLeft: 20,
        GameState:  menu,
        Level:     1,
    }
}

func (g *Game) NewLevel() {
    g.Board = Board{}
    g.GenerateBoard()
    for hasMatches(g.Board) {
        g.removeMatches()
        g.dropTiles()
        g.fillEmpty()
    }
    g.HintTimer = 0
}

func (g *Game) GenerateBoard() {
    // Placeholder: implement random board generation
}

func (g *Game) fillEmpty()   {}
func (g *Game) dropTiles()   {}
func hasMatches(b Board) bool { return false }
func (g *Game) removeMatches() {}
