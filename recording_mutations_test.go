// SPDX-License-Identifier: MIT
package box2d_test

import (
	"fmt"
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"testing"
)

// Each mutation phase has two steps. The first is a dispatch fence: its replay
// observations cannot be cancelled by the next phase's between-frame mutation.
func TestRecordingMutationParity(t *testing.T) {
	r := oracle.Open(t, "recording/MutationParity")
	wd := box2d.DefaultWorldDef()
	wd.Gravity = box2d.Vec2{}
	w := box2d.NewWorld(&wd)
	box2d.BTWorldCleanup(t, w)
	rec := box2d.NewRecording(0)
	ignored := box2d.NewRecording(0)
	w.StartRecording(rec)
	w.StartRecording(ignored)
	bd := box2d.DefaultBodyDef()
	bd.Name = "ground"
	ground := w.CreateBody(&bd)
	bd.Type = box2d.DynamicBody
	bd.Position = box2d.Pos{X: 2, Y: 3}
	bd.Name = "moving"
	body := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	circle := box2d.Circle{Radius: 0.5}
	shape := body.CreateCircleShape(&sd, &circle)
	sensorDef := box2d.DefaultShapeDef()
	sensorDef.IsSensor = true
	sensorCircle := box2d.Circle{Center: box2d.Vec2{X: 2, Y: 3}, Radius: 2}
	ground.CreateCircleShape(&sensorDef, &sensorCircle)
	md := box2d.DefaultMoverJointDef()
	md.Base.BodyA = ground
	md.Base.BodyB = body
	mover := w.CreateMoverJoint(&md)
	pd := box2d.DefaultPogoJointDef()
	pd.Base.BodyA = ground
	pd.Base.BodyB = body
	pd.Normal = box2d.Vec2{Y: 1}
	pd.RestLength = 3
	pd.MaxCompressionForce = 15
	pd.MaxTensionForce = 5
	pogo := w.CreatePogoJoint(&pd)
	wheelDef := box2d.DefaultWheelJointDef()
	wheelDef.Base.BodyA = ground
	wheelDef.Base.BodyB = body
	wheel := w.CreateWheelJoint(&wheelDef)
	wheel.SetSpringEnabled(false)
	points := []box2d.Vec2{{X: -4}, {X: -2}, {}, {X: 2}, {X: 4}}
	cd := box2d.DefaultChainDef()
	cd.Points = points
	cd.Ghost1 = box2d.Vec2{X: -6}
	cd.Ghost2 = box2d.Vec2{X: 6}
	cd.Materials = []box2d.SurfaceMaterial{box2d.DefaultSurfaceMaterial()}
	chain := ground.CreateChain(&cd)
	for phase := range 18 {
		switch phase {
		case 0:
			cap := box2d.Capsule{Center1: box2d.Vec2{X: -0.4}, Center2: box2d.Vec2{X: 0.6}, Radius: 0.3}
			shape.SetCapsule(&cap)
			body.SetMassData(box2d.MassData{Mass: 7, Center: box2d.Vec2{X: 0.2}, RotationalInertia: 2})
			body.SetName("mutated name")
			body.SetContactEventsEnabled(true)
			body.SetHitEventsEnabled(true)
			surface := box2d.DefaultSurfaceMaterial()
			surface.Friction = 0.65
			surface.Restitution = 0.35
			surface.RollingResistance = 0.15
			surface.TangentSpeed = 2
			surface.UserMaterialID = 0x123456789abcdef0
			surface.CustomColor = 0xabcdef
			shape.SetSurfaceMaterial(&surface)
			shape.SetSensorEventsEnabled(true)
			shape.SetPreSolveEventsEnabled(true)
			chain.SetAllSurfaceMaterials(&surface)
			surface.Friction = 0.85
			surface.UserMaterialID = 19
			chain.SetSurfaceMaterial(&surface, 2)
			mover.SetLinearVelocity(box2d.Vec2{X: 0.7, Y: -0.1})
			mover.SetMaxVelocityForce(box2d.Vec2{X: 3, Y: 4})
			pogo.SetRestLength(2.5)
			pogo.SetSpringHertz(2)
			pogo.SetSpringDampingRatio(0.8)
			mover.SetLocalFrameA(box2d.Transform{P: box2d.Vec2{X: 0.1, Y: 0.2}, Q: box2d.MakeRot(0.1)})
			mover.SetLocalFrameB(box2d.Transform{P: box2d.Vec2{X: -0.2}, Q: box2d.MakeRot(-0.2)})
			mover.SetCollideConnected(true)
			mover.SetConstraintTuning(17, 0.3)
			mover.SetForceThreshold(0.01)
			mover.SetTorqueThreshold(0.02)
			w.SetSleepingEnabled(false)
			w.SetContinuousEnabled(false)
			w.SetWarmStartingEnabled(false)
			w.SetRestitutionPropagationEnabled(false)
		case 1:
			body.UpdateMassFromShapes()
			polygon := box2d.MakeBox(0.7, 0.2)
			shape.SetPolygon(&polygon)
			shape.SetUserMaterial(0xfedcba9876543210)
			body.SetContactEventsEnabled(false)
			body.SetHitEventsEnabled(false)
			shape.SetSensorEventsEnabled(false)
			shape.SetPreSolveEventsEnabled(false)
		case 2:
			body.SetType(box2d.KinematicBody)
			body.SetLinearVelocity(box2d.Vec2{X: 0.25})
			segment := box2d.Segment{Point1: box2d.Vec2{X: -0.6}, Point2: box2d.Vec2{X: 0.8, Y: 0.3}}
			shape.SetSegment(&segment)
		case 3:
			body.SetType(box2d.DynamicBody)
			chainSeg := box2d.ChainSegment{Ghost1: box2d.Vec2{X: -2}, Segment: box2d.Segment{Point1: box2d.Vec2{X: -1}, Point2: box2d.Vec2{X: 1}}, Ghost2: box2d.Vec2{X: 2}, ChainID: -1}
			shape.SetChainSegment(&chainSeg)
		case 4:
			shape.SetCircle(&circle)
			body.Disable()
		case 5:
			body.Enable()
			body.SetMotionLocks(box2d.MotionLocks{LinearX: true, AngularZ: true})
			body.SetSleepEnabled(true)
			w.SetSleepingEnabled(true)
		case 6:
			body.SetAwake(false)
		case 7:
			mover.WakeBodies()
			body.SetMotionLocks(box2d.MotionLocks{})
		case 8:
			body.SetSleepEnabled(false)
			body.ApplyForceToCenter(box2d.Vec2{X: 40, Y: 20}, true)
			body.ApplyTorque(3, true)
			body.ClearForces()
		case 9:
			mover.Destroy()
			pogo.Destroy()
			chain.Destroy()
			shape.Destroy(true)
			replacement := body.CreateCircleShape(&sd, &circle)
			replacement.SetDensity(3, false)
			body.UpdateMassFromShapes()
		case 10:
			body.Destroy()
			bd.Name = "reused body"
			bd.Position = box2d.Pos{X: 4, Y: 6}
			body = w.CreateBody(&bd)
			shape = body.CreateCircleShape(&sd, &circle)
		case 11:
			body.SetTransform(box2d.Pos{X: 5, Y: 7}, box2d.MakeRot(0.3))
			body.SetLinearVelocity(box2d.Vec2{Y: -1})
			body.SetAngularVelocity(0.4)
		case 12:
			body.SetSleepEnabled(true)
			body.SetAwake(false)
		case 13:
			body.SetAwake(true)
		case 14:
			support := box2d.Circle{Center: box2d.Vec2{X: 5, Y: 7}, Radius: 1}
			ground.CreateCircleShape(&sd, &support)
			body.SetContactEventsEnabled(true)
			body.SetHitEventsEnabled(true)
			w.SetHitEventThreshold(0.01)
		case 15:
			body.SetAwake(false)
		case 16:
			ground.WakeTouching()
		case 17:
			point := body.Position().Add(box2d.Vec2{X: 0.3, Y: 0.2})
			body.ApplyForce(box2d.Vec2{X: 4, Y: 6}, point, true)
			body.ApplyForceToCenter(box2d.Vec2{X: 2, Y: 3}, true)
			body.ApplyTorque(1.5, true)
			body.ApplyLinearImpulse(box2d.Vec2{X: 0.2, Y: 0.1}, point, true)
			body.ApplyLinearImpulseToCenter(box2d.Vec2{X: 0.3, Y: 0.2}, true)
			body.ApplyAngularImpulse(0.4, true)
		}
		for sub := range 2 {
			w.Step(1.0/60.0, 4)
			key := fmt.Sprintf("live/%d", phase*2+sub+1)
			r.U64(key+"/state", w.StateHash())
			recordingProperties(r, w, key)
			recordingEvents(r, w, key)
		}
	}
	w.StopRecording()
	w.StopRecording()
	r.Bool("repeatedStart/ignoredEmpty", ignored.Size() == 0)
	body.SetName("after stop")
	w.Close()
	p, destroy := recordingParityPlayer(t, rec.Data(), 0)
	p.SetKeyframePolicy(1<<26, 2)
	for p.StepFrame() {
		key := fmt.Sprintf("replay/%d", p.Frame())
		r.U64(key+"/state", p.World().StateHash())
		recordingProperties(r, p.World(), key)
		recordingEvents(r, p.World(), key)
	}
	r.Bool("replay/diverged", p.HasDiverged())
	for i, target := range []int{36, 1, 14, 10, 0, 18, 35, 99, -5} {
		p.SeekFrame(target)
		key := fmt.Sprintf("seek/%d", i)
		recordingObserve(t, r, p, key)
		recordingProperties(r, p.World(), key)
		recordingEvents(r, p.World(), key)
	}
	destroy()
	// Reusing a stopped buffer replaces the old session rather than appending it.
	wd.Gravity = box2d.Vec2{Y: -2}
	w = box2d.NewWorld(&wd)
	box2d.BTWorldCleanup(t, w)
	w.StartRecording(rec)
	bd.Name = "new session"
	body = w.CreateBody(&bd)
	body.CreateCircleShape(&sd, &circle)
	w.Step(1.0/20.0, 3)
	w.StopRecording()
	w.Close()
	p, destroy = recordingParityPlayer(t, rec.Data(), 0)
	recordingObserve(t, r, p, "reuse/open")
	r.Bool("reuse/step", p.StepFrame())
	recordingObserve(t, r, p, "reuse/after")
	destroy()
}
