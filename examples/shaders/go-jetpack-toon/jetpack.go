package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	"github.com/go-gl/mathgl/mgl32"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
)

// InstallShader ...
func InstallShader(w *pixelgl.Window, uAmount *float32, uResolution, uMouse *mgl32.Vec2) {
	wc := w.GetCanvas()
	wc.BindUniform("u_resolution", uResolution)
	wc.BindUniform("u_mouse", uMouse)
	wc.BindUniform("u_amount", uAmount)
	wc.SetFragmentShader(toonFragShader)
	wc.UpdateShader()
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func loadSprite(path string) (pixel.Sprite, error) {
	pic, err := loadPicture(path)
	if err != nil {
		return *pixel.NewSprite(pic, pic.Bounds()), err
	}
	sprite := pixel.NewSprite(pic, pic.Bounds())
	return *sprite, nil
}

func loadTTF(path string, size float64, origin pixel.Vec) *text.Text {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	font, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	face := truetype.NewFace(font, &truetype.Options{
		Size:              size,
		GlyphCacheEntries: 1,
	})

	atlas := text.NewAtlas(face, text.ASCII)

	txt := text.New(origin, atlas)

	return txt

}

func run() {
	// Set up window configs
	cfg := pixelgl.WindowConfig{ // Default: 1024 x 768
		Title:  "Golang Jetpack!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// Importantr variables
	var jetX, jetY, velX, velY, radians float64
	flipped := 1.0
	jetpackOn := false
	gravity := 0.6 // Default: 0.004
	jetAcc := 0.9  // Default: 0.008
	tilt := 0.01   // Default: 0.001
	whichOn := false
	onNumber := 0
	jetpackOffName := "jetpack.png"
	jetpackOn1Name := "jetpack-on.png"
	jetpackOn2Name := "jetpack-on2.png"
	camVector := win.Bounds().Center()

	var uAmount float32
	var uMouse, uResolution mgl32.Vec2

	InstallShader(win, &uAmount, &uResolution, &uMouse)

	bg, _ := loadSprite("sky.png")

	// Jetpack - Rendering
	jetpackOff, err := loadSprite(jetpackOffName)
	if err != nil {
		panic(err)
	}
	jetpackOn1, err := loadSprite(jetpackOn1Name)
	if err != nil {
		panic(err)
	}
	jetpackOn2, err := loadSprite(jetpackOn2Name)
	if err != nil {
		panic(err)
	}

	// Tutorial Text
	txt := loadTTF("intuitive.ttf", 50, pixel.V(win.Bounds().Center().X-450, win.Bounds().Center().Y-200))
	fmt.Fprintf(txt, "Explore the Skies with WASD or Arrow Keys!")

	currentSprite := jetpackOff

	canvas := pixelgl.NewCanvas(win.Bounds())
	uResolution[0] = float32(win.Bounds().W())
	uResolution[1] = float32(win.Bounds().H())
	uAmount = 300.0
	// Game Loop
	for !win.Closed() {

		mpos := win.MousePosition()
		uMouse[0] = float32(mpos.X)
		uMouse[1] = float32(mpos.Y)
		win.SetTitle(fmt.Sprint(uAmount))
		win.Clear(colornames.Blue)
		canvas.Clear(colornames.Green)

		// Jetpack - Controls
		jetpackOn = win.Pressed(pixelgl.KeyUp) || win.Pressed(pixelgl.KeyW)

		if win.Pressed(pixelgl.KeyRight) || win.Pressed(pixelgl.KeyD) {
			jetpackOn = true
			flipped = -1
			radians -= tilt
			velX += tilt * 30
		} else if win.Pressed(pixelgl.KeyLeft) || win.Pressed(pixelgl.KeyA) {
			jetpackOn = true
			flipped = 1
			radians += tilt
			velX -= tilt * 30
		} else {
			if velX < 0 {
				radians -= tilt / 3
				velX += tilt * 10
			} else if velX > 0 {
				radians += tilt / 3
				velX -= tilt * 10
			}
		}
		if jetY < 0 {
			jetY = 0
			velY = -0.3 * velY
		}

		if jetpackOn {
			velY += jetAcc
			whichOn = !whichOn
			onNumber++
			if onNumber == 5 { // every 5 frames, toggle anijetMation
				onNumber = 0
				if whichOn {
					currentSprite = jetpackOn1
				} else {
					currentSprite = jetpackOn2
				}
			}
		} else {
			currentSprite = jetpackOff
			velY -= gravity
		}

		if win.Pressed(pixelgl.KeyEqual) {
			uAmount += 10
		}
		if win.Pressed(pixelgl.KeyMinus) {
			uAmount -= 10
		}

		positionVector := pixel.V(win.Bounds().Center().X+jetX, win.Bounds().Center().Y+jetY-372)
		jetMat := pixel.IM
		jetMat = jetMat.Scaled(pixel.ZV, 4)
		jetMat = jetMat.Moved(positionVector)
		jetMat = jetMat.ScaledXY(positionVector, pixel.V(flipped, 1))
		jetMat = jetMat.Rotated(positionVector, radians)

		jetX += velX
		jetY += velY

		// Camera
		camVector.X += (positionVector.X - camVector.X) * 0.2
		camVector.Y += (positionVector.Y - camVector.Y) * 0.2

		if camVector.X > 25085 {
			camVector.X = 25085
		} else if camVector.X < -14843 {
			camVector.X = -14843
		}

		if camVector.Y > 22500 {
			camVector.Y = 22500
		}

		cam := pixel.IM.Moved(win.Bounds().Center().Sub(camVector))

		canvas.SetMatrix(cam)

		// Drawing to the screen
		win.SetSmooth(true)
		bg.Draw(canvas, pixel.IM.Moved(pixel.V(win.Bounds().Center().X, win.Bounds().Center().Y+766)).Scaled(pixel.ZV, 10))
		txt.Draw(canvas, pixel.IM)
		win.SetSmooth(false)
		currentSprite.Draw(canvas, jetMat)
		canvas.Draw(win, pixel.IM.Moved(win.Bounds().Center()))
		win.Update()
	}

}

func main() {
	pixelgl.Run(run)
}

var toonFragShader = `
#version 330 core

in vec4 Color;
in vec2 texcoords;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;
uniform float u_amount;
uniform vec2 u_mouse;
uniform vec2 u_resolution;

#define HueLevCount 6
#define SatLevCount 7
#define ValLevCount 4

float HueLevels[HueLevCount];
float SatLevels[SatLevCount];
float ValLevels[ValLevCount];

vec3 RGBtoHSV( float r, float g, float b) {
   float minv, maxv, delta;
   vec3 res;

   minv = min(min(r, g), b);
   maxv = max(max(r, g), b);
   res.z = maxv;            // v

   delta = maxv - minv;

   if( maxv != 0.0 )
      res.y = delta / maxv;      // s
   else {
      // r = g = b = 0      // s = 0, v is undefined
      res.y = 0.0;
      res.x = -1.0;
      return res;
   }

   if( r == maxv )
      res.x = ( g - b ) / delta;      // between yellow & magenta
   else if( g == maxv )
      res.x = 2.0 + ( b - r ) / delta;   // between cyan & yellow
   else
      res.x = 4.0 + ( r - g ) / delta;   // between magenta & cyan

   res.x = res.x * 60.0;            // degrees
   if( res.x < 0.0 )
      res.x = res.x + 360.0;

   return res;
}

vec3 HSVtoRGB(float h, float s, float v ) {
   int i;
   float f, p, q, t;
   vec3 res;

   if( s == 0.0 ) {
      // achromatic (grey)
      res.x = v;
      res.y = v;
      res.z = v;
      return res;
   }

   h /= 60.0;         // sector 0 to 5
   i = int(floor( h ));
   f = h - float(i);         // factorial part of h
   p = v * ( 1.0 - s );
   q = v * ( 1.0 - s * f );
   t = v * ( 1.0 - s * ( 1.0 - f ) );

   if (i==0) {
		res.x = v;
		res.y = t;
		res.z = p;
   	} else if (i==1) {
         res.x = q;
         res.y = v;
         res.z = p;
	} else if (i==2) {
         res.x = p;
         res.y = v;
         res.z = t;
	} else if (i==3) {
         res.x = p;
         res.y = q;
         res.z = v;
	} else if (i==4) {
         res.x = t;
         res.y = p;
         res.z = v;
	} else if (i==5) {
         res.x = v;
         res.y = p;
         res.z = q;
   }

   return res;
}

float nearestLevel(float col, int mode) {

   if (mode==0) {
   		for (int i =0; i<HueLevCount-1; i++ ) {
		    if (col >= HueLevels[i] && col <= HueLevels[i+1]) {
		      return HueLevels[i+1];
		    }
		}
	 }

	if (mode==1) {
		for (int i =0; i<SatLevCount-1; i++ ) {
			if (col >= SatLevels[i] && col <= SatLevels[i+1]) {
	          return SatLevels[i+1];
	        }
		}
	}


	if (mode==2) {
		for (int i =0; i<ValLevCount-1; i++ ) {
			if (col >= ValLevels[i] && col <= ValLevels[i+1]) {
	          return ValLevels[i+1];
	        }
		}
	}


}

// averaged pixel intensity from 3 color channels
float avg_intensity(vec4 pix) {
 return (pix.r + pix.g + pix.b)/3.;
}

vec4 get_pixel(vec2 coords, float dx, float dy) {
 return texture(u_texture, coords + vec2(dx, dy));
}

// returns pixel color
float IsEdge(in vec2 coords){
  float dxtex = 1.0 / u_resolution.x ;
  float dytex = 1.0 / u_resolution.y ;

  float pix[9];

  int k = -1;
  float delta;

  // read neighboring pixel intensities
float pix0 = avg_intensity(get_pixel(coords,-1.0*dxtex, -1.0*dytex));
float pix1 = avg_intensity(get_pixel(coords,-1.0*dxtex, 0.0*dytex));
float pix2 = avg_intensity(get_pixel(coords,-1.0*dxtex, 1.0*dytex));
float pix3 = avg_intensity(get_pixel(coords,0.0*dxtex, -1.0*dytex));
float pix4 = avg_intensity(get_pixel(coords,0.0*dxtex, 0.0*dytex));
float pix5 = avg_intensity(get_pixel(coords,0.0*dxtex, 1.0*dytex));
float pix6 = avg_intensity(get_pixel(coords,1.0*dxtex, -1.0*dytex));
float pix7 = avg_intensity(get_pixel(coords,1.0*dxtex, 0.0*dytex));
float pix8 = avg_intensity(get_pixel(coords,1.0*dxtex, 1.0*dytex));
  // average color differences around neighboring pixels
  delta = (abs(pix1-pix7)+
          abs(pix5-pix3) +
          abs(pix0-pix8)+
          abs(pix2-pix6)
           )/4.;

  return clamp(5.5*delta,0.0,1.0);
}

void main(void)
{
    vec2 uv = (texcoords - u_texbounds.xy) / u_texbounds.zw;

	HueLevels[0] = 0.0;
	HueLevels[1] = 80.0;
	HueLevels[2] = 160.0;
	HueLevels[3] = 240.0;
	HueLevels[4] = 320.0;
	HueLevels[5] = 360.0;

	SatLevels[0] = 0.0;
	SatLevels[1] = 0.1;
	SatLevels[2] = 0.3;
	SatLevels[3] = 0.5;
	SatLevels[4] = 0.6;
	SatLevels[5] = 0.8;
	SatLevels[6] = 1.0;

	ValLevels[0] = 0.0;
	ValLevels[1] = 0.3;
	ValLevels[2] = 0.6;
	ValLevels[3] = 1.0;

    vec4 colorOrg = texture( u_texture, uv );
    vec3 vHSV =  RGBtoHSV(colorOrg.r,colorOrg.g,colorOrg.b);
    vHSV.x = nearestLevel(vHSV.x, 0);
    vHSV.y = nearestLevel(vHSV.y, 1);
    vHSV.z = nearestLevel(vHSV.z, 2);
    float edg = IsEdge(uv);
    vec3 vRGB = (edg >= 0.3)? vec3(0.0,0.0,0.0):HSVtoRGB(vHSV.x,vHSV.y,vHSV.z);
    fragColor = vec4(vRGB.x,vRGB.y,vRGB.z,1.0);
}
`
