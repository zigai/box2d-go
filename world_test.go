// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"github.com/zigai/box2d/internal/scenes"
)

const worldBodyCount = 10
const worldWorldCount = box2d.MaxWorlds / 2

type worldPairBox struct {
	BodyID  box2d.Body
	ShapeID box2d.Shape
}
type worldOracle struct {
	*oracle.Reader
	name     string
	t        *testing.T
	expected []uint64
	hashes   []uint64
}

func worldPackSteps(hashes []uint64) []byte {
	var packed []byte
	for i := 0; i < len(hashes); {
		end := i + 1
		for end < len(hashes) && hashes[end] == hashes[i] {
			end++
		}
		packed = binary.LittleEndian.AppendUint32(packed, uint32(end-i))
		packed = binary.LittleEndian.AppendUint64(packed, hashes[i])
		i = end
	}
	return packed
}
func worldUnpackSteps(packed []byte) ([]uint64, error) {
	if len(packed)%12 != 0 {
		return nil, fmt.Errorf("incomplete RLE record: %d bytes", len(packed))
	}
	var hashes []uint64
	for len(packed) != 0 {
		count := binary.LittleEndian.Uint32(packed[:4])
		if count == 0 || count > 1048576 || uint64(len(hashes))+uint64(count) > 1048576 {
			return nil, fmt.Errorf("invalid RLE run count: %d", count)
		}
		hash := binary.LittleEndian.Uint64(packed[4:12])
		for range count {
			hashes = append(hashes, hash)
		}
		packed = packed[12:]
	}
	return hashes, nil
}
func worldOpen(t *testing.T, name string) *worldOracle {
	t.Helper()
	o := &worldOracle{Reader: oracle.Open(t, "world/"+name), name: name, t: t}
	config := "float"
	if reflect.TypeFor[oracle.Real]().Bits() == 64 {
		config = "double"
	}
	data, err := os.ReadFile("testdata/oracle/" + config + "/world/" + name + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var packed []byte
	found := false
	for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
		var record struct {
			Key  string `json:"k"`
			Type string `json:"t"`
			Val  string `json:"v"`
		}
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatal(err)
		}
		if record.Key == "steps/rle-u32-u64-le" {
			if found || record.Type != "bytes" {
				t.Fatal("invalid step trace record")
			}
			found = true
			packed, err = hex.DecodeString(record.Val)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if !found {
		t.Fatal("missing step trace record")
	}
	o.expected, err = worldUnpackSteps(packed)
	if err != nil {
		t.Fatal(err)
	}
	return o
}
func (o *worldOracle) finish(t *testing.T) {
	t.Helper()
	if len(o.expected) != len(o.hashes) {
		t.Fatalf("world/%s: first divergent source step %d: got %d steps, C oracle %d", o.name, min(len(o.expected), len(o.hashes))+1, len(o.hashes), len(o.expected))
	}
	o.Bytes("steps/rle-u32-u64-le", worldPackSteps(o.hashes))
}
func (o *worldOracle) step(w box2d.World, source string) {
	transforms := box2d.WorldBodyTransforms(w)
	hash := oracle.HashInit
	for _, body := range transforms {
		transform := body.Transform
		hash = oracle.HashTransform(hash, transform.P.X, transform.P.Y, transform.Q.C, transform.Q.S)
	}
	o.hashes = append(o.hashes, hash)
	index := len(o.hashes) - 1
	if index >= len(o.expected) {
		o.t.Fatalf("world/%s: unexpected source step %d (%s)", o.name, index+1, source)
	}
	if hash != o.expected[index] {
		o.t.Fatalf("world/%s: first divergent source step %d (%s): transform hash got %016x, C oracle %016x", o.name, index+1, source, hash, o.expected[index])
	}
	if len(o.hashes)%60 == 0 {
		prefix := fmt.Sprintf("step/%d/%s/", len(o.hashes), source)
		o.I32(prefix+"bodyCount", int32(len(transforms)))
		for _, body := range transforms {
			transform := body.Transform
			o.I32(prefix+"body", body.Index)
			key := fmt.Sprintf("%sbody/%d/", prefix, body.Index)
			o.Pos(key+"position", transform.P.X, transform.P.Y)
			o.Rot(key+"rotation", transform.Q.C, transform.Q.S)
		}
	}
}
func worldContactData(b box2d.Body, data *[4]box2d.ContactData) int {
	values := slices.Collect(b.ContactData())
	count := copy(data[:], values)
	return count
}

func TestWorld(t *testing.T) {
	defer porttest.Guard(t)
	t.Run("HelloWorld", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "HelloWorld")
		porttest.Ensure(t, worldHelloWorld(t, o) == 0)
		o.finish(t)
	})
	t.Run("EmptyWorld", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "EmptyWorld")
		porttest.Ensure(t, worldEmptyWorld(t, o) == 0)
		o.finish(t)
	})
	t.Run("DestroyAllBodiesWorld", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "DestroyAllBodiesWorld")
		porttest.Ensure(t, worldDestroyAllBodiesWorld(t, o) == 0)
		o.finish(t)
	})
	t.Run("TestIsValid", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "TestIsValid")
		porttest.Ensure(t, worldTestIsValid(t, o) == 0)
		o.finish(t)
	})
	t.Run("TestWorldRecycle", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "TestWorldRecycle")
		porttest.Ensure(t, worldTestWorldRecycle(t, o) == 0)
		o.finish(t)
	})
	t.Run("TestWorldCoverage", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "TestWorldCoverage")
		porttest.Ensure(t, worldTestWorldCoverage(t, o) == 0)
		o.finish(t)
	})
	t.Run("TestSensor", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "TestSensor")
		porttest.Ensure(t, worldTestSensor(t, o) == 0)
		o.finish(t)
	})
	t.Run("TestSetWorkerCount", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "TestSetWorkerCount")
		porttest.Ensure(t, worldTestSetWorkerCount(t, o) == 0)
		o.finish(t)
	})
	t.Run("ChainSegmentShapeTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "ChainSegmentShapeTest")
		porttest.Ensure(t, worldChainSegmentShapeTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("SetBulletDriftTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "SetBulletDriftTest")
		porttest.Ensure(t, worldSetBulletDriftTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("DeferredMassFlagSyncTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "DeferredMassFlagSyncTest")
		porttest.Ensure(t, worldDeferredMassFlagSyncTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("EnableSleepFlagSyncTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "EnableSleepFlagSyncTest")
		porttest.Ensure(t, worldEnableSleepFlagSyncTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("EnableContactRecyclingTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "EnableContactRecyclingTest")
		porttest.Ensure(t, worldEnableContactRecyclingTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("EnlargedProxyDestroyedTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "EnlargedProxyDestroyedTest")
		porttest.Ensure(t, worldEnlargedProxyDestroyedTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("BroadPhasePairsTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "BroadPhasePairsTest")
		porttest.Ensure(t, worldBroadPhasePairsTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("BodySimLocatorTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "BodySimLocatorTest")
		porttest.Ensure(t, worldBodySimLocatorTest(t, o) == 0)
		o.finish(t)
	})
	t.Run("NonDynamicJointDisableTest", func(t *testing.T) {
		defer porttest.Guard(t)
		o := worldOpen(t, "NonDynamicJointDisableTest")
		porttest.Ensure(t, worldNonDynamicJointDisableTest(t, o) == 0)
		o.finish(t)
	})
}

func worldHelloWorld(t *testing.T, o *worldOracle) int {
	// Construct a world object, which will hold and simulate the rigid bodies.
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{0.0, -10.0}

	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid())
	o.Bool("L30", worldId.IsValid())

	// Define the ground body.
	var groundBodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	groundBodyDef.Position = box2d.Pos{oracle.Real(float32(0.0)), oracle.Real(float32(-10.0))}

	// Call the body factory which allocates memory for the ground body
	// from a pool and creates the ground box shape (also from a pool).
	// The body is also added to the world.
	var groundId box2d.Body = worldId.CreateBody(&groundBodyDef)
	porttest.Ensure(t, groundId.IsValid())
	o.Bool("L40", groundId.IsValid())

	// Define the ground box shape. The extents are the half-widths of the box.
	var groundBox box2d.Polygon = box2d.MakeBox(50.0, 10.0)

	// Add the box shape to the ground body.
	var groundShapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	groundId.CreatePolygonShape(&groundShapeDef, &groundBox)

	// Define the dynamic body. We set its position and call the body factory.
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{oracle.Real(float32(0.0)), oracle.Real(float32(4.0))}

	var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

	// Define another box shape for our dynamic body.
	var dynamicBox box2d.Polygon = box2d.MakeBox(1.0, 1.0)

	// Define the dynamic body shape
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	// Set the box density to be non-zero, so it will be dynamic.
	shapeDef.Density = 1.0

	// Override the default friction.
	shapeDef.Material.Friction = 0.3

	// Add the shape to the body.
	bodyId.CreatePolygonShape(&shapeDef, &dynamicBox)

	// Prepare for simulation. Typically we use a time step of 1/60 of a
	// second (60Hz) and 4 sub-steps. This provides a high quality simulation
	// in most game scenarios.
	var timeStep float32 = 1.0 / 60.0
	var subStepCount int = 4

	var position box2d.Pos = bodyId.Position()
	var rotation box2d.Rot = bodyId.Rotation()
	o.Pos("hello.position", position.X, position.Y)
	o.Rot("hello.rotation", rotation.C, rotation.S)

	// This is our little game loop.
	for i := 0; i < 90; i++ {
		// Instruct the world to perform a single step of simulation.
		// It is generally best to keep the time step and iterations fixed.
		worldId.Step(timeStep, subStepCount)
		o.step(worldId, "L85")

		// Now print the position and angle of the body.
		position = bodyId.Position()
		rotation = bodyId.Rotation()
		o.Pos("hello.position", position.X, position.Y)
		o.Rot("hello.rotation", rotation.C, rotation.S)

		// printf("%4.2f %4.2f %4.2f\n", position.X, position.Y, rotation.Angle());
	}

	// When the world destructor is called, all bodies and joints are freed. This can
	// create orphaned ids, so be careful about your world management.
	worldId.Destroy()

	porttest.Ensure(t, box2d.AbsFloat(float32(position.X)) < 0.01)
	o.Bool("L98", box2d.AbsFloat(float32(position.X)) < 0.01)
	o.F32("L98:value", box2d.AbsFloat(float32(position.X)))
	porttest.Ensure(t, box2d.AbsFloat(float32(position.Y-1.00)) < 0.01)
	o.Bool("L99", box2d.AbsFloat(float32(position.Y-1.00)) < 0.01)
	o.F32("L99:value", box2d.AbsFloat(float32(position.Y-1.00)))
	porttest.Ensure(t, box2d.AbsFloat(rotation.Angle()) < 0.01)
	o.Bool("L100", box2d.AbsFloat(rotation.Angle()) < 0.01)
	o.F32("L100:value", box2d.AbsFloat(rotation.Angle()))

	return 0
}

func worldEmptyWorld(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid() == true)
	o.Bool("L109", worldId.IsValid() == true)

	var timeStep float32 = 1.0 / 60.0
	var subStepCount int = 1

	for i := 0; i < 60; i++ {
		worldId.Step(timeStep, subStepCount)
		o.step(worldId, "L116")
	}

	worldId.Destroy()

	porttest.Ensure(t, worldId.IsValid() == false)
	o.Bool("L121", worldId.IsValid() == false)

	return 0
}

func worldDestroyAllBodiesWorld(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid() == true)
	o.Bool("L131", worldId.IsValid() == true)

	var count int = 0
	var creating bool = true

	var bodyIds [worldBodyCount]box2d.Body
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var square box2d.Polygon = box2d.MakeSquare(0.5)

	for i := 0; i < 2*worldBodyCount+10; i++ {
		if creating {
			if count < worldBodyCount {
				bodyIds[count] = worldId.CreateBody(&bodyDef)

				var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
				bodyIds[count].CreatePolygonShape(&shapeDef, &square)
				count += 1
			} else {
				creating = false
			}
		} else if count > 0 {
			bodyIds[count-1].Destroy()
			bodyIds[count-1] = box2d.Body{}
			count -= 1
		}

		worldId.Step(1.0/60.0, 3)
		o.step(worldId, "L165")
	}

	var counters box2d.Counters = worldId.Counters()
	porttest.Ensure(t, counters.BodyCount == 0)
	o.Bool("L169", counters.BodyCount == 0)
	o.I32("L169:value", int32(counters.BodyCount))

	worldId.Destroy()

	porttest.Ensure(t, worldId.IsValid() == false)
	o.Bool("L173", worldId.IsValid() == false)

	return 0
}

func worldTestIsValid(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid())
	o.Bool("L182", worldId.IsValid())

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()

	var bodyId1 box2d.Body = worldId.CreateBody(&bodyDef)
	porttest.Ensure(t, bodyId1.IsValid() == true)
	o.Bool("L187", bodyId1.IsValid() == true)

	var bodyId2 box2d.Body = worldId.CreateBody(&bodyDef)
	porttest.Ensure(t, bodyId2.IsValid() == true)
	o.Bool("L190", bodyId2.IsValid() == true)

	bodyId1.Destroy()
	porttest.Ensure(t, bodyId1.IsValid() == false)
	o.Bool("L193", bodyId1.IsValid() == false)

	bodyId2.Destroy()
	porttest.Ensure(t, bodyId2.IsValid() == false)
	o.Bool("L196", bodyId2.IsValid() == false)

	worldId.Destroy()

	porttest.Ensure(t, worldId.IsValid() == false)
	o.Bool("L200", worldId.IsValid() == false)
	porttest.Ensure(t, bodyId2.IsValid() == false)
	o.Bool("L201", bodyId2.IsValid() == false)
	porttest.Ensure(t, bodyId1.IsValid() == false)
	o.Bool("L202", bodyId1.IsValid() == false)

	return 0
}

func worldTestWorldRecycle(t *testing.T, o *worldOracle) int {

	var count int = 100

	var worldIds [worldWorldCount]box2d.World

	for i := 0; i < count; i++ {
		var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
		for j := 0; j < worldWorldCount; j++ {
			worldIds[j] = box2d.CreateWorld(&worldDef)
			box2d.BTWorldCleanup(t, worldIds[j])
			porttest.Ensure(t, worldIds[j].IsValid() == true)
			o.Bool("L223", worldIds[j].IsValid() == true)

			var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
			worldIds[j].CreateBody(&bodyDef)
		}

		for j := 0; j < worldWorldCount; j++ {
			var timeStep float32 = 1.0 / 60.0
			var subStepCount int = 1

			for k := 0; k < 10; k++ {
				worldIds[j].Step(timeStep, subStepCount)
				o.step(worldIds[j], "L236")
			}
		}

		for j := worldWorldCount - 1; j >= 0; j-- {
			worldIds[j].Destroy()
			porttest.Ensure(t, worldIds[j].IsValid() == false)
			o.Bool("L243", worldIds[j].IsValid() == false)
			worldIds[j] = box2d.World{}
		}
	}

	return 0
}

func worldCustomFilter(t *testing.T, o *worldOracle, shapeIdA box2d.Shape, shapeIdB box2d.Shape) bool {
	porttest.Ensure(t, true)
	o.Bool("L255", true)
	return true
}

func worldPreSolveStatic(t *testing.T, o *worldOracle, shapeIdA box2d.Shape, shapeIdB box2d.Shape, manifold *box2d.Manifold) {
}

// This test is here to ensure all API functions link correctly.

func worldTestWorldCoverage(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()

	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid())
	o.Bool("L273", worldId.IsValid())

	worldId.EnableSleeping(true)
	worldId.EnableSleeping(false)
	var flag bool = worldId.IsSleepingEnabled()
	porttest.Ensure(t, flag == false)
	o.Bool("L278", flag == false)

	worldId.EnableContinuous(false)
	worldId.EnableContinuous(true)
	flag = worldId.IsContinuousEnabled()
	porttest.Ensure(t, flag == true)
	o.Bool("L283", flag == true)

	worldId.SetRestitutionThreshold(0.0)
	worldId.SetRestitutionThreshold(2.0)
	var value float32 = worldId.RestitutionThreshold()
	porttest.Ensure(t, value == 2.0)
	o.Bool("L288", value == 2.0)
	o.F32("L288:value", value)

	worldId.SetHitEventThreshold(0.0)
	worldId.SetHitEventThreshold(100.0)
	value = worldId.HitEventThreshold()
	porttest.Ensure(t, value == 100.0)
	o.Bool("L293", value == 100.0)
	o.F32("L293:value", value)

	worldId.SetCustomFilterCallback(func(a, b box2d.Shape) bool { return worldCustomFilter(t, o, a, b) })
	worldId.SetPreSolveCallback(func(a, b box2d.Shape, m *box2d.Manifold) { worldPreSolveStatic(t, o, a, b, m) }, nil)

	var g box2d.Vec2 = box2d.Vec2{1.0, 2.0}
	worldId.SetGravity(g)
	var v box2d.Vec2 = worldId.Gravity()
	porttest.Ensure(t, v.X == g.X)
	o.Bool("L301", v.X == g.X)
	o.F32("L301:value", v.X)
	porttest.Ensure(t, v.Y == g.Y)
	o.Bool("L302", v.Y == g.Y)
	o.F32("L302:value", v.Y)

	var explosionDef box2d.ExplosionDef = box2d.DefaultExplosionDef()
	worldId.Explode(&explosionDef)

	worldId.SetContactTuning(10.0, 2.0, 4.0)

	worldId.SetMaximumLinearSpeed(10.0)
	value = worldId.MaximumLinearSpeed()
	porttest.Ensure(t, value == 10.0)
	o.Bool("L311", value == 10.0)
	o.F32("L311:value", value)

	worldId.EnableWarmStarting(true)
	flag = worldId.IsWarmStartingEnabled()
	porttest.Ensure(t, flag == true)
	o.Bool("L315", flag == true)

	var count int = worldId.AwakeBodyCount()
	porttest.Ensure(t, count == 0)
	o.Bool("L318", count == 0)
	o.I32("L318:value", int32(count))

	worldId.SetUserData(&value)
	var userData any = worldId.UserData()
	porttest.Ensure(t, userData == &value)
	o.Bool("L322", userData == &value)

	worldId.Step(1.0, 1)
	o.step(worldId, "L324")

	worldId.Destroy()

	return 0
}

func worldTestSensor(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	// Wall from x = 1 to x = 2
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.StaticBody
	bodyDef.Position.X = 1.5
	bodyDef.Position.Y = 11.0
	var wallId box2d.Body = worldId.CreateBody(&bodyDef)
	var box box2d.Polygon = box2d.MakeBox(0.5, 10.0)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.EnableSensorEvents = true
	wallId.CreatePolygonShape(&shapeDef, &box)

	// Bullet fired towards the wall
	bodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.IsBullet = true
	bodyDef.GravityScale = 0.0
	bodyDef.Position = box2d.Pos{oracle.Real(float32(7.39814)), oracle.Real(float32(4.0))}
	bodyDef.LinearVelocity = box2d.Vec2{-20.0, 0.0}
	var bulletId box2d.Body = worldId.CreateBody(&bodyDef)
	shapeDef = box2d.DefaultShapeDef()
	shapeDef.IsSensor = true
	shapeDef.EnableSensorEvents = true
	var circle box2d.Circle = box2d.Circle{box2d.Vec2{0.0, 0.0}, 0.1}
	bulletId.CreateCircleShape(&shapeDef, &circle)

	var beginCount int = 0
	var endCount int = 0

	for {
		var timeStep float32 = 1.0 / 60.0
		var subStepCount int = 4
		worldId.Step(timeStep, subStepCount)
		o.step(worldId, "L368")

		var bulletPos box2d.Pos = bulletId.Position()
		// printf( "Bullet pos: %g %g\n", bulletPos.X, bulletPos.Y );

		beginEvents := slices.Collect(worldId.SensorBeginEvents())
		endEvents := slices.Collect(worldId.SensorEndEvents())
		o.I32("sensor.beginCount", int32(len(beginEvents)))
		for _, event := range beginEvents {
			o.I32("sensor.begin.sensor", box2d.WorldShapeIndex(event.SensorShapeID))
			o.I32("sensor.begin.visitor", box2d.WorldShapeIndex(event.VisitorShapeID))
		}
		o.I32("sensor.endCount", int32(len(endEvents)))
		for _, event := range endEvents {
			o.I32("sensor.end.sensor", box2d.WorldShapeIndex(event.SensorShapeID))
			o.I32("sensor.end.visitor", box2d.WorldShapeIndex(event.VisitorShapeID))
		}
		o.Pos("sensor.position", bulletPos.X, bulletPos.Y)

		if len(beginEvents) > 0 {
			beginCount += 1
		}

		if len(endEvents) > 0 {
			endCount += 1
		}

		if bulletPos.X < -1.0 {
			break
		}
	}

	worldId.Destroy()

	porttest.Ensure(t, beginCount == 1)
	o.Bool("L393", beginCount == 1)
	o.I32("L393:value", int32(beginCount))
	porttest.Ensure(t, endCount == 1)
	o.Bool("L394", endCount == 1)
	o.I32("L394:value", int32(endCount))

	return 0
}

func worldTestSetWorkerCount(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.WorkerCount = 1
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)
	porttest.Ensure(t, worldId.IsValid())
	o.Bool("L404", worldId.IsValid())
	porttest.Ensure(t, worldId.WorkerCount() == 1)
	o.Bool("L405", worldId.WorkerCount() == 1)
	o.I32("L405:value", int32(worldId.WorkerCount()))

	scenes.CreateJunkyard(worldId)
	scenes.StepJunkyard(worldId, 1)
	o.step(worldId, "L408")

	worldId.SetWorkerCount(4)
	porttest.Ensure(t, worldId.WorkerCount() == 4)
	o.Bool("L411", worldId.WorkerCount() == 4)
	o.I32("L411:value", int32(worldId.WorkerCount()))

	scenes.StepJunkyard(worldId, 2)
	o.step(worldId, "L413")

	worldId.SetWorkerCount(4)
	porttest.Ensure(t, worldId.WorkerCount() == 4)
	o.Bool("L416", worldId.WorkerCount() == 4)
	o.I32("L416:value", int32(worldId.WorkerCount()))

	scenes.StepJunkyard(worldId, 3)
	o.step(worldId, "L418")

	worldId.SetWorkerCount(0)
	porttest.Ensure(t, worldId.WorkerCount() == 1)
	o.Bool("L421", worldId.WorkerCount() == 1)
	o.I32("L421:value", int32(worldId.WorkerCount()))

	scenes.StepJunkyard(worldId, 4)
	o.step(worldId, "L423")

	worldId.SetWorkerCount(-5)
	porttest.Ensure(t, worldId.WorkerCount() == 1)
	o.Bool("L426", worldId.WorkerCount() == 1)
	o.I32("L426:value", int32(worldId.WorkerCount()))

	scenes.StepJunkyard(worldId, 5)
	o.step(worldId, "L428")

	worldId.SetWorkerCount(box2d.MaxWorkers + 10)
	porttest.Ensure(t, worldId.WorkerCount() == box2d.MaxWorkers)
	o.Bool("L431", worldId.WorkerCount() == box2d.MaxWorkers)
	o.I32("L431:value", int32(worldId.WorkerCount()))

	scenes.StepJunkyard(worldId, 2)
	o.step(worldId, "L433")

	worldId.Destroy()

	return 0
}

func worldChainSegmentShapeTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{0.0, -10.0}
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	var groundId box2d.Body = worldId.CreateBody(&bodyDef)

	var cs box2d.ChainSegment = box2d.ChainSegment{}
	cs.Ghost1 = box2d.Vec2{2.0, 0.0}
	cs.Segment.Point1 = box2d.Vec2{1.0, 0.0}
	cs.Segment.Point2 = box2d.Vec2{-1.0, 0.0}
	cs.Ghost2 = box2d.Vec2{-2.0, 0.0}
	cs.ChainID = 99

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var orphanShape box2d.Shape = groundId.CreateChainSegmentShape(&shapeDef, &cs)
	porttest.Ensure(t, !orphanShape.IsNull())
	o.Bool("L458", !orphanShape.IsNull())

	porttest.Ensure(t, orphanShape.Type() == box2d.ChainSegmentShape)
	o.Bool("L460", orphanShape.Type() == box2d.ChainSegmentShape)
	o.I32("L460:value", int32(orphanShape.Type()))

	var parentChain box2d.Chain = orphanShape.ParentChain()
	porttest.Ensure(t, parentChain.IsNull())
	o.Bool("L463", parentChain.IsNull())

	var got box2d.ChainSegment = orphanShape.ChainSegment()
	porttest.EnsureSmall(t, got.Ghost1.X-cs.Ghost1.X, 1e-5)
	o.F32("L466", got.Ghost1.X-cs.Ghost1.X)
	porttest.EnsureSmall(t, got.Ghost1.Y-cs.Ghost1.Y, 1e-5)
	o.F32("L467", got.Ghost1.Y-cs.Ghost1.Y)
	porttest.EnsureSmall(t, got.Segment.Point1.X-cs.Segment.Point1.X, 1e-5)
	o.F32("L468", got.Segment.Point1.X-cs.Segment.Point1.X)
	porttest.EnsureSmall(t, got.Segment.Point1.Y-cs.Segment.Point1.Y, 1e-5)
	o.F32("L469", got.Segment.Point1.Y-cs.Segment.Point1.Y)
	porttest.EnsureSmall(t, got.Segment.Point2.X-cs.Segment.Point2.X, 1e-5)
	o.F32("L470", got.Segment.Point2.X-cs.Segment.Point2.X)
	porttest.EnsureSmall(t, got.Segment.Point2.Y-cs.Segment.Point2.Y, 1e-5)
	o.F32("L471", got.Segment.Point2.Y-cs.Segment.Point2.Y)
	porttest.EnsureSmall(t, got.Ghost2.X-cs.Ghost2.X, 1e-5)
	o.F32("L472", got.Ghost2.X-cs.Ghost2.X)
	porttest.EnsureSmall(t, got.Ghost2.Y-cs.Ghost2.Y, 1e-5)
	o.F32("L473", got.Ghost2.Y-cs.Ghost2.Y)
	porttest.Ensure(t, got.ChainID == -1)
	o.Bool("L474", got.ChainID == -1)
	o.I32("L474:value", int32(got.ChainID))

	var dynamicDef box2d.BodyDef = box2d.DefaultBodyDef()
	dynamicDef.Type = box2d.DynamicBody
	dynamicDef.Position = box2d.Pos{oracle.Real(float32(0.0)), oracle.Real(float32(2.0))}
	var circleBodyId box2d.Body = worldId.CreateBody(&dynamicDef)
	var circle box2d.Circle = box2d.Circle{box2d.Vec2{0.0, 0.0}, 0.5}
	var circleShapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	circleBodyId.CreateCircleShape(&circleShapeDef, &circle)

	for i := 0; i < 120; i++ {
		worldId.Step(1.0/60.0, 4)
		o.step(worldId, "L486")
	}

	var circlePos box2d.Pos = circleBodyId.Position()
	porttest.Ensure(t, circlePos.Y > 0.0)
	o.Bool("L490", circlePos.Y > 0.0)
	o.Pos("L490:value", circlePos.X, circlePos.Y)

	var cs2 box2d.ChainSegment = box2d.ChainSegment{}
	cs2.Ghost1 = box2d.Vec2{3.0, 0.0}
	cs2.Segment.Point1 = box2d.Vec2{2.0, 0.0}
	cs2.Segment.Point2 = box2d.Vec2{-2.0, 0.0}
	cs2.Ghost2 = box2d.Vec2{-3.0, 0.0}
	cs2.ChainID = -1

	orphanShape.SetChainSegment(&cs2)

	var got2 box2d.ChainSegment = orphanShape.ChainSegment()
	porttest.EnsureSmall(t, got2.Segment.Point1.X-cs2.Segment.Point1.X, 1e-5)
	o.F32("L502", got2.Segment.Point1.X-cs2.Segment.Point1.X)
	porttest.EnsureSmall(t, got2.Segment.Point1.Y-cs2.Segment.Point1.Y, 1e-5)
	o.F32("L503", got2.Segment.Point1.Y-cs2.Segment.Point1.Y)
	porttest.EnsureSmall(t, got2.Segment.Point2.X-cs2.Segment.Point2.X, 1e-5)
	o.F32("L504", got2.Segment.Point2.X-cs2.Segment.Point2.X)
	porttest.EnsureSmall(t, got2.Segment.Point2.Y-cs2.Segment.Point2.Y, 1e-5)
	o.F32("L505", got2.Segment.Point2.Y-cs2.Segment.Point2.Y)
	porttest.EnsureSmall(t, got2.Ghost1.X-cs2.Ghost1.X, 1e-5)
	o.F32("L506", got2.Ghost1.X-cs2.Ghost1.X)
	porttest.EnsureSmall(t, got2.Ghost1.Y-cs2.Ghost1.Y, 1e-5)
	o.F32("L507", got2.Ghost1.Y-cs2.Ghost1.Y)
	porttest.EnsureSmall(t, got2.Ghost2.X-cs2.Ghost2.X, 1e-5)
	o.F32("L508", got2.Ghost2.X-cs2.Ghost2.X)
	porttest.EnsureSmall(t, got2.Ghost2.Y-cs2.Ghost2.Y, 1e-5)
	o.F32("L509", got2.Ghost2.Y-cs2.Ghost2.Y)
	porttest.Ensure(t, got2.ChainID == -1)
	o.Bool("L510", got2.ChainID == -1)
	o.I32("L510:value", int32(got2.ChainID))

	var parentChain2 box2d.Chain = orphanShape.ParentChain()
	porttest.Ensure(t, parentChain2.IsNull())
	o.Bool("L513", parentChain2.IsNull())

	var convBody box2d.Body = worldId.CreateBody(&bodyDef)
	var convCircle box2d.Circle = box2d.Circle{box2d.Vec2{0.0, 0.0}, 0.25}
	var convShape box2d.Shape = convBody.CreateCircleShape(&shapeDef, &convCircle)
	porttest.Ensure(t, convShape.Type() == box2d.CircleShape)
	o.Bool("L518", convShape.Type() == box2d.CircleShape)
	o.I32("L518:value", int32(convShape.Type()))

	convShape.SetChainSegment(&cs2)

	porttest.Ensure(t, convShape.Type() == box2d.ChainSegmentShape)
	o.Bool("L522", convShape.Type() == box2d.ChainSegmentShape)
	o.I32("L522:value", int32(convShape.Type()))
	var got3 box2d.ChainSegment = convShape.ChainSegment()
	porttest.EnsureSmall(t, got3.Ghost1.X-cs2.Ghost1.X, 1e-5)
	o.F32("L524", got3.Ghost1.X-cs2.Ghost1.X)
	porttest.EnsureSmall(t, got3.Ghost1.Y-cs2.Ghost1.Y, 1e-5)
	o.F32("L525", got3.Ghost1.Y-cs2.Ghost1.Y)
	porttest.EnsureSmall(t, got3.Ghost2.X-cs2.Ghost2.X, 1e-5)
	o.F32("L526", got3.Ghost2.X-cs2.Ghost2.X)
	porttest.EnsureSmall(t, got3.Ghost2.Y-cs2.Ghost2.Y, 1e-5)
	o.F32("L527", got3.Ghost2.Y-cs2.Ghost2.Y)
	porttest.Ensure(t, got3.ChainID == -1)
	o.Bool("L528", got3.ChainID == -1)
	o.I32("L528:value", int32(got3.ChainID))

	var parentChain3 box2d.Chain = convShape.ParentChain()
	porttest.Ensure(t, parentChain3.IsNull())
	o.Bool("L531", parentChain3.IsNull())

	orphanShape.Destroy(true)
	worldId.Destroy()

	return 0
}

func worldDeferredMassFlagSyncTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	shapeDef.UpdateBodyMass = false

	var circle box2d.Circle = box2d.Circle{box2d.Vec2{0.0, 0.0}, 0.5}
	bodyId.CreateCircleShape(&shapeDef, &circle)

	bodyId.UpdateMassFromShapes()

	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L556")

	worldId.Destroy()
	return 0
}

func worldEnableSleepFlagSyncTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.EnableSleep = false
	var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

	porttest.Ensure(t, bodyId.IsSleepEnabled() == false)
	o.Bool("L572", bodyId.IsSleepEnabled() == false)

	bodyId.EnableSleep(true)
	porttest.Ensure(t, bodyId.IsSleepEnabled() == true)
	o.Bool("L575", bodyId.IsSleepEnabled() == true)

	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L577")

	worldId.Destroy()
	return 0
}

func worldEnableContactRecyclingTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	// Default is enabled
	var bodyA box2d.Body = worldId.CreateBody(&bodyDef)
	porttest.Ensure(t, bodyA.IsContactRecyclingEnabled() == true)
	o.Bool("L593", bodyA.IsContactRecyclingEnabled() == true)

	bodyA.EnableContactRecycling(false)
	porttest.Ensure(t, bodyA.IsContactRecyclingEnabled() == false)
	o.Bool("L596", bodyA.IsContactRecyclingEnabled() == false)

	bodyA.EnableContactRecycling(true)
	porttest.Ensure(t, bodyA.IsContactRecyclingEnabled() == true)
	o.Bool("L599", bodyA.IsContactRecyclingEnabled() == true)

	// Per-def opt-out at creation
	bodyDef.EnableContactRecycling = false
	var bodyB box2d.Body = worldId.CreateBody(&bodyDef)
	porttest.Ensure(t, bodyB.IsContactRecyclingEnabled() == false)
	o.Bool("L604", bodyB.IsContactRecyclingEnabled() == false)

	// Stepping after toggling must not trip the flag-sync validator
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L607")

	worldId.Destroy()
	return 0
}

func worldSetBulletDriftTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.IsBullet = false
		var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

		porttest.Ensure(t, bodyId.IsBullet() == false)
		o.Bool("L624", bodyId.IsBullet() == false)

		bodyId.SetBullet(true)
		porttest.Ensure(t, bodyId.IsBullet() == true)
		o.Bool("L627", bodyId.IsBullet() == true)

		var locks box2d.MotionLocks = box2d.MotionLocks{}
		locks.LinearX = true
		bodyId.SetMotionLocks(locks)

		porttest.Ensure(t, bodyId.IsBullet() == true)
		o.Bool("L633", bodyId.IsBullet() == true)
	}

	{
		var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
		bodyDef.Type = box2d.DynamicBody
		bodyDef.IsBullet = true
		var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

		porttest.Ensure(t, bodyId.IsBullet() == true)
		o.Bool("L642", bodyId.IsBullet() == true)

		bodyId.SetBullet(false)
		porttest.Ensure(t, bodyId.IsBullet() == false)
		o.Bool("L645", bodyId.IsBullet() == false)

		var locks box2d.MotionLocks = box2d.MotionLocks{}
		locks.LinearX = true
		bodyId.SetMotionLocks(locks)

		porttest.Ensure(t, bodyId.IsBullet() == false)
		o.Bool("L651", bodyId.IsBullet() == false)
	}

	worldId.Destroy()
	return 0
}

// Counts marked nodes, the root included. Free nodes are never marked.

func worldCountEnlargedNodes(t *testing.T, o *worldOracle, tree *box2d.DynamicTree) int {
	return box2d.WorldCountEnlargedNodes(tree)
}

func worldEnlargedProxyDestroyed(t *testing.T, o *worldOracle, workerCount int) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	worldDef.WorkerCount = workerCount
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var tree *box2d.DynamicTree = box2d.WorldDynamicTree(worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var circle box2d.Circle = box2d.Circle{box2d.Vec2{}, 0.25}

	// Resting bodies keep the tree deep enough that the mover has an ancestor that outlives it
	for i := 0; i < 8; i++ {
		bodyDef.Position = box2d.Pos{oracle.Real(float32(2.0) * float32(i)), oracle.Real(float32(0.0))}
		var restingId box2d.Body = worldId.CreateBody(&bodyDef)
		restingId.CreateCircleShape(&shapeDef, &circle)
	}

	// This body outruns its AABB margin every step
	bodyDef.Position = box2d.Pos{oracle.Real(float32(7.0)), oracle.Real(float32(1.0))}
	bodyDef.LinearVelocity = box2d.Vec2{6.0, 0.0}
	var moverId box2d.Body = worldId.CreateBody(&bodyDef)
	moverId.CreateCircleShape(&shapeDef, &circle)

	var timeStep float32 = 1.0 / 60.0
	worldId.Step(timeStep, 4)
	o.step(worldId, "L704")

	porttest.Ensure(t, worldCountEnlargedNodes(t, o, tree) > 0)
	o.Bool("L706", worldCountEnlargedNodes(t, o, tree) > 0)
	o.I32("L706:value", int32(worldCountEnlargedNodes(t, o, tree)))

	moverId.Destroy()

	porttest.Ensure(t, worldCountEnlargedNodes(t, o, tree) == 0)
	o.Bool("L710", worldCountEnlargedNodes(t, o, tree) == 0)
	o.I32("L710:value", int32(worldCountEnlargedNodes(t, o, tree)))

	worldId.Step(timeStep, 4)
	o.step(worldId, "L712")

	porttest.Ensure(t, worldCountEnlargedNodes(t, o, tree) == 0)
	o.Bool("L714", worldCountEnlargedNodes(t, o, tree) == 0)
	o.I32("L714:value", int32(worldCountEnlargedNodes(t, o, tree)))

	worldId.Destroy()
	return 0
}

func worldEnlargedProxyDestroyedTest(t *testing.T, o *worldOracle) int {
	porttest.Ensure(t, worldEnlargedProxyDestroyed(t, o, 1) == 0)
	o.Bool("L722", true)
	porttest.Ensure(t, worldEnlargedProxyDestroyed(t, o, 4) == 0)
	o.Bool("L723", true)

	return 0
}

// Broad phase pair creation through every route that marks a proxy: creation, teleport, a static
// shape that asks for contacts, refiltering, collide connected, enable, a type change, a static
// rebuild that must keep its pending marks, and the filters that reject a pair.
// Bodies never move on their own here, gravity is off and sleep is disabled, so a contact can only
// come from the route under test. The count is every live contact, touching or not, which is what
// the broad phase decides.

func worldCreatePairBox(t *testing.T, o *worldOracle, worldId box2d.World, bodyType box2d.BodyType, x float32, y float32, shapeDef *box2d.ShapeDef) worldPairBox {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = bodyType
	bodyDef.Position = box2d.Pos{oracle.Real(x), oracle.Real(y)}

	var box worldPairBox
	box.BodyID = worldId.CreateBody(&bodyDef)

	var square box2d.Polygon = box2d.MakeSquare(0.5)
	box.ShapeID = box.BodyID.CreatePolygonShape(shapeDef, &square)
	return box
}

func worldCreatePairWorld(t *testing.T, o *worldOracle, workerCount int) box2d.World {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	worldDef.EnableSleep = false
	worldDef.WorkerCount = workerCount
	id := box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, id)
	return id
}

func worldPairContactCount(t *testing.T, o *worldOracle, worldId box2d.World) int {
	return worldId.Counters().ContactCount
}

func worldStepPairWorld(t *testing.T, o *worldOracle, worldId box2d.World) {
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L771")
}

// Every kind of proxy created over a resting body

func worldPairCreation(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	// Dynamic created on static
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.5, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L784", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L784:value", int32(worldPairContactCount(t, o, worldId)))

	// Dynamic created on a resting dynamic
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L789", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L789:value", int32(worldPairContactCount(t, o, worldId)))
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.5, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L792", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L792:value", int32(worldPairContactCount(t, o, worldId)))

	// Kinematic created on a resting dynamic, the mark is on the kinematic side
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 20.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	worldCreatePairBox(t, o, worldId, box2d.KinematicBody, 20.5, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 3)
	o.Bool("L799", worldPairContactCount(t, o, worldId) == 3)
	o.I32("L799:value", int32(worldPairContactCount(t, o, worldId)))

	// A static shape created on a resting dynamic pairs by default and not when told to stay quiet
	var quietDef box2d.ShapeDef = box2d.DefaultShapeDef()
	quietDef.InvokeContactCreation = false
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 30.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 30.5, 0.0, &quietDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 3)
	o.Bool("L808", worldPairContactCount(t, o, worldId) == 3)
	o.I32("L808:value", int32(worldPairContactCount(t, o, worldId)))

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 29.5, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 4)
	o.Bool("L812", worldPairContactCount(t, o, worldId) == 4)
	o.I32("L812:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// A teleport marks the moved proxy, kinematic and static ones included

func worldPairTeleport(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)
	var mover worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.0, &shapeDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 20.0, 0.0, &shapeDef)
	var wall worldPairBox = worldCreatePairBox(t, o, worldId, box2d.StaticBody, 30.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L829", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L829:value", int32(worldPairContactCount(t, o, worldId)))

	mover.BodyID.SetTransform(box2d.Pos{oracle.Real(float32(0.5)), oracle.Real(float32(0.0))}, box2d.Rot{C: 1})
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L833", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L833:value", int32(worldPairContactCount(t, o, worldId)))

	wall.BodyID.SetTransform(box2d.Pos{oracle.Real(float32(20.5)), oracle.Real(float32(0.0))}, box2d.Rot{C: 1})
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L837", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L837:value", int32(worldPairContactCount(t, o, worldId)))

	// Kinematic onto a resting dynamic, the mark is on the kinematic side
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 40.0, 0.0, &shapeDef)
	var pusher worldPairBox = worldCreatePairBox(t, o, worldId, box2d.KinematicBody, 50.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L843", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L843:value", int32(worldPairContactCount(t, o, worldId)))

	pusher.BodyID.SetTransform(box2d.Pos{oracle.Real(float32(40.5)), oracle.Real(float32(0.0))}, box2d.Rot{C: 1})
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 3)
	o.Bool("L847", worldPairContactCount(t, o, worldId) == 3)
	o.I32("L847:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// A filter change must find the pairs the old filter rejected and drop the ones it now rejects

func worldPairRefilter(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)

	var pickyDef box2d.ShapeDef = box2d.DefaultShapeDef()
	pickyDef.Filter.CategoryBits = 2
	pickyDef.Filter.MaskBits = 1

	var closedDef box2d.ShapeDef = box2d.DefaultShapeDef()
	closedDef.Filter.CategoryBits = 4

	var openFilter box2d.Filter = box2d.DefaultFilter()
	openFilter.CategoryBits = 1

	var ground worldPairBox = worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &closedDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.5, 0.0, &pickyDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L871", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L871:value", int32(worldPairContactCount(t, o, worldId)))

	ground.ShapeID.SetFilter(openFilter)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L875", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L875:value", int32(worldPairContactCount(t, o, worldId)))

	ground.ShapeID.SetFilter(closedDef.Filter)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L879", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L879:value", int32(worldPairContactCount(t, o, worldId)))

	// Same on a dynamic shape
	var other worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.0, &closedDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.5, 0.0, &pickyDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L885", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L885:value", int32(worldPairContactCount(t, o, worldId)))

	other.ShapeID.SetFilter(openFilter)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L889", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L889:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// Collide connected is off by default and turning it on must find the pair, static partner too.
// The weld frames coincide so the joint does not move anything.

func worldPairCollideConnected(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	var boxA worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.0, &shapeDef)
	var boxB worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.5, 0.0, &shapeDef)

	var jointDef box2d.WeldJointDef = box2d.DefaultWeldJointDef()
	jointDef.Base.BodyIDA = boxA.BodyID
	jointDef.Base.BodyIDB = boxB.BodyID
	jointDef.Base.LocalFrameB.P = box2d.Vec2{-0.5, 0.0}
	var jointId box2d.Joint = worldId.CreateWeldJoint(&jointDef).Joint
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L911", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L911:value", int32(worldPairContactCount(t, o, worldId)))

	jointId.SetCollideConnected(true)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L915", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L915:value", int32(worldPairContactCount(t, o, worldId)))

	jointId.SetCollideConnected(false)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L919", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L919:value", int32(worldPairContactCount(t, o, worldId)))

	// Static partner, the mark goes on the dynamic body
	var wall worldPairBox = worldCreatePairBox(t, o, worldId, box2d.StaticBody, 10.0, 0.0, &shapeDef)
	var boxC worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.5, 0.0, &shapeDef)
	jointDef.Base.BodyIDA = wall.BodyID
	jointDef.Base.BodyIDB = boxC.BodyID
	var wallJointId box2d.Joint = worldId.CreateWeldJoint(&jointDef).Joint
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L928", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L928:value", int32(worldPairContactCount(t, o, worldId)))

	wallJointId.SetCollideConnected(true)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L932", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L932:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

func worldRejectTaggedShape(t *testing.T, o *worldOracle, shapeIdA box2d.Shape, shapeIdB box2d.Shape) bool {
	return shapeIdA.UserData() == nil && shapeIdB.UserData() == nil
}

// The filters that reject a pair. The benchmarks never exercise these, so this is the only guard.

func worldPairFilters(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	worldId.SetCustomFilterCallback(func(a, b box2d.Shape) bool { return worldRejectTaggedShape(t, o, a, b) })

	// Category and mask must pass in both directions
	var defA box2d.ShapeDef = box2d.DefaultShapeDef()
	defA.Filter.CategoryBits = 1
	defA.Filter.MaskBits = 2
	var defB box2d.ShapeDef = box2d.DefaultShapeDef()
	defB.Filter.CategoryBits = 2
	defB.Filter.MaskBits = 1
	var defC box2d.ShapeDef = box2d.DefaultShapeDef()
	defC.Filter.CategoryBits = 2
	defC.Filter.MaskBits = 4
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.0, &defA)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.5, 0.0, &defB)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, -0.5, 0.0, &defC)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L964", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L964:value", int32(worldPairContactCount(t, o, worldId)))

	// Group index overrides the bits, negative rejects and positive accepts
	var negativeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	negativeDef.Filter.GroupIndex = -1
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.0, &negativeDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 10.5, 0.0, &negativeDef)
	var positiveDef box2d.ShapeDef = defC
	positiveDef.Filter.GroupIndex = 1
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 20.0, 0.0, &positiveDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 20.5, 0.0, &positiveDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L976", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L976:value", int32(worldPairContactCount(t, o, worldId)))

	// Custom filter rejects a tagged shape
	var customDef box2d.ShapeDef = box2d.DefaultShapeDef()
	customDef.EnableCustomFiltering = true
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 30.0, 0.0, &customDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 30.5, 0.0, &customDef)
	customDef.UserData = 1
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 40.0, 0.0, &customDef)
	customDef.UserData = nil
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 40.5, 0.0, &customDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 3)
	o.Bool("L988", worldPairContactCount(t, o, worldId) == 3)
	o.I32("L988:value", int32(worldPairContactCount(t, o, worldId)))

	// Sensors never make contacts
	var sensorDef box2d.ShapeDef = box2d.DefaultShapeDef()
	sensorDef.IsSensor = true
	var plainDef box2d.ShapeDef = box2d.DefaultShapeDef()
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 50.0, 0.0, &sensorDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 50.5, 0.0, &plainDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 3)
	o.Bool("L997", worldPairContactCount(t, o, worldId) == 3)
	o.I32("L997:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// A body that outran its margin on the step it went to sleep still gets its contact. This is the
// hazard behind section G: the marks live in the tree and do not care which set the body is in.

func worldPairSleepEscape(t *testing.T, o *worldOracle, workerCount int, sleepBeforePairs bool) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	worldDef.WorkerCount = workerCount
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	bodyDef.Position = box2d.Pos{oracle.Real(float32(-2.0)), oracle.Real(float32(0.0))}
	bodyDef.LinearVelocity = box2d.Vec2{58.0, 0.0}
	var moverId box2d.Body = worldId.CreateBody(&bodyDef)
	var square box2d.Polygon = box2d.MakeSquare(0.5)
	moverId.CreatePolygonShape(&shapeDef, &square)

	// One step carries the mover 0.97, far past its margin and into the fat box of the wall with a
	// few centimeters to spare, but not into the wall itself
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L1026", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L1026:value", int32(worldPairContactCount(t, o, worldId)))

	if sleepBeforePairs {
		moverId.SetAwake(false)
		porttest.Ensure(t, moverId.IsAwake() == false)
		o.Bool("L1031", moverId.IsAwake() == false)
	}

	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1035", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1035:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// Enable recreates the proxies with marks, on either side of a resting pair. The fat boxes
// overlap across the gap while the shapes do not, so nothing pushes and nothing moves.

func worldPairEnable(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	var ground worldPairBox = worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)
	var box worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 1.1, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1051", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1051:value", int32(worldPairContactCount(t, o, worldId)))

	box.BodyID.Disable()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L1055", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L1055:value", int32(worldPairContactCount(t, o, worldId)))

	box.BodyID.Enable()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1059", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1059:value", int32(worldPairContactCount(t, o, worldId)))

	// The static side, the resting box has no mark of its own
	ground.BodyID.Disable()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L1064", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L1064:value", int32(worldPairContactCount(t, o, worldId)))

	ground.BodyID.Enable()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1068", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1068:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// A type change recreates the proxies with marks and must find the resting neighbors again

func worldPairSetType(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)
	var middle worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 1.1, 0.0, &shapeDef)
	worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 2.2, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L1084", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L1084:value", int32(worldPairContactCount(t, o, worldId)))

	// Kinematic drops the static pair and keeps the dynamic one
	middle.BodyID.SetType(box2d.KinematicBody)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1089", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1089:value", int32(worldPairContactCount(t, o, worldId)))

	middle.BodyID.SetType(box2d.DynamicBody)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L1093", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L1093:value", int32(worldPairContactCount(t, o, worldId)))

	// Static drops the static pair too and the dynamic pair becomes static versus dynamic
	middle.BodyID.SetType(box2d.StaticBody)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1098", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1098:value", int32(worldPairContactCount(t, o, worldId)))

	middle.BodyID.SetType(box2d.DynamicBody)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L1102", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L1102:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// Bullet is a flag only, the pair must be untouched

func worldPairSetBullet(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.0, &shapeDef)
	var box worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 1.1, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1117", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1117:value", int32(worldPairContactCount(t, o, worldId)))

	box.BodyID.SetBullet(true)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1121", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1121:value", int32(worldPairContactCount(t, o, worldId)))

	box.BodyID.SetBullet(false)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1125", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1125:value", int32(worldPairContactCount(t, o, worldId)))

	worldId.Destroy()
	return 0
}

// A static rebuild must keep the pending marks. The resting box has none of its own, so a lost
// mark means the pair is never found.

func worldPairRebuildStatic(t *testing.T, o *worldOracle, workerCount int) int {
	var worldId box2d.World = worldCreatePairWorld(t, o, workerCount)
	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()

	var box worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.0, &shapeDef)
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 10.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L1141", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L1141:value", int32(worldPairContactCount(t, o, worldId)))

	// No marks to keep
	worldId.RebuildStaticTree()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 0)
	o.Bool("L1146", worldPairContactCount(t, o, worldId) == 0)
	o.I32("L1146:value", int32(worldPairContactCount(t, o, worldId)))

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 1.1, 0.0, &shapeDef)
	worldId.RebuildStaticTree()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1151", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1151:value", int32(worldPairContactCount(t, o, worldId)))

	// A quiet static shape has no mark to keep
	var quietDef box2d.ShapeDef = box2d.DefaultShapeDef()
	quietDef.InvokeContactCreation = false
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, -1.1, 0.0, &quietDef)
	worldId.RebuildStaticTree()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1159", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1159:value", int32(worldPairContactCount(t, o, worldId)))

	// Overlapping the resting box, the contact must touch on the first step after the rebuild
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.0, 0.5, &shapeDef)
	worldId.RebuildStaticTree()
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 2)
	o.Bool("L1165", worldPairContactCount(t, o, worldId) == 2)
	o.I32("L1165:value", int32(worldPairContactCount(t, o, worldId)))

	var contactData [4]box2d.ContactData
	var touchingCount int = worldContactData(box.BodyID, &contactData)
	porttest.Ensure(t, touchingCount == 1)
	o.Bool("L1169", touchingCount == 1)
	o.I32("L1169:value", int32(touchingCount))
	porttest.Ensure(t, contactData[0].Manifold.PointCount > 0)
	o.Bool("L1170", contactData[0].Manifold.PointCount > 0)
	o.I32("L1170:value", int32(contactData[0].Manifold.PointCount))

	worldId.Destroy()
	return 0
}

// A static shape created over a sleeping body still gets its contact, parked with the sleeper
// until it wakes. The world counter only sees awake contacts, so this counts on the body.

func worldPairSleepingStatic(t *testing.T, o *worldOracle, workerCount int) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	worldDef.WorkerCount = workerCount
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var sleeper worldPairBox = worldCreatePairBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	sleeper.BodyID.SetAwake(false)
	porttest.Ensure(t, sleeper.BodyID.IsAwake() == false)
	o.Bool("L1189", sleeper.BodyID.IsAwake() == false)

	worldCreatePairBox(t, o, worldId, box2d.StaticBody, 0.5, 0.0, &shapeDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, sleeper.BodyID.ContactCapacity() == 1)
	o.Bool("L1193", sleeper.BodyID.ContactCapacity() == 1)
	o.I32("L1193:value", int32(sleeper.BodyID.ContactCapacity()))
	porttest.Ensure(t, sleeper.BodyID.IsAwake() == false)
	o.Bool("L1194", sleeper.BodyID.IsAwake() == false)

	// The quiet kind never pairs with a sleeper
	var quietDef box2d.ShapeDef = box2d.DefaultShapeDef()
	quietDef.InvokeContactCreation = false
	worldCreatePairBox(t, o, worldId, box2d.StaticBody, -0.5, 0.0, &quietDef)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, sleeper.BodyID.ContactCapacity() == 1)
	o.Bool("L1201", sleeper.BodyID.ContactCapacity() == 1)
	o.I32("L1201:value", int32(sleeper.BodyID.ContactCapacity()))

	// Waking brings the parked contact into the step and it touches at once
	sleeper.BodyID.SetAwake(true)
	worldStepPairWorld(t, o, worldId)
	porttest.Ensure(t, worldPairContactCount(t, o, worldId) == 1)
	o.Bool("L1206", worldPairContactCount(t, o, worldId) == 1)
	o.I32("L1206:value", int32(worldPairContactCount(t, o, worldId)))

	var contactData [4]box2d.ContactData
	var touchingCount int = worldContactData(sleeper.BodyID, &contactData)
	porttest.Ensure(t, touchingCount == 1)
	o.Bool("L1210", touchingCount == 1)
	o.I32("L1210:value", int32(touchingCount))
	porttest.Ensure(t, contactData[0].Manifold.PointCount > 0)
	o.Bool("L1211", contactData[0].Manifold.PointCount > 0)
	o.I32("L1211:value", int32(contactData[0].Manifold.PointCount))

	worldId.Destroy()
	return 0
}

// The contact sim carries a persistent locator for each body sim instead of a copy written by the
// collide task every step. These scenarios drive every site that changes a body's set or local index
// while its contacts are live, then step under validation so validateSolverSets and the prepare
// asserts are the real check. Worker count parity on the state hash guards the multithreaded paths.

func worldStepLocatorWorld(t *testing.T, o *worldOracle, worldId box2d.World, stepCount int) {
	for i := 0; i < stepCount; i++ {
		worldId.Step(1.0/60.0, 4)
		o.step(worldId, "L1226")
	}
}

func worldStepUntilAsleep(t *testing.T, o *worldOracle, worldId box2d.World, bodyId box2d.Body, maxSteps int) bool {
	for i := 0; i < maxSteps; i++ {
		worldId.Step(1.0/60.0, 4)
		o.step(worldId, "L1234")
		if bodyId.IsAwake() == false {
			return true
		}
	}

	return false
}

func worldCreateLocatorWorld(t *testing.T, o *worldOracle, workerCount int, enableSleep bool) box2d.World {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{0.0, -10.0}
	worldDef.EnableSleep = enableSleep
	worldDef.WorkerCount = workerCount
	id := box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, id)
	return id
}

func worldCreateLocatorGround(t *testing.T, o *worldOracle, worldId box2d.World, x float32, halfWidth float32) box2d.Body {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Position = box2d.Pos{oracle.Real(x), oracle.Real(float32(-0.5))}
	var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var box box2d.Polygon = box2d.MakeBox(halfWidth, 0.5)
	bodyId.CreatePolygonShape(&shapeDef, &box)
	return bodyId
}

func worldCreateLocatorBox(t *testing.T, o *worldOracle, worldId box2d.World, bodyType box2d.BodyType, x float32, y float32) box2d.Body {
	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = bodyType
	bodyDef.Position = box2d.Pos{oracle.Real(x), oracle.Real(y)}
	var bodyId box2d.Body = worldId.CreateBody(&bodyDef)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var square box2d.Polygon = box2d.MakeSquare(0.5)
	bodyId.CreatePolygonShape(&shapeDef, &square)
	return bodyId
}

// A pile sleeps on static ground and a dropped body wakes it. Covers first touch, the graph add and
// the wake of a whole sleeping set.

func worldLocatorSleepWake(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, true)
	worldCreateLocatorGround(t, o, worldId, 0.0, 10.0)

	var pile [3]box2d.Body
	for i := 0; i < 3; i++ {
		pile[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.5+float32(i))
	}

	porttest.Ensure(t, worldStepUntilAsleep(t, o, worldId, pile[0], 600))
	o.Bool("L1291", true)
	porttest.Ensure(t, pile[2].IsAwake() == false)
	o.Bool("L1292", pile[2].IsAwake() == false)

	var dropper box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 6.0)
	porttest.Ensure(t, pile[0].IsAwake() == false)
	o.Bool("L1295", pile[0].IsAwake() == false)

	worldStepLocatorWorld(t, o, worldId, 60)
	porttest.Ensure(t, pile[0].IsAwake())
	o.Bool("L1298", pile[0].IsAwake())
	porttest.Ensure(t, dropper.IsAwake())
	o.Bool("L1299", dropper.IsAwake())

	porttest.Ensure(t, worldStepUntilAsleep(t, o, worldId, dropper, 900))
	o.Bool("L1301", true)
	porttest.Ensure(t, pile[0].IsAwake() == false)
	o.Bool("L1302", pile[0].IsAwake() == false)

	*hash = worldId.StateHash()
	o.U64("LocatorSleepWake:hash", *hash)
	worldId.Destroy()
	return 0
}

// Destroying a body in the middle of the awake set swaps the last awake sim into the hole while that
// body's contacts are live.

func worldLocatorDestroyAwakeBody(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, false)
	worldCreateLocatorGround(t, o, worldId, 0.0, 20.0)

	var stack [4]box2d.Body
	for i := 0; i < 4; i++ {
		stack[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.5+float32(i))
	}

	// A second stack far away owns the tail of the awake set
	var neighbors [3]box2d.Body
	for i := 0; i < 3; i++ {
		neighbors[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.5+float32(i))
	}

	worldStepLocatorWorld(t, o, worldId, 120)
	var restY float32 = float32(neighbors[2].Position().Y)

	stack[0].Destroy()
	worldStepLocatorWorld(t, o, worldId, 30)

	porttest.EnsureSmall(t, float32(neighbors[2].Position().Y)-restY, 0.01)
	o.F32("L1335", float32(neighbors[2].Position().Y)-restY)

	*hash = worldId.StateHash()
	o.U64("LocatorDestroyAwakeBody:hash", *hash)
	worldId.Destroy()
	return 0
}

// Destroying a static body swaps the last static sim into the hole. The moved body still has awake
// contacts, so its side of them has to be re-encoded or the collide task reads past the static set.

func worldLocatorDestroyStaticBody(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, false)

	var platforms [3]box2d.Body
	var boxes [3]box2d.Body
	for i := 0; i < 3; i++ {
		var x float32 = float32(10.0) * float32(i)
		platforms[i] = worldCreateLocatorGround(t, o, worldId, x, 2.0)
		boxes[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, x, 0.5)
	}

	worldStepLocatorWorld(t, o, worldId, 60)
	var restY float32 = float32(boxes[2].Position().Y)

	platforms[0].Destroy()
	worldStepLocatorWorld(t, o, worldId, 40)

	porttest.EnsureSmall(t, float32(boxes[2].Position().Y)-restY, 0.01)
	o.F32("L1363", float32(boxes[2].Position().Y)-restY)
	porttest.Ensure(t, float32(boxes[0].Position().Y) < -1.0)
	o.Bool("L1364", float32(boxes[0].Position().Y) < -1.0)
	o.F32("L1364:value", float32(boxes[0].Position().Y))

	*hash = worldId.StateHash()
	o.U64("LocatorDestroyStaticBody:hash", *hash)
	worldId.Destroy()
	return 0
}

// A kinematic that never sleeps keeps a non touching contact to a body that falls asleep. That
// contact stays in the awake set with its sleeping side on the cold path, then the kinematic closes
// the gap and wakes it.

func worldLocatorSleepingNeighbor(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, true)
	worldCreateLocatorGround(t, o, worldId, 0.0, 10.0)

	var sleeper box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.5)

	// The fat boxes overlap across a tenth of a meter, well past the speculative distance
	var mover box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.KinematicBody, 1.1, 0.5)
	mover.EnableSleep(false)

	porttest.Ensure(t, worldStepUntilAsleep(t, o, worldId, sleeper, 600))
	o.Bool("L1385", true)
	porttest.Ensure(t, mover.IsAwake())
	o.Bool("L1386", mover.IsAwake())

	var contactData [4]box2d.ContactData
	porttest.Ensure(t, mover.ContactCapacity() == 1)
	o.Bool("L1389", mover.ContactCapacity() == 1)
	o.I32("L1389:value", int32(mover.ContactCapacity()))
	porttest.Ensure(t, worldContactData(mover, &contactData) == 0)
	o.Bool("L1390", worldContactData(mover, &contactData) == 0)

	worldStepLocatorWorld(t, o, worldId, 30)
	porttest.Ensure(t, sleeper.IsAwake() == false)
	o.Bool("L1393", sleeper.IsAwake() == false)

	mover.SetLinearVelocity(box2d.Vec2{-2.0, 0.0})
	worldStepLocatorWorld(t, o, worldId, 10)
	porttest.Ensure(t, sleeper.IsAwake())
	o.Bool("L1397", sleeper.IsAwake())

	*hash = worldId.StateHash()
	o.U64("LocatorSleepingNeighbor:hash", *hash)
	worldId.Destroy()
	return 0
}

// Type change, disable and enable all route through the body transfer, which pulls a body sim out of
// the middle of a set while other bodies hold contacts.

func worldLocatorTypeChange(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, false)
	worldCreateLocatorGround(t, o, worldId, 0.0, 20.0)

	var subject box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.5)
	worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 1.5)

	var neighbors [3]box2d.Body
	for i := 0; i < 3; i++ {
		neighbors[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 10.0, 0.5+float32(i))
	}

	worldStepLocatorWorld(t, o, worldId, 120)
	var restY float32 = float32(neighbors[2].Position().Y)

	subject.SetType(box2d.StaticBody)
	worldStepLocatorWorld(t, o, worldId, 30)
	porttest.Ensure(t, subject.Type() == box2d.StaticBody)
	o.Bool("L1425", subject.Type() == box2d.StaticBody)
	o.I32("L1425:value", int32(subject.Type()))

	subject.SetType(box2d.DynamicBody)
	worldStepLocatorWorld(t, o, worldId, 30)
	porttest.Ensure(t, subject.Type() == box2d.DynamicBody)
	o.Bool("L1429", subject.Type() == box2d.DynamicBody)
	o.I32("L1429:value", int32(subject.Type()))

	subject.Disable()
	worldStepLocatorWorld(t, o, worldId, 10)
	porttest.Ensure(t, subject.IsEnabled() == false)
	o.Bool("L1433", subject.IsEnabled() == false)

	subject.Enable()
	worldStepLocatorWorld(t, o, worldId, 30)
	porttest.Ensure(t, subject.IsEnabled())
	o.Bool("L1437", subject.IsEnabled())

	porttest.EnsureSmall(t, float32(neighbors[2].Position().Y)-restY, 0.01)
	o.F32("L1439", float32(neighbors[2].Position().Y)-restY)

	*hash = worldId.StateHash()
	o.U64("LocatorTypeChange:hash", *hash)
	worldId.Destroy()
	return 0
}

// A joint between two sleeping bodies merges their sets instead of waking them, while an awake
// kinematic holds a non touching contact to one of them across the merge.

func worldLocatorSetMerge(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldId box2d.World = worldCreateLocatorWorld(t, o, workerCount, true)
	worldCreateLocatorGround(t, o, worldId, 0.0, 20.0)

	var left box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, -5.0, 0.5)
	var right box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 5.0, 0.5)

	var watcher box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.KinematicBody, -3.9, 0.5)
	watcher.EnableSleep(false)

	porttest.Ensure(t, worldStepUntilAsleep(t, o, worldId, left, 600))
	o.Bool("L1459", true)
	porttest.Ensure(t, right.IsAwake() == false)
	o.Bool("L1460", right.IsAwake() == false)
	porttest.Ensure(t, watcher.IsAwake())
	o.Bool("L1461", watcher.IsAwake())
	porttest.Ensure(t, watcher.ContactCapacity() == 1)
	o.Bool("L1462", watcher.ContactCapacity() == 1)
	o.I32("L1462:value", int32(watcher.ContactCapacity()))

	var jointDef box2d.DistanceJointDef = box2d.DefaultDistanceJointDef()
	jointDef.Base.BodyIDA = left
	jointDef.Base.BodyIDB = right
	jointDef.Length = 10.0
	var jointId box2d.Joint = worldId.CreateDistanceJoint(&jointDef).Joint

	porttest.Ensure(t, left.IsAwake() == false)
	o.Bool("L1470", left.IsAwake() == false)
	porttest.Ensure(t, right.IsAwake() == false)
	o.Bool("L1471", right.IsAwake() == false)

	worldStepLocatorWorld(t, o, worldId, 30)
	porttest.Ensure(t, left.IsAwake() == false)
	o.Bool("L1474", left.IsAwake() == false)

	left.SetAwake(true)
	worldStepLocatorWorld(t, o, worldId, 10)
	porttest.Ensure(t, right.IsAwake())
	o.Bool("L1478", right.IsAwake())

	jointId.Destroy()
	worldStepLocatorWorld(t, o, worldId, 10)

	*hash = worldId.StateHash()
	o.U64("LocatorSetMerge:hash", *hash)
	worldId.Destroy()
	return 0
}

// A body moving far enough in one step carries the fast flag, which the collide task now reads off
// the sim. A public setter mid flight makes sure a flag sync does not drop it.

func worldLocatorFastBody(t *testing.T, o *worldOracle, workerCount int, hash *uint64) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	worldDef.WorkerCount = workerCount
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	worldCreateLocatorBox(t, o, worldId, box2d.StaticBody, 10.0, 0.0)
	var bullet box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, 0.0, 0.0)
	bullet.SetLinearVelocity(box2d.Vec2{40.0, 0.0})

	// A step carries it two thirds of a meter, well past half its minimum extent
	worldStepLocatorWorld(t, o, worldId, 4)
	porttest.Ensure(t, float32(bullet.Position().X) > 2.0)
	o.Bool("L1503", float32(bullet.Position().X) > 2.0)
	o.F32("L1503:value", float32(bullet.Position().X))

	bullet.EnableSleep(false)
	worldStepLocatorWorld(t, o, worldId, 4)

	bullet.EnableSleep(true)
	worldStepLocatorWorld(t, o, worldId, 30)

	// It stopped at the wall instead of tunneling through
	porttest.Ensure(t, float32(bullet.Position().X) < 9.1)
	o.Bool("L1512", float32(bullet.Position().X) < 9.1)
	o.F32("L1512:value", float32(bullet.Position().X))
	porttest.Ensure(t, float32(bullet.Position().X) > 8.5)
	o.Bool("L1513", float32(bullet.Position().X) > 8.5)
	o.F32("L1513:value", float32(bullet.Position().X))

	*hash = worldId.StateHash()
	o.U64("LocatorFastBody:hash", *hash)
	worldId.Destroy()
	return 0
}

// A row of boxes at rest in zero gravity is unaffected by sleeping, so the run that sleeps and wakes
// on impact has to land where the run that never sleeps lands. Positions rather than the state hash
// because the hash folds in set and local index, which sleeping is entitled to change.

func worldLocatorSleepParity(t *testing.T, o *worldOracle, workerCount int) int {
	var finalX [2][3]float32

	for pass := 0; pass < 2; pass++ {
		var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
		worldDef.Gravity = box2d.Vec2{}
		worldDef.EnableSleep = pass == 0
		worldDef.WorkerCount = workerCount
		var worldId box2d.World = box2d.CreateWorld(&worldDef)
		box2d.BTWorldCleanup(t, worldId)

		var boxes [3]box2d.Body
		for i := 0; i < 3; i++ {
			boxes[i] = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, float32(i), 0.0)
		}

		worldStepLocatorWorld(t, o, worldId, 60)
		porttest.Ensure(t, boxes[0].IsAwake() == (pass == 1))
		o.Bool("L1542", boxes[0].IsAwake() == (pass == 1))

		var impactor box2d.Body = worldCreateLocatorBox(t, o, worldId, box2d.DynamicBody, -5.0, 0.0)
		impactor.SetLinearVelocity(box2d.Vec2{5.0, 0.0})

		worldStepLocatorWorld(t, o, worldId, 120)
		porttest.Ensure(t, boxes[2].IsAwake())
		o.Bool("L1548", boxes[2].IsAwake())

		for i := 0; i < 3; i++ {
			finalX[pass][i] = float32(boxes[i].Position().X)
		}

		worldId.Destroy()
	}

	for i := 0; i < 3; i++ {
		porttest.EnsureSmall(t, finalX[0][i]-finalX[1][i], 0.01)
		o.F32("L1560", finalX[0][i]-finalX[1][i])
	}

	return 0
}

func worldBodySimLocatorTest(t *testing.T, o *worldOracle) int {
	var workerCounts [2]int = [2]int{1, 4}
	var hashes [2]uint64

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorSleepWake(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1573", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1575", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorDestroyAwakeBody(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1579", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1581", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorDestroyStaticBody(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1585", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1587", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorSleepingNeighbor(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1591", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1593", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorTypeChange(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1597", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1599", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorSetMerge(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1603", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1605", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorFastBody(t, o, workerCounts[i], &hashes[i]) == 0)
		o.Bool("L1609", true)
	}
	porttest.Ensure(t, hashes[0] == hashes[1])
	o.Bool("L1611", hashes[0] == hashes[1])

	for i := 0; i < 2; i++ {
		porttest.Ensure(t, worldLocatorSleepParity(t, o, workerCounts[i]) == 0)
		o.Bool("L1615", true)
	}

	return 0
}

func worldBroadPhasePairsTest(t *testing.T, o *worldOracle) int {
	var workerCounts [2]int = [2]int{1, 4}
	for i := 0; i < 2; i++ {
		var workerCount int = workerCounts[i]
		porttest.Ensure(t, worldPairCreation(t, o, workerCount) == 0)
		o.Bool("L1627", true)
		porttest.Ensure(t, worldPairTeleport(t, o, workerCount) == 0)
		o.Bool("L1628", true)
		porttest.Ensure(t, worldPairRefilter(t, o, workerCount) == 0)
		o.Bool("L1629", true)
		porttest.Ensure(t, worldPairCollideConnected(t, o, workerCount) == 0)
		o.Bool("L1630", true)
		porttest.Ensure(t, worldPairFilters(t, o, workerCount) == 0)
		o.Bool("L1631", true)
		porttest.Ensure(t, worldPairEnable(t, o, workerCount) == 0)
		o.Bool("L1632", true)
		porttest.Ensure(t, worldPairSetType(t, o, workerCount) == 0)
		o.Bool("L1633", true)
		porttest.Ensure(t, worldPairSetBullet(t, o, workerCount) == 0)
		o.Bool("L1634", true)
		porttest.Ensure(t, worldPairRebuildStatic(t, o, workerCount) == 0)
		o.Bool("L1635", true)
		porttest.Ensure(t, worldPairSleepEscape(t, o, workerCount, false) == 0)
		o.Bool("L1636", true)
		porttest.Ensure(t, worldPairSleepEscape(t, o, workerCount, true) == 0)
		o.Bool("L1637", true)
		porttest.Ensure(t, worldPairSleepingStatic(t, o, workerCount) == 0)
		o.Bool("L1638", true)
	}

	return 0
}

// A joint with no dynamic body lives in the static set even when one body is an awake kinematic.
// Disabling and enabling that body has to respect this, whether the joint was created that way or
// got there through body type changes. Box3D issue #164.

func worldNonDynamicJointDisableTest(t *testing.T, o *worldOracle) int {
	var worldDef box2d.WorldDef = box2d.DefaultWorldDef()
	worldDef.Gravity = box2d.Vec2{}
	var worldId box2d.World = box2d.CreateWorld(&worldDef)
	box2d.BTWorldCleanup(t, worldId)

	var bodyDef box2d.BodyDef = box2d.DefaultBodyDef()
	bodyDef.Type = box2d.DynamicBody
	var bodyIdA box2d.Body = worldId.CreateBody(&bodyDef)
	bodyDef.Position = box2d.Pos{oracle.Real(float32(2.0)), oracle.Real(float32(0.0))}
	var bodyIdB box2d.Body = worldId.CreateBody(&bodyDef)

	var shapeDef box2d.ShapeDef = box2d.DefaultShapeDef()
	var box box2d.Polygon = box2d.MakeBox(0.5, 0.5)
	bodyIdA.CreatePolygonShape(&shapeDef, &box)
	bodyIdB.CreatePolygonShape(&shapeDef, &box)

	var jointDef box2d.WeldJointDef = box2d.DefaultWeldJointDef()
	jointDef.Base.BodyIDA = bodyIdA
	jointDef.Base.BodyIDB = bodyIdB
	var weldId box2d.Joint = worldId.CreateWeldJoint(&jointDef).Joint

	bodyIdA.SetType(box2d.StaticBody)
	bodyIdB.SetType(box2d.KinematicBody)
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L1671")

	bodyIdB.Disable()
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L1674")
	bodyIdB.Enable()
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L1676")
	porttest.Ensure(t, weldId.IsValid())
	o.Bool("L1677", weldId.IsValid())

	var world *box2d.WorldWorld = box2d.WorldGetWorldFromID(worldId)
	porttest.Ensure(t, box2d.WorldJointSetIndex(world, weldId) == box2d.WorldStaticSet)
	o.Bool("L1680", box2d.WorldJointSetIndex(world, weldId) == box2d.WorldStaticSet)
	o.I32("L1680:value", int32(box2d.WorldJointSetIndex(world, weldId)))

	bodyIdB.SetType(box2d.DynamicBody)
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L1683")
	porttest.Ensure(t, box2d.WorldJointSetIndex(world, weldId) == box2d.WorldAwakeSet)
	o.Bool("L1684", box2d.WorldJointSetIndex(world, weldId) == box2d.WorldAwakeSet)
	o.I32("L1684:value", int32(box2d.WorldJointSetIndex(world, weldId)))

	// Created directly between a static and a kinematic body
	bodyDef.Type = box2d.KinematicBody
	bodyDef.Position = box2d.Pos{oracle.Real(float32(4.0)), oracle.Real(float32(0.0))}
	var bodyIdC box2d.Body = worldId.CreateBody(&bodyDef)
	bodyIdC.CreatePolygonShape(&shapeDef, &box)

	var revoluteDef box2d.RevoluteJointDef = box2d.DefaultRevoluteJointDef()
	revoluteDef.Base.BodyIDA = bodyIdA
	revoluteDef.Base.BodyIDB = bodyIdC
	var revoluteId box2d.Joint = worldId.CreateRevoluteJoint(&revoluteDef).Joint

	bodyIdC.Disable()
	bodyIdC.Enable()
	worldId.Step(1.0/60.0, 4)
	o.step(worldId, "L1699")
	porttest.Ensure(t, box2d.WorldJointSetIndex(world, revoluteId) == box2d.WorldStaticSet)
	o.Bool("L1700", box2d.WorldJointSetIndex(world, revoluteId) == box2d.WorldStaticSet)
	o.I32("L1700:value", int32(box2d.WorldJointSetIndex(world, revoluteId)))

	worldId.Destroy()
	return 0
}

func TestWorldStepEncoding(t *testing.T) {
	defer porttest.Guard(t)
	cases := []struct {
		name   string
		hashes []uint64
		packed string
	}{
		{"empty", nil, ""},
		{"single", []uint64{0x0807060504030201}, "010000000102030405060708"},
		{"repeated", []uint64{9, 9, 9}, "030000000900000000000000"},
		{"runs", []uint64{9, 9, 10, 9}, "020000000900000000000000010000000a00000000000000010000000900000000000000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer porttest.Guard(t)
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			if got := worldPackSteps(tc.hashes); !bytes.Equal(got, packed) {
				t.Fatalf("encoding: got %x, want %x", got, packed)
			}
			decoded, err := worldUnpackSteps(packed)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(decoded, tc.hashes) {
				t.Fatalf("decoding: got %v, want %v", decoded, tc.hashes)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		packed string
	}{
		{"truncated", "01000000"},
		{"zero_run", "000000000900000000000000"},
		{"oversized_run", "010010000900000000000000"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer porttest.Guard(t)
			packed, err := hex.DecodeString(tc.packed)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := worldUnpackSteps(packed); err == nil {
				t.Fatal("malformed RLE accepted")
			}
		})
	}
	t.Run("recycle_64000_steps", func(t *testing.T) {
		defer porttest.Guard(t)
		packed, err := hex.DecodeString("00fa00000102030405060708")
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := worldUnpackSteps(packed)
		if err != nil {
			t.Fatal(err)
		}
		if len(decoded) != 64000 {
			t.Fatalf("decoded %d steps, want 64000", len(decoded))
		}
		for step, hash := range decoded {
			if hash != 0x0807060504030201 {
				t.Fatalf("step %d: wrong hash %x", step+1, hash)
			}
		}
		if !bytes.Equal(worldPackSteps(decoded), packed) {
			t.Fatal("recycle trace did not re-encode exactly")
		}
	})
}
