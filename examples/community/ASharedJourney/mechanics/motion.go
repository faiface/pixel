package mechanics

import (
	"reflect"

	"github.com/faiface/pixel"
	"github.com/gandrin/ASharedJourney/menu"
	"github.com/gandrin/ASharedJourney/music"
	"github.com/gandrin/ASharedJourney/supervisor"
	"github.com/gandrin/ASharedJourney/tiles"
)

//move function receives as input the data from a player direction channel
func (m *Mechanics) Move(playDir *supervisor.PlayerDirections) *tiles.World {

	if m.world.Players[0].HasWon && m.world.Players[1].HasWon &&
		!reflect.DeepEqual(m.world.Players[0].WinningPosition, m.world.Players[1].WinningPosition) {
		music.Music.PlayEffect(music.SOUND_EFFECT_WIN_GAME)
		m.world = tiles.NextLevel()
	}

	if m.world.Players[0].InTheWater || m.world.Players[1].InTheWater {
		menu.Menu(menu.DrownedGameImage, "Oops ....", pixel.V(300, 150), true, music.SOUND_EFFECT_LOSE_GAME)
		m.world = tiles.RestartLevel()
	}

	if playDir.Player1.X != 0 || playDir.Player1.Y != 0 {
		// "Zzzz" tile...
		m.world.Holes[len(m.world.Holes)-1].Position.X = -100
		m.world.Holes[len(m.world.Holes)-1].Position.Y = -100

		m.movePlayer(&m.world.Players[0], playDir.Player1.Next)
		m.movePlayer(&m.world.Players[1], playDir.Player2.Next)

	}

	return m.copyToNewWorld()
}

func (m *Mechanics) movePlayer(player *tiles.SpriteWithPosition, getNextPosition func(pixel.Vec) pixel.Vec) {
	var canPlayerMove = true
	nextPlayerPosition := getNextPosition(player.Position)

	/// In the hole
	if player.InTheHole {
		player.InTheHole = false
		m.world.Holes[len(m.world.Holes)-1].Position = player.Position
		return
	}

	// Obstacles
	for _, obstacle := range m.world.Obstacles {
		if obstacle.Position.X == nextPlayerPosition.X && obstacle.Position.Y == nextPlayerPosition.Y {
			canPlayerMove = false
		}
	}

	// Movables
	if canPlayerMove {
		for n, mov := range m.world.Movables {
			if mov.Position.X == nextPlayerPosition.X && mov.Position.Y == nextPlayerPosition.Y {
				// There's a movable in that position
				movableNextPosition := getNextPosition(nextPlayerPosition)
				for _, playerTile := range m.world.Players {
					if playerTile.Position.X == movableNextPosition.X &&
						playerTile.Position.Y == movableNextPosition.Y {
						canPlayerMove = false
					}
				}
				for _, obstacleTile := range m.world.Obstacles {
					if obstacleTile.Position.X == movableNextPosition.X &&
						obstacleTile.Position.Y == movableNextPosition.Y {
						canPlayerMove = false
					}
				}
				for _, movableTile := range m.world.Movables {
					if movableTile.Position.X == movableNextPosition.X &&
						movableTile.Position.Y == movableNextPosition.Y {
						canPlayerMove = false
					}
				}
				for _, winStarTile := range m.world.WinStars {
					if winStarTile.Position.X == movableNextPosition.X && winStarTile.Position.Y == movableNextPosition.Y {
						player.HasWon = true
						player.WinningPosition = pixel.V(winStarTile.Position.X, winStarTile.Position.Y)
					}
				}
				if canPlayerMove {
					m.world.Movables[n].Position = movableNextPosition
				}
				for h, holeTile := range m.world.Holes {
					if holeTile.Position.X == movableNextPosition.X && holeTile.Position.Y == movableNextPosition.Y {
						// remove both obj (hole and movable)
						m.world.Movables[n].Position.X = -100
						m.world.Holes[h].Position.X = -100
						music.Music.PlayEffect(music.SOUND_EFFECT_SNORE)
					}
				}
			}
		}
	}

	if canPlayerMove {
		player.Position = nextPlayerPosition

		// Water
		for _, waterTile := range m.world.Water {
			if waterTile.Position.X == nextPlayerPosition.X && waterTile.Position.Y == nextPlayerPosition.Y {
				player.InTheWater = true
				music.Music.PlayEffect(music.SOUND_EFFECT_WATER)
			}
		}

		// Hole
		for _, holeTile := range m.world.Holes {
			if holeTile.Position.X == nextPlayerPosition.X && holeTile.Position.Y == nextPlayerPosition.Y {
				player.InTheHole = true
				m.world.Holes[len(m.world.Holes)-1].Position = nextPlayerPosition
				music.Music.PlayEffect(music.SOUND_EFFECT_SNORE)
			}
		}

		player.HasWon = false

		// Winning rule
		for _, winStarTile := range m.world.WinStars {
			if winStarTile.Position.X == nextPlayerPosition.X && winStarTile.Position.Y == nextPlayerPosition.Y {
				player.HasWon = true
				player.WinningPosition = pixel.V(winStarTile.Position.X, winStarTile.Position.Y)
			}
		}
	}
}

func (m *Mechanics) copyToNewWorld() *tiles.World {
	var newWorld = new(tiles.World)

	//copy player locations
	//copy world
	newWorld.BackgroundTiles = make([]tiles.SpriteWithPosition, len(m.world.BackgroundTiles))
	newWorld.Movables = make([]tiles.SpriteWithPosition, len(m.world.Movables))
	newWorld.Players = make([]tiles.SpriteWithPosition, len(m.world.Players))
	newWorld.Obstacles = make([]tiles.SpriteWithPosition, len(m.world.Obstacles))
	newWorld.Water = make([]tiles.SpriteWithPosition, len(m.world.Water))
	newWorld.Holes = make([]tiles.SpriteWithPosition, len(m.world.Holes))
	newWorld.WinStars = make([]tiles.SpriteWithPosition, len(m.world.WinStars))
	copy(newWorld.BackgroundTiles, m.world.BackgroundTiles)
	copy(newWorld.Movables, m.world.Movables)
	copy(newWorld.Players, m.world.Players)
	copy(newWorld.WinStars, m.world.WinStars)
	copy(newWorld.Water, m.world.Water)
	copy(newWorld.Obstacles, m.world.Obstacles)
	copy(newWorld.Holes, m.world.Holes)
	return newWorld
}
