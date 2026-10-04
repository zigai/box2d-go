// SPDX-License-Identifier: MIT
package box2d

import (
	"github.com/zigai/box2d/internal/oracle"
	"testing"
)

func TestWorldDrawTypedParity(t *testing.T) {
	frames := []struct {
		name string
		base Pos
	}{{"Origin", Pos{}}, {"Offset", Pos{32, -16}}}
	if IsDoublePrecision() {
		frames = append(frames, struct {
			name string
			base Pos
		}{"Far", Pos{10000000, -10000000}})
	}
	for _, frame := range frames {
		t.Run(frame.name, func(t *testing.T) {
			r := oracle.Open(t, "draw/"+frame.name)
			base := frame.base
			wd := DefaultWorldDef()
			wd.WorkerCount = 1
			w := NewWorld(&wd)
			BTWorldCleanup(t, w)
			inspect := func() { spatialInspect(r, w); r.I32("awakeBodies", int32(w.AwakeBodyCount())) }
			inspect()
			var bodies [6]Body
			for i := range 6 {
				bd := DefaultBodyDef()
				offset := Vec2{2 * float32(i), 2}
				if i == 0 {
					offset = Vec2{0, -.5}
				}
				if i == 1 {
					offset = Vec2{0, .45}
				}
				bd.Position = base.Add(offset)
				if i == 1 || i == 2 {
					bd.Type = DynamicBody
				}
				if i == 3 {
					bd.Type = KinematicBody
				}
				bd.Name = "fixture"
				if i == 1 {
					bd.Name = "contact box"
				}
				if i == 2 {
					bd.Name = "capsule"
					bd.Rotation = MakeRot(.25)
				}
				bodies[i] = w.CreateBody(&bd)
				sd := DefaultShapeDef()
				switch i {
				case 0, 1:
					halfWidth := float32(.5)
					if i == 0 {
						halfWidth = 3
					}
					p := MakeBox(halfWidth, .5)
					bodies[i].CreatePolygonShape(&sd, &p)
				case 2:
					c := Capsule{Vec2{-.3, 0}, Vec2{.7, 0}, .3}
					bodies[i].CreateCapsuleShape(&sd, &c)
				case 3:
					sd.IsSensor = true
					c := Circle{Vec2{.2, .1}, .6}
					bodies[i].CreateCircleShape(&sd, &c)
				case 4:
					sd.Material.CustomColor = 0x123456
					s := Segment{Vec2{-1, 0}, Vec2{1, .5}}
					bodies[i].CreateSegmentShape(&sd, &s)
				case 5:
					cd := DefaultChainDef()
					cd.Points = []Vec2{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}}
					cd.IsLoop = true
					bodies[i].CreateChain(&cd)
				}
				inspect()
			}
			jd := DefaultDistanceJointDef()
			jd.Base.BodyA = bodies[1]
			jd.Base.BodyB = bodies[2]
			jd.Length = 4
			jd.EnableSpring = true
			jd.Hertz = 2
			jd.EnableLimit = true
			jd.MinLength = 2
			jd.MaxLength = 5
			w.CreateDistanceJoint(&jd)
			inspect()
			for range 3 {
				w.Step(float32(1)/60, 4)
				inspect()
			}
			for option := range 17 {
				d := DefaultDebugDraw()
				dFuncs := &BTFuncDrawer{}
				d.Drawer = dFuncs
				commands := 0
				command := func(kind int, color HexColor) {
					commands++
					r.I32("command", int32(kind))
					r.U32("color", uint32(color))
				}
				transform := func(xf WorldTransform) { r.Pos("position", xf.P.X, xf.P.Y); r.Rot("rotation", xf.Q.C, xf.Q.S) }
				point := func(p Pos) { r.Pos("point", p.X, p.Y) }
				vertices := func(v []Vec2) {
					r.I32("vertices", int32(len(v)))
					for _, p := range v {
						r.Vec2("vertex", p.X, p.Y)
					}
				}
				dFuncs.Polygon = func(xf WorldTransform, v []Vec2, c HexColor) { command(1, c); transform(xf); vertices(v) }
				dFuncs.SolidPolygon = func(xf WorldTransform, v []Vec2, radius float32, c HexColor) {
					command(2, c)
					transform(xf)
					vertices(v)
					r.F32("radius", radius)
				}
				dFuncs.Circle = func(p Pos, radius float32, c HexColor) { command(3, c); point(p); r.F32("radius", radius) }
				dFuncs.SolidCircle = func(xf WorldTransform, p Vec2, radius float32, c HexColor) {
					command(4, c)
					transform(xf)
					r.Vec2("center", p.X, p.Y)
					r.F32("radius", radius)
				}
				dFuncs.SolidCapsule = func(p, q Pos, radius float32, c HexColor) { command(5, c); point(p); point(q); r.F32("radius", radius) }
				dFuncs.Line = func(p, q Pos, c HexColor) { command(6, c); point(p); point(q) }
				dFuncs.TransformFunc = func(xf WorldTransform) { command(7, 0); transform(xf) }
				dFuncs.Point = func(p Pos, size float32, c HexColor) { command(8, c); point(p); r.F32("size", size) }
				dFuncs.String = func(p Pos, text string, c HexColor) { command(9, c); point(p); r.Bytes("text", []byte(text)) }
				dFuncs.Bounds = func(a AABB, c HexColor) {
					command(10, c)
					r.Vec2("lower", a.LowerBound.X, a.LowerBound.Y)
					r.Vec2("upper", a.UpperBound.X, a.UpperBound.Y)
				}
				d.DrawShapes = option == 0 || option == 1 || option == 15 || option == 16
				d.ForceScale = .3
				d.JointScale = .7
				d.DrawChainNormals = option == 1 || option == 15
				d.DrawJoints = option == 2 || option == 3 || option == 15
				d.DrawJointExtras = option == 3 || option == 15
				d.DrawBounds = option == 4 || option == 15
				d.DrawMass = option == 5 || option == 15
				d.DrawBodyNames = option == 6 || option == 15
				d.DrawContacts = (option >= 7 && option <= 13) || option == 15
				d.DrawAnchorA = option == 13
				d.DrawGraphColors = option == 8 || option == 15
				d.DrawContactFeatures = option == 9 || option == 15
				d.DrawContactNormals = option == 10
				d.DrawContactForces = option == 11 || option == 15
				d.DrawFrictionForces = option == 12 || option == 15
				d.DrawIslands = option == 14 || option == 15
				if option == 16 {
					center := base.Vec2()
					d.DrawingBounds = AABB{Vec2{center.X - 1, center.Y - 1}, Vec2{center.X + 1, center.Y + 1}}
				}
				before := w.StateHash()
				r.U64("before", before)
				w.Draw(&d)
				r.I32("commands", int32(commands))
				after := w.StateHash()
				r.U64("after", after)
				if before != after {
					t.Fatalf("draw option %d changed canonical state: %x -> %x", option, before, after)
				}
			}
			for n := range 6 {
				bodies[5-n].Destroy()
				inspect()
			}
		})
	}
}
