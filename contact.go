// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// A begin touch event is generated when two shapes begin touching.
type ContactBeginTouchEvent struct {
	// Id of the first shape
	ShapeIDA Shape
	// Id of the second shape
	ShapeIDB Shape
	// The transient contact id. This contact maybe destroyed automatically when the world is modified or simulated.
	// Used Contact.IsValid before using this id.
	ContactID Contact
}

// An end touch event is generated when two shapes stop touching.
//
//	You will get an end event if you do anything that destroys contacts previous to the last
//	world step. These include things like setting the transform, destroying a body
//	or shape, or changing a filter or body type.
type ContactEndTouchEvent struct {
	// Id of the first shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	ShapeIDA Shape
	// Id of the second shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	ShapeIDB Shape
	// Id of the contact.
	// 	Warning: this contact may have been destroyed
	// 	See: Contact.IsValid
	ContactID Contact
}

// A hit touch event is generated when two shapes collide with a speed faster than the hit speed threshold.
// This may be reported for speculative contacts that have a confirmed impulse.
type ContactHitEvent struct {
	// Id of the first shape
	ShapeIDA Shape
	// Id of the second shape
	ShapeIDB Shape
	// Id of the contact.
	// 	Warning: this contact may have been destroyed
	// 	See: Contact.IsValid
	ContactID Contact
	// Point where the shapes hit. This is a mid-point between the two surfaces. It could be
	// a speculative point where the shapes were not touching at the beginning of the time
	// step. The point is reconstructed after the step, so when both bodies are moving it may
	// trail the impact by up to one step of motion.
	Point Pos
	// Normal vector pointing from shape A to shape B
	Normal Vec2
	// The speed the shapes are approaching. Always positive. Typically in meters per second.
	ApproachSpeed float32
}

// Contact events are buffered in the Box2D world and are available
// as event arrays after the time step is complete.
// Note: these may become invalid if bodies and/or shapes are destroyed
type ContactEvents struct {
	// Array of begin touch events
	BeginEvents []ContactBeginTouchEvent
	// Array of end touch events
	EndEvents []ContactEndTouchEvent
	// Array of hit events
	HitEvents []ContactHitEvent
	// Number of begin touch events
	BeginCount int
	// Number of end touch events
	EndCount int
	// Number of hit events
	HitCount int
}

// The contact data for two shapes. By convention the manifold normal points
// from shape A to shape B.
// See: Shape.ContactData() and Body.ContactData()
type ContactData struct {
	// The contact id. You may hold onto this to track a contact across time steps.
	// This id may become orphaned. Use Contact.IsValid before using it for other functions.
	ContactID Contact
	// The first shape id.
	ShapeIDA Shape
	// The second shape id.
	ShapeIDB Shape
	// The manifold copied from the contact.
	Manifold Manifold
}

type contactEdge struct {
	bodyID  int
	prevKey int
	nextKey int
}

type contact struct {
	edges       [2]contactEdge
	islandID    int
	islandIndex int
	setIndex    int
	colorIndex  int
	localIndex  int
	shapeIDA    int
	shapeIDB    int
	contactID   int
	flags       uint32
	generation  uint32
}

type contactSim struct {
	contactID          int
	cachedRotationA    Rot
	cachedRotationB    Rot
	cachedRelativePose Transform
	bodyIDA            int
	bodyIDB            int
	encodedBodySimA    int
	encodedBodySimB    int
	simFlags           uint32
	shapeIDA           int
	shapeIDB           int
	invMassA           float32
	invIA              float32
	invMassB           float32
	invIB              float32
	manifold           Manifold
	friction           float32
	restitution        float32
	rollingResistance  float32
	tangentSpeed       float32
	cache              SimplexCache
}

type contactRegister struct {
	fcn     manifoldFunc
	primary bool
}

type contactFlags uint32

const (
	contactTouchingFlag        uint32 = 0x00000001
	contactHitEventFlag        uint32 = 0x00000002
	contactEnableContactEvents uint32 = 0x00000004
	contactRecycleFlag         uint32 = 0x00000008
	simTouchingFlag            uint32 = 0x00010000
	simDisjoint                uint32 = 0x00020000
	simStartedTouching         uint32 = 0x00040000
	simStoppedTouching         uint32 = 0x00080000
	simEnableHitEvent          uint32 = 0x00100000
	simEnablePreSolveEvents    uint32 = 0x00200000
	simRelativeTransformValid  uint32 = 0x00400000
)

type manifoldFunc func(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold

var contactRegistry = [ShapeTypeCount][ShapeTypeCount]contactRegister{
	CircleShape: {
		CircleShape:       {fcn: circleManifold, primary: true},
		CapsuleShape:      {fcn: capsuleAndCircleManifold, primary: false},
		SegmentShape:      {fcn: segmentAndCircleManifold, primary: false},
		PolygonShape:      {fcn: polygonAndCircleManifold, primary: false},
		ChainSegmentShape: {fcn: chainSegmentAndCircleManifold, primary: false},
	},
	CapsuleShape: {
		CircleShape:       {fcn: capsuleAndCircleManifold, primary: true},
		CapsuleShape:      {fcn: capsuleManifold, primary: true},
		SegmentShape:      {fcn: segmentAndCapsuleManifold, primary: false},
		PolygonShape:      {fcn: polygonAndCapsuleManifold, primary: false},
		ChainSegmentShape: {fcn: chainSegmentAndCapsuleManifold, primary: false},
	},
	SegmentShape: {
		CircleShape:  {fcn: segmentAndCircleManifold, primary: true},
		CapsuleShape: {fcn: segmentAndCapsuleManifold, primary: true},
		PolygonShape: {fcn: segmentAndPolygonManifold, primary: true},
	},
	PolygonShape: {
		CircleShape:       {fcn: polygonAndCircleManifold, primary: true},
		CapsuleShape:      {fcn: polygonAndCapsuleManifold, primary: true},
		SegmentShape:      {fcn: segmentAndPolygonManifold, primary: false},
		PolygonShape:      {fcn: polygonManifold, primary: true},
		ChainSegmentShape: {fcn: chainSegmentAndPolygonManifold, primary: false},
	},
	ChainSegmentShape: {
		CircleShape:  {fcn: chainSegmentAndCircleManifold, primary: true},
		CapsuleShape: {fcn: chainSegmentAndCapsuleManifold, primary: true},
		PolygonShape: {fcn: chainSegmentAndPolygonManifold, primary: true},
	},
}

func circleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func capsuleAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func segmentAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func polygonAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func chainSegmentAndCircleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func capsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func segmentAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func polygonAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func chainSegmentAndCapsuleManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func segmentAndPolygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func polygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func chainSegmentAndPolygonManifold(shapeA, shapeB *shape, xf Transform, cache *SimplexCache) LocalManifold {
	todo()
	return LocalManifold{}
}

func getContactFullId(world *world, contactID Contact) *contact {
	todo()
	return nil
}

// Get the data for a contact. The manifold may have no points if the contact is not touching.
func (c Contact) Data() ContactData {
	todo()
	return ContactData{}
}

func canCollide(typeA ShapeType, typeB ShapeType) bool {
	todo()
	return false
}

func createContact(world *world, shapeA *shape, shapeB *shape) {
	todo()
}

func destroyContact(world *world, contact *contact) {
	todo()
}

func getContactSim(world *world, contact *contact) *contactSim {
	todo()
	return nil
}

func updateContact(world *world, contactSim *contactSim, shapeA *shape, transformA WorldTransform, centerOffsetA Vec2, shapeB *shape, transformB WorldTransform, centerOffsetB Vec2) bool {
	todo()
	return false
}

func mixFriction(frictionA float32, frictionB float32) float32 {
	todo()
	return 0
}

func mixRestitution(restitutionA float32, restitutionB float32) float32 {
	todo()
	return 0
}
