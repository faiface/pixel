package game

type command int

const (
	speedUp command = iota
	slowDown
	left
	right
	clearObstacle
)

var validCommands = []command{speedUp, slowDown, left, right, clearObstacle}

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
		s = updateBot(s, *b)
	case slowDown:
		b.a -= da
		*b = accelerate(*b)
		s = updateBot(s, *b)
	case left:
		b.Position.Lane++
		s = updateBot(s, *b)
	case right:
		b.Position.Lane--
		s = updateBot(s, *b)
	case clearObstacle:
		pos := b.Position
		pos.Pos++
		s = removeObstacle(s, pos)
	}

	if b := ActiveBot(s.Teams[teamID]); b != nil {
		s = moveBot(s, *b)
	}
	s = maybePassBaton(s, teamID)

	return s
}

func (c command) String() string {
	switch c {
	case speedUp:
		return "speed up"
	case slowDown:
		return "slow down"
	case left:
		return "go left"
	case right:
		return "go right"
	case clearObstacle:
		return "clear obstacle"
	}
	return "(unknown)"
}
