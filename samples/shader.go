// SPDX-FileCopyrightText: 2024 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

func checkOpenGL() {
	errCode := gl.GetError()
	if errCode != gl.NO_ERROR {
		panic(fmt.Sprintf("OpenGL error = %d", errCode))
	}
}

func printLogGL(object uint32) {
	var logLength int32
	switch {
	case gl.IsShader(object):
		gl.GetShaderiv(object, gl.INFO_LOG_LENGTH, &logLength)
	case gl.IsProgram(object):
		gl.GetProgramiv(object, gl.INFO_LOG_LENGTH, &logLength)
	default:
		fmt.Printf("PrintLogGL: Not a shader or a program\n")
		return
	}

	log := strings.Repeat("\x00", int(logLength)+1)
	if gl.IsShader(object) {
		gl.GetShaderInfoLog(object, logLength, nil, gl.Str(log))
	} else if gl.IsProgram(object) {
		gl.GetProgramInfoLog(object, logLength, nil, gl.Str(log))
	}

	fmt.Printf("PrintLogGL: %s", strings.TrimRight(log, "\x00"))
}

func createShaderFromString(source string, shaderType uint32) uint32 {
	shader := gl.CreateShader(shaderType)
	sources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, sources, nil)
	free()
	gl.CompileShader(shader)

	var success int32 = gl.FALSE
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)

	if success == gl.FALSE {
		fmt.Printf("Error compiling shader of type %d!\n", shaderType)
		printLogGL(shader)
		gl.DeleteShader(shader)
		return 0
	}

	return shader
}

func createProgramFromStrings(vertexString, fragmentString string) uint32 {
	vertex := createShaderFromString(vertexString, gl.VERTEX_SHADER)
	if vertex == 0 {
		return 0
	}

	fragment := createShaderFromString(fragmentString, gl.FRAGMENT_SHADER)
	if fragment == 0 {
		return 0
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertex)
	gl.AttachShader(program, fragment)

	gl.LinkProgram(program)

	var success int32 = gl.FALSE
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		fmt.Printf("glLinkProgram:")
		printLogGL(program)
		return 0
	}

	gl.DeleteShader(vertex)
	gl.DeleteShader(fragment)

	return program
}
