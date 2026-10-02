// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

// A begin touch event is generated when a shape starts to overlap a sensor shape.
type SensorBeginTouchEvent struct {
	// The id of the sensor shape
	SensorShapeID Shape
	// The id of the shape that began touching the sensor shape
	VisitorShapeID Shape
}

// An end touch event is generated when a shape stops overlapping a sensor shape.
//
//	These include things like setting the transform, destroying a body or shape, or changing
//	a filter. You will also get an end event if the sensor or visitor are destroyed.
//	Therefore you should always confirm the shape id is valid using Shape.IsValid.
type SensorEndTouchEvent struct {
	// The id of the sensor shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	SensorShapeID Shape
	// The id of the shape that stopped touching the sensor shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	VisitorShapeID Shape
}

// Sensor events are buffered in the world and are available
// as begin/end overlap event arrays after the time step is complete.
// Note: these may become invalid if bodies and/or shapes are destroyed
type SensorEvents struct {
	// Array of sensor begin touch events
	BeginEvents []SensorBeginTouchEvent
	// Array of sensor end touch events
	EndEvents []SensorEndTouchEvent
	// The number of begin touch events
	BeginCount int
	// The number of end touch events
	EndCount int
}

type sensorHit struct {
	sensorID  int
	visitorID int
}

type visitor struct {
	shapeID    int
	generation uint16
}

type sensor struct {
	hits      []visitor
	overlaps1 []visitor
	overlaps2 []visitor
	shapeID   int
}

type sensorTaskContext struct {
	eventBits bitSet
}

type sensorQueryContext struct {
	world       *world
	taskContext *sensorTaskContext
	sensor      *sensor
	sensorShape *shape
	transform   WorldTransform
}

func sensorQueryCallback(proxyID int, userData uint64, context *sensorQueryContext) bool {
	todo()
	return false
}

func compareVisitors(a visitor, b visitor) int {
	todo()
	return 0
}

func sensorTask(startIndex int, endIndex int, threadIndex int, context *world) {
	todo()
}

func overlapSensors(world *world) {
	todo()
}

func destroySensor(world *world, sensorShape *shape) {
	todo()
}
