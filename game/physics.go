package game

func accelerate(r Racer) Racer {
	if r.Kinetics.A < -MaxA {
		r.Kinetics.A = -MaxA
	}
	if r.Kinetics.A > MaxA {
		r.Kinetics.A = MaxA
	}

	r.Kinetics.V += r.Kinetics.A
	if r.Kinetics.V > MaxV {
		r.Kinetics.V = MaxV
	}
	if r.Kinetics.V < -MaxV {
		r.Kinetics.V = -MaxV
	}

	return r
}

func moveRacer(s State, r Racer) State {
	for i := 0; i < r.Kinetics.V; i++ {
		if o := collide(r.Position.Pos+1, r.Position.Lane, s); o != nil {
			return destroyRacer(s, r)
		} else {
			r.Position.Pos++
		}
	}

	s = updateRacer(s, r)
	return s
}

func collide(pos, lane int, s State) interface{} {
	for _, o := range s.Obstacles {
		if o.Position.Pos == pos && o.Position.Lane == lane {
			return o
		}
	}
	for _, t := range s.Teams {
		for _, r := range t.Racers {
			if r.Position.Pos == pos && r.Position.Lane == lane {
				return r
			}
		}
	}
	return nil
}

const (
	baseAccel    = 1
	MaxA         = 1
	MaxV         = 2
	PassDistance = 2
)
