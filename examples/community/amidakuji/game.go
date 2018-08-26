package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"
	"unsafe"

	gg "github.com/faiface/pixel/examples/community/amidakuji/glossary"
	"github.com/faiface/pixel/examples/community/amidakuji/glossary/jukebox"
	glfw "github.com/go-gl/glfw/v3.2/glfw"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/sqweek/dialog"
	"golang.org/x/image/colornames"
)

// Actor updates and draws itself. It acts as a game object.
type Actor interface {
	Drawer
	Updater
}

// Drawer draws itself.
type Drawer interface {
	Draw()
}

// Updater updates itself.
type Updater interface {
	Update()
}

// -------------------------------------------------------------------------
// Core game

// game is a path finder.
// Also it manages and draws everything about...
type game struct {
	// something system, somthing runtime
	window *pixelgl.Window // lazy init
	bg     pixel.RGBA
	camera *gg.Camera // lazy init
	fpsw   *gg.FPSWatch
	dtw    gg.DtWatch
	vsync  <-chan time.Time // lazy init
	// game state
	isRefreshedLadder   bool
	isRefreshedNametags bool
	isScalpelMode       bool
	// drawings
	mutex         sync.Mutex // It is unsafe to access any refd; ptrd object without a critical section.
	nPlayers      int
	ladder        *Ladder
	scalpel       *Scalpel
	paths         []Path
	emojis        []pixel.Sprite
	nametagPicks  Nametags
	nametagPrizes Nametags
	atlas         *text.Atlas
	galaxy        *gg.Galaxy
	explosions    *gg.Explosions
	// other user settings
	fontSize            float64
	winWidth            float64 // The screen width, not the game width.
	winHeight           float64
	initialZoomLevel    float64
	initialRotateDegree float64
}

type gameConfig struct {
	nParticipants       int
	nLevel              int
	winWidth            float64
	winHeight           float64
	width               float64
	height              float64
	initialZoomLevel    float64
	initialRotateDegree float64
	paddingTop          float64
	paddingRight        float64
	paddingBottom       float64
	paddingLeft         float64
	fontSize            float64
	nametagPicks        []string
	nametagPrizes       []string
}

// init game
func newGame(cfg gameConfig) *game {

	newEmojis := func(nParticipants int) (emojis []pixel.Sprite) {
		emojis = make([]pixel.Sprite, nParticipants)
		const dir = "emoji"
		randomNames, err := gg.AssetDir(dir) // The order is random because they're from a map.
		if err != nil {
			return nil
		}
		nRandomNames := len(randomNames)
		for participant := 0; participant < nParticipants; participant++ {
			emojis[participant] = *gg.NewSprite(dir + "/" + randomNames[participant%nRandomNames]) // val, not ptr
		}
		return emojis
	}

	g := game{
		bg:                  gg.RandomNiceColor(),
		fpsw:                gg.NewFPSWatchSimple(pixel.V(cfg.winWidth, cfg.winHeight), gg.Top, gg.Right),
		isRefreshedLadder:   false,
		isRefreshedNametags: false,
		isScalpelMode:       false,
		nPlayers:            cfg.nParticipants,
		ladder: NewLadder(
			cfg.nParticipants, cfg.nLevel,
			cfg.width, cfg.height,
			cfg.paddingTop, cfg.paddingRight,
			cfg.paddingBottom, cfg.paddingLeft,
		),
		scalpel:       &Scalpel{},
		paths:         make([]Path, cfg.nParticipants),
		emojis:        newEmojis(cfg.nParticipants),
		nametagPicks:  make([]Nametag, cfg.nParticipants), // val, not ptr
		nametagPrizes: make([]Nametag, cfg.nParticipants), // val, not ptr
		atlas: gg.NewAtlas(
			"", cfg.fontSize,
			[]rune(strings.Join(cfg.nametagPicks, "")+strings.Join(cfg.nametagPrizes, "")),
		), // A prepared set of images of characters or symbols to be drawn.
		galaxy:              gg.NewGalaxy(cfg.width, cfg.height, 400),
		explosions:          gg.NewExplosions(cfg.width, cfg.width, nil, 5),
		initialZoomLevel:    cfg.initialZoomLevel,
		initialRotateDegree: cfg.initialRotateDegree,
		winWidth:            cfg.winWidth,
		winHeight:           cfg.winHeight,
	}

	// init paths
	g.ResetPaths()

	// copy nametags
	copyNametagPicks := func(dstNametags []Nametag, srcNames []string) {
		positions := g.ladder.PtsAtLevelOfPicks()
		for i := 0; i < cfg.nParticipants; i++ {
			posAdjust := positions[i]
			posAdjust.Y += 5
			posAdjust.X -= 60
			dstNametags[i] = *NewNametagSimple(
				g.atlas, "", posAdjust,
				gg.Middle, gg.Right,
			) // val, not ptr
			if i < len(srcNames) {
				dstNametags[i].desc = srcNames[i]
			}
		}
	}
	copyNametagPrizes := func(dstNametags []Nametag, srcNames []string) {
		positions := g.ladder.PtsAtLevelOfPrizes()
		for i := 0; i < cfg.nParticipants; i++ {
			posAdjust := positions[i]
			posAdjust.Y += 5
			posAdjust.X += 15
			dstNametags[i] = *NewNametagSimple(
				g.atlas, "", posAdjust,
				gg.Middle, gg.Left,
			) // val, not ptr
			if i < len(srcNames) {
				dstNametags[i].desc = srcNames[i]
			}
		}
	}
	copyNametagPicks(g.nametagPicks, cfg.nametagPicks)
	copyNametagPrizes(g.nametagPrizes, cfg.nametagPrizes)
	// log.Println(g.nametagPicks[1].desc) //

	return &g
}

func (g *game) Draw() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// This was originally an argument of this function.
	var t pixel.BasicTarget
	t = g.window

	// ---------------------------------------------------
	// 1. canvas a game world
	t.SetMatrix(g.camera.Transform())

	// Draw()s in an order.
	g.galaxy.Draw(t)
	g.ladder.Draw(t)
	for iPath := range g.paths {
		g.paths[iPath].Draw(t)
	}
	g.nametagPicks.Draw(t)
	g.nametagPrizes.Draw(t)
	if g.explosions.IsExploding() {
		g.explosions.Draw(t)
	}
	for iEmoji := range g.emojis {
		g.emojis[iEmoji].Draw(
			t, pixel.IM.
				Scaled(pixel.ZV, 2).
				Rotated(pixel.ZV, -g.camera.Angle()).
				Moved(g.paths[iEmoji].PosTip()),
		)
	}
	if g.isScalpelMode {
		g.scalpel.Draw(t)
	}
	if g.isScalpelMode {
		UpdateDrawUnprojekt(g.window, g.ladder.bound, colornames.Blue, g.camera.Transform())
		UpdateDrawUnprojekt2(g.window, g.ladder.bound, colornames.Red, *g.camera)
	}

	// ---------------------------------------------------
	// 2. canvas a screen
	t.SetMatrix(pixel.IM)

	// Draw()s in an order.
	g.fpsw.Draw(g.window)
	if g.isScalpelMode {
		UpdateDrawProjekt(g.window, g.ladder.bound, colornames.Black, g.camera.Transform())
	}
}

func (g *game) Update(dt float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// The camera would and should update every frame.
	g.camera.Update(dt)

	// Update only if there is a need.
	// isRefreshedLadder be set to false if there was an update to the ladder or its scalpel.
	if !g.isRefreshedLadder {
		g.ladder.Update()
		g.scalpel.Update(*g.ladder)
		g.isRefreshedLadder = true
	}

	// Only update when there is a need.
	if !g.isRefreshedNametags {
		g.nametagPicks.Update()
		g.nametagPrizes.Update()
		g.isRefreshedNametags = true
	}

	// Only currently animating paths need to update each frame.
	for iPath := range g.paths {
		if g.paths[iPath].IsAnimating() {
			g.paths[iPath].Update(g.ladder.colors[iPath])
		}
	}

	// As long as it doesn't hurt the framerate.
	if g.fpsw.GetFPS() >= 10 {
		g.galaxy.Update(dt)
	}

	// Only update when there is at least one (animating) explosion.
	if g.explosions.IsExploding() {
		g.explosions.Update(dt)
	}
}

func (g *game) OnResize(width, height float64) {
	g.camera.SetScreenBound(pixel.R(0, 0, width, height))
	g.fpsw.SetPos(pixel.V(width, height), gg.Top, gg.Right)
	// g.explosions.SetBound(width, height)
}

// -------------------------------------------------------------------------
// Single path

// ClearPath of a participant.
func (g *game) ClearPath(participant int) {
	g.paths[participant] = *NewPathEmpty()
}

// ResetPath of a participant.
func (g *game) ResetPath(participant int) {
	// GeneratePath contains a path-finding algorithm. This function is used as a path finder.
	GeneratePath := func(g *game, participant int) Path {
		const icol int = 0  // level
		irow := participant // participant
		grid := g.ladder.grid
		route := []pixel.Vec{}
		prize := -1
		for level := icol; level < g.ladder.nLevel; level++ {
			route = append(route, grid[irow][level])
			prize = irow
			if irow+1 < g.ladder.nParticipants {
				if g.ladder.bridges[irow][level] {
					irow++ // cross the bridge ... to the left (south)
					route = append(route, grid[irow][level])
					prize = irow
					continue
				}
			}
			if irow-1 >= 0 {
				if g.ladder.bridges[irow-1][level] {
					irow-- // cross the bridge ... to the right (north)
					route = append(route, grid[irow][level])
					prize = irow
					continue
				}
			}
		}
		// log.Println(participant, prize, irow) //

		// A path found here is called a route or roads.
		return *NewPath(route, &prize) // val, not ptr
	}

	g.paths[participant] = GeneratePath(g, participant) // path-find
	g.paths[participant].OnPassedEachPoint = func(pt pixel.Vec, dir pixel.Vec) {
		g.explosions.ExplodeAt(pt, dir.Scaled(2))
	}
}

func (g *game) AnimatePath(participant int) {
	g.paths[participant].Animate()
}

func (g *game) AnimatePathInTime(participant int, sec float64) {
	g.paths[participant].AnimateInTime(sec)
}

// -------------------------------------------------------------------------
// All paths

func (g *game) ResetPaths() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for participant := 0; participant < g.nPlayers; participant++ {
		g.ResetPath(participant)
	}
}

// AnimatePaths in order.
func (g *game) AnimatePaths(thunkAnimatePath func(participant int)) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	for participant := 0; participant < g.nPlayers; participant++ {
		participantCurr := participant
		participantNext := participant + 1
		prize := g.paths[participantCurr].GetPrize()
		title := "Result"
		caption := fmt.Sprint(
			" 👆 Pick\t(No. ", participantCurr+1, ")\t", g.nametagPicks[participantCurr], "\t",
			"\r\n", "\r\n",
			" 🎁 Prize\t(No. ", prize+1, ")\t", g.nametagPrizes[prize], "\t",
			"\r\n",
		)
		g.paths[participant].OnFinishedAnimation = func() {
			if g.window.Monitor() == nil {
				dialog.Message("%s", caption).Title(title).Info()
			}
			if participantNext < g.nPlayers {
				thunkAnimatePath(participantNext)
			}
			g.paths[participantCurr].OnFinishedAnimation = nil
		}
	}
	thunkAnimatePath(0)
}

// -------------------------------------------------------------------------
// Game controls

func (g *game) Reset() {
	g.ladder.Reset()
	g.ResetPaths()
	g.isRefreshedLadder = false
}

// Shuffle in an approximate time.
func (g *game) Shuffle(times, inMillisecond int) {
	speed := g.galaxy.Speed()
	g.galaxy.SetSpeed(speed * 10)
	{
		i := 0
		for range time.Tick(
			(time.Millisecond * time.Duration(inMillisecond)) / time.Duration(times),
		) {
			g.bg = gg.RandomNiceColor()
			g.Reset()
			i++
			if i >= times {
				break
			}
		}
	}
	g.galaxy.SetSpeed(speed)
}

// Pause the game.
func (g *game) Pause() {
	for i := range g.paths {
		if g.paths[i].IsAnimating() {
			g.paths[i].Pause()
		}
	}
}

// Resume after pause.
func (g *game) Resume() {
	g.dtw.Dt()
	for i := range g.paths {
		if g.paths[i].IsAnimating() {
			g.paths[i].Resume()
		}
	}
}

func (g *game) SetFullScreenMode(on bool) {
	if on {
		monitor := pixelgl.PrimaryMonitor()
		width, height := monitor.Size()
		// log.Println(monitor.VideoModes()) //
		g.window.SetMonitor(monitor)
		go func(width, height float64) {
			g.OnResize(width, height)
		}(width, height)
	} else if !on { // off
		g.window.SetMonitor(nil)
	} else {
		panic(errors.New("it may be thread"))
	}
}

// -------------------------------------------------------------------------
// Read only methods

// WindowDeep is a hacky way to access a window in deep.
// It returns (window *glfw.Window) which is an unexported member inside a (*pixelgl.Window).
// Read only argument game ignores the pass lock by value warning.
func (g game) WindowDeep() (baseWindow *glfw.Window) {
	return *(**glfw.Window)(unsafe.Pointer(reflect.Indirect(reflect.ValueOf(g.window)).FieldByName("window").UnsafeAddr()))
}

// Read only argument game ignores the pass lock by value warning.
func (g game) BridgesCount() (sum int) {
	for _, row := range g.ladder.bridges {
		for _, col := range row {
			if col {
				sum++
			}
		}
	}
	return sum
}

// -------------------------------------------------------------------------
// Run on main thread

// Run the game window and its event loop on main thread.
func (g *game) Run() {
	pixelgl.Run(func() {
		g.RunLazyInit()
		g.RunEventLoop()
	})
}

func (g *game) RunLazyInit() {
	// This window will show up as soon as it is created.
	win, err := pixelgl.NewWindow(pixelgl.WindowConfig{
		Title:     title + "  (" + version + ")",
		Icon:      nil,
		Bounds:    pixel.R(0, 0, g.winWidth, g.winHeight),
		Monitor:   nil,
		Resizable: true,
		// Undecorated: true,
		VSync: false,
	})
	if err != nil {
		panic(err)
	}
	win.SetSmooth(true)

	MoveWindowToCenterOfPrimaryMonitor := func(win *pixelgl.Window) {
		vmodes := pixelgl.PrimaryMonitor().VideoModes()
		vmodesLast := vmodes[len(vmodes)-1]
		biggestResolution := pixel.R(0, 0, float64(vmodesLast.Width), float64(vmodesLast.Height))
		win.SetPos(biggestResolution.Center().Sub(win.Bounds().Center()))
	}
	MoveWindowToCenterOfPrimaryMonitor(win)

	// lazy init vars
	g.window = win
	g.camera = gg.NewCamera(g.ladder.bound.Center(), g.window.Bounds())

	// register callback
	windowGL := g.WindowDeep()
	windowGL.SetSizeCallback(func(_ *glfw.Window, width int, height int) {
		g.OnResize(float64(width), float64(height))
	})

	// time manager
	g.vsync = time.Tick(time.Second / 120)
	g.fpsw.Start()
	g.dtw.Start()

	// so-called loading
	{
		g.window.Clear(colornames.Brown)
		screenCenter := g.window.Bounds().Center()
		txt := text.New(screenCenter, gg.NewAtlas("", 36, nil))
		txt.WriteString("Loading...")
		txt.Draw(g.window, pixel.IM)
		g.window.Update()
	}
	g.NextFrame(g.dtw.Dt()) // Give it a blood pressure.
	g.NextFrame(g.dtw.Dt()) // Now the oxygenated blood will start to pump through its vein.
	// Do whatever you want after that...

	// from user setting
	g.camera.Zoom(float64(g.initialZoomLevel))
	g.camera.Rotate(g.initialRotateDegree)
}

func (g *game) RunEventLoop() {
	for g.window.Closed() != true { // Your average event loop in main thread.
		// Notice that all function calls as go routine are non-blocking, but the others will block the main thread.

		// ---------------------------------------------------
		// 0. dt
		dt := g.dtw.Dt()

		// ---------------------------------------------------
		// 1. handling events
		g.HandlingEvents(dt)

		// ---------------------------------------------------
		// 2. move on
		g.NextFrame(dt)

		// log.Println(g.window.Closed()) //

	} // for
} // func

func (g *game) HandlingEvents(dt float64) {
	// Notice that all function calls as go routine are non-blocking, but the others will block the main thread.

	// system
	if g.window.JustReleased(pixelgl.KeyEscape) {
		g.window.SetClosed(true)
	}
	if g.window.JustReleased(pixelgl.KeySpace) {
		g.Pause()
		dialog.Message("%s", "Pause").Title("PPAP").Info()
		g.Resume()
	}
	if g.window.JustReleased(pixelgl.KeyTab) {
		if g.window.Monitor() == nil {
			g.SetFullScreenMode(true)
		} else {
			g.SetFullScreenMode(false)
		}
	}

	// scalpel mode
	if g.window.JustReleased(pixelgl.MouseButtonRight) {
		go func() {
			g.isScalpelMode = !g.isScalpelMode
		}()
	}
	if g.window.JustReleased(pixelgl.MouseButtonLeft) {
		// ---------------------------------------------------
		if !jukebox.IsPlaying() {
			jukebox.Play()
		}

		// ---------------------------------------------------
		posWin := g.window.MousePosition()
		posGame := g.camera.Unproject(posWin)
		go func() {
			g.explosions.ExplodeAt(pixel.V(posGame.X, posGame.Y), pixel.V(10, 10))
		}()

		// ---------------------------------------------------
		if g.isScalpelMode {
			// strTitle := fmt.Sprint(posGame.X, ", ", posGame.Y) //
			strDlg := fmt.Sprint(
				"number of bridges: ", g.BridgesCount(), "\r\n", "\r\n",
				"camera angle in degree: ", (g.camera.Angle()/math.Pi)*180, "\r\n", "\r\n",
				"camera coordinates: ", g.camera.XY().X, g.camera.XY().Y, "\r\n", "\r\n",
				"game clock: ", g.dtw.GetTimeStarted(), "\r\n", "\r\n",
				"starfield speed: ", g.galaxy.Speed(), "\r\n", "\r\n",
				"mouse click coords in screen pos: ", posWin.X, posWin.Y, "\r\n", "\r\n",
				"mouse click coords in game pos: ", posGame.X, posGame.Y,
			)
			go func() {
				// g.window.SetTitle(strTitle) //
				dialog.Message("%s", strDlg).Title("MouseButtonLeft").Info()
			}()
		}
	}

	// game ctrl
	if g.window.JustReleased(pixelgl.Key1) { // shuffle
		go func() {
			g.Shuffle(10, 750)
		}()
	}
	if g.window.JustReleased(pixelgl.Key2) { // find path slow
		go func() {
			g.ResetPaths()
			g.AnimatePaths(g.AnimatePath)
		}()
	}
	if g.window.JustReleased(pixelgl.Key3) { // find path fast
		go func() {
			g.ResetPaths()
			g.AnimatePaths(func(participant int) {
				g.AnimatePathInTime(participant, 1)
			})
		}()
	}

	// camera
	if g.window.JustReleased(pixelgl.KeyEnter) {
		go func() {
			g.camera.Rotate(-90)
		}()
	}
	if g.window.Pressed(pixelgl.KeyRight) {
		go func(dt float64) { // This camera will go diagonal while the case is in middle of rotating the camera.
			g.camera.Move(pixel.V(1000*dt, 0).Rotated(-g.camera.Angle()))
		}(dt)
	}
	if g.window.Pressed(pixelgl.KeyLeft) {
		go func(dt float64) {
			g.camera.Move(pixel.V(-1000*dt, 0).Rotated(-g.camera.Angle()))
		}(dt)
	}
	if g.window.Pressed(pixelgl.KeyUp) {
		go func(dt float64) {
			g.camera.Move(pixel.V(0, 1000*dt).Rotated(-g.camera.Angle()))
		}(dt)
	}
	if g.window.Pressed(pixelgl.KeyDown) {
		go func(dt float64) {
			g.camera.Move(pixel.V(0, -1000*dt).Rotated(-g.camera.Angle()))
		}(dt)
	}
	{ // if scrolled
		zoomLevel := g.window.MouseScroll().Y
		go func() {
			g.camera.Zoom(zoomLevel)
		}()
	}
}

func (g *game) NextFrame(dt float64) {
	// ---------------------------------------------------
	// 1. update - calc state of game objects each frame
	g.Update(dt)
	g.fpsw.Poll()

	// ---------------------------------------------------
	// 2. draw on window
	g.window.Clear(g.bg) // clear canvas
	g.Draw()             // then draw

	// ---------------------------------------------------
	// 3. update window - always end with it
	g.window.Update()
	<-g.vsync
}

// -------------------------------------------------------------------------
// On compile

const title = "AMIDA KUJI"

var version = "undefined"

// -------------------------------------------------------------------------
// Entry point

func main() {
	defer func() {
		err := jukebox.Finalize()
		if err != nil {
			log.Fatal(err)
		}
	}()
	rand.Seed(time.Now().UnixNano())

	conf := askConf()
	if conf == nil {
		conf = map[string]interface{}{
			"window_width":  800.0,
			"window_height": 800.0,
			"max_player":    10.0,
			"max_level":     100.0,
			"width":         1500.0,
			"height":        1500.0,
			"zoom":          -4.0,
			"rotate_degree": 270.0,
			"margin_top":    50.0,
			"margin_right":  100.0,
			"margin_bottom": 50.0,
			"margin_left":   200.0,
			"font_size":     28.0,
			"picks":         []interface{}{"Bulbasaur", "Ivysaur", "Venusaur", "Charmander", "Charmeleon", "Charizard", "Squirtle", "Wartortle", "Blastoise", "Caterpie", "Metapod", "Butterfree", "Weedle", "Kakuna", "Beedrill", "Pidgey", "Pidgeotto", "Pidgeot", "Rattata"},
			"prizes":        []interface{}{"TM88", "TM89", "TM90", "TM91", "TM92", "HM01", "HM02", "HM03", "HM04", "HM05", "HM06"},
		}
	}

	newGame(gameConfig{
		winWidth:            conf["window_width"].(float64),
		winHeight:           conf["window_height"].(float64),
		nParticipants:       int(conf["max_player"].(float64)),
		nLevel:              int(conf["max_level"].(float64)),
		width:               conf["width"].(float64),
		height:              conf["height"].(float64),
		initialZoomLevel:    conf["zoom"].(float64),
		initialRotateDegree: conf["rotate_degree"].(float64),
		paddingTop:          conf["margin_top"].(float64),
		paddingRight:        conf["margin_right"].(float64),
		paddingBottom:       conf["margin_bottom"].(float64),
		paddingLeft:         conf["margin_left"].(float64),
		fontSize:            conf["font_size"].(float64),
		nametagPicks:        gg.ItfsToStrs(conf["picks"].([]interface{})),
		nametagPrizes:       gg.ItfsToStrs(conf["prizes"].([]interface{})),
	}).Run()
}

func askConf() (conf map[string]interface{}) {
	for { // Load JSON
		cwd, _ := os.Getwd()
		filepath, err := dialog.File().Title("Load User Settings").
			Filter("JSON Format (*.json)", "json").
			Filter("All Files (*.*)", "*").
			SetStartDir(cwd).Load()
		if err != nil {
			if err.Error() == "Cancelled" {
				conf = nil
				break
			}
			dialog.Message("%s", "Invalid file path."+"\r\n"+"\r\n"+fmt.Sprint(err)).Title("Failed to load JSON").Error()
			continue
		}
		bytes, err := ioutil.ReadFile(filepath)
		if err != nil {
			dialog.Message("%s", "Could not read the file."+"\r\n"+"\r\n"+fmt.Sprint(err)).Title("Failed to load JSON").Error()
			continue
		}
		err = json.Unmarshal(bytes, &conf)
		if err != nil {
			dialog.Message("%s", "The file is not valid JSON format."+"\r\n"+"\r\n"+fmt.Sprint(err)).Title("Failed to load JSON").Error()
			continue
		}
		break
	}
	return
}
