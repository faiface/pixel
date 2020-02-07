package game

import (
	"math/rand"
)

type command int

const (
	speedUp command = iota
	slowDown
	left
	right
)

func doCommand(cmd command, s State, sOld State, teamID int) State {
	da := 1
	da += rand.Intn(3) - 1

	b := activeBot(s.Teams[teamID])
	if b == nil {
		return s
	}

	switch cmd {
	case speedUp:
		b.a += da
		*b = accelerate(*b)
	case slowDown:
		b.a -= da
		*b = accelerate(*b)
	case left:
		b.Lane++
	case right:
		b.Lane--
	}

	return updateBot(s, sOld, teamID, *b)
}
