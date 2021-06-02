package main

import (
	"douji"
	"fmt"
)

// a self-playing middle game
type selfMiddleGame struct{}

func (smg selfMiddleGame) InOrOut(player *douji.Player, chips int) bool {
	fmt.Printf("Asking:%s. Press y for in, anything else for out.\n", player.Name)
	var answer string
	fmt.Scan(&answer)
	return answer == "y"
}

func (smg selfMiddleGame) CallOnce(p *douji.Player, step, end int, lastCall bool) int {
	fmt.Printf("%s, how much do you want to you call:", p.Name)
	for i := 0; i <= end; i += step {
		fmt.Print(i, " ")
	}
	if lastCall {
		fmt.Print(end*2, " ")
	}
	fmt.Print("?")
	var calling int
	var err error
	_, err = fmt.Scan(&calling)
	for err != nil || calling < 0 || (lastCall && calling > 2*end) || (!lastCall && calling > end) {
		return smg.CallOnce(p, step, end, lastCall)
	}
	fmt.Printf("%s called:%d\n", p.Name, calling)
	return calling
}

// for testing!
type inMemoryDb struct{}

func (imdb inMemoryDb) SaveGameStats(setId string, gameId int, pnp []douji.PlayerDTO) error {
	return nil
}

func (imdb inMemoryDb) LoadPlayerStatsByName(name string) *douji.Player {
	return douji.NewTestPlayer(name, name, 1000)
}

// SaveSet(s *Set) error
func (imdb inMemoryDb) CreatePlayer(name, password string, points int) (string, error) {
	return "", nil
}

func chooseDb() int {
	fmt.Println("Choose database mode:\n1 in-memory (for local testing, ok to ignore missing douji.env file.)\n2 for LeanCloud.")
	var mode int
	fmt.Scan(&mode)
	if mode == 1 || mode == 2 {
		return mode
	}
	fmt.Println("invalid db mode choice!")
	return chooseDb()
}

func main() {
	var db douji.Db
	dbMode := chooseDb()
	if dbMode == 1 {
		db = inMemoryDb{}
	} else {
		db = douji.NewLeanCloudDB()
	}

	players := []*douji.Player{
		db.LoadPlayerStatsByName("Liu"),
		db.LoadPlayerStatsByName("Sun"),
		db.LoadPlayerStatsByName("Gu"),
		db.LoadPlayerStatsByName("Wang"),
		db.LoadPlayerStatsByName("Pan"),
		db.LoadPlayerStatsByName("Mu"),
	}

	// douji.NewPlayer("Zhang San", "password1", 1000, db),
	// douji.NewPlayer("Liu Wu", "password1", 1000, db),
	// douji.NewPlayer("Sun", "password1", 1000, db),
	// douji.NewPlayer("Gu", "password2", 1000, db),
	// douji.NewPlayer("Wang", "password3", 1000, db),
	// douji.NewPlayer("Pan", "password4", 1000, db),
	// douji.NewPlayer("Mu", "password5", 1000, db),

	p, base := 0, 1
	s := douji.NewSet(2, true)
	// if err := db.SaveSet(&s); err != nil {
	// 	fmt.Errorf("error on saving set:%w", err)
	// }
	hiddenCount := 1
	s.Run(players, selfMiddleGame{}, db, base, hiddenCount, p)
}
