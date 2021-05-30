package douji

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

func (c Card) String() string {
	return fmt.Sprintf("%s%d", c.suit, c.rank)
}

// AddPlayer adds a new player to an in-process game.
func (g *Game) AddPlayer(p Player) {
	g.players = append(g.players, &p)
}

func NewGame(id int, players []*Player, base, hiddenCount, pot, step, end int, prevWinner *Player) *Game {
	mr := 4 // one hidden card game has 4 rounds after start.
	if hiddenCount > 1 {
		mr = 5 // with two hidden cards, there are an extra round.
	}
	return &Game{
		id:          id,
		players:     players,
		base:        base,
		hiddenCount: hiddenCount,
		pot:         pot,
		status:      ready,
		step:        step,
		end:         end,
		maxRound:    mr,
		prevWinner:  prevWinner,
	}
}

// start starts a game by assigning each player certain hidden cards and possibly one public card.
func (g *Game) start(cardDealer CardDealer) bool {
	if len(g.players) < 2 {
		return false // can't start until there are at least two players.
	}
	bombedPot := g.pot > 0
	for _, p := range g.players {
		if !bombedPot {
			p.points -= g.base // regardless of how many hidden cards, starting a game only costs one base point for each player unless last game was bombed.
		}
		c := cardDealer.DealOne()
		p.ReceivePrivateCard(c)
	}
	for _, p := range g.players {
		c := cardDealer.DealOne()
		if g.hiddenCount == 2 {
			p.ReceivePrivateCard(c) // two hidden cards, no initial public card.
		} else {
			p.ReceivePublicCard(c) // when there is only one hidden card, each gets a public card.
		}
	}
	if !bombedPot {
		g.pot += g.base * len(g.players) //
	}
	g.status = inProcess
	return true
}

func (g *Game) printCurrentStatus(i int) {
	fmt.Printf("Total pot:%d, Round: %d\n", g.pot, i)
	for _, p := range g.players {
		fmt.Println(p)
	}
	fmt.Println()
}

// getAskingPlayers gets asking players in the game with a given calling player index.
func (g *Game) getAskingPlayers(index int) []*Player {
	if index == len(g.players)-1 {
		return g.players[:index] // dealing order starts from the player next to the calling player.
	}
	return append(g.players[index+1:], g.players[:index]...) // deals in anti-clock wise order.
}

func dealARound(dealer CardDealer, inPlayers []*Player) {
	for _, p := range inPlayers {
		c := dealer.DealOne()
		p.ReceivePublicCard(c)
	}
}

// find the player with largest face score to be the calling player.
func (g *Game) getCallingPlayerByFaceScore() *Player {
	var cp *Player
	fs := 1
	for _, player := range g.players {
		pfs := player.FaceScore()
		if pfs > fs {
			cp = player
			fs = pfs
		}
	}
	return cp
}

func getPlayerIndex(p *Player, players []*Player) int {
	for i, player := range players {
		if player.id == p.id {
			return i
		}
	}
	panic(fmt.Errorf("cannot find an index for player name:%s, id:%s", p.Name, p.id))
}

func (g *Game) getTwoHiddenCallingPlayer(isFirstRound bool) *Player {
	// With two hidden cards, if there is a previous winner (i.e. 2nd game onwards), previous winner calls the first round.
	if isFirstRound && g.prevWinner != nil {
		p := g.prevWinner
		g.prevWinner = nil
		return p
	}
	// in the 1st round when no previous winner (i.e. the 1st game), let first player call.
	if isFirstRound {
		return g.players[0]
	}

	// in all other situations, calling player is chosen according to face value.
	return g.getCallingPlayerByFaceScore()
}

func (g *Game) getCallingPlayer(isFirstRound bool) *Player {
	if g.hiddenCount == 1 {
		return g.getCallingPlayerByFaceScore()
	}
	return g.getTwoHiddenCallingPlayer(isFirstRound)
}

// run a game and return its winner player with the finished game pot. Unless it's bombed pot, the ending pot is 0.
func (g *Game) run(print bool, md MiddleGame, cardDealer CardDealer) (*Player, int) {
	if !g.start(cardDealer) {
		panic("failed to start the game.")
	}
	if print {
		fmt.Println("Game started!")
		g.printCurrentStatus(0)
	}

	for i := 1; i <= g.maxRound; i++ {
		cp := g.getCallingPlayer(i == 1)
		callingPoint := md.CallOnce(cp, g.step, g.end, i == g.maxRound)
		for callingPoint == 0 {
			idx := getPlayerIndex(cp, g.players)
			g.players = g.getAskingPlayers(idx)
			if len(g.players) == 1 {
				break // one player left, game over, break from the inner for loop.
			}
			cp = g.getCallingPlayer(i == 1)
			callingPoint = md.CallOnce(cp, g.step, g.end, i == g.maxRound)
		}
		if len(g.players) == 1 {
			break // one player left, game over! break from the outer loop.
		}
		cp.points -= callingPoint  // update calling player's chips.
		g.pot += callingPoint      // update pot
		inPlayers := []*Player{cp} // calling player always remains in the game.
		callingIndex := getPlayerIndex(cp, g.players)
		askingPlayers := g.getAskingPlayers(callingIndex)
		for _, player := range askingPlayers {
			if md.InOrOut(player, callingPoint) {
				player.points -= callingPoint
				g.pot += callingPoint
				inPlayers = append(inPlayers, player)
			}
		}
		g.players = inPlayers
		if len(g.players) == 1 {
			break // game over as only calling player is left.
		}
		if i < g.maxRound { // deal a round before the last round.
			dealARound(cardDealer, inPlayers)
		}
		if print {
			g.printCurrentStatus(i)
		}
	}

	// more than 1 player in the final round,
	if len(g.players) > 1 {
		// check for four a kind!
		if fkp, ok := checkFourKind(g.players); ok {
			fkp.points += g.pot
			g.status = over
			return fkp, 0
		}

		// no single four kind, check for final score to determine a winner.
		sort.Slice(g.players, func(i, j int) bool {
			return g.players[i].FinalScore() > g.players[j].FinalScore()
		})
		// check for bombing pot.
		if g.players[0].FinalScore() == g.players[1].FinalScore() {
			g.status = bombing
			return nil, g.pot // it's a tie so no winner yet.
		}
	}

	g.players[0].points += g.pot // update winner's chips.
	g.pot = 0
	g.status = over
	return g.players[0], 0
}

// check whether there is any player having four a kind, if so return the player and true.
func checkFourKind(players []*Player) (*Player, bool) {
	var fourKindPlayers []*Player
	for _, p := range players {
		if p.isFourKind {
			fourKindPlayers = append(fourKindPlayers, p)
		}
	}
	if len(fourKindPlayers) == 1 {
		return fourKindPlayers[0], true // if only one player with four a kind, it's the winner otherwise always compare the final score.
	}
	return nil, false
}

func convertToPlayerDTO(players []*Player) []PlayerDTO {
	pdtos := make([]PlayerDTO, len(players))
	for i := 0; i < len(players); i++ {
		pdtos[i] = PlayerDTO{players[i].id, players[i].Name, players[i].points}
	}
	return pdtos
}

func (s Set) Run(players []*Player, md MiddleGame, db Db, base int, hiddenCount int, pot int) {
	step := 1 // s.step
	end := 5  // s.end
	var prevWinner *Player
	var wg sync.WaitGroup
	for i := 0; i < s.gameNumber; i++ {
		game := NewGame(i, players, base, hiddenCount, pot, step, end, prevWinner)
		prevWinner, pot = game.run(s.printStatus, md, NewDeck())
		pdtos := convertToPlayerDTO(players)
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if err := db.SaveGameStats(s.id, i+1, pdtos); err != nil {
				panic(err)
			}
		}(i)
		// this is just for debugging.
		if s.printStatus {
			if prevWinner != nil {
				fmt.Printf("Game %d winner is:%s\n", i, prevWinner.Name)
			}
			for _, player := range players {
				fmt.Printf("%s-%d-%v-%v\n", player.Name, player.FinalScore(), player.privateCards, player.publicCards)
			}
		}
		for _, p := range players {
			p.ClearHand()
		}
		if pot > 0 { // bombed pot!
			s.gameNumber++ // add an extra game when there is a bombed pot.
			step *= 2      // double the step with every bobmed pot.
			end *= 2       // double the end with every bombed pot.
		} else {
			step = 1 // s.step // revert back to the original data.
			end = 5  // s.end
		}
	}
	if s.printStatus {
		fmt.Println("\nSet is finished!")
		for _, p := range players {
			fmt.Printf("%s(%d)-%d\n", p.Name, p.points, p.FinalScore())
		}
	}
	wg.Wait() // make sure the last game result is saved before exit.
}

// creates unshuffled cards.
func createCards() []Card {
	var (
		// for quickness, use int to represent ranks.
		ranks = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 15} // J-11, Q-12, K-13, A-15
		suits = []string{"♦", "♣", "♥", "♠"}
	)
	cards := make([]Card, 0, 55)
	for _, r := range ranks {
		for _, s := range suits {
			cards = append(cards, Card{r, s})
		}
	}
	return append(cards, jokerB, jokerR, Card{21, "special"})
}

// NewDeck creates a new randomly shuffle deck of 55 cards.
func NewDeck() *Deck {
	cards := createCards()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
	return &Deck{cards}
}

// DealOne deals one card from the the deck.
func (d *Deck) DealOne() Card {
	if len(d.cards) == 0 {
		panic("empty deck, can't deal anymore")
	}
	c := d.cards[0]
	d.cards = d.cards[1:]
	return c
}

func NewSet(gameNumber int, printStatus bool) Set {
	return Set{gameNumber: gameNumber, printStatus: printStatus}
}
