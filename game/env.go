package game

import "math/rand"

type Obstacle struct {
	Position Position
}

func removeObstacle(s State, pos Position) State {
	for i, o := range s.Obstacles {
		if o.Position == pos {
			var os []Obstacle
			os = append(os, s.Obstacles[:i]...)
			os = append(os, s.Obstacles[i+1:]...)
			s.Obstacles = os
			break
		}
	}
	return s
}

func positionOpen(pos Position, ts []Team, os []Obstacle) bool {
	for _, t := range ts {
		for _, r := range t.Racers {
			if r.Position == pos {
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

func randomObstacles(teams []Team) []Obstacle {
	var os []Obstacle

	const numObstacles = 3 * NumTeams
	for i := 0; i < numObstacles; i++ {
		os = append(os, Obstacle{
			Position: randomOpenPosition(teams, os),
		})
	}

	return os
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
