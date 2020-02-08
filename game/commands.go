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

	r := ActiveRacer(s.Teams[teamID])
	if r == nil {
		return s
	}
	r.Kinetics.A = 0

	switch cmd {
	case speedUp:
		r.Kinetics.A = da
		*r = accelerate(*r)
		s = updateRacer(s, *r)
	case slowDown:
		r.Kinetics.A = -da
		*r = accelerate(*r)
		s = updateRacer(s, *r)
	case left:
		r.Position.Lane++
		s = updateRacer(s, *r)
	case right:
		r.Position.Lane--
		s = updateRacer(s, *r)
	case clearObstacle:
		pos := r.Position
		pos.Pos++
		s = removeObstacle(s, pos)
		r.Kinetics.V = 0
		s = updateRacer(s, *r)
	}

	if r := ActiveRacer(s.Teams[teamID]); r != nil {
		s = moveRacer(s, *r)
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
