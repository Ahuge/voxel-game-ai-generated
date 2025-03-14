package main

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Frustum represents a view frustum for culling
type Frustum struct {
	planes [6]mgl32.Vec4 // Left, Right, Bottom, Top, Near, Far
	// Track previously visible chunks to prevent flickering
	visibleLastFrame map[ChunkPos]bool
	frameCounter     int
}

// NewFrustum creates a new view frustum from projection and view matrices
func NewFrustum(projection, view mgl32.Mat4) *Frustum {
	frustum := &Frustum{
		visibleLastFrame: make(map[ChunkPos]bool),
	}
	frustum.Update(projection, view)
	return frustum
}

// Update updates the frustum planes based on the current projection and view matrices
func (f *Frustum) Update(projection, view mgl32.Mat4) {
	// Combine view and projection matrices
	clip := projection.Mul4(view)

	// Left plane
	f.planes[0] = mgl32.Vec4{
		clip.At(0, 3) + clip.At(0, 0),
		clip.At(1, 3) + clip.At(1, 0),
		clip.At(2, 3) + clip.At(2, 0),
		clip.At(3, 3) + clip.At(3, 0),
	}

	// Right plane
	f.planes[1] = mgl32.Vec4{
		clip.At(0, 3) - clip.At(0, 0),
		clip.At(1, 3) - clip.At(1, 0),
		clip.At(2, 3) - clip.At(2, 0),
		clip.At(3, 3) - clip.At(3, 0),
	}

	// Bottom plane
	f.planes[2] = mgl32.Vec4{
		clip.At(0, 3) + clip.At(0, 1),
		clip.At(1, 3) + clip.At(1, 1),
		clip.At(2, 3) + clip.At(2, 1),
		clip.At(3, 3) + clip.At(3, 1),
	}

	// Top plane
	f.planes[3] = mgl32.Vec4{
		clip.At(0, 3) - clip.At(0, 1),
		clip.At(1, 3) - clip.At(1, 1),
		clip.At(2, 3) - clip.At(2, 1),
		clip.At(3, 3) - clip.At(3, 1),
	}

	// Near plane
	f.planes[4] = mgl32.Vec4{
		clip.At(0, 3) + clip.At(0, 2),
		clip.At(1, 3) + clip.At(1, 2),
		clip.At(2, 3) + clip.At(2, 2),
		clip.At(3, 3) + clip.At(3, 2),
	}

	// Far plane
	f.planes[5] = mgl32.Vec4{
		clip.At(0, 3) - clip.At(0, 2),
		clip.At(1, 3) - clip.At(1, 2),
		clip.At(2, 3) - clip.At(2, 2),
		clip.At(3, 3) - clip.At(3, 2),
	}

	// Normalize all planes
	for i := range f.planes {
		length := float32(math.Sqrt(float64(
			f.planes[i].X()*f.planes[i].X() +
				f.planes[i].Y()*f.planes[i].Y() +
				f.planes[i].Z()*f.planes[i].Z())))
		f.planes[i] = f.planes[i].Mul(1.0 / length)
	}
}

// IsBoxVisible checks if an axis-aligned bounding box is visible in the frustum
func (f *Frustum) IsBoxVisible(minX, minY, minZ, maxX, maxY, maxZ float32) bool {
	// Add a much larger margin to the bounding box to prevent popping when turning quickly
	// Increased from 32.0 to 64.0 to fix aggressive culling issues
	const margin float32 = 64.0
	minX -= margin
	minY -= margin
	minZ -= margin
	maxX += margin
	maxY += margin
	maxZ += margin

	// Use a more conservative approach for frustum culling
	// Only cull if the box is completely outside ALL corners for at least one plane
	for i := 0; i < 6; i++ {
		plane := f.planes[i]
		p := plane.Vec3()
		d := plane.W()

		// Calculate distances from all 8 corners to the plane
		dist1 := p.X()*minX + p.Y()*minY + p.Z()*minZ + d
		dist2 := p.X()*maxX + p.Y()*minY + p.Z()*minZ + d
		dist3 := p.X()*minX + p.Y()*maxY + p.Z()*minZ + d
		dist4 := p.X()*maxX + p.Y()*maxY + p.Z()*minZ + d
		dist5 := p.X()*minX + p.Y()*minY + p.Z()*maxZ + d
		dist6 := p.X()*maxX + p.Y()*minY + p.Z()*maxZ + d
		dist7 := p.X()*minX + p.Y()*maxY + p.Z()*maxZ + d
		dist8 := p.X()*maxX + p.Y()*maxY + p.Z()*maxZ + d

		// If all corners are outside this plane, the box is not visible
		if dist1 < 0 && dist2 < 0 && dist3 < 0 && dist4 < 0 &&
			dist5 < 0 && dist6 < 0 && dist7 < 0 && dist8 < 0 {
			return false
		}
	}

	// If we get here, the box is at least partially inside all planes
	return true
}

// IsChunkVisible checks if a chunk is visible in the frustum
func (f *Frustum) IsChunkVisible(chunk *Chunk) bool {
	// Calculate chunk bounds in world space
	minX := float32(chunk.worldPos.X * ChunkSize)
	minZ := float32(chunk.worldPos.Z * ChunkSize)
	maxX := minX + ChunkSize
	maxZ := minZ + ChunkSize

	// Use full world height for Y bounds
	minY := float32(0)
	maxY := float32(WorldHeight)

	// Check if the chunk was visible in the last frame
	if wasVisible, exists := f.visibleLastFrame[chunk.worldPos]; exists && wasVisible {
		// If it was visible in the last frame, keep it visible for a few frames
		// to prevent flickering during camera rotation
		f.visibleLastFrame[chunk.worldPos] = true
		return true
	}

	// Standard visibility check
	isVisible := f.IsBoxVisible(minX, minY, minZ, maxX, maxY, maxZ)

	// Update visibility history
	f.visibleLastFrame[chunk.worldPos] = isVisible

	// Every 10 frames, clean up the visibility history to prevent memory leaks
	f.frameCounter++
	if f.frameCounter >= 10 {
		f.frameCounter = 0

		// Remove chunks that have been invisible for a while
		for pos, visible := range f.visibleLastFrame {
			if !visible {
				delete(f.visibleLastFrame, pos)
			}
		}
	}

	return isVisible
}
