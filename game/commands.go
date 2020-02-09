package game

import (
	"log"
)

type Command int

const (
	speedUp Command = iota
	slowDown
	left
	right
	clearObstacle
)

var validCommands = []Command{speedUp, slowDown, left, right, clearObstacle}

func PollCommands(s State) []Command {
	cmds := make([]Command, len(s.Teams))
	for i := range s.Teams {
		cmd := chooseCommand(s, i)
		log.Printf("team %d chose to %v", i, cmd)
		cmds[i] = cmd
	}
	return cmds
}

func doCommand(cmd Command, s State, teamID int) State {
	da := 1

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

func (c Command) String() string {
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
