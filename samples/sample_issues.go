// SPDX-License-Identifier: MIT

package main

import "github.com/zigai/box2d-go/box2d"

type UnstablePrismaticJoints struct {
	Sample
}

func CreateUnstablePrismaticJoints(context *SampleContext) Sampler {
	s := &UnstablePrismaticJoints{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.75}
		s.context.camera.zoom = 32.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -100.0, Y: 0.0}, Point2: box2d.Vec2{X: 100.0, Y: 0.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	var centerID box2d.Body
	{
		bd := box2d.DefaultBodyDef()
		bd.Type = box2d.DynamicBody
		bd.Position = box2d.Pos{X: 0, Y: 3}
		centerID = s.worldID.CreateBody(&bd)

		sd := box2d.DefaultShapeDef()

		var circle box2d.Circle
		circle.Center = box2d.Vec2{X: 0, Y: 0}

		circle.Radius = 0.5

		centerID.CreateCircleShape(&sd, &circle)
	}

	jd := box2d.DefaultPrismaticJointDef()
	jd.EnableSpring = true
	jd.Hertz = 10.0
	jd.DampingRatio = 2.0

	{
		bd := box2d.DefaultBodyDef()
		bd.Type = box2d.DynamicBody
		bd.Position = box2d.Pos{X: -3.5, Y: 3}

		leftID := s.worldID.CreateBody(&bd)

		sd := box2d.DefaultShapeDef()

		var circle box2d.Circle
		circle.Center = box2d.Vec2{X: 0, Y: 0}
		circle.Radius = 2.0
		leftID.CreateCircleShape(&sd, &circle)

		jd.Base.BodyA = centerID
		jd.Base.BodyB = leftID
		jd.TargetTranslation = -3.0
		s.worldID.CreatePrismaticJoint(&jd)
	}

	{
		bd := box2d.DefaultBodyDef()
		bd.Type = box2d.DynamicBody
		bd.Position = box2d.Pos{X: 3.5, Y: 3}
		rightID := s.worldID.CreateBody(&bd)

		sd := box2d.DefaultShapeDef()

		var circle box2d.Circle
		circle.Center = box2d.Vec2{X: 0, Y: 0}
		circle.Radius = 2.0

		rightID.CreateCircleShape(&sd, &circle)

		jd.Base.BodyA = centerID
		jd.Base.BodyB = rightID
		jd.TargetTranslation = 3.0
		s.worldID.CreatePrismaticJoint(&jd)
	}

	return s
}

type UnstableWindmill struct {
	Sample
}

func CreateUnstableWindmill(context *SampleContext) Sampler {
	s := &UnstableWindmill{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 1.75}
		s.context.camera.zoom = 32.0
	}

	{
		bodyDef := box2d.DefaultBodyDef()
		groundID := s.worldID.CreateBody(&bodyDef)

		shapeDef := box2d.DefaultShapeDef()
		segment := box2d.Segment{Point1: box2d.Vec2{X: -100.0, Y: -10.0}, Point2: box2d.Vec2{X: 100.0, Y: -10.0}}
		groundID.CreateSegmentShape(&shapeDef, &segment)
	}

	bdef := box2d.DefaultBodyDef()
	bdef.GravityScale = 0.0
	bdef.Type = box2d.DynamicBody
	sdef := box2d.DefaultShapeDef()
	sdef.Material = box2d.DefaultSurfaceMaterial()
	sdef.Material.Friction = 0.1

	bdef.Position = box2d.Pos{X: 10, Y: 10}
	center := s.worldID.CreateBody(&bdef)
	circle := box2d.Circle{Center: box2d.Vec2{X: 0, Y: 0}, Radius: 5}
	center.CreateCircleShape(&sdef, &circle)

	wjdef := box2d.DefaultWeldJointDef()

	wjdef.Base.BodyA = center

	var polygon box2d.Polygon

	bdef.Position = box2d.Pos{X: 10, Y: 0}
	body := s.worldID.CreateBody(&bdef)
	polygon = box2d.MakeBox(4, 5)
	body.CreatePolygonShape(&sdef, &polygon)
	wjdef.Base.LocalFrameA = box2d.Transform{P: box2d.Vec2{X: 0, Y: -5}, Q: box2d.RotIdentity}
	wjdef.Base.BodyB = body
	wjdef.Base.LocalFrameB = box2d.Transform{P: box2d.Vec2{X: 0, Y: 5}, Q: box2d.RotIdentity}
	s.worldID.CreateWeldJoint(&wjdef)

	bdef.Position = box2d.Pos{X: 20, Y: 10}
	body = s.worldID.CreateBody(&bdef)
	polygon = box2d.MakeBox(5, 4)
	body.CreatePolygonShape(&sdef, &polygon)
	wjdef.Base.LocalFrameA = box2d.Transform{P: box2d.Vec2{X: 5, Y: 0}, Q: box2d.RotIdentity}
	wjdef.Base.BodyB = body
	wjdef.Base.LocalFrameB = box2d.Transform{P: box2d.Vec2{X: -5, Y: 0}, Q: box2d.RotIdentity}
	s.worldID.CreateWeldJoint(&wjdef)

	bdef.Position = box2d.Pos{X: 10, Y: 20}
	body = s.worldID.CreateBody(&bdef)
	polygon = box2d.MakeBox(4, 5)
	body.CreatePolygonShape(&sdef, &polygon)
	wjdef.Base.LocalFrameA = box2d.Transform{P: box2d.Vec2{X: 0, Y: 5}, Q: box2d.RotIdentity}
	wjdef.Base.BodyB = body
	wjdef.Base.LocalFrameB = box2d.Transform{P: box2d.Vec2{X: 0, Y: -5}, Q: box2d.RotIdentity}
	s.worldID.CreateWeldJoint(&wjdef)

	bdef.Position = box2d.Pos{X: 0, Y: 10}
	body = s.worldID.CreateBody(&bdef)
	polygon = box2d.MakeBox(5, 4)
	body.CreatePolygonShape(&sdef, &polygon)
	wjdef.Base.LocalFrameA = box2d.Transform{P: box2d.Vec2{X: -5, Y: 0}, Q: box2d.RotIdentity}
	wjdef.Base.BodyB = body
	wjdef.Base.LocalFrameB = box2d.Transform{P: box2d.Vec2{X: 5, Y: 0}, Q: box2d.RotIdentity}
	s.worldID.CreateWeldJoint(&wjdef)

	return s
}

func init() {
	RegisterSample("Issues", "Unstable Prismatic Joints", CreateUnstablePrismaticJoints)
	RegisterSample("Issues", "Unstable Windmill", CreateUnstableWindmill)
}
