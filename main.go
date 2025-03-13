package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	width  = 800
	height = 600
	title  = "Voxel Game"
)

func init() {
	// GLFW event handling must be run on the main OS thread
	runtime.LockOSThread()
}

// Global variable to track the current game instance
var currentGame *Game

func main() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(width, height, title, nil, nil)
	if err != nil {
		log.Fatalln("Failed to create window:", err)
	}

	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalln("Failed to initialize OpenGL:", err)
	}

	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version:", version)

	// Configure global OpenGL state
	gl.Enable(gl.DEPTH_TEST)
	// Temporarily disable face culling to see if blocks become visible
	// gl.Enable(gl.CULL_FACE)
	// gl.CullFace(gl.BACK)
	// gl.FrontFace(gl.CCW)

	// Create game instance
	game := NewGame(window)

	// Set global game instance for collision detection
	currentGame = game

	// Set input callbacks
	window.SetKeyCallback(game.KeyCallback)
	window.SetCursorPosCallback(game.MouseCallback)
	window.SetMouseButtonCallback(game.MouseButtonCallback)
	window.SetScrollCallback(game.ScrollCallback)

	// Capture cursor for first-person camera control
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	// Main game loop
	for !window.ShouldClose() {
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		game.Update()
		game.Render()

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// Game represents the main game state
type Game struct {
	window       *glfw.Window
	camera       *Camera
	world        *World
	shader       *Shader
	player       *Player
	deltaTime    float64
	lastFrame    float64
	firstMouse   bool
	lastX, lastY float64
	// Pause menu related fields
	paused      bool
	menuOption  int
	menuOptions []string
}

// NewGame creates a new game instance
func NewGame(window *glfw.Window) *Game {
	game := &Game{
		window:      window,
		firstMouse:  true,
		paused:      false,
		menuOption:  0,
		menuOptions: []string{"Resume", "Regenerate World", "Exit"},
	}

	// Initialize camera
	game.camera = NewCamera(mgl32.Vec3{0, 35, 10}, mgl32.Vec3{0, 1, 0}, -90, -30)

	// Initialize shader
	game.shader = NewShader("shaders/vertex.glsl", "shaders/fragment.glsl")

	// Initialize world
	game.world = NewWorld()

	// Initialize player
	game.player = NewPlayer(game.camera)

	// Get window size
	w, h := window.GetSize()
	game.lastX = float64(w) / 2
	game.lastY = float64(h) / 2

	return game
}

// Update updates the game state
func (g *Game) Update() {
	currentFrame := glfw.GetTime()
	g.deltaTime = currentFrame - g.lastFrame
	g.lastFrame = currentFrame

	// Only update game elements if not paused
	if !g.paused {
		// Update player
		g.player.Update(g.deltaTime, g.window)

		// Update world
		g.world.Update(g.deltaTime)
	}
}

// Render renders the game
func (g *Game) Render() {
	// Use shader
	g.shader.Use()

	// Set view and projection matrices
	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(width)/float32(height), 0.1, 1000.0)
	view := g.camera.GetViewMatrix()

	g.shader.SetMat4("projection", projection)
	g.shader.SetMat4("view", view)

	// Render world
	g.world.Render(g.shader)

	// Render pause menu if game is paused
	if g.paused {
		g.renderPauseMenu()
	}
}

// renderPauseMenu renders the pause menu
func (g *Game) renderPauseMenu() {
	// This is a simple implementation that prints the menu to the console
	// In a real implementation, you would render this to the screen using OpenGL
	fmt.Println("\n===== PAUSE MENU =====")
	for i, option := range g.menuOptions {
		if i == g.menuOption {
			fmt.Printf("> %s <\n", option)
		} else {
			fmt.Printf("  %s  \n", option)
		}
	}
	fmt.Println("=====================")
}

// KeyCallback handles key input
func (g *Game) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Toggle pause menu on escape
	if key == glfw.KeyEscape && action == glfw.Press {
		g.paused = !g.paused

		// Toggle cursor visibility based on pause state
		if g.paused {
			window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		} else {
			window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		}
		return
	}

	// Handle menu navigation when paused
	if g.paused && action == glfw.Press {
		switch key {
		case glfw.KeyUp, glfw.KeyW:
			// Move up in menu
			g.menuOption--
			if g.menuOption < 0 {
				g.menuOption = len(g.menuOptions) - 1
			}
		case glfw.KeyDown, glfw.KeyS:
			// Move down in menu
			g.menuOption++
			if g.menuOption >= len(g.menuOptions) {
				g.menuOption = 0
			}
		case glfw.KeyEnter, glfw.KeySpace:
			// Select menu option
			g.selectMenuOption()
		}
		return
	}

	// Only pass key events to player if not paused
	if !g.paused {
		g.player.KeyCallback(window, key, scancode, action, mods)
	}
}

// MouseCallback handles mouse movement
func (g *Game) MouseCallback(window *glfw.Window, xpos, ypos float64) {
	// Only process mouse movement if game is not paused
	if !g.paused {
		if g.firstMouse {
			g.lastX = xpos
			g.lastY = ypos
			g.firstMouse = false
		}

		xoffset := xpos - g.lastX
		yoffset := g.lastY - ypos // Reversed: y ranges bottom to top

		g.lastX = xpos
		g.lastY = ypos

		g.camera.ProcessMouseMovement(float32(xoffset), float32(yoffset), true)
	}
}

// MouseButtonCallback handles mouse button input
func (g *Game) MouseButtonCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	// Only handle mouse input if not paused
	if !g.paused {
		if button == glfw.MouseButtonLeft && action == glfw.Press {
			// Handle block placing
			g.world.PlaceBlock(g.camera.Position, g.camera.Front)
		} else if button == glfw.MouseButtonRight && action == glfw.Press {
			// Handle block breaking (mining)
			g.world.BreakBlock(g.camera.Position, g.camera.Front)
		}
	}
}

// selectMenuOption handles the selection of menu options
func (g *Game) selectMenuOption() {
	switch g.menuOption {
	case 0: // Resume
		// Unpause the game and hide cursor
		g.paused = false
		g.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	case 1: // Regenerate World
		// Create a new world with a different seed
		// First, clean up the old world resources
		for _, chunk := range g.world.chunks {
			chunk.Delete()
		}

		// Generate a new seed based on current time
		newSeed := int64(glfw.GetTime() * 1000)

		// Create a new world with the new seed
		g.world = NewWorld()
		g.world.terrainGen = NewTerrainGenerator(newSeed)

		// Reset player position
		g.camera.Position = mgl32.Vec3{0, 35, 10}
		g.player.position = g.camera.Position
		g.player.velocity = mgl32.Vec3{0, 0, 0}

		// Unpause the game
		g.paused = false
		g.window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	case 2: // Exit
		// Close the window to exit the game
		g.window.SetShouldClose(true)
	}
}

// ScrollCallback handles scroll input
func (g *Game) ScrollCallback(window *glfw.Window, xoffset, yoffset float64) {
	// Only process scroll input if game is not paused
	if !g.paused {
		g.camera.ProcessMouseScroll(float32(yoffset))
	}
}
