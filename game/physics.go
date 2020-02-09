package game

type Position struct {
	Lane int
	Pos  int
}

type Kinetics struct {
	VX int
	VY int
	A  int
}

func accelerate(r Racer) Racer {
	if r.Kinetics.A < -MaxA {
		r.Kinetics.A = -MaxA
	}
	if r.Kinetics.A > MaxA {
		r.Kinetics.A = MaxA
	}

	if r.Kinetics.A > r.Battery.Charge {
		r.Kinetics.A = r.Battery.Charge
	}
	r.Battery.Charge -= abs(r.Kinetics.A)

	r.Kinetics.VX += r.Kinetics.A
	if r.Kinetics.VX > MaxV {
		r.Kinetics.VX = MaxV
	}
	if r.Kinetics.VX < -MaxV {
		r.Kinetics.VX = -MaxV
	}

	return r
}

func moveRacer(s State, r Racer) State {
	r.Battery.Charge--
	s = updateRacer(s, r)
	if r.Battery.Charge <= 0 {
		return destroyRacer(s, r)
	}

	for i := 0; i < r.Kinetics.VX; i++ {
		if o := collide(r.Position.Pos+1, r.Position.Lane, s); o != nil {
			return destroyRacer(s, r)
		} else {
			r.Position.Pos++
		}
	}

	r.Position.Lane += r.Kinetics.VY
	r.Kinetics.VY = 0

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
	MaxV         = 4
	PassDistance = 2
)
