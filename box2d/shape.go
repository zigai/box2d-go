// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"iter"
	"math"

	"github.com/zigai/box2d-go/internal/array"
	"github.com/zigai/box2d-go/internal/core"
	"github.com/zigai/box2d-go/internal/idpool"
)

const (
	// A circle with an offset.
	CircleShape ShapeType = 0
	// A capsule is an extruded circle.
	CapsuleShape ShapeType = 1
	// A line segment.
	SegmentShape ShapeType = 2
	// A convex polygon.
	PolygonShape ShapeType = 3
	// A line segment owned by a chain shape.
	ChainSegmentShape ShapeType = 4
	// The number of shape types.
	ShapeTypeCount ShapeType = 5
)

const (
	DefaultCategoryBits uint64 = 1
	DefaultMaskBits     uint64 = 0xffffffffffffffff
)

// This is used to filter collision on shapes. It affects shape-vs-shape collision
// and shape-versus-query collision (such as World.CastRay).
type Filter struct {
	// The collision category bits. Normally you would just set one bit. The category bits should
	// represent your application object types. For example:
	//
	// const (
	//    Static = 0x00000001
	//    Dynamic = 0x00000002
	//    Debris = 0x00000004
	//    Player = 0x00000008
	// )
	CategoryBits uint64
	// The collision mask bits. This states the categories that this
	// shape would accept for collision.
	// For example, you may want your player to only collide with static objects
	// and other players.
	//
	// maskBits := Static | Player
	MaskBits uint64
	// Collision groups allow a certain group of objects to never collide (negative)
	// or always collide (positive). A group index of zero has no effect. Non-zero group filtering
	// always wins against the mask bits.
	// For example, you may want ragdolls to collide with other ragdolls but you don't want
	// ragdoll self-collision. In this case you would give each ragdoll a unique negative group index
	// and apply that group index to all shapes on the ragdoll.
	GroupIndex int
}

// The query filter is used to filter collisions between queries and shapes. For example,
// you may want a ray-cast representing a projectile to hit players and the static environment
// but not debris.
type QueryFilter struct {
	// The collision category bits of this query. Normally you would just set one bit.
	CategoryBits uint64
	// The collision mask bits. This states the shape categories that this
	// query would accept for collision.
	MaskBits uint64
}

// Shape type.
type ShapeType int

// Surface materials allow chain shapes to have per segment surface properties.
type SurfaceMaterial struct {
	// The Coulomb (dry) friction coefficient, usually in the range [0,1].
	Friction float32
	// The coefficient of restitution (bounce) usually in the range [0,1].
	// https://en.wikipedia.org/wiki/Coefficient_of_restitution
	Restitution float32
	// The rolling resistance usually in the range [0,1].
	RollingResistance float32
	// The tangent speed for conveyor belts
	TangentSpeed float32
	// User material identifier. This is passed with query results and to friction and restitution
	// combining functions. It is not used internally.
	UserMaterialID uint64
	// Custom debug draw color.
	CustomColor uint32
}

// Used to create a shape.
// This is a temporary object used to bundle shape creation parameters. You may use
// the same shape definition to create multiple shapes.
// Must be initialized using DefaultShapeDef().
type ShapeDef struct {
	// Use this to store application specific shape data.
	UserData any
	// The surface material for this shape.
	Material SurfaceMaterial
	// The density, usually in kg/m^2.
	// This is not part of the surface material because this is for the interior, which may have
	// other considerations, such as being hollow. For example a wood barrel may be hollow or full of water.
	Density float32
	// Collision filtering data.
	Filter Filter
	// Enable custom filtering. Only one of the two shapes needs to enable custom filtering. See WorldDef.
	EnableCustomFiltering bool
	// A sensor shape generates overlap events but never generates a collision response.
	// Sensors do not have continuous collision. Instead, use a ray or shape cast for those scenarios.
	// Sensors still contribute to the body mass if they have non-zero density.
	// Note: Sensor events are disabled by default.
	// See: SetSensorEventsEnabled
	IsSensor bool
	// Enable sensor events for this shape. This applies to sensors and non-sensors. Both shapes involved must have this
	// flag set to true. False by default, even for sensors.
	EnableSensorEvents bool
	// Enable contact events for this shape. Only applies to kinematic and dynamic bodies. Only one shape involved needs
	// this flag set to true. Ignored for sensors. False by default.
	EnableContactEvents bool
	// Enable hit events for this shape. Only applies to kinematic and dynamic bodies. Only one shape involved needs
	// this flag set to true. Ignored for sensors. False by default.
	EnableHitEvents bool
	// Enable pre-solve contact events for this shape. Only applies to dynamic bodies. These are expensive
	// and must be carefully handled due to multithreading. Ignored for sensors.
	EnablePreSolveEvents bool
	// When shapes are created they will scan the environment for collision the next time step. This can significantly
	// slow down static body creation when there are many static shapes.
	// This is flag is ignored for dynamic and kinematic shapes which always invoke contact creation.
	InvokeContactCreation bool
	// Should the body update the mass properties when this shape is created. Default is true.
	// Warning: if this is false, you MUST call Body.UpdateMassFromShapes or Body.SetMassData before simulating the
	// world.
	UpdateBodyMass bool
	internalValue  int32
}

// Used to create a chain of line segments. This is designed to eliminate ghost collisions with some limitations.
// - chains are one-sided
// - chains have no mass and should be used on static bodies
// - chains have a counter-clockwise winding order (normal points right of segment direction)
// - chains are either a loop or open
// - the distance between any two points must be greater than LinearSlop
// - a chain shape should not self intersect (this is not validated)
// - an open chain shape needs a leading and trailing ghost point.
// - you may overlap two open chains using the ghost points to get smooth collision.
// - a chain shape creates multiple line segment shapes on the body
// https://en.wikipedia.org/wiki/Polygonal_chain
// Must be initialized using DefaultChainDef().
// Warning: Do not use chain shapes unless you understand the limitations. This is an advanced feature.
type ChainDef struct {
	// Use this to store application specific shape data.
	UserData any
	// Array of chain points. Each consecutive pair forms a segment. For a loop
	// the last and first point form a segment. At least 2 points for an open chain
	// and at least 3 for a loop. Cloned.
	Points []Vec2
	// Leading ghost point for an open chain. Distance from first point must be greater
	// than LinearSlop. Ignored for loops.
	Ghost1 Vec2
	// Ending ghost point for an open chain. Distance from last point must be greater
	// than LinearSlop. Ignored for loops.
	Ghost2 Vec2
	// One material for the whole chain or one for each segment. Cloned.
	Materials []SurfaceMaterial
	// Contact filtering data.
	Filter Filter
	// Indicates a closed chain formed by connecting the first and last point.
	IsLoop bool
	// Enable sensors to detect this chain. False by default.
	EnableSensorEvents bool
	internalValue      int32
}

type shape struct {
	id                    int
	bodyID                int
	prevShapeID           int
	nextShapeID           int
	sensorIndex           int
	shapeType             ShapeType
	material              SurfaceMaterial
	density               float32
	aabbMargin            float32
	aabb                  AABB
	localCentroid         Vec2
	proxyKey              int
	filter                Filter
	userData              any
	capsule               Capsule
	circle                Circle
	polygon               Polygon
	segment               Segment
	chainSegment          ChainSegment
	generation            uint16
	enableSensorEvents    bool
	enableContactEvents   bool
	enableCustomFiltering bool
	enableHitEvents       bool
	enablePreSolveEvents  bool
}

type chainShape struct {
	id           int
	bodyID       int
	nextChainID  int
	segmentCount int
	generation   uint16
	shapeIndices []int
}

type shapeExtent struct {
	minExtent float32
	maxExtent float32
}

func getShape(world *World, shapeID Shape) *shape {
	id := int(shapeID.index1) - 1
	shape := &world.shapes[id]
	core.Assert(
		shape.id == id && shape.generation == shapeID.generation,
		"shape.id == id && shape.generation == shapeID.generation",
	)

	return shape
}

func getChainShape(world *World, chainID Chain) *chainShape {
	id := int(chainID.index1) - 1
	chain := &world.chainShapes[id]
	core.Assert(
		chain.id == id && chain.generation == chainID.generation,
		"chain.id == id && chain.generation == chainID.generation",
	)

	return chain
}

func computeShapeMargin(shape *shape) float32 {
	var margin float32

	switch shape.shapeType {
	case CapsuleShape:
		{
			margin = float32(
				0.5*shape.capsule.Center2.Distance(shape.capsule.Center1),
			) + shape.capsule.Radius
		}

	case CircleShape:
		{
			margin = shape.circle.Radius
		}

	case PolygonShape:
		{
			poly := &shape.polygon
			maxExtentSqr := float32(0.0)

			count := poly.Count
			for i := range count {
				distanceSqr := poly.Vertices[i].DistanceSquared(poly.Centroid)
				maxExtentSqr = maxFloat(maxExtentSqr, distanceSqr)
			}

			margin = float32(math.Sqrt(float64(maxExtentSqr)))
		}

	case SegmentShape:
		{
			margin = 0.5 * shape.segment.Point1.Distance(shape.segment.Point2)
		}

	case ChainSegmentShape:
		{
			margin = 0.5 * shape.chainSegment.Segment.Point1.Distance(
				shape.chainSegment.Segment.Point2,
			)
		}

	case ShapeTypeCount:
		fallthrough // The count sentinel uses the invalid-geometry path.
	default:
		core.Validate(false, "false")
		return MaxAABBMargin()
	}

	return minFloat(MaxAABBMargin(), AABBMarginFraction*margin)
}

func updateShapeAABBs(shape *shape, fatAABB *AABB, transform WorldTransform, proxyType BodyType) {
	// Compute a bounding box with a speculative margin
	speculativeDistance := SpeculativeDistance()
	aabbMargin := shape.aabbMargin

	aabb := computeFatShapeAABB(shape, transform, speculativeDistance)
	shape.aabb = aabb

	// Smaller margin for static bodies. Cannot be zero due to TOI tolerance.
	margin := aabbMargin
	if proxyType == StaticBody {
		margin = speculativeDistance
	}

	fatAABB.LowerBound.X = aabb.LowerBound.X - margin
	fatAABB.LowerBound.Y = aabb.LowerBound.Y - margin
	fatAABB.UpperBound.X = aabb.UpperBound.X + margin
	fatAABB.UpperBound.Y = aabb.UpperBound.Y + margin
}

func setShapeGeometry(shape *shape, geometry any, shapeType ShapeType) {
	switch shapeType {
	case CapsuleShape:
		shape.capsule = *checkedShapeGeometry[Capsule](geometry, "capsule shape requires *Capsule geometry")
	case CircleShape:
		shape.circle = *checkedShapeGeometry[Circle](geometry, "circle shape requires *Circle geometry")
	case PolygonShape:
		shape.polygon = *checkedShapeGeometry[Polygon](geometry, "polygon shape requires *Polygon geometry")
	case SegmentShape:
		shape.segment = *checkedShapeGeometry[Segment](geometry, "segment shape requires *Segment geometry")
	case ChainSegmentShape:
		shape.chainSegment = *checkedShapeGeometry[ChainSegment](
			geometry,
			"chain segment shape requires *ChainSegment geometry",
		)
	case ShapeTypeCount:
		fallthrough // The count sentinel uses the invalid-geometry path.
	default:
		core.Assert(false, "false")
	}
}

func checkedShapeGeometry[T any](geometry any, mismatch string) *T {
	value, ok := geometry.(*T)
	if !ok {
		panic(mismatch)
	}

	return value
}

func createShapeInternal(
	world *World,
	body *body,
	transform WorldTransform,
	def *ShapeDef,
	geometry any,
	shapeType ShapeType,
) *shape {
	shapeID := idpool.AllocWithGeneration(&world.shapeIDPool, func(id int) uint16 {
		return world.shapes[id].generation
	})

	if shapeID == len(world.shapes) {
		world.shapes = array.Push(world.shapes, shape{})
		world.fatAABBs = array.Push(world.fatAABBs, AABB{})
	} else {
		core.Assert(world.shapes[shapeID].id == nullIndex, "world.shapes[shapeID].id == nullIndex")
	}

	core.Assert(
		len(world.fatAABBs) == len(world.shapes),
		"len(world.fatAABBs) == len(world.shapes)",
	)

	shape := &world.shapes[shapeID]

	setShapeGeometry(shape, geometry, shapeType)

	shape.id = shapeID
	shape.bodyID = body.id
	shape.shapeType = shapeType
	shape.density = def.Density
	shape.material = def.Material
	shape.filter = def.Filter
	shape.userData = def.UserData
	shape.enableSensorEvents = def.EnableSensorEvents
	shape.enableContactEvents = def.EnableContactEvents
	shape.enableCustomFiltering = def.EnableCustomFiltering
	shape.enableHitEvents = def.EnableHitEvents
	shape.enablePreSolveEvents = def.EnablePreSolveEvents
	shape.proxyKey = nullIndex
	shape.localCentroid = getShapeCentroid(shape)
	shape.aabbMargin = computeShapeMargin(shape)
	shape.aabb = AABB{Vec2{}, Vec2{}}
	world.fatAABBs[shapeID] = AABB{Vec2{}, Vec2{}}
	shape.generation += 1

	if body.setIndex != disabledSet {
		proxyType := body.bodyType
		createShapeProxy(
			world,
			shape,
			proxyType,
			transform,
			def.InvokeContactCreation || def.IsSensor,
		)
	}

	// Add to shape doubly linked list
	if body.headShapeID != nullIndex {
		headShape := &world.shapes[body.headShapeID]
		headShape.prevShapeID = shapeID
	}

	shape.prevShapeID = nullIndex
	shape.nextShapeID = body.headShapeID
	body.headShapeID = shapeID
	body.shapeCount += 1

	if def.IsSensor {
		shape.sensorIndex = len(world.sensors)
		sensor := sensor{}
		sensor.shapeID = shapeID
		sensor.hits = make([]visitor, 0, 4)
		sensor.overlaps1 = make([]visitor, 0, 16)
		sensor.overlaps2 = make([]visitor, 0, 16)
		world.sensors = array.Push(world.sensors, sensor)
	} else {
		shape.sensorIndex = nullIndex
	}

	validateSolverSets(world)

	return shape
}

func isValidMaterial(material *SurfaceMaterial) bool {
	return IsValidFloat(material.Friction) && material.Friction >= 0.0 &&
		IsValidFloat(material.Restitution) &&
		material.Restitution >= 0.0 &&
		IsValidFloat(material.RollingResistance) &&
		material.RollingResistance >= 0.0 &&
		IsValidFloat(material.TangentSpeed)
}

func isValidCircle(circle *Circle) bool {
	return circle.Center.IsValid() && IsValidFloat(circle.Radius) && circle.Radius >= 0.0
}

func isValidCapsule(capsule *Capsule) bool {
	return capsule.Center1.IsValid() && capsule.Center2.IsValid() && IsValidFloat(capsule.Radius) &&
		capsule.Radius >= 0.0
}

func isValidSegment(segment *Segment) bool {
	return segment.Point1.IsValid() && segment.Point2.IsValid()
}

func isValidChainSegment(chainSegment *ChainSegment) bool {
	return chainSegment.Ghost1.IsValid() && isValidSegment(&chainSegment.Segment) &&
		chainSegment.Ghost2.IsValid()
}

func isValidPolygon(polygon *Polygon) bool {
	if polygon.Count < 1 || polygon.Count > MaxPolygonVertices {
		return false
	}

	if !IsValidFloat(polygon.Radius) || polygon.Radius < 0.0 {
		return false
	}

	if !polygon.Centroid.IsValid() {
		return false
	}

	for i := range polygon.Count {
		if !polygon.Vertices[i].IsValid() || !polygon.Normals[i].IsValid() {
			return false
		}
	}

	return true
}

func createShape(bodyID Body, def *ShapeDef, geometry any, shapeType ShapeType) Shape {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")
	checkInput(
		IsValidFloat(def.Density) && def.Density >= 0.0,
		"IsValidFloat( def.Density ) && def.Density >= 0.0",
	)
	checkInput(isValidMaterial(&def.Material), "isValidMaterial( &def.Material )")

	world := getWorldLocked(bodyID.world)
	if world == nil {
		return Shape{}
	}

	body := getBodyFullID(world, bodyID)
	transform := getBodyTransformQuick(world, body)

	shape := createShapeInternal(world, body, transform, def, geometry, shapeType)

	if def.UpdateBodyMass {
		updateBodyMassData(world, body)
	} else if (body.flags & dirtyMass) == 0 {
		body.flags |= dirtyMass
		syncBodyFlags(world, body)
	}

	validateSolverSets(world)

	id := Shape{index1: int32(shape.id + 1), world: bodyID.world, generation: shape.generation}

	return id
}

// Create a circle shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape.
func (b Body) CreateCircleShape(def *ShapeDef, circle *Circle) Shape {
	checkInput(isValidCircle(circle), "isValidCircle( circle )")

	id := createShape(b, def, circle, CircleShape)

	world := b.world
	if world.recording != nil {
		recWriteRetCreateCircleShape(
			world.recording,
			&recArgsCreateCircleShape{body: b, def: *def, circle: *circle},
			id,
		)
	}

	return id
}

// Create a capsule shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape, this will be Shape{} if the length is too small.
func (b Body) CreateCapsuleShape(def *ShapeDef, capsule *Capsule) Shape {
	checkInput(isValidCapsule(capsule), "isValidCapsule( capsule )")

	lengthSqr := capsule.Center1.DistanceSquared(capsule.Center2)
	if lengthSqr <= LinearSlop()*LinearSlop() {
		return Shape{}
	}

	id := createShape(b, def, capsule, CapsuleShape)

	world := b.world
	if world.recording != nil {
		recWriteRetCreateCapsuleShape(
			world.recording,
			&recArgsCreateCapsuleShape{body: b, def: *def, capsule: *capsule},
			id,
		)
	}

	return id
}

// Create a polygon shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape.
func (b Body) CreatePolygonShape(def *ShapeDef, polygon *Polygon) Shape {
	checkInput(isValidPolygon(polygon), "isValidPolygon( polygon )")

	id := createShape(b, def, polygon, PolygonShape)

	world := b.world
	if world.recording != nil {
		recWriteRetCreatePolygonShape(
			world.recording,
			&recArgsCreatePolygonShape{body: b, def: *def, polygon: *polygon},
			id,
		)
	}

	return id
}

// Create a line segment shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id or Shape{} if the segment is too short.
func (b Body) CreateSegmentShape(def *ShapeDef, segment *Segment) Shape {
	checkInput(isValidSegment(segment), "isValidSegment( segment )")

	lengthSqr := segment.Point1.DistanceSquared(segment.Point2)
	if lengthSqr <= LinearSlop()*LinearSlop() {
		core.Assert(false, "false")
		return Shape{}
	}

	id := createShape(b, def, segment, SegmentShape)

	world := b.world
	if world.recording != nil {
		recWriteRetCreateSegmentShape(
			world.recording,
			&recArgsCreateSegmentShape{body: b, def: *def, segment: *segment},
			id,
		)
	}

	return id
}

// Create an orphaned chain segment shape and attach it to a body. The shape definition and
// geometry are fully cloned. The caller is responsible for the segment's ghost vertices and
// lifetime. The segment is not owned by any chainShape (Shape.ParentChain returns
// Chain{}). Contacts are not created until the next time step.
// Returns: the shape id, or Shape{} if the segment is too short.
func (b Body) CreateChainSegmentShape(def *ShapeDef, chainSegment *ChainSegment) Shape {
	checkInput(isValidChainSegment(chainSegment), "isValidChainSegment( chainSegment )")

	lengthSqr := chainSegment.Segment.Point1.DistanceSquared(chainSegment.Segment.Point2)
	if lengthSqr <= LinearSlop()*LinearSlop() {
		core.Assert(false, "false")
		return Shape{}
	}

	// No parent chain shape
	local := *chainSegment
	local.ChainID = nullIndex

	id := createShape(b, def, &local, ChainSegmentShape)

	world := b.world
	if world.recording != nil {
		recWriteRetCreateChainSegmentShape(
			world.recording,
			&recArgsCreateChainSegmentShape{body: b, def: *def, chainSegment: local},
			id,
		)
	}

	return id
}

func destroyShapeInternal(world *World, shape *shape, body *body) {
	shapeID := shape.id

	// Remove the shape from the body's doubly linked list.
	if shape.prevShapeID != nullIndex {
		prevShape := &world.shapes[shape.prevShapeID]
		prevShape.nextShapeID = shape.nextShapeID
	}

	if shape.nextShapeID != nullIndex {
		nextShape := &world.shapes[shape.nextShapeID]
		nextShape.prevShapeID = shape.prevShapeID
	}

	if shapeID == body.headShapeID {
		body.headShapeID = shape.nextShapeID
	}

	body.shapeCount -= 1

	// Remove from broad-phase.
	destroyShapeProxy(shape, &world.broadPhase)

	// Destroy any contacts associated with the shape.
	contactKey := body.headContactKey
	for contactKey != nullIndex {
		contactID := contactKey >> 1
		edgeIndex := contactKey & 1

		contact := &world.contacts[contactID]
		contactKey = contact.edges[edgeIndex].nextKey

		if contact.shapeIDA == shapeID || contact.shapeIDB == shapeID {
			destroyContact(world, contact)
		}
	}

	if shape.sensorIndex != nullIndex {
		sensor := &world.sensors[shape.sensorIndex]
		for i := range len(sensor.overlaps2) {
			ref := &sensor.overlaps2[i]
			event := SensorEndTouchEvent{
				SensorShape: Shape{
					index1:     int32(shapeID + 1),
					world:      world,
					generation: shape.generation,
				},
				VisitorShape: Shape{
					index1:     int32(ref.shapeID + 1),
					world:      world,
					generation: ref.generation,
				},
			}

			world.sensorEndEvents[world.endEventArrayIndex] = array.Push(
				world.sensorEndEvents[world.endEventArrayIndex],
				event,
			)
		}

		// Destroy sensor
		sensor.hits = nil
		sensor.overlaps1 = nil
		sensor.overlaps2 = nil

		var movedIndex int

		world.sensors, movedIndex = array.RemoveSwap(world.sensors, shape.sensorIndex)
		if movedIndex != nullIndex {
			// Fixup moved sensor
			movedSensor := &world.sensors[shape.sensorIndex]
			otherSensorShape := &world.shapes[movedSensor.shapeID]
			otherSensorShape.sensorIndex = shape.sensorIndex
		}
	}

	// Return shape to free list.
	world.shapeIDPool.Free(shapeID)

	shape.id = nullIndex

	validateSolverSets(world)
}

// Destroy a shape. You may defer the body mass update which can improve performance if several shapes on a
//
//	body are destroyed at once.
//	See: Body.UpdateMassFromShapes
func (s Shape) Destroy(updateBodyMass bool) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	shape := getShape(world, s)

	// Cannot destroy a chain segment that has a parent chain shape
	if shape.shapeType == ChainSegmentShape && shape.chainSegment.ChainID != nullIndex {
		core.Assert(false, "false")
		return
	}

	if world.recording != nil {
		recWriteDestroyShape(
			world.recording,
			&recArgsDestroyShape{shape: s, updateBodyMass: updateBodyMass},
		)
	}

	body := &world.bodies[shape.bodyID]
	destroyShapeInternal(world, shape, body)

	if updateBodyMass {
		updateBodyMassData(world, body)
	}
}

func createChainSegments(
	world *World,
	body *body,
	transform WorldTransform,
	def *ChainDef,
	chainShape *chainShape,
) {
	chainID := chainShape.id
	segmentCount := chainShape.segmentCount
	materialCount := len(def.Materials)
	shapeDef := DefaultShapeDef()
	shapeDef.UserData = def.UserData
	shapeDef.Filter = def.Filter
	shapeDef.EnableSensorEvents = def.EnableSensorEvents
	shapeDef.EnableContactEvents = false
	shapeDef.EnableHitEvents = false

	points := def.Points
	n := segmentCount
	chainShape.shapeIndices = make([]int, n)

	tolSqr := LinearSlop() * LinearSlop()
	_ = tolSqr

	if def.IsLoop {
		var chainSegment ChainSegment

		prevIndex := n - 1
		for i := range n {
			chainSegment.Ghost1 = points[prevIndex]
			chainSegment.Segment.Point1 = points[i]
			chainSegment.Segment.Point2 = points[(i+1)%n]
			chainSegment.Ghost2 = points[(i+2)%n]

			core.Assert(
				chainSegment.Ghost1.DistanceSquared(chainSegment.Segment.Point1) > tolSqr,
				"DistanceSquared(chainSegment.Ghost1, chainSegment.Segment.Point1) > tolSqr",
			)
			core.Assert(
				chainSegment.Segment.Point1.DistanceSquared(chainSegment.Segment.Point2) > tolSqr,
				"DistanceSquared(chainSegment.Segment.Point1, chainSegment.Segment.Point2) > tolSqr",
			)

			chainSegment.ChainID = chainID
			prevIndex = i

			materialIndex := i
			if materialCount == 1 {
				materialIndex = 0
			}

			shapeDef.Material = def.Materials[materialIndex]

			shape := createShapeInternal(
				world,
				body,
				transform,
				&shapeDef,
				&chainSegment,
				ChainSegmentShape,
			)
			chainShape.shapeIndices[i] = shape.id
		}

		return
	}

	var chainSegment ChainSegment

	for i := range n {
		core.Assert(i+1 < len(def.Points), "i + 1 < len(def.Points)")

		if i == 0 {
			chainSegment.Ghost1 = def.Ghost1
		} else {
			chainSegment.Ghost1 = points[i-1]
		}

		chainSegment.Segment.Point1 = points[i+0]

		chainSegment.Segment.Point2 = points[i+1]
		if i == n-1 {
			chainSegment.Ghost2 = def.Ghost2
		} else {
			chainSegment.Ghost2 = points[i+2]
		}

		core.Assert(
			chainSegment.Ghost1.DistanceSquared(chainSegment.Segment.Point1) > tolSqr,
			"DistanceSquared(chainSegment.Ghost1, chainSegment.Segment.Point1) > tolSqr",
		)
		core.Assert(
			chainSegment.Segment.Point1.DistanceSquared(chainSegment.Segment.Point2) > tolSqr,
			"DistanceSquared(chainSegment.Segment.Point1, chainSegment.Segment.Point2) > tolSqr",
		)
		core.Assert(
			chainSegment.Segment.Point2.DistanceSquared(chainSegment.Ghost2) > tolSqr,
			"DistanceSquared(chainSegment.Segment.Point2, chainSegment.Ghost2) > tolSqr",
		)

		chainSegment.ChainID = chainID

		materialIndex := i
		if materialCount == 1 {
			materialIndex = 0
		}

		shapeDef.Material = def.Materials[materialIndex]

		shape := createShapeInternal(
			world,
			body,
			transform,
			&shapeDef,
			&chainSegment,
			ChainSegmentShape,
		)
		chainShape.shapeIndices[i] = shape.id
	}
}

// Create a chain shape
// See: ChainDef for details.
func (b Body) CreateChain(def *ChainDef) Chain {
	checkInput(def.internalValue == secretCookie, "def->internalValue == B2_SECRET_COOKIE")

	validPointCount := len(def.Points) >= 2
	if def.IsLoop {
		validPointCount = len(def.Points) >= 3
	}

	checkInput(validPointCount, "def.IsLoop ? len(def.Points) >= 3 : len(def.Points) >= 2")

	segmentCount := len(def.Points) - 1
	if def.IsLoop {
		segmentCount = len(def.Points)
	}

	checkInput(
		len(def.Materials) == 1 || len(def.Materials) == segmentCount,
		"len(def.Materials) == 1 || len(def.Materials) == segmentCount",
	)

	for i := range len(def.Points) {
		checkInput(def.Points[i].IsValid(), "IsValidVec2( def.Points[i] )")
	}

	for i := range len(def.Materials) {
		checkInput(isValidMaterial(&def.Materials[i]), "isValidMaterial( &def.Materials[i] )")
	}

	if !def.IsLoop {
		// These are set to INFINITY in the default def as a sentinel to ensure they are set.
		checkInput(def.Ghost1.IsValid(), "IsValidVec2( def.Ghost1 )")
		checkInput(def.Ghost2.IsValid(), "IsValidVec2( def.Ghost2 )")
	}

	world := getWorldLocked(b.world)
	if world == nil {
		return Chain{}
	}

	body := getBodyFullID(world, b)
	transform := getBodyTransformQuick(world, body)

	chainID := idpool.AllocWithGeneration(&world.chainIDPool, func(id int) uint16 {
		return world.chainShapes[id].generation
	})

	if chainID == len(world.chainShapes) {
		world.chainShapes = array.Push(world.chainShapes, chainShape{})
	} else {
		core.Assert(
			world.chainShapes[chainID].id == nullIndex,
			"world.chainShapes[chainID].id == nullIndex",
		)
	}

	chainShape := &world.chainShapes[chainID]

	chainShape.id = chainID
	chainShape.bodyID = body.id
	chainShape.nextChainID = body.headChainID
	chainShape.segmentCount = segmentCount
	chainShape.generation += 1

	body.headChainID = chainID
	createChainSegments(world, body, transform, def, chainShape)

	id := Chain{index1: int32(chainID + 1), world: world, generation: chainShape.generation}

	if world.recording != nil {
		recWriteRetCreateChain(world.recording, &recArgsCreateChain{body: b, def: *def}, id)
	}

	return id
}

func freeChainData(chain *chainShape) {
	chain.shapeIndices = nil
}

// Destroy a chain shape.
func (c Chain) Destroy() {
	world := getWorldLocked(c.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteDestroyChain(world.recording, &recArgsDestroyChain{chain: c})
	}

	chain := getChainShape(world, c)

	body := &world.bodies[chain.bodyID]

	// Remove the chain from the body's singly linked list.
	chainIDPtr := &body.headChainID
	found := false

	for *chainIDPtr != nullIndex {
		if *chainIDPtr == chain.id {
			*chainIDPtr = chain.nextChainID
			found = true

			break
		}

		chainIDPtr = &world.chainShapes[*chainIDPtr].nextChainID
	}

	core.Assert(found, "found == true")

	if !found {
		return
	}

	count := chain.segmentCount
	for i := range count {
		shapeID := chain.shapeIndices[i]
		shape := &world.shapes[shapeID]
		destroyShapeInternal(world, shape, body)
	}

	freeChainData(chain)

	// Return chain to free list.
	world.chainIDPool.Free(chain.id)
	chain.id = nullIndex

	validateSolverSets(world)
}

// Get the world that owns this chain shape.
func (c Chain) World() *World {
	return c.world
}

// Get the number of segments on this chain.
func (c Chain) SegmentCount() int {
	world := getWorldLocked(c.world)
	if world == nil {
		return 0
	}

	chain := getChainShape(world, c)

	return chain.segmentCount
}

// Get chain segment shape ids.
func (c Chain) Segments() iter.Seq[Shape] {
	world := getWorldLocked(c.world)
	if world == nil {
		return func(yield func(Shape) bool) {}
	}

	chain := getChainShape(world, c)

	return func(yield func(Shape) bool) {
		for i := range chain.segmentCount {
			shapeID := chain.shapeIndices[i]

			shape := &world.shapes[shapeID]
			if !yield(
				Shape{index1: int32(shapeID + 1), world: c.world, generation: shape.generation},
			) {
				return
			}
		}
	}
}

// Set a chain material by segment index.
func (c Chain) SetSurfaceMaterial(material *SurfaceMaterial, segmentIndex int) {
	checkInput(isValidMaterial(material), "isValidMaterial( material )")

	world := getWorldLocked(c.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteChainSetSurfaceMaterial(
			world.recording,
			&recArgsChainSetSurfaceMaterial{
				chain:        c,
				material:     *material,
				segmentIndex: int32(segmentIndex),
			},
		)
	}

	chainShape := getChainShape(world, c)
	core.Assert(
		0 <= segmentIndex && segmentIndex < chainShape.segmentCount,
		"0 <= segmentIndex && segmentIndex < chainShape.segmentCount",
	)
	shapeID := chainShape.shapeIndices[segmentIndex]
	shape := &world.shapes[shapeID]
	shape.material = *material
}

// Set the chain material on all segments.
func (c Chain) SetAllSurfaceMaterials(material *SurfaceMaterial) {
	checkInput(isValidMaterial(material), "isValidMaterial( material )")

	world := getWorldLocked(c.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteChainSetAllSurfaceMaterials(
			world.recording,
			&recArgsChainSetAllSurfaceMaterials{chain: c, material: *material},
		)
	}

	chainShape := getChainShape(world, c)

	for i := range chainShape.segmentCount {
		shapeID := chainShape.shapeIndices[i]
		shape := &world.shapes[shapeID]
		shape.material = *material
	}
}

// Get a chain material by segment index.
func (c Chain) SurfaceMaterial(segmentIndex int) SurfaceMaterial {
	world := c.world
	chainShape := getChainShape(world, c)
	core.Assert(
		0 <= segmentIndex && segmentIndex < chainShape.segmentCount,
		"0 <= segmentIndex && segmentIndex < chainShape.segmentCount",
	)
	shapeID := chainShape.shapeIndices[segmentIndex]
	shape := &world.shapes[shapeID]

	return shape.material
}

func computeShapeAABB(shape *shape, xf WorldTransform) AABB {
	switch shape.shapeType {
	case CapsuleShape:
		return ComputeCapsuleAABB(&shape.capsule, xf)
	case CircleShape:
		return ComputeCircleAABB(&shape.circle, xf)
	case PolygonShape:
		return ComputePolygonAABB(&shape.polygon, xf)
	case SegmentShape:
		return ComputeSegmentAABB(&shape.segment, xf)
	case ChainSegmentShape:
		return ComputeSegmentAABB(&shape.chainSegment.Segment, xf)
	case ShapeTypeCount:
		fallthrough // The count sentinel uses the invalid-geometry path.
	default:
		{
			core.Assert(false, "false")

			empty := AABB{}

			return empty
		}
	}
}

func getShapeCentroid(shape *shape) Vec2 {
	switch shape.shapeType {
	case CapsuleShape:
		return shape.capsule.Center1.Lerp(shape.capsule.Center2, 0.5)
	case CircleShape:
		return shape.circle.Center
	case PolygonShape:
		return shape.polygon.Centroid
	case SegmentShape:
		return shape.segment.Point1.Lerp(shape.segment.Point2, 0.5)
	case ChainSegmentShape:
		return shape.chainSegment.Segment.Point1.Lerp(shape.chainSegment.Segment.Point2, 0.5)
	case ShapeTypeCount:
		return Vec2{}
	default:
		return Vec2{}
	}
}

func getShapeProjectedPerimeter(shape *shape, line Vec2) float32 {
	switch shape.shapeType {
	case CapsuleShape:
		{
			axis := shape.capsule.Center2.Sub(shape.capsule.Center1)
			projectedLength := absFloat(axis.Dot(line))

			return projectedLength + float32(2.0*shape.capsule.Radius)
		}

	case CircleShape:
		return 2.0 * shape.circle.Radius

	case PolygonShape:
		{
			points := shape.polygon.Vertices
			count := shape.polygon.Count
			core.Assert(count > 0, "count > 0")

			value := points[0].Dot(line)
			lower := value

			upper := value
			for i := 1; i < count; i++ {
				value = points[i].Dot(line)
				lower = minFloat(lower, value)
				upper = maxFloat(upper, value)
			}

			return (upper - lower) + float32(2.0*shape.polygon.Radius)
		}

	case SegmentShape:
		{
			value1 := shape.segment.Point1.Dot(line)
			value2 := shape.segment.Point2.Dot(line)

			return absFloat(value2 - value1)
		}

	case ChainSegmentShape:
		{
			value1 := shape.chainSegment.Segment.Point1.Dot(line)
			value2 := shape.chainSegment.Segment.Point2.Dot(line)

			return absFloat(value2 - value1)
		}

	case ShapeTypeCount:
		return 0.0
	default:
		return 0.0
	}
}

func computeShapeMass(shape *shape) MassData {
	switch shape.shapeType {
	case CapsuleShape:
		return ComputeCapsuleMass(&shape.capsule, shape.density)
	case CircleShape:
		return ComputeCircleMass(&shape.circle, shape.density)
	case PolygonShape:
		return ComputePolygonMass(&shape.polygon, shape.density)
	case SegmentShape, ChainSegmentShape, ShapeTypeCount:
		return MassData{} // Edges and the count sentinel have no area mass.
	default:
		return MassData{}
	}
}

func computeShapeExtent(shape *shape, localCenter Vec2) shapeExtent {
	extent := shapeExtent{}

	switch shape.shapeType {
	case CapsuleShape:
		{
			radius := shape.capsule.Radius
			extent.minExtent = radius
			c1 := shape.capsule.Center1.Sub(localCenter)
			c2 := shape.capsule.Center2.Sub(localCenter)
			extent.maxExtent = float32(
				math.Sqrt(float64(maxFloat(c1.LengthSquared(), c2.LengthSquared()))),
			) + radius
		}

	case CircleShape:
		{
			radius := shape.circle.Radius
			extent.minExtent = radius
			extent.maxExtent = shape.circle.Center.Sub(localCenter).Length() + radius
		}

	case PolygonShape:
		{
			poly := &shape.polygon
			minExtent := Huge()
			maxExtentSqr := float32(0.0)

			count := poly.Count
			for i := range count {
				v := poly.Vertices[i]
				planeOffset := poly.Normals[i].Dot(v.Sub(poly.Centroid))
				minExtent = minFloat(minExtent, planeOffset)

				distanceSqr := v.Sub(localCenter).LengthSquared()
				maxExtentSqr = maxFloat(maxExtentSqr, distanceSqr)
			}

			extent.minExtent = minExtent + poly.Radius
			extent.maxExtent = float32(math.Sqrt(float64(maxExtentSqr))) + poly.Radius
		}

	case SegmentShape:
		{
			extent.minExtent = 0.0
			c1 := shape.segment.Point1.Sub(localCenter)
			c2 := shape.segment.Point2.Sub(localCenter)
			extent.maxExtent = float32(
				math.Sqrt(float64(maxFloat(c1.LengthSquared(), c2.LengthSquared()))),
			)
		}

	case ChainSegmentShape:
		{
			extent.minExtent = 0.0
			c1 := shape.chainSegment.Segment.Point1.Sub(localCenter)
			c2 := shape.chainSegment.Segment.Point2.Sub(localCenter)
			extent.maxExtent = float32(
				math.Sqrt(float64(maxFloat(c1.LengthSquared(), c2.LengthSquared()))),
			)
		}

	case ShapeTypeCount:
		// The count sentinel has no extent.
	default:
		break
	}

	return extent
}

func rayCastShape(input *RayCastInput, shape *shape, transform Transform) CastOutput {
	localInput := *input
	localInput.Origin = transform.InvApply(input.Origin)
	localInput.Translation = transform.Q.InvRotate(input.Translation)

	output := CastOutput{}

	switch shape.shapeType {
	case CapsuleShape:
		output = RayCastCapsule(&shape.capsule, &localInput)
	case CircleShape:
		output = RayCastCircle(&shape.circle, &localInput)
	case PolygonShape:
		output = RayCastPolygon(&shape.polygon, &localInput)
	case SegmentShape:
		output = RayCastSegment(&shape.segment, &localInput, false)
	case ChainSegmentShape:
		output = RayCastSegment(&shape.chainSegment.Segment, &localInput, true)
	case ShapeTypeCount:
		return output
	default:
		return output
	}

	// The output point stays in the frame of the input transform, a caller chosen frame that is
	// typically re-centered near the origin.
	output.Point = transform.Apply(output.Point)
	output.Normal = transform.Q.Rotate(output.Normal)

	return output
}

func shapeCastShape(input *ShapeCastInput, shape *shape, transform Transform) CastOutput {
	output := CastOutput{}

	if input.Proxy.Count == 0 {
		return output
	}

	localInput := *input

	for i := range localInput.Proxy.Count {
		localInput.Proxy.Points[i] = transform.InvApply(input.Proxy.Points[i])
	}

	localInput.Translation = transform.Q.InvRotate(input.Translation)

	switch shape.shapeType {
	case CapsuleShape:
		output = ShapeCastCapsule(&shape.capsule, &localInput)
	case CircleShape:
		output = ShapeCastCircle(&shape.circle, &localInput)
	case PolygonShape:
		output = ShapeCastPolygon(&shape.polygon, &localInput)
	case SegmentShape:
		output = ShapeCastSegment(&shape.segment, &localInput)
	case ChainSegmentShape:
		{
			// Check for back side collision
			approximateCentroid := localInput.Proxy.Points[0]
			for i := 1; i < localInput.Proxy.Count; i++ {
				approximateCentroid = approximateCentroid.Add(localInput.Proxy.Points[i])
			}

			approximateCentroid = approximateCentroid.Scale(1.0 / float32(localInput.Proxy.Count))

			edge := shape.chainSegment.Segment.Point2.Sub(shape.chainSegment.Segment.Point1)
			r := approximateCentroid.Sub(shape.chainSegment.Segment.Point1)

			if r.Cross(edge) < 0.0 {
				// Shape cast starts behind
				return output
			}

			output = ShapeCastSegment(&shape.chainSegment.Segment, &localInput)
		}

	case ShapeTypeCount:
		return output
	default:
		return output
	}

	// Same frame contract as b2RayCastShape, the point stays in the input transform frame
	output.Point = transform.Apply(output.Point)
	output.Normal = transform.Q.Rotate(output.Normal)

	return output
}

func collideMover(mover *Capsule, shape *shape, transform Transform) PlaneResult {
	var localMover Capsule

	localMover.Center1 = transform.InvApply(mover.Center1)
	localMover.Center2 = transform.InvApply(mover.Center2)
	localMover.Radius = mover.Radius

	result := PlaneResult{}

	switch shape.shapeType {
	case CapsuleShape:
		result = collideMoverAndCapsule(&localMover, &shape.capsule)
	case CircleShape:
		result = collideMoverAndCircle(&localMover, &shape.circle)
	case PolygonShape:
		result = collideMoverAndPolygon(&localMover, &shape.polygon)
	case SegmentShape:
		result = collideMoverAndSegment(&localMover, &shape.segment)
	case ChainSegmentShape:
		result = collideMoverAndSegment(&localMover, &shape.chainSegment.Segment)
	case ShapeTypeCount:
		return result
	default:
		return result
	}

	if !result.Hit {
		return result
	}

	result.Plane.Normal = transform.Q.Rotate(result.Plane.Normal)

	return result
}

func createShapeProxy(
	world *World,
	shape *shape,
	bodyType BodyType,
	transform WorldTransform,
	forcePairCreation bool,
) {
	core.Assert(shape.proxyKey == nullIndex, "shape.proxyKey == nullIndex")

	fatAABB := &world.fatAABBs[shape.id]
	updateShapeAABBs(shape, fatAABB, transform, bodyType)

	// Create proxies in the broad-phase.
	shape.proxyKey = world.broadPhase.createProxy(
		bodyType,
		*fatAABB,
		shape.filter.CategoryBits,
		shape.id,
		forcePairCreation,
	)
	core.Assert(
		proxyType(shape.proxyKey) < BodyTypeCount,
		"proxyType( shape.proxyKey ) < BodyTypeCount",
	)
}

func destroyShapeProxy(shape *shape, bp *broadPhase) {
	if shape.proxyKey != nullIndex {
		bp.destroyProxy(shape.proxyKey)
		shape.proxyKey = nullIndex
	}
}

func makeShapeDistanceProxy(shape *shape) ShapeProxy {
	switch shape.shapeType {
	case CapsuleShape:
		return MakeProxy([]Vec2{shape.capsule.Center1, shape.capsule.Center2}, shape.capsule.Radius)
	case CircleShape:
		return MakeProxy([]Vec2{shape.circle.Center}, shape.circle.Radius)
	case PolygonShape:
		return MakeProxy(shape.polygon.Vertices[:shape.polygon.Count], shape.polygon.Radius)
	case SegmentShape:
		return MakeProxy([]Vec2{shape.segment.Point1, shape.segment.Point2}, 0)
	case ChainSegmentShape:
		return MakeProxy(
			[]Vec2{shape.chainSegment.Segment.Point1, shape.chainSegment.Segment.Point2},
			0,
		)
	case ShapeTypeCount:
		fallthrough // The count sentinel uses the invalid-geometry path.
	default:
		{
			core.Assert(false, "false")

			empty := ShapeProxy{}

			return empty
		}
	}
}

// Get the id of the body that a shape is attached to.
func (s Shape) Body() Body {
	world := s.world
	shape := getShape(world, s)

	return makeBodyID(world, shape.bodyID)
}

// Get the world that owns this shape.
func (s Shape) World() *World {
	return s.world
}

// Set the user data for a shape.
func (s Shape) SetUserData(userData any) {
	world := s.world
	shape := getShape(world, s)
	shape.userData = userData
}

// Get the user data for a shape. This is useful when you get a shape id
// from an event or query.
func (s Shape) UserData() any {
	world := s.world
	shape := getShape(world, s)

	return shape.userData
}

// Returns true if the shape is a sensor. It is not possible to change a shape
// from sensor to solid dynamically because this breaks the contract for
// sensor events.
func (s Shape) IsSensor() bool {
	world := s.world
	shape := getShape(world, s)

	return shape.sensorIndex != nullIndex
}

// Test a world point for overlap with a shape.
func (s Shape) TestPoint(point Pos) bool {
	world := s.world
	shape := getShape(world, s)

	transform := getBodyTransform(world, shape.bodyID)
	localPoint := transform.InvApply(point)

	var result bool

	switch shape.shapeType {
	case CapsuleShape:
		result = PointInCapsule(&shape.capsule, localPoint)

	case CircleShape:
		result = PointInCircle(&shape.circle, localPoint)

	case PolygonShape:
		result = PointInPolygon(&shape.polygon, localPoint)

	case SegmentShape, ChainSegmentShape, ShapeTypeCount:
		result = false // Edges and the count sentinel contain no area.
	default:
		result = false
	}

	if world.recording != nil {
		recBuf := recBuffer{}
		recWSHAPEID(&recBuf, s)
		recWPOSITION(&recBuf, point)
		recWBOOL(&recBuf, result)
		recCommitRecord(world.recording, 0xE7, recBuf.data, recBuf.size)
		recBufFree(&recBuf)
	}

	return result
}

// Ray cast a single shape. The ray starts at the origin and extends by the translation. The output
// point is a world position.
func (s Shape) RayCast(origin Pos, translation Vec2) WorldCastOutput {
	core.Assert(origin.IsValid(), "IsValidPosition( origin )")
	core.Assert(translation.IsValid(), "IsValidVec2( translation )")

	world := s.world
	shape := getShape(world, s)

	// Re-center on the origin so the cast runs in float precision
	transform := getBodyTransform(world, shape.bodyID).Relative(origin)

	// The ray starts at the origin, so its origin in the re-centered frame is zero
	input := RayCastInput{Vec2{}, translation, 1.0}

	// Lift the re-centered float result back to a world position
	local := rayCastShape(&input, shape, transform)

	var output WorldCastOutput

	output.Normal = local.Normal
	output.Point = origin.Add(local.Point)
	output.Fraction = local.Fraction
	output.Iterations = local.Iterations
	output.Hit = local.Hit

	if world.recording != nil {
		recBuf := recBuffer{}
		recWSHAPEID(&recBuf, s)
		recWPOSITION(&recBuf, origin)
		recWVEC2(&recBuf, translation)
		recWWORLDCASTOUTPUT(&recBuf, output)
		recCommitRecord(world.recording, 0xE8, recBuf.data, recBuf.size)
		recBufFree(&recBuf)
	}

	return output
}

// Set the mass density of a shape, usually in kg/m^2.
// This will optionally update the mass properties on the parent body.
// See: ShapeDef.Density, Body.UpdateMassFromShapes.
func (s Shape) SetDensity(density float32, updateBodyMass bool) {
	checkInput(IsValidFloat(density) && density >= 0.0, "IsValidFloat( density ) && density >= 0.0")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeSetDensity(
			world.recording,
			&recArgsShapeSetDensity{shape: s, density: density, updateBodyMass: updateBodyMass},
		)
	}

	shape := getShape(world, s)
	if density == shape.density {
		// early return to avoid expensive function
		return
	}

	shape.density = density

	if updateBodyMass {
		body := &world.bodies[shape.bodyID]
		updateBodyMassData(world, body)
	}
}

// Get the density of a shape, usually in kg/m^2.
func (s Shape) Density() float32 {
	world := s.world
	shape := getShape(world, s)

	return shape.density
}

// Set the friction on a shape.
func (s Shape) SetFriction(friction float32) {
	checkInput(
		IsValidFloat(friction) && friction >= 0.0,
		"IsValidFloat( friction ) && friction >= 0.0",
	)

	world := s.world
	core.Assert(!world.locked, "world.locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteShapeSetFriction(
			world.recording,
			&recArgsShapeSetFriction{shape: s, friction: friction},
		)
	}

	shape := getShape(world, s)
	shape.material.Friction = friction
}

// Get the friction of a shape.
func (s Shape) Friction() float32 {
	world := s.world
	shape := getShape(world, s)

	return shape.material.Friction
}

// Set the shape restitution (bounciness).
func (s Shape) SetRestitution(restitution float32) {
	checkInput(
		IsValidFloat(restitution) && restitution >= 0.0,
		"IsValidFloat( restitution ) && restitution >= 0.0",
	)

	world := s.world
	core.Assert(!world.locked, "world.locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteShapeSetRestitution(
			world.recording,
			&recArgsShapeSetRestitution{shape: s, restitution: restitution},
		)
	}

	shape := getShape(world, s)
	shape.material.Restitution = restitution
}

// Get the shape restitution.
func (s Shape) Restitution() float32 {
	world := s.world
	shape := getShape(world, s)

	return shape.material.Restitution
}

// Set the user material identifier.
func (s Shape) SetUserMaterial(material uint64) {
	world := s.world
	core.Assert(!world.locked, "world.locked == false")

	if world.locked {
		return
	}

	if world.recording != nil {
		recWriteShapeSetUserMaterial(
			world.recording,
			&recArgsShapeSetUserMaterial{shape: s, material: material},
		)
	}

	shape := getShape(world, s)
	shape.material.UserMaterialID = material
}

// Get the user material identifier.
func (s Shape) UserMaterial() uint64 {
	world := s.world
	shape := getShape(world, s)

	return shape.material.UserMaterialID
}

// Get the shape surface material.
func (s Shape) SurfaceMaterial() SurfaceMaterial {
	world := s.world
	shape := getShape(world, s)

	return shape.material
}

// Set the shape surface material.
func (s Shape) SetSurfaceMaterial(surfaceMaterial *SurfaceMaterial) {
	checkInput(isValidMaterial(surfaceMaterial), "isValidMaterial( surfaceMaterial )")

	world := s.world
	if world.recording != nil {
		recWriteShapeSetSurfaceMaterial(
			world.recording,
			&recArgsShapeSetSurfaceMaterial{shape: s, material: *surfaceMaterial},
		)
	}

	shape := getShape(world, s)
	shape.material = *surfaceMaterial
}

// Get the shape filter.
func (s Shape) Filter() Filter {
	world := s.world
	shape := getShape(world, s)

	return shape.filter
}

func resetProxy(world *World, shape *shape, destroyProxy bool) {
	body := &world.bodies[shape.bodyID]

	shapeID := shape.id

	// destroy all contacts associated with this shape
	contactKey := body.headContactKey
	for contactKey != nullIndex {
		contactID := contactKey >> 1
		edgeIndex := contactKey & 1

		contact := &world.contacts[contactID]
		contactKey = contact.edges[edgeIndex].nextKey

		if contact.shapeIDA == shapeID || contact.shapeIDB == shapeID {
			destroyContact(world, contact)
		}
	}

	transform := getBodyTransformQuick(world, body)

	if shape.proxyKey != nullIndex {
		proxyType := proxyType(shape.proxyKey)
		fatAABB := &world.fatAABBs[shapeID]
		updateShapeAABBs(shape, fatAABB, transform, proxyType)

		if destroyProxy {
			world.broadPhase.destroyProxy(shape.proxyKey)

			forcePairCreation := true
			shape.proxyKey = world.broadPhase.createProxy(
				proxyType,
				*fatAABB,
				shape.filter.CategoryBits,
				shapeID,
				forcePairCreation,
			)
		} else {
			world.broadPhase.moveProxy(shape.proxyKey, *fatAABB)
		}
	} else {
		proxyType := body.bodyType
		updateShapeAABBs(shape, &world.fatAABBs[shapeID], transform, proxyType)
	}

	validateSolverSets(world)
}

// Set the current filter. This is almost as expensive as recreating the shape. This may cause
// contacts to be immediately destroyed. However contacts are not created until the next world step.
// Sensor overlap state is also not updated until the next world step.
// See: ShapeDef.Filter.
func (s Shape) SetFilter(filter Filter) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeSetFilter(world.recording, &recArgsShapeSetFilter{shape: s, filter: filter})
	}

	shape := getShape(world, s)
	if filter.MaskBits == shape.filter.MaskBits &&
		filter.CategoryBits == shape.filter.CategoryBits &&
		filter.GroupIndex == shape.filter.GroupIndex {
		return
	}

	// If the category bits change, I need to destroy the proxy because it affects the tree sorting.
	destroyProxy := filter.CategoryBits != shape.filter.CategoryBits

	shape.filter = filter

	resetProxy(world, shape, destroyProxy)

	// note: this does not immediately update sensor overlaps. Instead sensor
	// overlaps are updated the next time step
}

// Enable sensor events for this shape.
// See: ShapeDef.EnableSensorEvents.
func (s Shape) SetSensorEventsEnabled(flag bool) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeEnableSensorEvents(
			world.recording,
			&recArgsShapeEnableSensorEvents{shape: s, flag: flag},
		)
	}

	shape := getShape(world, s)
	shape.enableSensorEvents = flag
}

// Returns true if sensor events are enabled.
func (s Shape) SensorEventsEnabled() bool {
	world := s.world
	shape := getShape(world, s)

	return shape.enableSensorEvents
}

// Enable contact events for this shape. Only applies to kinematic and dynamic bodies. Ignored for sensors.
// See: ShapeDef.EnableContactEvents
// Warning: changing this at run-time may lead to lost begin/end events.
func (s Shape) SetContactEventsEnabled(flag bool) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeEnableContactEvents(
			world.recording,
			&recArgsShapeEnableContactEvents{shape: s, flag: flag},
		)
	}

	shape := getShape(world, s)
	shape.enableContactEvents = flag
}

// Returns true if contact events are enabled.
func (s Shape) ContactEventsEnabled() bool {
	world := s.world
	shape := getShape(world, s)

	return shape.enableContactEvents
}

// Enable pre-solve contact events for this shape. Only applies to dynamic bodies. These are expensive
// and must be carefully handled due to multithreading. Ignored for sensors.
// See: PreSolveFunc.
func (s Shape) SetPreSolveEventsEnabled(flag bool) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeEnablePreSolveEvents(
			world.recording,
			&recArgsShapeEnablePreSolveEvents{shape: s, flag: flag},
		)
	}

	shape := getShape(world, s)
	shape.enablePreSolveEvents = flag
}

// Returns true if pre-solve events are enabled.
func (s Shape) PreSolveEventsEnabled() bool {
	world := s.world
	shape := getShape(world, s)

	return shape.enablePreSolveEvents
}

// Enable contact hit events for this shape. Ignored for sensors.
// See: WorldDef.hitEventThreshold.
func (s Shape) SetHitEventsEnabled(flag bool) {
	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeEnableHitEvents(
			world.recording,
			&recArgsShapeEnableHitEvents{shape: s, flag: flag},
		)
	}

	shape := getShape(world, s)
	shape.enableHitEvents = flag
}

// Returns true if hit events are enabled.
func (s Shape) HitEventsEnabled() bool {
	world := s.world
	shape := getShape(world, s)

	return shape.enableHitEvents
}

// Get the type of a shape.
func (s Shape) Type() ShapeType {
	world := s.world
	shape := getShape(world, s)

	return shape.shapeType
}

// Get a copy of the shape's circle. Asserts the type is correct.
func (s Shape) Circle() Circle {
	world := s.world
	shape := getShape(world, s)
	core.Assert(shape.shapeType == CircleShape, "shape.shapeType == CircleShape")

	return shape.circle
}

// Get a copy of the shape's line segment. Asserts the type is correct.
func (s Shape) Segment() Segment {
	world := s.world
	shape := getShape(world, s)
	core.Assert(shape.shapeType == SegmentShape, "shape.shapeType == SegmentShape")

	return shape.segment
}

// Get a copy of the shape's chain segment. These come from chain shapes.
// Asserts the type is correct.
func (s Shape) ChainSegment() ChainSegment {
	world := s.world
	shape := getShape(world, s)
	core.Assert(shape.shapeType == ChainSegmentShape, "shape.shapeType == ChainSegmentShape")

	return shape.chainSegment
}

// Get a copy of the shape's capsule. Asserts the type is correct.
func (s Shape) Capsule() Capsule {
	world := s.world
	shape := getShape(world, s)
	core.Assert(shape.shapeType == CapsuleShape, "shape.shapeType == CapsuleShape")

	return shape.capsule
}

// Get a copy of the shape's convex polygon. Asserts the type is correct.
func (s Shape) Polygon() Polygon {
	world := s.world
	shape := getShape(world, s)
	core.Assert(shape.shapeType == PolygonShape, "shape.shapeType == PolygonShape")

	return shape.polygon
}

// Allows you to change a shape to be a circle or update the current circle.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes.
func (s Shape) SetCircle(circle *Circle) {
	checkInput(isValidCircle(circle), "isValidCircle( circle )")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeSetCircle(world.recording, &recArgsShapeSetCircle{shape: s, circle: *circle})
	}

	shape := getShape(world, s)
	shape.circle = *circle
	shape.shapeType = CircleShape
	shape.aabbMargin = computeShapeMargin(shape)

	destroyProxy := true
	resetProxy(world, shape, destroyProxy)
}

// Allows you to change a shape to be a capsule or update the current capsule.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes.
func (s Shape) SetCapsule(capsule *Capsule) {
	checkInput(isValidCapsule(capsule), "isValidCapsule( capsule )")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	lengthSqr := capsule.Center1.DistanceSquared(capsule.Center2)
	if lengthSqr <= LinearSlop()*LinearSlop() {
		return
	}

	if world.recording != nil {
		recWriteShapeSetCapsule(
			world.recording,
			&recArgsShapeSetCapsule{shape: s, capsule: *capsule},
		)
	}

	shape := getShape(world, s)
	shape.capsule = *capsule
	shape.shapeType = CapsuleShape
	shape.aabbMargin = computeShapeMargin(shape)

	destroyProxy := true
	resetProxy(world, shape, destroyProxy)
}

// Allows you to change a shape to be a segment or update the current segment.
func (s Shape) SetSegment(segment *Segment) {
	checkInput(isValidSegment(segment), "isValidSegment( segment )")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeSetSegment(
			world.recording,
			&recArgsShapeSetSegment{shape: s, segment: *segment},
		)
	}

	shape := getShape(world, s)
	shape.segment = *segment
	shape.shapeType = SegmentShape
	shape.aabbMargin = computeShapeMargin(shape)

	destroyProxy := true
	resetProxy(world, shape, destroyProxy)
}

// Allows you to change a shape to be a polygon or update the current polygon.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes.
func (s Shape) SetPolygon(polygon *Polygon) {
	checkInput(isValidPolygon(polygon), "isValidPolygon( polygon )")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeSetPolygon(
			world.recording,
			&recArgsShapeSetPolygon{shape: s, polygon: *polygon},
		)
	}

	shape := getShape(world, s)
	shape.polygon = *polygon
	shape.shapeType = PolygonShape
	shape.aabbMargin = computeShapeMargin(shape)

	destroyProxy := true
	resetProxy(world, shape, destroyProxy)
}

// Allows you to change a shape to be an orphaned chain segment or update the current chain
// segment, including its ghost vertices. The chainId on the input is ignored. The resulting
// shape is always orphaned. Asserts if the shape is already a chain segment
// owned by a chainShape (chainId != B2_NULL_INDEX).
func (s Shape) SetChainSegment(chainSegment *ChainSegment) {
	checkInput(isValidChainSegment(chainSegment), "isValidChainSegment( chainSegment )")

	world := getWorldLocked(s.world)
	if world == nil {
		return
	}

	shape := getShape(world, s)

	// Cannot modify a chain segment that has a parent chain shape
	if shape.shapeType == ChainSegmentShape && shape.chainSegment.ChainID != nullIndex {
		core.Assert(false, "false")
		return
	}

	lengthSqr := chainSegment.Segment.Point1.DistanceSquared(chainSegment.Segment.Point2)
	if lengthSqr <= LinearSlop()*LinearSlop() {
		return
	}

	if world.recording != nil {
		recWriteShapeSetChainSegment(
			world.recording,
			&recArgsShapeSetChainSegment{shape: s, chainSegment: *chainSegment},
		)
	}

	shape.chainSegment = *chainSegment
	shape.chainSegment.ChainID = nullIndex
	shape.shapeType = ChainSegmentShape
	shape.aabbMargin = computeShapeMargin(shape)

	destroyProxy := true
	resetProxy(world, shape, destroyProxy)
}

// Get the parent chain id if the shape type is a chain segment, otherwise
// returns Chain{}.
func (s Shape) ParentChain() Chain {
	world := s.world

	shape := getShape(world, s)
	if shape.shapeType == ChainSegmentShape {
		chainID := shape.chainSegment.ChainID
		if chainID != nullIndex {
			chain := &world.chainShapes[chainID]
			id := Chain{index1: int32(chainID + 1), world: s.world, generation: chain.generation}

			return id
		}
	}

	return Chain{}
}

// Get the maximum capacity required for retrieving all the touching contacts on a shape.
func (s Shape) ContactCapacity() int {
	world := getWorldLocked(s.world)
	if world == nil {
		return 0
	}

	shape := getShape(world, s)
	if shape.sensorIndex != nullIndex {
		return 0
	}

	body := &world.bodies[shape.bodyID]

	// Conservative and fast
	return body.contactCount
}

// Get the touching contact data for a shape. The provided shapeId will be either shapeIdA or
// shapeIdB on the contact data.
// Note: Box2D uses speculative collision so some contact points may be separated.
func (s Shape) ContactData() iter.Seq[ContactData] {
	world := getWorldLocked(s.world)
	if world == nil {
		return func(yield func(ContactData) bool) {}
	}

	shape := getShape(world, s)
	if shape.sensorIndex != nullIndex {
		return func(yield func(ContactData) bool) {}
	}

	body := &world.bodies[shape.bodyID]

	return func(yield func(ContactData) bool) {
		contactKey := body.headContactKey
		for contactKey != nullIndex {
			contactID := contactKey >> 1
			edgeIndex := contactKey & 1
			contact := &world.contacts[contactID]
			// Does contact involve this shape and is it touching?
			if (contact.shapeIDA == int(s.index1)-1 || contact.shapeIDB == int(s.index1)-1) &&
				(contact.flags&contactTouchingFlag) != 0 {
				shapeA := &world.shapes[contact.shapeIDA]
				shapeB := &world.shapes[contact.shapeIDB]
				contactSim := getContactSim(world, contact)

				data := ContactData{
					Contact: Contact{
						index1:     int32(contact.contactID + 1),
						world:      s.world,
						generation: contact.generation,
					},
					ShapeA: Shape{
						index1:     int32(shapeA.id + 1),
						world:      s.world,
						generation: shapeA.generation,
					},
					ShapeB: Shape{
						index1:     int32(shapeB.id + 1),
						world:      s.world,
						generation: shapeB.generation,
					},
					Manifold: contactSim.manifold,
				}
				if !yield(data) {
					return
				}
			}

			contactKey = contact.edges[edgeIndex].nextKey
		}
	}
}

// Get the maximum capacity required for retrieving all the overlapped shapes on a sensor shape.
// This returns 0 if the provided shape is not a sensor.
// shapeId: the id of a sensor shape
// Returns: the required capacity to get all the overlaps in Shape.SensorData.
func (s Shape) SensorCapacity() int {
	world := getWorldLocked(s.world)
	if world == nil {
		return 0
	}

	shape := getShape(world, s)
	if shape.sensorIndex == nullIndex {
		return 0
	}

	sensor := &world.sensors[shape.sensorIndex]

	return len(sensor.overlaps2)
}

// Get the overlap data for a sensor shape computed the previous world step.
// shapeId: the id of a sensor shape
// Warning: overlaps may contain destroyed shapes so use shape_IsValid to confirm each overlap.
func (s Shape) SensorData() iter.Seq[Shape] {
	world := getWorldLocked(s.world)
	if world == nil {
		return func(yield func(Shape) bool) {}
	}

	shape := getShape(world, s)
	if shape.sensorIndex == nullIndex {
		return func(yield func(Shape) bool) {}
	}

	sensor := &world.sensors[shape.sensorIndex]

	return func(yield func(Shape) bool) {
		for _, ref := range sensor.overlaps2 {
			visitorID := Shape{
				index1:     int32(ref.shapeID + 1),
				world:      s.world,
				generation: ref.generation,
			}
			if !yield(visitorID) {
				return
			}
		}
	}
}

// Get the current world AABB.
func (s Shape) AABB() AABB {
	world := s.world
	if world == nil {
		return AABB{}
	}

	shape := getShape(world, s)

	return shape.aabb
}

// Compute the mass data for a shape.
func (s Shape) ComputeMassData() MassData {
	world := s.world
	if world == nil {
		return MassData{}
	}

	shape := getShape(world, s)

	return computeShapeMass(shape)
}

// Get the closest point on a shape to a target point. Target and result are world positions.
// todo need sample.
func (s Shape) ClosestPoint(target Pos) Pos {
	world := s.world
	if world == nil {
		return Pos{}
	}

	shape := getShape(world, s)
	body := &world.bodies[shape.bodyID]
	transform := getBodyTransformQuick(world, body)

	// The target rides in as the frame of proxy B, so the relative pose is differenced in double
	// and the result stays exact far from the origin
	zero := Vec2{}
	targetTransform := WorldTransform{target, Rot{C: 1}}

	var input DistanceInput

	input.ProxyA = makeShapeDistanceProxy(shape)
	input.ProxyB = MakeProxy([]Vec2{zero}, 0)
	input.Transform = transform.InvMul(targetTransform)
	input.UseRadii = true

	cache := SimplexCache{}
	output := ShapeDistance(&input, &cache, nil)

	return transform.Apply(output.PointA)
}

// Apply a wind force to the body for this shape using the density of air. This considers
// the projected area of the shape in the wind direction. This also considers
// the relative velocity of the shape.
// shapeId: the shape id
// wind: the wind velocity in world space
// drag: the drag coefficient, the force that opposes the relative velocity
// lift: the lift coefficient, the force that is perpendicular to the relative velocity
// wake: should this wake the body
func (s Shape) ApplyWind(wind Vec2, drag float32, lift float32, wake bool) {
	checkInput(wind.IsValid(), "IsValidVec2( wind )")
	checkInput(IsValidFloat(drag), "IsValidFloat( drag )")
	checkInput(IsValidFloat(lift), "IsValidFloat( lift )")

	world := s.world
	if world == nil {
		return
	}

	if world.recording != nil {
		recWriteShapeApplyWind(
			world.recording,
			&recArgsShapeApplyWind{shape: s, wind: wind, drag: drag, lift: lift, wake: wake},
		)
	}

	shape := getShape(world, s)

	shapeType := shape.shapeType
	if shapeType != CircleShape && shapeType != CapsuleShape && shapeType != PolygonShape {
		return
	}

	body := &world.bodies[shape.bodyID]

	if body.bodyType != DynamicBody || body.setIndex == disabledSet {
		return
	}

	if body.setIndex >= firstSleepingSet && !wake {
		return
	}

	sim := getBodySim(world, body)

	if body.setIndex != awakeSet {
		// Must wake for state to exist
		wakeBody(world, body)
	}

	core.Assert(body.setIndex == awakeSet, "body.setIndex == awakeSet")

	state := getBodyState(world, body)

	lengthUnits := LengthUnitsPerMeter()
	volumeUnits := lengthUnits * lengthUnits * lengthUnits

	// In 2D I'm assuming unit depth
	airDensity := 1.2250 / volumeUnits
	force, torque := computeShapeWindForce(shape, sim, state, wind, drag, lift, airDensity)

	sim.force = sim.force.Add(force)
	sim.torque += torque
}

func computeShapeWindForce(
	shape *shape,
	sim *bodySim,
	state *bodyState,
	wind Vec2,
	drag float32, lift float32, airDensity float32,
) (Vec2, float32) {
	transform := sim.transform
	force := Vec2{}
	torque := float32(0.0)

	switch shape.shapeType {
	case CircleShape:
		{
			radius := shape.circle.Radius
			centroid := shape.localCentroid
			lever := transform.Q.Rotate(centroid.Sub(sim.localCenter))
			shapeVelocity := state.linearVelocity.Add(lever.ScalarCross(state.angularVelocity))
			relativeVelocity := wind.MulSub(drag, shapeVelocity)
			speed, direction := relativeVelocity.LengthAndNormalize()
			projectedArea := 2.0 * radius
			force = direction.Scale(0.5 * airDensity * projectedArea * speed * speed)
			torque = lever.Cross(force)
		}

	case CapsuleShape:
		{
			centroid := shape.localCentroid
			lever := transform.Q.Rotate(centroid.Sub(sim.localCenter))
			shapeVelocity := state.linearVelocity.Add(lever.ScalarCross(state.angularVelocity))
			relativeVelocity := wind.MulSub(drag, shapeVelocity)
			speed, direction := relativeVelocity.LengthAndNormalize()

			d := shape.capsule.Center2.Sub(shape.capsule.Center1)
			d = transform.Q.Rotate(d)

			radius := shape.capsule.Radius
			projectedArea := float32(2.0*radius) + absFloat(d.Cross(direction))

			// Normal that opposes the wind
			normal := d.Normalize().LeftPerp()
			if normal.Dot(direction) > 0 {
				normal = normal.Neg()
			}

			// portion of wind that is perpendicular to surface
			liftDirection := direction.ScalarCross(normal.Cross(direction))

			forceMagnitude := 0.5 * airDensity * projectedArea * speed * speed
			force = direction.MulAdd(lift, liftDirection).Scale(forceMagnitude)

			edgeLever := lever.MulAdd(radius, normal)
			torque = edgeLever.Cross(force)
		}

	case PolygonShape:
		{
			centroid := shape.localCentroid
			lever := transform.Q.Rotate(centroid.Sub(sim.localCenter))
			shapeVelocity := state.linearVelocity.Add(lever.ScalarCross(state.angularVelocity))
			relativeVelocity := wind.MulSub(drag, shapeVelocity)
			speed, direction := relativeVelocity.LengthAndNormalize()

			// polygon radius is ignored for simplicity
			count := shape.polygon.Count
			vertices := shape.polygon.Vertices

			v1 := vertices[count-1]
			for i := range count {
				v2 := vertices[i]
				d := v2.Sub(v1)
				edgeCenter := v1.Lerp(v2, 0.5)
				v1 = v2

				d = transform.Q.Rotate(d)

				projectedArea := d.Cross(direction)
				if projectedArea <= 0.0 {
					// back facing
					continue
				}

				normal := d.Normalize().RightPerp()

				// portion of wind that is perpendicular to surface
				liftDirection := direction.ScalarCross(normal.Cross(direction))

				forceMagnitude := 0.5 * airDensity * projectedArea * speed * speed
				f := direction.MulAdd(lift, liftDirection).Scale(forceMagnitude)

				edgeLever := transform.Q.Rotate(edgeCenter.Sub(sim.localCenter))

				force = force.Add(f)
				torque += edgeLever.Cross(f)
			}
		}

	case SegmentShape, ChainSegmentShape, ShapeTypeCount:
		// Edges and the count sentinel present no wind-catching area.
	default:
		break
	}

	return force, torque
}

func getShapeRadius(shape *shape) float32 {
	switch shape.shapeType {
	case CapsuleShape:
		return shape.capsule.Radius
	case CircleShape:
		return shape.circle.Radius
	case PolygonShape:
		return shape.polygon.Radius
	case SegmentShape, ChainSegmentShape, ShapeTypeCount:
		return 0.0 // Edges and the count sentinel have no radius.
	default:
		return 0.0
	}
}

func shouldShapesCollide(filterA Filter, filterB Filter) bool {
	if filterA.GroupIndex == filterB.GroupIndex && filterA.GroupIndex != 0 {
		return filterA.GroupIndex > 0
	}

	return (filterA.MaskBits&filterB.CategoryBits) != 0 &&
		(filterA.CategoryBits&filterB.MaskBits) != 0
}

func shouldQueryCollide(shapeFilter Filter, queryFilter QueryFilter) bool {
	return (shapeFilter.CategoryBits&queryFilter.MaskBits) != 0 &&
		(shapeFilter.MaskBits&queryFilter.CategoryBits) != 0
}
