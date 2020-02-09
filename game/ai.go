package game

func chooseCommand(s State, teamID int) Command {
	return chooseCommandHelper(s, teamID, aiDepth)
}

func chooseCommandHelper(s State, teamID int, depth int) Command {
	bestCmd, bestN := speedUp, 0

	for _, cmd := range validCommands {
		n := score(cmd, s, teamID, depth)
		if n > bestN {
			bestCmd, bestN = cmd, n
		}
	}

	return bestCmd
}

func score(cmd Command, s State, teamID int, depth int) int {
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
		return b.Position.Pos*100 + b.Battery.Charge
	}

	depth--
	cmd2 := chooseCommandHelper(s, teamID, depth)
	return score(cmd2, s, teamID, depth)
}

const (
	aiDepth = 4
)
