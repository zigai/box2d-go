// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
)

func (draw *Draw) DrawScreenString(x, y float32, color box2d.HexColor, format string, args ...any) {
	text := fmt.Sprintf(format, args...)

	col := imCol32(uint8(color>>16), uint8(color>>8), uint8(color), 255)

	var ascent float32
	if baked := imgui.GetFontBaked(); baked != nil {
		ascent = baked.Ascent()
	}
	vp := imgui.MainViewport()
	pos := vp.Pos()
	imgui.BackgroundDrawListV(vp).AddTextVec2(imgui.Vec2{X: pos.X + x, Y: pos.Y + y - ascent}, col, text)
}

func (draw *Draw) DrawString(camera *Camera, p box2d.Pos, color box2d.HexColor, format string, args ...any) {
	ps := ConvertWorldToScreen(camera, p)
	draw.DrawScreenString(ps.X, ps.Y, color, format, args...)
}
