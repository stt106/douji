package douji

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestNewDeck(t *testing.T) {
	d := NewDeck()
	if len(d.cards) != 55 {
		t.Errorf("expected to get 55 cards in a new deck but got:%d", len(d.cards))
	}
}

func TestAddNewPlayer(t *testing.T) {
	g := NewGame(0, []*Player{}, 1, 1, 0, 1, 5, nil)
	p := NewTestPlayer("p1", "100", 1)

	g.AddPlayer(*p)
	if len(g.players) != 1 {
		t.Errorf("expected to have 1 player in the game now but got: %d", len(g.players))
	}
	if !reflect.DeepEqual(g.players[0], p) {
		t.Errorf("expected added player to the game but didn't.")
	}
	// if diff := cmp.Diff(g.players[0], p); diff != "" {
	// 	t.Error(diff)
	// }
}

func NewTestPlayer(name, id string, points int) *Player {
	return &Player{Name: name, id: id, points: points}
}

func TestStartGame(t *testing.T) {
	for _, tc := range []struct {
		hiddenCount int
		players     []*Player
		pcc         int
	}{
		{1, []*Player{
			NewTestPlayer("Liu", "1", 100),
			NewTestPlayer("Wang", "2", 100),
			NewTestPlayer("Gu", "3", 100),
		}, 1},
		{2, []*Player{
			NewTestPlayer("Liu", "1", 100),
			NewTestPlayer("Wang", "2", 100),
			NewTestPlayer("Gu", "3", 100),
		}, 0},
	} {
		base := 1
		g := NewGame(0, tc.players, base, tc.hiddenCount, 0, 1, 5, nil)
		g.start(NewDeck())
		for _, p := range tc.players {
			if len(p.publicCards) != tc.pcc {
				t.Errorf("expected to get %d public card after game start for player:%s but got:%d", tc.pcc, p.Name, len(p.publicCards))
			}
			if len(p.privateCards) != tc.hiddenCount {
				t.Errorf("expected to get %d hidden card after game start for player:%s but got:%d", tc.hiddenCount, p.Name, len(p.privateCards))
			}
			if p.points != 100-base {
				t.Errorf("expected chips to be %d for player:%s but got:%d", 100-base, p.Name, p.points)
			}
		}
	}
}

func getFourTestingPlayers() []*Player {
	return []*Player{
		{Name: "Liu", points: 100, id: "0"},
		{Name: "Wang", points: 100, id: "1"},
		{Name: "Gu", points: 100, id: "2"},
		{Name: "Sun", points: 100, id: "3"},
	}
}

func getStubs(players []*Player, ctrl *gomock.Controller) (*MockCardDealer, *MockMiddleGame) {
	mg := NewMockMiddleGame(ctrl)
	cardDealer := NewMockCardDealer(ctrl)

	// private cards for each player starting from index 0.
	gomock.InOrder(
		cardDealer.EXPECT().DealOne().Return(Card{rank: 10}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 9}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 8}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 7}),

		// public cards.
		cardDealer.EXPECT().DealOne().Return(Card{rank: 10}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 9}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 8}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 7}),
	)

	// first round.
	callingChip := 1
	mg.EXPECT().
		CallOnce(players[0], 1, 5, false). // Liu is calling 1 for 2 times.
		Return(callingChip).
		Times(2)

	mg.EXPECT().InOrOut(players[1], callingChip).Return(true)
	mg.EXPECT().InOrOut(players[2], callingChip).Return(true)
	mg.EXPECT().InOrOut(players[3], callingChip).Return(true)

	// deal 2nd public card
	gomock.InOrder(
		// public cards.
		cardDealer.EXPECT().DealOne().Return(Card{rank: 11}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 4}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 2}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 5}),
	)
	// Liu called 1 again for the 2nd time.
	mg.EXPECT().InOrOut(players[1], 1).Return(true)
	mg.EXPECT().InOrOut(players[2], 1).Return(false) // Gu is out of the game
	mg.EXPECT().InOrOut(players[3], 1).Return(true)

	gomock.InOrder(
		// public cards.
		cardDealer.EXPECT().DealOne().Return(Card{rank: 4}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 13}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 15}),
	)

	// 3nd public card; Sun called 1.
	mg.EXPECT().CallOnce(players[3], 1, 5, false).Return(1)
	mg.EXPECT().InOrOut(players[0], 1).Return(true)
	mg.EXPECT().InOrOut(players[1], 1).Return(false) // Wang is out of the game

	gomock.InOrder(
		// public cards.
		cardDealer.EXPECT().DealOne().Return(Card{rank: 13}),
		cardDealer.EXPECT().DealOne().Return(Card{rank: 13}),
	)
	// Sun called 2.
	mg.EXPECT().
		CallOnce(players[3], 1, 5, true).
		Return(2)
	mg.EXPECT().InOrOut(players[0], 2).Return(true)

	return cardDealer, mg
}

func TestRun(t *testing.T) {
	players := getFourTestingPlayers()
	hiddenCount := 1
	base := 1
	cardDealer, mg := getStubs(players, gomock.NewController(t))
	game := NewGame(0, players, base, hiddenCount, 0, 1, 5, nil)
	winner, pot := game.run(false, mg, cardDealer)
	if winner.Name != "Liu" {
		t.Errorf("Expected Liu wins the game but got:%s", winner.Name)
	}
	if winner.points != 111 {
		t.Errorf("Expected winner to have 110 chips but got:%d", winner.points)
	}
	if pot != 0 {
		t.Errorf("Expected to have a zero pot at the end game but got:%d", pot)
	}
	if players[2].points != 98 {
		t.Errorf("expected to have 98 chips left for player:%s but got:%d", players[2].Name, players[2].points)
	}
	if players[1].points != 97 {
		t.Errorf("Expected %s to have 97 chips left but got:%d", players[1].Name, players[1].points)
	}
	if players[3].points != 94 {
		t.Errorf("Expected %s to have 94 chips left but got:%d", players[3].Name, players[3].points)
	}
}

func Test_PublicScoreFourKind(t *testing.T) {
	type result struct {
		fs   int
		isfk bool
	}
	testCases := []struct {
		p      Player
		result result
		desc   string
	}{
		{
			desc: "four kind without wild card.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, {rank: 10}, {rank: 10}}, privateCards: []Card{{rank: 11}}},
			},
			result: result{fs: 100, isfk: true},
		},
		{
			desc: "foud kind with wild card.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, wildCard, {rank: 10}}, privateCards: []Card{{rank: 11}}},
			},
			result: result{fs: 100, isfk: true},
		},
		{
			desc: "foud kind with four 2s.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 2}, {rank: 2}, wildCard, {rank: 2}}, privateCards: []Card{{rank: 11}}},
			},
			result: result{fs: 68, isfk: true},
		},
	}
	for _, tc := range testCases {
		ps := tc.p.PublicScore()
		if !tc.p.isFourKind {
			t.Fatalf("%s:should be four kind.", tc.desc)
		}
		if ps != tc.result.fs {
			t.Fatalf("Public score should be %d but got:%d", tc.result.fs, ps)
		}
	}
}

func Test_FinalScoreFourKind(t *testing.T) {
	type result struct {
		fs   int
		fkr  int
		isfk bool
	}
	testCases := []struct {
		p      Player
		result result
		desc   string
	}{
		{
			desc: "four kind without wild card",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, {rank: 11}, {rank: 10}}, privateCards: []Card{{rank: 10}}},
			},
			result: result{fs: 111, isfk: true, fkr: 10},
		},
		{
			desc: "four kind with wild card",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, wildCard, {rank: 10}}, privateCards: []Card{{rank: 11}}},
			},
			result: result{fs: 111, isfk: true, fkr: 10},
		},
		{
			desc: "four kind with wild card also with two pairs",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 3}, {rank: 10}, wildCard, {rank: 10}}, privateCards: []Card{{rank: 3}, {rank: 10}}},
			},
			result: result{fs: 106, isfk: true, fkr: 10},
		},
		{
			desc: "four kind with wild card also with two pairs of 2s",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 2}, {rank: 10}, wildCard, {rank: 10}}, privateCards: []Card{{rank: 2}, {rank: 10}}},
			},
			result: result{fs: 104, isfk: true, fkr: 10},
		},
		{
			desc: "four kind with four 2s.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 2}, {rank: 2}, wildCard, {rank: 12}}, privateCards: []Card{{rank: 2}}},
			},
			result: result{fs: 80, isfk: true, fkr: 2},
		},
		{
			desc: "four kind while also having a wild card.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 3}, {rank: 3}, wildCard, {rank: 12}}, privateCards: []Card{{rank: 3}, {rank: 3}}},
			},
			result: result{fs: 300, isfk: true, fkr: 3},
		},
		{
			desc: "four kind in the public cards.",
			p: Player{
				Hand: Hand{publicCards: []Card{{rank: 3}, {rank: 3}, {rank: 3}, {rank: 3}}, privateCards: []Card{{rank: 13}}},
			},
			result: result{fs: 85, isfk: true, fkr: 3},
		},
	}
	for _, tc := range testCases {
		fs := tc.p.FinalScore()
		if tc.p.isFourKind != tc.result.isfk {
			t.Fatalf("%s:should be four kind", tc.desc)
		}
		if fs != tc.result.fs {
			t.Fatalf("%s:Final score should be %d but got %d", tc.desc, tc.result.fs, fs)
		}
		// if tc.p.fourKindRank != tc.result.fkr {
		// 	t.Fatalf("expected to have four kind rank:%d but got %d for %s.", tc.result.fkr, tc.p.fourKindRank, tc.desc)
		// }
	}
}

func TestPlayer_PublicScore(t *testing.T) {
	tests := []struct {
		name   string
		fields Player
		want   int
	}{
		{
			name: "single card",
			want: 10,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}}},
			},
		},
		{
			name: "no players",
			want: 0,
		},
		{
			name: "two cards, no extra",
			want: 23,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}}},
			},
		},
		{
			name: "three cards, no extra",
			want: 26,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}, {rank: 3}}},
			},
		},
		{
			name: "four cards",
			want: 36,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}, {rank: 5}, {rank: 8}}},
			},
		},
		{
			name: "three a kind in four cards, no wild card",
			want: 65,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, {rank: 5}, {rank: 10}}},
			},
		},
		{
			name: "three a kind in four cards, with wild card",
			want: 65,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, {rank: 5}, wildCard}},
			},
		},
		{
			name: "three a kind of 2s in four cards, with wild card",
			want: 41,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 2}, {rank: 2}, {rank: 5}, wildCard}},
			},
		},
		{
			name: "three a kind of 3s in four cards, with wild card",
			want: 41,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 3}, {rank: 3}, {rank: 2}, wildCard}},
			},
		},
		{
			name: "three a kind in three cards, no wild card",
			want: 45,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 5}, {rank: 5}, {rank: 5}}},
			},
		},
		{
			name: "three a kind in three cards, with wild card",
			want: 45,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 5}, wildCard, {rank: 5}}},
			},
		},
		{
			name: "two jokers in 3 cards",
			want: 71,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, {rank: 5}, jokerR}},
			},
		},
		{
			name: "two jokers in 3 cards with wild card",
			want: 71,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, {rank: 5}, wildCard}},
			},
		},
		{
			name: "two jokers in 4 cards",
			want: 73,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, {rank: 5}, jokerR, {rank: 2}}},
			},
		},
		{
			name: "two jokers in 2 cards",
			want: 66,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, jokerR}},
			},
		},
		{
			name: "jokerS with wild card",
			want: 66,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, wildCard}},
			},
		},
		{
			name: "jokerB with wild card",
			want: 66,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerR, wildCard}},
			},
		},
		{
			name: "two jokers in 3 cards and has wild card",
			want: 85,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, wildCard, jokerR}},
			},
		},
		{
			name: "two jokers in 4 cards and has wild card",
			want: 88,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, wildCard, jokerR, {rank: 3}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Name:   tt.fields.Name,
				points: tt.fields.points,
				Hand: Hand{
					privateCards: tt.fields.privateCards,
					publicCards:  tt.fields.publicCards,
				},
			}
			if got := p.PublicScore(); got != tt.want {
				t.Errorf("Player.PublicScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayer_FinalScore(t *testing.T) {
	type fields struct {
		name         string
		chips        int
		privateCards []Card
		publicCards  []Card
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "1 hidden card and 4 public cards, no extra points",
			fields: fields{
				privateCards: []Card{{rank: 5}},
				publicCards:  []Card{{rank: 10}, {rank: 12}, {rank: 4}, {rank: 5}},
			},
			want: 36,
		},
		{
			name: "2 hidden cards and 4 public cards, no extra points",
			fields: fields{
				privateCards: []Card{{rank: 10}, {rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 11}, {rank: 13}, {rank: 6}},
			},
			want: 62,
		},
		{
			name: "2 hidden cards and 4 public cards with 2 pairs, no extra points",
			fields: fields{
				privateCards: []Card{{rank: 5}, {rank: 6}},
				publicCards:  []Card{{rank: 7}, {rank: 7}, {rank: 15}, {rank: 6}},
			},
			want: 46,
		},
		{
			name: "2 hidden cards and 4 public cards, a three kind without wild card",
			fields: fields{
				privateCards: []Card{{rank: 4}, {rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 95,
		},
		{
			name: "1 hidden cards and 4 public cards, a three kind without wild card",
			fields: fields{
				privateCards: []Card{{rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 91,
		},
		{
			name: "2 hidden cards and 4 public cards, two jokers",
			fields: fields{
				privateCards: []Card{jokerB, {rank: 11}},
				publicCards:  []Card{jokerR, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 116,
		},
		{
			name: "2 hidden cards and 4 public cards, two jokers are hidden",
			fields: fields{
				privateCards: []Card{jokerB, jokerR},
				publicCards:  []Card{{rank: 9}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 114,
		},
		{
			name: "1 hidden cards and 4 public cards, two jokers",
			fields: fields{
				privateCards: []Card{jokerB},
				publicCards:  []Card{{rank: 10}, {rank: 5}, {rank: 12}, jokerR},
			},
			want: 93,
		},
		{
			name: "1 hidden cards and 4 public cards, two jokers and three a kind",
			fields: fields{
				privateCards: []Card{jokerB},
				publicCards:  []Card{{rank: 5}, {rank: 5}, {rank: 5}, jokerR},
			},
			want: 45 + 66,
		},
		{
			name: "2 hidden cards and 4 public cards, two pairs of three a kind without wild card.",
			fields: fields{
				privateCards: []Card{{rank: 3}, {rank: 3}},
				publicCards:  []Card{{rank: 5}, {rank: 5}, {rank: 3}, {rank: 5}},
			},
			want: 45 + 39,
		},
		{
			name: "1 hidden cards and 4 public cards, two pairs with a wild card leads to three a kind",
			fields: fields{
				privateCards: []Card{{rank: 3}},
				publicCards:  []Card{{rank: 5}, {rank: 5}, {rank: 3}, wildCard},
			},
			want: 45 + 6,
		},
		{
			name: "2 hidden cards and 4 public cards, two pairs with a wild card leads to three a kind",
			fields: fields{
				privateCards: []Card{{rank: 4}, {rank: 4}},
				publicCards:  []Card{{rank: 5}, {rank: 5}, {rank: 13}, wildCard},
			},
			want: 45 + 21,
		},
		{
			name: "2 hidden cards and 4 public cards, wild card leads to both three a kind and two jokers",
			fields: fields{
				privateCards: []Card{{rank: 4}, {rank: 4}},
				publicCards:  []Card{{rank: 5}, {rank: 5}, jokerR, wildCard},
			},
			want: 84,
		},
		{
			name: "2 hidden cards and 4 public cards, wild card leads to two jokers with a pure three a kind",
			fields: fields{
				privateCards: []Card{{rank: 4}, {rank: 4}},
				publicCards:  []Card{{rank: 5}, {rank: 4}, jokerR, wildCard},
			},
			want: 5 + 19 + 4*4 + 60,
		},
		{
			name: "1 hidden cards and 4 public cards, wild card leads to two jokers with a pure three a kind",
			fields: fields{
				privateCards: []Card{{rank: 4}},
				publicCards:  []Card{{rank: 4}, {rank: 4}, jokerR, wildCard},
			},
			want: 19 + 4*4 + 60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Name:   tt.fields.name,
				points: tt.fields.chips,
				Hand: Hand{
					privateCards: tt.fields.privateCards,
					publicCards:  tt.fields.publicCards,
				},
			}
			if got := p.FinalScore(); got != tt.want {
				t.Errorf("Player.FinalScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayer_FaceScore(t *testing.T) {
	type fields struct {
		name         string
		chips        int
		privateCards []Card
		publicCards  []Card
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "one public card",
			fields: fields{
				publicCards: []Card{{rank: 10}},
			},
			want: 10,
		},
		{
			name: "two public card",
			fields: fields{
				publicCards: []Card{{rank: 10}, {rank: 15}},
			},
			want: 15,
		},
		{
			name: "three public card",
			fields: fields{
				publicCards: []Card{{rank: 10}, {rank: 12}, {rank: 5}},
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Name:   tt.fields.name,
				points: tt.fields.chips,
				Hand: Hand{
					privateCards: tt.fields.privateCards,
					publicCards:  tt.fields.publicCards,
				},
			}
			if got := p.FaceScore(); got != tt.want {
				t.Errorf("Player.FaceScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayer_ReceivePrivateCard(t *testing.T) {
	tests := []struct {
		name     string
		fields   Player
		args     Card
		expected int
	}{
		{
			name: "empty hand receiving one hidden card",
			fields: Player{
				Hand: Hand{privateCards: []Card{}},
			},
			args:     Card{rank: 10},
			expected: 1,
		},
		{
			name: "one private card and receiving another private card",
			fields: Player{
				Hand: Hand{privateCards: []Card{{rank: 5}}},
			},
			args:     Card{rank: 10},
			expected: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Player{
				Name:   tt.fields.Name,
				points: tt.fields.points,
				Hand: Hand{
					privateCards: tt.fields.privateCards,
					publicCards:  tt.fields.publicCards,
				},
			}
			p.ReceivePrivateCard(tt.args)
			if len(p.privateCards) != tt.expected {
				t.Errorf("Expected to get %d private card but got:%d", tt.expected, len(p.privateCards))
			}
		})
	}
}

func TestGame_GetAskingPlayers(t *testing.T) {
	type fields struct {
		players     []*Player
		base        int
		hiddenCount int
	}
	tests := []struct {
		name   string
		fields fields
		args   int
		want   []*Player
	}{
		{
			name:   "first player being the calling player",
			fields: fields{players: []*Player{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}}},
			args:   0,
			want:   []*Player{{Name: "p2"}, {Name: "p3"}},
		},
		{
			name:   "middle player being the calling player",
			fields: fields{players: []*Player{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}, {Name: "p4"}}},
			args:   1,
			want:   []*Player{{Name: "p3"}, {Name: "p4"}, {Name: "p1"}},
		},
		{
			name:   "2nd middle player being the calling player",
			fields: fields{players: []*Player{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}, {Name: "p4"}}},
			args:   2,
			want:   []*Player{{Name: "p4"}, {Name: "p1"}, {Name: "p2"}},
		},
		{
			name:   "last player being the calling player",
			fields: fields{players: []*Player{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}, {Name: "p4"}}},
			args:   3,
			want:   []*Player{{Name: "p1"}, {Name: "p2"}, {Name: "p3"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				players: tt.fields.players,
				// middleGame:  tt.fields.middleGame,
				base:        tt.fields.base,
				hiddenCount: tt.fields.hiddenCount,
				// pointRange:  tt.fields.pointRange,
			}
			if got := g.getAskingPlayers(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Game.GetAskingPlayers() = %v, want %v", got, tt.want)
			}
		})
	}
}

var ps int

func BenchmarkPublicScore(b *testing.B) {
	tests := []struct {
		name   string
		fields Player
		want   int
	}{
		{
			name: "no players",
			want: 0,
		},
		{
			name: "two cards, no extra",
			want: 23,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}}},
			},
		},
		{
			name: "three cards, no extra",
			want: 26,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}, {rank: 3}}},
			},
		},
		{
			name: "four cards",
			want: 36,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 13}, {rank: 5}, {rank: 8}}},
			},
		},
		{
			name: "three a kind in four cards",
			want: 65,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 10}, {rank: 10}, {rank: 5}, {rank: 10}}},
			},
		},
		{
			name: "three a kind in three cards",
			want: 45,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{{rank: 5}, {rank: 5}, {rank: 5}}},
			},
		},
		{
			name: "two jokers in 3 cards",
			want: 71,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, {rank: 5}, jokerR}},
			},
		},
		{
			name: "two jokers in 4 cards",
			want: 73,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, {rank: 5}, jokerR, {rank: 2}}},
			},
		},
		{
			name: "two jokers in 2 cards",
			want: 66,
			fields: Player{
				Name: "P1",
				Hand: Hand{publicCards: []Card{jokerB, jokerR}},
			},
		},
	}
	for _, tt := range tests {
		for i := 0; i < b.N; i++ {
			b.Run(tt.name, func(b *testing.B) {
				p := &Player{
					Name:   tt.fields.Name,
					points: tt.fields.points,
					Hand: Hand{
						privateCards: tt.fields.privateCards,
						publicCards:  tt.fields.publicCards,
					},
				}
				ps = p.PublicScore()
			})
		}
	}
}

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		players := getFourTestingPlayers()
		hiddenCount := 1
		base := 1
		cardDealer, mg := getStubs(players, gomock.NewController(b))
		game := NewGame(0, players, base, hiddenCount, 0, 1, 5, nil)
		game.run(false, mg, cardDealer)
	}
}

func Benchmark_FinalScore(b *testing.B) {
	type fields struct {
		name         string
		chips        int
		privateCards []Card
		publicCards  []Card
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "1 hidden card and 4 public cards, no extra points",
			fields: fields{
				privateCards: []Card{{rank: 5}},
				publicCards:  []Card{{rank: 10}, {rank: 12}, {rank: 4}, {rank: 5}},
			},
			want: 36,
		},
		{
			name: "2 hidden cards and 4 public cards, no extra points",
			fields: fields{
				privateCards: []Card{{rank: 10}, {rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 11}, {rank: 13}, {rank: 6}},
			},
			want: 62,
		},
		{
			name: "2 hidden cards and 4 public cards, a three kind",
			fields: fields{
				privateCards: []Card{{rank: 4}, {rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 95,
		},
		{
			name: "1 hidden cards and 4 public cards, a three kind",
			fields: fields{
				privateCards: []Card{{rank: 12}},
				publicCards:  []Card{{rank: 10}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 91,
		},
		{
			name: "2 hidden cards and 4 public cards, two jokers",
			fields: fields{
				privateCards: []Card{{rank: 17, suit: "joker S"}, {rank: 11}},
				publicCards:  []Card{{rank: 19, suit: "joker B"}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 116,
		},
		{
			name: "2 hidden cards and 4 public cards, two jokers are hidden",
			fields: fields{
				privateCards: []Card{{rank: 17, suit: "joker S"}, {rank: 19, suit: "joker B"}},
				publicCards:  []Card{{rank: 9}, {rank: 15}, {rank: 12}, {rank: 12}},
			},
			want: 114,
		},
		{
			name: "1 hidden cards and 4 public cards, two jokers",
			fields: fields{
				privateCards: []Card{{rank: 17, suit: "joker S"}},
				publicCards:  []Card{{rank: 10}, {rank: 5}, {rank: 12}, {rank: 19, suit: "joker B"}},
			},
			want: 93,
		},
		{
			name: "1 hidden cards and 4 public cards, two jokers and three a kind",
			fields: fields{
				privateCards: []Card{{rank: 17, suit: "joker S"}},
				publicCards:  []Card{{rank: 5}, {rank: 5}, {rank: 5}, {rank: 19, suit: "joker B"}},
			},
			want: 45 + 66,
		},
	}
	for _, tt := range tests {
		for i := 0; i < b.N; i++ {
			b.Run(tt.name, func(b *testing.B) {
				p := &Player{
					Name:   tt.fields.name,
					points: tt.fields.chips,
					Hand: Hand{
						privateCards: tt.fields.privateCards,
						publicCards:  tt.fields.publicCards,
					},
				}
				p.FinalScore()
			})
		}
	}
}

func TestGame_getCallingPlayer(t *testing.T) {
	type fields struct {
		players     []*Player
		hiddenCount int
		prevWinner  *Player
	}
	type args struct {
		isFirstRound bool
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Player
	}{
		{
			name: "one hidden card, first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{publicCards: []Card{{rank: 4}}}},
					{Name: "p2", id: "2", Hand: Hand{publicCards: []Card{{rank: 5}}}},
					{Name: "p3", id: "3", Hand: Hand{publicCards: []Card{{rank: 6}}}},
				},
				hiddenCount: 1,
			},
			args: args{true},
			want: &Player{Name: "p3", id: "3", Hand: Hand{publicCards: []Card{{rank: 6}}}},
		},
		{
			name: "one hidden card, non-first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{publicCards: []Card{{rank: 4}, {rank: 3}}}},
					{Name: "p2", id: "2", Hand: Hand{publicCards: []Card{{rank: 5}, {rank: 5}}}},
					{Name: "p3", id: "3", Hand: Hand{publicCards: []Card{{rank: 6}, {rank: 2}}}},
				},
				hiddenCount: 1,
			},
			args: args{false},
			want: &Player{Name: "p2", id: "2", Hand: Hand{publicCards: []Card{{rank: 5}, {rank: 5}}}},
		},
		{
			name: "two hidden cards, first game first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{privateCards: []Card{{rank: 4}, {rank: 3}}}},
					{Name: "p2", id: "2", Hand: Hand{privateCards: []Card{{rank: 5}, {rank: 5}}}},
					{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 12}}}},
				},
				hiddenCount: 2,
				// prevWinner:  &Player{Name: "p4", id: 4, Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}}},
			},
			args: args{true},
			want: &Player{Name: "p1", id: "1", Hand: Hand{privateCards: []Card{{rank: 4}, {rank: 3}}}},
		},
		{
			name: "two hidden cards, first game non-first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{privateCards: []Card{{rank: 4}, {rank: 3}}, publicCards: []Card{{rank: 10}}}},
					{Name: "p2", id: "2", Hand: Hand{privateCards: []Card{{rank: 5}, {rank: 5}}, publicCards: []Card{{rank: 13}}}},
					{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 12}}, publicCards: []Card{{rank: 6}}}},
				},
				hiddenCount: 2,
				// prevWinner:  &Player{Name: "p3", id: 3, Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 12}}}},
			},
			args: args{false},
			want: &Player{Name: "p2", id: "2", Hand: Hand{privateCards: []Card{{rank: 5}, {rank: 5}}, publicCards: []Card{{rank: 13}}}},
		},
		{
			name: "two hidden cards, non-first game first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{privateCards: []Card{{rank: 4}, {rank: 3}}}},
					{Name: "p2", id: "2", Hand: Hand{privateCards: []Card{{rank: 5}, {rank: 5}}}},
					{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}}},
				},
				hiddenCount: 2,
				prevWinner:  &Player{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}}},
			},
			args: args{true},
			want: &Player{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}}},
		},
		{
			name: "two hidden cards, non-first game non-first round",
			fields: fields{
				players: []*Player{
					{Name: "p1", id: "1", Hand: Hand{privateCards: []Card{{rank: 4}, {rank: 3}}, publicCards: []Card{{rank: 10}}}},
					{Name: "p2", id: "2", Hand: Hand{privateCards: []Card{{rank: 5}, {rank: 5}}, publicCards: []Card{{rank: 11}}}},
					{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}, publicCards: []Card{{rank: 12}}}},
				},
				hiddenCount: 2,
				prevWinner:  &Player{Name: "p36", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 12}}}},
			},
			args: args{false},
			want: &Player{Name: "p3", id: "3", Hand: Hand{privateCards: []Card{{rank: 6}, {rank: 2}}, publicCards: []Card{{rank: 12}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{
				players:     tt.fields.players,
				hiddenCount: tt.fields.hiddenCount,
				prevWinner:  tt.fields.prevWinner,
			}
			if got := g.getCallingPlayer(tt.args.isFirstRound); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Game.getCallingPlayer() = %v, want %v", got, tt.want)
			}
		})
	}
}
