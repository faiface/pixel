package game

import (
	"log"
)

type State struct {
	Teams       []Team
	SpawnPoints map[int]SpawnPoint // keys are racer IDs
	Obstacles   []Obstacle
	Derelicts   []Obstacle
	GameOver    bool
}

type Team struct {
	id     int
	Racers []Racer
	Baton  Baton
	won    bool
	Lane   int
}

type Racer struct {
	ID       int
	TeamID   int
	Position Position
	Kinetics Kinetics
	Battery  Battery
}

func ActiveRacer(t Team) *Racer {
	for _, r := range t.Racers {
		if r.ID == t.Baton.HolderID {
			rr := r
			return &rr
		}
	}
	return nil
}

func updateRacer(s State, r Racer) State {
	t := s.Teams[r.TeamID]
	for i, rr := range t.Racers {
		if rr.ID == r.ID {
			racers := append([]Racer{}, t.Racers[:i]...)
			racers = append(racers, r)
			racers = append(racers, t.Racers[i+1:]...)
			t.Racers = racers
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

func destroyRacer(s State, r Racer) State {
	// insert derelict where racer was
	s.Derelicts = append(s.Derelicts, Obstacle{Position: r.Position})

	// spawn racer back at starting position
	r.Position = s.SpawnPoints[r.ID].Pos
	r.Kinetics = Kinetics{}
	r.Battery.Charge = r.Battery.Capacity

	return updateRacer(s, r)
}

type Battery struct {
	Capacity int
	Charge   int
}

type Baton struct {
	HolderID int
}

func checkWin(s State, sOld State) State {
	var winners []int
	for _, t := range s.Teams {
		if r := ActiveRacer(t); r != nil {
			if won(*r, s) {
				winners = append(winners, t.id)
			}
		}
	}

	switch n := len(winners); {
	case n == 1:
		log.Printf("team %d won", winners[0])
		s.GameOver = true
	case n > 1:
		log.Printf("%d-way tie between teams: %v", len(winners), winners)
		s.GameOver = true
	}

	return s
}

func maybePassBaton(s State, teamID int) State {
	t := s.Teams[teamID]
	h := ActiveRacer(t)
	if h == nil {
		return s
	}

	for i, r := range t.Racers {
		if h.ID >= r.ID || h.Position.Lane != r.Position.Lane {
			continue
		}
		if abs(r.Position.Pos-h.Position.Pos) <= PassDistance {
			h.Kinetics.VX = 0
			h.Kinetics.A = 0
			s = updateRacer(s, *h)
			newH := t.Racers[i]
			newH.Kinetics.VX = 1
			t.Baton.HolderID = newH.ID
			s = updateTeam(s, t)
			return updateRacer(s, newH)
		}
	}

	return s
}

func won(r Racer, s State) bool {
	return r.Position.Pos >= Steps
}

func gameOver(s State) bool {
	for _, t := range s.Teams {
		if t.won {
			return true
		}
	}
	return false
}

func legalMove(s State, teamID int, cmd Command) bool {
	r := ActiveRacer(s.Teams[teamID])
	if r == nil {
		return false
	}

	switch cmd {
	case left:
		return r.Position.Lane < NumLanes-1
	case right:
		return r.Position.Lane > 0

	}
	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func NewState() State {
	spawns := make(map[int]SpawnPoint)

	var teams []Team
	for i := 0; i < NumTeams; i++ {
		var racers []Racer
		for j := 0; j < numRacers; j++ {
			r := Racer{
				ID:     i*NumTeams + j,
				TeamID: i,
				Position: Position{
					Lane: i,
					Pos:  j * (Steps / numRacers),
				},
				Battery: Battery{
					Capacity: baseCharge,
					Charge:   baseCharge,
				},
			}
			spawns[r.ID] = SpawnPoint{
				TeamID: i,
				Pos:    r.Position,
			}
			racers = append(racers, r)
		}
		teams = append(teams, Team{
			id:     i,
			Racers: racers,
			Baton:  Baton{HolderID: i * NumTeams},
			Lane:   i,
		})
	}

	return State{
		Teams:       teams,
		SpawnPoints: spawns,
		Obstacles:   randomObstacles(teams),
	}
}

const (
	Steps      = 80
	numRacers  = 3
	NumTeams   = 8
	NumLanes   = NumTeams
	baseCharge = 14
)
