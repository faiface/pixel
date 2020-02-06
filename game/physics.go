package game

import "math/rand"

func moveBot(b *Bot, s State) {
	if b.a == 0 {
		b.a = 1
	}
	b.a += rand.Intn(3) - 1
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

	newPos := b.Pos + b.v
	if !collide(b.id, newPos, b.Lane, s) {
		b.Pos = newPos
	}
}

func collide(id, pos, lane int, s State) bool {
	for _, o := range s.Obstacles {
		if o.Pos == pos && o.Lane == lane {
			return true
		}
	}
	return false
}

const (
	passDistance = 2
	baseAccel    = 1
)
