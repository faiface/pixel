package supervisor

import (
	"log"

	"github.com/faiface/pixel/pixelgl"
	"github.com/gandrin/ASharedJourney/shared"
)

type Event string

//get the key values that was pressed
func catchEvent() *Event {
	var event = new(Event)
	//check if key was just pressed
	if shared.Win.Pressed(pixelgl.KeyR) {
		*event = "RESTART"
	} else if shared.Win.Pressed(pixelgl.Key0) {
		*event = "KEY0"
	} else if shared.Win.Pressed(pixelgl.Key1) {
		log.Printf("1")
		*event = "KEY1"
	} else if shared.Win.Pressed(pixelgl.Key2) {
		*event = "KEY2"
	} else if shared.Win.Pressed(pixelgl.Key3) {
		*event = "KEY3"
	} else if shared.Win.Pressed(pixelgl.Key4) {
		*event = "KEY4"
	} else if shared.Win.Pressed(pixelgl.Key5) {
		*event = "KEY5"
	} else if shared.Win.Pressed(pixelgl.Key6) {
		*event = "KEY6"
	} else if shared.Win.Pressed(pixelgl.Key7) {
		*event = "KEY7"
	} else if shared.Win.Pressed(pixelgl.Key8) {
		*event = "KEY8"
	} else if shared.Win.Pressed(pixelgl.Key9) {
		*event = "KEY9"
	}

	return event
}
