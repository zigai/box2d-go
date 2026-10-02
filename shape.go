// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "iter"

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

// Shape type
type ShapeType int

const (
	// A circle with an offset
	CircleShape ShapeType = 0
	// A capsule is an extruded circle
	CapsuleShape ShapeType = 1
	// A line segment
	SegmentShape ShapeType = 2
	// A convex polygon
	PolygonShape ShapeType = 3
	// A line segment owned by a chain shape
	ChainSegmentShape ShapeType = 4
	// The number of shape types
	ShapeTypeCount ShapeType = 5
)

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
	// See: EnableSensorEvents
	IsSensor bool
	// Enable sensor events for this shape. This applies to sensors and non-sensors. Both shapes involved must have this flag set
	// to true. False by default, even for sensors.
	EnableSensorEvents bool
	// Enable contact events for this shape. Only applies to kinematic and dynamic bodies. Only one shape involved needs this
	// flag set to true. Ignored for sensors. False by default.
	EnableContactEvents bool
	// Enable hit events for this shape. Only applies to kinematic and dynamic bodies. Only one shape involved needs this flag
	// set to true. Ignored for sensors. False by default.
	EnableHitEvents bool
	// Enable pre-solve contact events for this shape. Only applies to dynamic bodies. These are expensive
	// and must be carefully handled due to multithreading. Ignored for sensors.
	EnablePreSolveEvents bool
	// When shapes are created they will scan the environment for collision the next time step. This can significantly slow down
	// static body creation when there are many static shapes.
	// This is flag is ignored for dynamic and kinematic shapes which always invoke contact creation.
	InvokeContactCreation bool
	// Should the body update the mass properties when this shape is created. Default is true.
	// Warning: if this is false, you MUST call Body.UpdateMassFromShapes or Body.SetMassData before simulating the world.
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
	// the last and first point form a segment. Cloned.
	Points []Vec2
	// The point count. At least 2 for an open chain and at least 3 for a loop.
	// SegmentCount = PointCount for loops, PointCount - 1 for open chains
	PointCount int
	// Leading ghost point for an open chain. Distance from first point must be greater
	// than LinearSlop. Ignored for loops.
	Ghost1 Vec2
	// Ending ghost point for an open chain. Distance from last point must be greater
	// than LinearSlop. Ignored for loops.
	Ghost2 Vec2
	// One material for the whole chain or one for each segment. Cloned.
	Materials []SurfaceMaterial
	// The material count. Must be 1 or the segment count.
	MaterialCount int
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

type sensorOverlaps struct {
	overlaps []int
}

const (
	DefaultCategoryBits uint64 = 1
	DefaultMaskBits     uint64 = 0xffffffffffffffff
)

func getShape(world *world, shapeID Shape) *shape {
	todo()
	return nil
}

func getChainShape(world *world, chainID Chain) *chainShape {
	todo()
	return nil
}

func computeShapeMargin(shape *shape) float32 {
	todo()
	return 0
}

func updateShapeAABBs(shape *shape, fatAABB *AABB, transform WorldTransform, proxyType BodyType) {
	todo()
}

func createShapeInternal(world *world, body *body, transform WorldTransform, def *ShapeDef, geometry any, shapeType ShapeType) *shape {
	todo()
	return nil
}

func isValidMaterial(material *SurfaceMaterial) bool {
	todo()
	return false
}

func isValidCircle(circle *Circle) bool {
	todo()
	return false
}

func isValidCapsule(capsule *Capsule) bool {
	todo()
	return false
}

func isValidSegment(segment *Segment) bool {
	todo()
	return false
}

func isValidChainSegment(chainSegment *ChainSegment) bool {
	todo()
	return false
}

func isValidPolygon(polygon *Polygon) bool {
	todo()
	return false
}

func createShape(bodyID Body, def *ShapeDef, geometry any, shapeType ShapeType) Shape {
	todo()
	return Shape{}
}

// Create a circle shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape
func (b Body) CreateCircleShape(def *ShapeDef, circle *Circle) Shape {
	todo()
	return Shape{}
}

// Create a capsule shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape, this will be Shape{} if the length is too small.
func (b Body) CreateCapsuleShape(def *ShapeDef, capsule *Capsule) Shape {
	todo()
	return Shape{}
}

// Create a polygon shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id for accessing the shape
func (b Body) CreatePolygonShape(def *ShapeDef, polygon *Polygon) Shape {
	todo()
	return Shape{}
}

// Create a line segment shape and attach it to a body. The shape definition and geometry are fully cloned.
// Contacts are not created until the next time step.
// Returns: the shape id or Shape{} if the segment is too short.
func (b Body) CreateSegmentShape(def *ShapeDef, segment *Segment) Shape {
	todo()
	return Shape{}
}

// Create an orphaned chain segment shape and attach it to a body. The shape definition and
// geometry are fully cloned. The caller is responsible for the segment's ghost vertices and
// lifetime. The segment is not owned by any chainShape (Shape.ParentChain returns
// Chain{}). Contacts are not created until the next time step.
// Returns: the shape id, or Shape{} if the segment is too short.
func (b Body) CreateChainSegmentShape(def *ShapeDef, chainSegment *ChainSegment) Shape {
	todo()
	return Shape{}
}

func destroyShapeInternal(world *world, shape *shape, body *body) {
	todo()
}

// Destroy a shape. You may defer the body mass update which can improve performance if several shapes on a
//
//	body are destroyed at once.
//	See: Body.UpdateMassFromShapes
func (s Shape) Destroy(updateBodyMass bool) {
	todo()
}

// Create a chain shape
// See: ChainDef for details
func (b Body) CreateChain(def *ChainDef) Chain {
	todo()
	return Chain{}
}

func freeChainData(chain *chainShape) {
	todo()
}

// Destroy a chain shape
func (c Chain) Destroy() {
	todo()
}

// Get the world that owns this chain shape
func (c Chain) World() World {
	todo()
	return World{}
}

// Get the number of segments on this chain
func (c Chain) SegmentCount() int {
	todo()
	return 0
}

// Get chain segment shape ids.
func (c Chain) Segments() iter.Seq[Shape] {
	todo()
	return nil
}

// Set a chain material by segment index.
func (c Chain) SetSurfaceMaterial(material *SurfaceMaterial, segmentIndex int) {
	todo()
}

// Set the chain material on all segments.
func (c Chain) SetAllSurfaceMaterials(material *SurfaceMaterial) {
	todo()
}

// Get a chain material by segment index.
func (c Chain) SurfaceMaterial(segmentIndex int) SurfaceMaterial {
	todo()
	return SurfaceMaterial{}
}

func computeShapeAABB(shape *shape, xf WorldTransform) AABB {
	todo()
	return AABB{}
}

func getShapeCentroid(shape *shape) Vec2 {
	todo()
	return Vec2{}
}

func getShapePerimeter(shape *shape) float32 {
	todo()
	return 0
}

func getShapeProjectedPerimeter(shape *shape, line Vec2) float32 {
	todo()
	return 0
}

func computeShapeMass(shape *shape) MassData {
	todo()
	return MassData{}
}

func computeShapeExtent(shape *shape, localCenter Vec2) shapeExtent {
	todo()
	return shapeExtent{}
}

func rayCastShape(input *RayCastInput, shape *shape, transform Transform) CastOutput {
	todo()
	return CastOutput{}
}

func shapeCastShape(input *ShapeCastInput, shape *shape, transform Transform) CastOutput {
	todo()
	return CastOutput{}
}

func collideMover(mover *Capsule, shape *shape, transform Transform) PlaneResult {
	todo()
	return PlaneResult{}
}

func createShapeProxy(world *world, shape *shape, bodyType BodyType, transform WorldTransform, forcePairCreation bool) {
	todo()
}

func destroyShapeProxy(shape *shape, bp *broadPhase) {
	todo()
}

func makeShapeDistanceProxy(shape *shape) ShapeProxy {
	todo()
	return ShapeProxy{}
}

// Get the id of the body that a shape is attached to
func (s Shape) Body() Body {
	todo()
	return Body{}
}

// Get the world that owns this shape
func (s Shape) World() World {
	todo()
	return World{}
}

// Set the user data for a shape
func (s Shape) SetUserData(userData any) {
	todo()
}

// Get the user data for a shape. This is useful when you get a shape id
// from an event or query.
func (s Shape) UserData() any {
	todo()
	return nil
}

// Returns true if the shape is a sensor. It is not possible to change a shape
// from sensor to solid dynamically because this breaks the contract for
// sensor events.
func (s Shape) IsSensor() bool {
	todo()
	return false
}

// Test a world point for overlap with a shape
func (s Shape) TestPoint(point Pos) bool {
	todo()
	return false
}

// Ray cast a single shape. The ray starts at the origin and extends by the translation. The output
// point is a world position.
func (s Shape) RayCast(origin Pos, translation Vec2) WorldCastOutput {
	todo()
	return WorldCastOutput{}
}

// Set the mass density of a shape, usually in kg/m^2.
// This will optionally update the mass properties on the parent body.
// See: ShapeDef.Density, Body.UpdateMassFromShapes
func (s Shape) SetDensity(density float32, updateBodyMass bool) {
	todo()
}

// Get the density of a shape, usually in kg/m^2
func (s Shape) Density() float32 {
	todo()
	return 0
}

// Set the friction on a shape
func (s Shape) SetFriction(friction float32) {
	todo()
}

// Get the friction of a shape
func (s Shape) Friction() float32 {
	todo()
	return 0
}

// Set the shape restitution (bounciness)
func (s Shape) SetRestitution(restitution float32) {
	todo()
}

// Get the shape restitution
func (s Shape) Restitution() float32 {
	todo()
	return 0
}

// Set the user material identifier
func (s Shape) SetUserMaterial(material uint64) {
	todo()
}

// Get the user material identifier
func (s Shape) UserMaterial() uint64 {
	todo()
	return 0
}

// Get the shape surface material
func (s Shape) SurfaceMaterial() SurfaceMaterial {
	todo()
	return SurfaceMaterial{}
}

// Set the shape surface material
func (s Shape) SetSurfaceMaterial(surfaceMaterial *SurfaceMaterial) {
	todo()
}

// Get the shape filter
func (s Shape) Filter() Filter {
	todo()
	return Filter{}
}

func resetProxy(world *world, shape *shape, destroyProxy bool) {
	todo()
}

// Set the current filter. This is almost as expensive as recreating the shape. This may cause
// contacts to be immediately destroyed. However contacts are not created until the next world step.
// Sensor overlap state is also not updated until the next world step.
// See: ShapeDef.Filter
func (s Shape) SetFilter(filter Filter) {
	todo()
}

// Enable sensor events for this shape.
// See: ShapeDef.EnableSensorEvents
func (s Shape) EnableSensorEvents(flag bool) {
	todo()
}

// Returns true if sensor events are enabled.
func (s Shape) AreSensorEventsEnabled() bool {
	todo()
	return false
}

// Enable contact events for this shape. Only applies to kinematic and dynamic bodies. Ignored for sensors.
// See: ShapeDef.EnableContactEvents
// Warning: changing this at run-time may lead to lost begin/end events
func (s Shape) EnableContactEvents(flag bool) {
	todo()
}

// Returns true if contact events are enabled
func (s Shape) AreContactEventsEnabled() bool {
	todo()
	return false
}

// Enable pre-solve contact events for this shape. Only applies to dynamic bodies. These are expensive
// and must be carefully handled due to multithreading. Ignored for sensors.
// See: PreSolveFunc
func (s Shape) EnablePreSolveEvents(flag bool) {
	todo()
}

// Returns true if pre-solve events are enabled
func (s Shape) ArePreSolveEventsEnabled() bool {
	todo()
	return false
}

// Enable contact hit events for this shape. Ignored for sensors.
// See: WorldDef.hitEventThreshold
func (s Shape) EnableHitEvents(flag bool) {
	todo()
}

// Returns true if hit events are enabled
func (s Shape) AreHitEventsEnabled() bool {
	todo()
	return false
}

// Get the type of a shape
func (s Shape) Type() ShapeType {
	todo()
	return 0
}

// Get a copy of the shape's circle. Asserts the type is correct.
func (s Shape) Circle() Circle {
	todo()
	return Circle{}
}

// Get a copy of the shape's line segment. Asserts the type is correct.
func (s Shape) Segment() Segment {
	todo()
	return Segment{}
}

// Get a copy of the shape's chain segment. These come from chain shapes.
// Asserts the type is correct.
func (s Shape) ChainSegment() ChainSegment {
	todo()
	return ChainSegment{}
}

// Get a copy of the shape's capsule. Asserts the type is correct.
func (s Shape) Capsule() Capsule {
	todo()
	return Capsule{}
}

// Get a copy of the shape's convex polygon. Asserts the type is correct.
func (s Shape) Polygon() Polygon {
	todo()
	return Polygon{}
}

// Allows you to change a shape to be a circle or update the current circle.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes
func (s Shape) SetCircle(circle *Circle) {
	todo()
}

// Allows you to change a shape to be a capsule or update the current capsule.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes
func (s Shape) SetCapsule(capsule *Capsule) {
	todo()
}

// Allows you to change a shape to be a segment or update the current segment.
func (s Shape) SetSegment(segment *Segment) {
	todo()
}

// Allows you to change a shape to be a polygon or update the current polygon.
// This does not modify the mass properties.
// See: Body.UpdateMassFromShapes
func (s Shape) SetPolygon(polygon *Polygon) {
	todo()
}

// Allows you to change a shape to be an orphaned chain segment or update the current chain
// segment, including its ghost vertices. The chainId on the input is ignored. The resulting
// shape is always orphaned. Asserts if the shape is already a chain segment
// owned by a chainShape (chainId != B2_NULL_INDEX).
func (s Shape) SetChainSegment(chainSegment *ChainSegment) {
	todo()
}

// Get the parent chain id if the shape type is a chain segment, otherwise
// returns Chain{}.
func (s Shape) ParentChain() Chain {
	todo()
	return Chain{}
}

// Get the maximum capacity required for retrieving all the touching contacts on a shape
func (s Shape) ContactCapacity() int {
	todo()
	return 0
}

// Get the touching contact data for a shape. The provided shapeId will be either shapeIdA or shapeIdB on the contact data.
// Note: Box2D uses speculative collision so some contact points may be separated.
func (s Shape) ContactData() iter.Seq[ContactData] {
	todo()
	return nil
}

// Get the maximum capacity required for retrieving all the overlapped shapes on a sensor shape.
// This returns 0 if the provided shape is not a sensor.
// shapeId: the id of a sensor shape
// Returns: the required capacity to get all the overlaps in Shape.SensorData
func (s Shape) SensorCapacity() int {
	todo()
	return 0
}

// Get the overlap data for a sensor shape computed the previous world step.
// shapeId: the id of a sensor shape
// Warning: overlaps may contain destroyed shapes so use shape_IsValid to confirm each overlap
func (s Shape) SensorData() iter.Seq[Shape] {
	todo()
	return nil
}

// Get the current world AABB
func (s Shape) AABB() AABB {
	todo()
	return AABB{}
}

// Compute the mass data for a shape
func (s Shape) ComputeMassData() MassData {
	todo()
	return MassData{}
}

// Get the closest point on a shape to a target point. Target and result are world positions.
// todo need sample
func (s Shape) ClosestPoint(target Pos) Pos {
	todo()
	return Pos{}
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
	todo()
}

func getShapeRadius(shape *shape) float32 {
	todo()
	return 0
}

func shouldShapesCollide(filterA Filter, filterB Filter) bool {
	todo()
	return false
}

func shouldQueryCollide(shapeFilter Filter, queryFilter QueryFilter) bool {
	todo()
	return false
}
