package game

import "log"

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
		if o := collide(b.Position.Pos+1, b.Position.Lane, s); o != nil {
			log.Printf("bot %d crashed into %#v!", b.ID, o)
			break
		} else {
			b.Position.Pos++
		}
	}

	s = updateBot(s, teamID, b)
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
	passDistance = 3
	baseAccel    = 1
)
