package main

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Block type constants
const (
	Air = iota
	Dirt
	Grass
	Stone
	Wood
	Leaves
	Water
	Sand
	Snow
	Gravel
	Clay
	Bedrock
	NumBlockTypes
)

// BlockFace represents a face of a block
type BlockFace int

// Block face constants
const (
	FaceNorth BlockFace = iota
	FaceSouth
	FaceEast
	FaceWest
	FaceTop
	FaceBottom
	NumFaces
)

// BlockInfo contains information about a block type
type BlockInfo struct {
	Name           string
	Solid          bool
	Transparent    bool
	TextureIndexes [NumFaces]int // Texture index for each face
	Color          mgl32.Vec3    // Base color for the block
}

// BlockRegistry stores information about all block types
type BlockRegistry struct {
	Blocks [NumBlockTypes]BlockInfo
}

// NewBlockRegistry creates a new block registry with default block types
func NewBlockRegistry() *BlockRegistry {
	registry := &BlockRegistry{}

	// Initialize block types
	registry.Blocks[Air] = BlockInfo{
		Name:        "Air",
		Solid:       false,
		Transparent: true,
		Color:       mgl32.Vec3{1.0, 1.0, 1.0},
	}

	registry.Blocks[Dirt] = BlockInfo{
		Name:        "Dirt",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.6, 0.3, 0.1},
	}

	registry.Blocks[Grass] = BlockInfo{
		Name:        "Grass",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.3, 0.7, 0.2},
	}

	registry.Blocks[Stone] = BlockInfo{
		Name:        "Stone",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.5, 0.5, 0.5},
	}

	registry.Blocks[Wood] = BlockInfo{
		Name:        "Wood",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.6, 0.4, 0.2},
	}

	registry.Blocks[Leaves] = BlockInfo{
		Name:        "Leaves",
		Solid:       true,
		Transparent: true,
		Color:       mgl32.Vec3{0.2, 0.6, 0.2},
	}

	registry.Blocks[Water] = BlockInfo{
		Name:        "Water",
		Solid:       false,
		Transparent: true,
		Color:       mgl32.Vec3{0.0, 0.3, 0.8},
	}

	registry.Blocks[Sand] = BlockInfo{
		Name:        "Sand",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.9, 0.8, 0.6},
	}

	registry.Blocks[Snow] = BlockInfo{
		Name:        "Snow",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.9, 0.9, 0.9},
	}

	registry.Blocks[Gravel] = BlockInfo{
		Name:        "Gravel",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.5, 0.5, 0.5},
	}

	registry.Blocks[Clay] = BlockInfo{
		Name:        "Clay",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.7, 0.7, 0.8},
	}

	registry.Blocks[Bedrock] = BlockInfo{
		Name:        "Bedrock",
		Solid:       true,
		Transparent: false,
		Color:       mgl32.Vec3{0.2, 0.2, 0.2},
	}

	return registry
}

// GetBlockInfo returns information about a block type
func (r *BlockRegistry) GetBlockInfo(blockType byte) BlockInfo {
	if blockType >= NumBlockTypes {
		return r.Blocks[Air]
	}
	return r.Blocks[blockType]
}

// IsSolid returns whether a block type is solid
func (r *BlockRegistry) IsSolid(blockType byte) bool {
	return r.GetBlockInfo(blockType).Solid
}

// IsTransparent returns whether a block type is transparent
func (r *BlockRegistry) IsTransparent(blockType byte) bool {
	return r.GetBlockInfo(blockType).Transparent
}