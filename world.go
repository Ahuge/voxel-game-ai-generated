package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png" // Register PNG decoder
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Render distance constants
const (
	RenderDist = 10 // Increased from 16 to allow for better visibility
)

// World represents the voxel world
type World struct {
	chunks         map[ChunkPos]*Chunk
	blockRegistry  *BlockRegistry
	terrainGen     *TerrainGenerator
	blockTexture   uint32
	textures       map[string]uint32
	playerChunkPos ChunkPos
	frustum        *Frustum
	chunkMeshCache map[ChunkPos]bool // Tracks which chunks have valid meshes
}

// NewWorld creates a new world instance
func NewWorld(seed int64) *World {
	world := &World{
		chunks:         make(map[ChunkPos]*Chunk),
		blockRegistry:  NewBlockRegistry(),
		terrainGen:     NewTerrainGenerator(seed), // Seed for terrain generation
		textures:       make(map[string]uint32),
		frustum:        NewFrustum(mgl32.Ident4(), mgl32.Ident4()), // Will be updated later
		chunkMeshCache: make(map[ChunkPos]bool),
	}

	// Load block textures
	world.loadTextures()

	// Generate initial chunks around origin
	world.playerChunkPos = ChunkPos{X: 0, Z: 0}
	world.generateInitialChunks()

	return world
}

func (w *World) loadTexture(path string) ([]byte, int, int) {
	imgFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, 0, 0
	}
	defer imgFile.Close()

	// Use image package to decode the image
	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return nil, 0, 0
	}

	// Get image dimensions
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert image to RGBA
	rgbaImg := image.NewRGBA(bounds)
	draw.Draw(rgbaImg, bounds, img, bounds.Min, draw.Src)

	return rgbaImg.Pix, width, height
}

// loadTextures loads the block textures
func (w *World) loadTextures() {
	// In a real implementation, this would load textures from files
	// For simplicity, we'll just create a simple texture here
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	// Create a simple checkerboard texture
	width, height := 16, 16
	pixels := make([]byte, width*height*4)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 4
			if (x/4+y/4)%2 == 0 {
				pixels[i] = 0x80   // R
				pixels[i+1] = 0x80 // G
				pixels[i+2] = 0x80 // B
				pixels[i+3] = 0xFF // A
			} else {
				pixels[i] = 0xFF   // R
				pixels[i+1] = 0xFF // G
				pixels[i+2] = 0xFF // B
				pixels[i+3] = 0xFF // A
			}
		}
	}

	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	gl.GenerateMipmap(gl.TEXTURE_2D)

	w.blockTexture = texture

	filepath.Walk("./textures", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".png") {
			fmt.Println("Found PNG:", path)
			byteImage, imgWidth, imgHeight := w.loadTexture(path)
			if byteImage == nil {
				fmt.Println("Failed to load texture:", path)
				return nil
			}

			var textureImage uint32
			gl.GenTextures(1, &textureImage)
			gl.BindTexture(gl.TEXTURE_2D, textureImage)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

			gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(imgWidth), int32(imgHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(byteImage))
			gl.GenerateMipmap(gl.TEXTURE_2D)

			// Store texture with block name (without extension)
			blockName := strings.Split(filepath.Base(path), ".")[0]
			fmt.Println("Loaded texture for block:", blockName)
			w.textures[blockName] = textureImage
		}
		return nil
	})
}

// generateInitialChunks generates chunks around the player
func (w *World) generateInitialChunks() {
	for x := -RenderDist; x <= RenderDist; x++ {
		for z := -RenderDist; z <= RenderDist; z++ {
			pos := ChunkPos{X: w.playerChunkPos.X + x, Z: w.playerChunkPos.Z + z}
			w.GetOrCreateChunk(pos)
		}
	}
}

// GetOrCreateChunk gets an existing chunk or creates a new one if it doesn't exist
func (w *World) GetOrCreateChunk(pos ChunkPos) *Chunk {
	// Check if chunk already exists
	if chunk, exists := w.chunks[pos]; exists {
		return chunk
	}

	// Create new chunk
	chunk := NewChunk(pos, w.blockRegistry)

	// Generate terrain
	w.terrainGen.GenerateChunkTerrain(chunk)

	// Generate structures (trees, etc.)
	w.terrainGen.GenerateStructures(chunk)

	// Store chunk in map
	w.chunks[pos] = chunk

	return chunk
}

// Update updates the world state
func (w *World) Update(deltaTime float64) {
	// Update chunks that need updating
	for _, chunk := range w.chunks {
		if chunk.needsUpdate {
			chunk.BuildMesh(w)
		}
	}
}

// Render renders the world
func (w *World) Render(shader *Shader) {
	// Set texture unit for the shader
	shader.SetInt("texture1", 0)

	// Set model matrix (identity for world)
	model := mgl32.Ident4()
	shader.SetMat4("model", model)

	// Get the current view and projection matrices from the shader
	view := shader.GetMat4("view")
	projection := shader.GetMat4("projection")

	// Update the frustum with current matrices
	w.frustum.Update(projection, view)

	// Set view position for lighting in the shader
	// This assumes the camera position is available
	if currentGame != nil && currentGame.camera != nil {
		shader.SetVec3("viewPos", currentGame.camera.Position)
	}

	// Count of rendered chunks (for debugging)
	renderedChunks := 0

	// Render only visible chunks
	for _, chunk := range w.chunks {
		// Skip empty chunks
		if chunk.IsEmpty() {
			continue
		}

		// Check if chunk is visible in frustum
		if !w.frustum.IsChunkVisible(chunk) {
			continue
		}

		// Render the chunk
		chunk.Render(w)
		renderedChunks++
	}

	// Enable debugging output to monitor chunk rendering
	// This helps identify when frustum culling is too aggressive
	if renderedChunks == 0 {
		fmt.Printf("WARNING: No chunks rendered! Player at chunk (%d, %d)\n", w.playerChunkPos.X, w.playerChunkPos.Z)
	} else if renderedChunks < 5 {
		// Only a few chunks being rendered might indicate a problem
		fmt.Printf("Low chunk count: %d/%d chunks rendered\n", renderedChunks, len(w.chunks))
	}

	// For periodic debugging
	if currentGame != nil && currentGame.frameCount%120 == 0 {
		fmt.Printf("Rendered %d/%d chunks\n", renderedChunks, len(w.chunks))
	}
}

// GetBlock returns the block type at the given world coordinates
func (w *World) GetBlock(x, y, z int) BlockType {
	// Calculate chunk position
	chunkX := x / ChunkSize
	chunkZ := z / ChunkSize

	// Calculate local coordinates within chunk
	localX := x % ChunkSize
	localZ := z % ChunkSize

	// Handle negative coordinates
	if x < 0 && localX != 0 {
		chunkX--
		localX = ChunkSize + localX
	}
	if z < 0 && localZ != 0 {
		chunkZ--
		localZ = ChunkSize + localZ
	}

	// Check if chunk exists
	chunk, exists := w.chunks[ChunkPos{X: chunkX, Z: chunkZ}]
	if !exists {
		return Air
	}

	// Check if y is in bounds
	if y < 0 || y >= WorldHeight {
		return Air
	}

	return chunk.GetBlock(localX, y, localZ)
}

// SetBlock sets the block type at the given world coordinates
func (w *World) SetBlock(x, y, z int, blockType BlockType) {
	// Calculate chunk position
	chunkX := x / ChunkSize
	chunkZ := z / ChunkSize

	// Calculate local coordinates within chunk
	localX := x % ChunkSize
	localZ := z % ChunkSize

	// Handle negative coordinates
	if x < 0 && localX != 0 {
		chunkX--
		localX = ChunkSize + localX
	}
	if z < 0 && localZ != 0 {
		chunkZ--
		localZ = ChunkSize + localZ
	}

	// Check if chunk exists, create if not
	chunkPos := ChunkPos{X: chunkX, Z: chunkZ}
	chunk := w.GetOrCreateChunk(chunkPos)

	// Check if y is in bounds
	if y < 0 || y >= WorldHeight {
		return
	}

	// Set block
	chunk.SetBlock(localX, y, localZ, blockType)

	// Mark neighboring chunks for update if block is on the edge
	if localX == 0 {
		if neighbor, exists := w.chunks[ChunkPos{X: chunkX - 1, Z: chunkZ}]; exists {
			neighbor.needsUpdate = true
		}
	} else if localX == ChunkSize-1 {
		if neighbor, exists := w.chunks[ChunkPos{X: chunkX + 1, Z: chunkZ}]; exists {
			neighbor.needsUpdate = true
		}
	}

	if localZ == 0 {
		if neighbor, exists := w.chunks[ChunkPos{X: chunkX, Z: chunkZ - 1}]; exists {
			neighbor.needsUpdate = true
		}
	} else if localZ == ChunkSize-1 {
		if neighbor, exists := w.chunks[ChunkPos{X: chunkX, Z: chunkZ + 1}]; exists {
			neighbor.needsUpdate = true
		}
	}
}

// BreakBlock breaks (removes) a block at the position the player is looking at
func (w *World) BreakBlock(playerPos, playerDir mgl32.Vec3) {
	// Simple raycast to find block
	blockPos := w.raycastBlock(playerPos, playerDir, 5.0) // 5.0 is max distance
	if blockPos != nil {
		w.SetBlock(blockPos[0], blockPos[1], blockPos[2], Air)
	}
}

// PlaceBlock places a block adjacent to the block the player is looking at
func (w *World) PlaceBlock(playerPos, playerDir mgl32.Vec3) {
	// Simple raycast to find block
	blockPos := w.raycastBlock(playerPos, playerDir, 5.0) // 5.0 is max distance
	if blockPos != nil {
		// Find the face that was hit and place block adjacent to it
		// This is a simplified version - in a real game you'd need to determine which face was hit
		// For now, just place a block one unit back along the ray
		dir := playerDir.Normalize()
		placeX := blockPos[0] - int(dir.X())
		placeY := blockPos[1] - int(dir.Y())
		placeZ := blockPos[2] - int(dir.Z())

		// Don't place if it would overlap with the player
		playerBlockX := int(playerPos.X())
		playerBlockY := int(playerPos.Y())
		playerBlockZ := int(playerPos.Z())
		if placeX == playerBlockX && placeY == playerBlockY && placeZ == playerBlockZ {
			return
		}
		if placeX == playerBlockX && placeY == playerBlockY+1 && placeZ == playerBlockZ {
			return // Don't place at player's head position
		}

		// Place a dirt block for now - in a real game you'd select the block type
		w.SetBlock(placeX, placeY, placeZ, Dirt)
	}
}

// raycastBlock performs a simple raycast to find a block
func (w *World) raycastBlock(start, dir mgl32.Vec3, maxDist float32) []int {
	// Normalize direction
	dir = dir.Normalize()

	// Step along ray
	for dist := 0.0; dist < float64(maxDist); dist += 0.1 {
		// Calculate position along ray
		pos := start.Add(dir.Mul(float32(dist)))

		// Convert to block coordinates
		blockX := int(pos.X())
		blockY := int(pos.Y())
		blockZ := int(pos.Z())

		// Check if this block is solid
		block := w.GetBlock(blockX, blockY, blockZ)
		if block != Air && w.blockRegistry.IsSolid(block) {
			return []int{blockX, blockY, blockZ}
		}
	}

	return nil
}

// UpdatePlayerPosition updates which chunks are loaded based on player position
func (w *World) UpdatePlayerPosition(playerPos mgl32.Vec3) {
	// Calculate which chunk the player is in
	newChunkPos := ChunkPos{
		X: int(playerPos.X()) / ChunkSize,
		Z: int(playerPos.Z()) / ChunkSize,
	}

	// If player moved to a new chunk, update loaded chunks
	if newChunkPos != w.playerChunkPos {
		w.playerChunkPos = newChunkPos
		w.updateLoadedChunks()
	}
}

// updateLoadedChunks loads and unloads chunks based on player position
func (w *World) updateLoadedChunks() {
	// Determine which chunks should be loaded
	chunksToLoad := make(map[ChunkPos]bool)
	for x := -RenderDist; x <= RenderDist; x++ {
		for z := -RenderDist; z <= RenderDist; z++ {
			pos := ChunkPos{X: w.playerChunkPos.X + x, Z: w.playerChunkPos.Z + z}
			chunksToLoad[pos] = true
		}
	}

	// Unload chunks that are too far away
	for pos, chunk := range w.chunks {
		if !chunksToLoad[pos] {
			// Delete chunk resources
			chunk.Delete()
			// Remove from map
			delete(w.chunks, pos)
		}
	}

	// Load new chunks
	for pos := range chunksToLoad {
		if _, exists := w.chunks[pos]; !exists {
			w.GetOrCreateChunk(pos)
		}
	}
}
