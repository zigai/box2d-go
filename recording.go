// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
)

const (
	snapFNVInit     uint64 = 14695981039346656037
	snapFNVPrime    uint64 = 1099511628211
	recMagic        uint32 = 0x43523242
	recVersionMajor uint16 = 3
	recVersionMinor uint16 = 12
)

type recHeader struct {
	magic             uint32
	versionMajor      uint16
	versionMinor      uint16
	reserved2         uint32
	lengthScale       float32
	reserved3         uint8
	pointerWidth      uint8
	bigEndian         uint8
	validationEnabled uint8
	reserved1         uint32
	snapshotSize      uint64
}

type recBuffer struct {
	data      []byte
	capacity  int
	size      int
	countOnly bool
}

// User-owned recording buffer. The world appends into it while recording; the user saves and destroys it.
type Recording struct {
	buffer            recBuffer
	recordStart       int
	lock              *sync.Mutex
	accumulatedBounds AABB
	haveBounds        bool
}

type recQueryWriter struct {
	userFcn struct {
		overlapFcn OverlapResultFunc
		castFcn    CastResultFunc
		planeFcn   PlaneResultFunc
	}
	buf         recBuffer
	countOffset int
	hitCount    uint32
}

func recBufAppend(buf *recBuffer, data []byte, size int) {
	if size <= 0 {
		return
	}
	if buf.countOnly {
		buf.size += size
		return
	}
	if size > len(data) {
		panic("box2d: short recording source")
	}
	if buf.size+size > buf.capacity {
		newCap := max(buf.capacity*2, buf.size+size+64)
		b := make([]byte, newCap)
		copy(b, buf.data[:buf.size])
		buf.data = b
		buf.capacity = newCap
	}
	copy(buf.data[buf.size:buf.size+size], data[:size])
	buf.size += size
}

func recBufFree(buf *recBuffer) {
	*buf = recBuffer{}
}

func recWU8(buf *recBuffer, v uint8) {
	b := [1]byte{}
	b[0] = v
	recBufAppend(buf, b[:], 1)
}

func recWU16(buf *recBuffer, v uint16) {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], v)
	recBufAppend(buf, b[:], 2)
}

func recWU32(buf *recBuffer, v uint32) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	recBufAppend(buf, b[:], 4)
}

func recWU64(buf *recBuffer, v uint64) {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], v)
	recBufAppend(buf, b[:], 8)
}

func recWI32(buf *recBuffer, v int32) {
	recWU32(buf, uint32(v))
}

func recWF32(buf *recBuffer, v float32) {
	recWU32(buf, math.Float32bits(v))
}

func recWBOOL(buf *recBuffer, v bool) {
	if v {
		recWU8(buf, 1)
	} else {
		recWU8(buf, 0)
	}
}

func recWVEC2(buf *recBuffer, v Vec2) {
	recWF32(buf, v.X)
	recWF32(buf, v.Y)
}

func recWROT(buf *recBuffer, v Rot) {
	recWF32(buf, v.C)
	recWF32(buf, v.S)
}

func recWXF(buf *recBuffer, v Transform) {
	recWVEC2(buf, v.P)
	recWROT(buf, v.Q)
}

func recWF64(buf *recBuffer, v float64) {
	recWU64(buf, math.Float64bits(v))
}

func recWPOSITION(buf *recBuffer, v Pos) {
	if doublePrecision {
		recWF64(buf, float64(v.X))
		recWF64(buf, float64(v.Y))
	} else {
		recWF32(buf, float32(v.X))
		recWF32(buf, float32(v.Y))
	}
}

func recWWORLDXF(buf *recBuffer, v WorldTransform) {
	recWPOSITION(buf, v.P)
	recWROT(buf, v.Q)
}

func recWBODYID(buf *recBuffer, v Body) {
	recWU64(buf, v.ID())
}

func recWSHAPEID(buf *recBuffer, v Shape) {
	recWU64(buf, v.ID())
}

func recWCHAINID(buf *recBuffer, v Chain) {
	recWU64(buf, v.ID())
}

func recWJOINTID(buf *recBuffer, v Joint) {
	recWU64(buf, v.ID())
}

func recWCIRCLE(buf *recBuffer, v Circle) {
	recWVEC2(buf, v.Center)
	recWF32(buf, v.Radius)
}

func recWCAPSULE(buf *recBuffer, v Capsule) {
	recWVEC2(buf, v.Center1)
	recWVEC2(buf, v.Center2)
	recWF32(buf, v.Radius)
}

func recWSEGMENT(buf *recBuffer, v Segment) {
	recWVEC2(buf, v.Point1)
	recWVEC2(buf, v.Point2)
}

func recWPOLYGON(buf *recBuffer, v Polygon) {
	for _, p := range v.Vertices {
		recWVEC2(buf, p)
	}
	for _, n := range v.Normals {
		recWVEC2(buf, n)
	}
	recWVEC2(buf, v.Centroid)
	recWF32(buf, v.Radius)
	recWI32(buf, int32(v.Count))
}

func recWCHAINSEG(buf *recBuffer, v ChainSegment) {
	recWVEC2(buf, v.Ghost1)
	recWSEGMENT(buf, v.Segment)
	recWVEC2(buf, v.Ghost2)
	recWI32(buf, int32(v.ChainID))
}

func recWFILTER(buf *recBuffer, v Filter) {
	recWU64(buf, v.CategoryBits)
	recWU64(buf, v.MaskBits)
	recWI32(buf, int32(v.GroupIndex))
}

func recWMATERIAL(buf *recBuffer, v SurfaceMaterial) {
	recWF32(buf, v.Friction)
	recWF32(buf, v.Restitution)
	recWF32(buf, v.RollingResistance)
	recWF32(buf, v.TangentSpeed)
	recWU64(buf, v.UserMaterialID)
	recWU32(buf, v.CustomColor)
}

func recWMASSDATA(buf *recBuffer, v MassData) {
	recWF32(buf, v.Mass)
	recWVEC2(buf, v.Center)
	recWF32(buf, v.RotationalInertia)
}

func recWLOCKS(buf *recBuffer, v MotionLocks) {
	recWBOOL(buf, v.LinearX)
	recWBOOL(buf, v.LinearY)
	recWBOOL(buf, v.AngularZ)
}

func recWSTR(buf *recBuffer, s string) {
	n := min(len(s), 65534)
	if i := strings.IndexByte(s[:n], 0); i >= 0 {
		n = i
	}
	recWU16(buf, uint16(n))
	recBufAppend(buf, []byte(s[:n]), n)
}

func recWBODYDEF(buf *recBuffer, v BodyDef) {
	recWI32(buf, int32(v.Type))
	recWPOSITION(buf, v.Position)
	recWROT(buf, v.Rotation)
	recWVEC2(buf, v.LinearVelocity)
	recWF32(buf, v.AngularVelocity)
	recWF32(buf, v.LinearDamping)
	recWF32(buf, v.AngularDamping)
	recWF32(buf, v.GravityScale)
	recWF32(buf, v.SleepThreshold)
	recWF32(buf, v.SafetyFactor)
	recWSTR(buf, v.Name)
	recWU64(buf, 0)
	recWLOCKS(buf, v.MotionLocks)
	recWBOOL(buf, v.EnableSleep)
	recWBOOL(buf, v.IsAwake)
	recWBOOL(buf, v.IsBullet)
	recWBOOL(buf, v.IsEnabled)
	recWBOOL(buf, v.AllowFastRotation)
	recWBOOL(buf, v.EnableContactRecycling)
}

func recWSHAPEDEF(buf *recBuffer, v ShapeDef) {
	recWU64(buf, 0)
	recWMATERIAL(buf, v.Material)
	recWF32(buf, v.Density)
	recWFILTER(buf, v.Filter)
	recWBOOL(buf, v.EnableCustomFiltering)
	recWBOOL(buf, v.IsSensor)
	recWBOOL(buf, v.EnableSensorEvents)
	recWBOOL(buf, v.EnableContactEvents)
	recWBOOL(buf, v.EnableHitEvents)
	recWBOOL(buf, v.EnablePreSolveEvents)
	recWBOOL(buf, v.InvokeContactCreation)
	recWBOOL(buf, v.UpdateBodyMass)
}

func recWCHAINDEF(buf *recBuffer, v ChainDef) {
	recWU64(buf, 0)
	recWI32(buf, int32(len(v.Points)))
	for _, p := range v.Points {
		recWVEC2(buf, p)
	}
	recWVEC2(buf, v.Ghost1)
	recWVEC2(buf, v.Ghost2)
	recWI32(buf, int32(len(v.Materials)))
	for _, m := range v.Materials {
		recWMATERIAL(buf, m)
	}
	recWFILTER(buf, v.Filter)
	recWBOOL(buf, v.IsLoop)
	recWBOOL(buf, v.EnableSensorEvents)
}

func recWEXPLOSIONDEF(buf *recBuffer, v ExplosionDef) {
	recWU64(buf, v.MaskBits)
	recWPOSITION(buf, v.Position)
	recWF32(buf, v.Radius)
	recWF32(buf, v.Falloff)
	recWF32(buf, v.ImpulsePerLength)
}

func recWJointBase(buf *recBuffer, base *JointDef) {
	recWU64(buf, 0)
	recWBODYID(buf, base.BodyA)
	recWBODYID(buf, base.BodyB)
	recWXF(buf, base.LocalFrameA)
	recWXF(buf, base.LocalFrameB)
	recWF32(buf, base.ForceThreshold)
	recWF32(buf, base.TorqueThreshold)
	recWF32(buf, base.ConstraintHertz)
	recWF32(buf, base.ConstraintDampingRatio)
	recWF32(buf, base.DrawScale)
	recWBOOL(buf, base.CollideConnected)
}

func recWDISTANCEJOINTDEF(buf *recBuffer, v DistanceJointDef) {
	recWJointBase(buf, &v.Base)
	recWF32(buf, v.Length)
	recWBOOL(buf, v.EnableSpring)
	recWF32(buf, v.LowerSpringForce)
	recWF32(buf, v.UpperSpringForce)
	recWF32(buf, v.Hertz)
	recWF32(buf, v.DampingRatio)
	recWBOOL(buf, v.EnableLimit)
	recWF32(buf, v.MinLength)
	recWF32(buf, v.MaxLength)
	recWBOOL(buf, v.EnableMotor)
	recWF32(buf, v.MaxMotorForce)
	recWF32(buf, v.MotorSpeed)
}

func recWFILTERJOINTDEF(buf *recBuffer, v FilterJointDef) {
	recWJointBase(buf, &v.Base)
}

func recWMOTORJOINTDEF(buf *recBuffer, v MotorJointDef) {
	recWJointBase(buf, &v.Base)
	recWVEC2(buf, v.LinearVelocity)
	recWF32(buf, v.MaxVelocityForce)
	recWF32(buf, v.AngularVelocity)
	recWF32(buf, v.MaxVelocityTorque)
	recWF32(buf, v.LinearHertz)
	recWF32(buf, v.LinearDampingRatio)
	recWF32(buf, v.MaxSpringForce)
	recWF32(buf, v.AngularHertz)
	recWF32(buf, v.AngularDampingRatio)
	recWF32(buf, v.MaxSpringTorque)
}

func recWMOVERJOINTDEF(buf *recBuffer, v MoverJointDef) {
	recWJointBase(buf, &v.Base)
	recWVEC2(buf, v.LinearVelocity)
	recWVEC2(buf, v.MaxVelocityForce)
}

func recWPOGOJOINTDEF(buf *recBuffer, v PogoJointDef) {
	recWJointBase(buf, &v.Base)
	recWVEC2(buf, v.Normal)
	recWF32(buf, v.Hertz)
	recWF32(buf, v.DampingRatio)
	recWF32(buf, v.RestLength)
	recWF32(buf, v.MaxTensionForce)
	recWF32(buf, v.MaxCompressionForce)
	recWF32(buf, v.Impulse)
	recWF32(buf, v.Velocity)
}

func recWPRISMATICJOINTDEF(buf *recBuffer, v PrismaticJointDef) {
	recWJointBase(buf, &v.Base)
	recWBOOL(buf, v.EnableSpring)
	recWF32(buf, v.Hertz)
	recWF32(buf, v.DampingRatio)
	recWF32(buf, v.TargetTranslation)
	recWBOOL(buf, v.EnableLimit)
	recWF32(buf, v.LowerTranslation)
	recWF32(buf, v.UpperTranslation)
	recWBOOL(buf, v.EnableMotor)
	recWF32(buf, v.MaxMotorForce)
	recWF32(buf, v.MotorSpeed)
}

func recWREVOLUTEJOINTDEF(buf *recBuffer, v RevoluteJointDef) {
	recWJointBase(buf, &v.Base)
	recWF32(buf, v.TargetAngle)
	recWBOOL(buf, v.EnableSpring)
	recWF32(buf, v.Hertz)
	recWF32(buf, v.DampingRatio)
	recWBOOL(buf, v.EnableLimit)
	recWF32(buf, v.LowerAngle)
	recWF32(buf, v.UpperAngle)
	recWBOOL(buf, v.EnableMotor)
	recWF32(buf, v.MaxMotorTorque)
	recWF32(buf, v.MotorSpeed)
}

func recWWELDJOINTDEF(buf *recBuffer, v WeldJointDef) {
	recWJointBase(buf, &v.Base)
	recWF32(buf, v.LinearHertz)
	recWF32(buf, v.AngularHertz)
	recWF32(buf, v.LinearDampingRatio)
	recWF32(buf, v.AngularDampingRatio)
}

func recWWHEELJOINTDEF(buf *recBuffer, v WheelJointDef) {
	recWJointBase(buf, &v.Base)
	recWBOOL(buf, v.EnableSpring)
	recWF32(buf, v.Hertz)
	recWF32(buf, v.DampingRatio)
	recWBOOL(buf, v.EnableLimit)
	recWF32(buf, v.LowerTranslation)
	recWF32(buf, v.UpperTranslation)
	recWBOOL(buf, v.EnableMotor)
	recWF32(buf, v.MaxMotorTorque)
	recWF32(buf, v.MotorSpeed)
}

func recWAABB(buf *recBuffer, v AABB) {
	recWVEC2(buf, v.LowerBound)
	recWVEC2(buf, v.UpperBound)
}

func recWQUERYFILTER(buf *recBuffer, v QueryFilter) {
	recWU64(buf, v.CategoryBits)
	recWU64(buf, v.MaskBits)
}

func recWSHAPEPROXY(buf *recBuffer, v ShapeProxy) {
	count := min(max(v.Count, 0), MaxPolygonVertices)
	recWI32(buf, int32(count))
	for i := 0; i < count; i++ {
		recWVEC2(buf, v.Points[i])
	}
	recWF32(buf, v.Radius)
}

func recWWORLDCASTOUTPUT(buf *recBuffer, v WorldCastOutput) {
	recWVEC2(buf, v.Normal)
	recWPOSITION(buf, v.Point)
	recWF32(buf, v.Fraction)
	recWI32(buf, int32(v.Iterations))
	recWBOOL(buf, v.Hit)
}

func recWRAYRESULT(buf *recBuffer, v RayResult) {
	recWSHAPEID(buf, v.Shape)
	recWPOSITION(buf, v.Point)
	recWVEC2(buf, v.Normal)
	recWF32(buf, v.Fraction)
	recWI32(buf, int32(v.NodeVisits))
	recWI32(buf, int32(v.LeafVisits))
	recWBOOL(buf, v.Hit)
}

func recWPLANERESULT(buf *recBuffer, v PlaneResult) {
	recWVEC2(buf, v.Plane.Normal)
	recWF32(buf, v.Plane.Offset)
	recWVEC2(buf, v.Point)
	recWBOOL(buf, v.Hit)
}

func recWTREESTATS(buf *recBuffer, v TreeStats) {
	recWI32(buf, int32(v.NodeVisits))
	recWI32(buf, int32(v.LeafVisits))
}

func recReserveU32(buf *recBuffer) int {
	offset := buf.size
	recWU32(buf, 0)
	return offset
}

func recPatchU32(buf *recBuffer, offset int, v uint32) {
	assert(offset >= 0 && offset <= buf.size-4, "offset >= 0 && offset + 4 <= buf->size")
	binary.LittleEndian.PutUint32(buf.data[offset:offset+4], v)
}

func recCommitRecord(rec *Recording, opcode uint8, payload []byte, payloadSize int) {
	assert(payloadSize >= 0 && payloadSize < 1<<24, "payloadSize >= 0 && payloadSize < ( 1 << 24 )")
	rec.lock.Lock()
	defer rec.lock.Unlock()
	recBeginRecord(rec, opcode)
	recBufAppend(&rec.buffer, payload, payloadSize)
	recEndRecord(rec)
}

func recQueryBegin(w *recQueryWriter) {
	*w = recQueryWriter{}
}

func recQueryCommit(rec *Recording, opcode uint8, w *recQueryWriter) {
	recCommitRecord(rec, opcode, w.buf.data, w.buf.size)
	recBufFree(&w.buf)
}

func recOverlapTrampoline(id Shape, ctx *recQueryWriter) bool {
	ret := ctx.userFcn.overlapFcn(id)
	recWSHAPEID(&ctx.buf, id)
	recWBOOL(&ctx.buf, ret)
	ctx.hitCount++
	return ret
}

func recCastTrampoline(id Shape, point Pos, normal Vec2, fraction float32, ctx *recQueryWriter) float32 {
	ret := ctx.userFcn.castFcn(id, point, normal, fraction)
	recWSHAPEID(&ctx.buf, id)
	recWPOSITION(&ctx.buf, point)
	recWVEC2(&ctx.buf, normal)
	recWF32(&ctx.buf, fraction)
	recWF32(&ctx.buf, ret)
	ctx.hitCount++
	return ret
}

func recPlaneTrampoline(id Shape, plane *PlaneResult, ctx *recQueryWriter) bool {
	ret := ctx.userFcn.planeFcn(id, plane)
	recWSHAPEID(&ctx.buf, id)
	recWPLANERESULT(&ctx.buf, *plane)
	recWBOOL(&ctx.buf, ret)
	ctx.hitCount++
	return ret
}

func recBeginRecord(rec *Recording, opcode uint8) {
	recWU8(&rec.buffer, opcode)
	rec.recordStart = rec.buffer.size
	recBufAppend(&rec.buffer, []byte{0, 0, 0}, 3)
}

func recEndRecord(rec *Recording) {
	n := rec.buffer.size - rec.recordStart - 3
	assert(n >= 0 && n < 1<<24, "payloadSize >= 0 && payloadSize < ( 1 << 24 )")
	p := rec.buffer.data[rec.recordStart:]
	p[0] = byte(n)
	p[1] = byte(n >> 8)
	p[2] = byte(n >> 16)
}

// Create a recording buffer. The buffer grows on demand. byteCapacity pre-sizes it to avoid
// reallocations during a known-length session. Pass 0 for a small default.
func NewRecording(byteCapacity int) *Recording {
	if byteCapacity <= 0 {
		byteCapacity = 65536
	}
	return &Recording{buffer: recBuffer{data: make([]byte, byteCapacity), capacity: byteCapacity}, lock: new(sync.Mutex)}
}

// Get the recorded bytes, for saving to a file or transmitting. Valid until the
// next recording call.
func (r *Recording) Data() []byte {
	return r.buffer.data[:r.buffer.size]
}

// Get the number of recorded bytes.
func (r *Recording) Size() int {
	return r.buffer.size
}

func recAccumulateBounds(rec *Recording, bounds AABB) {
	if rec.haveBounds {
		rec.accumulatedBounds = rec.accumulatedBounds.Union(bounds)
	} else {
		rec.accumulatedBounds = bounds
	}
	rec.haveBounds = true
}

func startRecordingIntoBuffer(world *World, recording *Recording) {
	recording.buffer.size = 0
	recording.recordStart = 0
	recording.haveBounds = false
	blob := recBuffer{}
	serializeWorld(world, &blob)
	recWU32(&recording.buffer, recMagic)
	recWU16(&recording.buffer, recVersionMajor)
	recWU16(&recording.buffer, recVersionMinor)
	recWU32(&recording.buffer, 0)
	recWF32(&recording.buffer, LengthUnitsPerMeter())
	recWU8(&recording.buffer, 0)
	recWU8(&recording.buffer, 0)
	recWU8(&recording.buffer, 0)
	recWBOOL(&recording.buffer, validationEnabled)
	recWU32(&recording.buffer, 0)
	recWU64(&recording.buffer, uint64(blob.size))
	recBufAppend(&recording.buffer, blob.data, blob.size)
	recBufFree(&blob)
	world.recording = recording
	var seed AABB
	if computeWorldBounds(world, &seed) {
		recAccumulateBounds(recording, seed)
	}
	recWriteStateHash(recording, &recArgsStateHash{hash: hashWorldState(world)})
}

func stopRecordingInternal(world *World) {
	if world.recording == nil {
		return
	}
	rec := world.recording
	world.recording = nil
	rb := recArgsRecordingBounds{}
	if rec.haveBounds {
		rb.bounds = rec.accumulatedBounds
	}
	recWriteRecordingBounds(rec, &rb)
	recWriteDestroyWorld(rec, &recArgsDestroyWorld{})
}

// WriteTo writes the recorded bytes to w.
func (r *Recording) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(r.Data())
	return int64(n), err
}

// ReadRecording reads a recording written by WriteTo into a new recording buffer. It returns
// ErrCorrupt for data too short to hold a header and ErrBadMagic for data that is not a recording.
func ReadRecording(src io.Reader) (*Recording, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	if len(data) < 32 {
		return nil, fmt.Errorf("%w: recording too small", ErrCorrupt)
	}
	if binary.LittleEndian.Uint32(data) != recMagic {
		return nil, fmt.Errorf("%w: recording magic 0x%08X", ErrBadMagic, binary.LittleEndian.Uint32(data))
	}
	r := NewRecording(len(data))
	copy(r.buffer.data, data)
	r.buffer.size = len(data)
	return r, nil
}

func hashWorldState(world *World) uint64 {
	hash := snapFNVInit
	for i := range world.bodies {
		b := &world.bodies[i]
		if b.id != i {
			continue
		}
		sim := getBodySim(world, b)
		hash = fnvMixPosition(hash, sim.transform.P)
		hash = fnvMixFloat(hash, sim.transform.Q.C)
		hash = fnvMixFloat(hash, sim.transform.Q.S)
		if state := getBodyState(world, b); state != nil {
			hash = fnvMixFloat(hash, state.linearVelocity.X)
			hash = fnvMixFloat(hash, state.linearVelocity.Y)
			hash = fnvMixFloat(hash, state.angularVelocity)
		}
	}
	return hash
}

func fnvMixPosition(hash uint64, p Pos) uint64 {
	if doublePrecision {
		hash = (hash ^ math.Float64bits(float64(p.X))) * snapFNVPrime
		return (hash ^ math.Float64bits(float64(p.Y))) * snapFNVPrime
	}
	hash = fnvMixFloat(hash, float32(p.X))
	return fnvMixFloat(hash, float32(p.Y))
}
