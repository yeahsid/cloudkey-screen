package display

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/yeahsid/cloudkey-screen/fonts"
)

// Colors from Black to White
var colors = []color.Gray{
	color.Gray{0x00},
	color.Gray{0x11},
	color.Gray{0x22},
	color.Gray{0x33},
	color.Gray{0x44},
	color.Gray{0x55},
	color.Gray{0x66},
	color.Gray{0x77},
	color.Gray{0x88},
	color.Gray{0x99},
	color.Gray{0xaa},
	color.Gray{0xbb},
	color.Gray{0xcc},
	color.Gray{0xdd},
	color.Gray{0xee},
	color.Gray{0xff},
}

// Increase Fades Out, Decrease Faces In
// No need to fade EVERY step
var fades = []color.Alpha{
	color.Alpha{0xff},
	// color.Alpha{0xee},
	// color.Alpha{0xdd},
	color.Alpha{0xcc},
	// color.Alpha{0xbb},
	// color.Alpha{0xaa},
	color.Alpha{0x99},
	// color.Alpha{0x88},
	// color.Alpha{0x77},
	color.Alpha{0x66},
	// color.Alpha{0x55},
	// color.Alpha{0x44},
	color.Alpha{0x33},
	// color.Alpha{0x22},
	// color.Alpha{0x11},
	color.Alpha{0x00},
}

// clearScreen clears... the... screen
func clearScreen() {
	draw.Draw(fb, fb.Bounds(), image.NewUniform(color.Gray{0}), image.ZP, draw.Src)
}

// colorTest the screen for funsies
func colorTest() {
	for x := range colors {
		fmt.Printf("%d\r", x)
		draw.Draw(fb, fb.Bounds(), image.NewUniform(colors[x]), image.ZP, draw.Src)
		time.Sleep(32 * time.Millisecond)
	}
	for x := range colors {
		fmt.Printf("%d\r", x)
		draw.Draw(fb, fb.Bounds(), image.NewUniform(colors[len(colors)-1-x]), image.ZP, draw.Src)
		time.Sleep(32 * time.Millisecond)
	}
}

// Fade the current screen, in or out (default out)
func fade(inverse bool) {
	capture := image.NewGray(fb.Bounds())
	draw.Draw(capture, capture.Bounds(), fb, image.ZP, draw.Src)

	for x := range colors {
		y := x
		if inverse {
			y = len(fades) - 1 - x
		}

		fmt.Printf("%d\r", y)

		bg := image.NewGray(fb.Bounds())
		draw.Draw(bg, bg.Bounds(), image.NewUniform(color.Gray{0}), image.ZP, draw.Src)
		draw.DrawMask(bg, bg.Bounds(), capture, image.ZP, image.NewUniform(fades[y]), image.ZP, draw.Over)

		// Put it on the RITZ!
		draw.Draw(fb, fb.Bounds(), bg, image.ZP, draw.Over)
		time.Sleep(8 * time.Millisecond)
	}
}

// startFadeCarousel Fast and smooth (default)
func startFadeCarousel(delay float64) {
	for {
		for s := range screens {
			capture := image.NewGray(fb.Bounds())
			draw.Draw(capture, capture.Bounds(), fb, image.ZP, draw.Src)
			// Fade Old Screen Out
			for x := range fades {
				bg := image.NewGray(fb.Bounds())
				draw.Draw(bg, bg.Bounds(), image.NewUniform(color.Gray{0}), image.ZP, draw.Src)
				draw.DrawMask(bg, bg.Bounds(), capture, image.ZP, image.NewUniform(fades[x]), image.ZP, draw.Over)
				draw.Draw(fb, fb.Bounds(), bg, image.ZP, draw.Over)
				time.Sleep(8 * time.Millisecond)
			}

			// Fade New Screen In
			for x := len(fades) - 1; x >= 0; x-- {
				bg := image.NewGray(fb.Bounds())
				draw.Draw(bg, bg.Bounds(), image.NewUniform(color.Gray{0}), image.ZP, draw.Src)
				draw.DrawMask(bg, bg.Bounds(), screens[s], image.ZP, image.NewUniform(fades[x]), image.ZP, draw.Over)
				draw.Draw(fb, fb.Bounds(), bg, image.ZP, draw.Over)
				time.Sleep(8 * time.Millisecond)
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// startXCarousel Very slow and CPU intensive on arm
func startXCarousel(delay float64) {
	capture := image.NewGray(fb.Bounds())
	for i := 0; i < 2; i++ {
		for s := range screens {
			for x := fb.Bounds().Max.X; x > -1; x-- {
				// Offset current framebuffer 1 pixel to the left (slide out)
				draw.Draw(capture, image.Rect(-1, 0, -1+screens[s].Bounds().Max.X, screens[s].Bounds().Max.Y), fb, image.ZP, draw.Src)

				// Print new screen directly on the capture as it slides out
				draw.Draw(capture, image.Rect(x, 0, x+screens[s].Bounds().Max.X, screens[s].Bounds().Max.Y), screens[s], image.ZP, draw.Src)

				// Send it all to the framebuffer
				draw.Draw(fb, fb.Bounds(), capture, image.ZP, draw.Over)
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// startYCarousel slow and cpu intensive in bursts on arm
func startYCarousel(delay float64) {
	capture := image.NewGray(fb.Bounds())
	for i := 0; i < 2; i++ {
		for s := range screens {
			for y := fb.Bounds().Max.Y; y > -1; y-- {
				// Offset current framebuffer 1 pixel to the left (slide out)
				draw.Draw(capture, image.Rect(0, -1, screens[s].Bounds().Max.X, -1+screens[s].Bounds().Max.Y), fb, image.ZP, draw.Src)

				// Print new screen directly on the capture as it slides out
				draw.Draw(capture, image.Rect(0, y, screens[s].Bounds().Max.X, y+screens[s].Bounds().Max.Y), screens[s], image.ZP, draw.Src)

				// Send it all to the framebuffer
				draw.Draw(fb, fb.Bounds(), capture, image.ZP, draw.Over)
			}
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
}

// Write draws text to a x,y coordinate on the image
func write(screen draw.Image, text string, x, y int, size float64, fontname string) {
	font := fonts.Load(fontname)
	// Setup new context
	c := freetype.NewContext()
	c.SetFont(font)            // Set the font
	c.SetFontSize(size)        // Set font size
	c.SetDPI(72)               // Fixed DPI
	c.SetClip(screen.Bounds()) // Clip the text?
	c.SetDst(screen)           // Send it where?
	c.SetSrc(image.White)      // Color of Foreground

	_, err := c.DrawString(text, freetype.Pt(x, y+int(c.PointToFixed(math.Round(float64(size)+1))>>6))) // y is center of line, shift to top of line
	if err != nil {
		log.Println(err)
		return
	}
}

func center(screen draw.Image, text string, x, y int, size float64, fontname string) {
	font := fonts.Load(fontname)
	// Setup new context
	c := freetype.NewContext()
	c.SetFont(font)        // Set the font
	c.SetFontSize(size)    // Set font size
	c.SetDPI(72)           // Fixed DPI
	c.SetClip(fb.Bounds()) // Clip the text?
	c.SetDst(fb)           // Send it where?
	c.SetSrc(image.White)  // Color of Foreground

	// Truetype stuff
	opts := truetype.Options{}
	opts.DPI = 72
	opts.Size = size + 1
	face := truetype.NewFace(font, &opts)

	var widths int
	// Calculate the widths and print to image

	for _, t := range text {
		awidth, ok := face.GlyphAdvance(rune(t))
		if ok != true {
			return
		}
		iwidthf := int(float64(awidth) / 64)
		widths = widths + iwidthf
	}
	write(fb, text, width/2-widths/2, y, size, fontname)
}
