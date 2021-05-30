package douji

import "github.com/leancloud/go-sdk/leancloud"

type Card struct {
	rank int
	suit string
}

type Set struct {
	leancloud.Object
	id         string `json:"id"`
	gameNumber int    `json:"gameNumber"` // number of games
	// step        int
	// end         int
	printStatus bool
}

type gameStatus int

const (
	ready gameStatus = iota
	inProcess
	bombing
	over
)

type Game struct {
	id          int
	players     []*Player
	base        int
	hiddenCount int
	pot         int // chips from all players.
	step        int
	end         int
	maxRound    int
	status      gameStatus // maybe don't need this?
	prevWinner  *Player
}

type Deck struct {
	cards []Card
}

type CardDealer interface {
	DealOne() Card
}

type Hand struct {
	privateCards []Card
	publicCards  []Card
	isFourKind   bool
	hasWildCard  bool
	hasJokerR    bool
	hasJokerB    bool
}

type Player struct {
	id     string
	Name   string
	points int
	Hand
}

// These are the operations that need to wait for players input.
type MiddleGame interface {
	// whether stay in the game or quit when another player calls a certain amount.
	InOrOut(player *Player, callingChip int) bool

	// CallOnce lets a calling player either call a certain amount or choose to quit the game by calling 0.
	// Each round calling points are [0, step, 2step...end]. In the final round, it can choose from [0, step, 2step...end, end*2]
	// If last game is bombed, then the calling points is doubled i.e. [0, 2step, 4step, 6step ... 2*end] and in the final round the limit is 4*end.
	// 0 is always an option since it indicates quitting the game.
	CallOnce(player *Player, step, end int, lastCall bool) int
}
