// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

const (
	recKeyframeIntervalDefault int  = 16
	recKeyframeBudgetDefault   uint = 512 * 1024 * 1024
)

// The kind of a recorded spatial query, matching the public query and cast functions.
type ReplayQueryType int

const (
	ReplayQueryOverlapAABB ReplayQueryType = iota
	ReplayQueryOverlapShape
	ReplayQueryCastRay
	ReplayQueryCastShape
	ReplayQueryCollideMover
	ReplayQueryCastRayClosest
	ReplayQueryCastMover
	ReplayQueryShapeTestPoint
	ReplayQueryShapeRayCast
)

type recQueryKind int

const (
	recqOverlapAABB recQueryKind = iota
	recqOverlapShape
	recqCastRay
	recqCastShape
	recqCollideMover
	recqCastRayClosest
	recqCastMover
	recqShapeTestPoint
	recqShapeRayCast
)

// Static metadata describing a recording, resolved once when the player opens the file.
type ReplayInfo struct {
	FrameCount   int
	WorkerCount  int
	TimeStep     float32
	SubStepCount int
	LengthScale  float32
	Bounds       AABB
}

// A spatial query recorded during a replayed frame, exposed for inspection.
type ReplayQueryInfo struct {
	Type        ReplayQueryType
	Filter      QueryFilter
	AABB        AABB
	Origin      Pos
	Translation Vec2
	Shape       Shape
	HitCount    int
}

// One result of a recorded spatial query.
type ReplayQueryHit struct {
	Shape    Shape
	Point    Pos
	Normal   Vec2
	Fraction float32
}

type recRecordedHit struct {
	id          Shape
	point       Pos
	normal      Vec2
	fraction    float32
	plane       PlaneResult
	userReturnF float32
	userReturnB bool
}

type recDrawQuery struct {
	kind         int
	filter       QueryFilter
	aabb         AABB
	proxy        ShapeProxy
	mover        Capsule
	origin       Pos
	translation  Vec2
	boolResult   bool
	castFraction float32
	castOut      WorldCastOutput
	shape        Shape
	hitStart     int
	hitCount     int
}

type recReader struct {
	data             []byte
	size             int
	cursor           int
	replayWorldID    *World
	workerCount      int
	ok               bool
	diverged         bool
	chainPoints      []Vec2
	chainPointCap    int
	chainMaterials   []SurfaceMaterial
	chainMaterialCap int
	hits             []recRecordedHit
	hitCap           int
	owner            *Replay
}

type recKeyframe struct {
	image         []byte
	imageSize     int
	imageCapacity int
	frame         int
	cursor        int
	bodyIDs       []Body
	bodyIDCount   int
	divergeFrame  int
	diverged      bool
}

// Opaque handle for incremental playback of a recording.
type Replay struct {
	data                 []byte
	size                 int
	headerEnd            int
	lengthScale          float32
	previousLengthScale  float32
	frame                int
	frameCount           int
	recordedWorkerCount  int
	recordedDt           float32
	recordedSubStepCount int
	bounds               AABB
	divergeFrame         int
	atEnd                bool
	rdr                  recReader
	frameQueries         []recDrawQuery
	frameQueryCount      int
	frameQueryCap        int
	frameHits            []recRecordedHit
	frameHitCount        int
	frameHitCap          int
	bodyIDs              []Body
	bodyIDCount          int
	bodyIDCap            int
	frame0Image          []byte
	frame0Size           int
	frame0BodyIDs        []Body
	frame0BodyIDCount    int
	keyframes            []recKeyframe
	keyframeCount        int
	keyframeCapacity     int
	keyframeBudget       uint
	keyframeBytes        uint
	keyframeMinInterval  int
	keyframeInterval     int
	lastKeyframeFrame    int
}

type recReplayQueryCtx struct {
	rdr    *recReader
	hits   []recRecordedHit
	count  int
	cursor int
}

func recRdrCheck(rdr *recReader, size int) {
	if size < 0 || rdr.cursor < 0 || rdr.cursor > rdr.size || size > rdr.size-rdr.cursor || rdr.size > len(rdr.data) {
		rdr.ok = false
	}
}

func recRU8(rdr *recReader) uint8 {
	recRdrCheck(rdr, 1)

	if !rdr.ok {
		return 0
	}

	v := uint8(rdr.data[rdr.cursor+0]) << 0
	rdr.cursor += 1

	return v
}

func recRU16(rdr *recReader) uint16 {
	recRdrCheck(rdr, 2)

	if !rdr.ok {
		return 0
	}

	v := uint16(rdr.data[rdr.cursor+0])<<0 | uint16(rdr.data[rdr.cursor+1])<<8
	rdr.cursor += 2

	return v
}

func recRU24(rdr *recReader) uint32 {
	recRdrCheck(rdr, 3)

	if !rdr.ok {
		return 0
	}

	v := uint32(rdr.data[rdr.cursor+0])<<0 | uint32(rdr.data[rdr.cursor+1])<<8 | uint32(rdr.data[rdr.cursor+2])<<16
	rdr.cursor += 3

	return v
}

func recRU32(rdr *recReader) uint32 {
	recRdrCheck(rdr, 4)

	if !rdr.ok {
		return 0
	}

	v := uint32(rdr.data[rdr.cursor+0])<<0 | uint32(rdr.data[rdr.cursor+1])<<8 | uint32(rdr.data[rdr.cursor+2])<<16 | uint32(rdr.data[rdr.cursor+3])<<24
	rdr.cursor += 4

	return v
}

func recRU64(rdr *recReader) uint64 {
	recRdrCheck(rdr, 8)

	if !rdr.ok {
		return 0
	}

	v := uint64(rdr.data[rdr.cursor+0])<<0 | uint64(rdr.data[rdr.cursor+1])<<8 | uint64(rdr.data[rdr.cursor+2])<<16 | uint64(rdr.data[rdr.cursor+3])<<24 | uint64(rdr.data[rdr.cursor+4])<<32 | uint64(rdr.data[rdr.cursor+5])<<40 | uint64(rdr.data[rdr.cursor+6])<<48 | uint64(rdr.data[rdr.cursor+7])<<56
	rdr.cursor += 8

	return v
}

func recRI32(rdr *recReader) int32 {
	return int32(recRU32(rdr))
}

func recRF32(rdr *recReader) float32 {
	return math.Float32frombits(recRU32(rdr))
}

func recRBOOL(rdr *recReader) bool {
	return recRU8(rdr) != 0
}

func recRVEC2(rdr *recReader) Vec2 {
	var v Vec2

	v.X = recRF32(rdr)
	v.Y = recRF32(rdr)

	return v
}

func recRROT(rdr *recReader) Rot {
	var r Rot

	r.C = recRF32(rdr)
	r.S = recRF32(rdr)

	return r
}

func recRXF(rdr *recReader) Transform {
	var xf Transform

	xf.P = recRVEC2(rdr)
	xf.Q = recRROT(rdr)

	return xf
}

func recRF64(rdr *recReader) float64 {
	return math.Float64frombits(recRU64(rdr))
}

func recRPOSITION(rdr *recReader) Pos {
	if doublePrecision {
		return Pos{X: posScalar(recRF64(rdr)), Y: posScalar(recRF64(rdr))}
	}

	return Pos{X: posScalar(recRF32(rdr)), Y: posScalar(recRF32(rdr))}
}

func recRWORLDXF(rdr *recReader) WorldTransform {
	var t WorldTransform

	t.P = recRPOSITION(rdr)
	t.Q = recRROT(rdr)

	return t
}

func recRBODYID(rdr *recReader) Body {
	x := recRU64(rdr)
	return Body{index1: int32(x >> 32), generation: uint16(x)}
}

func recRSHAPEID(rdr *recReader) Shape {
	x := recRU64(rdr)
	return Shape{index1: int32(x >> 32), generation: uint16(x)}
}

func recRCHAINID(rdr *recReader) Chain {
	x := recRU64(rdr)
	return Chain{index1: int32(x >> 32), generation: uint16(x)}
}

func recRJOINTID(rdr *recReader) Joint {
	x := recRU64(rdr)
	return Joint{index1: int32(x >> 32), generation: uint16(x)}
}

func recRdrBlob(rdr *recReader, out []byte, size int) {
	recRdrCheck(rdr, size)

	if !rdr.ok {
		clear(out)
		return
	}

	copy(out[:size], rdr.data[rdr.cursor:rdr.cursor+size])
	rdr.cursor += size
}

func recRCIRCLE(rdr *recReader) Circle {
	recRdrCheck(rdr, 12)

	if !rdr.ok {
		return Circle{}
	}

	return Circle{Center: recRVEC2(rdr), Radius: recRF32(rdr)}
}

func recRCAPSULE(rdr *recReader) Capsule {
	recRdrCheck(rdr, 20)

	if !rdr.ok {
		return Capsule{}
	}

	return Capsule{Center1: recRVEC2(rdr), Center2: recRVEC2(rdr), Radius: recRF32(rdr)}
}

func recRSEGMENT(rdr *recReader) Segment {
	recRdrCheck(rdr, 16)

	if !rdr.ok {
		return Segment{}
	}

	return Segment{Point1: recRVEC2(rdr), Point2: recRVEC2(rdr)}
}

func recRPOLYGON(rdr *recReader) Polygon {
	recRdrCheck(rdr, 144)

	if !rdr.ok {
		return Polygon{}
	}

	var p Polygon
	for i := range p.Vertices {
		p.Vertices[i] = recRVEC2(rdr)
	}

	for i := range p.Normals {
		p.Normals[i] = recRVEC2(rdr)
	}

	p.Centroid = recRVEC2(rdr)
	p.Radius = recRF32(rdr)
	p.Count = int(recRI32(rdr))

	return p
}

func recRCHAINSEG(rdr *recReader) ChainSegment {
	recRdrCheck(rdr, 36)

	if !rdr.ok {
		return ChainSegment{}
	}

	return ChainSegment{Ghost1: recRVEC2(rdr), Segment: recRSEGMENT(rdr), Ghost2: recRVEC2(rdr), ChainID: int(recRI32(rdr))}
}

func recRFILTER(rdr *recReader) Filter {
	var f Filter

	f.CategoryBits = recRU64(rdr)
	f.MaskBits = recRU64(rdr)
	f.GroupIndex = int(recRI32(rdr))

	return f
}

func recRMATERIAL(rdr *recReader) SurfaceMaterial {
	m := DefaultSurfaceMaterial()
	m.Friction = recRF32(rdr)
	m.Restitution = recRF32(rdr)
	m.RollingResistance = recRF32(rdr)
	m.TangentSpeed = recRF32(rdr)
	m.UserMaterialID = recRU64(rdr)
	m.CustomColor = recRU32(rdr)

	return m
}

func recRMASSDATA(rdr *recReader) MassData {
	var md MassData

	md.Mass = recRF32(rdr)
	md.Center = recRVEC2(rdr)
	md.RotationalInertia = recRF32(rdr)

	return md
}

func recRLOCKS(rdr *recReader) MotionLocks {
	var locks MotionLocks

	locks.LinearX = recRBOOL(rdr)
	locks.LinearY = recRBOOL(rdr)
	locks.AngularZ = recRBOOL(rdr)

	return locks
}

func recRSTR(rdr *recReader) string {
	length := int(recRU16(rdr))
	if length == 0xffff {
		return ""
	}

	n := min(length, NameLength)
	recRdrCheck(rdr, length)

	var name string

	if rdr.ok && n > 0 {
		value := rdr.data[rdr.cursor : rdr.cursor+n]
		if end := bytes.IndexByte(value, 0); end >= 0 {
			value = value[:end]
		}

		name = string(value)
	}

	rdr.cursor += length

	return name
}

func recRBODYDEF(rdr *recReader) BodyDef {
	def := DefaultBodyDef()
	def.Type = BodyType(recRI32(rdr))
	def.Position = recRPOSITION(rdr)
	def.Rotation = recRROT(rdr)
	def.LinearVelocity = recRVEC2(rdr)
	def.AngularVelocity = recRF32(rdr)
	def.LinearDamping = recRF32(rdr)
	def.AngularDamping = recRF32(rdr)
	def.GravityScale = recRF32(rdr)
	def.SleepThreshold = recRF32(rdr)
	def.SafetyFactor = recRF32(rdr)

	// recRSTR handles the over-length clamp and skips the full recorded length, so the
	// cursor stays aligned even for names longer than NameLength. Valid until the create call.
	def.Name = recRSTR(rdr)

	_ = recRU64(rdr) // userData (not preserved)
	def.MotionLocks = recRLOCKS(rdr)
	def.EnableSleep = recRBOOL(rdr)
	def.IsAwake = recRBOOL(rdr)
	def.IsBullet = recRBOOL(rdr)
	def.IsEnabled = recRBOOL(rdr)
	def.AllowFastRotation = recRBOOL(rdr)
	def.EnableContactRecycling = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recRSHAPEDEF(rdr *recReader) ShapeDef {
	def := DefaultShapeDef()
	_ = recRU64(rdr) // userData (not preserved)
	def.Material = recRMATERIAL(rdr)
	def.Density = recRF32(rdr)
	def.Filter = recRFILTER(rdr)
	def.EnableCustomFiltering = recRBOOL(rdr)
	def.IsSensor = recRBOOL(rdr)
	def.EnableSensorEvents = recRBOOL(rdr)
	def.EnableContactEvents = recRBOOL(rdr)
	def.EnableHitEvents = recRBOOL(rdr)
	def.EnablePreSolveEvents = recRBOOL(rdr)
	def.InvokeContactCreation = recRBOOL(rdr)
	def.UpdateBodyMass = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recRCHAINDEF(rdr *recReader) ChainDef {
	def := DefaultChainDef()
	_ = recRU64(rdr)

	pointCount := max(int(recRI32(rdr)), 0)

	if !recReserveScratch(rdr, &rdr.chainPoints, &rdr.chainPointCap, pointCount, 8) {
		pointCount = 0
	}

	for i := range pointCount {
		rdr.chainPoints[i] = recRVEC2(rdr)
	}

	def.Points = rdr.chainPoints[:pointCount]
	def.Ghost1 = recRVEC2(rdr)
	def.Ghost2 = recRVEC2(rdr)

	materialCount := max(int(recRI32(rdr)), 0)

	if !recReserveScratch(rdr, &rdr.chainMaterials, &rdr.chainMaterialCap, materialCount, 32) {
		materialCount = 0
	}

	for i := range materialCount {
		rdr.chainMaterials[i] = recRMATERIAL(rdr)
	}

	def.Materials = rdr.chainMaterials[:materialCount]
	def.Filter = recRFILTER(rdr)
	def.IsLoop = recRBOOL(rdr)
	def.EnableSensorEvents = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recREXPLOSIONDEF(rdr *recReader) ExplosionDef {
	def := DefaultExplosionDef()
	def.MaskBits = recRU64(rdr)
	def.Position = recRPOSITION(rdr)
	def.Radius = recRF32(rdr)
	def.Falloff = recRF32(rdr)
	def.ImpulsePerLength = recRF32(rdr)

	return def
}

func recRJointBase(rdr *recReader, base *JointDef) {
	_ = recRU64(rdr) // userData
	base.BodyA = recRBODYID(rdr)
	base.BodyB = recRBODYID(rdr)
	base.LocalFrameA = recRXF(rdr)
	base.LocalFrameB = recRXF(rdr)
	base.ForceThreshold = recRF32(rdr)
	base.TorqueThreshold = recRF32(rdr)
	base.ConstraintHertz = recRF32(rdr)
	base.ConstraintDampingRatio = recRF32(rdr)
	base.DrawScale = recRF32(rdr)
	base.CollideConnected = recRBOOL(rdr)
	base.UserData = nil
}

func recRDISTANCEJOINTDEF(rdr *recReader) DistanceJointDef {
	def := DefaultDistanceJointDef()
	recRJointBase(rdr, &def.Base)
	def.Length = recRF32(rdr)
	def.EnableSpring = recRBOOL(rdr)
	def.LowerSpringForce = recRF32(rdr)
	def.UpperSpringForce = recRF32(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.MinLength = recRF32(rdr)
	def.MaxLength = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorForce = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRFILTERJOINTDEF(rdr *recReader) FilterJointDef {
	def := DefaultFilterJointDef()
	recRJointBase(rdr, &def.Base)

	return def
}

func recRMOTORJOINTDEF(rdr *recReader) MotorJointDef {
	def := DefaultMotorJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearVelocity = recRVEC2(rdr)
	def.MaxVelocityForce = recRF32(rdr)
	def.AngularVelocity = recRF32(rdr)
	def.MaxVelocityTorque = recRF32(rdr)
	def.LinearHertz = recRF32(rdr)
	def.LinearDampingRatio = recRF32(rdr)
	def.MaxSpringForce = recRF32(rdr)
	def.AngularHertz = recRF32(rdr)
	def.AngularDampingRatio = recRF32(rdr)
	def.MaxSpringTorque = recRF32(rdr)

	return def
}

func recRMOVERJOINTDEF(rdr *recReader) MoverJointDef {
	def := DefaultMoverJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearVelocity = recRVEC2(rdr)
	def.MaxVelocityForce = recRVEC2(rdr)

	return def
}

func recRPRISMATICJOINTDEF(rdr *recReader) PrismaticJointDef {
	def := DefaultPrismaticJointDef()
	recRJointBase(rdr, &def.Base)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.TargetTranslation = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerTranslation = recRF32(rdr)
	def.UpperTranslation = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorForce = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRPOGOJOINTDEF(rdr *recReader) PogoJointDef {
	def := DefaultPogoJointDef()
	recRJointBase(rdr, &def.Base)
	def.Normal = recRVEC2(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.RestLength = recRF32(rdr)
	def.MaxTensionForce = recRF32(rdr)
	def.MaxCompressionForce = recRF32(rdr)
	def.Impulse = recRF32(rdr)
	def.Velocity = recRF32(rdr)

	return def
}

func recRREVOLUTEJOINTDEF(rdr *recReader) RevoluteJointDef {
	def := DefaultRevoluteJointDef()
	recRJointBase(rdr, &def.Base)
	def.TargetAngle = recRF32(rdr)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerAngle = recRF32(rdr)
	def.UpperAngle = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorTorque = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRWELDJOINTDEF(rdr *recReader) WeldJointDef {
	def := DefaultWeldJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearHertz = recRF32(rdr)
	def.AngularHertz = recRF32(rdr)
	def.LinearDampingRatio = recRF32(rdr)
	def.AngularDampingRatio = recRF32(rdr)

	return def
}

func recRWHEELJOINTDEF(rdr *recReader) WheelJointDef {
	def := DefaultWheelJointDef()
	recRJointBase(rdr, &def.Base)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerTranslation = recRF32(rdr)
	def.UpperTranslation = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorTorque = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRAABB(rdr *recReader) AABB {
	var v AABB

	v.LowerBound = recRVEC2(rdr)
	v.UpperBound = recRVEC2(rdr)

	return v
}

func recRQUERYFILTER(rdr *recReader) QueryFilter {
	var f QueryFilter

	f.CategoryBits = recRU64(rdr)
	f.MaskBits = recRU64(rdr)

	return f
}

func recRSHAPEPROXY(rdr *recReader) ShapeProxy {
	var p ShapeProxy

	count := min(max(int(recRI32(rdr)), 0), MaxPolygonVertices)

	p.Count = count
	for i := range count {
		p.Points[i] = recRVEC2(rdr)
	}

	p.Radius = recRF32(rdr)

	return p
}

func recRWORLDCASTOUTPUT(rdr *recReader) WorldCastOutput {
	var v WorldCastOutput

	v.Normal = recRVEC2(rdr)
	v.Point = recRPOSITION(rdr)
	v.Fraction = recRF32(rdr)
	v.Iterations = int(recRI32(rdr))
	v.Hit = recRBOOL(rdr)

	return v
}

func recRRAYRESULT(rdr *recReader) RayResult {
	var v RayResult
	// ShapeID has no world; recMakeShapeID is applied at compare time
	v.Shape = recRSHAPEID(rdr)
	v.Point = recRPOSITION(rdr)
	v.Normal = recRVEC2(rdr)
	v.Fraction = recRF32(rdr)
	v.NodeVisits = int(recRI32(rdr))
	v.LeafVisits = int(recRI32(rdr))
	v.Hit = recRBOOL(rdr)

	return v
}

func recRPLANERESULT(rdr *recReader) PlaneResult {
	var v PlaneResult

	v.Plane.Normal = recRVEC2(rdr)
	v.Plane.Offset = recRF32(rdr)
	v.Point = recRVEC2(rdr)
	v.Hit = recRBOOL(rdr)

	return v
}

func recRTREESTATS(rdr *recReader) TreeStats {
	var v TreeStats

	v.NodeVisits = int(recRI32(rdr))
	v.LeafVisits = int(recRI32(rdr))

	return v
}

func recReserveScratch[T any](rdr *recReader, data *[]T, capacity *int, need int, elemSize int) bool {
	remaining := rdr.size - rdr.cursor
	if need < 0 || remaining < 0 || need > remaining || elemSize <= 0 || need > math.MaxInt32/elemSize {
		rdr.ok = false
		return false
	}

	if need <= *capacity {
		return true
	}

	newCap := need
	if need <= math.MaxInt32/elemSize-8 {
		newCap = need + 8
	}

	*data = make([]T, newCap)
	*capacity = newCap

	return true
}

func recGrow[T any](data *[]T, capacity *int, need int, keep int, elemSize int) {
	if need <= *capacity {
		return
	}

	newCap := 8
	if *capacity != 0 {
		newCap = 2 * *capacity
	}

	if newCap < need {
		newCap = need
	}

	grown := make([]T, newCap)
	if *data != nil && keep > 0 {
		copy(grown[:keep], (*data)[:keep])
	}

	*data = grown
	*capacity = newCap
}

func recEnsureHits(rdr *recReader, n int) {
	recReserveScratch(rdr, &rdr.hits, &rdr.hitCap, n, int(unsafe.Sizeof(recRecordedHit{})))
}

func recMakeBodyID(rdr *recReader, recorded Body) Body {
	return Body{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeShapeID(rdr *recReader, recorded Shape) Shape {
	return Shape{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeChainID(rdr *recReader, recorded Chain) Chain {
	return Chain{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeJointID(rdr *recReader, recorded Joint) Joint {
	return Joint{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recCheckID(rdr *recReader, kind string, gotIndex int, gotGen uint32, recIndex int, recGen uint32) {
	if gotIndex != recIndex || gotGen != recGen {
		rdr.replayWorldID.logger.Warn("box2d: replay id mismatch", "kind", kind, "recordedIndex", recIndex, "recordedGeneration", recGen, "index", gotIndex, "generation", gotGen)
		rdr.ok = false
	}
}

func recCheckBodyID(rdr *recReader, got Body, rec Body) {
	recCheckID(rdr, "body", int(got.index1), uint32(got.generation), int(rec.index1), uint32(rec.generation))
}

func recCheckShapeID(rdr *recReader, got Shape, rec Shape) {
	recCheckID(rdr, "shape", int(got.index1), uint32(got.generation), int(rec.index1), uint32(rec.generation))
}

func recCheckChainID(rdr *recReader, got Chain, rec Chain) {
	recCheckID(rdr, "chain", int(got.index1), uint32(got.generation), int(rec.index1), uint32(rec.generation))
}

func recCheckJointID(rdr *recReader, got Joint, rec Joint) {
	recCheckID(rdr, "joint", int(got.index1), uint32(got.generation), int(rec.index1), uint32(rec.generation))
}

func recDispatchDestroyWorld(a *recArgsDestroyWorld, rdr *recReader) {
	_ = a
	_ = rdr
	// The recorded session ended here. The player owns the replay world's lifetime and tears it
	// down in Destroy/Restart, so a viewer can keep drawing the final step. There is
	// one world per recording and this is always the last record, so leaving it alive is safe.
}

func recDispatchStep(a *recArgsStep, rdr *recReader) {
	_ = a
	rdr.replayWorldID.Step(a.dt, int(a.subStepCount))
}

func recDispatchWorldEnableSleeping(a *recArgsWorldEnableSleeping, rdr *recReader) {
	rdr.replayWorldID.SetSleepingEnabled(a.flag)
}

func recDispatchWorldEnableContinuous(a *recArgsWorldEnableContinuous, rdr *recReader) {
	rdr.replayWorldID.SetContinuousEnabled(a.flag)
}

func recDispatchWorldSetRestitutionThreshold(a *recArgsWorldSetRestitutionThreshold, rdr *recReader) {
	rdr.replayWorldID.SetRestitutionThreshold(a.value)
}

func recDispatchWorldSetHitEventThreshold(a *recArgsWorldSetHitEventThreshold, rdr *recReader) {
	rdr.replayWorldID.SetHitEventThreshold(a.value)
}

func recDispatchWorldSetGravity(a *recArgsWorldSetGravity, rdr *recReader) {
	rdr.replayWorldID.SetGravity(a.gravity)
}

func recDispatchWorldExplode(a *recArgsWorldExplode, rdr *recReader) {
	rdr.replayWorldID.Explode(&a.def)
}

func recDispatchWorldSetContactTuning(a *recArgsWorldSetContactTuning, rdr *recReader) {
	rdr.replayWorldID.SetContactTuning(a.hertz, a.dampingRatio, a.pushSpeed)
}

func recDispatchWorldSetContactRecycleDistance(a *recArgsWorldSetContactRecycleDistance, rdr *recReader) {
	rdr.replayWorldID.SetContactRecycleDistance(a.recycleDistance)
}

func recDispatchWorldSetMaximumLinearSpeed(a *recArgsWorldSetMaximumLinearSpeed, rdr *recReader) {
	rdr.replayWorldID.SetMaximumLinearSpeed(a.maximumLinearSpeed)
}

func recDispatchWorldEnableWarmStarting(a *recArgsWorldEnableWarmStarting, rdr *recReader) {
	rdr.replayWorldID.SetWarmStartingEnabled(a.flag)
}

func recDispatchWorldRebuildStaticTree(a *recArgsWorldRebuildStaticTree, rdr *recReader) {
	_ = a

	rdr.replayWorldID.RebuildStaticTree()
}

func recDispatchWorldSetRestitutionIterations(a *recArgsWorldSetRestitutionIterations, rdr *recReader) {
	rdr.replayWorldID.SetRestitutionIterations(int(a.iterations))
}

func recDispatchWorldEnableRestitutionPropagation(a *recArgsWorldEnableRestitutionPropagation, rdr *recReader) {
	rdr.replayWorldID.SetRestitutionPropagationEnabled(a.flag)
}

func recTrackBodyCreate(player *Replay, id Body) {
	recGrow(&player.bodyIDs, &player.bodyIDCap, player.bodyIDCount+1, player.bodyIDCount, 8)
	player.bodyIDs[player.bodyIDCount] = id
	player.bodyIDCount++
}

func recTrackBodyDestroy(player *Replay, id Body) {
	for i := range player.bodyIDCount {
		if player.bodyIDs[i] == id {
			player.bodyIDs[i] = Body{}
			return
		}
	}
}

func recSeedBodyIDs(player *Replay) {
	world := player.rdr.replayWorldID
	player.bodyIDCount = 0

	for i := range world.bodies {
		if world.bodies[i].id != i {
			continue
		}

		recTrackBodyCreate(player, makeBodyId(world, i))
	}
}

func recDispatchCreateBody(a *recArgsCreateBody, rdr *recReader) {
	// Recorded id is appended after args (written before recEndRecord)
	recId := recRBODYID(rdr)
	gotId := rdr.replayWorldID.CreateBody(&a.def)
	recCheckBodyID(rdr, gotId, recId)

	if rdr.owner != nil {
		recTrackBodyCreate(rdr.owner, gotId)
	}
}

func recDispatchDestroyBody(a *recArgsDestroyBody, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	if rdr.owner != nil {
		recTrackBodyDestroy(rdr.owner, id)
	}

	id.Destroy()
}

func recDispatchBodySetTransform(a *recArgsBodySetTransform, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	id.SetTransform(a.position, a.rotation)
}

func recDispatchBodySetLinearVelocity(a *recArgsBodySetLinearVelocity, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	id.SetLinearVelocity(a.v)
}

func recDispatchBodySetType(a *recArgsBodySetType, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetType(BodyType(a.bodyType))
}

func recDispatchBodySetName(a *recArgsBodySetName, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetName(a.name)
}

func recDispatchBodySetAngularVelocity(a *recArgsBodySetAngularVelocity, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAngularVelocity(a.w)
}

func recDispatchBodySetTargetTransform(a *recArgsBodySetTargetTransform, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetTargetTransform(a.target, a.timeStep, a.wake)
}

func recDispatchBodyApplyForce(a *recArgsBodyApplyForce, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyForce(a.force, a.point, a.wake)
}

func recDispatchBodyApplyForceToCenter(a *recArgsBodyApplyForceToCenter, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyForceToCenter(a.force, a.wake)
}

func recDispatchBodyApplyTorque(a *recArgsBodyApplyTorque, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyTorque(a.torque, a.wake)
}

func recDispatchBodyClearForces(a *recArgsBodyClearForces, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ClearForces()
}

func recDispatchBodyApplyLinearImpulse(a *recArgsBodyApplyLinearImpulse, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyLinearImpulse(a.impulse, a.point, a.wake)
}

func recDispatchBodyApplyLinearImpulseToCenter(a *recArgsBodyApplyLinearImpulseToCenter, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyLinearImpulseToCenter(a.impulse, a.wake)
}

func recDispatchBodyApplyAngularImpulse(a *recArgsBodyApplyAngularImpulse, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyAngularImpulse(a.impulse, a.wake)
}

func recDispatchBodySetMassData(a *recArgsBodySetMassData, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetMassData(a.massData)
}

func recDispatchBodyUpdateMassFromShapes(a *recArgsBodyUpdateMassFromShapes, rdr *recReader) {
	recMakeBodyID(rdr, a.body).UpdateMassFromShapes()
}

func recDispatchBodySetLinearDamping(a *recArgsBodySetLinearDamping, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetLinearDamping(a.damping)
}

func recDispatchBodySetAngularDamping(a *recArgsBodySetAngularDamping, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAngularDamping(a.damping)
}

func recDispatchBodySetGravityScale(a *recArgsBodySetGravityScale, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetGravityScale(a.scale)
}

func recDispatchBodySetAwake(a *recArgsBodySetAwake, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAwake(a.awake)
}

func recDispatchBodyWakeTouching(a *recArgsBodyWakeTouching, rdr *recReader) {
	recMakeBodyID(rdr, a.body).WakeTouching()
}

func recDispatchBodyEnableSleep(a *recArgsBodyEnableSleep, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSleepEnabled(a.flag)
}

func recDispatchBodySetSleepThreshold(a *recArgsBodySetSleepThreshold, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSleepThreshold(a.threshold)
}

func recDispatchBodySetSafetyFactor(a *recArgsBodySetSafetyFactor, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSafetyFactor(a.value)
}

func recDispatchBodyDisable(a *recArgsBodyDisable, rdr *recReader) {
	recMakeBodyID(rdr, a.body).Disable()
}

func recDispatchBodyEnable(a *recArgsBodyEnable, rdr *recReader) {
	recMakeBodyID(rdr, a.body).Enable()
}

func recDispatchBodySetMotionLocks(a *recArgsBodySetMotionLocks, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetMotionLocks(a.locks)
}

func recDispatchBodySetBullet(a *recArgsBodySetBullet, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetBullet(a.flag)
}

func recDispatchBodyEnableContactRecycling(a *recArgsBodyEnableContactRecycling, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetContactRecyclingEnabled(a.flag)
}

func recDispatchBodyEnableContactEvents(a *recArgsBodyEnableContactEvents, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetContactEventsEnabled(a.flag)
}

func recDispatchBodyEnableHitEvents(a *recArgsBodyEnableHitEvents, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetHitEventsEnabled(a.flag)
}

func recDispatchCreateCircleShape(a *recArgsCreateCircleShape, rdr *recReader) {
	recId := recRSHAPEID(rdr)
	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreateCircleShape(&a.def, &a.circle)
	recCheckShapeID(rdr, gotId, recId)
}

func recDispatchCreateCapsuleShape(a *recArgsCreateCapsuleShape, rdr *recReader) {
	recId := recRSHAPEID(rdr)
	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreateCapsuleShape(&a.def, &a.capsule)
	recCheckShapeID(rdr, gotId, recId)
}

func recDispatchCreateSegmentShape(a *recArgsCreateSegmentShape, rdr *recReader) {
	recId := recRSHAPEID(rdr)
	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreateSegmentShape(&a.def, &a.segment)
	recCheckShapeID(rdr, gotId, recId)
}

func recDispatchCreatePolygonShape(a *recArgsCreatePolygonShape, rdr *recReader) {
	recId := recRSHAPEID(rdr)
	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreatePolygonShape(&a.def, &a.polygon)
	recCheckShapeID(rdr, gotId, recId)
}

func recDispatchCreateChainSegmentShape(a *recArgsCreateChainSegmentShape, rdr *recReader) {
	recId := recRSHAPEID(rdr)
	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreateChainSegmentShape(&a.def, &a.chainSegment)
	recCheckShapeID(rdr, gotId, recId)
}

func recDispatchDestroyShape(a *recArgsDestroyShape, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).Destroy(a.updateBodyMass)
}

func recDispatchShapeSetDensity(a *recArgsShapeSetDensity, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetDensity(a.density, a.updateBodyMass)
}

func recDispatchShapeSetFriction(a *recArgsShapeSetFriction, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetFriction(a.friction)
}

func recDispatchShapeSetRestitution(a *recArgsShapeSetRestitution, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetRestitution(a.restitution)
}

func recDispatchShapeSetUserMaterial(a *recArgsShapeSetUserMaterial, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetUserMaterial(a.material)
}

func recDispatchShapeSetSurfaceMaterial(a *recArgsShapeSetSurfaceMaterial, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSurfaceMaterial(&a.material)
}

func recDispatchShapeSetFilter(a *recArgsShapeSetFilter, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetFilter(a.filter)
}

func recDispatchShapeEnableSensorEvents(a *recArgsShapeEnableSensorEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSensorEventsEnabled(a.flag)
}

func recDispatchShapeEnableContactEvents(a *recArgsShapeEnableContactEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetContactEventsEnabled(a.flag)
}

func recDispatchShapeEnablePreSolveEvents(a *recArgsShapeEnablePreSolveEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetPreSolveEventsEnabled(a.flag)
}

func recDispatchShapeEnableHitEvents(a *recArgsShapeEnableHitEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetHitEventsEnabled(a.flag)
}

func recDispatchShapeSetCircle(a *recArgsShapeSetCircle, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetCircle(&a.circle)
}

func recDispatchShapeSetCapsule(a *recArgsShapeSetCapsule, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetCapsule(&a.capsule)
}

func recDispatchShapeSetSegment(a *recArgsShapeSetSegment, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSegment(&a.segment)
}

func recDispatchShapeSetPolygon(a *recArgsShapeSetPolygon, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetPolygon(&a.polygon)
}

func recDispatchShapeSetChainSegment(a *recArgsShapeSetChainSegment, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetChainSegment(&a.chainSegment)
}

func recDispatchShapeApplyWind(a *recArgsShapeApplyWind, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).ApplyWind(a.wind, a.drag, a.lift, a.wake)
}

func recDispatchCreateChain(a *recArgsCreateChain, rdr *recReader) {
	recId := recRCHAINID(rdr)
	if !rdr.ok {
		// A corrupt point/material count left the def degenerate, do not build a chain from it
		return
	}

	bodyId := recMakeBodyID(rdr, a.body)
	gotId := bodyId.CreateChain(&a.def)
	recCheckChainID(rdr, gotId, recId)
}

func recDispatchDestroyChain(a *recArgsDestroyChain, rdr *recReader) {
	recMakeChainID(rdr, a.chain).Destroy()
}

func recDispatchChainSetSurfaceMaterial(a *recArgsChainSetSurfaceMaterial, rdr *recReader) {
	recMakeChainID(rdr, a.chain).SetSurfaceMaterial(&a.material, int(a.segmentIndex))
}

func recDispatchChainSetAllSurfaceMaterials(a *recArgsChainSetAllSurfaceMaterials, rdr *recReader) {
	recMakeChainID(rdr, a.chain).SetAllSurfaceMaterials(&a.material)
}

func recDispatchCreateDistanceJoint(a *recArgsCreateDistanceJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateDistanceJoint(&def).Joint, recId)
}

func recDispatchCreateFilterJoint(a *recArgsCreateFilterJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateFilterJoint(&def).Joint, recId)
}

func recDispatchCreateMotorJoint(a *recArgsCreateMotorJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateMotorJoint(&def).Joint, recId)
}

func recDispatchCreateMoverJoint(a *recArgsCreateMoverJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateMoverJoint(&def).Joint, recId)
}

func recDispatchCreatePrismaticJoint(a *recArgsCreatePrismaticJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreatePrismaticJoint(&def).Joint, recId)
}

func recDispatchCreatePogoJoint(a *recArgsCreatePogoJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreatePogoJoint(&def).Joint, recId)
}

func recDispatchCreateRevoluteJoint(a *recArgsCreateRevoluteJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateRevoluteJoint(&def).Joint, recId)
}

func recDispatchCreateWeldJoint(a *recArgsCreateWeldJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateWeldJoint(&def).Joint, recId)
}

func recDispatchCreateWheelJoint(a *recArgsCreateWheelJoint, rdr *recReader) {
	recId := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateWheelJoint(&def).Joint, recId)
}

func recDispatchDestroyJoint(a *recArgsDestroyJoint, rdr *recReader) {
	recMakeJointID(rdr, a.joint).Destroy()
}

func recDispatchJointSetLocalFrameA(a *recArgsJointSetLocalFrameA, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetLocalFrameA(a.localFrame)
}

func recDispatchJointSetLocalFrameB(a *recArgsJointSetLocalFrameB, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetLocalFrameB(a.localFrame)
}

func recDispatchJointSetCollideConnected(a *recArgsJointSetCollideConnected, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetCollideConnected(a.shouldCollide)
}

func recDispatchJointWakeBodies(a *recArgsJointWakeBodies, rdr *recReader) {
	recMakeJointID(rdr, a.joint).WakeBodies()
}

func recDispatchJointSetConstraintTuning(a *recArgsJointSetConstraintTuning, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetConstraintTuning(a.hertz, a.dampingRatio)
}

func recDispatchJointSetForceThreshold(a *recArgsJointSetForceThreshold, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetForceThreshold(a.threshold)
}

func recDispatchJointSetTorqueThreshold(a *recArgsJointSetTorqueThreshold, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetTorqueThreshold(a.threshold)
}

func recDispatchDistanceJointSetLength(a *recArgsDistanceJointSetLength, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLength(a.length)
}

func recDispatchDistanceJointEnableSpring(a *recArgsDistanceJointEnableSpring, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchDistanceJointSetSpringForceRange(a *recArgsDistanceJointSetSpringForceRange, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringForceRange(a.lowerForce, a.upperForce)
}

func recDispatchDistanceJointSetSpringHertz(a *recArgsDistanceJointSetSpringHertz, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchDistanceJointSetSpringDampingRatio(a *recArgsDistanceJointSetSpringDampingRatio, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchDistanceJointEnableLimit(a *recArgsDistanceJointEnableLimit, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchDistanceJointSetLengthRange(a *recArgsDistanceJointSetLengthRange, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLengthRange(a.minLength, a.maxLength)
}

func recDispatchDistanceJointEnableMotor(a *recArgsDistanceJointEnableMotor, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchDistanceJointSetMotorSpeed(a *recArgsDistanceJointSetMotorSpeed, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchDistanceJointSetMaxMotorForce(a *recArgsDistanceJointSetMaxMotorForce, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorForce(a.force)
}

func recDispatchMotorJointSetLinearVelocity(a *recArgsMotorJointSetLinearVelocity, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearVelocity(a.velocity)
}

func recDispatchMotorJointSetAngularVelocity(a *recArgsMotorJointSetAngularVelocity, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularVelocity(a.velocity)
}

func recDispatchMotorJointSetMaxVelocityForce(a *recArgsMotorJointSetMaxVelocityForce, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityForce(a.maxForce)
}

func recDispatchMotorJointSetMaxVelocityTorque(a *recArgsMotorJointSetMaxVelocityTorque, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityTorque(a.maxTorque)
}

func recDispatchMotorJointSetLinearHertz(a *recArgsMotorJointSetLinearHertz, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearHertz(a.hertz)
}

func recDispatchMotorJointSetLinearDampingRatio(a *recArgsMotorJointSetLinearDampingRatio, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearDampingRatio(a.damping)
}

func recDispatchMotorJointSetAngularHertz(a *recArgsMotorJointSetAngularHertz, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularHertz(a.hertz)
}

func recDispatchMotorJointSetAngularDampingRatio(a *recArgsMotorJointSetAngularDampingRatio, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularDampingRatio(a.damping)
}

func recDispatchMotorJointSetMaxSpringForce(a *recArgsMotorJointSetMaxSpringForce, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxSpringForce(a.maxForce)
}

func recDispatchMotorJointSetMaxSpringTorque(a *recArgsMotorJointSetMaxSpringTorque, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxSpringTorque(a.maxTorque)
}

func recDispatchMoverJointSetLinearVelocity(a *recArgsMoverJointSetLinearVelocity, rdr *recReader) {
	MoverJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearVelocity(a.velocity)
}

func recDispatchMoverJointSetMaxVelocityForce(a *recArgsMoverJointSetMaxVelocityForce, rdr *recReader) {
	MoverJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityForce(a.maxForce)
}

func recDispatchPogoJointSetRestLength(a *recArgsPogoJointSetRestLength, rdr *recReader) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetRestLength(a.length)
}

func recDispatchPogoJointSetSpringHertz(a *recArgsPogoJointSetSpringHertz, rdr *recReader) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchPogoJointSetSpringDampingRatio(a *recArgsPogoJointSetSpringDampingRatio, rdr *recReader) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchPrismaticJointEnableSpring(a *recArgsPrismaticJointEnableSpring, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchPrismaticJointSetSpringHertz(a *recArgsPrismaticJointSetSpringHertz, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchPrismaticJointSetSpringDampingRatio(a *recArgsPrismaticJointSetSpringDampingRatio, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchPrismaticJointSetTargetTranslation(a *recArgsPrismaticJointSetTargetTranslation, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetTargetTranslation(a.translation)
}

func recDispatchPrismaticJointEnableLimit(a *recArgsPrismaticJointEnableLimit, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchPrismaticJointSetLimits(a *recArgsPrismaticJointSetLimits, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchPrismaticJointEnableMotor(a *recArgsPrismaticJointEnableMotor, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchPrismaticJointSetMotorSpeed(a *recArgsPrismaticJointSetMotorSpeed, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchPrismaticJointSetMaxMotorForce(a *recArgsPrismaticJointSetMaxMotorForce, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorForce(a.force)
}

func recDispatchRevoluteJointEnableSpring(a *recArgsRevoluteJointEnableSpring, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchRevoluteJointSetSpringHertz(a *recArgsRevoluteJointSetSpringHertz, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchRevoluteJointSetSpringDampingRatio(a *recArgsRevoluteJointSetSpringDampingRatio, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchRevoluteJointSetTargetAngle(a *recArgsRevoluteJointSetTargetAngle, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetTargetAngle(a.angle)
}

func recDispatchRevoluteJointEnableLimit(a *recArgsRevoluteJointEnableLimit, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchRevoluteJointSetLimits(a *recArgsRevoluteJointSetLimits, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchRevoluteJointEnableMotor(a *recArgsRevoluteJointEnableMotor, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchRevoluteJointSetMotorSpeed(a *recArgsRevoluteJointSetMotorSpeed, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchRevoluteJointSetMaxMotorTorque(a *recArgsRevoluteJointSetMaxMotorTorque, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorTorque(a.torque)
}

func recDispatchWeldJointSetLinearHertz(a *recArgsWeldJointSetLinearHertz, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearHertz(a.hertz)
}

func recDispatchWeldJointSetLinearDampingRatio(a *recArgsWeldJointSetLinearDampingRatio, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearDampingRatio(a.dampingRatio)
}

func recDispatchWeldJointSetAngularHertz(a *recArgsWeldJointSetAngularHertz, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularHertz(a.hertz)
}

func recDispatchWeldJointSetAngularDampingRatio(a *recArgsWeldJointSetAngularDampingRatio, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularDampingRatio(a.dampingRatio)
}

func recDispatchWheelJointEnableSpring(a *recArgsWheelJointEnableSpring, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchWheelJointSetSpringHertz(a *recArgsWheelJointSetSpringHertz, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchWheelJointSetSpringDampingRatio(a *recArgsWheelJointSetSpringDampingRatio, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchWheelJointEnableLimit(a *recArgsWheelJointEnableLimit, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchWheelJointSetLimits(a *recArgsWheelJointSetLimits, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchWheelJointEnableMotor(a *recArgsWheelJointEnableMotor, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchWheelJointSetMotorSpeed(a *recArgsWheelJointSetMotorSpeed, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchWheelJointSetMaxMotorTorque(a *recArgsWheelJointSetMaxMotorTorque, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorTorque(a.torque)
}

func recF32Differs(a float32, b float32) bool {
	return math.Float32bits(a) != math.Float32bits(b)
}

func recVec2Differs(a Vec2, b Vec2) bool {
	return recF32Differs(a.X, b.X) || recF32Differs(a.Y, b.Y)
}

func recGrowFrameQueries(player *Replay) {
	recGrow(&player.frameQueries, &player.frameQueryCap, player.frameQueryCount+1, player.frameQueryCount, int(unsafe.Sizeof(recDrawQuery{})))
}

func recGrowFrameHits(player *Replay, need int) {
	recGrow(&player.frameHits, &player.frameHitCap, player.frameHitCount+need, player.frameHitCount, int(unsafe.Sizeof(recRecordedHit{})))
}

func recStashQueryBegin(player *Replay, kind int, hits []recRecordedHit, hitCount int) *recDrawQuery {
	recGrowFrameQueries(player)
	q := &player.frameQueries[player.frameQueryCount]
	*q = recDrawQuery{kind: kind, hitStart: player.frameHitCount, hitCount: hitCount}
	recGrowFrameHits(player, hitCount)
	copy(player.frameHits[player.frameHitCount:player.frameHitCount+hitCount], hits[:hitCount])
	player.frameHitCount += hitCount
	player.frameQueryCount++

	return q
}

func recReplayOverlapTrampoline(id Shape, ctx *recReplayQueryCtx) bool {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return false
	}

	h := &rc.hits[rc.cursor]

	rc.cursor++
	if id.index1 != h.id.index1 || id.generation != h.id.generation {
		rc.rdr.diverged = true
	}

	return h.userReturnB
}

func recDispatchQueryOverlapAABB(a *recArgsQueryOverlapAABB, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	_ = recRTREESTATS(rdr)
	if !rdr.ok {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.OverlapAABB(a.origin, a.aabb, a.filter, func(id Shape) bool { return recReplayOverlapTrampoline(id, &rc) })

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqOverlapAABB), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.aabb = a.aabb
	}
}

func recDispatchQueryOverlapShape(a *recArgsQueryOverlapShape, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	_ = recRTREESTATS(rdr)
	if !rdr.ok {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.OverlapShape(a.origin, &a.proxy, a.filter, func(id Shape) bool { return recReplayOverlapTrampoline(id, &rc) })

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqOverlapShape), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.proxy = a.proxy
	}
}

func recReplayCastTrampoline(id Shape, point Pos, normal Vec2, fraction float32, ctx *recReplayQueryCtx) float32 {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return 0
	}

	h := &rc.hits[rc.cursor]
	rc.cursor++
	// Compare positions through the full width delta, truncating both sides would pass vacuously
	// far from the origin
	if id.index1 != h.id.index1 || id.generation != h.id.generation ||
		recVec2Differs(point.Sub(h.point), Vec2{}) || recVec2Differs(normal, h.normal) || recF32Differs(fraction, h.fraction) {
		rc.rdr.diverged = true
	}

	return h.userReturnF
}

func recDispatchQueryCastRay(a *recArgsQueryCastRay, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].point = recRPOSITION(rdr)
		rdr.hits[i].normal = recRVEC2(rdr)
		rdr.hits[i].fraction = recRF32(rdr)
		rdr.hits[i].userReturnF = recRF32(rdr)
	}

	_ = recRTREESTATS(rdr)
	if !rdr.ok {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CastRay(a.origin, a.translation, a.filter, func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
		return recReplayCastTrampoline(id, point, normal, fraction, &rc)
	})

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastRay), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.translation = a.translation
	}
}

func recDispatchQueryCastShape(a *recArgsQueryCastShape, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].point = recRPOSITION(rdr)
		rdr.hits[i].normal = recRVEC2(rdr)
		rdr.hits[i].fraction = recRF32(rdr)
		rdr.hits[i].userReturnF = recRF32(rdr)
	}

	_ = recRTREESTATS(rdr)
	if !rdr.ok {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CastShape(a.origin, &a.proxy, a.translation, a.filter, func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
		return recReplayCastTrampoline(id, point, normal, fraction, &rc)
	})

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastShape), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.proxy = a.proxy
		q.translation = a.translation
	}
}

func recReplayPlaneTrampoline(id Shape, plane *PlaneResult, ctx *recReplayQueryCtx) bool {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return true
	}

	h := &rc.hits[rc.cursor]

	rc.cursor++
	if id.index1 != h.id.index1 || id.generation != h.id.generation ||
		recVec2Differs(plane.Plane.Normal, h.plane.Plane.Normal) || recF32Differs(plane.Plane.Offset, h.plane.Plane.Offset) ||
		recVec2Differs(plane.Point, h.plane.Point) || plane.Hit != h.plane.Hit {
		rc.rdr.diverged = true
	}

	return h.userReturnB
}

func recDispatchQueryCollideMover(a *recArgsQueryCollideMover, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].plane = recRPLANERESULT(rdr)
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	if !rdr.ok {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CollideMover(a.origin, &a.mover, a.filter, func(id Shape, plane *PlaneResult) bool { return recReplayPlaneTrampoline(id, plane, &rc) })

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCollideMover), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.mover = a.mover
	}
}

func recDispatchQueryCastRayClosest(a *recArgsQueryCastRayClosest, rdr *recReader) {
	rec := recRRAYRESULT(rdr)
	if !rdr.ok {
		return
	}

	got := rdr.replayWorldID.CastRayClosest(a.origin, a.translation, a.filter)
	if got.Hit != rec.Hit || (got.Hit && (got.Shape.index1 != rec.Shape.index1 || got.Shape.generation != rec.Shape.generation ||
		recVec2Differs(got.Point.Sub(rec.Point), Vec2{}) || recVec2Differs(got.Normal, rec.Normal) || recF32Differs(got.Fraction, rec.Fraction))) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		h := recRecordedHit{id: recMakeShapeID(rdr, rec.Shape), point: rec.Point, normal: rec.Normal, fraction: rec.Fraction}

		n := 0
		if rec.Hit {
			n = 1
		}

		q := recStashQueryBegin(rdr.owner, int(recqCastRayClosest), []recRecordedHit{h}, n)
		q.filter = a.filter
		q.origin = a.origin
		q.translation = a.translation
	}
}

func recDispatchQueryCastMover(a *recArgsQueryCastMover, rdr *recReader) {
	rec := recRF32(rdr)
	if !rdr.ok {
		return
	}

	got := rdr.replayWorldID.CastMover(a.origin, &a.mover, a.translation, a.filter)
	if recF32Differs(got, rec) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastMover), nil, 0)
		q.filter = a.filter
		q.origin = a.origin
		q.mover = a.mover
		q.translation = a.translation
		q.castFraction = rec
	}
}

func recDispatchShapeTestPoint(a *recArgsShapeTestPoint, rdr *recReader) {
	rec := recRBOOL(rdr)
	if !rdr.ok {
		return
	}

	id := recMakeShapeID(rdr, a.shape)

	got := id.TestPoint(a.point)
	if got != rec {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqShapeTestPoint), nil, 0)
		q.shape = id
		q.origin = a.point
		q.boolResult = rec
	}
}

func recDispatchShapeRayCast(a *recArgsShapeRayCast, rdr *recReader) {
	rec := recRWORLDCASTOUTPUT(rdr)
	if !rdr.ok {
		return
	}

	id := recMakeShapeID(rdr, a.shape)

	got := id.RayCast(a.origin, a.translation)
	if got.Hit != rec.Hit || (got.Hit && (recVec2Differs(got.Normal, rec.Normal) ||
		recVec2Differs(got.Point.Sub(rec.Point), Vec2{}) || recF32Differs(got.Fraction, rec.Fraction))) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqShapeRayCast), nil, 0)
		q.shape = id
		q.origin = a.origin
		q.translation = a.translation
		q.castOut = rec
	}
}

func recDispatchStateHash(a *recArgsStateHash, rdr *recReader) {
	world := rdr.replayWorldID

	computed := hashWorldState(world)
	if computed != a.hash {
		world.logger.Warn("box2d: replay state hash mismatch", "recorded", a.hash, "computed", computed)

		rdr.diverged = true
	}
}

func recDispatchRecordingBounds(a *recArgsRecordingBounds, rdr *recReader) {
	rdr.owner.bounds = a.bounds
}

func recDispatchOne(player *Replay) int {
	rdr := &player.rdr
	if rdr.cursor >= rdr.size || !rdr.ok {
		return -1
	}

	opcode := recRU8(rdr)
	if !rdr.ok {
		return -1
	}

	payloadSize := recRU24(rdr)
	if !rdr.ok {
		return -1
	}

	payloadStart := rdr.cursor

	switch opcode {
	case 0x01:
		var a recArgsDestroyWorld
		if rdr.ok {
			recDispatchDestroyWorld(&a, rdr)
		}
	case 0x80:
		var a recArgsStep

		a.dt = recRF32(rdr)

		a.subStepCount = recRI32(rdr)
		if rdr.ok {
			recDispatchStep(&a, rdr)
		}
	case 0x02:
		var a recArgsWorldEnableSleeping

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWorldEnableSleeping(&a, rdr)
		}
	case 0x03:
		var a recArgsWorldEnableContinuous

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWorldEnableContinuous(&a, rdr)
		}
	case 0x04:
		var a recArgsWorldSetRestitutionThreshold

		a.value = recRF32(rdr)
		if rdr.ok {
			recDispatchWorldSetRestitutionThreshold(&a, rdr)
		}
	case 0x05:
		var a recArgsWorldSetHitEventThreshold

		a.value = recRF32(rdr)
		if rdr.ok {
			recDispatchWorldSetHitEventThreshold(&a, rdr)
		}
	case 0x06:
		var a recArgsWorldSetGravity

		a.gravity = recRVEC2(rdr)
		if rdr.ok {
			recDispatchWorldSetGravity(&a, rdr)
		}
	case 0x07:
		var a recArgsWorldExplode

		a.def = recREXPLOSIONDEF(rdr)
		if rdr.ok {
			recDispatchWorldExplode(&a, rdr)
		}
	case 0x08:
		var a recArgsWorldSetContactTuning

		a.hertz = recRF32(rdr)
		a.dampingRatio = recRF32(rdr)

		a.pushSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchWorldSetContactTuning(&a, rdr)
		}
	case 0x09:
		var a recArgsWorldSetContactRecycleDistance

		a.recycleDistance = recRF32(rdr)
		if rdr.ok {
			recDispatchWorldSetContactRecycleDistance(&a, rdr)
		}
	case 0x0A:
		var a recArgsWorldSetMaximumLinearSpeed

		a.maximumLinearSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchWorldSetMaximumLinearSpeed(&a, rdr)
		}
	case 0x0B:
		var a recArgsWorldEnableWarmStarting

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWorldEnableWarmStarting(&a, rdr)
		}
	case 0x0C:
		var a recArgsWorldRebuildStaticTree
		if rdr.ok {
			recDispatchWorldRebuildStaticTree(&a, rdr)
		}
	case 0x0D:
		var a recArgsWorldSetRestitutionIterations

		a.iterations = recRI32(rdr)
		if rdr.ok {
			recDispatchWorldSetRestitutionIterations(&a, rdr)
		}
	case 0x0E:
		var a recArgsWorldEnableRestitutionPropagation

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWorldEnableRestitutionPropagation(&a, rdr)
		}
	case 0x10:
		var a recArgsCreateBody

		a.def = recRBODYDEF(rdr)
		if rdr.ok {
			recDispatchCreateBody(&a, rdr)
		}
	case 0x11:
		var a recArgsDestroyBody

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchDestroyBody(&a, rdr)
		}
	case 0x20:
		var a recArgsBodySetTransform

		a.body = recRBODYID(rdr)
		a.position = recRPOSITION(rdr)

		a.rotation = recRROT(rdr)
		if rdr.ok {
			recDispatchBodySetTransform(&a, rdr)
		}
	case 0x21:
		var a recArgsBodySetLinearVelocity

		a.body = recRBODYID(rdr)

		a.v = recRVEC2(rdr)
		if rdr.ok {
			recDispatchBodySetLinearVelocity(&a, rdr)
		}
	case 0x22:
		var a recArgsBodySetType

		a.body = recRBODYID(rdr)

		a.bodyType = recRI32(rdr)
		if rdr.ok {
			recDispatchBodySetType(&a, rdr)
		}
	case 0x23:
		var a recArgsBodySetName

		a.body = recRBODYID(rdr)

		a.name = recRSTR(rdr)
		if rdr.ok {
			recDispatchBodySetName(&a, rdr)
		}
	case 0x24:
		var a recArgsBodySetAngularVelocity

		a.body = recRBODYID(rdr)

		a.w = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetAngularVelocity(&a, rdr)
		}
	case 0x25:
		var a recArgsBodySetTargetTransform

		a.body = recRBODYID(rdr)
		a.target = recRWORLDXF(rdr)
		a.timeStep = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodySetTargetTransform(&a, rdr)
		}
	case 0x26:
		var a recArgsBodyApplyForce

		a.body = recRBODYID(rdr)
		a.force = recRVEC2(rdr)
		a.point = recRPOSITION(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyForce(&a, rdr)
		}
	case 0x27:
		var a recArgsBodyApplyForceToCenter

		a.body = recRBODYID(rdr)
		a.force = recRVEC2(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyForceToCenter(&a, rdr)
		}
	case 0x28:
		var a recArgsBodyApplyTorque

		a.body = recRBODYID(rdr)
		a.torque = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyTorque(&a, rdr)
		}
	case 0x29:
		var a recArgsBodyClearForces

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchBodyClearForces(&a, rdr)
		}
	case 0x2A:
		var a recArgsBodyApplyLinearImpulse

		a.body = recRBODYID(rdr)
		a.impulse = recRVEC2(rdr)
		a.point = recRPOSITION(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyLinearImpulse(&a, rdr)
		}
	case 0x2B:
		var a recArgsBodyApplyLinearImpulseToCenter

		a.body = recRBODYID(rdr)
		a.impulse = recRVEC2(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyLinearImpulseToCenter(&a, rdr)
		}
	case 0x2C:
		var a recArgsBodyApplyAngularImpulse

		a.body = recRBODYID(rdr)
		a.impulse = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyApplyAngularImpulse(&a, rdr)
		}
	case 0x2D:
		var a recArgsBodySetMassData

		a.body = recRBODYID(rdr)

		a.massData = recRMASSDATA(rdr)
		if rdr.ok {
			recDispatchBodySetMassData(&a, rdr)
		}
	case 0x2E:
		var a recArgsBodyUpdateMassFromShapes

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchBodyUpdateMassFromShapes(&a, rdr)
		}
	case 0x2F:
		var a recArgsBodySetLinearDamping

		a.body = recRBODYID(rdr)

		a.damping = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetLinearDamping(&a, rdr)
		}
	case 0x30:
		var a recArgsBodySetAngularDamping

		a.body = recRBODYID(rdr)

		a.damping = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetAngularDamping(&a, rdr)
		}
	case 0x31:
		var a recArgsBodySetGravityScale

		a.body = recRBODYID(rdr)

		a.scale = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetGravityScale(&a, rdr)
		}
	case 0x32:
		var a recArgsBodySetAwake

		a.body = recRBODYID(rdr)

		a.awake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodySetAwake(&a, rdr)
		}
	case 0x33:
		var a recArgsBodyWakeTouching

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchBodyWakeTouching(&a, rdr)
		}
	case 0x34:
		var a recArgsBodyEnableSleep

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyEnableSleep(&a, rdr)
		}
	case 0x35:
		var a recArgsBodySetSleepThreshold

		a.body = recRBODYID(rdr)

		a.threshold = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetSleepThreshold(&a, rdr)
		}
	case 0x36:
		var a recArgsBodyDisable

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchBodyDisable(&a, rdr)
		}
	case 0x37:
		var a recArgsBodyEnable

		a.body = recRBODYID(rdr)
		if rdr.ok {
			recDispatchBodyEnable(&a, rdr)
		}
	case 0x38:
		var a recArgsBodySetMotionLocks

		a.body = recRBODYID(rdr)

		a.locks = recRLOCKS(rdr)
		if rdr.ok {
			recDispatchBodySetMotionLocks(&a, rdr)
		}
	case 0x39:
		var a recArgsBodySetBullet

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodySetBullet(&a, rdr)
		}
	case 0x3A:
		var a recArgsBodyEnableContactRecycling

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyEnableContactRecycling(&a, rdr)
		}
	case 0x3B:
		var a recArgsBodyEnableContactEvents

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyEnableContactEvents(&a, rdr)
		}
	case 0x3C:
		var a recArgsBodyEnableHitEvents

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchBodyEnableHitEvents(&a, rdr)
		}
	case 0x3D:
		var a recArgsBodySetSafetyFactor

		a.body = recRBODYID(rdr)

		a.value = recRF32(rdr)
		if rdr.ok {
			recDispatchBodySetSafetyFactor(&a, rdr)
		}
	case 0x40:
		var a recArgsCreateCircleShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.circle = recRCIRCLE(rdr)
		if rdr.ok {
			recDispatchCreateCircleShape(&a, rdr)
		}
	case 0x41:
		var a recArgsCreateCapsuleShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.capsule = recRCAPSULE(rdr)
		if rdr.ok {
			recDispatchCreateCapsuleShape(&a, rdr)
		}
	case 0x42:
		var a recArgsCreateSegmentShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.segment = recRSEGMENT(rdr)
		if rdr.ok {
			recDispatchCreateSegmentShape(&a, rdr)
		}
	case 0x43:
		var a recArgsCreatePolygonShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.polygon = recRPOLYGON(rdr)
		if rdr.ok {
			recDispatchCreatePolygonShape(&a, rdr)
		}
	case 0x44:
		var a recArgsCreateChainSegmentShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.chainSegment = recRCHAINSEG(rdr)
		if rdr.ok {
			recDispatchCreateChainSegmentShape(&a, rdr)
		}
	case 0x45:
		var a recArgsDestroyShape

		a.shape = recRSHAPEID(rdr)

		a.updateBodyMass = recRBOOL(rdr)
		if rdr.ok {
			recDispatchDestroyShape(&a, rdr)
		}
	case 0x50:
		var a recArgsShapeSetDensity

		a.shape = recRSHAPEID(rdr)
		a.density = recRF32(rdr)

		a.updateBodyMass = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeSetDensity(&a, rdr)
		}
	case 0x51:
		var a recArgsShapeSetFriction

		a.shape = recRSHAPEID(rdr)

		a.friction = recRF32(rdr)
		if rdr.ok {
			recDispatchShapeSetFriction(&a, rdr)
		}
	case 0x52:
		var a recArgsShapeSetRestitution

		a.shape = recRSHAPEID(rdr)

		a.restitution = recRF32(rdr)
		if rdr.ok {
			recDispatchShapeSetRestitution(&a, rdr)
		}
	case 0x53:
		var a recArgsShapeSetUserMaterial

		a.shape = recRSHAPEID(rdr)

		a.material = recRU64(rdr)
		if rdr.ok {
			recDispatchShapeSetUserMaterial(&a, rdr)
		}
	case 0x54:
		var a recArgsShapeSetSurfaceMaterial

		a.shape = recRSHAPEID(rdr)

		a.material = recRMATERIAL(rdr)
		if rdr.ok {
			recDispatchShapeSetSurfaceMaterial(&a, rdr)
		}
	case 0x55:
		var a recArgsShapeSetFilter

		a.shape = recRSHAPEID(rdr)

		a.filter = recRFILTER(rdr)
		if rdr.ok {
			recDispatchShapeSetFilter(&a, rdr)
		}
	case 0x56:
		var a recArgsShapeEnableSensorEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeEnableSensorEvents(&a, rdr)
		}
	case 0x57:
		var a recArgsShapeEnableContactEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeEnableContactEvents(&a, rdr)
		}
	case 0x58:
		var a recArgsShapeEnablePreSolveEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeEnablePreSolveEvents(&a, rdr)
		}
	case 0x59:
		var a recArgsShapeEnableHitEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeEnableHitEvents(&a, rdr)
		}
	case 0x5A:
		var a recArgsShapeSetCircle

		a.shape = recRSHAPEID(rdr)

		a.circle = recRCIRCLE(rdr)
		if rdr.ok {
			recDispatchShapeSetCircle(&a, rdr)
		}
	case 0x5B:
		var a recArgsShapeSetCapsule

		a.shape = recRSHAPEID(rdr)

		a.capsule = recRCAPSULE(rdr)
		if rdr.ok {
			recDispatchShapeSetCapsule(&a, rdr)
		}
	case 0x5C:
		var a recArgsShapeSetSegment

		a.shape = recRSHAPEID(rdr)

		a.segment = recRSEGMENT(rdr)
		if rdr.ok {
			recDispatchShapeSetSegment(&a, rdr)
		}
	case 0x5D:
		var a recArgsShapeSetPolygon

		a.shape = recRSHAPEID(rdr)

		a.polygon = recRPOLYGON(rdr)
		if rdr.ok {
			recDispatchShapeSetPolygon(&a, rdr)
		}
	case 0x5E:
		var a recArgsShapeSetChainSegment

		a.shape = recRSHAPEID(rdr)

		a.chainSegment = recRCHAINSEG(rdr)
		if rdr.ok {
			recDispatchShapeSetChainSegment(&a, rdr)
		}
	case 0x5F:
		var a recArgsShapeApplyWind

		a.shape = recRSHAPEID(rdr)
		a.wind = recRVEC2(rdr)
		a.drag = recRF32(rdr)
		a.lift = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if rdr.ok {
			recDispatchShapeApplyWind(&a, rdr)
		}
	case 0x70:
		var a recArgsCreateChain

		a.body = recRBODYID(rdr)

		a.def = recRCHAINDEF(rdr)
		if rdr.ok {
			recDispatchCreateChain(&a, rdr)
		}
	case 0x71:
		var a recArgsDestroyChain

		a.chain = recRCHAINID(rdr)
		if rdr.ok {
			recDispatchDestroyChain(&a, rdr)
		}
	case 0x72:
		var a recArgsChainSetSurfaceMaterial

		a.chain = recRCHAINID(rdr)
		a.material = recRMATERIAL(rdr)

		a.segmentIndex = recRI32(rdr)
		if rdr.ok {
			recDispatchChainSetSurfaceMaterial(&a, rdr)
		}
	case 0x73:
		var a recArgsChainSetAllSurfaceMaterials

		a.chain = recRCHAINID(rdr)

		a.material = recRMATERIAL(rdr)
		if rdr.ok {
			recDispatchChainSetAllSurfaceMaterials(&a, rdr)
		}
	case 0x90:
		var a recArgsDestroyJoint

		a.joint = recRJOINTID(rdr)
		if rdr.ok {
			recDispatchDestroyJoint(&a, rdr)
		}
	case 0x91:
		var a recArgsJointSetLocalFrameA

		a.joint = recRJOINTID(rdr)

		a.localFrame = recRXF(rdr)
		if rdr.ok {
			recDispatchJointSetLocalFrameA(&a, rdr)
		}
	case 0x92:
		var a recArgsJointSetLocalFrameB

		a.joint = recRJOINTID(rdr)

		a.localFrame = recRXF(rdr)
		if rdr.ok {
			recDispatchJointSetLocalFrameB(&a, rdr)
		}
	case 0x93:
		var a recArgsJointSetCollideConnected

		a.joint = recRJOINTID(rdr)

		a.shouldCollide = recRBOOL(rdr)
		if rdr.ok {
			recDispatchJointSetCollideConnected(&a, rdr)
		}
	case 0x94:
		var a recArgsJointWakeBodies

		a.joint = recRJOINTID(rdr)
		if rdr.ok {
			recDispatchJointWakeBodies(&a, rdr)
		}
	case 0x95:
		var a recArgsJointSetConstraintTuning

		a.joint = recRJOINTID(rdr)
		a.hertz = recRF32(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchJointSetConstraintTuning(&a, rdr)
		}
	case 0x96:
		var a recArgsJointSetForceThreshold

		a.joint = recRJOINTID(rdr)

		a.threshold = recRF32(rdr)
		if rdr.ok {
			recDispatchJointSetForceThreshold(&a, rdr)
		}
	case 0x97:
		var a recArgsJointSetTorqueThreshold

		a.joint = recRJOINTID(rdr)

		a.threshold = recRF32(rdr)
		if rdr.ok {
			recDispatchJointSetTorqueThreshold(&a, rdr)
		}
	case 0x98:
		var a recArgsCreateDistanceJoint

		a.def = recRDISTANCEJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateDistanceJoint(&a, rdr)
		}
	case 0x99:
		var a recArgsDistanceJointSetLength

		a.joint = recRJOINTID(rdr)

		a.length = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetLength(&a, rdr)
		}
	case 0x9A:
		var a recArgsDistanceJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if rdr.ok {
			recDispatchDistanceJointEnableSpring(&a, rdr)
		}
	case 0x9B:
		var a recArgsDistanceJointSetSpringForceRange

		a.joint = recRJOINTID(rdr)
		a.lowerForce = recRF32(rdr)

		a.upperForce = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetSpringForceRange(&a, rdr)
		}
	case 0x9C:
		var a recArgsDistanceJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetSpringHertz(&a, rdr)
		}
	case 0x9D:
		var a recArgsDistanceJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetSpringDampingRatio(&a, rdr)
		}
	case 0x9E:
		var a recArgsDistanceJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if rdr.ok {
			recDispatchDistanceJointEnableLimit(&a, rdr)
		}
	case 0x9F:
		var a recArgsDistanceJointSetLengthRange

		a.joint = recRJOINTID(rdr)
		a.minLength = recRF32(rdr)

		a.maxLength = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetLengthRange(&a, rdr)
		}
	case 0xA0:
		var a recArgsDistanceJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if rdr.ok {
			recDispatchDistanceJointEnableMotor(&a, rdr)
		}
	case 0xA1:
		var a recArgsDistanceJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetMotorSpeed(&a, rdr)
		}
	case 0xA2:
		var a recArgsDistanceJointSetMaxMotorForce

		a.joint = recRJOINTID(rdr)

		a.force = recRF32(rdr)
		if rdr.ok {
			recDispatchDistanceJointSetMaxMotorForce(&a, rdr)
		}
	case 0xA3:
		var a recArgsCreateFilterJoint

		a.def = recRFILTERJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateFilterJoint(&a, rdr)
		}
	case 0xA4:
		var a recArgsCreateMotorJoint

		a.def = recRMOTORJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateMotorJoint(&a, rdr)
		}
	case 0xA5:
		var a recArgsMotorJointSetLinearVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRVEC2(rdr)
		if rdr.ok {
			recDispatchMotorJointSetLinearVelocity(&a, rdr)
		}
	case 0xA6:
		var a recArgsMotorJointSetAngularVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetAngularVelocity(&a, rdr)
		}
	case 0xA7:
		var a recArgsMotorJointSetMaxVelocityForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetMaxVelocityForce(&a, rdr)
		}
	case 0xA8:
		var a recArgsMotorJointSetMaxVelocityTorque

		a.joint = recRJOINTID(rdr)

		a.maxTorque = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetMaxVelocityTorque(&a, rdr)
		}
	case 0xA9:
		var a recArgsMotorJointSetLinearHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetLinearHertz(&a, rdr)
		}
	case 0xAA:
		var a recArgsMotorJointSetLinearDampingRatio

		a.joint = recRJOINTID(rdr)

		a.damping = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetLinearDampingRatio(&a, rdr)
		}
	case 0xAB:
		var a recArgsMotorJointSetAngularHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetAngularHertz(&a, rdr)
		}
	case 0xAC:
		var a recArgsMotorJointSetAngularDampingRatio

		a.joint = recRJOINTID(rdr)

		a.damping = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetAngularDampingRatio(&a, rdr)
		}
	case 0xAD:
		var a recArgsMotorJointSetMaxSpringForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetMaxSpringForce(&a, rdr)
		}
	case 0xAE:
		var a recArgsMotorJointSetMaxSpringTorque

		a.joint = recRJOINTID(rdr)

		a.maxTorque = recRF32(rdr)
		if rdr.ok {
			recDispatchMotorJointSetMaxSpringTorque(&a, rdr)
		}
	case 0xAF:
		var a recArgsCreateMoverJoint

		a.def = recRMOVERJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateMoverJoint(&a, rdr)
		}
	case 0xB0:
		var a recArgsMoverJointSetLinearVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRVEC2(rdr)
		if rdr.ok {
			recDispatchMoverJointSetLinearVelocity(&a, rdr)
		}
	case 0xB1:
		var a recArgsMoverJointSetMaxVelocityForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRVEC2(rdr)
		if rdr.ok {
			recDispatchMoverJointSetMaxVelocityForce(&a, rdr)
		}
	case 0xB2:
		var a recArgsCreatePogoJoint

		a.def = recRPOGOJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreatePogoJoint(&a, rdr)
		}
	case 0xB3:
		var a recArgsPogoJointSetRestLength

		a.joint = recRJOINTID(rdr)

		a.length = recRF32(rdr)
		if rdr.ok {
			recDispatchPogoJointSetRestLength(&a, rdr)
		}
	case 0xB4:
		var a recArgsPogoJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchPogoJointSetSpringHertz(&a, rdr)
		}
	case 0xB5:
		var a recArgsPogoJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchPogoJointSetSpringDampingRatio(&a, rdr)
		}
	case 0xB6:
		var a recArgsCreatePrismaticJoint

		a.def = recRPRISMATICJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreatePrismaticJoint(&a, rdr)
		}
	case 0xB7:
		var a recArgsPrismaticJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if rdr.ok {
			recDispatchPrismaticJointEnableSpring(&a, rdr)
		}
	case 0xB8:
		var a recArgsPrismaticJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetSpringHertz(&a, rdr)
		}
	case 0xB9:
		var a recArgsPrismaticJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetSpringDampingRatio(&a, rdr)
		}
	case 0xBA:
		var a recArgsPrismaticJointSetTargetTranslation

		a.joint = recRJOINTID(rdr)

		a.translation = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetTargetTranslation(&a, rdr)
		}
	case 0xBB:
		var a recArgsPrismaticJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if rdr.ok {
			recDispatchPrismaticJointEnableLimit(&a, rdr)
		}
	case 0xBC:
		var a recArgsPrismaticJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetLimits(&a, rdr)
		}
	case 0xBD:
		var a recArgsPrismaticJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if rdr.ok {
			recDispatchPrismaticJointEnableMotor(&a, rdr)
		}
	case 0xBE:
		var a recArgsPrismaticJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetMotorSpeed(&a, rdr)
		}
	case 0xBF:
		var a recArgsPrismaticJointSetMaxMotorForce

		a.joint = recRJOINTID(rdr)

		a.force = recRF32(rdr)
		if rdr.ok {
			recDispatchPrismaticJointSetMaxMotorForce(&a, rdr)
		}
	case 0xC0:
		var a recArgsCreateRevoluteJoint

		a.def = recRREVOLUTEJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateRevoluteJoint(&a, rdr)
		}
	case 0xC1:
		var a recArgsRevoluteJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if rdr.ok {
			recDispatchRevoluteJointEnableSpring(&a, rdr)
		}
	case 0xC2:
		var a recArgsRevoluteJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetSpringHertz(&a, rdr)
		}
	case 0xC3:
		var a recArgsRevoluteJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetSpringDampingRatio(&a, rdr)
		}
	case 0xC4:
		var a recArgsRevoluteJointSetTargetAngle

		a.joint = recRJOINTID(rdr)

		a.angle = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetTargetAngle(&a, rdr)
		}
	case 0xC5:
		var a recArgsRevoluteJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if rdr.ok {
			recDispatchRevoluteJointEnableLimit(&a, rdr)
		}
	case 0xC6:
		var a recArgsRevoluteJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetLimits(&a, rdr)
		}
	case 0xC7:
		var a recArgsRevoluteJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if rdr.ok {
			recDispatchRevoluteJointEnableMotor(&a, rdr)
		}
	case 0xC8:
		var a recArgsRevoluteJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetMotorSpeed(&a, rdr)
		}
	case 0xC9:
		var a recArgsRevoluteJointSetMaxMotorTorque

		a.joint = recRJOINTID(rdr)

		a.torque = recRF32(rdr)
		if rdr.ok {
			recDispatchRevoluteJointSetMaxMotorTorque(&a, rdr)
		}
	case 0xCA:
		var a recArgsCreateWeldJoint

		a.def = recRWELDJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateWeldJoint(&a, rdr)
		}
	case 0xCB:
		var a recArgsWeldJointSetLinearHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchWeldJointSetLinearHertz(&a, rdr)
		}
	case 0xCC:
		var a recArgsWeldJointSetLinearDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchWeldJointSetLinearDampingRatio(&a, rdr)
		}
	case 0xCD:
		var a recArgsWeldJointSetAngularHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchWeldJointSetAngularHertz(&a, rdr)
		}
	case 0xCE:
		var a recArgsWeldJointSetAngularDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchWeldJointSetAngularDampingRatio(&a, rdr)
		}
	case 0xCF:
		var a recArgsCreateWheelJoint

		a.def = recRWHEELJOINTDEF(rdr)
		if rdr.ok {
			recDispatchCreateWheelJoint(&a, rdr)
		}
	case 0xD0:
		var a recArgsWheelJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWheelJointEnableSpring(&a, rdr)
		}
	case 0xD1:
		var a recArgsWheelJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if rdr.ok {
			recDispatchWheelJointSetSpringHertz(&a, rdr)
		}
	case 0xD2:
		var a recArgsWheelJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if rdr.ok {
			recDispatchWheelJointSetSpringDampingRatio(&a, rdr)
		}
	case 0xD3:
		var a recArgsWheelJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWheelJointEnableLimit(&a, rdr)
		}
	case 0xD4:
		var a recArgsWheelJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if rdr.ok {
			recDispatchWheelJointSetLimits(&a, rdr)
		}
	case 0xD5:
		var a recArgsWheelJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if rdr.ok {
			recDispatchWheelJointEnableMotor(&a, rdr)
		}
	case 0xD6:
		var a recArgsWheelJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if rdr.ok {
			recDispatchWheelJointSetMotorSpeed(&a, rdr)
		}
	case 0xD7:
		var a recArgsWheelJointSetMaxMotorTorque

		a.joint = recRJOINTID(rdr)

		a.torque = recRF32(rdr)
		if rdr.ok {
			recDispatchWheelJointSetMaxMotorTorque(&a, rdr)
		}
	case 0xE0:
		var a recArgsQueryOverlapAABB

		a.origin = recRPOSITION(rdr)
		a.aabb = recRAABB(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryOverlapAABB(&a, rdr)
		}
	case 0xE1:
		var a recArgsQueryOverlapShape

		a.origin = recRPOSITION(rdr)
		a.proxy = recRSHAPEPROXY(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryOverlapShape(&a, rdr)
		}
	case 0xE2:
		var a recArgsQueryCastRay

		a.origin = recRPOSITION(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastRay(&a, rdr)
		}
	case 0xE3:
		var a recArgsQueryCastShape

		a.origin = recRPOSITION(rdr)
		a.proxy = recRSHAPEPROXY(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastShape(&a, rdr)
		}
	case 0xE4:
		var a recArgsQueryCollideMover

		a.origin = recRPOSITION(rdr)
		a.mover = recRCAPSULE(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCollideMover(&a, rdr)
		}
	case 0xE5:
		var a recArgsQueryCastRayClosest

		a.origin = recRPOSITION(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastRayClosest(&a, rdr)
		}
	case 0xE6:
		var a recArgsQueryCastMover

		a.origin = recRPOSITION(rdr)
		a.mover = recRCAPSULE(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastMover(&a, rdr)
		}
	case 0xE7:
		var a recArgsShapeTestPoint

		a.shape = recRSHAPEID(rdr)

		a.point = recRPOSITION(rdr)
		if rdr.ok {
			recDispatchShapeTestPoint(&a, rdr)
		}
	case 0xE8:
		var a recArgsShapeRayCast

		a.shape = recRSHAPEID(rdr)
		a.origin = recRPOSITION(rdr)

		a.translation = recRVEC2(rdr)
		if rdr.ok {
			recDispatchShapeRayCast(&a, rdr)
		}
	case 0xF1:
		var a recArgsStateHash

		a.hash = recRU64(rdr)
		if rdr.ok {
			recDispatchStateHash(&a, rdr)
		}
	case 0xF2:
		var a recArgsRecordingBounds

		a.bounds = recRAABB(rdr)
		if rdr.ok {
			recDispatchRecordingBounds(&a, rdr)
		}
	default:
		rdr.replayWorldID.logger.Warn("box2d: replay skipped unknown opcode", "opcode", opcode, "bytes", payloadSize)

		if uint64(payloadSize) > uint64(rdr.size-payloadStart) {
			rdr.ok = false
		} else {
			rdr.cursor = payloadStart + int(payloadSize)
		}
	}

	return int(opcode)
}

func recScanFile(player *Replay) {
	data := player.data
	size := player.size
	cursor := player.headerEnd
	frameCount := 0
	gotStep := false

	for cursor <= size-4 {
		opcode := data[cursor]
		payloadSize := uint32(data[cursor+1]) | uint32(data[cursor+2])<<8 | uint32(data[cursor+3])<<16

		payloadStart := cursor + 4
		if int(payloadSize) > size-payloadStart {
			break
		}

		if opcode == 0x80 {
			frameCount++

			if !gotStep && payloadSize >= 8 {
				player.recordedDt = math.Float32frombits(binary.LittleEndian.Uint32(data[payloadStart:]))
				player.recordedSubStepCount = int(int32(binary.LittleEndian.Uint32(data[payloadStart+4:])))
				gotStep = true
			}
		} else if opcode == 0xf2 && payloadSize >= 16 {
			rdr := recReader{data: data, size: size, cursor: payloadStart, ok: true}
			player.bounds = recRAABB(&rdr)
		}

		cursor = payloadStart + int(payloadSize)
	}

	player.frameCount = frameCount
}

// Open a recording for incremental playback. The player copies the bytes,
// so you may free or destroy the source buffer immediately after this call.
func NewReplay(data []byte, workerCount int) (*Replay, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("%w: recording too small", ErrCorrupt)
	}
	// Validate the header before copying anything
	magic := binary.LittleEndian.Uint32(data)
	major := binary.LittleEndian.Uint16(data[4:])
	minor := binary.LittleEndian.Uint16(data[6:])
	lengthScale := math.Float32frombits(binary.LittleEndian.Uint32(data[12:]))

	if magic != recMagic {
		return nil, fmt.Errorf("%w: recording magic 0x%08X", ErrBadMagic, magic)
	}

	if major != recVersionMajor || minor != recVersionMinor {
		return nil, fmt.Errorf("%w: recording %d.%d, runtime %d.%d", ErrVersion, major, minor, recVersionMajor, recVersionMinor)
	}

	if data[18] != 0 {
		return nil, fmt.Errorf("%w: big-endian recording", ErrLayout)
	}

	snapshotSize := binary.LittleEndian.Uint64(data[24:])
	if snapshotSize == 0 || snapshotSize > uint64(len(data)-32) {
		return nil, fmt.Errorf("%w: missing or oversized snapshot", ErrCorrupt)
	}

	headerEnd := 32 + int(snapshotSize)

	owned := append([]byte(nil), data...)
	player := &Replay{
		data: owned, size: len(owned), headerEnd: headerEnd, lengthScale: lengthScale,
		previousLengthScale: LengthUnitsPerMeter(), divergeFrame: -1,
		keyframeBudget: recKeyframeBudgetDefault, keyframeMinInterval: recKeyframeIntervalDefault, keyframeInterval: recKeyframeIntervalDefault,
	}
	player.rdr = recReader{data: owned, size: len(owned), cursor: headerEnd, workerCount: workerCount, ok: true, owner: player}
	// Override the global length scale with the recording's so replay reproduces the same constants.
	// This is global engine state and affects the caller's other worlds, so the previous value was
	// captured above and is restored in Destroy.
	if lengthScale > 0 {
		SetLengthUnitsPerMeter(lengthScale)
	}

	recScanFile(player)

	world, err := NewWorldFromSnapshot(owned[32:headerEnd], workerCount)

	player.rdr.replayWorldID = world
	if err != nil {
		player.Close()
		return nil, fmt.Errorf("box2d: recording snapshot: %w", err)
	}

	player.recordedWorkerCount = workerCount
	player.frame0Image = owned[32:headerEnd]
	player.frame0Size = int(snapshotSize)
	recSeedBodyIDs(player)

	player.frame0BodyIDCount = player.bodyIDCount
	if player.bodyIDCount > 0 {
		player.frame0BodyIDs = append([]Body(nil), player.bodyIDs[:player.bodyIDCount]...)
	}

	return player, nil
}

func freeKeyframe(kf *recKeyframe) {
	kf.image = nil
	kf.bodyIDs = nil
}

func recCaptureKeyframe(player *Replay) {
	world := player.rdr.replayWorldID

	var buf recBuffer
	serializeWorld(world, &buf)

	bodyBytes := uint(player.bodyIDCount) * 8
	newBytes := uint(buf.capacity) + bodyBytes
	// Make room under the budget: doubling the spacing drops the off-grid keyframes, roughly halving
	// the bytes, until the new keyframe fits or only it remains. The budget is soft in the corner
	// where a single snapshot already exceeds it.
	for player.keyframeCount > 0 && player.keyframeBytes+newBytes > player.keyframeBudget {
		player.keyframeInterval *= 2
		kept := 0

		var keptBytes uint

		for i := range player.keyframeCount {
			kf := &player.keyframes[i]
			if kf.frame%player.keyframeInterval == 0 {
				player.keyframes[kept] = *kf
				keptBytes += uint(kf.imageCapacity) + uint(kf.bodyIDCount)*8
				kept++
			} else {
				freeKeyframe(kf)
			}
		}

		progress := kept < player.keyframeCount
		clear(player.keyframes[kept:player.keyframeCount])
		player.keyframeCount = kept
		player.keyframeBytes = keptBytes

		if !progress {
			break
		}
	}

	recGrow(&player.keyframes, &player.keyframeCapacity, player.keyframeCount+1, player.keyframeCount, int(unsafe.Sizeof(recKeyframe{})))
	kf := &player.keyframes[player.keyframeCount]

	*kf = recKeyframe{
		image: buf.data, imageSize: buf.size, imageCapacity: buf.capacity, frame: player.frame, cursor: player.rdr.cursor,
		divergeFrame: player.divergeFrame, diverged: player.rdr.diverged, bodyIDCount: player.bodyIDCount,
	}
	if bodyBytes > 0 {
		kf.bodyIDs = append([]Body(nil), player.bodyIDs[:player.bodyIDCount]...)
	}

	player.keyframeBytes += newBytes
	player.keyframeCount++
	player.lastKeyframeFrame = player.frame
}

func replayRestoreKeyframe(player *Replay, kf *recKeyframe) {
	if player.rdr.replayWorldID.UnmarshalBinary(kf.image[:kf.imageSize]) != nil {
		player.rdr.ok = false
		return
	}

	player.rdr.cursor = kf.cursor
	player.rdr.ok = true
	player.rdr.diverged = kf.diverged
	player.frame = kf.frame
	player.divergeFrame = kf.divergeFrame
	player.atEnd = false
	recGrow(&player.bodyIDs, &player.bodyIDCap, kf.bodyIDCount, 0, 8)

	player.bodyIDCount = kf.bodyIDCount
	if kf.bodyIDCount > 0 {
		copy(player.bodyIDs[:kf.bodyIDCount], kf.bodyIDs[:kf.bodyIDCount])
	}
}

// Advance the replay by one recorded step.
func (r *Replay) StepFrame() bool {
	if r.atEnd {
		return false
	}

	r.frameQueryCount = 0
	r.frameHitCount = 0
	// Run this frame's Step, then consume the records that trail it (StateHash, queries, any
	// between-frame mutators) up to the next Step. The queries and hash for a frame are recorded
	// after its Step, so grouping them with that Step keeps them paired with the world state they
	// were computed against. Stopping before the next Step is what advances exactly one frame.
	stepped := false

	for {
		if r.rdr.cursor >= r.rdr.size || !r.rdr.ok {
			r.atEnd = true
			return stepped
		}

		if stepped && r.rdr.data[r.rdr.cursor] == 0x80 {
			if r.frame > r.lastKeyframeFrame && r.frame%r.keyframeInterval == 0 {
				recCaptureKeyframe(r)
			}

			return true
		}

		opcode := recDispatchOne(r)
		if opcode < 0 {
			r.atEnd = true
			return stepped
		}

		if opcode == 0x80 {
			r.frame++
			stepped = true
		}

		if r.divergeFrame < 0 && r.rdr.diverged {
			r.divergeFrame = r.frame
		}
	}
}

// Rewind the player to the first step, recreating the replay world from the file.
func (r *Replay) Restart() {
	if r.rdr.replayWorldID.UnmarshalBinary(r.frame0Image[:r.frame0Size]) != nil {
		r.rdr.ok = false
		return
	}

	r.rdr.cursor = r.headerEnd
	r.rdr.ok = true
	r.rdr.diverged = false
	r.frame = 0
	r.divergeFrame = -1
	r.atEnd = false
	// Frame 0 is the pre-step snapshot, so it has no recorded queries. Clear the per-frame store so
	// the last stepped frame's queries do not linger on a load or a backward scrub to the start.
	r.frameQueryCount = 0
	r.frameHitCount = 0

	r.bodyIDCount = r.frame0BodyIDCount
	if r.frame0BodyIDCount > 0 {
		copy(r.bodyIDs[:r.frame0BodyIDCount], r.frame0BodyIDs[:r.frame0BodyIDCount])
	}
}

// Get the replayed world.
func (r *Replay) World() *World {
	if r == nil {
		return nil
	}

	return r.rdr.replayWorldID
}

// Get the number of steps replayed so far.
func (r *Replay) Frame() int {
	if r == nil {
		return 0
	}

	return r.frame
}

// Seek to a recorded step. Seeking backward rewinds and re-runs from the start, so the
// cost grows with the target frame. Clamps to the recording bounds.
func (r *Replay) SeekFrame(targetFrame int) {
	if r == nil {
		return
	}

	if targetFrame < 0 {
		targetFrame = 0
	}
	// Resume from the nearest keyframe strictly below the target when it beats the current cursor.
	// A backward seek must restore since the cursor cannot rewind. A forward seek restores only when
	// a keyframe sits ahead of the cursor, capping a long forward fling at one keyframe interval of
	// replay instead of every intervening frame. Strictly below so the step loop still runs the
	// target frame and regenerates its per-frame query store, body list, and divergence latch
	// exactly as a plain forward replay would.
	var best *recKeyframe

	for i := range r.keyframeCount {
		kf := &r.keyframes[i]
		if kf.frame < targetFrame && (best == nil || kf.frame > best.frame) {
			best = kf
		}
	}

	if targetFrame < r.frame {
		if best != nil {
			replayRestoreKeyframe(r, best)
		} else {
			r.Restart()
		}
	} else if best != nil && best.frame > r.frame {
		replayRestoreKeyframe(r, best)
	}

	for r.frame < targetFrame && r.StepFrame() {
	}
}

// Get static metadata for the recording (frame count, recorded tuning, time).
func (r *Replay) Info() ReplayInfo {
	if r == nil {
		return ReplayInfo{}
	}

	return ReplayInfo{
		FrameCount: r.frameCount, WorkerCount: r.recordedWorkerCount, TimeStep: r.recordedDt,
		SubStepCount: r.recordedSubStepCount, LengthScale: r.lengthScale, Bounds: r.bounds,
	}
}

// Returns true once the end of the recording has been reached.
func (r *Replay) IsAtEnd() bool {
	if r == nil {
		return true
	}

	return r.atEnd
}

// Returns true if a recorded state hash failed to reproduce, meaning replay diverged.
func (r *Replay) HasDiverged() bool {
	if r == nil {
		return false
	}

	return r.rdr.diverged
}

// Get the first step at which replay diverged, or -1 if it has not diverged.
func (r *Replay) DivergeFrame() int {
	if r == nil {
		return -1
	}

	return r.divergeFrame
}

// Tune the keyframe ring used to speed up backward seeking. A keyframe is a periodic snapshot the
// player restores from instead of replaying from the start, trading memory for seek speed.
// A zero budget or a non-positive interval keeps that value. Clears the existing ring, so call
// Restart afterward to repopulate it under the new policy.
func (r *Replay) SetKeyframePolicy(budgetBytes uint, minIntervalFrames int) {
	if r == nil {
		return
	}

	if budgetBytes > 0 {
		r.keyframeBudget = budgetBytes
	}

	if minIntervalFrames > 0 {
		r.keyframeMinInterval = minIntervalFrames
	}

	for i := range r.keyframeCount {
		freeKeyframe(&r.keyframes[i])
	}

	r.keyframeCount = 0
	r.keyframeBytes = 0
	r.keyframeInterval = r.keyframeMinInterval
	r.lastKeyframeFrame = 0
}

// Get the keyframe memory budget in bytes.
func (r *Replay) KeyframeBudget() uint {
	if r == nil {
		return 0
	}

	return r.keyframeBudget
}

// Get the finest keyframe spacing in frames.
func (r *Replay) KeyframeMinInterval() int {
	if r == nil {
		return 0
	}

	return r.keyframeMinInterval
}

// Get the current keyframe spacing in frames. Starts at the min interval and doubles as the ring
// evicts to stay under budget, so it reflects the effective backward-seek granularity right now.
func (r *Replay) KeyframeInterval() int {
	if r == nil {
		return 0
	}

	return r.keyframeInterval
}

// Get the memory currently held by keyframe snapshots, in bytes.
func (r *Replay) KeyframeBytes() uint {
	if r == nil {
		return 0
	}

	return r.keyframeBytes
}

// Close a player and free its replay world and file buffer.
func (r *Replay) Close() {
	if r == nil {
		return
	}

	if r.rdr.replayWorldID.IsValid() {
		r.rdr.replayWorldID.Close()
	}

	r.data = nil
	r.rdr.data = nil
	r.rdr.chainPoints = nil
	r.rdr.chainMaterials = nil
	r.rdr.hits = nil
	r.frameQueries = nil
	r.frameHits = nil
	r.bodyIDs = nil
	r.frame0Image = nil

	r.frame0BodyIDs = nil
	for i := range r.keyframeCount {
		freeKeyframe(&r.keyframes[i])
	}

	r.keyframes = nil
	SetLengthUnitsPerMeter(r.previousLengthScale)
}

func recDrawHitAABBs(player *Replay, q *recDrawQuery, draw *DebugDraw) {
	for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
		id := player.frameHits[hi].id
		if !id.IsValid() {
			continue
		}

		b := id.AABB()
		lower, upper := b.LowerBound, b.UpperBound
		vs := []Vec2{lower, {X: upper.X, Y: lower.Y}, upper, {X: lower.X, Y: upper.Y}}
		draw.DrawPolygon(WorldTransform{Q: Rot{C: 1}}, vs, ColorMagenta)
	}
}

// Draw spatial queries recorded during the most recently replayed frame.
// Call after World_Draw so queries are layered on top of the world.
func (r *Replay) DrawFrameQueries(draw *DebugDraw, queryIndex int) {
	if r == nil || draw == nil {
		return
	}

	checkInput(draw.Drawer != nil, "draw.Drawer != nil")
	// queryIndex < 0 draws all queries, otherwise just the one selected in the viewer
	for qi := range r.frameQueryCount {
		if queryIndex >= 0 && qi != queryIndex {
			continue
		}

		q := &r.frameQueries[qi]
		switch recQueryKind(q.kind) {
		case recqCastRay, recqCastRayClosest:
			// Ray origin to endpoint
			origin := q.origin
			end := origin.Add(q.translation)
			draw.DrawLine(origin, end, ColorYellow)
			// Per-hit point + short normal
			for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
				h := &r.frameHits[hi]
				point := h.point
				draw.DrawPoint(point, 4, ColorYellow)
				np := point.Add(h.normal.Scale(0.2))
				draw.DrawLine(point, np, ColorLightYellow)
			}
		case recqCastShape:
			// Shape cast: draw per-hit points along the swept path
			for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
				h := &r.frameHits[hi]
				point := h.point
				draw.DrawPoint(point, 4, ColorSkyBlue)
				np := point.Add(h.normal.Scale(0.2))
				draw.DrawLine(point, np, ColorLightSkyBlue)
			}
		case recqCastMover:
			// The mover capsule is relative to the query origin
			c1 := q.origin.Add(q.mover.Center1)
			c2 := q.origin.Add(q.mover.Center2)
			draw.DrawSolidCapsule(c1, c2, q.mover.Radius, ColorLightSkyBlue)
		case recqOverlapAABB:
			// The query box is relative to the query origin
			lower, upper := q.aabb.LowerBound, q.aabb.UpperBound
			vs := []Vec2{lower, {X: upper.X, Y: lower.Y}, upper, {X: lower.X, Y: upper.Y}}
			draw.DrawPolygon(WorldTransform{P: q.origin, Q: Rot{C: 1}}, vs, ColorLimeGreen)
			recDrawHitAABBs(r, q, draw)
		case recqOverlapShape:
			// The proxy points are relative to the query origin
			if q.proxy.Count == 1 {
				draw.DrawCircle(q.origin.Add(q.proxy.Points[0]), q.proxy.Radius, ColorLimeGreen)
			} else if q.proxy.Count >= 2 {
				draw.DrawPolygon(WorldTransform{P: q.origin, Q: Rot{C: 1}}, q.proxy.Points[:q.proxy.Count], ColorLimeGreen)
			}

			recDrawHitAABBs(r, q, draw)
		case recqCollideMover:
			// The mover capsule and the collision planes are relative to the query origin
			c1 := q.origin.Add(q.mover.Center1)
			c2 := q.origin.Add(q.mover.Center2)
			draw.DrawSolidCapsule(c1, c2, q.mover.Radius, ColorTan)
			// Per-hit plane point and normal
			for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
				h := &r.frameHits[hi]
				if h.plane.Hit {
					point := q.origin.Add(h.plane.Point)
					np := point.Add(h.plane.Plane.Normal.Scale(0.2))
					draw.DrawLine(point, np, ColorOrange)
				}
			}
		case recqShapeTestPoint:
			c := ColorRed
			if q.boolResult {
				c = ColorAqua
			}

			draw.DrawPoint(q.origin, 6, c)
		case recqShapeRayCast:
			origin := q.origin
			end := origin.Add(q.translation)
			draw.DrawLine(origin, end, ColorViolet)

			if q.castOut.Hit {
				draw.DrawPoint(q.castOut.Point, 4, ColorViolet)
			}
		}
	}
}

// Get the number of spatial queries recorded for the most recently replayed frame.
func (r *Replay) FrameQueryCount() int {
	if r == nil {
		return 0
	}

	return r.frameQueryCount
}

// Get a recorded query from the most recently replayed frame by index.
func (r *Replay) FrameQuery(index int) ReplayQueryInfo {
	if r == nil || index < 0 || index >= r.frameQueryCount {
		return ReplayQueryInfo{}
	}

	q := &r.frameQueries[index]

	return ReplayQueryInfo{
		Type: ReplayQueryType(q.kind), Filter: q.filter, AABB: q.aabb, Origin: q.origin,
		Translation: q.translation, Shape: q.shape, HitCount: q.hitCount,
	}
}

// Get one result of a recorded query from the most recently replayed frame.
func (r *Replay) FrameQueryHit(queryIndex int, hitIndex int) ReplayQueryHit {
	if r == nil || queryIndex < 0 || queryIndex >= r.frameQueryCount {
		return ReplayQueryHit{}
	}

	q := &r.frameQueries[queryIndex]
	if hitIndex < 0 || hitIndex >= q.hitCount {
		return ReplayQueryHit{}
	}

	h := &r.frameHits[q.hitStart+hitIndex]

	return ReplayQueryHit{Shape: h.id, Point: h.point, Normal: h.normal, Fraction: h.fraction}
}

// Get the number of body slots tracked for the outliner. This is the creation-order span and
// includes holes for destroyed bodies, so it only grows as the replay advances.
func (r *Replay) BodyCount() int {
	if r == nil {
		return 0
	}

	return r.bodyIDCount
}

// Get a tracked body by creation ordinal. Returns the zero Body for a destroyed slot, an ordinal not
// yet reached at the current frame, or an out-of-range index. Validate with Body.IsValid.
func (r *Replay) Body(index int) Body {
	if r == nil || index < 0 || index >= r.bodyIDCount {
		return Body{}
	}

	return r.bodyIDs[index]
}

// Replay a recording by re-running the engine and checking that recorded ids and state match.
// It returns ErrDiverged when the replay diverges and ErrCorrupt when a frame cannot be read.
func ValidateReplay(data []byte, workerCount int) error {
	player, err := NewReplay(data, workerCount)
	if err != nil {
		return err
	}
	defer player.Close()

	for player.StepFrame() {
		if player.rdr.diverged {
			break
		}
	}

	if player.rdr.diverged {
		return ErrDiverged
	}

	if !player.rdr.ok {
		return ErrCorrupt
	}

	return nil
}
