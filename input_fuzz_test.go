// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/zigai/box2d"
)

// Native Go format validation is supplemental, not interchange with C's native image.
// Only known-good serialized state is truncated; arbitrary ids/physics payloads are
// never dispatched into APIs whose preconditions require valid objects.
func malformedInputSource(t testing.TB) (box2d.World, []byte, []byte) {
	t.Helper()
	def := box2d.DefaultWorldDef()
	def.WorkerCount = 4
	world := box2d.CreateWorld(&def)
	t.Cleanup(func() {
		if world.IsValid() {
			world.Destroy()
		}
	})
	bodyDef := box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{X: 2, Y: 4}
	body := world.CreateBody(&bodyDef)
	shapeDef := box2d.DefaultShapeDef()
	box := box2d.MakeBox(0.5, 0.5)
	body.CreatePolygonShape(&shapeDef, &box)
	world.Step(1.0/60.0, 4)
	image := world.Snapshot()
	recording := box2d.CreateRecording(0)
	world.StartRecording(recording)
	world.StopRecording()
	data := bytes.Clone(recording.Data())
	recording.Destroy()
	return world, image, data
}

func requireSnapshotCategory(t testing.TB, err error, category error) {
	t.Helper()
	var snapshotError *box2d.SnapshotError
	if !errors.As(err, &snapshotError) || snapshotError.Err == nil {
		t.Fatalf("expected typed snapshot error, got %T: %v", err, err)
	}
	if category != nil && !errors.Is(err, category) {
		t.Fatalf("snapshot category: got %v, want %v", err, category)
	}
}

func checkMalformedSnapshotHeader(t testing.TB, selector uint8, delta uint32) {
	t.Helper()
	world, image, _ := malformedInputSource(t)
	before := world.StateHash()
	bad := bytes.Clone(image)
	offset := int(selector%3) * 4
	binary.LittleEndian.PutUint32(bad[offset:], binary.LittleEndian.Uint32(bad[offset:])^(delta|1))
	original := bytes.Clone(bad)
	category := []error{box2d.ErrBadMagic, box2d.ErrVersion, box2d.ErrLayout}[selector%3]
	err := world.Restore(bad)
	requireSnapshotCategory(t, err, category)
	if world.StateHash() != before {
		t.Fatal("header rejection changed populated world")
	}
	failed, err := box2d.CreateWorldFromSnapshot(bad, 4)
	if failed.IsValid() {
		failed.Destroy()
		t.Fatal("bad header returned valid world")
	}
	requireSnapshotCategory(t, err, category)
	if !bytes.Equal(bad, original) {
		t.Fatal("snapshot rejection mutated input bytes")
	}
	clone, err := box2d.CreateWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("independent valid construction after rejection: %v", err)
	}
	defer clone.Destroy()
	world.Step(1.0/60.0, 4)
	clone.Step(1.0/60.0, 4)
	if clone.StateHash() != world.StateHash() {
		t.Fatal("rejection poisoned subsequent native simulation")
	}
}

func checkMalformedSavedPayload(t testing.TB, cut uint32) {
	t.Helper()
	world, image, recording := malformedInputSource(t)
	before := world.StateHash()
	// Every generated input retains its native header but ends before the complete
	// payload. This bounds allocations by the one-body image and exercises all
	// count/element length boundaries without inventing arbitrary native objects.
	end := 16 + int(cut%uint32(len(image)-16))
	bad := bytes.Clone(image[:end])
	original := bytes.Clone(bad)
	failed, err := box2d.CreateWorldFromSnapshot(bad, 4)
	if failed.IsValid() {
		failed.Destroy()
		t.Fatalf("truncated payload accepted at byte %d/%d", end, len(image))
	}
	requireSnapshotCategory(t, err, nil)
	if !bytes.Equal(bad, original) {
		t.Fatal("constructor changed malformed input bytes")
	}

	// In-place rebuilds retain their original world until the caller destroys it.
	// Teardown must also tolerate a partially replaced id pool or solver-set array.
	restored, err := box2d.CreateWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("valid restore target construction: %v", err)
	}
	t.Cleanup(func() {
		if restored.IsValid() {
			restored.Destroy()
		}
	})
	requireSnapshotCategory(t, restored.Restore(bad), nil)
	restored.Destroy()
	if restored.IsValid() {
		t.Fatal("failed restore cleanup left world valid")
	}
	if !bytes.Equal(bad, original) {
		t.Fatal("restore changed malformed input bytes")
	}

	// Replay has no dedicated typed header error, but embedded snapshot failure
	// must preserve the SnapshotError cause and restore its global length scale.
	data := append(bytes.Clone(recording[:32]), bad...)
	binary.LittleEndian.PutUint64(data[24:], uint64(len(bad)))
	scale := box2d.GetLengthUnitsPerMeter()
	binary.LittleEndian.PutUint32(data[12:], math.Float32bits(scale*2))
	originalReplay := bytes.Clone(data)
	player, err := box2d.CreateReplay(data, 4)
	if player != nil {
		player.Destroy()
		t.Fatal("truncated embedded snapshot returned usable player")
	}
	requireSnapshotCategory(t, err, nil)
	if box2d.GetLengthUnitsPerMeter() != scale {
		t.Fatal("failed replay leaked recording length scale")
	}
	if !bytes.Equal(data, originalReplay) {
		t.Fatal("failed replay changed input bytes")
	}
	if world.StateHash() != before {
		t.Fatal("failed constructors changed independent live world")
	}

	valid, err := box2d.CreateReplay(recording, 4)
	if err != nil {
		t.Fatalf("valid replay after failure: %v", err)
	}
	valid.WorldID().Step(1.0/60.0, 4)
	world.Step(1.0/60.0, 4)
	if valid.WorldID().StateHash() != world.StateHash() {
		valid.Destroy()
		t.Fatal("failed replay poisoned independent simulation")
	}
	valid.Destroy()
	if box2d.GetLengthUnitsPerMeter() != scale {
		t.Fatal("valid replay cleanup changed length scale")
	}
}

func TestMalformedSavedInputCorpus(t *testing.T) {
	for _, seed := range []struct {
		selector uint8
		delta    uint32
	}{{0, 1}, {1, 0xffffffff}, {2, 0x80000000}} {
		t.Run([]string{"magic", "version", "layout"}[seed.selector], func(t *testing.T) {
			checkMalformedSnapshotHeader(t, seed.selector, seed.delta)
		})
	}
	for _, cut := range []uint32{0, 1, 15, 31, 63, 127, 255, 511, 1023, 0xffffffff} {
		t.Run("truncated-payload", func(t *testing.T) { checkMalformedSavedPayload(t, cut) })
	}
}

func TestMalformedSavedPayloadReleasesWorldSlots(t *testing.T) {
	world, image, _ := malformedInputSource(t)
	bad := bytes.Clone(image[:16])
	original := bytes.Clone(bad)
	before := world.StateHash()
	scale := box2d.GetLengthUnitsPerMeter()
	// More failures than available slots detect retained worlds even when the
	// failed constructor returns the correct zero World and typed error.
	for attempt := range box2d.MaxWorlds + 1 {
		failed, err := box2d.CreateWorldFromSnapshot(bad, 4)
		if failed.IsValid() {
			failed.Destroy()
			t.Fatalf("attempt %d returned usable world", attempt)
		}
		requireSnapshotCategory(t, err, nil)
	}
	if !bytes.Equal(bad, original) || world.StateHash() != before || box2d.GetLengthUnitsPerMeter() != scale {
		t.Fatal("repeated failed constructors changed input, independent world, or length scale")
	}
	valid, err := box2d.CreateWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("valid constructor after repeated failures: %v", err)
	}
	defer valid.Destroy()
	valid.Step(1.0/60.0, 4)
	world.Step(1.0/60.0, 4)
	if valid.StateHash() != world.StateHash() {
		t.Fatal("repeated failed constructors poisoned subsequent simulation")
	}
}

func FuzzMalformedSnapshotHeader(f *testing.F) {
	f.Add(uint8(0), uint32(1))
	f.Add(uint8(1), uint32(0xffffffff))
	f.Add(uint8(2), uint32(0x80000000))
	f.Fuzz(func(t *testing.T, selector uint8, delta uint32) {
		checkMalformedSnapshotHeader(t, selector, delta)
	})
}

func FuzzMalformedSavedPayload(f *testing.F) {
	for _, cut := range []uint32{0, 1, 15, 31, 63, 127, 255, 511, 1023, 0xffffffff} {
		f.Add(cut)
	}
	f.Fuzz(func(t *testing.T, cut uint32) { checkMalformedSavedPayload(t, cut) })
}

// Panic metadata and build-tag behavior are Go-defined evidence. The paired
// TestInvalidInput retains the C guard return/state/log contract.
func TestInputGuardAssertionErrorNative(t *testing.T) {
	restore := box2d.InputSaveInputHooks()
	t.Cleanup(restore)
	world, _, _ := malformedInputSource(t)
	before, gravity := world.StateHash(), world.Gravity()
	for _, bad := range inputBadValues {
		logs, assertions := 0, 0
		var condition, file string
		var line int
		box2d.SetLogFunc(func(string) { logs++ })
		box2d.SetAssertFunc(func(c, f string, l int) bool {
			assertions++
			condition, file, line = c, f, l
			return true
		})
		var caught any
		func() {
			defer func() { caught = recover() }()
			world.SetGravity(box2d.Vec2{X: bad})
		}()
		if logs != 1 {
			t.Fatalf("guard log multiplicity: %d, want 1", logs)
		}
		if box2d.InputAssertsEnabled() {
			assertion, ok := caught.(*box2d.AssertError)
			if !ok {
				t.Fatalf("guard panic: got %T, want *AssertError", caught)
			}
			if assertions != 1 || condition == "" || file == "" || line <= 0 ||
				assertion.Condition != condition || assertion.File != file || assertion.Line != line {
				t.Fatalf("assertion metadata/count: %#v, count=%d", assertion, assertions)
			}
		} else if caught != nil || assertions != 0 {
			t.Fatalf("noassert guard invoked assertion: panic=%v calls=%d", caught, assertions)
		}
		if world.StateHash() != before || world.Gravity() != gravity {
			t.Fatal("rejected guarded setter changed world")
		}
	}
	restore()
	world.Step(1.0/60.0, 4)
	if world.StateHash() == before {
		t.Fatal("valid step did not advance populated world after guard failures")
	}
}
