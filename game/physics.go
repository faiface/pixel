package game

import "math/rand"

func accelerate(b *Bot) {
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
}

func moveBot(b *Bot, s State) {
	for i := 0; i < b.v; i++ {
		if !collide(b.id, b.Pos+1, b.Lane, s) {
			b.Pos++
		}
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
