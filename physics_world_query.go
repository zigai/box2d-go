// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"fmt"
	"math"
	"math/bits"
	"strconv"
)

type (
	OverlapResultFunc func(shape Shape) bool
	CastResultFunc    func(shape Shape, point Pos, normal Vec2, fraction float32) float32
	PlaneResultFunc   func(shape Shape, plane *PlaneResult) bool
)

type RayResult struct {
	Shape      Shape
	Point      Pos
	Normal     Vec2
	Fraction   float32
	NodeVisits int
	LeafVisits int
	Hit        bool
}

type ExplosionDef struct {
	MaskBits         uint64
	Position         Pos
	Radius           float32
	Falloff          float32
	ImpulsePerLength float32
}

type HexColor int

const (
	ColorAliceBlue            HexColor = 0xF0F8FF
	ColorAntiqueWhite         HexColor = 0xFAEBD7
	ColorAqua                 HexColor = 0x00FFFF
	ColorAquamarine           HexColor = 0x7FFFD4
	ColorAzure                HexColor = 0xF0FFFF
	ColorBeige                HexColor = 0xF5F5DC
	ColorBisque               HexColor = 0xFFE4C4
	ColorBlack                HexColor = 0x000000
	ColorBlanchedAlmond       HexColor = 0xFFEBCD
	ColorBlue                 HexColor = 0x0000FF
	ColorBlueViolet           HexColor = 0x8A2BE2
	ColorBrown                HexColor = 0xA52A2A
	ColorBurlywood            HexColor = 0xDEB887
	ColorCadetBlue            HexColor = 0x5F9EA0
	ColorChartreuse           HexColor = 0x7FFF00
	ColorChocolate            HexColor = 0xD2691E
	ColorCoral                HexColor = 0xFF7F50
	ColorCornflowerBlue       HexColor = 0x6495ED
	ColorCornsilk             HexColor = 0xFFF8DC
	ColorCrimson              HexColor = 0xDC143C
	ColorCyan                 HexColor = 0x00FFFF
	ColorDarkBlue             HexColor = 0x00008B
	ColorDarkCyan             HexColor = 0x008B8B
	ColorDarkGoldenRod        HexColor = 0xB8860B
	ColorDarkGray             HexColor = 0xA9A9A9
	ColorDarkGreen            HexColor = 0x006400
	ColorDarkKhaki            HexColor = 0xBDB76B
	ColorDarkMagenta          HexColor = 0x8B008B
	ColorDarkOliveGreen       HexColor = 0x556B2F
	ColorDarkOrange           HexColor = 0xFF8C00
	ColorDarkOrchid           HexColor = 0x9932CC
	ColorDarkRed              HexColor = 0x8B0000
	ColorDarkSalmon           HexColor = 0xE9967A
	ColorDarkSeaGreen         HexColor = 0x8FBC8F
	ColorDarkSlateBlue        HexColor = 0x483D8B
	ColorDarkSlateGray        HexColor = 0x2F4F4F
	ColorDarkTurquoise        HexColor = 0x00CED1
	ColorDarkViolet           HexColor = 0x9400D3
	ColorDeepPink             HexColor = 0xFF1493
	ColorDeepSkyBlue          HexColor = 0x00BFFF
	ColorDimGray              HexColor = 0x696969
	ColorDodgerBlue           HexColor = 0x1E90FF
	ColorFireBrick            HexColor = 0xB22222
	ColorFloralWhite          HexColor = 0xFFFAF0
	ColorForestGreen          HexColor = 0x228B22
	ColorFuchsia              HexColor = 0xFF00FF
	ColorGainsboro            HexColor = 0xDCDCDC
	ColorGhostWhite           HexColor = 0xF8F8FF
	ColorGold                 HexColor = 0xFFD700
	ColorGoldenRod            HexColor = 0xDAA520
	ColorGray                 HexColor = 0x808080
	ColorGreen                HexColor = 0x008000
	ColorGreenYellow          HexColor = 0xADFF2F
	ColorHoneyDew             HexColor = 0xF0FFF0
	ColorHotPink              HexColor = 0xFF69B4
	ColorIndianRed            HexColor = 0xCD5C5C
	ColorIndigo               HexColor = 0x4B0082
	ColorIvory                HexColor = 0xFFFFF0
	ColorKhaki                HexColor = 0xF0E68C
	ColorLavender             HexColor = 0xE6E6FA
	ColorLavenderBlush        HexColor = 0xFFF0F5
	ColorLawnGreen            HexColor = 0x7CFC00
	ColorLemonChiffon         HexColor = 0xFFFACD
	ColorLightBlue            HexColor = 0xADD8E6
	ColorLightCoral           HexColor = 0xF08080
	ColorLightCyan            HexColor = 0xE0FFFF
	ColorLightGoldenRodYellow HexColor = 0xFAFAD2
	ColorLightGray            HexColor = 0xD3D3D3
	ColorLightGreen           HexColor = 0x90EE90
	ColorLightPink            HexColor = 0xFFB6C1
	ColorLightSalmon          HexColor = 0xFFA07A
	ColorLightSeaGreen        HexColor = 0x20B2AA
	ColorLightSkyBlue         HexColor = 0x87CEFA
	ColorLightSlateGray       HexColor = 0x778899
	ColorLightSteelBlue       HexColor = 0xB0C4DE
	ColorLightYellow          HexColor = 0xFFFFE0
	ColorLime                 HexColor = 0x00FF00
	ColorLimeGreen            HexColor = 0x32CD32
	ColorLinen                HexColor = 0xFAF0E6
	ColorMagenta              HexColor = 0xFF00FF
	ColorMaroon               HexColor = 0x800000
	ColorMediumAquaMarine     HexColor = 0x66CDAA
	ColorMediumBlue           HexColor = 0x0000CD
	ColorMediumOrchid         HexColor = 0xBA55D3
	ColorMediumPurple         HexColor = 0x9370DB
	ColorMediumSeaGreen       HexColor = 0x3CB371
	ColorMediumSlateBlue      HexColor = 0x7B68EE
	ColorMediumSpringGreen    HexColor = 0x00FA9A
	ColorMediumTurquoise      HexColor = 0x48D1CC
	ColorMediumVioletRed      HexColor = 0xC71585
	ColorMidnightBlue         HexColor = 0x191970
	ColorMintCream            HexColor = 0xF5FFFA
	ColorMistyRose            HexColor = 0xFFE4E1
	ColorMoccasin             HexColor = 0xFFE4B5
	ColorNavajoWhite          HexColor = 0xFFDEAD
	ColorNavy                 HexColor = 0x000080
	ColorOldLace              HexColor = 0xFDF5E6
	ColorOlive                HexColor = 0x808000
	ColorOliveDrab            HexColor = 0x6B8E23
	ColorOrange               HexColor = 0xFFA500
	ColorOrangeRed            HexColor = 0xFF4500
	ColorOrchid               HexColor = 0xDA70D6
	ColorPaleGoldenRod        HexColor = 0xEEE8AA
	ColorPaleGreen            HexColor = 0x98FB98
	ColorPaleTurquoise        HexColor = 0xAFEEEE
	ColorPaleVioletRed        HexColor = 0xDB7093
	ColorPapayaWhip           HexColor = 0xFFEFD5
	ColorPeachPuff            HexColor = 0xFFDAB9
	ColorPeru                 HexColor = 0xCD853F
	ColorPink                 HexColor = 0xFFC0CB
	ColorPlum                 HexColor = 0xDDA0DD
	ColorPowderBlue           HexColor = 0xB0E0E6
	ColorPurple               HexColor = 0x800080
	ColorRebeccaPurple        HexColor = 0x663399
	ColorRed                  HexColor = 0xFF0000
	ColorRosyBrown            HexColor = 0xBC8F8F
	ColorRoyalBlue            HexColor = 0x4169E1
	ColorSaddleBrown          HexColor = 0x8B4513
	ColorSalmon               HexColor = 0xFA8072
	ColorSandyBrown           HexColor = 0xF4A460
	ColorSeaGreen             HexColor = 0x2E8B57
	ColorSeaShell             HexColor = 0xFFF5EE
	ColorSienna               HexColor = 0xA0522D
	ColorSilver               HexColor = 0xC0C0C0
	ColorSkyBlue              HexColor = 0x87CEEB
	ColorSlateBlue            HexColor = 0x6A5ACD
	ColorSlateGray            HexColor = 0x708090
	ColorSnow                 HexColor = 0xFFFAFA
	ColorSpringGreen          HexColor = 0x00FF7F
	ColorSteelBlue            HexColor = 0x4682B4
	ColorTan                  HexColor = 0xD2B48C
	ColorTeal                 HexColor = 0x008080
	ColorThistle              HexColor = 0xD8BFD8
	ColorTomato               HexColor = 0xFF6347
	ColorTurquoise            HexColor = 0x40E0D0
	ColorViolet               HexColor = 0xEE82EE
	ColorWheat                HexColor = 0xF5DEB3
	ColorWhite                HexColor = 0xFFFFFF
	ColorWhiteSmoke           HexColor = 0xF5F5F5
	ColorYellow               HexColor = 0xFFFF00
	ColorYellowGreen          HexColor = 0x9ACD32
	ColorBox2DRed             HexColor = 0xDC3132
	ColorBox2DBlue            HexColor = 0x30AEBF
	ColorBox2DGreen           HexColor = 0x8CC924
	ColorBox2DYellow          HexColor = 0xFFEE8C
)

// Drawer receives debug drawing commands from World.Draw and Replay.DrawFrameQueries.
type Drawer interface {
	DrawPolygon(transform WorldTransform, vertices []Vec2, color HexColor)
	DrawSolidPolygon(transform WorldTransform, vertices []Vec2, radius float32, color HexColor)
	DrawCircle(center Pos, radius float32, color HexColor)
	DrawSolidCircle(transform WorldTransform, center Vec2, radius float32, color HexColor)
	DrawSolidCapsule(p1, p2 Pos, radius float32, color HexColor)
	DrawLine(p1, p2 Pos, color HexColor)
	DrawTransform(transform WorldTransform)
	DrawPoint(p Pos, size float32, color HexColor)
	DrawString(p Pos, s string, color HexColor)
	DrawAABB(aabb AABB, color HexColor)
}

// DebugDraw selects what World.Draw draws and sends the commands to its Drawer.
type DebugDraw struct {
	Drawer

	DrawingBounds       AABB
	ForceScale          float32
	JointScale          float32
	DrawContacts        bool
	DrawAnchorA         bool
	DrawShapes          bool
	DrawChainNormals    bool
	DrawJoints          bool
	DrawJointExtras     bool
	DrawBounds          bool
	DrawMass            bool
	DrawBodyNames       bool
	DrawGraphColors     bool
	DrawContactFeatures bool
	DrawContactNormals  bool
	DrawContactForces   bool
	DrawFrictionForces  bool
	DrawIslands         bool
}

type drawContext struct {
	world *World
	draw  *DebugDraw
}

type worldQueryContext struct {
	world  *World
	fcn    OverlapResultFunc
	filter QueryFilter
}

type worldOverlapContext struct {
	world  *World
	fcn    OverlapResultFunc
	filter QueryFilter
	proxy  *ShapeProxy
	origin Pos
}

type worldRayCastContext struct {
	world    *World
	fcn      CastResultFunc
	filter   QueryFilter
	fraction float32
	origin   Pos
}

type worldShapeCastContext struct {
	world    *World
	fcn      CastResultFunc
	filter   QueryFilter
	fraction float32
	origin   Pos
	input    ShapeCastInput
}

type worldMoverCastContext struct {
	world    *World
	filter   QueryFilter
	fraction float32
	origin   Pos
	input    ShapeCastInput
}

type worldMoverContext struct {
	world  *World
	fcn    PlaneResultFunc
	filter QueryFilter
	mover  Capsule
	origin Pos
}

type explosionContext struct {
	world            *World
	position         Pos
	radius           float32
	falloff          float32
	impulsePerLength float32
}

func drawShape(draw *DebugDraw, s *shape, transform WorldTransform, color HexColor, drawChainNormals bool) {
	switch s.shapeType {
	case CapsuleShape:
		capsule := &s.capsule
		p1 := transform.Apply(capsule.Center1)
		p2 := transform.Apply(capsule.Center2)
		draw.DrawSolidCapsule(p1, p2, capsule.Radius, color)
	case CircleShape:
		circle := &s.circle
		draw.DrawSolidCircle(transform, circle.Center, circle.Radius, color)
	case PolygonShape:
		poly := &s.polygon
		draw.DrawSolidPolygon(transform, poly.Vertices[:poly.Count], poly.Radius, color)
	case SegmentShape:
		segment := &s.segment
		p1 := transform.Apply(segment.Point1)
		p2 := transform.Apply(segment.Point2)
		draw.DrawLine(p1, p2, color)
	case ChainSegmentShape:
		segment := &s.chainSegment.Segment
		p1 := transform.Apply(segment.Point1)
		p2 := transform.Apply(segment.Point2)
		draw.DrawLine(p1, p2, color)
		draw.DrawPoint(p2, 4, color)

		if drawChainNormals {
			c := p1.Lerp(p2, 0.5)
			e := p2.Sub(p1).Normalize()
			n := e.RightPerp()
			L := float32(0.2) * LengthUnitsPerMeter()
			draw.DrawLine(c, c.Add(n.Scale(L)), ColorPaleGreen)
		}
	}
}

func drawQueryCallback(proxyID int, userData uint64, context *drawContext) bool {
	shapeID := int(userData)
	world := context.world
	draw := context.draw
	shape := &world.shapes[shapeID]
	assert(shape.id == shapeID, "shape->id == shapeId")
	body := &world.bodies[shape.bodyID]
	bodySim := getBodySim(world, body)
	setBit(&world.debugBodySet, uint32(shape.bodyID))

	if draw.DrawShapes {
		var color HexColor
		if shape.material.CustomColor != 0 {
			color = HexColor(shape.material.CustomColor)
		} else if body.bodyType == DynamicBody && body.mass == 0 {
			// Bad body
			color = ColorRed
		} else if body.setIndex == disabledSet {
			color = ColorSlateGray
		} else if shape.sensorIndex != nullIndex {
			color = ColorWheat
		} else if bodySim.flags&hadTimeOfImpact != 0 {
			color = ColorLime
		} else if bodySim.flags&isBullet != 0 && body.setIndex == awakeSet {
			color = ColorTurquoise
		} else if body.flags&isSpeedCapped != 0 {
			color = ColorYellow
		} else if bodySim.flags&isFast != 0 {
			color = ColorSalmon
		} else if body.bodyType == StaticBody {
			color = ColorPaleGreen
		} else if body.bodyType == KinematicBody {
			color = ColorRoyalBlue
		} else if body.setIndex == awakeSet {
			color = ColorPink
		} else {
			color = ColorGray
		}

		drawShape(draw, shape, bodySim.transform, color, draw.DrawChainNormals)
	}

	if draw.DrawBounds {
		draw.DrawAABB(world.fatAABBs[shapeID], ColorGold)
	}

	return true
}

// todo this has varying order for moving shapes, causing flicker when overlapping shapes are moving
// solution: display order by shape id modulus 3, keep 3 buckets in GLSolid* and flush in 3 passes.
func (w *World) Draw(draw *DebugDraw) {
	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return
	}

	checkInput(draw.Drawer != nil, "draw.Drawer != nil")
	assert(draw.DrawingBounds.IsValid(), "b2IsValidAABB( draw->drawingBounds )")

	const axisScale float32 = 0.3

	speculativeColor := ColorGainsboro
	addColor := ColorGreen
	persistColor := ColorBlue
	normalColor := ColorDimGray
	impulseColor := ColorMagenta
	frictionColor := ColorYellow
	bodyCapacity := getIDCapacity(&world.bodyIDPool)
	setBitCountAndClear(&world.debugBodySet, uint32(bodyCapacity))
	jointCapacity := getIDCapacity(&world.jointIDPool)
	setBitCountAndClear(&world.debugJointSet, uint32(jointCapacity))
	contactCapacity := getIDCapacity(&world.contactIDPool)
	setBitCountAndClear(&world.debugContactSet, uint32(contactCapacity))
	islandCapacity := getIDCapacity(&world.islandIDPool)
	setBitCountAndClear(&world.debugIslandSet, uint32(islandCapacity))

	drawContext := drawContext{world: world, draw: draw}
	for i := range int(BodyTypeCount) {
		world.broadPhase.trees[i].QueryAll(draw.DrawingBounds, func(proxyID int, userData uint64) bool { return drawQueryCallback(proxyID, userData, &drawContext) })
	}

	wordCount := world.debugBodySet.blockCount

	bodyBits := world.debugBodySet.bits
	for k := range wordCount {
		word := bodyBits[k]
		for word != 0 {
			ctz := uint32(bits.TrailingZeros64(word))
			bodyID := 64*k + ctz

			body := &world.bodies[int(bodyID)]
			if draw.DrawBodyNames && body.name[0] != 0 {
				offset := Vec2{X: 0.1, Y: 0.1}
				bodySim := getBodySim(world, body)
				transform := WorldTransform{P: bodySim.center, Q: bodySim.transform.Q}
				p := transform.Apply(offset)

				nameLength := 0
				for nameLength < len(body.name) && body.name[nameLength] != 0 {
					nameLength++
				}

				draw.DrawString(p, string(body.name[:nameLength]), ColorWhiteSmoke)
			}

			if draw.DrawMass && body.bodyType == DynamicBody {
				offset := Vec2{X: 0.1, Y: 0.1}
				bodySim := getBodySim(world, body)
				transform := WorldTransform{P: bodySim.center, Q: bodySim.transform.Q}
				draw.DrawLine(bodySim.center0, bodySim.center, ColorWhiteSmoke)
				draw.DrawTransform(transform)
				p := transform.Apply(offset)

				buffer := fmt.Sprintf("  %.2f", body.mass)
				if len(buffer) > 31 {
					buffer = buffer[:31]
				}

				draw.DrawString(p, buffer, ColorWhite)
			}

			if draw.DrawJoints {
				jointKey := body.headJointKey
				for jointKey != nullIndex {
					jointID := jointKey >> 1
					edgeIndex := jointKey & 1
					joint := &world.joints[jointID]
					// avoid double draw
					if !getBit(&world.debugJointSet, uint32(jointID)) {
						drawJoint(draw, world, joint)
						setBit(&world.debugJointSet, uint32(jointID))
					}

					jointKey = joint.edges[edgeIndex].nextKey
				}
			}

			linearSlop := LinearSlop()

			if draw.DrawContacts && body.bodyType == DynamicBody {
				contactKey := body.headContactKey
				for contactKey != nullIndex {
					contactID := contactKey >> 1
					edgeIndex := contactKey & 1
					contact := &world.contacts[contactID]
					contactKey = contact.edges[edgeIndex].nextKey
					// avoid double draw
					if !getBit(&world.debugContactSet, uint32(contactID)) {
						contactSim := getContactSim(world, contact)
						bodyA := &world.bodies[contact.edges[0].bodyID]
						bodySimA := getBodySim(world, bodyA)
						bodyB := &world.bodies[contact.edges[1].bodyID]
						bodySimB := getBodySim(world, bodyB)
						pointCount := contactSim.manifold.PointCount

						normal := contactSim.manifold.Normal
						for j := range pointCount {
							mp := &contactSim.manifold.Points[j]

							var p Pos
							if draw.DrawAnchorA {
								p = bodySimA.center.Add(mp.AnchorA)
							} else {
								p = bodySimB.center.Add(mp.AnchorB)
							}

							if draw.DrawGraphColors && contact.colorIndex != nullIndex {
								// graph color
								pointSize := float32(5)
								if contact.colorIndex == overflowIndex {
									pointSize = 7.5
								}

								draw.DrawPoint(p, pointSize, GraphColor(contact.colorIndex))
							} else if mp.Separation > linearSlop {
								// Speculative
								draw.DrawPoint(p, 5, speculativeColor)
							} else if !mp.Persisted {
								// Add
								draw.DrawPoint(p, 10, addColor)
							} else if mp.Persisted {
								// Persist
								draw.DrawPoint(p, 5, persistColor)
							}

							if draw.DrawContactNormals {
								p1 := p
								p2 := p1.Add(normal.Scale(axisScale))
								draw.DrawLine(p1, p2, normalColor)

								buffer := fmt.Sprintf(" %.2f", mp.Separation)
								if len(buffer) > 31 {
									buffer = buffer[:31]
								}

								draw.DrawString(p1, buffer, ColorWhite)
							} else if draw.DrawContactForces {
								// todo validate
								// multiply by one-half due to relax iteration
								force := float32(0.5) * mp.TotalNormalImpulse * world.invDT
								p1 := p
								p2 := p1.Add(normal.Scale(draw.ForceScale * force))
								draw.DrawLine(p1, p2, impulseColor)

								buffer := fmt.Sprintf("%.1f", force)
								if len(buffer) > 31 {
									buffer = buffer[:31]
								}

								draw.DrawString(p1, buffer, ColorWhite)
							}

							if draw.DrawContactFeatures {
								buffer := strconv.FormatUint(uint64(mp.ID), 10)
								draw.DrawString(p, buffer, ColorOrange)
							}

							if draw.DrawFrictionForces {
								force := float32(0.5) * mp.TangentImpulse * world.invH
								tangent := normal.RightPerp()
								p1 := p
								p2 := p1.Add(tangent.Scale(draw.ForceScale * force))
								draw.DrawLine(p1, p2, frictionColor)

								buffer := fmt.Sprintf("%.1f", force)
								if len(buffer) > 31 {
									buffer = buffer[:31]
								}

								draw.DrawString(p1, buffer, ColorWhite)
							}
						}

						setBit(&world.debugContactSet, uint32(contactID))
					}

					contactKey = contact.edges[edgeIndex].nextKey
				}
			}

			if draw.DrawIslands {
				islandID := body.islandID
				if islandID != nullIndex && !getBit(&world.debugIslandSet, uint32(islandID)) {
					island := &world.islands[islandID]
					if island.setIndex == nullIndex {
						continue
					}

					shapeCount := 0
					aabb := AABB{LowerBound: Vec2{X: math.MaxFloat32, Y: math.MaxFloat32}, UpperBound: Vec2{X: -math.MaxFloat32, Y: -math.MaxFloat32}}

					for bodyIndex := range len(island.bodies) {
						islandBodyID := island.bodies[bodyIndex]
						islandBody := &world.bodies[islandBodyID]

						shapeID := islandBody.headShapeID
						for shapeID != nullIndex {
							shape := &world.shapes[shapeID]
							aabb = aabb.Union(world.fatAABBs[shapeID])
							shapeCount++
							shapeID = shape.nextShapeID
						}
					}

					if shapeCount > 0 {
						draw.DrawAABB(aabb, ColorOrangeRed)
					}

					setBit(&world.debugIslandSet, uint32(islandID))
				}
			}
			// Clear the smallest set bit
			word = word & (word - 1)
		}
	}
}

func treeQueryCallback(proxyID int, userData uint64, context *worldQueryContext) bool {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return true
	}

	id := Shape{index1: int32(shapeID + 1), world: world, generation: shape.generation}

	return context.fcn(id)
}

func (w *World) OverlapAABB(origin Pos, aabb AABB, filter QueryFilter, fcn OverlapResultFunc) TreeStats {
	var treeStats TreeStats

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return treeStats
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")
	assert(aabb.IsValid(), "b2IsValidAABB( aabb )")

	var recWriter recQueryWriter
	if world.recording != nil {
		recQueryBegin(&recWriter)
		recWriter.userFcn.overlapFcn = fcn
		recWPOSITION(&recWriter.buf, origin)
		recWAABB(&recWriter.buf, aabb)
		recWQUERYFILTER(&recWriter.buf, filter)
		recWriter.countOffset = recReserveU32(&recWriter.buf)
		fcn = func(id Shape) bool { return recOverlapTrampoline(id, &recWriter) }
	}
	// Lift to a world float box with outward rounding so the conservative tree test never misses
	worldBox := offsetAABB(aabb, origin)

	worldContext := worldQueryContext{world: world, fcn: fcn, filter: filter}
	for i := range int(BodyTypeCount) {
		treeResult := world.broadPhase.trees[i].Query(worldBox, filter.MaskBits, func(proxyID int, userData uint64) bool { return treeQueryCallback(proxyID, userData, &worldContext) })
		treeStats.NodeVisits += treeResult.NodeVisits
		treeStats.LeafVisits += treeResult.LeafVisits
	}

	if world.recording != nil {
		recPatchU32(&recWriter.buf, recWriter.countOffset, recWriter.hitCount)
		recWTREESTATS(&recWriter.buf, treeStats)
		recQueryCommit(world.recording, 0xE0, &recWriter)
	}

	return treeStats
}

func treeOverlapCallback(proxyID int, userData uint64, context *worldOverlapContext) bool {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return true
	}
	// Re-center on the query origin so the distance test stays in float precision far from the world origin
	body := &world.bodies[shape.bodyID]
	transform := getBodyTransformQuick(world, body).Relative(context.origin)
	input := DistanceInput{ProxyA: *context.proxy, ProxyB: makeShapeDistanceProxy(shape), Transform: transform, UseRadii: true}

	var cache SimplexCache

	output := ShapeDistance(&input, &cache, nil)

	tolerance := float32(0.1) * LinearSlop()
	if output.Distance > tolerance {
		return true
	}

	id := Shape{index1: int32(shape.id + 1), world: world, generation: shape.generation}

	return context.fcn(id)
}

func (w *World) OverlapShape(origin Pos, proxy *ShapeProxy, filter QueryFilter, fcn OverlapResultFunc) TreeStats {
	var treeStats TreeStats

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return treeStats
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")

	var recWriter recQueryWriter
	if world.recording != nil {
		recQueryBegin(&recWriter)
		recWriter.userFcn.overlapFcn = fcn
		recWPOSITION(&recWriter.buf, origin)
		recWSHAPEPROXY(&recWriter.buf, *proxy)
		recWQUERYFILTER(&recWriter.buf, filter)
		recWriter.countOffset = recReserveU32(&recWriter.buf)
		fcn = func(id Shape) bool { return recOverlapTrampoline(id, &recWriter) }
	}
	// Relative box lifted to world float with outward rounding, conservative for the tree
	aabb := offsetAABB(MakeAABB(proxy.Points[:proxy.Count], proxy.Radius), origin)

	worldContext := worldOverlapContext{world: world, fcn: fcn, filter: filter, proxy: proxy, origin: origin}
	for i := range int(BodyTypeCount) {
		treeResult := world.broadPhase.trees[i].Query(aabb, filter.MaskBits, func(proxyID int, userData uint64) bool { return treeOverlapCallback(proxyID, userData, &worldContext) })
		treeStats.NodeVisits += treeResult.NodeVisits
		treeStats.LeafVisits += treeResult.LeafVisits
	}

	if world.recording != nil {
		recPatchU32(&recWriter.buf, recWriter.countOffset, recWriter.hitCount)
		recWTREESTATS(&recWriter.buf, treeStats)
		recQueryCommit(world.recording, 0xE1, &recWriter)
	}

	return treeStats
}

func rayCastCallback(input *RayCastInput, proxyID int, userData uint64, context *worldRayCastContext) float32 {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return input.MaxFraction
	}

	body := &world.bodies[shape.bodyID]
	bodyTransform := getBodyTransformQuick(world, body)
	transform := bodyTransform.Relative(context.origin)
	localInput := *input
	localInput.Origin = Vec2Zero

	output := rayCastShape(&localInput, shape, transform)
	if output.Hit {
		id := Shape{index1: int32(shapeID + 1), world: world, generation: shape.generation}
		point := context.origin.Add(output.Point)
		fraction := context.fcn(id, point, output.Normal, output.Fraction)
		// The user may return -1 to skip this shape
		if 0 <= fraction && fraction <= 1 {
			context.fraction = fraction
		}

		return fraction
	}

	return input.MaxFraction
}

func (w *World) CastRay(origin Pos, translation Vec2, filter QueryFilter, fcn CastResultFunc) TreeStats {
	var treeStats TreeStats

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return treeStats
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")
	assert(translation.IsValid(), "b2IsValidVec2( translation )")

	var recWriter recQueryWriter
	if world.recording != nil {
		recQueryBegin(&recWriter)
		recWriter.userFcn.castFcn = fcn
		recWPOSITION(&recWriter.buf, origin)
		recWVEC2(&recWriter.buf, translation)
		recWQUERYFILTER(&recWriter.buf, filter)
		recWriter.countOffset = recReserveU32(&recWriter.buf)
		fcn = func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
			return recCastTrampoline(id, point, normal, fraction, &recWriter)
		}
	}
	// Tree traversal sees the origin truncated to float, displacing the ray by up to one
	// coordinate ULP, a graze sized miss tolerance at extreme range. Per-shape casts
	// re-difference against the full precision origin carried on the context.
	input := RayCastInput{Origin: origin.Vec2(), Translation: translation, MaxFraction: 1}

	worldContext := worldRayCastContext{world: world, fcn: fcn, filter: filter, fraction: 1, origin: origin}
	for i := range int(BodyTypeCount) {
		treeResult := world.broadPhase.trees[i].CastRay(&input, filter.MaskBits, func(input *RayCastInput, proxyID int, userData uint64) float32 {
			return rayCastCallback(input, proxyID, userData, &worldContext)
		})
		treeStats.NodeVisits += treeResult.NodeVisits
		treeStats.LeafVisits += treeResult.LeafVisits

		if worldContext.fraction == 0 {
			break
		}

		input.MaxFraction = worldContext.fraction
	}

	if world.recording != nil {
		recPatchU32(&recWriter.buf, recWriter.countOffset, recWriter.hitCount)
		recWTREESTATS(&recWriter.buf, treeStats)
		recQueryCommit(world.recording, 0xE2, &recWriter)
	}

	return treeStats
}

// This callback finds the closest hit. This is the most common callback used in games.
func rayCastClosestFcn(shapeID Shape, point Pos, normal Vec2, fraction float32, result *RayResult) float32 {
	// Ignore initial overlap
	if fraction == 0 {
		return -1
	}

	result.Shape = shapeID
	result.Point = point
	result.Normal = normal
	result.Fraction = fraction
	result.Hit = true

	return fraction
}

func (w *World) CastRayClosest(origin Pos, translation Vec2, filter QueryFilter) RayResult {
	var result RayResult

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return result
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")
	assert(translation.IsValid(), "b2IsValidVec2( translation )")
	input := RayCastInput{Origin: origin.Vec2(), Translation: translation, MaxFraction: 1}

	worldContext := worldRayCastContext{world: world, fcn: func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
		return rayCastClosestFcn(id, point, normal, fraction, &result)
	}, filter: filter, fraction: 1, origin: origin}
	for i := range int(BodyTypeCount) {
		treeResult := world.broadPhase.trees[i].CastRay(&input, filter.MaskBits, func(input *RayCastInput, proxyID int, userData uint64) float32 {
			return rayCastCallback(input, proxyID, userData, &worldContext)
		})
		result.NodeVisits += treeResult.NodeVisits
		result.LeafVisits += treeResult.LeafVisits

		if worldContext.fraction == 0 {
			break
		}

		input.MaxFraction = worldContext.fraction
	}

	if world.recording != nil {
		var recBuf recBuffer
		recWPOSITION(&recBuf, origin)
		recWVEC2(&recBuf, translation)
		recWQUERYFILTER(&recBuf, filter)
		recWRAYRESULT(&recBuf, result)
		recCommitRecord(world.recording, 0xE5, recBuf.data, recBuf.size)
		recBufFree(&recBuf)
	}

	return result
}

func shapeCastCallback(input *BoxCastInput, proxyID int, userData uint64, context *worldShapeCastContext) float32 {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return input.MaxFraction
	}
	// Rebuild from the origin relative input, taking only the advancing fraction from the tree.
	// The tree input is world float and would lose the cast far from the origin.
	localInput := context.input
	localInput.MaxFraction = input.MaxFraction
	body := &world.bodies[shape.bodyID]
	transform := getBodyTransformQuick(world, body)
	localTransform := transform.Relative(context.origin)

	output := shapeCastShape(&localInput, shape, localTransform)
	if output.Hit {
		id := Shape{index1: int32(shapeID + 1), world: world, generation: shape.generation}
		point := context.origin.Add(output.Point)
		fraction := context.fcn(id, point, output.Normal, output.Fraction)
		// The user may return -1 to skip this shape
		if 0 <= fraction && fraction <= 1 {
			context.fraction = fraction
		}

		return fraction
	}

	return input.MaxFraction
}

func (w *World) CastShape(origin Pos, proxy *ShapeProxy, translation Vec2, filter QueryFilter, fcn CastResultFunc) TreeStats {
	var treeStats TreeStats

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return treeStats
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")
	assert(translation.IsValid(), "b2IsValidVec2( translation )")

	var recWriter recQueryWriter
	if world.recording != nil {
		recQueryBegin(&recWriter)
		recWriter.userFcn.castFcn = fcn
		recWPOSITION(&recWriter.buf, origin)
		recWSHAPEPROXY(&recWriter.buf, *proxy)
		recWVEC2(&recWriter.buf, translation)
		recWQUERYFILTER(&recWriter.buf, filter)
		recWriter.countOffset = recReserveU32(&recWriter.buf)
		fcn = func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
			return recCastTrampoline(id, point, normal, fraction, &recWriter)
		}
	}

	worldContext := worldShapeCastContext{world: world, fcn: fcn, filter: filter, fraction: 1, origin: origin, input: ShapeCastInput{Proxy: *proxy, Translation: translation, MaxFraction: 1}}
	// Create a conservative world space box from the proxy and origin.
	localBox := MakeAABB(proxy.Points[:proxy.Count], proxy.Radius)
	box := offsetAABB(localBox, origin)

	treeInput := BoxCastInput{Box: box, Translation: translation, MaxFraction: 1}
	for i := range int(BodyTypeCount) {
		treeResult := world.broadPhase.trees[i].CastBox(&treeInput, filter.MaskBits, func(input *BoxCastInput, proxyID int, userData uint64) float32 {
			return shapeCastCallback(input, proxyID, userData, &worldContext)
		})
		treeStats.NodeVisits += treeResult.NodeVisits
		treeStats.LeafVisits += treeResult.LeafVisits

		if worldContext.fraction == 0 {
			break
		}

		treeInput.MaxFraction = worldContext.fraction
	}

	if world.recording != nil {
		recPatchU32(&recWriter.buf, recWriter.countOffset, recWriter.hitCount)
		recWTREESTATS(&recWriter.buf, treeStats)
		recQueryCommit(world.recording, 0xE3, &recWriter)
	}

	return treeStats
}

func moverCastCallback(input *BoxCastInput, proxyID int, userData uint64, context *worldMoverCastContext) float32 {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return context.fraction
	}
	// Rebuild from the origin relative input, taking only the advancing fraction from the tree
	localInput := context.input
	localInput.MaxFraction = input.MaxFraction
	body := &world.bodies[shape.bodyID]
	bodyTransform := getBodyTransformQuick(world, body)
	transform := bodyTransform.Relative(context.origin)

	output := shapeCastShape(&localInput, shape, transform)
	if output.Fraction == 0 {
		// Ignore overlapping shapes
		return context.fraction
	}

	context.fraction = output.Fraction

	return output.Fraction
}

func (w *World) CastMover(origin Pos, mover *Capsule, translation Vec2, filter QueryFilter) float32 {
	assert(origin.IsValid(), "b2IsValidPosition( origin )")
	assert(translation.IsValid(), "b2IsValidVec2( translation )")
	assert(mover.Radius > float32(2)*LinearSlop(), "mover->radius > 2.0f * B2_LINEAR_SLOP")

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return 1
	}

	worldContext := worldMoverCastContext{world: world, filter: filter, fraction: 1, origin: origin}
	worldContext.input.Proxy.Points[0] = mover.Center1
	worldContext.input.Proxy.Points[1] = mover.Center2
	worldContext.input.Proxy.Count = 2
	worldContext.input.Proxy.Radius = mover.Radius
	worldContext.input.Translation = translation
	worldContext.input.MaxFraction = 1
	worldContext.input.CanEncroach = true
	// Bound the capsule in origin relative space then lift to a conservative world float box
	centers := [2]Vec2{mover.Center1, mover.Center2}
	box := offsetAABB(MakeAABB(centers[:], mover.Radius), origin)

	treeInput := BoxCastInput{Box: box, Translation: translation, MaxFraction: 1}
	for i := range int(BodyTypeCount) {
		world.broadPhase.trees[i].CastBox(&treeInput, filter.MaskBits, func(input *BoxCastInput, proxyID int, userData uint64) float32 {
			return moverCastCallback(input, proxyID, userData, &worldContext)
		})

		if worldContext.fraction == 0 {
			break
		}

		treeInput.MaxFraction = worldContext.fraction
	}

	if world.recording != nil {
		var recBuf recBuffer
		recWPOSITION(&recBuf, origin)
		recWCAPSULE(&recBuf, *mover)
		recWVEC2(&recBuf, translation)
		recWQUERYFILTER(&recBuf, filter)
		recWF32(&recBuf, worldContext.fraction)
		recCommitRecord(world.recording, 0xE6, recBuf.data, recBuf.size)
		recBufFree(&recBuf)
	}

	return worldContext.fraction
}

func treeCollideCallback(proxyID int, userData uint64, context *worldMoverContext) bool {
	shapeID := int(userData)
	world := context.world

	shape := &world.shapes[shapeID]
	if !shouldQueryCollide(shape.filter, context.filter) {
		return true
	}
	// Re-center on the query origin, the mover and the resulting planes are origin relative
	body := &world.bodies[shape.bodyID]
	transform := getBodyTransformQuick(world, body).Relative(context.origin)
	result := collideMover(&context.mover, shape, transform)
	// todo handle deep overlap
	if result.Hit && result.Plane.Normal.IsNormalized() {
		id := Shape{index1: int32(shape.id + 1), world: world, generation: shape.generation}
		return context.fcn(id, &result)
	}

	return true
}

// It is tempting to use a shape proxy for the mover, but this makes handling deep overlap difficult and the generality may
// not be worth it.
func (w *World) CollideMover(origin Pos, mover *Capsule, filter QueryFilter, fcn PlaneResultFunc) {
	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return
	}

	assert(origin.IsValid(), "b2IsValidPosition( origin )")

	var recWriter recQueryWriter
	if world.recording != nil {
		recQueryBegin(&recWriter)
		recWriter.userFcn.planeFcn = fcn
		recWPOSITION(&recWriter.buf, origin)
		recWCAPSULE(&recWriter.buf, *mover)
		recWQUERYFILTER(&recWriter.buf, filter)
		recWriter.countOffset = recReserveU32(&recWriter.buf)
		fcn = func(id Shape, plane *PlaneResult) bool { return recPlaneTrampoline(id, plane, &recWriter) }
	}

	r := Vec2{X: mover.Radius, Y: mover.Radius}
	// Relative box lifted to world float with outward rounding, conservative for the tree
	relBox := AABB{LowerBound: mover.Center1.Min(mover.Center2).Sub(r), UpperBound: mover.Center1.Max(mover.Center2).Add(r)}
	aabb := offsetAABB(relBox, origin)

	worldContext := worldMoverContext{world: world, fcn: fcn, filter: filter, mover: *mover, origin: origin}
	for i := range int(BodyTypeCount) {
		world.broadPhase.trees[i].Query(aabb, filter.MaskBits, func(proxyID int, userData uint64) bool { return treeCollideCallback(proxyID, userData, &worldContext) })
	}

	if world.recording != nil {
		recPatchU32(&recWriter.buf, recWriter.countOffset, recWriter.hitCount)
		// CollideMover returns void; no TREESTATS tail
		recQueryCommit(world.recording, 0xE4, &recWriter)
	}
}

func explosionCallback(proxyID int, userData uint64, context *explosionContext) bool {
	shapeID := int(userData)
	world := context.world
	shape := &world.shapes[shapeID]
	body := &world.bodies[shape.bodyID]
	assert(body.bodyType == DynamicBody, "body->type == b2_dynamicBody")
	xf := getBodyTransformQuick(world, body)
	// Re-center the explosion into the shape local frame so distance and direction stay precise
	// far from the origin. Everything below runs in that near-origin frame.
	localPosition := xf.InvApply(context.position)
	input := DistanceInput{ProxyA: makeShapeDistanceProxy(shape), ProxyB: MakeProxy([]Vec2{localPosition}, 0), Transform: TransformIdentity, UseRadii: true}

	var cache SimplexCache

	output := ShapeDistance(&input, &cache, nil)
	radius := context.radius

	falloff := context.falloff
	if output.Distance > radius+falloff {
		return true
	}

	wakeBody(world, body)

	if body.setIndex != awakeSet {
		return true
	}

	closestPoint := output.PointA
	if output.Distance == 0 {
		closestPoint = getShapeCentroid(shape)
	}

	direction := closestPoint.Sub(localPosition)
	if direction.LengthSquared() > float32(100)*float32(0x1p-23)*float32(0x1p-23) {
		direction = direction.Normalize()
	} else {
		direction = Vec2{X: 1}
	}

	localLine := direction.LeftPerp()
	perimeter := getShapeProjectedPerimeter(shape, localLine)

	scale := float32(1)
	if output.Distance > radius && falloff > 0 {
		scale = clampFloat((radius+falloff-output.Distance)/falloff, 0, 1)
	}

	magnitude := context.impulsePerLength * perimeter * scale
	impulse := xf.Q.Rotate(direction).Scale(magnitude)
	localIndex := body.localIndex
	set := &world.solverSets[awakeSet]
	state := &set.bodyStates[localIndex]
	bodySim := &set.bodySims[localIndex]
	state.linearVelocity = state.linearVelocity.MulAdd(bodySim.invMass, impulse)
	// Lever arm from the center of mass to the closest point, rotated to world
	r := xf.Q.Rotate(closestPoint.Sub(bodySim.localCenter))
	state.angularVelocity += float32(bodySim.invInertia * r.Cross(impulse))

	return true
}

func (w *World) Explode(def *ExplosionDef) {
	maskBits := def.MaskBits
	position := def.Position
	radius := def.Radius
	falloff := def.Falloff
	impulsePerLength := def.ImpulsePerLength

	checkInput(position.IsValid(), "b2IsValidPosition( position )")
	checkInput(IsValidFloat(radius) && radius >= 0, "b2IsValidFloat( radius ) && radius >= 0.0f")
	checkInput(IsValidFloat(falloff) && falloff >= 0, "b2IsValidFloat( falloff ) && falloff >= 0.0f")
	checkInput(IsValidFloat(impulsePerLength), "b2IsValidFloat( impulsePerLength )")

	world := w
	assert(!world.locked, "world->locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteWorldExplode(world.recording, &recArgsWorldExplode{def: *def})
	}

	explosionContext := explosionContext{world: world, position: position, radius: radius, falloff: falloff, impulsePerLength: impulsePerLength}
	// The broad-phase tree is float, so translate a local query box out to world with outward rounding
	extent := radius + falloff
	localBox := AABB{LowerBound: Vec2{X: -extent, Y: -extent}, UpperBound: Vec2{X: extent, Y: extent}}
	aabb := offsetAABB(localBox, position)
	world.broadPhase.trees[DynamicBody].Query(aabb, maskBits, func(proxyID int, userData uint64) bool {
		return explosionCallback(proxyID, userData, &explosionContext)
	})
}

func validateConnectivity(world *World) {
	if !validationEnabled {
		return
	}

	bodies := world.bodies
	bodyCapacity := len(world.bodies)

	for bodyIndex := range bodyCapacity {
		body := &bodies[bodyIndex]
		if body.id == nullIndex {
			validateFreeID(&world.bodyIDPool, bodyIndex)
			continue
		}

		validateUsedID(&world.bodyIDPool, bodyIndex)

		assert(bodyIndex == body.id, "bodyIndex == body->id")

		// Need to get the root island because islands are not merged until the next time step
		bodyIslandID := body.islandID
		bodySetIndex := body.setIndex

		contactKey := body.headContactKey
		for contactKey != nullIndex {
			contactID := contactKey >> 1
			edgeIndex := contactKey & 1

			contact := &world.contacts[contactID]

			touching := (contact.flags & contactTouchingFlag) != 0
			if touching {
				if bodySetIndex != staticSet {
					contactIslandID := contact.islandID
					assert(contactIslandID == bodyIslandID, "contactIslandId == bodyIslandId")
				}
			} else {
				assert(contact.islandID == nullIndex, "contact->islandId == B2_NULL_INDEX")
			}

			contactKey = contact.edges[edgeIndex].nextKey
		}

		jointKey := body.headJointKey
		for jointKey != nullIndex {
			jointID := jointKey >> 1
			edgeIndex := jointKey & 1

			joint := &world.joints[jointID]

			otherEdgeIndex := edgeIndex ^ 1

			otherBody := &world.bodies[joint.edges[otherEdgeIndex].bodyId]

			if bodySetIndex == disabledSet || otherBody.setIndex == disabledSet {
				assert(joint.islandId == nullIndex, "joint->islandId == B2_NULL_INDEX")
			} else if bodySetIndex == staticSet {
				// Intentional nesting
				if otherBody.setIndex == staticSet {
					assert(joint.islandId == nullIndex, "joint->islandId == B2_NULL_INDEX")
				}
			} else if body.bodyType != DynamicBody && otherBody.bodyType != DynamicBody {
				assert(joint.islandId == nullIndex, "joint->islandId == B2_NULL_INDEX")
			} else {
				jointIslandID := joint.islandId
				assert(jointIslandID == bodyIslandID, "jointIslandId == bodyIslandId")
			}

			jointKey = joint.edges[edgeIndex].nextKey
		}
	}
}

func validateSolverSets(world *World) {
	if !validationEnabled {
		return
	}

	assert(getIDCapacity(&world.bodyIDPool) == len(world.bodies), "b2GetIdCapacity( &world->bodyIdPool ) == world->bodies.count")
	assert(getIDCapacity(&world.contactIDPool) == len(world.contacts), "b2GetIdCapacity( &world->contactIdPool ) == world->contacts.count")
	assert(getIDCapacity(&world.jointIDPool) == len(world.joints), "b2GetIdCapacity( &world->jointIdPool ) == world->joints.count")
	assert(getIDCapacity(&world.islandIDPool) == len(world.islands), "b2GetIdCapacity( &world->islandIdPool ) == world->islands.count")
	assert(getIDCapacity(&world.solverSetIDPool) == len(world.solverSets), "b2GetIdCapacity( &world->solverSetIdPool ) == world->solverSets.count")

	activeSetCount := 0
	totalBodyCount := 0
	totalJointCount := 0
	totalContactCount := 0
	totalIslandCount := 0

	// Validate all solver sets
	setCount := len(world.solverSets)
	for setIndex := range setCount {
		set := &world.solverSets[setIndex]
		if set.setIndex != nullIndex {
			activeSetCount += 1

			switch setIndex {
			case staticSet:
				assert(len(set.contactSims) == 0, "set->contactSims.count == 0")
				assert(len(set.islandSims) == 0, "set->islandSims.count == 0")
				assert(len(set.bodyStates) == 0, "set->bodyStates.count == 0")
			case disabledSet:
				assert(len(set.islandSims) == 0, "set->islandSims.count == 0")
				assert(len(set.bodyStates) == 0, "set->bodyStates.count == 0")
			case awakeSet:
				assert(len(set.bodySims) == len(set.bodyStates), "set->bodySims.count == set->bodyStates.count")
				assert(len(set.jointSims) == 0, "set->jointSims.count == 0")
			default:
				assert(len(set.bodyStates) == 0, "set->bodyStates.count == 0")
			}

			// Validate bodies
			{
				bodies := world.bodies

				assert(len(set.bodySims) >= 0, "set->bodySims.count >= 0")

				totalBodyCount += len(set.bodySims)
				for i := range len(set.bodySims) {
					bodySim := &set.bodySims[i]

					bodyID := bodySim.bodyID
					assert(0 <= bodyID && bodyID < len(world.bodies), "0 <= bodyId && bodyId < world->bodies.count")
					body := &bodies[bodyID]
					assert(body.setIndex == setIndex, "body->setIndex == setIndex")
					assert(body.localIndex == i, "body->localIndex == i")

					syncedFlags := body.flags &^ (isFast | bodyTransientFlags)
					assert((bodySim.flags&syncedFlags) == syncedFlags, "( bodySim->flags & syncedFlags ) == syncedFlags")

					bodyState := getBodyState(world, body)
					if bodyState != nil {
						assert((bodyState.flags&syncedFlags) == syncedFlags, "( bodyState->flags & syncedFlags ) == syncedFlags")
					}

					if body.bodyType == DynamicBody {
						assert(body.flags&dynamicFlag != 0, "body->flags & b2_dynamicFlag")
					}

					if setIndex == disabledSet {
						assert(body.headContactKey == nullIndex, "body->headContactKey == B2_NULL_INDEX")
					}

					// Validate body shapes
					prevShapeID := nullIndex

					shapeID := body.headShapeID
					for shapeID != nullIndex {
						shape := &world.shapes[shapeID]
						assert(shape.id == shapeID, "shape->id == shapeId")
						assert(shape.prevShapeID == prevShapeID, "shape->prevShapeId == prevShapeId")

						switch setIndex {
						case disabledSet:
							assert(shape.proxyKey == nullIndex, "shape->proxyKey == B2_NULL_INDEX")
						case staticSet:
							assert(proxyType(shape.proxyKey) == StaticBody, "B2_PROXY_TYPE( shape->proxyKey ) == b2_staticBody")
						default:
							proxyType := proxyType(shape.proxyKey)
							assert(proxyType == KinematicBody || proxyType == DynamicBody, "proxyType == b2_kinematicBody || proxyType == b2_dynamicBody")
						}

						prevShapeID = shapeID
						shapeID = shape.nextShapeID
					}

					// Validate body contacts
					contactKey := body.headContactKey
					for contactKey != nullIndex {
						contactID := contactKey >> 1
						edgeIndex := contactKey & 1

						contact := &world.contacts[contactID]
						assert(contact.setIndex != staticSet, "contact->setIndex != b2_staticSet")
						assert(contact.edges[0].bodyID == bodyID || contact.edges[1].bodyID == bodyID, "contact->edges[0].bodyId == bodyId || contact->edges[1].bodyId == bodyId")
						contactKey = contact.edges[edgeIndex].nextKey
					}

					// Validate body joints
					jointKey := body.headJointKey
					for jointKey != nullIndex {
						jointID := jointKey >> 1
						edgeIndex := jointKey & 1

						joint := &world.joints[jointID]

						otherEdgeIndex := edgeIndex ^ 1

						otherBody := &world.bodies[joint.edges[otherEdgeIndex].bodyId]

						if setIndex == disabledSet || otherBody.setIndex == disabledSet {
							assert(joint.setIndex == disabledSet, "joint->setIndex == b2_disabledSet")
						} else if setIndex == staticSet && otherBody.setIndex == staticSet {
							assert(joint.setIndex == staticSet, "joint->setIndex == b2_staticSet")
						} else if body.bodyType != DynamicBody && otherBody.bodyType != DynamicBody {
							assert(joint.setIndex == staticSet, "joint->setIndex == b2_staticSet")
						} else if setIndex == awakeSet {
							assert(joint.setIndex == awakeSet, "joint->setIndex == b2_awakeSet")
						} else if setIndex >= firstSleepingSet {
							assert(joint.setIndex == setIndex, "joint->setIndex == setIndex")
						}

						jointSim := getJointSim(world, joint)
						assert(jointSim.jointId == jointID, "jointSim->jointId == jointId")
						assert(jointSim.bodyIdA == joint.edges[0].bodyId, "jointSim->bodyIdA == joint->edges[0].bodyId")
						assert(jointSim.bodyIdB == joint.edges[1].bodyId, "jointSim->bodyIdB == joint->edges[1].bodyId")

						jointKey = joint.edges[edgeIndex].nextKey
					}
				}
			}

			// Validate contacts
			{
				assert(len(set.contactSims) >= 0, "set->contactSims.count >= 0")

				totalContactCount += len(set.contactSims)
				for i := range len(set.contactSims) {
					contactSim := &set.contactSims[i]

					contact := &world.contacts[contactSim.contactID]
					if setIndex == awakeSet {
						// contact should be non-touching if awake
						// or it could be this contact hasn't been transferred yet
						assert(contactSim.manifold.PointCount == 0 ||
							(contactSim.simFlags&simStartedTouching) != 0, "contactSim->manifold.pointCount == 0 || ( contactSim->simFlags & b2_simStartedTouching ) != 0")
					}

					assert(contact.setIndex == setIndex, "contact->setIndex == setIndex")
					assert(contact.colorIndex == nullIndex, "contact->colorIndex == B2_NULL_INDEX")
					assert(contact.localIndex == i, "contact->localIndex == i")

					edgeBodyA := &world.bodies[contact.edges[0].bodyID]
					edgeBodyB := &world.bodies[contact.edges[1].bodyID]

					assert(contactSim.encodedBodySimA == encodeBodySimIndex(edgeBodyA), "contactSim->encodedBodySimA == b2EncodeBodySimIndex( edgeBodyA )")
					assert(contactSim.encodedBodySimB == encodeBodySimIndex(edgeBodyB), "contactSim->encodedBodySimB == b2EncodeBodySimIndex( edgeBodyB )")
				}
			}

			// Validate joints
			{
				assert(len(set.jointSims) >= 0, "set->jointSims.count >= 0")

				totalJointCount += len(set.jointSims)
				for i := range len(set.jointSims) {
					jointSim := &set.jointSims[i]
					joint := &world.joints[jointSim.jointId]
					assert(joint.setIndex == setIndex, "joint->setIndex == setIndex")
					assert(joint.colorIndex == nullIndex, "joint->colorIndex == B2_NULL_INDEX")
					assert(joint.localIndex == i, "joint->localIndex == i")
				}
			}

			// Validate islands
			{
				assert(len(set.islandSims) >= 0, "set->islandSims.count >= 0")

				totalIslandCount += len(set.islandSims)
				for i := range len(set.islandSims) {
					islandSim := &set.islandSims[i]
					island := &world.islands[islandSim.islandID]
					assert(island.setIndex == setIndex, "island->setIndex == setIndex")
					assert(island.localIndex == i, "island->localIndex == i")
				}
			}
		} else {
			assert(len(set.bodySims) == 0, "set->bodySims.count == 0")
			assert(len(set.contactSims) == 0, "set->contactSims.count == 0")
			assert(len(set.jointSims) == 0, "set->jointSims.count == 0")
			assert(len(set.islandSims) == 0, "set->islandSims.count == 0")
			assert(len(set.bodyStates) == 0, "set->bodyStates.count == 0")
		}
	}

	setIdCount := getIDCount(&world.solverSetIDPool)
	assert(activeSetCount == setIdCount, "activeSetCount == setIdCount")

	bodyIdCount := getIDCount(&world.bodyIDPool)
	assert(totalBodyCount == bodyIdCount, "totalBodyCount == bodyIdCount")

	islandIdCount := getIDCount(&world.islandIDPool)
	assert(totalIslandCount == islandIdCount, "totalIslandCount == islandIdCount")

	// Validate constraint graph
	for colorIndex := range GraphColorCount {
		color := &world.constraintGraph.colors[colorIndex]
		bitCount := 0

		assert(len(color.contactSims) >= 0, "color->contactSims.count >= 0")

		totalContactCount += len(color.contactSims)
		for i := range len(color.contactSims) {
			contactSim := &color.contactSims[i]
			contact := &world.contacts[contactSim.contactID]
			// contact should be touching in the constraint graph or awaiting transfer to non-touching
			assert(contactSim.manifold.PointCount > 0 ||
				(contactSim.simFlags&(simStoppedTouching|simDisjoint)) != 0, "contactSim->manifold.pointCount > 0 || ( contactSim->simFlags & ( b2_simStoppedTouching | b2_simDisjoint ) ) != 0")
			assert(contact.setIndex == awakeSet, "contact->setIndex == b2_awakeSet")
			assert(contact.colorIndex == colorIndex, "contact->colorIndex == colorIndex")
			assert(contact.localIndex == i, "contact->localIndex == i")

			bodyIDA := contact.edges[0].bodyID
			bodyIDB := contact.edges[1].bodyID

			edgeBodyA := &world.bodies[bodyIDA]
			edgeBodyB := &world.bodies[bodyIDB]

			assert(contactSim.encodedBodySimA == encodeBodySimIndex(edgeBodyA), "contactSim->encodedBodySimA == b2EncodeBodySimIndex( edgeBodyA )")
			assert(contactSim.encodedBodySimB == encodeBodySimIndex(edgeBodyB), "contactSim->encodedBodySimB == b2EncodeBodySimIndex( edgeBodyB )")

			if colorIndex < overflowIndex {
				assert(getBit(&color.bodySet, uint32(bodyIDA)) == (edgeBodyA.bodyType == DynamicBody), "b2GetBit( &color->bodySet, bodyIdA ) == ( edgeBodyA->type == b2_dynamicBody )")
				assert(getBit(&color.bodySet, uint32(bodyIDB)) == (edgeBodyB.bodyType == DynamicBody), "b2GetBit( &color->bodySet, bodyIdB ) == ( edgeBodyB->type == b2_dynamicBody )")

				if edgeBodyA.bodyType == DynamicBody {
					bitCount++
				}

				if edgeBodyB.bodyType == DynamicBody {
					bitCount++
				}
			}
		}

		assert(len(color.jointSims) >= 0, "color->jointSims.count >= 0")

		totalJointCount += len(color.jointSims)
		for i := range len(color.jointSims) {
			jointSim := &color.jointSims[i]
			joint := &world.joints[jointSim.jointId]
			assert(joint.setIndex == awakeSet, "joint->setIndex == b2_awakeSet")
			assert(joint.colorIndex == colorIndex, "joint->colorIndex == colorIndex")
			assert(joint.localIndex == i, "joint->localIndex == i")

			bodyIDA := joint.edges[0].bodyId
			bodyIDB := joint.edges[1].bodyId

			if colorIndex < overflowIndex {
				bodyA := &world.bodies[bodyIDA]
				bodyB := &world.bodies[bodyIDB]

				assert(getBit(&color.bodySet, uint32(bodyIDA)) == (bodyA.bodyType == DynamicBody), "b2GetBit( &color->bodySet, bodyIdA ) == ( bodyA->type == b2_dynamicBody )")
				assert(getBit(&color.bodySet, uint32(bodyIDB)) == (bodyB.bodyType == DynamicBody), "b2GetBit( &color->bodySet, bodyIdB ) == ( bodyB->type == b2_dynamicBody )")

				if bodyA.bodyType == DynamicBody {
					bitCount++
				}

				if bodyB.bodyType == DynamicBody {
					bitCount++
				}
			}
		}

		// Validate the bit population for this graph color
		assert(bitCount == countSetBits(&color.bodySet), "bitCount == b2CountSetBits( &color->bodySet )")
	}

	contactIdCount := getIDCount(&world.contactIDPool)
	assert(totalContactCount == contactIdCount, "totalContactCount == contactIdCount")
	assert(totalContactCount == int(world.broadPhase.pairSet.count), "totalContactCount == (int)world->broadPhase.pairSet.count")

	jointIdCount := getIDCount(&world.jointIDPool)
	assert(totalJointCount == jointIdCount, "totalJointCount == jointIdCount")
}

func validateContacts(world *World) {
	if !validationEnabled {
		return
	}

	contactCount := len(world.contacts)
	assert(contactCount == getIDCapacity(&world.contactIDPool), "contactCount == b2GetIdCapacity( &world->contactIdPool )")

	allocatedContactCount := 0

	for contactIndex := range contactCount {
		contact := &world.contacts[contactIndex]
		if contact.contactID == nullIndex {
			continue
		}

		assert(contact.contactID == contactIndex, "contact->contactId == contactIndex")

		allocatedContactCount += 1

		touching := (contact.flags & contactTouchingFlag) != 0

		setID := contact.setIndex

		if setID == awakeSet {
			if touching {
				assert(0 <= contact.colorIndex && contact.colorIndex < GraphColorCount, "0 <= contact->colorIndex && contact->colorIndex < B2_GRAPH_COLOR_COUNT")
			} else {
				assert(contact.colorIndex == nullIndex, "contact->colorIndex == B2_NULL_INDEX")
			}
		} else if setID >= firstSleepingSet {
			// Only touching contacts allowed in a sleeping set
			assert(touching, "touching == true")
		} else {
			// Sleeping and non-touching contacts belong in the disabled set
			assert(!touching && setID == disabledSet, "touching == false && setId == b2_disabledSet")
		}

		contactSim := getContactSim(world, contact)
		assert(contactSim.contactID == contactIndex, "contactSim->contactId == contactIndex")
		assert(contactSim.bodyIDA == contact.edges[0].bodyID, "contactSim->bodyIdA == contact->edges[0].bodyId")
		assert(contactSim.bodyIDB == contact.edges[1].bodyID, "contactSim->bodyIdB == contact->edges[1].bodyId")

		simTouching := (contactSim.simFlags & simTouchingFlag) != 0
		assert(touching == simTouching, "touching == simTouching")

		assert(0 <= contactSim.manifold.PointCount && contactSim.manifold.PointCount <= 2, "0 <= contactSim->manifold.pointCount && contactSim->manifold.pointCount <= 2")
	}

	contactIdCount := getIDCount(&world.contactIDPool)
	assert(allocatedContactCount == contactIdCount, "allocatedContactCount == contactIdCount")
}
