package game

import "log"

func chooseCommand(t Team, s State) command {
	h := t.Baton.Holder
	if collide(h.Pos+1, h.Lane, s) {
		return left
	}

	var nextBot *Bot
	for i, b := range t.Bots {
		if b.id == h.id+1 {
			nextBot = &t.Bots[i]
			break
		}
	}

	if nextBot != nil {
		if h.Lane != nextBot.Lane {
			if abs(nextBot.Pos-h.Pos) < h.v {
				log.Println("WHOOOOOOA")
				return slowDown
			}
		}
	}

	return speedUp
}
