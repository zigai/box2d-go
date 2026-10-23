// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"cmp"
	"math/bits"
	"slices"

	"github.com/zigai/box2d/internal/array"
	"github.com/zigai/box2d/internal/bitset"
	"github.com/zigai/box2d/internal/core"
)

// A begin touch event is generated when a shape starts to overlap a sensor shape.
type SensorBeginTouchEvent struct {
	// The id of the sensor shape
	SensorShape Shape
	// The id of the shape that began touching the sensor shape
	VisitorShape Shape
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
	SensorShape Shape
	// The id of the shape that stopped touching the sensor shape
	// 	Warning: this shape may have been destroyed
	// 	See: Shape.IsValid
	VisitorShape Shape
}

// Sensor events are buffered in the world and are available
// as begin/end overlap event arrays after the time step is complete.
// Note: these may become invalid if bodies and/or shapes are destroyed.
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
	eventBits bitset.Set
}

type sensorQueryContext struct {
	world       *World
	taskContext *sensorTaskContext
	sensor      *sensor
	sensorShape *shape
	transform   WorldTransform
}

func sensorQueryCallback(userData uint64, context *sensorQueryContext) bool {
	shapeID := int(userData)
	sensorShape := context.sensorShape

	sensorShapeID := sensorShape.id
	if shapeID == sensorShapeID {
		return true
	}

	world := context.world
	otherShape := &world.shapes[shapeID]
	// Are sensor events enabled on the other shape?
	if !otherShape.enableSensorEvents {
		return true
	}
	// Skip shapes on the same body
	if otherShape.bodyID == sensorShape.bodyID {
		return true
	}
	// Check filter
	if !shouldShapesCollide(sensorShape.filter, otherShape.filter) {
		return true
	}
	// Custom user filter
	if sensorShape.enableCustomFiltering || otherShape.enableCustomFiltering {
		customFilterFcn := world.customFilterFcn
		if customFilterFcn != nil {
			idA := Shape{index1: int32(sensorShapeID + 1), world: world, generation: sensorShape.generation}
			idB := Shape{index1: int32(shapeID + 1), world: world, generation: otherShape.generation}

			shouldCollide := customFilterFcn(idA, idB)
			if !shouldCollide {
				return true
			}
		}
	}
	// The relative pose is differenced in double so sensor overlap stays exact far from the origin
	otherTransform := getBodyTransform(world, otherShape.bodyID)

	var input DistanceInput

	input.ProxyA = makeShapeDistanceProxy(sensorShape)
	input.ProxyB = makeShapeDistanceProxy(otherShape)
	input.Transform = context.transform.InvMul(otherTransform)
	input.UseRadii = true

	var cache SimplexCache

	output := ShapeDistance(&input, &cache, nil)

	overlaps := output.Distance < 10.0*float32(0x1p-23)
	if !overlaps {
		return true
	}
	// Record the overlap
	sensor := context.sensor

	var shapeRef *visitor

	sensor.overlaps2, shapeRef = array.Emplace(sensor.overlaps2)
	shapeRef.shapeID = shapeID
	shapeRef.generation = otherShape.generation

	return true
}

func compareVisitors(a visitor, b visitor) int {
	if a.shapeID < b.shapeID {
		return -1
	}

	if a.shapeID == b.shapeID {
		return 0
	}

	return 1
}

func sensorTask(startIndex int, endIndex int, threadIndex int, context *World) {
	world := context
	taskContext := &world.sensorTaskContexts[threadIndex]

	core.Assert(startIndex < endIndex, "startIndex < endIndex")

	trees := &world.broadPhase.trees
	for sensorIndex := startIndex; sensorIndex < endIndex; sensorIndex++ {
		sensor := &world.sensors[sensorIndex]
		sensorShape := &world.shapes[sensor.shapeID]
		// Swap overlap arrays
		sensor.overlaps1, sensor.overlaps2 = sensor.overlaps2, sensor.overlaps1
		sensor.overlaps2 = sensor.overlaps2[:0]
		// Append sensor hits
		hitCount := len(sensor.hits)
		for i := range hitCount {
			sensor.overlaps2 = array.Push(sensor.overlaps2, sensor.hits[i])
		}
		// Clear the hits
		sensor.hits = sensor.hits[:0]

		body := &world.bodies[sensorShape.bodyID]
		if body.setIndex == disabledSet || !sensorShape.enableSensorEvents {
			if len(sensor.overlaps1) != 0 {
				// This sensor is dropping all overlaps because it has been disabled.
				taskContext.eventBits.SetBit(uint32(sensorIndex))
			}

			continue
		}

		transform := getBodyTransformQuick(world, body)
		queryContext := sensorQueryContext{
			world:       world,
			taskContext: taskContext,
			sensor:      sensor,
			sensorShape: sensorShape,
			transform:   transform,
		}
		core.Assert(sensorShape.sensorIndex == sensorIndex, "sensorShape->sensorIndex == sensorIndex")
		queryBounds := sensorShape.aabb
		// Query all trees
		callback := func(_ int, userData uint64) bool {
			return sensorQueryCallback(userData, &queryContext)
		}
		trees[0].Query(queryBounds, sensorShape.filter.MaskBits, callback)
		trees[1].Query(queryBounds, sensorShape.filter.MaskBits, callback)
		trees[2].Query(queryBounds, sensorShape.filter.MaskBits, callback)
		// Sort the overlaps to enable finding begin and end events.
		slices.SortFunc(sensor.overlaps2, compareVisitors)
		// Remove duplicates from overlaps2 (sorted). Duplicates are possible due to the hit events appended earlier.
		uniqueCount := 0
		overlapCount := len(sensor.overlaps2)

		overlapData := sensor.overlaps2
		for i := range overlapCount {
			if uniqueCount == 0 || overlapData[i].shapeID != overlapData[uniqueCount-1].shapeID {
				overlapData[uniqueCount] = overlapData[i]
				uniqueCount++
			}
		}

		sensor.overlaps2 = sensor.overlaps2[:uniqueCount]
		if sensorOverlapsChanged(sensor.overlaps1, sensor.overlaps2) {
			// something changed
			taskContext.eventBits.SetBit(uint32(sensorIndex))
		}
	}
}

func sensorOverlapsChanged(previous, current []visitor) bool {
	if len(previous) != len(current) {
		return true
	}

	for i := range previous {
		if previous[i].shapeID != current[i].shapeID || previous[i].generation != current[i].generation {
			return true
		}
	}

	return false
}

func overlapSensors(world *World) {
	sensorCount := len(world.sensors)
	if sensorCount == 0 {
		return
	}

	core.Assert(world.workerCount > 0, "world->workerCount > 0")

	for i := range world.workerCount {
		world.sensorTaskContexts[i].eventBits.SetBitCountAndClear(uint32(sensorCount))
	}
	// Parallel-for sensors overlaps
	minRange := 16
	parallelFor(world, func(startIndex, endIndex, threadIndex int) {
		sensorTask(startIndex, endIndex, threadIndex, world)
	}, sensorCount, minRange)

	bitSet := &world.sensorTaskContexts[0].eventBits
	for i := 1; i < world.workerCount; i++ {
		bitSet.InPlaceUnion(&world.sensorTaskContexts[i].eventBits)
	}
	// Iterate sensors bits and publish events
	// Process sensor state changes. Iterate over set bits
	bitData := bitSet.Bits

	blockCount := bitSet.BlockCount
	for k := range blockCount {
		word := bitData[k]
		for word != 0 {
			ctz := uint32(bits.TrailingZeros64(word))
			sensorIndex := int(64*k + ctz)
			publishSensorEvents(world, sensorIndex)
			// Clear the smallest set bit
			word &= word - 1
		}
	}
}

func publishSensorEvents(world *World, sensorIndex int) {
	sensor := &world.sensors[sensorIndex]
	sensorShape := &world.shapes[sensor.shapeID]
	sensorID := Shape{index1: int32(sensor.shapeID + 1), world: world, generation: sensorShape.generation}
	count1 := len(sensor.overlaps1)
	count2 := len(sensor.overlaps2)
	refs1 := sensor.overlaps1
	refs2 := sensor.overlaps2
	// overlaps1 can have overlaps that end
	// overlaps2 can have overlaps that begin
	index1, index2 := 0, 0
	for index1 < count1 && index2 < count2 {
		r1 := &refs1[index1]
		r2 := &refs2[index2]
		order := cmp.Compare(r1.shapeID, r2.shapeID)

		if order == 0 {
			order = cmp.Compare(r1.generation, r2.generation)
		}

		switch order {
		case -1:
			// end
			visitorID := Shape{index1: int32(r1.shapeID + 1), world: world, generation: r1.generation}
			event := SensorEndTouchEvent{SensorShape: sensorID, VisitorShape: visitorID}
			world.sensorEndEvents[world.endEventArrayIndex] = array.Push(world.sensorEndEvents[world.endEventArrayIndex], event)
			index1++
		case 1:
			// begin
			visitorID := Shape{index1: int32(r2.shapeID + 1), world: world, generation: r2.generation}
			event := SensorBeginTouchEvent{SensorShape: sensorID, VisitorShape: visitorID}
			world.sensorBeginEvents = array.Push(world.sensorBeginEvents, event)
			index2++
		case 0:
			// persisted
			index1++
			index2++
		}
	}

	for index1 < count1 {
		// end
		r1 := &refs1[index1]
		visitorID := Shape{index1: int32(r1.shapeID + 1), world: world, generation: r1.generation}
		event := SensorEndTouchEvent{SensorShape: sensorID, VisitorShape: visitorID}
		world.sensorEndEvents[world.endEventArrayIndex] = array.Push(world.sensorEndEvents[world.endEventArrayIndex], event)
		index1++
	}

	for index2 < count2 {
		// begin
		r2 := &refs2[index2]
		visitorID := Shape{index1: int32(r2.shapeID + 1), world: world, generation: r2.generation}
		event := SensorBeginTouchEvent{SensorShape: sensorID, VisitorShape: visitorID}
		world.sensorBeginEvents = array.Push(world.sensorBeginEvents, event)
		index2++
	}
}

func destroySensor(world *World, sensorShape *shape) {
	sensor := &world.sensors[sensorShape.sensorIndex]
	for i := range len(sensor.overlaps2) {
		ref := &sensor.overlaps2[i]
		event := SensorEndTouchEvent{
			SensorShape:  Shape{index1: int32(sensorShape.id + 1), world: world, generation: sensorShape.generation},
			VisitorShape: Shape{index1: int32(ref.shapeID + 1), world: world, generation: ref.generation},
		}
		world.sensorEndEvents[world.endEventArrayIndex] = array.Push(world.sensorEndEvents[world.endEventArrayIndex], event)
	}
	// Destroy sensor
	sensor.hits = nil
	sensor.overlaps1 = nil
	sensor.overlaps2 = nil

	var movedIndex int

	world.sensors, movedIndex = array.RemoveSwap(world.sensors, sensorShape.sensorIndex)
	if movedIndex != nullIndex {
		// Fixup moved sensor
		movedSensor := &world.sensors[sensorShape.sensorIndex]
		otherSensorShape := &world.shapes[movedSensor.shapeID]
		otherSensorShape.sensorIndex = sensorShape.sensorIndex
	}
}
