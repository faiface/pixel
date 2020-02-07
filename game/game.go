package game

import "log"

func UpdateState(s State, sOld State) State {
	for i := range s.Teams {
		s = doCommand(chooseCommand(s, i), s, i)
		if b := activeBot(s.Teams[i]); b != nil {
			s = moveBot(s, i, *b)
		}
		s = maybePassBaton(s, i)
	}

	for _, t := range s.Teams {
		if b := activeBot(t); b != nil && won(*b, s) {
			log.Printf("team %d won", t.id)
			s.GameOver = true
		}
	}

	return s
}

func maybePassBaton(s State, teamID int) State {
	t := s.Teams[teamID]
	h := activeBot(t)
	if h == nil {
		return s
	}

	for i, b := range t.Bots {
		if h.ID >= b.ID || h.Lane != b.Lane {
			continue
		}
		if abs(b.Pos-h.Pos) <= passDistance {
			h.v = 0
			h.a = 0
			s = updateBot(s, teamID, *h)
			newH := t.Bots[i]
			newH.a = baseAccel
			t.Baton.HolderID = newH.ID
			s = updateTeam(s, t)
			return updateBot(s, teamID, newH)
		}
	}

	return s
}

func activeBot(t Team) *Bot {
	for _, b := range t.Bots {
		if b.ID == t.Baton.HolderID {
			return &b
		}
	}
	return nil
}

func updateBot(s State, teamID int, b Bot) State {
	t := s.Teams[teamID]
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

func (t Team) BatonHolder() *Bot {
	for _, b := range t.Bots {
		if b.ID == t.Baton.HolderID {
			return &b
		}
	}
	return nil
}

type Bot struct {
	ID   int
	Lane int
	Pos  int
	v    int
	a    int
}

type Baton struct {
	HolderID int
}

type Obstacle struct {
	Lane int
	Pos  int
}

func NewState() State {
	var teams []Team
	for i := 0; i < NumTeams; i++ {
		var bots []Bot
		for j := 0; j < numBots; j++ {
			b := Bot{
				ID:   i*NumTeams + j,
				Lane: i,
				Pos:  j * (Steps / numBots),
			}
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

const (
	Steps    = 60
	numBots  = 5
	NumTeams = 4
	NumLanes = 4
	maxA     = 2
	maxV     = 8
)
