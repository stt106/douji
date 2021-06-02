package douji

import (
	"github.com/leancloud/go-sdk/leancloud"
)

type GameStats struct {
	leancloud.Object
	SetId    string `json:"set_id"`
	GameId   int    `json:"game_id"`
	Name     string `json:"player_name"`
	PlayerId string `json:"player_id"`
	Points   int    `json:"points"`
}

// LeanCloudDB is a wrapper of LeanCloud which is a serverless cloud provider.
type LeanCloudDB struct {
	client *leancloud.Client
}

func NewLeanCloudDB() LeanCloudDB {
	client := leancloud.NewEnvClient()
	return LeanCloudDB{client: client}
}

func (lc LeanCloudDB) SaveSet(s *Set) error {
	or, err := lc.client.Class(set).Create(s)
	if err != nil {
		return err
	}
	s.id = or.ID
	return nil
}

func (lc LeanCloudDB) CreatePlayer(name string, password string, points int) (string, error) {
	player, err := lc.client.Users.SignUp(name, password)
	if err != nil {
		panic(err)
	}

	err = lc.client.User(player).Set("points", points, leancloud.UseUser(player))
	if err != nil {
		panic(err)
	}
	return player.ID, nil
}

func (lc LeanCloudDB) LoadPlayerStatsByName(name string) *Player {
	ret := []PlayerDTO{}
	if err := lc.client.Class(gameStatsClass).NewQuery().EqualTo("player_name", name).Order("-createdAt").Find(&ret); err != nil {
		panic(err)
	}
	return &Player{Name: name, points: ret[0].Points, id: ret[0].Id}
}

const (
	gameStatsClass = "GameStat"
	set            = "Set"
	// player         = "Player"
)

func (lc LeanCloudDB) SaveGameStats(setId string, gameId int, pnp []PlayerDTO) error {
	for _, p := range pnp {
		gs := GameStats{SetId: setId, GameId: gameId, Name: p.Name, Points: p.Points, PlayerId: p.Id}
		if _, err := lc.client.Class(gameStatsClass).Create(&gs); err != nil {
			panic(err)
		}
	}
	return nil
}
