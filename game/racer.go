package game

type Racer struct {
	ID     int
	TeamID int

	Position
	Kinetics
	Battery
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

func destroyRacer(s State, r Racer) State {
	// insert derelict where racer was
	s.Derelicts = append(s.Derelicts, Obstacle{Position: r.Position})

	// spawn racer back at starting position
	r.Position = s.SpawnPoints[r.ID].Pos
	r.Kinetics = Kinetics{}
	r.Battery.Charge = r.Battery.Capacity

	return updateRacer(s, r)
}
