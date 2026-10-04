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
func malformedInputSource(t testing.TB) (*box2d.World, []byte, []byte) {
	t.Helper()
	def := box2d.DefaultWorldDef()
	def.WorkerCount = 4
	world := box2d.NewWorld(&def)
	t.Cleanup(func() {
		if world.IsValid() {
			world.Close()
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
	image := box2d.BTSnapshot(world)
	recording := box2d.NewRecording(0)
	world.StartRecording(recording)
	world.StopRecording()
	data := bytes.Clone(recording.Data())
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
	err := world.UnmarshalBinary(bad)
	requireSnapshotCategory(t, err, category)
	if world.StateHash() != before {
		t.Fatal("header rejection changed populated world")
	}
	failed, err := box2d.NewWorldFromSnapshot(bad, 4)
	if failed.IsValid() {
		failed.Close()
		t.Fatal("bad header returned valid world")
	}
	requireSnapshotCategory(t, err, category)
	if !bytes.Equal(bad, original) {
		t.Fatal("snapshot rejection mutated input bytes")
	}
	clone, err := box2d.NewWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("independent valid construction after rejection: %v", err)
	}
	defer clone.Close()
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
	failed, err := box2d.NewWorldFromSnapshot(bad, 4)
	if failed.IsValid() {
		failed.Close()
		t.Fatalf("truncated payload accepted at byte %d/%d", end, len(image))
	}
	requireSnapshotCategory(t, err, nil)
	if !bytes.Equal(bad, original) {
		t.Fatal("constructor changed malformed input bytes")
	}

	// In-place rebuilds retain their original world until the caller destroys it.
	// Teardown must also tolerate a partially replaced id pool or solver-set array.
	restored, err := box2d.NewWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("valid restore target construction: %v", err)
	}
	t.Cleanup(func() {
		if restored.IsValid() {
			restored.Close()
		}
	})
	requireSnapshotCategory(t, restored.UnmarshalBinary(bad), nil)
	restored.Close()
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
	scale := box2d.LengthUnitsPerMeter()
	binary.LittleEndian.PutUint32(data[12:], math.Float32bits(scale*2))
	originalReplay := bytes.Clone(data)
	player, err := box2d.NewReplay(data, 4)
	if player != nil {
		player.Close()
		t.Fatal("truncated embedded snapshot returned usable player")
	}
	requireSnapshotCategory(t, err, nil)
	if box2d.LengthUnitsPerMeter() != scale {
		t.Fatal("failed replay leaked recording length scale")
	}
	if !bytes.Equal(data, originalReplay) {
		t.Fatal("failed replay changed input bytes")
	}
	if world.StateHash() != before {
		t.Fatal("failed constructors changed independent live world")
	}

	valid, err := box2d.NewReplay(recording, 4)
	if err != nil {
		t.Fatalf("valid replay after failure: %v", err)
	}
	valid.World().Step(1.0/60.0, 4)
	world.Step(1.0/60.0, 4)
	if valid.World().StateHash() != world.StateHash() {
		valid.Close()
		t.Fatal("failed replay poisoned independent simulation")
	}
	valid.Close()
	if box2d.LengthUnitsPerMeter() != scale {
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

func TestMalformedSavedPayloadLeavesWorldsIntact(t *testing.T) {
	world, image, _ := malformedInputSource(t)
	bad := bytes.Clone(image[:16])
	original := bytes.Clone(bad)
	before := world.StateHash()
	scale := box2d.LengthUnitsPerMeter()
	failed, err := box2d.NewWorldFromSnapshot(bad, 4)
	if failed.IsValid() {
		failed.Close()
		t.Fatal("failed constructor returned usable world")
	}
	requireSnapshotCategory(t, err, nil)
	if !bytes.Equal(bad, original) || world.StateHash() != before || box2d.LengthUnitsPerMeter() != scale {
		t.Fatal("failed constructor changed input, independent world, or length scale")
	}
	valid, err := box2d.NewWorldFromSnapshot(image, 4)
	if err != nil {
		t.Fatalf("valid constructor after failure: %v", err)
	}
	defer valid.Close()
	valid.Step(1.0/60.0, 4)
	world.Step(1.0/60.0, 4)
	if valid.StateHash() != world.StateHash() {
		t.Fatal("failed constructor poisoned subsequent simulation")
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

// Panic metadata is Go-defined evidence. The paired TestInvalidInput retains the C
// guard rejection and state contract.
func TestInputGuardAssertionErrorNative(t *testing.T) {
	world, _, _ := malformedInputSource(t)
	before, gravity := world.StateHash(), world.Gravity()
	for _, bad := range inputBadValues {
		var caught any
		func() {
			defer func() { caught = recover() }()
			world.SetGravity(box2d.Vec2{X: bad})
		}()
		assertion, ok := caught.(*box2d.AssertError)
		if !ok {
			t.Fatalf("guard panic: got %T, want *AssertError", caught)
		}
		if assertion.Condition == "" || assertion.File == "" || assertion.Line <= 0 {
			t.Fatalf("assertion metadata: %#v", assertion)
		}
		if world.StateHash() != before || world.Gravity() != gravity {
			t.Fatal("rejected guarded setter changed world")
		}
	}
	world.Step(1.0/60.0, 4)
	if world.StateHash() == before {
		t.Fatal("valid step did not advance populated world after guard failures")
	}
}
