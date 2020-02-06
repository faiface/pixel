package main

import (
	"log"
	"math/rand"
)

type state struct {
	teams    []team
	gameOver bool
}

func newState() state {
	var teams []team
	for i := 0; i < numTeams; i++ {
		var bots []bot
		for j := 0; j < numBots; j++ {
			b := bot{
				id:  i*numTeams + j,
				pos: j * (steps / numBots),
			}
			bots = append(bots, b)
		}
		teams = append(teams, team{
			bots:  bots,
			baton: baton{holder: &bots[0]},
		})
	}

	return state{
		teams: teams,
	}
}

type team struct {
	bots  []bot
	baton baton
	won   bool
}

type bot struct {
	id  int
	pos int
	v   int
	a   int
}

type baton struct {
	holder *bot
}

func updateState(sOld state) state {
	s := sOld

	for i, t := range s.teams {
		b := t.baton.holder

		if b.a == 0 {
			b.a = 10
		}
		b.a += rand.Intn(3) - 1

		b.v += b.a
		b.pos += b.v

		maybePassBaton(&s.teams[i])
	}

	for _, t := range s.teams {
		if won(*t.baton.holder, s) {
			s.gameOver = true
		}
	}

	return s
}

func maybePassBaton(t *team) {
	for i, b := range t.bots {
		h := t.baton.holder
		if h == &t.bots[i] {
			continue
		}
		if abs(b.pos-h.pos) <= 10 {
			log.Printf("pass from %v to %v!", h.id, b.id)
			t.baton.holder.v = 0
			t.baton.holder.a = 0
			t.baton.holder = &t.bots[i]
			t.bots[i].a = 10
			return
		}
	}
}

func won(b bot, s state) bool {
	return b.pos == steps
}

func gameOver(s state) bool {
	for _, t := range s.teams {
		if t.won {
			return true
		}
	}
	return false
}

const (
	steps    = 400
	numBots  = 10
	numTeams = 1
)

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
