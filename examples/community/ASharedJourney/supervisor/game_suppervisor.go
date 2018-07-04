//retrieve player input and send the directions of motions to the mechanics
package supervisor

import (
	"time"

	"github.com/gandrin/ASharedJourney/shared"
)

type PlayerDirections struct {
	Player1 Direction
	Player2 Direction
}

type GameEvent struct {
	PlayerDirections *PlayerDirections
	Event            *Event
}

type GameSupervisor struct {
	GameEventsChannel chan *GameEvent
}

var Sup *GameSupervisor

//Start initialises the game and specify the game mode
func Start() chan *GameEvent {
	Sup = new(GameSupervisor)
	Sup.GameEventsChannel = make(chan *GameEvent, 1)
	return Sup.GameEventsChannel
}

//Play launches game supervisor (should be launched last)

func (gameSupervisor *GameSupervisor) Play() {
	var nextMove *PlayerDirections
	var nextEvent *Event
	nextGameEvent := new(GameEvent)
	for play := true; play; play = shared.Continue() {
		time.Sleep(shared.KeyPressedDelayMs * time.Millisecond)

		nextEvent = catchEvent()
		//get the players key move
		nextMove = Move()
		if nextMove.Player1.X != 0 || nextMove.Player1.Y != 0 {
			//new move
			shared.AddAction()
		}
		nextGameEvent.PlayerDirections = nextMove
		nextGameEvent.Event = nextEvent
		gameSupervisor.GameEventsChannel <- nextGameEvent
	}
}
