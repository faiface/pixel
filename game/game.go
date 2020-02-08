package game

import (
	"log"
	"math/rand"
)

func UpdateState(s State, sOld State) State {
	for i := range s.Teams {
		cmd := smartChooseCommand(s, i)
		log.Printf("team %d chose to %v", i, cmd)
		s = doCommand(cmd, s, i)
	}

	for _, t := range s.Teams {
		if b := ActiveBot(t); b != nil && won(*b, s) {
			log.Printf("team %d won", t.id)
			s.GameOver = true
		}
	}

	return s
}

func maybePassBaton(s State, teamID int) State {
	t := s.Teams[teamID]
	h := ActiveBot(t)
	if h == nil {
		return s
	}

	for i, b := range t.Bots {
		if h.ID >= b.ID || h.Position.Lane != b.Position.Lane {
			continue
		}
		if abs(b.Position.Pos-h.Position.Pos) <= PassDistance {
			h.v = 0
			h.a = 0
			s = updateBot(s, *h)
			newH := t.Bots[i]
			newH.a = baseAccel
			t.Baton.HolderID = newH.ID
			s = updateTeam(s, t)
			return updateBot(s, newH)
		}
	}

	return s
}

func ActiveBot(t Team) *Bot {
	for _, b := range t.Bots {
		if b.ID == t.Baton.HolderID {
			return &b
		}
	}
	return nil
}

func updateBot(s State, b Bot) State {
	t := s.Teams[b.TeamID]
	for i, bb := range t.Bots {
		if bb.ID == b.ID {
			bots := append([]Bot{}, t.Bots[:i]...)
			bots = append(bots, b)
			bots = append(bots, t.Bots[i+1:]...)
			t.Bots = bots
			break
		}
	}

	s = updateTeam(s, t)
	return s
}

func updateTeam(s State, t Team) State {
	teams := append([]Team{}, s.Teams[:t.id]...)
	teams = append(teams, t)
	teams = append(teams, s.Teams[t.id+1:]...)
	s.Teams = teams

	return s
}

func destroyBot(s State, b Bot) State {
	// insert obstacle where bot was
	s.Obstacles = append(s.Obstacles, Obstacle{Position: b.Position})

	// spawn bot back at starting position
	b.Position = b.StartPos

	return updateBot(s, b)
}

func removeObstacle(s State, pos Position) State {
	for i, o := range s.Obstacles {
		if o.Position == pos {
			var os []Obstacle
			os = append(os, s.Obstacles[:i]...)
			os = append(os, s.Obstacles[i+1:]...)
			s.Obstacles = os
			//s.Obstacles = append([]Obstacle{}, append(s.Obstacles[:i], s.Obstacles[i+1:]...)...)
			break
		}
	}
	return s
}

func won(b Bot, s State) bool {
	return b.Position.Pos >= Steps
}

func gameOver(s State) bool {
	for _, t := range s.Teams {
		if t.won {
			return true
		}
	}
	return false
}

func legalMove(s State, teamID int, cmd command) bool {
	b := ActiveBot(s.Teams[teamID])
	if b == nil {
		return false
	}

	switch cmd {
	case left:
		return b.Position.Lane < NumLanes-1
	case right:
		return b.Position.Lane > 0

	}
	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

type State struct {
	Teams     []Team
	Obstacles []Obstacle
	GameOver  bool
}

type Team struct {
	id    int
	Bots  []Bot
	Baton Baton
	won   bool
	Lane  int
}

type Bot struct {
	ID       int
	TeamID   int
	Position Position
	StartPos Position
	v        int
	a        int
}

type Position struct {
	Lane int
	Pos  int
}

type Battery struct {
	Capacity int
	Charge   int
}

type Baton struct {
	HolderID int
}

type Obstacle struct {
	Position Position
}

func NewState() State {
	var teams []Team
	for i := 0; i < NumTeams; i++ {
		var bots []Bot
		for j := 0; j < numBots; j++ {
			b := Bot{
				ID:     i*NumTeams + j,
				TeamID: i,
				StartPos: Position{
					Lane: i,
					Pos:  j * (Steps / numBots),
				},
			}
			b.Position = b.StartPos
			bots = append(bots, b)
		}
		teams = append(teams, Team{
			id:    i,
			Bots:  bots,
			Baton: Baton{HolderID: i * NumTeams},
			Lane:  i,
		})
	}

	return State{
		Teams:     teams,
		Obstacles: randomObstacles(teams),
	}
}

func randomObstacles(teams []Team) []Obstacle {
	var os []Obstacle

	const numObstacles = 5 * NumTeams
	for i := 0; i < numObstacles; i++ {
		os = append(os, Obstacle{
			Position: randomOpenPosition(teams, os),
		})
	}

	return os
}

func randomOpenPosition(ts []Team, os []Obstacle) Position {
	for {
		p := Position{
			Pos:  rand.Intn(Steps-8) + 4,
			Lane: rand.Intn(NumLanes),
		}
		if positionOpen(p, ts, os) {
			return p
		}
	}
}

func positionOpen(pos Position, ts []Team, os []Obstacle) bool {
	for _, t := range ts {
		for _, b := range t.Bots {
			if b.Position == pos {
				return false
			}
		}
	}
	for _, o := range os {
		if o.Position == pos {
			return false
		}
	}
	return true
}

var (
	staticObstacles = []Obstacle{
		{
			Position: Position{
				Lane: 0,
				Pos:  Steps / 3,
			},
		},
		{
			Position: Position{
				Lane: 1,
				Pos:  Steps * 2 / 3,
			},
		},
		{
			Position: Position{
				Lane: 2,
				Pos:  Steps / 2,
			},
		},
		{
			Position: Position{
				Lane: 3,
				Pos:  Steps * 3 / 4,
			},
		},
	}
)

const (
	Steps    = 50
	numBots  = 5
	NumTeams = 8
	NumLanes = NumTeams
)
