package game

func chooseCommand(s State, teamID int) command {
	t := s.Teams[teamID]
	h := t.BatonHolder()
	if collide(h.Position.Pos+1, h.Position.Lane, s) != nil {
		if h.Position.Lane <= t.Lane && h.Position.Lane < NumLanes-1 {
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
		if h.Position.Lane != nextBot.Position.Lane {
			if abs(nextBot.Position.Pos-h.Position.Pos) < h.v {
				return slowDown
			}
		}
	}

	return speedUp
}
