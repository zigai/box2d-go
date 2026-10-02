// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"testing"
)

var inputLoopTriangle = []box2d.Vec2{{X: -2, Y: 0}, {X: 2, Y: 0}, {X: 0, Y: 3}}
var inputLoopSquare = []box2d.Vec2{{X: -2, Y: 0}, {X: 2, Y: 0}, {X: 2, Y: 4}, {X: -2, Y: 4}}
var inputOpenLine = []box2d.Vec2{{X: 0, Y: 0}, {X: 2, Y: 1}, {X: 4, Y: 1}, {X: 6, Y: 0}, {X: 8, Y: 0}, {X: 10, Y: 1}}
var inputOpenPair = []box2d.Vec2{{X: 0, Y: 0}, {X: 4, Y: 0}}
var inputGhostBegin = box2d.Vec2{X: -2, Y: -1}
var inputGhostEnd = box2d.Vec2{X: 12, Y: 2}

func inputCreateTaggedChain(bodyID box2d.Body, points []box2d.Vec2, pointCount, materialCount int, isLoop bool) box2d.Chain {
	materials := make([]box2d.SurfaceMaterial, 8)
	for i := 0; i < materialCount; i++ {
		materials[i] = box2d.DefaultSurfaceMaterial()
		materials[i].UserMaterialID = uint64(i + 1)
	}
	chainDef := box2d.DefaultChainDef()
	chainDef.Points = points
	chainDef.PointCount = pointCount
	chainDef.Materials = materials
	chainDef.MaterialCount = materialCount
	chainDef.IsLoop = isLoop
	chainDef.Ghost1 = inputGhostBegin
	chainDef.Ghost2 = inputGhostEnd
	return bodyID.CreateChain(&chainDef)
}

func inputReadSegmentTags(chainID box2d.Chain, tags []uint64, count int) {
	segments := make([]box2d.Shape, 0, count)
	for segment := range chainID.Segments() {
		if len(segments) == count {
			break
		}
		segments = append(segments, segment)
	}
	for i := 0; i < count; i++ {
		tags[i] = segments[i].SurfaceMaterial().UserMaterialID
	}
}

func inputSamePoint(a, b box2d.Vec2) bool {
	return a.X == b.X && a.Y == b.Y
}

func TestChain(t *testing.T) {

	t.Run("SegmentCountTest", func(t *testing.T) {

		r := oracle.Open(t, "input/chain/SegmentCountTest")
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyID := worldID.CreateBody(&bodyDef)

		triangle := inputCreateTaggedChain(bodyID, inputLoopTriangle, 3, 1, true).SegmentCount()
		square := inputCreateTaggedChain(bodyID, inputLoopSquare, 4, 1, true).SegmentCount()
		pair := inputCreateTaggedChain(bodyID, inputOpenPair, 2, 1, false).SegmentCount()
		line := inputCreateTaggedChain(bodyID, inputOpenLine, 6, 1, false).SegmentCount()

		worldID.Destroy()

		{
			ok := triangle == 3
			r.Bool("check.120", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}
		{
			ok := square == 4
			r.Bool("check.121", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}
		{
			ok := pair == 1
			r.Bool("check.122", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}
		{
			ok := line == 5
			r.Bool("check.123", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

	})
	t.Run("SegmentGeometryTest", func(t *testing.T) {

		r := oracle.Open(t, "input/chain/SegmentGeometryTest")
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyID := worldID.CreateBody(&bodyDef)

		loopID := inputCreateTaggedChain(bodyID, inputLoopSquare, 4, 1, true)
		openID := inputCreateTaggedChain(bodyID, inputOpenLine, 6, 1, false)

		var loopSegments [4]box2d.Shape
		var openSegments [5]box2d.Shape
		loopCount := 0
		for segment := range loopID.Segments() {
			if loopCount == len(loopSegments) {
				break
			}
			loopSegments[loopCount] = segment
			loopCount++
		}
		openCount := 0
		for segment := range openID.Segments() {
			if openCount == len(openSegments) {
				break
			}
			openSegments[openCount] = segment
			openCount++
		}

		var loop [4]box2d.ChainSegment
		var open [5]box2d.ChainSegment
		for i := 0; i < 4; i++ {
			loop[i] = loopSegments[i].ChainSegment()
		}
		for i := 0; i < 5; i++ {
			open[i] = openSegments[i].ChainSegment()
		}

		worldID.Destroy()

		{
			ok := loopCount == 4
			r.Bool("check.159", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}
		{
			ok := openCount == 5
			r.Bool("check.160", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

		// ENSURE stringizes into a format string, so keep a percent out of the condition
		for i := 0; i < 4; i++ {
			next := (i + 1) % 4
			after := (i + 2) % 4
			prev := (i + 3) % 4
			{
				ok := inputSamePoint(loop[i].Segment.Point1, inputLoopSquare[i])
				r.Bool("check.168", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := inputSamePoint(loop[i].Segment.Point2, inputLoopSquare[next])
				r.Bool("check.169", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := inputSamePoint(loop[i].Ghost1, inputLoopSquare[prev])
				r.Bool("check.170", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := inputSamePoint(loop[i].Ghost2, inputLoopSquare[after])
				r.Bool("check.171", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

		for i := 0; i < 5; i++ {
			{
				ok := inputSamePoint(open[i].Segment.Point1, inputOpenLine[i])
				r.Bool("check.176", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := inputSamePoint(open[i].Segment.Point2, inputOpenLine[i+1])
				r.Bool("check.177", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

		// The def ghosts extend the chain past its ends without consuming a point or a segment
		{
			ok := inputSamePoint(open[0].Ghost1, inputGhostBegin)
			r.Bool("check.181", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}
		{
			ok := inputSamePoint(open[4].Ghost2, inputGhostEnd)
			r.Bool("check.182", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

		for i := 1; i < 5; i++ {
			{
				ok := inputSamePoint(open[i].Ghost1, inputOpenLine[i-1])
				r.Bool("check.186", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}
		for i := 0; i < 4; i++ {
			{
				ok := inputSamePoint(open[i].Ghost2, inputOpenLine[i+2])
				r.Bool("check.190", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

	})
	t.Run("LoopMaterialTest", func(t *testing.T) {

		r := oracle.Open(t, "input/chain/LoopMaterialTest")
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		assertCount := 0
		beginCountingAsserts := func() {
			assertCount = 0
			box2d.SetAssertFunc(func(string, string, int) bool { assertCount++; return false })
		}
		endCountingAsserts := func() int { restore(); return assertCount }
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyID := worldID.CreateBody(&bodyDef)
		chainID := inputCreateTaggedChain(bodyID, inputLoopSquare, 4, 4, true)

		var created [4]uint64
		inputReadSegmentTags(chainID, created[:], 4)

		material := box2d.DefaultSurfaceMaterial()
		material.UserMaterialID = 99

		beginCountingAsserts()
		chainID.SetSurfaceMaterial(&material, 1)
		assertsTripped := endCountingAsserts()

		var updated [4]uint64
		var reported [4]uint64
		inputReadSegmentTags(chainID, updated[:], 4)
		for i := 0; i < 4; i++ {
			reported[i] = chainID.SurfaceMaterial(i).UserMaterialID
		}

		worldID.Destroy()

		{
			ok := assertsTripped == 0
			r.Bool("check.225", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

		for i := 0; i < 4; i++ {
			expected := uint64(i + 1)
			if i == 1 {
				expected = 99
			}
			{
				ok := created[i] == uint64(i+1)
				r.Bool("check.230", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := updated[i] == expected
				r.Bool("check.231", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}

			// The getter and the segment shape cannot disagree, there is one copy
			{
				ok := reported[i] == expected
				r.Bool("check.234", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

	})
	t.Run("OpenMaterialTest", func(t *testing.T) {

		r := oracle.Open(t, "input/chain/OpenMaterialTest")
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		assertCount := 0
		beginCountingAsserts := func() {
			assertCount = 0
			box2d.SetAssertFunc(func(string, string, int) bool { assertCount++; return false })
		}
		endCountingAsserts := func() int { restore(); return assertCount }
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyID := worldID.CreateBody(&bodyDef)
		chainID := inputCreateTaggedChain(bodyID, inputOpenLine, 6, 5, false)

		var created [5]uint64
		inputReadSegmentTags(chainID, created[:], 5)

		material := box2d.DefaultSurfaceMaterial()
		material.UserMaterialID = 99

		beginCountingAsserts()
		chainID.SetSurfaceMaterial(&material, 0)
		material.UserMaterialID = 98
		chainID.SetSurfaceMaterial(&material, 4)
		assertsTripped := endCountingAsserts()

		var updated [5]uint64
		var reported [5]uint64
		inputReadSegmentTags(chainID, updated[:], 5)
		for i := 0; i < 5; i++ {
			reported[i] = chainID.SurfaceMaterial(i).UserMaterialID
		}

		worldID.Destroy()

		{
			ok := assertsTripped == 0
			r.Bool("check.273", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

		for i := 0; i < 5; i++ {
			expected := uint64(i + 1)
			if i == 0 {
				expected = 99
			} else if i == 4 {
				expected = 98
			}

			{
				ok := created[i] == uint64(i+1)
				r.Bool("check.287", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := updated[i] == expected
				r.Bool("check.288", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := reported[i] == expected
				r.Bool("check.289", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

	})
	t.Run("SingleMaterialTest", func(t *testing.T) {

		r := oracle.Open(t, "input/chain/SingleMaterialTest")
		restore := box2d.InputSaveInputHooks()
		t.Cleanup(restore)
		assertCount := 0
		beginCountingAsserts := func() {
			assertCount = 0
			box2d.SetAssertFunc(func(string, string, int) bool { assertCount++; return false })
		}
		endCountingAsserts := func() int { restore(); return assertCount }
		worldDef := box2d.DefaultWorldDef()
		worldID := box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldID)

		bodyDef := box2d.DefaultBodyDef()
		bodyID := worldID.CreateBody(&bodyDef)
		chainID := inputCreateTaggedChain(bodyID, inputOpenLine, 6, 1, false)

		var created [5]uint64
		inputReadSegmentTags(chainID, created[:], 5)

		material := box2d.DefaultSurfaceMaterial()
		material.UserMaterialID = 99

		beginCountingAsserts()
		chainID.SetAllSurfaceMaterials(&material)
		assertsTripped := endCountingAsserts()

		var broadcast [5]uint64
		inputReadSegmentTags(chainID, broadcast[:], 5)

		// One material at creation must not narrow the segment index, which is what the old count
		// bound would have rejected here
		material.UserMaterialID = 98
		beginCountingAsserts()
		chainID.SetSurfaceMaterial(&material, 2)
		assertsTripped += endCountingAsserts()

		var single [5]uint64
		var reported [5]uint64
		inputReadSegmentTags(chainID, single[:], 5)
		for i := 0; i < 5; i++ {
			reported[i] = chainID.SurfaceMaterial(i).UserMaterialID
		}

		worldID.Destroy()

		{
			ok := assertsTripped == 0
			r.Bool("check.336", ok)
			if !(ok) {
				t.Fatal("condition false")
			}
		}

		for i := 0; i < 5; i++ {
			expected := uint64(99)
			if i == 2 {
				expected = 98
			}
			{
				ok := created[i] == 1
				r.Bool("check.341", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := broadcast[i] == 99
				r.Bool("check.342", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := single[i] == expected
				r.Bool("check.343", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
			{
				ok := reported[i] == expected
				r.Bool("check.344", ok)
				if !(ok) {
					t.Fatal("condition false")
				}
			}
		}

	})
}
