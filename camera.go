package main

import (
	"math"

	// "github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Camera constants
const (
	YAW         = -90.0
	PITCH       = 0.0
	SPEED       = 30.0
	SENSITIVITY = 0.1
	ZOOM        = 45.0
)

// Camera represents a first-person camera
type Camera struct {
	// Camera attributes
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	Right    mgl32.Vec3
	WorldUp  mgl32.Vec3

	// Euler angles
	Yaw   float32
	Pitch float32

	// Camera options
	MovementSpeed    float32
	MouseSensitivity float32
	Zoom             float32
}

// NewCamera creates a new camera instance
func NewCamera(position, up mgl32.Vec3, yaw, pitch float32) *Camera {
	camera := &Camera{
		Position:         position,
		WorldUp:          up,
		Yaw:              yaw,
		Pitch:            pitch,
		MovementSpeed:    SPEED,
		MouseSensitivity: SENSITIVITY,
		Zoom:             ZOOM,
	}

	// Initialize vectors
	camera.updateCameraVectors()

	return camera
}

// GetViewMatrix returns the view matrix calculated using Euler angles
func (c *Camera) GetViewMatrix() mgl32.Mat4 {
	return mgl32.LookAtV(c.Position, c.Position.Add(c.Front), c.Up)
}

// ProcessKeyboard processes keyboard input for camera movement
func (c *Camera) ProcessKeyboard(direction CameraMovement, deltaTime float32) {
	velocity := c.MovementSpeed * deltaTime

	switch direction {
	case FORWARD:
		c.Position = c.Position.Add(c.Front.Mul(velocity))
	case BACKWARD:
		c.Position = c.Position.Sub(c.Front.Mul(velocity))
	case LEFT:
		c.Position = c.Position.Sub(c.Right.Mul(velocity))
	case RIGHT:
		c.Position = c.Position.Add(c.Right.Mul(velocity))
	case UP:
		c.Position = c.Position.Add(c.WorldUp.Mul(velocity))
	case DOWN:
		c.Position = c.Position.Sub(c.WorldUp.Mul(velocity))
	}
}

// ProcessMouseMovement processes mouse input for camera rotation
func (c *Camera) ProcessMouseMovement(xoffset, yoffset float32, constrainPitch bool) {
	xoffset *= c.MouseSensitivity
	yoffset *= c.MouseSensitivity

	c.Yaw += xoffset
	c.Pitch += yoffset

	// Constrain pitch
	if constrainPitch {
		if c.Pitch > 89.0 {
			c.Pitch = 89.0
		}
		if c.Pitch < -89.0 {
			c.Pitch = -89.0
		}
	}

	// Update camera vectors
	c.updateCameraVectors()
}

// ProcessMouseScroll processes mouse scroll input for camera zoom
func (c *Camera) ProcessMouseScroll(yoffset float32) {
	c.Zoom -= yoffset
	if c.Zoom < 1.0 {
		c.Zoom = 1.0
	}
	if c.Zoom > 45.0 {
		c.Zoom = 45.0
	}
}

// updateCameraVectors calculates the front, right and up vectors from the camera's Euler angles
func (c *Camera) updateCameraVectors() {
	// Calculate the new front vector
	front := mgl32.Vec3{
		float32(math.Cos(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.Pitch)))),
		float32(math.Sin(float64(mgl32.DegToRad(c.Yaw))) * math.Cos(float64(mgl32.DegToRad(c.Pitch)))),
	}
	c.Front = front.Normalize()

	// Re-calculate the right and up vectors
	c.Right = c.Front.Cross(c.WorldUp).Normalize()
	c.Up = c.Right.Cross(c.Front).Normalize()
}

// CameraMovement defines camera movement directions
type CameraMovement int

// Camera movement constants
const (
	FORWARD CameraMovement = iota
	BACKWARD
	LEFT
	RIGHT
	UP
	DOWN
)