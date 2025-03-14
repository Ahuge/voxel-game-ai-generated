package main

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Mesh represents a 3D mesh with vertices, indices, and OpenGL buffers
type Mesh struct {
	vertices  []float32
	indices   []uint32
	colors    []float32
	texCoords []float32
	normals   []float32

	vao uint32
	vbo uint32
	ebo uint32
	cbo uint32
	tbo uint32
	nbo uint32

	indexCount int32
}

// NewMesh creates a new mesh instance
func NewMesh() *Mesh {
	mesh := &Mesh{}

	// Generate VAO
	gl.GenVertexArrays(1, &mesh.vao)

	// Generate buffers
	gl.GenBuffers(1, &mesh.vbo) // Vertex buffer
	gl.GenBuffers(1, &mesh.ebo) // Element buffer
	gl.GenBuffers(1, &mesh.cbo) // Color buffer
	gl.GenBuffers(1, &mesh.tbo) // Texture coordinate buffer
	gl.GenBuffers(1, &mesh.nbo) // Normal buffer

	return mesh
}

// AddVertex adds a vertex to the mesh
func (m *Mesh) AddVertex(x, y, z float32) {
	m.vertices = append(m.vertices, x, y, z)
}

// AddIndex adds an index to the mesh
func (m *Mesh) AddIndex(index uint32) {
	m.indices = append(m.indices, index)
}

// AddColor adds a color to the mesh
func (m *Mesh) AddColor(r, g, b, a float32) {
	m.colors = append(m.colors, r, g, b, a)
}

// AddTexCoord adds a texture coordinate to the mesh
func (m *Mesh) AddTexCoord(u, v float32) {
	m.texCoords = append(m.texCoords, u, v)
}

// AddNormal adds a normal to the mesh
func (m *Mesh) AddNormal(nx, ny, nz float32) {
	m.normals = append(m.normals, nx, ny, nz)
}

// AddQuad adds a quad (two triangles) to the mesh
func (m *Mesh) AddQuad(v1, v2, v3, v4 mgl32.Vec3, color mgl32.Vec3, normal mgl32.Vec3, alpha float32) {
	// Get current vertex count
	baseIndex := uint32(len(m.vertices) / 3)

	// Add vertices
	m.AddVertex(v1.X(), v1.Y(), v1.Z())
	m.AddVertex(v2.X(), v2.Y(), v2.Z())
	m.AddVertex(v3.X(), v3.Y(), v3.Z())
	m.AddVertex(v4.X(), v4.Y(), v4.Z())

	// Add colors
	for i := 0; i < 4; i++ {
		m.AddColor(color.X(), color.Y(), color.Z(), alpha)
	}

	// Add normals
	for i := 0; i < 4; i++ {
		m.AddNormal(normal.X(), normal.Y(), normal.Z())
	}

	// Add texture coordinates
	m.AddTexCoord(0, 0)
	m.AddTexCoord(1, 0)
	m.AddTexCoord(1, 1)
	m.AddTexCoord(0, 1)

	// Add indices for two triangles - fixed winding order to be consistent with CCW
	m.AddIndex(baseIndex)
	m.AddIndex(baseIndex + 1)
	m.AddIndex(baseIndex + 2)

	m.AddIndex(baseIndex)
	m.AddIndex(baseIndex + 2)
	m.AddIndex(baseIndex + 3)
}

// Clear clears all mesh data
func (m *Mesh) Clear() {
	m.vertices = m.vertices[:0]
	m.indices = m.indices[:0]
	m.colors = m.colors[:0]
	m.texCoords = m.texCoords[:0]
	m.normals = m.normals[:0]
	m.indexCount = 0
}

// Upload uploads mesh data to GPU
func (m *Mesh) Upload() {
	// Update index count
	m.indexCount = int32(len(m.indices))

	// Bind VAO
	gl.BindVertexArray(m.vao)

	// Upload vertices
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.vertices)*4, gl.Ptr(m.vertices), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	// Upload colors
	gl.BindBuffer(gl.ARRAY_BUFFER, m.cbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.colors)*4, gl.Ptr(m.colors), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 4, gl.FLOAT, false, 0, nil)

	// Upload texture coordinates
	gl.BindBuffer(gl.ARRAY_BUFFER, m.tbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.texCoords)*4, gl.Ptr(m.texCoords), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, 0, nil)

	// Upload normals
	gl.BindBuffer(gl.ARRAY_BUFFER, m.nbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(m.normals)*4, gl.Ptr(m.normals), gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(3)
	gl.VertexAttribPointer(3, 3, gl.FLOAT, false, 0, nil)

	// Upload indices
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(m.indices)*4, gl.Ptr(m.indices), gl.STATIC_DRAW)

	// Unbind VAO
	gl.BindVertexArray(0)
}

// Draw draws the mesh
func (m *Mesh) Draw() {
	if m.indexCount == 0 {
		return
	}

	// Bind VAO
	gl.BindVertexArray(m.vao)

	// Draw elements
	gl.DrawElements(gl.TRIANGLES, m.indexCount, gl.UNSIGNED_INT, nil)

	// Unbind VAO
	gl.BindVertexArray(0)
}

// Delete deletes all OpenGL resources
func (m *Mesh) Delete() {
	gl.DeleteBuffers(1, &m.vbo)
	gl.DeleteBuffers(1, &m.ebo)
	gl.DeleteBuffers(1, &m.cbo)
	gl.DeleteBuffers(1, &m.tbo)
	gl.DeleteBuffers(1, &m.nbo)
	gl.DeleteVertexArrays(1, &m.vao)
}
