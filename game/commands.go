package game

type command int

const (
	speedUp command = iota
	slowDown
	left
	right
)

func doCommand(cmd command, s State, teamID int) State {
	da := 1
	//da += rand.Intn(3) - 1

	b := ActiveBot(s.Teams[teamID])
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
		b.Position.Lane++
	case right:
		b.Position.Lane--
	}

	s = updateBot(s, teamID, *b)
	return s
}
