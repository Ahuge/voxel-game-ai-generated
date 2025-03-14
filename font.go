package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"

	// "io/ioutil"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Character holds information about a single character in the font
type Character struct {
	TextureID uint32     // ID of the glyph texture
	Size      mgl32.Vec2 // Size of glyph
	Bearing   mgl32.Vec2 // Offset from baseline to left/top of glyph
	Advance   float32    // Offset to advance to next glyph
}

// Font represents a bitmap font for rendering text
type Font struct {
	Characters map[rune]Character
	Texture    uint32
	FaceSize   int
}

// NewFont creates a new bitmap font
func NewFont() *Font {
	// Create a new font
	f := &Font{
		Characters: make(map[rune]Character),
		FaceSize:   13, // Default size for basicfont
	}

	// Generate a texture atlas for the font
	f.generateFontAtlas()

	return f
}

// generateFontAtlas creates a texture atlas for the font
func (f *Font) generateFontAtlas() {
	// We'll use Go's basic font for simplicity
	face := basicfont.Face7x13
	f.FaceSize = 13

	// Create an image to draw the characters on
	atlas := image.NewRGBA(image.Rect(0, 0, 128, 128))
	draw.Draw(atlas, atlas.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Create a drawer to draw text on the image
	d := &font.Drawer{
		Dst:  atlas,
		Src:  image.White,
		Face: face,
	}

	// Draw each ASCII character to the atlas
	for r := rune(32); r < 128; r++ { // ASCII printable characters
		x := (int(r) % 16) * 8
		y := (int(r) / 16) * 16

		d.Dot = fixed.P(x, y+12) // +12 to account for baseline
		d.DrawString(string(r))

		// Store character information
		advance, ok := face.GlyphAdvance(r)
		if !ok {
			advance = fixed.I(7) // Default width for basicfont
		}

		f.Characters[r] = Character{
			Size:    mgl32.Vec2{7, 13}, // Fixed size for basicfont
			Bearing: mgl32.Vec2{0, 0},
			Advance: float32(advance.Round()),
		}
	}

	// Generate OpenGL texture
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Upload texture data
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		128,
		128,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(atlas.Pix),
	)

	f.Texture = texture

	// For debugging: save the atlas as a PNG
	// saveAtlasToFile(atlas, "font_atlas.png")
}

// saveAtlasToFile saves the font atlas to a file (for debugging)
func saveAtlasToFile(img *image.RGBA, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	png.Encode(file, img)
}

// Delete releases the font resources
func (f *Font) Delete() {
	gl.DeleteTextures(1, &f.Texture)
}
