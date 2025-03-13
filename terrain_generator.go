package main

import (
	"math"

	"github.com/ojrac/opensimplex-go"
)

// BiomeType represents different biome types in the world
type BiomeType int

// Biome type constants
const (
	BiomePlains BiomeType = iota
	BiomeMountains
	BiomeDesert
	BiomeForest
	BiomeOcean
	NumBiomeTypes
)

// BiomeInfo contains information about a biome
type BiomeInfo struct {
	Name           string
	BaseHeight     float64
	HeightVariation float64
	TopBlock       byte
	FillerBlock    byte
	UnderwaterBlock byte
}

// TerrainGenerator handles terrain generation
type TerrainGenerator struct {
	noise        opensimplex.Noise
	biomeNoise   opensimplex.Noise
	biomes       [NumBiomeTypes]BiomeInfo
	seed         int64
	seaLevel     int
}

// NewTerrainGenerator creates a new terrain generator with the given seed
func NewTerrainGenerator(seed int64) *TerrainGenerator {
	gen := &TerrainGenerator{
		noise:      opensimplex.New(seed),
		biomeNoise: opensimplex.New(seed + 1), // Use a different seed for biome noise
		seed:       seed,
		seaLevel:   64, // Set sea level at half of world height
	}

	// Initialize biome types
	gen.biomes[BiomePlains] = BiomeInfo{
		Name:           "Plains",
		BaseHeight:     0.1,
		HeightVariation: 0.05,
		TopBlock:       Grass,
		FillerBlock:    Dirt,
		UnderwaterBlock: Sand,
	}

	gen.biomes[BiomeMountains] = BiomeInfo{
		Name:           "Mountains",
		BaseHeight:     0.5,
		HeightVariation: 0.5,
		TopBlock:       Stone,
		FillerBlock:    Stone,
		UnderwaterBlock: Gravel,
	}

	gen.biomes[BiomeDesert] = BiomeInfo{
		Name:           "Desert",
		BaseHeight:     0.1,
		HeightVariation: 0.05,
		TopBlock:       Sand,
		FillerBlock:    Sand,
		UnderwaterBlock: Sand,
	}

	gen.biomes[BiomeForest] = BiomeInfo{
		Name:           "Forest",
		BaseHeight:     0.2,
		HeightVariation: 0.1,
		TopBlock:       Grass,
		FillerBlock:    Dirt,
		UnderwaterBlock: Dirt,
	}

	gen.biomes[BiomeOcean] = BiomeInfo{
		Name:           "Ocean",
		BaseHeight:     -0.5,
		HeightVariation: 0.05,
		TopBlock:       Sand,
		FillerBlock:    Sand,
		UnderwaterBlock: Sand,
	}

	return gen
}

// GenerateChunkTerrain fills a chunk with terrain blocks
func (g *TerrainGenerator) GenerateChunkTerrain(chunk *Chunk) {
	// Generate terrain for each column in the chunk
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			// Calculate world coordinates
			worldX := float64(chunk.worldPos.X*ChunkSize + x)
			worldZ := float64(chunk.worldPos.Z*ChunkSize + z)

			// Get biome at this position
			biome := g.getBiomeAt(worldX, worldZ)

			// Generate terrain height
			height := g.generateTerrainHeight(worldX, worldZ, biome)

			// Fill blocks in this column
			g.generateTerrainColumn(chunk, x, z, height, biome)
		}
	}
}

// getBiomeAt determines the biome at the given coordinates
func (g *TerrainGenerator) getBiomeAt(x, z float64) BiomeInfo {
	// Scale coordinates for biome noise
	x1, z1 := x*0.001, z*0.001

	// Use noise to blend between biomes
	humidity := (g.biomeNoise.Eval2(x1, z1) + 1.0) * 0.5
	temperature := (g.biomeNoise.Eval2(x1+1000, z1+1000) + 1.0) * 0.5

	// Determine biome based on humidity and temperature
	if temperature < 0.3 {
		// Cold biomes
		if humidity < 0.3 {
			return g.biomes[BiomeMountains]
		} else {
			return g.biomes[BiomeForest]
		}
	} else if temperature < 0.7 {
		// Temperate biomes
		if humidity < 0.4 {
			return g.biomes[BiomePlains]
		} else if humidity < 0.8 {
			return g.biomes[BiomeForest]
		} else {
			return g.biomes[BiomePlains]
		}
	} else {
		// Hot biomes
		if humidity < 0.2 {
			return g.biomes[BiomeDesert]
		} else if humidity < 0.6 {
			return g.biomes[BiomePlains]
		} else {
			return g.biomes[BiomeForest]
		}
	}
}

// generateTerrainHeight calculates the terrain height at the given coordinates
func (g *TerrainGenerator) generateTerrainHeight(x, z float64, biome BiomeInfo) int {
	// Scale coordinates for different noise layers
	x1, z1 := x*0.001, z*0.001  // Continent shape - very large scale
	x2, z2 := x*0.01, z*0.01    // Base terrain - large scale features
	x3, z3 := x*0.05, z*0.05    // Hills - medium scale features
	x4, z4 := x*0.2, z*0.2      // Details - small scale features

	// Continent shape (very large scale features)
	continentShape := (g.noise.Eval2(x1, z1) + 1.0) * 0.5

	// Apply continent edge falloff
	distFromCenter := math.Sqrt(x*x+z*z) * 0.0005
	continentFactor := 1.0 / (1.0 + distFromCenter*distFromCenter*0.1)
	continentShape = continentShape*0.8 + continentFactor*0.2

	// Ocean depth
	oceanDepth := 0.0
	if continentShape < 0.3 {
		// Create oceans where continent shape is low
		oceanDepth = (0.3 - continentShape) * 3.0
		// Use ocean biome for deep water
		if oceanDepth > 0.5 {
			biome = g.biomes[BiomeOcean]
		}
	}

	// Base terrain (large scale features)
	baseNoise := g.noise.Eval2(x2, z2)

	// Hills (medium scale features)
	hillsNoise := g.noise.Eval2(x3, z3) * 0.5

	// Details (small scale features)
	detailsNoise := g.noise.Eval2(x4, z4) * 0.25

	// Combine noise layers with proper weighting
	value := baseNoise + hillsNoise + detailsNoise

	// Normalize to 0-1 range
	value = (value + 1.0) * 0.5

	// Apply biome-specific height adjustments
	value = value * biome.HeightVariation + biome.BaseHeight

	// Apply ocean depth
	value -= oceanDepth

	// Convert to height value (0 to WorldHeight-1)
	height := int(value * float64(WorldHeight*0.8))

	// Ensure height is at least 1 for land, or seaLevel-10 for ocean
	minHeight := 1
	if biome.Name == "Ocean" {
		minHeight = g.seaLevel - 10
	}

	// Clamp to valid range
	if height < minHeight {
		height = minHeight
	}
	if height >= WorldHeight {
		height = WorldHeight - 1
	}

	return height
}

// generateTerrainColumn fills a column of blocks based on the terrain height
func (g *TerrainGenerator) generateTerrainColumn(chunk *Chunk, x, z, height int, biome BiomeInfo) {
	// Bedrock layer at the bottom
	chunk.blocks[x][0][z] = Bedrock

	// Fill the rest of the column
	for y := 1; y < WorldHeight; y++ {
		if y <= height {
			// Below terrain surface
			if y == height {
				// Surface block
				if y < g.seaLevel {
					// Underwater surface
					chunk.blocks[x][y][z] = biome.UnderwaterBlock
				} else {
					// Above water surface
					chunk.blocks[x][y][z] = biome.TopBlock
				}
			} else if y > height - 4 {
				// Filler layer (3 blocks deep)
				chunk.blocks[x][y][z] = biome.FillerBlock
			} else if y > height - 8 {
				// Transition layer (mix of filler and stone)
				if (x+y+z) % 2 == 0 {
					chunk.blocks[x][y][z] = biome.FillerBlock
				} else {
					chunk.blocks[x][y][z] = Stone
				}
			} else {
				// Stone layer
				chunk.blocks[x][y][z] = Stone
			}
		} else if y <= g.seaLevel {
			// Water blocks up to sea level
			// chunk.blocks[x][y][z] = Water
		} else {
			// Air above terrain
			chunk.blocks[x][y][z] = Air
		}
	}
}

// GenerateStructures adds structures like trees, caves, etc. to a chunk
func (g *TerrainGenerator) GenerateStructures(chunk *Chunk) {
	// This is a placeholder for future structure generation
	// For now, we'll just add some simple trees in forest biomes
	for x := 2; x < ChunkSize-2; x += 4 {
		for z := 2; z < ChunkSize-2; z += 4 {
			// Calculate world coordinates
			worldX := float64(chunk.worldPos.X*ChunkSize + x)
			worldZ := float64(chunk.worldPos.Z*ChunkSize + z)

			// Get biome at this position
			biome := g.getBiomeAt(worldX, worldZ)

			// Only generate trees in forest biomes
			if biome.Name == "Forest" {
				// Use noise to determine if we should place a tree here
				treeNoise := g.noise.Eval2(worldX*0.1, worldZ*0.1)
				if treeNoise > 0.7 {
					// Find the surface height
					for y := WorldHeight - 1; y >= 0; y-- {
						if chunk.blocks[x][y][z] == Grass {
							// Generate a simple tree
							g.generateTree(chunk, x, y+1, z)
							break
						}
					}
				}
			}
		}
	}
}

// generateTree creates a simple tree at the given position
func (g *TerrainGenerator) generateTree(chunk *Chunk, x, y, z int) {
	// Check if there's enough space for the tree
	if y+4 >= WorldHeight {
		return
	}

	// Generate trunk (3 blocks tall)
	for i := 0; i < 3; i++ {
		chunk.blocks[x][y+i][z] = Wood
	}

	// Generate leaves (a simple 3x3x3 cube)
	for lx := -1; lx <= 1; lx++ {
		for ly := 0; ly <= 2; ly++ {
			for lz := -1; lz <= 1; lz++ {
				// Skip the center column (trunk)
				if lx == 0 && lz == 0 {
					continue
				}
				
				// Calculate leaf position
				leafX := x + lx
				leafY := y + 1 + ly // Start leaves at y+1
				leafZ := z + lz
				
				// Check bounds
				if leafX >= 0 && leafX < ChunkSize && 
				   leafY >= 0 && leafY < WorldHeight && 
				   leafZ >= 0 && leafZ < ChunkSize {
					// Only place leaves where there's air
					if chunk.blocks[leafX][leafY][leafZ] == Air {
						chunk.blocks[leafX][leafY][leafZ] = Leaves
					}
				}
			}
		}
	}
}