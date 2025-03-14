package main

import (
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

	// Handle movement - WASD for horizontal movement
	// Apply movements separately and check collisions after each step
	// to prevent clipping through blocks when moving diagonally or quickly
	if p.keys[glfw.KeyW] {
		// Move forward in the direction the camera is facing, but only on the XZ plane
		forward := mgl32.Vec3{p.camera.Front.X(), 0, p.camera.Front.Z()}.Normalize()
		p.position = p.position.Add(forward.Mul(p.speed * dt))
		// Check collisions after forward movement
		// p.handleCollisions()
	}
	if p.keys[glfw.KeyS] {
		// Move backward in the direction the camera is facing, but only on the XZ plane
		forward := mgl32.Vec3{p.camera.Front.X(), 0, p.camera.Front.Z()}.Normalize()
		p.position = p.position.Sub(forward.Mul(p.speed * dt))
		// Check collisions after backward movement
		// p.handleCollisions()
	}
	if p.keys[glfw.KeyA] {
		// Move left relative to the camera direction
		p.position = p.position.Sub(p.camera.Right.Mul(p.speed * dt))
		// Check collisions after left movement
		// p.handleCollisions()
	}
	if p.keys[glfw.KeyD] {
		// Move right relative to the camera direction
		p.position = p.position.Add(p.camera.Right.Mul(p.speed * dt))
		// Check collisions after right movement
		// p.handleCollisions()
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
		// Check collisions after vertical movement
		// p.handleCollisions()
	}

	// Update position based on velocity
	p.position = p.position.Add(p.velocity.Mul(dt))

	// Final collision check after velocity-based movement
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
	// Slightly increased hitbox size for collision detection to prevent clipping
	playerWidth := float32(0.4)   // Increased from 0.3 to 0.35 for safer collision detection
	playerHeight := float32(1.95) // Increased from 1.8 to 1.85 for safer collision detection

	// Get the world from the game instance
	world := p.getWorld()
	if world == nil {
		return
	}

	// Store original position to detect if we're moving too fast
	originalPos := p.position

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
	// Expanded check range to ensure we don't miss any blocks
	for y := -1; y < 3; y++ { // Check below feet, at feet, at head level, and above head
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

						// Add a small buffer to prevent exact edge cases
						const buffer float32 = 0.01

						// Resolve along axis with minimum penetration
						if minPenetrationX < minPenetrationY && minPenetrationX < minPenetrationZ {
							// Resolve along X axis
							if penetrationX1 < penetrationX2 {
								p.position[0] = minX - playerWidth - buffer
							} else {
								p.position[0] = maxX + playerWidth + buffer
							}
						} else if minPenetrationZ < minPenetrationY {
							// Resolve along Z axis
							if penetrationZ1 < penetrationZ2 {
								p.position[2] = minZ - playerWidth - buffer
							} else {
								p.position[2] = maxZ + playerWidth + buffer
							}
						} else {
							// Resolve along Y axis
							if penetrationY1 < penetrationY2 {
								p.position[1] = minY - playerHeight + 0.5 - buffer
								p.velocity[1] = 0
							} else {
								p.position[1] = maxY + 0.5 + buffer
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

	// Additional safety check - if we're inside a block, move up
	// Check the exact position we're in
	block := world.GetBlock(int(math.Floor(float64(p.position[0]))),
		int(math.Floor(float64(p.position[1]))),
		int(math.Floor(float64(p.position[2]))))

	blockAbove := world.GetBlock(int(math.Floor(float64(p.position[0]))),
		int(math.Floor(float64(p.position[1]))+1),
		int(math.Floor(float64(p.position[2]))))

	// If we're still inside a block after all collision resolution, emergency fix
	if (block != Air && world.blockRegistry.IsSolid(block)) || (blockAbove != Air && world.blockRegistry.IsSolid(blockAbove)) {
		// We're inside a block - emergency escape upward
		p.position[1] = float32(math.Ceil(float64(p.position[1]))) + 0.5
		p.velocity[1] = 0
		p.onGround = true
	}

	// Check if we're moving too fast (teleporting through blocks)
	distMoved := originalPos.Sub(p.position).Len()
	if distMoved > 1.0 {
		// We moved more than 1 block in a single frame - do additional collision checks
		// This helps catch cases where we might have passed through a thin wall
		steps := int(math.Ceil(float64(distMoved))) + 1
		stepSize := distMoved / float32(steps)

		// Interpolate between original and current position
		for i := 1; i < steps; i++ {
			t := float32(i) * stepSize
			interpPos := originalPos.Add(p.position.Sub(originalPos).Mul(t / distMoved))

			// Check if this interpolated position is inside a block
			interpBlock := world.GetBlock(
				int(math.Floor(float64(interpPos[0]))),
				int(math.Floor(float64(interpPos[1]))),
				int(math.Floor(float64(interpPos[2]))))

			if interpBlock != Air && world.blockRegistry.IsSolid(interpBlock) {
				// Found collision during interpolation - move player to safe position
				p.position = interpPos
				p.position[1] = float32(math.Ceil(float64(interpPos[1]))) + 0.5
				break
			}
		}
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
