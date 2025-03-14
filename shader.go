package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Shader represents an OpenGL shader program
type Shader struct {
	ID uint32
}

// NewShader creates a new shader program from vertex and fragment shader files
func NewShader(vertexPath, fragmentPath string) *Shader {
	// Read vertex shader code from file
	vertexCode, err := ioutil.ReadFile(vertexPath)
	if err != nil {
		panic(fmt.Errorf("failed to read vertex shader file: %v", err))
	}

	// Read fragment shader code from file
	fragmentCode, err := ioutil.ReadFile(fragmentPath)
	if err != nil {
		panic(fmt.Errorf("failed to read fragment shader file: %v", err))
	}

	// Convert to string
	vertexCodeStr := string(vertexCode)
	fragmentCodeStr := string(fragmentCode)

	// Compile shaders
	vertexShader := compileShader(vertexCodeStr, gl.VERTEX_SHADER)
	fragmentShader := compileShader(fragmentCodeStr, gl.FRAGMENT_SHADER)

	// Create shader program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check for linking errors
	var success int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		panic(fmt.Errorf("failed to link shader program: %v", log))
	}

	// Delete shaders as they're linked into the program and no longer needed
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &Shader{ID: program}
}

// Use activates the shader program
func (s *Shader) Use() {
	gl.UseProgram(s.ID)
}

// SetBool sets a boolean uniform value
func (s *Shader) SetBool(name string, value bool) {
	var intValue int32
	if value {
		intValue = 1
	}
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), intValue)
}

// SetInt sets an integer uniform value
func (s *Shader) SetInt(name string, value int32) {
	gl.Uniform1i(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value)
}

// SetFloat sets a float uniform value
func (s *Shader) SetFloat(name string, value float32) {
	gl.Uniform1f(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), value)
}

// SetVec3 sets a vec3 uniform value
func (s *Shader) SetVec3(name string, value mgl32.Vec3) {
	gl.Uniform3fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), 1, &value[0])
}

// SetMat4 sets a mat4 uniform value
func (s *Shader) SetMat4(name string, value mgl32.Mat4) {
	gl.UniformMatrix4fv(gl.GetUniformLocation(s.ID, gl.Str(name+"\x00")), 1, false, &value[0])
}

// compileShader compiles a shader from source code
func compileShader(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	// Check for compilation errors
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		shaderTypeStr := "vertex"
		if shaderType == gl.FRAGMENT_SHADER {
			shaderTypeStr = "fragment"
		}

		panic(fmt.Errorf("failed to compile %s shader: %v", shaderTypeStr, log))
	}

	return shader
}
