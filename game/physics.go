package game

func accelerate(b Bot) Bot {
	if b.a < -MaxA {
		b.a = -MaxA
	}
	if b.a > MaxA {
		b.a = MaxA
	}

	b.v += b.a
	if b.v > MaxV {
		b.v = MaxV
	}
	if b.v < -MaxV {
		b.v = -MaxV
	}

	return b
}

func moveBot(s State, b Bot) State {
	for i := 0; i < b.v; i++ {
		if o := collide(b.Position.Pos+1, b.Position.Lane, s); o != nil {
			return destroyBot(s, b)
		} else {
			b.Position.Pos++
		}
	}

	s = updateBot(s, b)
	return s
}

func collide(pos, lane int, s State) interface{} {
	for _, o := range s.Obstacles {
		if o.Position.Pos == pos && o.Position.Lane == lane {
			return o
		}
	}
	for _, t := range s.Teams {
		for _, b := range t.Bots {
			if b.Position.Pos == pos && b.Position.Lane == lane {
				return b
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
