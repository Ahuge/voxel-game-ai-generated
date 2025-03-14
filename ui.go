package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// UI represents a user interface system for rendering text and UI elements
type UI struct {
	mesh   *Mesh
	shader *Shader
	font   *Font
}

// NewUI creates a new UI instance
func NewUI() *UI {
	// Create UI shader
	shader := NewShader("shaders/ui_vertex.glsl", "shaders/ui_fragment.glsl")

	// Create mesh for UI elements
	mesh := NewMesh()

	// Create font for text rendering
	font := NewFont()

	return &UI{
		mesh:   mesh,
		shader: shader,
		font:   font,
	}
}

// DrawText renders text at the specified position with the given color
func (ui *UI) DrawText(text string, x, y float32, scale float32, color mgl32.Vec3, alpha float32) {
	// Bind font texture
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, ui.font.Texture)
	ui.shader.SetInt("texture1", 0)

	// Set proper blending for text rendering
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Position for rendering characters
	xpos := x
	ypos := y

	// For each character in the text
	for _, r := range text {
		// Skip non-printable characters
		if r < 32 || r >= 128 {
			continue
		}

		// Get character info
		ch, ok := ui.font.Characters[r]
		if !ok {
			continue
		}

		// Calculate character position and size
		charWidth := ch.Size.X() * scale
		charHeight := ch.Size.Y() * scale

		// Calculate texture coordinates based on character position in atlas
		x := float32((int(r)%16)*8) / 128.0
		y := float32((int(r)/16)*16) / 128.0
		w := ch.Size.X() / 128.0
		h := ch.Size.Y() / 128.0

		// Create a quad for this character
		v1 := mgl32.Vec3{xpos, ypos, 0}
		v2 := mgl32.Vec3{xpos + charWidth, ypos, 0}
		v3 := mgl32.Vec3{xpos + charWidth, ypos + charHeight, 0}
		v4 := mgl32.Vec3{xpos, ypos + charHeight, 0}

		// Get current vertex count
		baseIndex := uint32(len(ui.mesh.vertices) / 3)

		// Add vertices
		ui.mesh.AddVertex(v1.X(), v1.Y(), v1.Z())
		ui.mesh.AddVertex(v2.X(), v2.Y(), v2.Z())
		ui.mesh.AddVertex(v3.X(), v3.Y(), v3.Z())
		ui.mesh.AddVertex(v4.X(), v4.Y(), v4.Z())

		// Add colors
		for i := 0; i < 4; i++ {
			ui.mesh.AddColor(color.X(), color.Y(), color.Z(), alpha)
		}

		// Add normals
		normal := mgl32.Vec3{0, 0, 1} // Facing forward
		for i := 0; i < 4; i++ {
			ui.mesh.AddNormal(normal.X(), normal.Y(), normal.Z())
		}

		// Add texture coordinates with proper UV mapping for this character
		ui.mesh.texCoords = append(ui.mesh.texCoords, x, y)
		ui.mesh.texCoords = append(ui.mesh.texCoords, x+w, y)
		ui.mesh.texCoords = append(ui.mesh.texCoords, x+w, y+h)
		ui.mesh.texCoords = append(ui.mesh.texCoords, x, y+h)

		// Add indices for two triangles
		ui.mesh.AddIndex(baseIndex)
		ui.mesh.AddIndex(baseIndex + 1)
		ui.mesh.AddIndex(baseIndex + 2)

		ui.mesh.AddIndex(baseIndex)
		ui.mesh.AddIndex(baseIndex + 2)
		ui.mesh.AddIndex(baseIndex + 3)

		// Advance cursor for next character
		xpos += ch.Advance * scale
	}
}

// RenderPauseMenu renders the pause menu on screen
func (ui *UI) RenderPauseMenu(menuOptions []string, selectedOption int, windowWidth, windowHeight int) {
	// Clear the mesh for new frame
	ui.mesh.Clear()

	// Use UI shader
	ui.shader.Use()

	// Set up orthographic projection for 2D rendering
	// Use a proper orthographic projection for UI rendering
	projection := mgl32.Ortho(0, float32(windowWidth), float32(windowHeight), 0, -1, 1)
	ui.shader.SetMat4("projection", projection)

	// Set model and view to identity matrices for 2D rendering
	model := mgl32.Ident4()
	view := mgl32.Ident4()
	ui.shader.SetMat4("model", model)
	ui.shader.SetMat4("view", view)

	// Save current OpenGL state
	depthTestWasEnabled := gl.IsEnabled(gl.DEPTH_TEST)
	var previousBlendSrc, previousBlendDst int32
	gl.GetIntegerv(gl.BLEND_SRC_ALPHA, &previousBlendSrc)
	gl.GetIntegerv(gl.BLEND_DST_ALPHA, &previousBlendDst)

	// Disable depth testing for UI rendering
	gl.Disable(gl.DEPTH_TEST)

	// Draw semi-transparent background overlay
	overlayColor := mgl32.Vec3{0.0, 0.0, 0.0}                      // Black
	ui.renderOverlay(windowWidth, windowHeight, overlayColor, 0.5) // 70% opacity

	// Calculate menu dimensions
	menuWidth := float32(300)
	menuHeight := float32(len(menuOptions)*40 + 60) // 40 pixels per option + padding
	menuX := float32(windowWidth)/2 - menuWidth/2
	menuY := float32(windowHeight)/2 - menuHeight/2

	// Draw menu background
	menuBgColor := mgl32.Vec3{0.2, 0.2, 0.2}                                            // Dark gray
	ui.renderRect(mgl32.Vec3{menuX, menuY, 0}, menuWidth, menuHeight, menuBgColor, 0.9) // 90% opacity

	// Draw menu title
	titleY := menuY + 20
	titleWidth := menuWidth - 40
	titleHeight := float32(30)
	titleColor := mgl32.Vec3{0.9, 0.9, 0.9} // Light gray
	ui.renderRect(mgl32.Vec3{menuX + 20, titleY, 0}, titleWidth, titleHeight, titleColor, 0.2)

	// Render title text
	titleText := "GAME PAUSED"
	titleTextColor := mgl32.Vec3{1.0, 1.0, 1.0}               // White
	titleX := menuX + menuWidth/2 - float32(len(titleText)*4) // Approximate centering
	ui.DrawText(titleText, titleX, titleY+8, 1.0, titleTextColor, 1.0)

	// Draw menu options
	for i, option := range menuOptions {
		optionY := menuY + 60 + float32(i*40)
		optionWidth := menuWidth - 40
		optionHeight := float32(30)

		// Different color for selected option
		optionColor := mgl32.Vec3{0.3, 0.3, 0.3} // Gray for unselected
		textColor := mgl32.Vec3{0.8, 0.8, 0.8}   // Light gray for text
		if i == selectedOption {
			optionColor = mgl32.Vec3{0.4, 0.6, 1.0} // Blue for selected
			textColor = mgl32.Vec3{1.0, 1.0, 1.0}   // White for selected text
		}

		// Render option background
		ui.renderRect(mgl32.Vec3{menuX + 20, optionY, 0}, optionWidth, optionHeight, optionColor, 0.8)

		// Render the option text
		ui.DrawText(option, menuX+40, optionY+10, 1.0, textColor, 1.0)
	}

	// Upload mesh data to GPU
	ui.mesh.Upload()

	// Draw the UI
	ui.mesh.Draw()

	// Restore previous OpenGL state
	if depthTestWasEnabled {
		gl.Enable(gl.DEPTH_TEST)
	} else {
		gl.Disable(gl.DEPTH_TEST)
	}

	// Restore previous blend function
	gl.BlendFunc(uint32(previousBlendSrc), uint32(previousBlendDst))
}

// renderRect renders a rectangle at the specified position with the given dimensions and color
func (ui *UI) renderRect(pos mgl32.Vec3, width, height float32, color mgl32.Vec3, alpha float32) {
	// Calculate vertices
	v1 := mgl32.Vec3{pos.X(), pos.Y(), pos.Z()}
	v2 := mgl32.Vec3{pos.X() + width, pos.Y(), pos.Z()}
	v3 := mgl32.Vec3{pos.X() + width, pos.Y() + height, pos.Z()}
	v4 := mgl32.Vec3{pos.X(), pos.Y() + height, pos.Z()}

	// Add quad to mesh
	normal := mgl32.Vec3{0, 0, 1} // Facing forward
	ui.mesh.AddQuad(v1, v2, v3, v4, color, normal, alpha)
}

// renderOverlay renders a full-screen overlay with the given color and opacity
func (ui *UI) renderOverlay(windowWidth, windowHeight int, color mgl32.Vec3, alpha float32) {
	ui.renderRect(mgl32.Vec3{0, 0, 0}, float32(windowWidth), float32(windowHeight), color, alpha)
}
