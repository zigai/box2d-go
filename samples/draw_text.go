// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

func (draw *Draw) DrawScreenStringf(
	x float32,
	y float32,
	color box2d.HexColor,
	format string,
	args ...any,
) {
	const (
		redShift   = 16
		greenShift = 8
	)

	text := fmt.Sprintf(format, args...)

	col := imCol32(uint8(color>>redShift), uint8(color>>greenShift), uint8(color))

	var ascent float32
	if baked := imgui.GetFontBaked(); baked != nil {
		ascent = baked.Ascent()
	}

	vp := imgui.MainViewport()
	pos := vp.Pos()
	imgui.BackgroundDrawListV(vp).
		AddTextVec2(imgui.Vec2{X: pos.X + x, Y: pos.Y + y - ascent}, col, text)
}

func (draw *Draw) DrawStringf(
	camera *Camera,
	p box2d.Pos,
	color box2d.HexColor,
	format string,
	args ...any,
) {
	ps := ConvertWorldToScreen(camera, p)
	draw.DrawScreenStringf(ps.X, ps.Y, color, format, args...)
}
