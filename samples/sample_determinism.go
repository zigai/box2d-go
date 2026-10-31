// SPDX-FileCopyrightText: 2022 Erin Catto
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/scenes"
)

type FallingHinges struct {
	Sample
	data scenes.FallingHingeData
	done bool
}

func CreateFallingHingesSample(context *SampleContext) Sampler {
	s := &FallingHinges{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 7.5}
		s.context.camera.zoom = 10.0
	}

	s.data = scenes.CreateFallingHinges(s.worldID)
	s.done = false
	return s
}

func (s *FallingHinges) Destroy() {
	scenes.DestroyFallingHinges(&s.data)
	s.Sample.Destroy()
}

func (s *FallingHinges) Step() {
	s.Sample.Step()

	if !s.context.pause && !s.done {
		s.done = scenes.UpdateFallingHinges(s.worldID, &s.data)

		filter := box2d.DefaultQueryFilter()
		scanBox := box2d.AABB{
			LowerBound: box2d.Vec2{X: 5.0, Y: 1.0},
			UpperBound: box2d.Vec2{X: 7.0, Y: 2.5},
		}
		s.worldID.OverlapAABB(box2d.Pos{}, scanBox, filter, func(box2d.Shape) bool { return true })

		origin := box2d.Pos{X: 0.0, Y: 12.0}
		translation := box2d.Vec2{X: 0.0, Y: -14.0}

		if s.stepCount == 30 {
			s.worldID.CastRayClosest(origin, translation, filter)
		}

		if s.stepCount < 5 || 100 < s.stepCount {
			origin = box2d.Pos{X: -10.0, Y: 2.0}
			translation = box2d.Vec2{X: 20.0, Y: 0.0}
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
		s.DrawScreenTextLine("sleep step = %d, hash = 0x%08X", s.data.SleepStep, s.data.Hash)
	}
}

type rollbackSnapshot struct {
	image []byte
	step  int
}

const rollbackSnapshotInterval = 20

type rollbackResult int

const (
	rollbackRunning rollbackResult = iota
	rollbackMatch
	rollbackMismatch
)

type Rollback struct {
	Sample
	snapshots          []rollbackSnapshot
	data               scenes.FallingHingeData
	referenceSleepStep int
	referenceHash      uint32
	selected           int
	rollbackStep       int
	result             rollbackResult
	done               bool
}

func CreateRollback(context *SampleContext) Sampler {
	s := &Rollback{}
	s.init(context, s, true)

	if !s.context.restart {
		s.context.camera.center = box2d.Pos{X: 0.0, Y: 7.5}
		s.context.camera.zoom = 10.0
	}

	s.data = scenes.CreateFallingHinges(s.worldID)
	s.referenceSleepStep = 0
	s.referenceHash = 0
	s.selected = 0
	s.rollbackStep = -1
	s.result = rollbackRunning
	s.done = false

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
	if s.referenceHash == 0 {
		imText("reference run: step %d", s.stepCount)
		return true
	}

	imText("reference sleep step %d", s.referenceSleepStep)
	imText("reference hash 0x%08X", s.referenceHash)

	totalSize := 0
	for _, snapshot := range s.snapshots {
		totalSize += len(snapshot.image)
	}
	imTextDisabled("%d snapshots, %.0f kB", len(s.snapshots), float32(totalSize)/1024.0)

	if imgui.BeginListBoxV(
		"##snapshots",
		imgui.Vec2{X: -fltMin, Y: 8.0 * imgui.TextLineHeightWithSpacing()},
	) {
		for i := range s.snapshots {
			label := fmt.Sprintf(
				"step %d, %d kB",
				s.snapshots[i].step,
				len(s.snapshots[i].image)/1024,
			)

			if imgui.SelectableBoolV(label, i == s.selected, 0, imgui.Vec2{}) {
				s.selected = i
			}
		}

		imgui.EndListBox()
	}

	if imgui.ButtonV("Roll Back", imgui.Vec2{X: 8.0 * imgui.FontSize(), Y: 0.0}) {
		s.RollBack(s.selected)
	}

	switch s.result {
	case rollbackRunning:
		imText("resimulating from step %d", s.rollbackStep)
	case rollbackMatch:
		imTextColored(
			imgui.Vec4{X: 0.4, Y: 0.8, Z: 0.4, W: 1.0},
			"match: sleep step %d",
			s.data.SleepStep,
		)
		imTextColored(imgui.Vec4{X: 0.4, Y: 0.8, Z: 0.4, W: 1.0}, "match: hash 0x%08X", s.data.Hash)
	default:
		imTextColored(
			imgui.Vec4{X: 0.9, Y: 0.3, Z: 0.3, W: 1.0},
			"MISMATCH: sleep step %d",
			s.data.SleepStep,
		)
		imTextColored(
			imgui.Vec4{X: 0.9, Y: 0.3, Z: 0.3, W: 1.0},
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

	if s.referenceHash == 0 {
		s.DrawScreenTextLine("reference run")
	} else if s.result == rollbackRunning {
		s.DrawScreenTextLine("rolled back to step %d", s.rollbackStep)
	} else {
		s.DrawScreenTextLine(
			"sleep step = %d, hash = 0x%08X%s",
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

func init() {
	RegisterSample("Determinism", "Falling Hinges", CreateFallingHingesSample)
	RegisterSample("Determinism", "Rollback", CreateRollback)
}
