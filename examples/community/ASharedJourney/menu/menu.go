package menu

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"time"

	"github.com/gandrin/ASharedJourney/assets_manager"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/gandrin/ASharedJourney/music"
	"github.com/gandrin/ASharedJourney/shared"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

//level image names
const MainMenuImage = "menu.png"
const WinLevelMenuImage = "win.png"
const FinishedGameImage = "end.png"
const DrownedGameImage = "oops.png"
const RulesMenuImage = "splashScreen.png"

//draw menu to screen while player while player hasn't pressed enter
func Menu(pictureName string, menuText string, positionText pixel.Vec, blocking bool, exitSoundEffect music.SoundEffect) {

	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(positionText, basicAtlas)
	basicTxt.Color = colornames.White
	fmt.Fprintln(basicTxt, menuText)

	//get picture
	pic, err := loadPicture(pictureName)
	if err != nil {
		log.Fatal(err)
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())

	mat := pixel.IM
	mat = mat.Moved(shared.Win.Bounds().Center())
	imageMatrix := mat.ScaledXY(shared.Win.Bounds().Center(), pixel.V(0.7, 0.7))

	//clear background
	shared.Win.Clear(colornames.Black)
	sprite.Draw(shared.Win, imageMatrix)

	//text
	basicTxt.Draw(shared.Win, pixel.IM.Scaled(basicTxt.Orig, 3))
	shared.Win.Update()

	//menu loop
	if blocking {
		for !shared.Win.JustPressed(pixelgl.KeyEnter) && !shared.Win.Closed() {
			time.Sleep(50 * time.Millisecond)
			shared.Win.Update()
		}
		music.Music.PlayEffect(exitSoundEffect)
	}

}

func loadPicture(path string) (pixel.Picture, error) {
	byteImage, err := assetsManager.Asset("assets/" + path)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(byteImage))
	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(img), nil
}
