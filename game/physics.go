package game

import "math/rand"

func moveBot(b *Bot) {
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
	b.Pos += b.v
}

const (
	passDistance = 10
	baseAccel    = 10
)
