package game

import (
	"github.com/faiface/pixel/pixelgl"
)

type Command int

const (
	coast Command = iota
	speedUp
	slowDown
	left
	right
	clearObstacle
)

var validCommands = []Command{coast, speedUp, slowDown, left, right, clearObstacle}

func CommandLoop(w *pixelgl.Window, s State, stateCA chan<- State) {
	cmdC := make(chan []Command)
	go func() { cmdC <- pollCommands(s) }()

	stateCB := make(chan State)

	turn := 1
	sOld := s

	for !w.Closed() {
		switch {
		case w.Pressed(pixelgl.KeyQ):
			w.SetClosed(true)
			return
		case w.JustPressed(pixelgl.KeyEnter) || w.Pressed(pixelgl.KeySpace) || true:
			for i, cmd := range <-cmdC {
				s = doCommand(cmd, s, i)
			}
			s = checkWin(s, sOld)
			turn++
			if s.GameOver {
				s = NewState()
				sOld = s
				turn = 1
			}
			go func() {
				s := <-stateCB
				cmdC <- pollCommands(s)
			}()
			stateCA <- s
			stateCB <- s
		}

		w.UpdateInput()
	}
}

func doCommand(cmd Command, s State, teamID int) State {
	da := 1

	r := ActiveRacer(s.Teams[teamID])
	if r == nil {
		return s
	}
	r.Kinetics.A = 0

	switch cmd {
	case coast:
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
	case coast:
		return "coast"
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
