package douji

import (
	"fmt"
)

var (
	wildCard = Card{rank: 2, suit: "♥"}
	jokerR   = Card{rank: 19, suit: "joker Red"}
	jokerB   = Card{rank: 17, suit: "joker Black"}
)

func isCard(card Card, targetCard Card) bool {
	return card.rank == targetCard.rank && card.suit == targetCard.suit
}

func NewPlayer(name, password string, points int, db Db) *Player {
	id, err := db.CreatePlayer(name, password, points)
	if err != nil {
		panic(err)
	}
	return &Player{Name: name, points: points, id: id}
}

// This shall be in the test file but because main package can't access functions in test files; it's moved here.
func NewTestPlayer(name, id string, points int) *Player {
	return &Player{Name: name, id: id, points: points}
}

// func GetPlayerById(db Db, id string) *Player {
// 	pdto := db.LoadPlayerStats(id)
// 	return NewPlayer(pdto.Name)
// }

func (p Player) String() string {
	return fmt.Sprintf("%s(%d-%d)-**: - %v", p.Name, p.points, p.PublicScore(), p.publicCards)
}

// ReceivePublicCard receives a new public card for the player.
func (p *Player) ReceivePublicCard(c Card) {
	p.publicCards = append(p.publicCards, c)
}

// returns the number of N kind according to a rank frequency map.
func hasNKind(freqMap map[int]int, n int) int {
	counter := 0
	for _, v := range freqMap {
		if v == n {
			counter++
		}
	}
	return counter
}

// ReceivePrivateCard receives a hidden card for the player.
func (p *Player) ReceivePrivateCard(c Card) {
	p.privateCards = append(p.privateCards, c)
}

// FaceScore returns the last public card score for a player.
func (p *Player) FaceScore() int {
	pcs := p.publicCards
	if len(pcs) == 0 {
		return 0
	}
	return pcs[len(pcs)-1].rank
}

// PublicScore returns the total score of a player's public cards
func (p *Player) PublicScore() int {
	return p.calculateScore(p.publicCards)
}

// Checks whether there is any additional points from jokers including wild card. Two jokers gets additional 30 points.
// It returns extra points and whether wild card has been used to get the extra point, if any.
func (p *Player) checkJokerExtra() (int, bool) {
	switch {
	case p.hasJokerR && p.hasJokerB && p.hasWildCard:
		return 47, true // it's like having two red Joker and one black Joker i.e. treating the wild card as red Joker (from rank 2 to rank 19).
	case p.hasJokerR && p.hasJokerB:
		return 30, false
	case p.hasJokerR && p.hasWildCard:
		return 45, true
	case p.hasJokerB && p.hasWildCard:
		return 47, true
	default:
		return 0, false
	}
}

// getRankSumAndFreq gets the rank sum and rank frequency; it also checks whether it contains jokers and wild card, if so setting the corresponding flag.
func (p *Player) getRankSumAndFreq(cards []Card) (map[int]int, int) {
	freqMap := make(map[int]int, len(cards))
	sum := 0
	for _, c := range cards {
		if !p.hasWildCard {
			p.hasWildCard = isCard(c, wildCard)
		}
		if !p.hasJokerR {
			p.hasJokerR = isCard(c, jokerR)
		}
		if !p.hasJokerB {
			p.hasJokerB = isCard(c, jokerB)
		}
		freqMap[c.rank]++
		sum += c.rank
	}
	return freqMap, sum
}

func (p *Player) calculateScore(cards []Card) int {
	freqMap, score := p.getRankSumAndFreq(cards)
	if p.hasWildCard {
		for k, freq := range freqMap {
			if freq == 4 && k != 2 {
				p.isFourKind = true
				return 300 // five a kind rules everything else!!!!
			}
			if freq == 3 && k != 2 {
				p.isFourKind = true
				freqMap[k]++
				score += k - 2
				p.hasWildCard = false // wild card has been used to get a four a kind!
				freqMap[2]--          // reduce rank 2 frequency since wild card has been used.
				break
			}
		}
	}
	extra, used := p.checkJokerExtra()
	score += extra
	if used {
		p.hasWildCard = false // wild card has been used to get double jokers!
	}

	// check for four a kind.
	if hasNKind(freqMap, 4) > 0 {
		p.isFourKind = true
		score += 60  // four a kind gets extra 60 points.
		return score // when there is a four a kind, no need to check for three a kind.
	}
	if p.hasWildCard {
		freqTwoRank := 0
		for k, freq := range freqMap {
			if freq == 2 && k != 2 && k > freqTwoRank { // don't treat it as three a kind if it's just a pair of 2s.
				freqTwoRank = k // take the larger freq two rank
			}
		}
		if freqTwoRank > 0 {
			freqMap[freqTwoRank]++
			score += freqTwoRank - 2
		}
	}

	// check for three a kind. Possible to have more than one 3 a kind in a hand of 6 cards.
	score += 30 * hasNKind(freqMap, 3)
	return score
}

// FinalScore returns the final score of both public and private cards for a player in a game.
// Final score is only relevant in the final round where there are still more than one player.
func (p *Player) FinalScore() int {
	all := append(p.privateCards, p.publicCards...)
	return p.calculateScore(all)
}

func (p *Player) ClearHand() {
	p.Hand = Hand{}
}
