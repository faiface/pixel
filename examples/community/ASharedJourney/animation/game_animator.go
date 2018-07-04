//package for managing animations and refreshing screen
package animation

import (
	"fmt"
	"log"
	"time"

	"github.com/gandrin/ASharedJourney/mechanics"
	"github.com/gandrin/ASharedJourney/shared"
)

type Animator struct {
	backgound [][]backgroundImage //placeholder

	fromAnim chan *mechanics.Motion

	player1 *playerSprite
	player2 *playerSprite

	tileSet *FakeTileSet
}

var anim *Animator

func Start(motion chan *mechanics.Motion, level int, pType1 mechanics.PlayerType, pType2 mechanics.PlayerType, tileset *FakeTileSet) {
	anim = new(Animator)
	anim.fromAnim = motion

	//player animation
	anim.player1 = NewPlayerSprite(pType1)
	anim.player2 = NewPlayerSprite(pType2)

	//load tileset
	anim.tileSet = tileset

	//load background
	anim.backgound = anim.generateBackground(level)
}

func (anim *Animator) muxChannel() *mechanics.Motion {
	var nextMotion *mechanics.Motion = nil
	select {
	case m, ok := <-anim.fromAnim:
		if ok {
			nextMotion = m
		} else {
			log.Fatal("Channel closed!")
		}
	default:
		fmt.Println("No value ready, moving on.")
	}
	return nextMotion
}

//play main animation loop
func (anim *Animator) Play() {
	for play := true; play; play = shared.Continue() {
		time.Sleep(shared.FrameRefreshDelayMs * time.Millisecond)
		motion := anim.muxChannel()
		anim.animate(motion)
	}
}

func (anim *Animator) animate(motion *mechanics.Motion) {

	//check if motion called
	if motion != nil {
		//move players and events
		//log.Printf("Move players")
	}

	//move all animations
	//log.Printf("animting")
}
