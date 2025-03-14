package main

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Player represents the player in the game
type Player struct {
	camera   *Camera
	speed    float32
	jump     float32
	gravity  float32
	onGround bool
	velocity mgl32.Vec3
	position mgl32.Vec3
	keys     map[glfw.Key]bool
}

// NewPlayer creates a new player instance
func NewPlayer(camera *Camera) *Player {
	return &Player{
		camera:   camera,
		speed:    5.0,
		jump:     5.0,
		gravity:  10.0,
		onGround: false,
		velocity: mgl32.Vec3{0, 0, 0},
		position: camera.Position,
		keys:     make(map[glfw.Key]bool),
	}
}

// Update updates the player state
func (p *Player) Update(deltaTime float64, window *glfw.Window) {
	dt := float32(deltaTime)

	// Store original position for collision detection
	// originalPos := p.position

	// Apply gravity
	if !p.onGround {
		p.velocity[1] -= p.gravity * dt
	}

	//// Shift for going down, E for going up
	//if p.keys[glfw.KeyLeftShift] {
	//	//p.velocity.Mul(p.speed)
	//	p.speed = 25
	//}

	// Handle movement - WASD for horizontal movement
	if p.keys[glfw.KeyW] {
		// Move forward in the direction the camera is facing, but only on the XZ plane
		forward := mgl32.Vec3{p.camera.Front.X(), 0, p.camera.Front.Z()}.Normalize()
		p.position = p.position.Add(forward.Mul(p.speed * dt))
	}
	if p.keys[glfw.KeyS] {
		// Move backward in the direction the camera is facing, but only on the XZ plane
		forward := mgl32.Vec3{p.camera.Front.X(), 0, p.camera.Front.Z()}.Normalize()
		p.position = p.position.Sub(forward.Mul(p.speed * dt))
	}
	if p.keys[glfw.KeyA] {
		// Move left relative to the camera direction
		p.position = p.position.Sub(p.camera.Right.Mul(p.speed * dt))
	}
	if p.keys[glfw.KeyD] {
		// Move right relative to the camera direction
		p.position = p.position.Add(p.camera.Right.Mul(p.speed * dt))
	}

	// Space for jumping
	if p.keys[glfw.KeySpace] {
		if p.onGround {
			p.velocity[1] += p.jump
			p.onGround = false
		}
	}

	if p.keys[glfw.KeyE] {
		p.position[1] += p.speed * dt
	}

	// Update position based on velocity
	p.position = p.position.Add(p.velocity.Mul(dt))

	// Check for collisions and adjust position
	p.handleCollisions()

	// Update camera position
	p.camera.Position = p.position

	// Update player chunk position in world for chunk loading/unloading
	if currentGame != nil && currentGame.world != nil {
		currentGame.world.UpdatePlayerPosition(p.position)
	}
}

// handleCollisions checks for collisions with blocks and adjusts player position
func (p *Player) handleCollisions() {
	// Player dimensions (hitbox)
	playerWidth := float32(0.3)  // Half-width of player
	playerHeight := float32(1.8) // Height of player

	// Get the world from the game instance
	world := p.getWorld()
	if world == nil {
		return
	}

	// Check feet position
	feetY := p.position[1] - 0.5 // Offset from center to feet

	// Check if we're standing on a block
	blockX := int(math.Floor(float64(p.position[0])))
	blockY := int(math.Floor(float64(feetY)))
	blockZ := int(math.Floor(float64(p.position[2])))

	// Check block below feet
	blockBelow := world.GetBlock(blockX, blockY-1, blockZ)
	if blockBelow != Air && world.blockRegistry.IsSolid(blockBelow) && p.onGround == true {
		// Standing on a block
		p.position[1] = float32(blockY) + 0.5 // Position player on top of block
		p.velocity[1] = 0
		p.onGround = true
	} else {
		// Not standing on a block
		p.onGround = false
	}

	// Check for collisions in all directions
	// Check blocks at player's position and head height
	for y := 0; y < 2; y++ { // Check at feet and head level
		checkY := blockY + y
		if checkY < 0 || checkY >= WorldHeight {
			continue
		}

		// Check surrounding blocks in X and Z directions
		for xOffset := -1; xOffset <= 1; xOffset++ {
			for zOffset := -1; zOffset <= 1; zOffset++ {
				// Skip the center block when checking at feet level (we're standing in it)
				if xOffset == 0 && zOffset == 0 && y == 0 {
					continue
				}

				checkX := blockX + xOffset
				checkZ := blockZ + zOffset

				block := world.GetBlock(checkX, checkY, checkZ)
				if block != Air && world.blockRegistry.IsSolid(block) {
					// Calculate block boundaries
					minX := float32(checkX)
					maxX := float32(checkX + 1)
					minZ := float32(checkZ)
					maxZ := float32(checkZ + 1)
					minY := float32(checkY)
					maxY := float32(checkY + 1)

					// Calculate player boundaries
					playerMinX := p.position[0] - playerWidth
					playerMaxX := p.position[0] + playerWidth
					playerMinZ := p.position[2] - playerWidth
					playerMaxZ := p.position[2] + playerWidth
					playerMinY := p.position[1] - 0.5                // Feet position
					playerMaxY := p.position[1] + playerHeight - 0.5 // Head position

					// Check for collision
					if playerMaxX > minX && playerMinX < maxX &&
						playerMaxY > minY && playerMinY < maxY &&
						playerMaxZ > minZ && playerMinZ < maxZ {

						// Collision detected, resolve it
						// Calculate penetration depths
						penetrationX1 := playerMaxX - minX
						penetrationX2 := maxX - playerMinX
						penetrationZ1 := playerMaxZ - minZ
						penetrationZ2 := maxZ - playerMinZ
						penetrationY1 := playerMaxY - minY
						penetrationY2 := maxY - playerMinY

						// Find minimum penetration
						minPenetrationX := float32(math.Min(float64(penetrationX1), float64(penetrationX2)))
						minPenetrationZ := float32(math.Min(float64(penetrationZ1), float64(penetrationZ2)))
						minPenetrationY := float32(math.Min(float64(penetrationY1), float64(penetrationY2)))

						// Resolve along axis with minimum penetration
						if minPenetrationX < minPenetrationY && minPenetrationX < minPenetrationZ {
							// Resolve along X axis
							if penetrationX1 < penetrationX2 {
								p.position[0] = minX - playerWidth
							} else {
								p.position[0] = maxX + playerWidth
							}
						} else if minPenetrationZ < minPenetrationY {
							// Resolve along Z axis
							if penetrationZ1 < penetrationZ2 {
								p.position[2] = minZ - playerWidth
							} else {
								p.position[2] = maxZ + playerWidth
							}
						} else {
							// Resolve along Y axis
							if penetrationY1 < penetrationY2 {
								p.position[1] = minY - playerHeight + 0.5
								p.velocity[1] = 0
							} else {
								p.position[1] = maxY + 0.5
								if p.velocity[1] < 0 {
									p.velocity[1] = 0
									p.onGround = true
								}
							}
						}
					}
				}
			}
		}
	}
	block := world.GetBlock(int(math.Round(float64(p.position[0]))), int(math.Round(float64(p.position[1]))), int(math.Round(float64(p.position[2]))))
	blockAbove := world.GetBlock(int(math.Round(float64(p.position[0]))), int(math.Round(float64(p.position[1]))+1), int(math.Round(float64(p.position[2]))))
	if block == Dirt && world.blockRegistry.IsSolid(block) {
		fmt.Printf("We are inside dirt\n")
	}
	if (block != Air && world.blockRegistry.IsSolid(block)) || (blockAbove != Air && world.blockRegistry.IsSolid(blockAbove)) {
		if block != Air {
			fmt.Sprintf("Hack fix position block\n")
			p.position[1] = p.position[1] + 1.5 // Position player on top of block
		}
		if blockAbove != Air {
			fmt.Sprintf("Hack fix position blockAbove\n")
			p.position[1] = p.position[1] + 1.5 // Position player on top of block
		}
		p.velocity[1] = 0
		p.onGround = true
	}
	if (block != Air && world.blockRegistry.IsSolid(block)) || (blockAbove != Air && world.blockRegistry.IsSolid(blockAbove)) {
		fmt.Printf("BadDay\n")
	}

	// Simple ground collision as a fallback
	if p.position[1] < 0.5 {
		p.position[1] = 0.5
		p.velocity[1] = 0
		p.onGround = true
	}
}

// getWorld returns the world from the game instance
func (p *Player) getWorld() *World {
	// This is a bit of a hack since we don't have direct access to the world
	// We'll use a global variable to access the current game instance
	if currentGame != nil {
		return currentGame.world
	}
	return nil
}

// KeyCallback handles key input
func (p *Player) KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Track key states
	if action == glfw.Press {
		p.keys[key] = true
		if key == glfw.KeyLeftShift {
			p.speed = p.speed * p.speed
		}
	} else if action == glfw.Release {
		p.keys[key] = false
		if key == glfw.KeyLeftShift {
			p.speed = float32(math.Sqrt(float64(p.speed)))
		}
	}
}
