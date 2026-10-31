// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

func jointBodyState(r *oracle.Reader, phase string, b box2d.Body) {
	x := b.Transform()
	r.Pos(phase+"/position", oracle.Real(x.P.X), oracle.Real(x.P.Y))
	r.Rot(phase+"/rotation", x.Q.C, x.Q.S)

	velocity := b.LinearVelocity()
	r.Vec2(phase+"/velocity", velocity.X, velocity.Y)
	r.F32(phase+"/angularVelocity", b.AngularVelocity())
	r.Bool(phase+"/awake", b.Awake())
	r.Bool(phase+"/enabled", b.Enabled())
	r.I32(phase+"/type", int32(b.Type()))
}

func jointScalarGetter(r *oracle.Reader, key string, get func() float32) {
	assertions := 0

	var value float32

	rejectedInput(&assertions, func() { value = get() })
	r.I32(key+"Assertions", int32(assertions))

	if assertions == 0 {
		r.F32(key, value)
	}
}

func observeJoint(
	r *oracle.Reader,
	phase string,
	w *box2d.World,
	j box2d.Joint,
	a box2d.Body,
	b box2d.Body,
) {
	jointBodyState(r, phase+"/bodyA", a)
	jointBodyState(r, phase+"/bodyB", b)
	r.I32(phase+"/jointCount", int32(w.Counters().JointCount))
	r.I32(phase+"/contactCount", int32(w.Counters().ContactCount))
	r.I32(phase+"/type", int32(j.Type()))
	r.Bool(phase+"/bodyA", j.BodyA() == a)
	r.Bool(phase+"/bodyB", j.BodyB() == b)
	r.Bool(phase+"/world", j.World() == w)
	fa, fb := j.LocalFrameA(), j.LocalFrameB()
	r.Vec2(phase+"/frameA/position", fa.P.X, fa.P.Y)
	r.Rot(phase+"/frameA/rotation", fa.Q.C, fa.Q.S)
	r.Vec2(phase+"/frameB/position", fb.P.X, fb.P.Y)
	r.Rot(phase+"/frameB/rotation", fb.Q.C, fb.Q.S)
	r.Bool(phase+"/collideConnected", j.CollideConnected())
	force := j.ConstraintForce()
	r.Vec2(phase+"/constraintForce", force.X, force.Y)
	jointScalarGetter(r, phase+"/constraintTorque", j.ConstraintTorque)
	jointScalarGetter(r, phase+"/linearSeparation", j.LinearSeparation)
	jointScalarGetter(r, phase+"/angularSeparation", j.AngularSeparation)
	hertz, damping := j.ConstraintTuning()
	r.F32(phase+"/constraintHertz", hertz)
	r.F32(phase+"/constraintDamping", damping)
	r.F32(phase+"/forceThreshold", j.ForceThreshold())
	r.F32(phase+"/torqueThreshold", j.TorqueThreshold())

	switch j.Type() {
	case box2d.DistanceJointType:
		f := box2d.DistanceJoint{Joint: j}
		r.F32(phase+"/Length", f.Length())
		r.Bool(phase+"/IsSpringEnabled", f.SpringEnabled())
		r.F32(phase+"/SpringHertz", f.SpringHertz())
		r.F32(phase+"/SpringDampingRatio", f.SpringDampingRatio())
		r.F32(phase+"/SpringForce", f.SpringForce())
		r.Bool(phase+"/IsLimitEnabled", f.LimitEnabled())
		r.F32(phase+"/MinLength", f.MinLength())
		r.F32(phase+"/MaxLength", f.MaxLength())
		r.F32(phase+"/CurrentLength", f.CurrentLength())
		r.Bool(phase+"/IsMotorEnabled", f.MotorEnabled())
		r.F32(phase+"/MotorSpeed", f.MotorSpeed())
		r.F32(phase+"/MaxMotorForce", f.MaxMotorForce())
		r.F32(phase+"/MotorForce", f.MotorForce())
		lo, hi := f.SpringForceRange()
		r.F32(phase+"/LowerSpringForce", lo)
		r.F32(phase+"/UpperSpringForce", hi)
	case box2d.FilterJointType:
	case box2d.MotorJointType:
		f := box2d.MotorJoint{Joint: j}
		velocity := f.LinearVelocity()
		r.Vec2(phase+"/LinearVelocity", velocity.X, velocity.Y)
		r.F32(phase+"/AngularVelocity", f.AngularVelocity())
		r.F32(phase+"/MaxVelocityForce", f.MaxVelocityForce())
		r.F32(phase+"/MaxVelocityTorque", f.MaxVelocityTorque())
		r.F32(phase+"/LinearHertz", f.LinearHertz())
		r.F32(phase+"/LinearDampingRatio", f.LinearDampingRatio())
		r.F32(phase+"/MaxSpringForce", f.MaxSpringForce())
		r.F32(phase+"/AngularHertz", f.AngularHertz())
		r.F32(phase+"/AngularDampingRatio", f.AngularDampingRatio())
		r.F32(phase+"/MaxSpringTorque", f.MaxSpringTorque())
	case box2d.MoverJointType:
		f := box2d.MoverJoint{Joint: j}
		velocity, force := f.LinearVelocity(), f.MaxVelocityForce()
		r.Vec2(phase+"/LinearVelocity", velocity.X, velocity.Y)
		r.Vec2(phase+"/MaxVelocityForce", force.X, force.Y)
	case box2d.PogoJointType:
		f := box2d.PogoJoint{Joint: j}
		r.F32(phase+"/RestLength", f.RestLength())
		r.F32(phase+"/SpringHertz", f.SpringHertz())
		r.F32(phase+"/SpringDampingRatio", f.SpringDampingRatio())
		r.F32(phase+"/Length", f.Length())
		r.F32(phase+"/Impulse", f.Impulse())
		r.F32(phase+"/Velocity", f.Velocity())
	case box2d.PrismaticJointType:
		f := box2d.PrismaticJoint{Joint: j}
		r.Bool(phase+"/IsSpringEnabled", f.SpringEnabled())
		r.F32(phase+"/SpringHertz", f.SpringHertz())
		r.F32(phase+"/SpringDampingRatio", f.SpringDampingRatio())
		r.F32(phase+"/TargetTranslation", f.TargetTranslation())
		r.Bool(phase+"/IsLimitEnabled", f.LimitEnabled())
		r.F32(phase+"/LowerLimit", f.LowerLimit())
		r.F32(phase+"/UpperLimit", f.UpperLimit())
		r.Bool(phase+"/IsMotorEnabled", f.MotorEnabled())
		r.F32(phase+"/MotorSpeed", f.MotorSpeed())
		r.F32(phase+"/MaxMotorForce", f.MaxMotorForce())
		r.F32(phase+"/MotorForce", f.MotorForce())
		r.F32(phase+"/Translation", f.Translation())
		r.F32(phase+"/Speed", f.Speed())
	case box2d.RevoluteJointType:
		f := box2d.RevoluteJoint{Joint: j}
		r.Bool(phase+"/IsSpringEnabled", f.SpringEnabled())
		r.F32(phase+"/SpringHertz", f.SpringHertz())
		r.F32(phase+"/SpringDampingRatio", f.SpringDampingRatio())
		r.F32(phase+"/TargetAngle", f.TargetAngle())
		r.F32(phase+"/Angle", f.Angle())
		r.Bool(phase+"/IsLimitEnabled", f.LimitEnabled())
		r.F32(phase+"/LowerLimit", f.LowerLimit())
		r.F32(phase+"/UpperLimit", f.UpperLimit())
		r.Bool(phase+"/IsMotorEnabled", f.MotorEnabled())
		r.F32(phase+"/MotorSpeed", f.MotorSpeed())
		r.F32(phase+"/MaxMotorTorque", f.MaxMotorTorque())
		r.F32(phase+"/MotorTorque", f.MotorTorque())
	case box2d.WeldJointType:
		f := box2d.WeldJoint{Joint: j}
		r.F32(phase+"/LinearHertz", f.LinearHertz())
		r.F32(phase+"/LinearDampingRatio", f.LinearDampingRatio())
		r.F32(phase+"/AngularHertz", f.AngularHertz())
		r.F32(phase+"/AngularDampingRatio", f.AngularDampingRatio())
	case box2d.WheelJointType:
		f := box2d.WheelJoint{Joint: j}
		r.Bool(phase+"/IsSpringEnabled", f.SpringEnabled())
		r.F32(phase+"/SpringHertz", f.SpringHertz())
		r.F32(phase+"/SpringDampingRatio", f.SpringDampingRatio())
		r.Bool(phase+"/IsLimitEnabled", f.LimitEnabled())
		r.F32(phase+"/LowerLimit", f.LowerLimit())
		r.F32(phase+"/UpperLimit", f.UpperLimit())
		r.Bool(phase+"/IsMotorEnabled", f.MotorEnabled())
		r.F32(phase+"/MotorSpeed", f.MotorSpeed())
		r.F32(phase+"/MaxMotorTorque", f.MaxMotorTorque())
		r.F32(phase+"/MotorTorque", f.MotorTorque())
	}
}

func createParityJoint(
	family box2d.JointType,
	w *box2d.World,
	a box2d.Body, b box2d.Body,
	fa box2d.Transform, fb box2d.Transform,
	partition int,
	loaded bool,
) box2d.Joint {
	switch family {
	case box2d.DistanceJointType:
		d := box2d.DefaultDistanceJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb
		d.Length = 1

		return w.CreateDistanceJoint(&d).Joint
	case box2d.FilterJointType:
		d := box2d.DefaultFilterJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateFilterJoint(&d).Joint
	case box2d.MotorJointType:
		d := box2d.DefaultMotorJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateMotorJoint(&d).Joint
	case box2d.MoverJointType:
		d := box2d.DefaultMoverJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateMoverJoint(&d).Joint
	case box2d.PogoJointType:
		return createParityPogoJoint(w, a, b, fa, fb, partition, loaded)
	case box2d.PrismaticJointType:
		d := box2d.DefaultPrismaticJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreatePrismaticJoint(&d).Joint
	case box2d.RevoluteJointType:
		d := box2d.DefaultRevoluteJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateRevoluteJoint(&d).Joint
	case box2d.WeldJointType:
		d := box2d.DefaultWeldJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateWeldJoint(&d).Joint
	case box2d.WheelJointType:
		d := box2d.DefaultWheelJointDef()
		d.Base.BodyA, d.Base.BodyB = a, b
		d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb

		return w.CreateWheelJoint(&d).Joint
	}

	panic("unknown joint family")
}

func mutateParityJoint(j box2d.Joint, stage int) {
	forceLimit := float32(3)
	if stage == 5 {
		forceLimit = 0
	}

	if stage == 6 {
		forceLimit = 100
	}

	switch j.Type() {
	case box2d.DistanceJointType:
		mutateParityDistanceJoint(j, stage, forceLimit)
	case box2d.FilterJointType:
	case box2d.MotorJointType:
		mutateParityMotorJoint(j, stage, forceLimit)
	case box2d.MoverJointType:
		mutateParityMoverJoint(j, stage, forceLimit)
	case box2d.PogoJointType:
		mutateParityPogoJoint(j, stage)
	case box2d.PrismaticJointType:
		mutateParityPrismaticJoint(j, stage, forceLimit)
	case box2d.RevoluteJointType:
		mutateParityRevoluteJoint(j, stage, forceLimit)
	case box2d.WeldJointType:
		mutateParityWeldJoint(j, stage)
	case box2d.WheelJointType:
		mutateParityWheelJoint(j, stage, forceLimit)
	}
}

func jointRunSteps(
	r *oracle.Reader,
	action string,
	count int,
	loaded bool,
	w *box2d.World,
	j box2d.Joint,
	a box2d.Body, b box2d.Body,
) {
	for step := range count {
		i := step + 1

		if loaded {
			b.ApplyForceToCenter(box2d.Vec2{X: 6, Y: -9}, true)
			b.ApplyTorque(2.5, true)
		}

		w.Step(1.0/60.0, 4)

		if i == 1 || i == 4 || i == 12 || i == 60 {
			observeJoint(r, fmt.Sprintf("%s/step%d", action, i), w, j, a, b)
		}
	}
}

func jointBoundary(j box2d.Joint, a box2d.Body, b box2d.Body, value float32) {
	// Exact binary offset anchors and a quarter turn retain a rotated mechanism
	// without rounding the initial translation away from an equality boundary.
	fa := box2d.Transform{P: box2d.Vec2{X: 0.25, Y: -0.5}, Q: box2d.Rot{C: 1}}
	fb := box2d.Transform{P: box2d.Vec2{X: -0.25, Y: 0.5}, Q: box2d.Rot{C: 1}}

	j.SetLocalFrameA(fa)
	j.SetLocalFrameB(fb)
	a.SetTransform(box2d.Pos{X: 7, Y: -3}, box2d.Rot{S: 1})
	axis := box2d.Rot{S: 1}

	q := axis
	if j.Type() == box2d.RevoluteJointType {
		q = axis.Mul(box2d.MakeRot(value)).Mul(box2d.Rot{C: fb.Q.C, S: -fb.Q.S})
	}

	delta := axis.Rotate(box2d.Vec2{X: value})
	if j.Type() == box2d.RevoluteJointType {
		delta = box2d.Vec2{}
	}

	offset := delta.Sub(q.Rotate(fb.P))
	b.SetTransform(a.WorldPoint(fa.P).Add(offset), q)
	b.SetLinearVelocity(box2d.Vec2{})
	b.SetAngularVelocity(0)

	switch j.Type() {
	case box2d.DistanceJointType:
		box2d.DistanceJoint{Joint: j}.SetLengthRange(0.5, 1.5)
	case box2d.PrismaticJointType:
		box2d.PrismaticJoint{Joint: j}.SetLimits(-0.5, 0.5)
	case box2d.WheelJointType:
		box2d.WheelJoint{Joint: j}.SetLimits(-0.5, 0.5)
	case box2d.RevoluteJointType:
		f := box2d.RevoluteJoint{Joint: j}

		lower, upper := float32(-0.5), float32(0.5)
		if value == -0.5 {
			lower = f.Angle()
		}

		if value == 0.5 {
			upper = f.Angle()
		}

		f.SetLimits(lower, upper)
	case box2d.FilterJointType,
		box2d.MotorJointType,
		box2d.MoverJointType,
		box2d.PogoJointType,
		box2d.WeldJointType:
		// These joint families have no boundary limits to configure.
	}
}

// All expectations are exact typed observations emitted by joints.c calling
// the pinned C engine. Existing upstream world and scene cases remain unchanged.
func TestJointParity(t *testing.T) {
	for family := range 9 {
		for partner := range 3 {
			for load := range 2 {
				for partition := range 3 {
					name := fmt.Sprintf(
						"%d/partner%d/loaded%d/partition%d",
						family,
						partner,
						load,
						partition,
					)
					t.Run(name, func(t *testing.T) {
						runJointParityScenario(t, name, family, partner, load, partition)
					})
				}
			}
		}
	}
}

func jointCollisionObserve(
	r *oracle.Reader,
	phase string,
	w *box2d.World,
	a box2d.Body, b box2d.Body,
	sa box2d.Shape, sb box2d.Shape,
) {
	r.I32(phase+"/contactCount", int32(w.Counters().ContactCount))
	r.I32(phase+"/jointCount", int32(w.Counters().JointCount))
	jointBodyState(r, phase+"/bodyA", a)
	jointBodyState(r, phase+"/bodyB", b)

	contacts := make([]box2d.ContactData, 0, 2)
	for c := range a.ContactData() {
		contacts = append(contacts, c)
	}

	r.I32(phase+"/touchingCount", int32(len(contacts)))

	for i, c := range contacts {
		key := fmt.Sprintf("%s/contact%d", phase, i)
		r.Bool(key+"/shapeA", c.ShapeA == sa)
		r.Bool(key+"/shapeB", c.ShapeB == sb)
		r.Bool(key+"/valid", c.Contact.IsValid())
		m := c.Manifold
		r.Vec2(key+"/normal", m.Normal.X, m.Normal.Y)
		r.I32(key+"/pointCount", int32(m.PointCount))

		for p := range m.PointCount {
			r.F32(fmt.Sprintf("%s/point%d/separation", key, p), m.Points[p].Separation)
			r.F32(fmt.Sprintf("%s/point%d/normalImpulse", key, p), m.Points[p].NormalImpulse)
		}
	}
}

func TestJointCollisionPolicy(t *testing.T) {
	for family := range 9 {
		t.Run(strconv.Itoa(family), func(t *testing.T) {
			r := oracle.Open(t, fmt.Sprintf("joints/collision/%d", family))
			wd := box2d.DefaultWorldDef()
			wd.Gravity = box2d.Vec2{}
			w := box2d.NewWorld(&wd)
			t.Cleanup(w.Close)

			bd := box2d.DefaultBodyDef()
			bd.Type = box2d.DynamicBody
			bd.Position = box2d.Pos{X: 7, Y: -3}
			a := w.CreateBody(&bd)
			bd.Position = box2d.Pos{X: 7.5, Y: -3}
			b := w.CreateBody(&bd)
			sd := box2d.DefaultShapeDef()
			c := box2d.Circle{Radius: 1}
			sa, sb := a.CreateCircleShape(&sd, &c), b.CreateCircleShape(&sd, &c)

			w.Step(1.0/60.0, 4)
			jointCollisionObserve(r, "beforeJoint", w, a, b, sa, sb)
			a.SetAwake(false)
			b.SetAwake(false)

			identity := box2d.Transform{Q: box2d.Rot{C: 1}}
			j := createParityJoint(box2d.JointType(family), w, a, b, identity, identity, 1, false)
			jointCollisionObserve(r, "createdSleepingPair", w, a, b, sa, sb)
			w.Step(1.0/60.0, 4)
			jointCollisionObserve(r, "suppressed", w, a, b, sa, sb)

			for i, action := range []string{"allow", "reject", "allowAgain"} {
				a.SetTransform(box2d.Pos{X: 7, Y: -3}, box2d.Rot{C: 1})
				b.SetTransform(box2d.Pos{X: 7.5, Y: -3}, box2d.Rot{C: 1})
				a.SetLinearVelocity(box2d.Vec2{})
				b.SetLinearVelocity(box2d.Vec2{})
				j.SetCollideConnected(i != 1)
				j.WakeBodies()
				r.Bool(action+"/collideConnected", j.CollideConnected())
				w.Step(1.0/60.0, 4)
				jointCollisionObserve(r, action, w, a, b, sa, sb)
			}

			j.Destroy()
			a.SetTransform(box2d.Pos{X: 7, Y: -3}, box2d.Rot{C: 1})
			b.SetTransform(box2d.Pos{X: 7.5, Y: -3}, box2d.Rot{C: 1})
			w.Step(1.0/60.0, 4)
			jointCollisionObserve(r, "afterDestroy", w, a, b, sa, sb)
		})
	}
}

func mutateParityDistanceJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.DistanceJoint{Joint: j}
	f.SetSpringEnabled(stage != 0)
	f.SetLimitEnabled(stage >= 2)
	f.SetMotorEnabled(stage >= 3)
	f.SetSpringHertz(func() float32 {
		if stage == 1 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetSpringDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
	f.SetMotorSpeed(func() float32 {
		if stage == 4 {
			return -4
		}

		return 4
	}())
	f.SetMaxMotorForce(forceLimit)
	f.SetLength(func() float32 {
		if stage == 7 {
			return 0.7
		}

		return 1.3
	}())
	f.SetLengthRange(0.7, 1.3)
	f.SetSpringForceRange(-forceLimit, forceLimit)
}

func mutateParityMotorJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.MotorJoint{Joint: j}
	f.SetLinearHertz(func() float32 {
		if stage == 0 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetLinearDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
	f.SetAngularHertz(func() float32 {
		if stage == 0 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetAngularDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
	f.SetLinearVelocity(box2d.Vec2{X: func() float32 {
		if stage == 4 {
			return -4
		}

		return 4
	}(), Y: -2})
	f.SetMaxVelocityForce(forceLimit)
	f.SetMaxVelocityTorque(forceLimit)
	f.SetMaxSpringForce(forceLimit)
	f.SetMaxSpringTorque(forceLimit)
	f.SetAngularVelocity(func() float32 {
		if stage == 4 {
			return -3
		}

		return 3
	}())
}

func mutateParityMoverJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.MoverJoint{Joint: j}
	f.SetLinearVelocity(box2d.Vec2{X: func() float32 {
		if stage == 4 {
			return -4
		}

		return 4
	}(), Y: -2})
	f.SetMaxVelocityForce(box2d.Vec2{X: forceLimit, Y: 0.5 * forceLimit})
}

func mutateParityPogoJoint(j box2d.Joint, stage int) {
	f := box2d.PogoJoint{Joint: j}
	f.SetRestLength(func() float32 {
		if stage == 7 {
			return 0.3
		}

		return 1.3
	}())
	f.SetSpringHertz(func() float32 {
		if stage == 0 {
			return 0
		}

		return 5
	}())
	f.SetSpringDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
}

func parityJointTuning(stage int) (float32, float32, float32, float32) {
	hertz, damping, speed, target := float32(3), float32(0.2), float32(4), float32(0.5)

	switch stage {
	case 1:
		hertz = 0
	case 4:
		speed = -4
	case 7:
		hertz, damping, target = 9, 1, -0.5
	}

	return hertz, damping, speed, target
}

func mutateParityPrismaticJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.PrismaticJoint{Joint: j}
	hertz, damping, speed, target := parityJointTuning(stage)
	f.SetSpringEnabled(stage != 0)
	f.SetLimitEnabled(stage >= 2)
	f.SetMotorEnabled(stage >= 3)
	f.SetSpringHertz(hertz)
	f.SetSpringDampingRatio(damping)
	f.SetMotorSpeed(speed)
	f.SetMaxMotorForce(forceLimit)
	f.SetLimits(-0.5, 0.5)
	f.SetTargetTranslation(target)
}

func mutateParityRevoluteJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.RevoluteJoint{Joint: j}
	hertz, damping, speed, target := parityJointTuning(stage)
	f.SetSpringEnabled(stage != 0)
	f.SetLimitEnabled(stage >= 2)
	f.SetMotorEnabled(stage >= 3)
	f.SetSpringHertz(hertz)
	f.SetSpringDampingRatio(damping)
	f.SetMotorSpeed(speed)
	f.SetMaxMotorTorque(forceLimit)
	f.SetLimits(-0.5, 0.5)
	f.SetTargetAngle(target)
}

func mutateParityWeldJoint(j box2d.Joint, stage int) {
	f := box2d.WeldJoint{Joint: j}
	f.SetLinearHertz(func() float32 {
		if stage == 0 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetLinearDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
	f.SetAngularHertz(func() float32 {
		if stage == 0 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetAngularDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
}

func mutateParityWheelJoint(j box2d.Joint, stage int, forceLimit float32) {
	f := box2d.WheelJoint{Joint: j}
	f.SetSpringEnabled(stage != 0)
	f.SetLimitEnabled(stage >= 2)
	f.SetMotorEnabled(stage >= 3)
	f.SetSpringHertz(func() float32 {
		if stage == 1 {
			return 0
		}

		if stage == 7 {
			return 9
		}

		return 3
	}())
	f.SetSpringDampingRatio(func() float32 {
		if stage == 7 {
			return 1
		}

		return 0.2
	}())
	f.SetMotorSpeed(func() float32 {
		if stage == 4 {
			return -4
		}

		return 4
	}())
	f.SetMaxMotorTorque(forceLimit)
	f.SetLimits(-0.5, 0.5)
}

func runJointParityScenario(
	t *testing.T,
	name string,
	family int,
	partner int,
	load int,
	partition int,
) {
	t.Helper()

	loaded := load != 0
	r := oracle.Open(t, "joints/"+name)
	wd := box2d.DefaultWorldDef()
	wd.Gravity = box2d.Vec2{}
	w := box2d.NewWorld(&wd)
	t.Cleanup(w.Close)

	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.BodyType(partner)
	bd.Position = box2d.Pos{X: 7, Y: -3}
	bd.Rotation = box2d.MakeRot(0.35)
	bd.IsAwake = loaded
	a := w.CreateBody(&bd)
	bd.Type = box2d.DynamicBody
	bd.Position = box2d.Pos{X: 7.5, Y: -2}.Add(box2d.Vec2{X: 1.5 * float32(partition)})
	bd.Rotation = box2d.MakeRot(-0.2)
	b := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	sd.Density = 2
	circle := box2d.Circle{Center: box2d.Vec2{X: 0.1, Y: 0.2}, Radius: 0.4}
	a.CreateCircleShape(&sd, &circle)
	b.CreateCircleShape(&sd, &circle)

	if partition == 0 {
		b.Disable()
	}

	jointBodyState(r, "beforeCreate/bodyA", a)
	jointBodyState(r, "beforeCreate/bodyB", b)

	fa := box2d.Transform{P: box2d.Vec2{X: 0.2, Y: -0.3}, Q: box2d.MakeRot(0.2)}
	fb := box2d.Transform{P: box2d.Vec2{X: -0.1, Y: 0.25}, Q: box2d.MakeRot(-0.1)}

	j := createParityJoint(box2d.JointType(family), w, a, b, fa, fb, partition, loaded)
	if !j.IsValid() {
		t.Fatal("valid family constructor rejected mechanism")
	}

	observeJoint(r, "created", w, j, a, b)

	if partition == 0 {
		mutateParityJoint(j, 3)
		j.WakeBodies()
		observeJoint(r, "createdDisabled/mutatedAndWoken", w, j, a, b)
		jointRunSteps(r, "createdDisabled", 12, loaded, w, j, a, b)
		b.Enable()
		observeJoint(r, "createdDisabled/enabled", w, j, a, b)
	}

	jointRunSteps(r, "initial", 60, loaded, w, j, a, b)

	stages := []string{
		"featuresOff",
		"springOn",
		"limitsOn",
		"positiveMotor",
		"negativeMotor",
		"zeroCaps",
		"highCaps",
		"retargetAndDamp",
	}
	for stage, action := range stages {
		mutateParityJoint(j, stage)
		observeJoint(r, action+"/afterSetters", w, j, a, b)
		j.WakeBodies()
		jointRunSteps(r, action, 12, loaded, w, j, a, b)
	}

	switch family {
	case int(box2d.DistanceJointType),
		int(box2d.PrismaticJointType),
		int(box2d.RevoluteJointType),
		int(box2d.WheelJointType):
		values := []float32{-0.8, -0.5, 0, 0.5, 0.8}
		labels := []string{"belowLower", "atLower", "inside", "atUpper", "aboveUpper"}
		lengths := []float32{0.25, 0.5, 1, 1.5, 1.75}

		for i, value := range values {
			if family == int(box2d.DistanceJointType) {
				value = lengths[i]
			}

			jointBoundary(j, a, b, value)
			observeJoint(r, labels[i], w, j, a, b)
			j.WakeBodies()
			jointRunSteps(r, labels[i], 12, loaded, w, j, a, b)
		}
	}

	b.SetAwake(false)
	jointBodyState(r, "beforeSleepingMutation/bodyA", a)
	jointBodyState(r, "beforeSleepingMutation/bodyB", b)
	mutateParityJoint(j, 4)
	observeJoint(r, "sleepingFamilyMutation", w, j, a, b)
	j.SetConstraintTuning(17, 0.6)
	j.SetForceThreshold(4)
	j.SetTorqueThreshold(2)

	fa.P = box2d.Vec2{X: -0.35, Y: 0.15}
	fa.Q = box2d.MakeRot(-0.4)
	fb.P = box2d.Vec2{X: 0.3, Y: -0.2}
	fb.Q = box2d.MakeRot(0.45)

	j.SetLocalFrameA(fa)
	j.SetLocalFrameB(fb)
	j.SetUserData(&wd)
	r.Bool("userData", j.UserData() == &wd)
	observeJoint(r, "commonSetters", w, j, a, b)
	j.WakeBodies()
	observeJoint(r, "explicitWake", w, j, a, b)
	jointRunSteps(r, "changedFrames", 12, loaded, w, j, a, b)
	j.SetCollideConnected(true)
	observeJoint(r, "collisionEnabled", w, j, a, b)
	jointRunSteps(r, "collisionEnabled", 12, loaded, w, j, a, b)
	j.SetCollideConnected(false)
	observeJoint(r, "collisionDisabled", w, j, a, b)
	jointRunSteps(r, "collisionDisabled", 12, loaded, w, j, a, b)
	b.Disable()
	observeJoint(r, "bodyDisabled", w, j, a, b)
	jointRunSteps(r, "bodyDisabled", 12, loaded, w, j, a, b)
	b.Enable()
	observeJoint(r, "bodyEnabled", w, j, a, b)
	jointRunSteps(r, "bodyEnabled", 12, loaded, w, j, a, b)
	a.SetType(box2d.StaticBody)
	b.SetType(box2d.KinematicBody)
	observeJoint(r, "nonDynamicPair", w, j, a, b)
	jointRunSteps(r, "nonDynamicPair", 12, false, w, j, a, b)
	b.Disable()
	b.Enable()
	jointRunSteps(r, "nonDynamicReenabled", 12, false, w, j, a, b)
	b.SetType(box2d.DynamicBody)
	observeJoint(r, "dynamicRestored", w, j, a, b)
	jointRunSteps(r, "dynamicRestored", 12, loaded, w, j, a, b)
	j.Destroy()
	r.I32("jointDestroyed/count", int32(w.Counters().JointCount))
	b.SetLinearVelocity(box2d.Vec2{X: 2, Y: -1})
	b.SetAngularVelocity(1.5)

	for step := range 12 {
		i := step + 1

		w.Step(1.0/60.0, 4)

		if i == 1 || i == 12 {
			jointBodyState(r, fmt.Sprintf("freeMotion/step%d", i), b)
		}
	}

	j = createParityJoint(box2d.JointType(family), w, a, b, fa, fb, partition, loaded)
	observeJoint(r, "recreated", w, j, a, b)
	b.Destroy()
	r.I32("bodyDestroyed/jointCount", int32(w.Counters().JointCount))
	w.Step(1.0/60.0, 4)
	jointBodyState(r, "survivor", a)
}

func createParityPogoJoint(
	w *box2d.World,
	a box2d.Body, b box2d.Body,
	fa box2d.Transform, fb box2d.Transform,
	partition int,
	loaded bool,
) box2d.Joint {
	d := box2d.DefaultPogoJointDef()
	d.Base.BodyA, d.Base.BodyB = a, b
	d.Base.LocalFrameA, d.Base.LocalFrameB = fa, fb
	d.Normal = box2d.Vec2{X: 0.6, Y: 0.8}

	d.RestLength, d.Hertz, d.DampingRatio = 1, 3, 0.2
	if partition != 0 {
		d.MaxTensionForce = 2
	}

	d.MaxCompressionForce = 4
	if partition == 2 {
		d.MaxCompressionForce = 100
	}

	if loaded {
		d.Impulse, d.Velocity = 0.2, -0.5
	}

	return w.CreatePogoJoint(&d).Joint
}
