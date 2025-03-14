package main

import (
	// "github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Chunk constants
const (
	ChunkSize   = 64
	WorldHeight = 128 // Increased from 10 to allow for more vertical terrain
)

// ChunkPos represents a chunk position in the world
type ChunkPos struct {
	X, Z int
}

// Chunk represents a section of the world
type Chunk struct {
	blocks        [ChunkSize][WorldHeight][ChunkSize]BlockType
	mesh          *Mesh
	worldPos      ChunkPos
	needsUpdate   bool
	blockRegistry *BlockRegistry
}

// NewChunk creates a new chunk at the given position
func NewChunk(pos ChunkPos, registry *BlockRegistry) *Chunk {
	return &Chunk{
		worldPos:      pos,
		needsUpdate:   true,
		blockRegistry: registry,
	}
}

// GetBlock returns the block type at the given local coordinates
func (c *Chunk) GetBlock(x, y, z int) BlockType {
	// Check bounds
	if x < 0 || x >= ChunkSize || y < 0 || y >= WorldHeight || z < 0 || z >= ChunkSize {
		return Air
	}
	return c.blocks[x][y][z]
}

// SetBlock sets the block type at the given local coordinates
func (c *Chunk) SetBlock(x, y, z int, blockType BlockType) {
	// Check bounds
	if x < 0 || x >= ChunkSize || y < 0 || y >= WorldHeight || z < 0 || z >= ChunkSize {
		return
	}
	c.blocks[x][y][z] = blockType
	c.needsUpdate = true
}

// IsEmpty returns true if the chunk contains only air blocks
func (c *Chunk) IsEmpty() bool {
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < WorldHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				if c.blocks[x][y][z] != Air {
					return false
				}
			}
		}
	}
	return true
}

// BuildMesh builds the chunk's mesh for rendering
func (c *Chunk) BuildMesh() {
	// Create a new mesh if needed
	if c.mesh == nil {
		c.mesh = NewMesh()
	} else {
		c.mesh.Clear()
	}

	// For each block in the chunk
	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < WorldHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				blockType := c.blocks[x][y][z]

				// Skip air blocks
				if blockType == Air {
					continue
				}

				// Get block info
				blockInfo := c.blockRegistry.GetBlockInfo(blockType)

				// Calculate world position
				worldX := float32(c.worldPos.X*ChunkSize + x)
				worldY := float32(y)
				worldZ := float32(c.worldPos.Z*ChunkSize + z)

				// Add faces for this block
				c.addBlockFaces(worldX, worldY, worldZ, x, y, z, blockType, blockInfo.Color)
			}
		}
	}

	// Upload mesh data to GPU
	c.mesh.Upload()

	// Mark chunk as updated
	c.needsUpdate = false
}

// addBlockFaces adds visible faces of a block to the mesh
func (c *Chunk) addBlockFaces(worldX, worldY, worldZ float32, x, y, z int, blockType BlockType, color mgl32.Vec3) {
	// Check each of the 6 faces

	// Top face (+Y)
	if c.shouldRenderFace(x, y+1, z, blockType) {
		v1 := mgl32.Vec3{worldX, worldY + 1, worldZ}
		v2 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ}
		v3 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ + 1}
		v4 := mgl32.Vec3{worldX, worldY + 1, worldZ + 1}
		normal := mgl32.Vec3{0, 1, 0}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}

	// Bottom face (-Y)
	if c.shouldRenderFace(x, y-1, z, blockType) {
		v1 := mgl32.Vec3{worldX, worldY, worldZ + 1}
		v2 := mgl32.Vec3{worldX + 1, worldY, worldZ + 1}
		v3 := mgl32.Vec3{worldX + 1, worldY, worldZ}
		v4 := mgl32.Vec3{worldX, worldY, worldZ}
		normal := mgl32.Vec3{0, -1, 0}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}

	// Front face (+Z)
	if c.shouldRenderFace(x, y, z+1, blockType) {
		v1 := mgl32.Vec3{worldX, worldY, worldZ + 1}
		v2 := mgl32.Vec3{worldX, worldY + 1, worldZ + 1}
		v3 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ + 1}
		v4 := mgl32.Vec3{worldX + 1, worldY, worldZ + 1}
		normal := mgl32.Vec3{0, 0, 1}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}

	// Back face (-Z)
	if c.shouldRenderFace(x, y, z-1, blockType) {
		v1 := mgl32.Vec3{worldX + 1, worldY, worldZ}
		v2 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ}
		v3 := mgl32.Vec3{worldX, worldY + 1, worldZ}
		v4 := mgl32.Vec3{worldX, worldY, worldZ}
		normal := mgl32.Vec3{0, 0, -1}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}

	// Right face (+X)
	if c.shouldRenderFace(x+1, y, z, blockType) {
		v1 := mgl32.Vec3{worldX + 1, worldY, worldZ + 1}
		v2 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ + 1}
		v3 := mgl32.Vec3{worldX + 1, worldY + 1, worldZ}
		v4 := mgl32.Vec3{worldX + 1, worldY, worldZ}
		normal := mgl32.Vec3{1, 0, 0}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}

	// Left face (-X)
	if c.shouldRenderFace(x-1, y, z, blockType) {
		v1 := mgl32.Vec3{worldX, worldY, worldZ}
		v2 := mgl32.Vec3{worldX, worldY + 1, worldZ}
		v3 := mgl32.Vec3{worldX, worldY + 1, worldZ + 1}
		v4 := mgl32.Vec3{worldX, worldY, worldZ + 1}
		normal := mgl32.Vec3{-1, 0, 0}
		c.mesh.AddQuad(v1, v2, v3, v4, color, normal, 1.0)
	}
}

// shouldRenderFace determines if a face should be rendered
func (c *Chunk) shouldRenderFace(x, y, z int, blockType BlockType) bool {
	// Get the adjacent block
	adjBlockType := c.GetBlock(x, y, z)

	// If the adjacent block is air, always render the face
	if adjBlockType == Air {
		return true
	}

	// If the adjacent block is transparent and not the same type as this block,
	// render the face
	return c.blockRegistry.IsTransparent(adjBlockType) && adjBlockType != blockType
}

// Render renders the chunk
func (c *Chunk) Render() {
	// Build mesh if needed
	if c.needsUpdate {
		c.BuildMesh()
	}

	// Draw mesh
	if c.mesh != nil {
		c.mesh.Draw()
	}
}

// Delete frees all resources used by the chunk
func (c *Chunk) Delete() {
	if c.mesh != nil {
		c.mesh.Delete()
		c.mesh = nil
	}
}
