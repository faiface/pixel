package game

func chooseCommand(s State, teamID int) command {
	t := s.Teams[teamID]
	h := t.BatonHolder()
	if collide(h.Pos+1, h.Lane, s) {
		if h.Lane <= t.Lane {
			return left
		}
		return right
	}

	var nextBot *Bot
	for i, b := range t.Bots {
		if b.ID == h.ID+1 {
			nextBot = &t.Bots[i]
			break
		}
	}

	if nextBot != nil {
		if h.Lane != nextBot.Lane {
			if abs(nextBot.Pos-h.Pos) < h.v {
				return slowDown
			}
		}
	}

	return speedUp
}
