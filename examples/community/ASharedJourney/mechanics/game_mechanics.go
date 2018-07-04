//defines the behavior of the game once we have defined the desired direction
package mechanics

import (
	"fmt"
	"log"
	"time"

	"github.com/gandrin/ASharedJourney/shared"
	"github.com/gandrin/ASharedJourney/supervisor"
	"github.com/gandrin/ASharedJourney/tiles"
	"github.com/gandrin/ASharedJourney/menu"
	"github.com/faiface/pixel"
	"github.com/gandrin/ASharedJourney/music"
	"image/color"
)

type Mechanics struct {
	world tiles.World
	//communication channel to animator
	toAnimate chan *tiles.World
	//communication channel from supervisor
	gameEventChannel chan *supervisor.GameEvent
}

//game mechanics
var Mecha *Mechanics

//initialise the game mechanics structure
func Start(
	gameEventChannel chan *supervisor.GameEvent,
	baseWorld tiles.World,
) chan *tiles.World {
	Mecha = new(Mechanics)
	//build return channel to animator
	var toAnim chan *tiles.World
	toAnim = make(chan *tiles.World, 1)

	Mecha.toAnimate = toAnim
	Mecha.gameEventChannel = gameEventChannel
	Mecha.world = baseWorld

	//log.Print("Mecanics loaded")
	return Mecha.toAnimate
}

//synchronisation objects
func (m *Mechanics) muxChannel() *supervisor.GameEvent {
	select {
	case nextGameEvent, ok := <-m.gameEventChannel:
		if !ok {
			fmt.Println("Channel  closed!")
			log.Fatal()
		}
		return nextGameEvent
	default:
		nextEvent := supervisor.Event("NONE")
		nextGameEvent := new(supervisor.GameEvent)
		nextGameEvent.PlayerDirections = new(supervisor.PlayerDirections)
		nextGameEvent.PlayerDirections.Player1.X = 0
		nextGameEvent.PlayerDirections.Player1.Y = 0
		nextGameEvent.PlayerDirections.Player2.X = 0
		nextGameEvent.PlayerDirections.Player2.Y = 0
		nextGameEvent.Event = &nextEvent
		fmt.Println("No player direction mecha is faster than supervisor ")
		return nextGameEvent
		//set motion to default values
	}
}

//call mechanics
func (m *Mechanics) Play() {

	for play := true; play; play = shared.Continue() {
		//delay to not call and overload cpu
		time.Sleep(shared.MechanicsRefreshDelayMs * time.Millisecond)

		gameEvent := m.muxChannel()
		m.toAnimate <- m.Move(gameEvent.PlayerDirections)
		m.handleGameEvent(gameEvent.Event)
	}
}

func (m *Mechanics) handleGameEvent(event *supervisor.Event) {
	switch *event {
	case "RESTART":
		m.world = tiles.RestartLevel()
		menu.Menu(menu.WinLevelMenuImage, "Reloading level ...", pixel.V(180, 150), true, music.SOUND_EFFECT_START_GAME)
		shared.Win.Clear(color.Black)
		break
	case "KEY0":
		tiles.SetNexLevel(0)
		m.world = tiles.NextLevel()
		break
	case "KEY1":
		tiles.SetNexLevel(1)
		m.world = tiles.NextLevel()
		break
	case "KEY2":
		tiles.SetNexLevel(2)
		m.world = tiles.NextLevel()
		break
	case "KEY3":
		tiles.SetNexLevel(3)
		m.world = tiles.NextLevel()
		break
	case "KEY4":
		tiles.SetNexLevel(4)
		m.world = tiles.NextLevel()
		break
	case "KEY5":
		tiles.SetNexLevel(5)
		m.world = tiles.NextLevel()
		break
	case "KEY6":
		tiles.SetNexLevel(6)
		m.world = tiles.NextLevel()
		break
	case "KEY7":
		tiles.SetNexLevel(7)
		m.world = tiles.NextLevel()
		break
	case "KEY8":
		tiles.SetNexLevel(8)
		m.world = tiles.NextLevel()
		break
	case "KEY9":
		tiles.SetNexLevel(9)
		m.world = tiles.NextLevel()
		break

	default:
		//No event
	}
}
