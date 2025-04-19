package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

// CreateDVDLogo creates a DVD-like logo with the given color
func CreateDVDLogo(width, height int, bgColor color.RGBA) *ebiten.Image {
	// Create a new ebiten image
	ebitenImg := ebiten.NewImage(width, height)

	// Fill the background with the specified color
	ebitenImg.Fill(bgColor)

	// Add text "DVD" to the image
	// Convert basicfont.Face to text.Face for text/v2
	face := text.NewGoXFace(basicfont.Face7x13)

	// Calculate text position to center it
	// We need to use Advance for measuring text width with the new API
	textWidth := text.Advance("DVD", face)
	textX := (width - int(textWidth)) / 2
	metrics := face.Metrics()
	textY := height/2 + int(metrics.HAscent)/2

	// Draw text with a slight shadow effect - shadow first
	shadowOpts := &text.DrawOptions{}
	shadowOpts.GeoM.Translate(float64(textX+1), float64(textY+1))
	shadowOpts.ColorScale.ScaleWithColor(color.RGBA{0, 0, 0, 128})
	text.Draw(ebitenImg, "DVD", face, shadowOpts)

	// Then draw the main text
	textOpts := &text.DrawOptions{}
	textOpts.GeoM.Translate(float64(textX), float64(textY))
	textOpts.ColorScale.ScaleWithColor(color.White)
	text.Draw(ebitenImg, "DVD", face, textOpts)

	return ebitenImg
}
