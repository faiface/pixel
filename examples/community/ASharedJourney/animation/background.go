package animation

import (

	"log"

	"github.com/gandrin/ASharedJourney/shared"
)

//placeholder
type backgroundImage int

func (a *Animator) generateBackground(level int) [][]backgroundImage {
	//todo draw background depending on selected level
	return make([][]backgroundImage, 3, 3)
}

func (a *Animator) redrawBackGroundTile(pos shared.Position) {
	//todo refresh background position
	//log.Print("position to redraw ", pos)
}
