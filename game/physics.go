package game

func accelerate(b Bot) Bot {
	if b.a < -maxA {
		b.a = -maxA
	}
	if b.a > maxA {
		b.a = maxA
	}

	b.v += b.a
	if b.v > maxV {
		b.v = maxV
	}
	if b.v < -maxV {
		b.v = -maxV
	}

	return b
}

func moveBot(s State, teamID int, b Bot) State {
	for i := 0; i < b.v; i++ {
		if !collide(b.Pos+1, b.Lane, s) {
			b.Pos++
		}
	}

	s = updateBot(s, s, teamID, b)
	return s
}

func collide(pos, lane int, s State) bool {
	for _, o := range s.Obstacles {
		if o.Pos == pos && o.Lane == lane {
			return true
		}
	}
	for _, t := range s.Teams {
		for _, b := range t.Bots {
			if b.Pos == pos && b.Lane == lane {
				return true
			}
		}
	}
	return false
}

const (
	passDistance = 1
	baseAccel    = 1
)
