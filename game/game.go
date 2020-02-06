package game

import (
	"log"
	"math/rand"
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
			Bots:  bots,
			Baton: Baton{Holder: &bots[0]},
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
		},
	}
}

type Team struct {
	Bots  []Bot
	Baton Baton
	won   bool
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

func doCommand(cmd command, b *Bot) {
	da := 1
	da += rand.Intn(3) - 1

	switch cmd {
	case speedUp:
		b.a += da
		accelerate(b)
	case slowDown:
		b.a -= da
		accelerate(b)
	case left:
		b.Lane++
	}
}

func chooseCommand(t Team, s State) command {
	h := t.Baton.Holder
	if collide(h.Pos+1, h.Lane, s) {
		return left
	}

	var nextBot *Bot
	for i, b := range t.Bots {
		if b.id == h.id+1 {
			nextBot = &t.Bots[i]
			break
		}
	}

	if nextBot != nil {
		if h.Lane != nextBot.Lane {
			if abs(nextBot.Pos-h.Pos) < h.v {
				log.Println("WHOOOOOOA")
				return slowDown
			}
		}
	}

	return speedUp
}

type command int

const (
	speedUp command = iota
	slowDown
	left
)

func maybePassBaton(t *Team) {
	for i, b := range t.Bots {
		h := t.Baton.Holder
		if h.id >= b.id || h.Lane != b.Lane {
			continue
		}
		if abs(b.Pos-h.Pos) <= passDistance {
			log.Printf("pass from %v to %v!", h.id, b.id)
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
	NumTeams = 2
	maxA     = 3
	maxV     = 10
)

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
