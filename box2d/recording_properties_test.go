// SPDX-License-Identifier: MIT

package box2d_test

import (
	"fmt"
	"math"
	"strings"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

// Keep small recordings directly readable; large scenes use named field
// digests with the same typed getter calls, identity paths, and exact bits.
func recordingProperties(r *oracle.Reader, world *box2d.World, prefix string) {
	if world.Counters().BodyCount <= 32 {
		recordingObserveProperties(r, world, prefix)
		return
	}

	fields := &recordingPropertyFields{prefix: prefix, byName: make(map[string]int)}
	recordingObserveProperties(fields, world, prefix)
	fields.compare(r)
}

func recordingObserveProperties(r recordingPropertyObserver, world *box2d.World, prefix string) {
	recordingObserveWorldProperties(r, world, prefix)
	recordingObserveBodyProperties(r, world, prefix)
	recordingObserveShapeProperties(r, world, prefix)
	recordingObserveJointProperties(r, world, prefix)
}

func recordingObserveWorldProperties(
	r recordingPropertyObserver,
	world *box2d.World,
	prefix string,
) {
	r.Bool(prefix+"/IsSleepingEnabled", world.SleepingEnabled())
	r.Bool(prefix+"/IsContinuousEnabled", world.ContinuousEnabled())
	r.Bool(prefix+"/IsWarmStartingEnabled", world.WarmStartingEnabled())
	r.Bool(prefix+"/IsRestitutionPropagationEnabled", world.RestitutionPropagationEnabled())
	r.F32(prefix+"/RestitutionThreshold", world.RestitutionThreshold())
	r.I32(prefix+"/RestitutionIterations", int32(world.RestitutionIterations()))
	r.F32(prefix+"/HitEventThreshold", world.HitEventThreshold())
	r.F32(prefix+"/ContactRecycleDistance", world.ContactRecycleDistance())
	r.F32(prefix+"/MaximumLinearSpeed", world.MaximumLinearSpeed())
	vGravity := world.Gravity()
	r.Vec2(prefix+"/Gravity", vGravity.X, vGravity.Y)

	h, d, s := box2d.RecordingContactTuning(world)
	r.F32(prefix+"/contactHertz", h)
	r.F32(prefix+"/contactDamping", d)
	r.F32(prefix+"/contactSpeed", s)
}

func recordingObserveBodyProperties(
	r recordingPropertyObserver,
	world *box2d.World,
	prefix string,
) {
	for _, id := range box2d.RecordingBodies(world) {
		key := fmt.Sprintf("%s/body%d", prefix, id.ID()>>32)
		r.U32(key+"/generation", uint32(uint16(id.ID())))
		r.Bool(key+"/owner", id.World() == world)
		r.I32(key+"/Type", int32(id.Type()))
		r.F32(key+"/Mass", id.Mass())
		r.F32(key+"/RotationalInertia", id.RotationalInertia())
		vLocalCenter := id.LocalCenter()
		r.Vec2(key+"/LocalCenter", vLocalCenter.X, vLocalCenter.Y)

		vLinearVelocity := id.LinearVelocity()
		r.Vec2(key+"/LinearVelocity", vLinearVelocity.X, vLinearVelocity.Y)
		r.F32(key+"/AngularVelocity", id.AngularVelocity())
		r.F32(key+"/LinearDamping", id.LinearDamping())
		r.F32(key+"/AngularDamping", id.AngularDamping())
		r.F32(key+"/GravityScale", id.GravityScale())
		r.F32(key+"/SleepThreshold", id.SleepThreshold())
		r.F32(key+"/SafetyFactor", id.SafetyFactor())
		r.Bool(key+"/IsEnabled", id.Enabled())
		r.Bool(key+"/IsAwake", id.Awake())
		r.Bool(key+"/IsSleepEnabled", id.SleepEnabled())
		r.Bool(key+"/IsBullet", id.IsBullet())
		r.Bool(key+"/IsContactRecyclingEnabled", id.ContactRecyclingEnabled())
		r.Bytes(key+"/name", []byte(id.Name()))
		locks := id.MotionLocks()
		r.Bool(key+"/lockX", locks.LinearX)
		r.Bool(key+"/lockY", locks.LinearY)
		r.Bool(key+"/lockZ", locks.AngularZ)
	}
}

func recordingObserveShapeProperties(
	r recordingPropertyObserver,
	world *box2d.World,
	prefix string,
) {
	for _, id := range box2d.RecordingShapes(world) {
		key := fmt.Sprintf("%s/shape%d", prefix, id.ID()>>32)
		r.U32(key+"/generation", uint32(uint16(id.ID())))
		r.Bool(key+"/owner", id.World() == world)
		r.I32(key+"/Type", int32(id.Type()))
		r.F32(key+"/Density", id.Density())
		r.U64(key+"/UserMaterial", id.UserMaterial())
		r.Bool(key+"/IsSensor", id.IsSensor())
		r.Bool(key+"/AreSensorEventsEnabled", id.SensorEventsEnabled())
		r.Bool(key+"/AreContactEventsEnabled", id.ContactEventsEnabled())
		r.Bool(key+"/AreHitEventsEnabled", id.HitEventsEnabled())
		r.Bool(key+"/ArePreSolveEventsEnabled", id.PreSolveEventsEnabled())
		m := id.SurfaceMaterial()
		r.F32(key+"/material/Friction", m.Friction)
		r.F32(key+"/material/Restitution", m.Restitution)
		r.F32(key+"/material/RollingResistance", m.RollingResistance)
		r.F32(key+"/material/TangentSpeed", m.TangentSpeed)
		r.U64(key+"/material/UserMaterialID", m.UserMaterialID)
		r.U32(key+"/material/CustomColor", m.CustomColor)

		f := id.Filter()
		r.U64(key+"/filter/category", f.CategoryBits)
		r.U64(key+"/filter/mask", f.MaskBits)
		r.I32(key+"/filter/group", int32(f.GroupIndex))

		switch id.Type() {
		case box2d.CircleShape:
			v := id.Circle()
			r.Vec2(key+"/geometry/Center", v.Center.X, v.Center.Y)
			r.F32(key+"/geometry/Radius", v.Radius)
		case box2d.CapsuleShape:
			v := id.Capsule()
			r.Vec2(key+"/geometry/Center1", v.Center1.X, v.Center1.Y)
			r.Vec2(key+"/geometry/Center2", v.Center2.X, v.Center2.Y)
			r.F32(key+"/geometry/Radius", v.Radius)
		case box2d.SegmentShape:
			v := id.Segment()
			r.Vec2(key+"/geometry/Point1", v.Point1.X, v.Point1.Y)
			r.Vec2(key+"/geometry/Point2", v.Point2.X, v.Point2.Y)
		case box2d.ChainSegmentShape:
			v := id.ChainSegment()
			r.Vec2(key+"/geometry/Ghost1", v.Ghost1.X, v.Ghost1.Y)
			r.Vec2(key+"/geometry/Segment.Point1", v.Segment.Point1.X, v.Segment.Point1.Y)
			r.Vec2(key+"/geometry/Segment.Point2", v.Segment.Point2.X, v.Segment.Point2.Y)
			r.Vec2(key+"/geometry/Ghost2", v.Ghost2.X, v.Ghost2.Y)
		case box2d.PolygonShape:
			v := id.Polygon()
			r.I32(key+"/geometry/count", int32(v.Count))
			r.F32(key+"/geometry/radius", v.Radius)
			r.Vec2(key+"/geometry/centroid", v.Centroid.X, v.Centroid.Y)

			for j := range v.Count {
				r.Vec2(fmt.Sprintf("%s/geometry/v%d", key, j), v.Vertices[j].X, v.Vertices[j].Y)
				r.Vec2(fmt.Sprintf("%s/geometry/n%d", key, j), v.Normals[j].X, v.Normals[j].Y)
			}
		case box2d.ShapeTypeCount:
		}
	}
}

func recordingObserveJointProperties(
	r recordingPropertyObserver,
	world *box2d.World,
	prefix string,
) {
	for _, id := range box2d.RecordingJoints(world) {
		key := fmt.Sprintf("%s/joint%d", prefix, id.ID()>>32)
		r.U32(key+"/generation", uint32(uint16(id.ID())))
		r.Bool(key+"/owner", id.World() == world)
		r.I32(key+"/Type", int32(id.Type()))
		r.Bool(key+"/CollideConnected", id.CollideConnected())
		r.F32(key+"/ForceThreshold", id.ForceThreshold())
		r.F32(key+"/TorqueThreshold", id.TorqueThreshold())
		a, b := id.LocalFrameA(), id.LocalFrameB()
		r.Vec2(key+"/frameA/p", a.P.X, a.P.Y)
		r.Rot(key+"/frameA/q", a.Q.C, a.Q.S)
		r.Vec2(key+"/frameB/p", b.P.X, b.P.Y)
		r.Rot(key+"/frameB/q", b.Q.C, b.Q.S)

		hertz, damping := id.ConstraintTuning()
		r.F32(key+"/hertz", hertz)
		r.F32(key+"/damping", damping)

		switch id.Type() {
		case box2d.DistanceJointType:
			v := box2d.DistanceJoint{Joint: id}
			r.F32(key+"/Length", v.Length())
			r.F32(key+"/MinLength", v.MinLength())
			r.F32(key+"/MaxLength", v.MaxLength())
			r.F32(key+"/SpringHertz", v.SpringHertz())
			r.F32(key+"/SpringDampingRatio", v.SpringDampingRatio())
			r.F32(key+"/MotorSpeed", v.MotorSpeed())
			r.F32(key+"/MaxMotorForce", v.MaxMotorForce())
			r.Bool(key+"/IsSpringEnabled", v.SpringEnabled())
			r.Bool(key+"/IsLimitEnabled", v.LimitEnabled())
			r.Bool(key+"/IsMotorEnabled", v.MotorEnabled())
			low, high := v.SpringForceRange()
			r.F32(key+"/forceLow", low)
			r.F32(key+"/forceHigh", high)
		case box2d.MotorJointType:
			v := box2d.MotorJoint{Joint: id}
			vLinearVelocity := v.LinearVelocity()
			r.Vec2(key+"/LinearVelocity", vLinearVelocity.X, vLinearVelocity.Y)
			r.F32(key+"/AngularVelocity", v.AngularVelocity())
			r.F32(key+"/MaxVelocityForce", v.MaxVelocityForce())
			r.F32(key+"/MaxVelocityTorque", v.MaxVelocityTorque())
			r.F32(key+"/LinearHertz", v.LinearHertz())
			r.F32(key+"/LinearDampingRatio", v.LinearDampingRatio())
			r.F32(key+"/AngularHertz", v.AngularHertz())
			r.F32(key+"/AngularDampingRatio", v.AngularDampingRatio())
			r.F32(key+"/MaxSpringForce", v.MaxSpringForce())
			r.F32(key+"/MaxSpringTorque", v.MaxSpringTorque())
		case box2d.MoverJointType:
			v := box2d.MoverJoint{Joint: id}
			vLinearVelocity := v.LinearVelocity()
			r.Vec2(key+"/LinearVelocity", vLinearVelocity.X, vLinearVelocity.Y)

			vMaxVelocityForce := v.MaxVelocityForce()
			r.Vec2(key+"/MaxVelocityForce", vMaxVelocityForce.X, vMaxVelocityForce.Y)
		case box2d.PogoJointType:
			v := box2d.PogoJoint{Joint: id}
			r.F32(key+"/RestLength", v.RestLength())
			r.F32(key+"/SpringHertz", v.SpringHertz())
			r.F32(key+"/SpringDampingRatio", v.SpringDampingRatio())
		case box2d.PrismaticJointType:
			v := box2d.PrismaticJoint{Joint: id}
			r.F32(key+"/SpringHertz", v.SpringHertz())
			r.F32(key+"/SpringDampingRatio", v.SpringDampingRatio())
			r.F32(key+"/TargetTranslation", v.TargetTranslation())
			r.F32(key+"/LowerLimit", v.LowerLimit())
			r.F32(key+"/UpperLimit", v.UpperLimit())
			r.F32(key+"/MotorSpeed", v.MotorSpeed())
			r.F32(key+"/MaxMotorForce", v.MaxMotorForce())
			r.Bool(key+"/IsSpringEnabled", v.SpringEnabled())
			r.Bool(key+"/IsLimitEnabled", v.LimitEnabled())
			r.Bool(key+"/IsMotorEnabled", v.MotorEnabled())
		case box2d.RevoluteJointType:
			v := box2d.RevoluteJoint{Joint: id}
			r.F32(key+"/SpringHertz", v.SpringHertz())
			r.F32(key+"/SpringDampingRatio", v.SpringDampingRatio())
			r.F32(key+"/TargetAngle", v.TargetAngle())
			r.F32(key+"/LowerLimit", v.LowerLimit())
			r.F32(key+"/UpperLimit", v.UpperLimit())
			r.F32(key+"/MotorSpeed", v.MotorSpeed())
			r.F32(key+"/MaxMotorTorque", v.MaxMotorTorque())
			r.Bool(key+"/IsSpringEnabled", v.SpringEnabled())
			r.Bool(key+"/IsLimitEnabled", v.LimitEnabled())
			r.Bool(key+"/IsMotorEnabled", v.MotorEnabled())
		case box2d.WeldJointType:
			v := box2d.WeldJoint{Joint: id}
			r.F32(key+"/LinearHertz", v.LinearHertz())
			r.F32(key+"/LinearDampingRatio", v.LinearDampingRatio())
			r.F32(key+"/AngularHertz", v.AngularHertz())
			r.F32(key+"/AngularDampingRatio", v.AngularDampingRatio())
		case box2d.WheelJointType:
			v := box2d.WheelJoint{Joint: id}
			r.F32(key+"/SpringHertz", v.SpringHertz())
			r.F32(key+"/SpringDampingRatio", v.SpringDampingRatio())
			r.F32(key+"/LowerLimit", v.LowerLimit())
			r.F32(key+"/UpperLimit", v.UpperLimit())
			r.F32(key+"/MotorSpeed", v.MotorSpeed())
			r.F32(key+"/MaxMotorTorque", v.MaxMotorTorque())
			r.Bool(key+"/IsSpringEnabled", v.SpringEnabled())
			r.Bool(key+"/IsLimitEnabled", v.LimitEnabled())
			r.Bool(key+"/IsMotorEnabled", v.MotorEnabled())
		case box2d.FilterJointType:
		}
	}
}

type recordingPropertyObserver interface {
	Bool(key string, value bool)
	I32(key string, value int32)
	U32(key string, value uint32)
	U64(key string, value uint64)
	F32(key string, value float32)
	Vec2(key string, x, y float32)
	Rot(key string, c, s float32)
	Bytes(key string, value []byte)
}

type recordingPropertyField struct {
	name string
	hash uint64
}

type recordingPropertyFields struct {
	prefix string
	fields []recordingPropertyField
	byName map[string]int
}

func (f *recordingPropertyFields) field(key string, kind string) *uint64 {
	relative := strings.TrimPrefix(key, f.prefix+"/")

	family := relative
	if before, after, ok := strings.Cut(relative, "/"); ok {
		for _, stem := range []string{"body", "shape", "joint"} {
			if strings.HasPrefix(before, stem) {
				family = stem + "/" + after
				break
			}
		}
	}

	i, found := f.byName[family]
	if !found {
		i = len(f.fields)
		f.byName[family] = i
		f.fields = append(f.fields, recordingPropertyField{name: family, hash: oracle.HashInit})
	}

	hash := &f.fields[i].hash
	*hash = oracle.HashU32(*hash, uint32(len(relative)))
	*hash = oracle.HashBytes(*hash, []byte(relative))
	*hash = oracle.HashBytes(*hash, []byte(kind+"\x00"))

	return hash
}

func (f *recordingPropertyFields) Bool(key string, value bool) {
	hash := f.field(key, "bool")

	bits := uint32(0)
	if value {
		bits = 1
	}

	*hash = oracle.HashU32(*hash, bits)
}

func (f *recordingPropertyFields) I32(key string, value int32) {
	hash := f.field(key, "i32")
	*hash = oracle.HashU32(*hash, uint32(value))
}

func (f *recordingPropertyFields) U32(key string, value uint32) {
	hash := f.field(key, "u32")
	*hash = oracle.HashU32(*hash, value)
}

func (f *recordingPropertyFields) U64(key string, value uint64) {
	hash := f.field(key, "u64")
	*hash = oracle.HashU64(*hash, value)
}

func (f *recordingPropertyFields) F32(key string, value float32) {
	hash := f.field(key, "f32")
	*hash = oracle.HashU32(*hash, math.Float32bits(value))
}

func (f *recordingPropertyFields) Vec2(key string, x float32, y float32) {
	hash := f.field(key, "vec2")
	*hash = oracle.HashF32(oracle.HashF32(*hash, x), y)
}

func (f *recordingPropertyFields) Rot(key string, c float32, s float32) {
	hash := f.field(key, "rot")
	*hash = oracle.HashRot(*hash, c, s)
}

func (f *recordingPropertyFields) Bytes(key string, value []byte) {
	hash := f.field(key, "bytes")
	*hash = oracle.HashBytes(oracle.HashU64(*hash, uint64(len(value))), value)
}

func (f *recordingPropertyFields) compare(r *oracle.Reader) {
	r.I32(f.prefix+"/digest/fieldCount", int32(len(f.fields)))

	for _, field := range f.fields {
		r.U64(f.prefix+"/digest/"+field.name, field.hash)
	}
}
