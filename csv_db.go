package douji

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type csvDb struct {
	file string
	// writer *csv.Writer
}

func NewCSV() csvDb {
	return csvDb{file: "douji.csv"}
}

func (c csvDb) SaveGameStats(setId string, gameId int, players []PlayerDTO) error {
	file, err := os.OpenFile(c.file, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic("cannot open csv file.")
	}
	defer file.Close()
	w := csv.NewWriter(file)
	for _, p := range players {
		// d := fmt.Sprintf("%d,%d,%s,%d", setId, gameId, p.Name, p.points)
		// sd := strings.Split(d, ",")
		d := dataToWrite(setId, gameId, p.Name, p.Points)
		fmt.Println(d)

		if err := w.Write(d); err != nil {
			panic(err)
		}
	}
	w.Flush()
	return nil
}

func (csv csvDb) CreatePlayer(name, password string, points int) (string, error) {
	return "", nil
}

func (c csvDb) LoadPlayerStatsByName(name string) *Player {
	// TODO: fix this!
	return &Player{}
	// return new PlayerDTO{ID: playerID, }
}

func dataToWrite(setId string, gameId int, name string, points int) []string {
	gid := fmt.Sprintf("%d", gameId)
	p := fmt.Sprintf("%d", points)
	return []string{setId, gid, name, p, time.Now().Local().String()}
}
