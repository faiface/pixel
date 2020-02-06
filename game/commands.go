package game

import "math/rand"

func doCommand(cmd command, b *Bot) {
	da := 1
	da += rand.Intn(3) - 1

	switch cmd {
	case speedUp:
		b.a += da
		accelerate(b)
	case slowDown:
		b.a -= da
		accelerate(b)
	case left:
		b.Lane++
	case right:
		b.Lane--
	}
}

type command int

const (
	speedUp command = iota
	slowDown
	left
	right
)
