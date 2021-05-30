package douji

type PlayerDTO struct {
	Id     string `json:"player_id"`
	Name   string `json:"player_name"`
	Points int    `json:"points"`
}

type Db interface {
	SaveGameStats(setId string, gameId int, pnp []PlayerDTO) error
	LoadPlayerStatsByName(name string) *Player
	SaveSet(s *Set) error
	CreatePlayer(name, password string, points int) (string, error)
}
