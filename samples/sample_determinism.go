// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

const (
	rollbackSnapshotInterval = 20
	hingeCameraCenterY       = 7.5
)

const (
	rollbackRunning rollbackResult = iota
	rollbackMatch
	rollbackMismatch
)

type rollbackSnapshot struct {
	image []byte
	step  int
}

type rollbackResult int

type Rollback struct {
	Sample

	snapshots          []rollbackSnapshot      //exhaustruct:optional
	data               scenes.FallingHingeData //exhaustruct:optional
	referenceSleepStep int
	referenceHash      uint32
	selected           int
	rollbackStep       int
	result             rollbackResult
	done               bool
}

type FallingHinges struct {
	Sample

	data scenes.FallingHingeData //exhaustruct:optional
	done bool
}

func CreateFallingHingesSample(context *SampleContext) Sampler {
	s := &FallingHinges{
		context:         context,
		camera:          &context.camera,
		draw:            context.draw,
		screenTextX:     initialScreenTextX,
		mouseForceScale: initialMouseForceScale,
		done:            false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: hingeCameraCenterY}
		s.context.camera.zoom = 10.0
	}

	s.data = scenes.CreateFallingHinges(s.worldID)

	return s
}

func (s *FallingHinges) Destroy() {
	scenes.DestroyFallingHinges(&s.data)
	s.Sample.Destroy()
}

func (s *FallingHinges) Step() {
	const (
		scanMinX                = 5.0
		scanMinY                = 1.0
		scanMaxX                = 7.0
		scanMaxY                = 2.5
		verticalRayOriginY      = 12.0
		verticalRayTranslationY = -14.0
		verticalRayStep         = 30
		horizontalRayOriginX    = -10.0
		horizontalRayOriginY    = 2.0
		horizontalRayLength     = 20.0
		horizontalRayEarlySteps = 5
		horizontalRayResumeStep = 100
	)

	s.Sample.Step()

	if !s.context.pause && !s.done {
		s.done = scenes.UpdateFallingHinges(s.worldID, &s.data)

		filter := box2d.DefaultQueryFilter()
		scanBox := box2d.AABB{
			LowerBound: box2d.Vec2{X: scanMinX, Y: scanMinY},
			UpperBound: box2d.Vec2{X: scanMaxX, Y: scanMaxY},
		}

		var scanOrigin box2d.Pos
		s.worldID.OverlapAABB(scanOrigin, scanBox, filter, func(box2d.Shape) bool { return true })

		origin := box2d.Pos{X: 0.0, Y: verticalRayOriginY}
		translation := box2d.Vec2{X: 0.0, Y: verticalRayTranslationY}

		if s.stepCount == verticalRayStep {
			s.worldID.CastRayClosest(origin, translation, filter)
		}

		if s.stepCount < horizontalRayEarlySteps || horizontalRayResumeStep < s.stepCount {
			origin = box2d.Pos{X: horizontalRayOriginX, Y: horizontalRayOriginY}
			translation = box2d.Vec2{X: horizontalRayLength, Y: 0.0}
			s.worldID.CastRay(
				origin,
				translation,
				filter,
				func(box2d.Shape, box2d.Pos, box2d.Vec2, float32) float32 { return 1.0 },
			)
		}

		if s.done {
			fmt.Printf("sleep step = %d, hash = 0x%08X\n", s.data.SleepStep, s.data.Hash)

			s.FinishRecording()
		}
	} else {
		s.DrawScreenTextLinef("sleep step = %d, hash = 0x%08X", s.data.SleepStep, s.data.Hash)
	}
}

func CreateRollback(context *SampleContext) Sampler {
	s := &Rollback{
		context:            context,
		camera:             &context.camera,
		draw:               context.draw,
		screenTextX:        initialScreenTextX,
		mouseForceScale:    initialMouseForceScale,
		referenceSleepStep: 0,
		referenceHash:      0,
		selected:           0,
		rollbackStep:       -1,
		result:             rollbackRunning,
		done:               false,
	}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: hingeCameraCenterY}
		s.context.camera.zoom = 10.0
	}

	s.data = scenes.CreateFallingHinges(s.worldID)

	s.Capture()

	return s
}

func (s *Rollback) Destroy() {
	s.snapshots = nil
	scenes.DestroyFallingHinges(&s.data)
	s.Sample.Destroy()
}

func (s *Rollback) Capture() {
	image, err := s.worldID.MarshalBinary()
	if err != nil {
		panic(err)
	}

	s.snapshots = append(s.snapshots, rollbackSnapshot{image: image, step: s.stepCount})
}

func (s *Rollback) RollBack(index int) {
	snapshot := &s.snapshots[index]
	if err := s.worldID.UnmarshalBinary(snapshot.image); err != nil {
		panic(err)
	}

	s.data.StepCount = snapshot.step
	s.data.SleepStep = -1
	s.data.Hash = 0

	s.rollbackStep = snapshot.step
	s.result = rollbackRunning
	s.done = false
}

func (s *Rollback) DrawControls() bool {
	const (
		bytesPerKilobyte           = 1024
		snapshotListRows           = 8.0
		rollbackButtonWidthInFonts = 8.0
		matchRedBlue               = 0.4
		matchGreen                 = 0.8
		mismatchRed                = 0.9
		mismatchGreenBlue          = 0.3
	)

	if s.referenceHash == 0 {
		imTextf("reference run: step %d", s.stepCount)
		return true
	}

	imTextf("reference sleep step %d", s.referenceSleepStep)
	imTextf("reference hash 0x%08X", s.referenceHash)

	totalSize := 0
	for _, snapshot := range s.snapshots {
		totalSize += len(snapshot.image)
	}

	imTextDisabledf("%d snapshots, %.0f kB", len(s.snapshots), float32(totalSize)/bytesPerKilobyte)

	if imgui.BeginListBoxV(
		"##snapshots",
		imgui.Vec2{X: -fltMin, Y: snapshotListRows * imgui.TextLineHeightWithSpacing()},
	) {
		var automaticSize imgui.Vec2

		for i := range s.snapshots {
			label := fmt.Sprintf(
				"step %d, %d kB",
				s.snapshots[i].step,
				len(s.snapshots[i].image)/bytesPerKilobyte,
			)

			if imgui.SelectableBoolV(label, i == s.selected, 0, automaticSize) {
				s.selected = i
			}
		}

		imgui.EndListBox()
	}

	if imgui.ButtonV(
		"Roll Back",
		imgui.Vec2{X: rollbackButtonWidthInFonts * imgui.FontSize(), Y: 0.0},
	) {
		s.RollBack(s.selected)
	}

	switch s.result {
	case rollbackRunning:
		imTextf("resimulating from step %d", s.rollbackStep)
	case rollbackMatch:
		imTextColoredf(
			imgui.Vec4{X: matchRedBlue, Y: matchGreen, Z: matchRedBlue, W: 1.0},
			"match: sleep step %d",
			s.data.SleepStep,
		)
		imTextColoredf(
			imgui.Vec4{X: matchRedBlue, Y: matchGreen, Z: matchRedBlue, W: 1.0},
			"match: hash 0x%08X", s.data.Hash,
		)
	case rollbackMismatch:
		fallthrough
	default:
		imTextColoredf(
			imgui.Vec4{X: mismatchRed, Y: mismatchGreenBlue, Z: mismatchGreenBlue, W: 1.0},
			"MISMATCH: sleep step %d",
			s.data.SleepStep,
		)
		imTextColoredf(
			imgui.Vec4{X: mismatchRed, Y: mismatchGreenBlue, Z: mismatchGreenBlue, W: 1.0},
			"MISMATCH: hash 0x%08X",
			s.data.Hash,
		)
	}

	return true
}

func (s *Rollback) Step() {
	s.Sample.Step()

	if !s.context.pause && !s.done {
		if s.referenceHash == 0 && s.stepCount%rollbackSnapshotInterval == 0 {
			s.Capture()
		}

		s.done = scenes.UpdateFallingHinges(s.worldID, &s.data)

		s.finishRollbackRun()
	}

	if s.referenceHash == 0 {
		s.DrawScreenTextLinef("reference run")
		return
	}

	if s.result == rollbackRunning {
		s.DrawScreenTextLinef("rolled back to step %d", s.rollbackStep)
		return
	}

	s.DrawScreenTextLinef(
		"sleep step = %d, hash = 0x%08X%s",
		s.data.SleepStep,
		s.data.Hash,
		mismatchSuffix(s.result),
	)
}

func (s *Rollback) finishRollbackRun() {
	if s.done {
		if s.referenceHash == 0 {
			s.referenceSleepStep = s.data.SleepStep
			s.referenceHash = s.data.Hash

			s.FinishRecording()
		}

		if s.data.SleepStep == s.referenceSleepStep && s.data.Hash == s.referenceHash {
			s.result = rollbackMatch
		} else {
			s.result = rollbackMismatch
		}

		fmt.Printf(
			"sleep step = %d, hash = 0x%08X%s\n",
			s.data.SleepStep,
			s.data.Hash,
			mismatchSuffix(s.result),
		)
	}
}

func mismatchSuffix(result rollbackResult) string {
	if result == rollbackMatch {
		return ""
	}

	return " MISMATCH"
}

func registerDeterminismSamples() {
	RegisterSample("Determinism", "Falling Hinges", CreateFallingHingesSample)
	RegisterSample("Determinism", "Rollback", CreateRollback)
}
