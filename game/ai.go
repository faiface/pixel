package game

func chooseCommand(s State, teamID int) command {
	t := s.Teams[teamID]
	h := ActiveRacer(t)
	if collide(h.Position.Pos+1, h.Position.Lane, s) != nil {
		if h.Position.Lane <= t.Lane && h.Position.Lane < NumLanes-1 {
			return left
		}
		return right
	}

	var nextRacer *Racer
	for i, b := range t.Racers {
		if b.ID == h.ID+1 {
			nextRacer = &t.Racers[i]
			break
		}
	}

	if nextRacer != nil {
		if h.Position.Lane != nextRacer.Position.Lane {
			if abs(nextRacer.Position.Pos-h.Position.Pos) < h.Kinetics.V {
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
		n := score(cmd, s, teamID, depth)
		if n > bestN {
			bestCmd, bestN = cmd, n
		}
	}

	return bestCmd
}

func score(cmd command, s State, teamID int, depth int) int {
	if !legalMove(s, teamID, cmd) {
		return -1
	}
	s = doCommand(cmd, s, teamID)
	if depth == 0 {
		t := s.Teams[teamID]
		b := ActiveRacer(t)
		if b == nil {
			return 0
		}
		return b.Position.Pos
	}

	depth--
	cmd2 := smartChooseHelper(s, teamID, depth)
	return score(cmd2, s, teamID, depth)
}
