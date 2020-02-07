package game

import (
	"log"
)

type State struct {
	Teams     []Team
	Obstacles []Obstacle
	GameOver  bool
}

func NewState() State {
	var teams []Team
	for i := 0; i < NumTeams; i++ {
		var bots []Bot
		for j := 0; j < numBots; j++ {
			b := Bot{
				id:   i*NumTeams + j,
				Lane: i,
				Pos:  j * (Steps / numBots),
			}
			bots = append(bots, b)
		}
		teams = append(teams, Team{
			id:    i,
			Bots:  bots,
			Baton: Baton{Holder: &bots[0]},
			Lane:  i,
		})
	}

	return State{
		Teams: teams,
		Obstacles: []Obstacle{
			{
				Lane: 0,
				Pos:  Steps / 3,
			},
			{
				Lane: 1,
				Pos:  Steps * 2 / 3,
			},
			{
				Lane: 2,
				Pos:  Steps / 2,
			},
			{
				Lane: 3,
				Pos:  Steps * 3 / 4,
			},
		},
	}
}

type Team struct {
	id    int
	Bots  []Bot
	Baton Baton
	won   bool
	Lane  int
}

type Bot struct {
	id   int
	Lane int
	Pos  int
	v    int
	a    int
}

type Baton struct {
	Holder *Bot
}

type Obstacle struct {
	Lane int
	Pos  int
}

func UpdateState(sOld State) State {
	s := sOld

	for i, t := range s.Teams {
		doCommand(chooseCommand(t, sOld), t.Baton.Holder)
		moveBot(t.Baton.Holder, sOld)
		maybePassBaton(&s.Teams[i])
	}

	for _, t := range s.Teams {
		if won(*t.Baton.Holder, s) {
			s.GameOver = true
		}
	}

	return s
}

func maybePassBaton(t *Team) {
	for i, b := range t.Bots {
		h := t.Baton.Holder
		if h.id >= b.id || h.Lane != b.Lane {
			continue
		}
		if abs(b.Pos-h.Pos) <= passDistance {
			log.Printf("team %v pass from %v to %v!", t.id, h.id, b.id)
			t.Baton.Holder.v = 0
			t.Baton.Holder.a = 0
			t.Baton.Holder = &t.Bots[i]
			t.Bots[i].a = baseAccel
			return
		}
	}
}

func won(b Bot, s State) bool {
	return b.Pos >= Steps
}

func gameOver(s State) bool {
	for _, t := range s.Teams {
		if t.won {
			return true
		}
	}
	return false
}

const (
	Steps    = 50
	numBots  = 5
	NumTeams = 4
	NumLanes = 6
	maxA     = 3
	maxV     = 10
)

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
