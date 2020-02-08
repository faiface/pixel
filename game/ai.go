package game

func chooseCommand(s State, teamID int) command {
	t := s.Teams[teamID]
	h := ActiveBot(t)
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

func smartChooseCommand(s State, teamID int) command {
	return smartChooseHelper(s, teamID, 2)
}

func smartChooseHelper(s State, teamID int, depth int) command {
	bestCmd, bestN := speedUp, 0

	for _, cmd := range validCommands {
		if !legalMove(s, teamID, cmd) {
			continue
		}
		ss := doCommand(cmd, s, teamID)
		n := score(ss, teamID, depth)
		if n > bestN {
			bestCmd, bestN = cmd, n
		}
	}

	return bestCmd
}

func score(s State, teamID int, depth int) int {
	if depth == 0 {
		t := s.Teams[teamID]
		b := ActiveBot(t)
		if b == nil {
			return 0
		}
		return b.Position.Pos
	}

	cmd := smartChooseHelper(s, teamID, depth-1)
	ss := doCommand(cmd, s, teamID)
	return score(ss, teamID, depth-1)
}
