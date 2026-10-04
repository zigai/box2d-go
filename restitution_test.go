// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

const restitutionTimeStep float32 = 1.0 / 60.0
const restitutionSubStepCount = 4
const restitutionRestitutionIterations = 4
const restitutionImpactSpeed float32 = 5.0

func TestBTWorldCleanupSkipped(t *testing.T) {

	var ids [2]*World
	t.Run("skip", func(t *testing.T) {

		def := DefaultWorldDef()
		for i := range ids {
			ids[i] = NewWorld(&def)
			BTWorldCleanup(t, ids[i])
			if !(ids[i].IsValid()) {
				t.Fatal("condition false")
			}
		}
		t.SkipNow()
	})
	for _, id := range ids {
		if id.IsValid() {
			t.Fatal("skipped test left a live world")
		}
	}
}

func TestBTWorldCleanupExplicitDestroy(t *testing.T) {

	var replacement *World
	t.Run("destroy and recycle", func(child *testing.T) {

		def := DefaultWorldDef()
		id := NewWorld(&def)
		BTWorldCleanup(child, id)
		if !(id.IsValid()) {
			child.Fatal("condition false")
		}
		id.Close()
		if !(!id.IsValid()) {
			child.Fatal("condition false")
		}
		replacement = NewWorld(&def)
		BTWorldCleanup(t, replacement)
		if !(replacement.IsValid()) {
			child.Fatal("condition false")
		}
	})
	if !(replacement.IsValid()) {
		t.Fatal("condition false")
	}
	t.Run("null", func(t *testing.T) {

		BTWorldCleanup(t, nil)
	})
}

func restitutionMakeWorld(r *restitutionOracle, gravityY float32) *World {
	var worldDef WorldDef = DefaultWorldDef()
	worldDef.Gravity = Vec2{float32(0.0), gravityY}
	worldDef.EnableSleep = false
	id := NewWorld(&worldDef)
	BTWorldCleanup(r.t, id)
	return id
}

func restitutionMakeWorldIterations(r *restitutionOracle, gravityY float32, restitutionIterations int) *World {
	var worldDef WorldDef = DefaultWorldDef()
	worldDef.Gravity = Vec2{float32(0.0), gravityY}
	worldDef.EnableSleep = false
	worldDef.RestitutionIterations = restitutionIterations
	id := NewWorld(&worldDef)
	BTWorldCleanup(r.t, id)
	return id
}

func restitutionMakeGround(r *restitutionOracle, worldID *World, restitution float32) {
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(-float32(1.0))}
	var groundID Body = worldID.CreateBody(&bodyDef)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Material.Friction = float32(0.0)
	shapeDef.Material.Restitution = restitution
	var box Polygon = MakeBox(float32(40.0), float32(1.0))
	groundID.CreatePolygonShape(&shapeDef, &box)
}

func restitutionMakeBall(r *restitutionOracle, worldID *World, x float32, y float32, velocityY float32, restitution float32) Body {
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: oracle.Real(x), Y: oracle.Real(y)}
	bodyDef.LinearVelocity = Vec2{float32(0.0), velocityY}
	var bodyID Body = worldID.CreateBody(&bodyDef)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Material.Friction = float32(0.0)
	shapeDef.Material.Restitution = restitution
	var circle Circle = Circle{Vec2{float32(0.0), float32(0.0)}, float32(0.5)}
	bodyID.CreateCircleShape(&shapeDef, &circle)
	return bodyID
}

func restitutionMeasureHeadOn(r *restitutionOracle, restitution float32, densityB float32, momentumError *float32) float32 {
	var worldID *World = restitutionMakeWorld(r, float32(0.0))
	var circle Circle = Circle{Vec2{float32(0.0), float32(0.0)}, float32(0.5)}
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = float32(0.0)
	shapeDef.Material.Restitution = restitution
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: oracle.Real(-float32(1.0)), Y: oracle.Real(float32(0.0))}
	bodyDef.LinearVelocity = Vec2{float32(float32(0.5) * restitutionImpactSpeed), float32(0.0)}
	var idA Body = worldID.CreateBody(&bodyDef)
	idA.CreateCircleShape(&shapeDef, &circle)
	bodyDef.Position = Pos{X: oracle.Real(float32(1.0)), Y: oracle.Real(float32(0.0))}
	bodyDef.LinearVelocity = Vec2{float32(-float32(0.5) * restitutionImpactSpeed), float32(0.0)}
	var idB Body = worldID.CreateBody(&bodyDef)
	shapeDef.Density = densityB
	idB.CreateCircleShape(&shapeDef, &circle)
	var massA float32 = idA.Mass()
	var massB float32 = idB.Mass()
	var momentum0 float32 = (float32(massA*idA.LinearVelocity().X) + float32(massB*idB.LinearVelocity().X))
	for i := 0; i < 120; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
	}
	var vA float32 = idA.LinearVelocity().X
	var vB float32 = idB.LinearVelocity().X
	var momentum1 float32 = (float32(massA*vA) + float32(massB*vB))
	r.end(worldID)
	worldID.Close()
	*momentumError = ((momentum1 - momentum0) / (massA + massB))
	r.F32("MeasureHeadOn/momentumError", *momentumError)
	r.F32("MeasureHeadOn/return", ((vB - vA) / restitutionImpactSpeed))
	return ((vB - vA) / restitutionImpactSpeed)
}

func restitutionMeasureGroundBounce(r *restitutionOracle, restitution float32, gap float32) float32 {
	var worldID *World = restitutionMakeWorld(r, float32(0.0))
	restitutionMakeGround(r, worldID, float32(0.0))
	var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), (float32(0.5) + gap), -restitutionImpactSpeed, restitution)
	for i := 0; i < 60; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
	}
	var speed float32 = ballID.LinearVelocity().Y
	r.end(worldID)
	worldID.Close()
	r.F32("MeasureGroundBounce/return", (speed / restitutionImpactSpeed))
	return (speed / restitutionImpactSpeed)
}

func restitutionMeasureSupportedBounce(r *restitutionOracle, restitution float32, supportCount int, propagate bool) float32 {
	var worldDef WorldDef = DefaultWorldDef()
	worldDef.Gravity = Vec2{float32(0.0), float32(0.0)}
	worldDef.EnableSleep = false
	worldDef.EnableRestitutionPropagation = propagate
	worldDef.RestitutionIterations = restitutionRestitutionIterations
	var worldID *World = NewWorld(&worldDef)
	BTWorldCleanup(r.t, worldID)
	restitutionMakeGround(r, worldID, float32(0.0))
	for i := 0; i < supportCount; i++ {
		restitutionMakeBall(r, worldID, float32(0.0), (float32(0.5) + float32(float32(1.0)*float32(i))), float32(0.0), float32(0.0))
	}
	var y float32 = ((float32(0.5) + float32(float32(1.0)*float32(supportCount))) + float32(float32(float32(0.5)*restitutionImpactSpeed)*restitutionTimeStep))
	var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), y, -restitutionImpactSpeed, restitution)
	for i := 0; i < 60; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
	}
	var speed float32 = ballID.LinearVelocity().Y
	r.end(worldID)
	worldID.Close()
	r.F32("MeasureSupportedBounce/return", (speed / restitutionImpactSpeed))
	return (speed / restitutionImpactSpeed)
}

func restitutionMeasureFlatBounce(r *restitutionOracle, restitution float32, subStepCount int, spin *float32) float32 {
	var worldID *World = restitutionMakeWorldIterations(r, float32(0.0), restitutionRestitutionIterations)
	restitutionMakeGround(r, worldID, float32(0.0))
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real((float32(0.25) + float32(float32(float32(0.5)*restitutionImpactSpeed)*restitutionTimeStep)))}
	bodyDef.LinearVelocity = Vec2{float32(0.0), -restitutionImpactSpeed}
	var boxID Body = worldID.CreateBody(&bodyDef)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = float32(0.0)
	shapeDef.Material.Restitution = restitution
	var box Polygon = MakeBox(float32(1.0), float32(0.25))
	boxID.CreatePolygonShape(&shapeDef, &box)
	for i := 0; i < 60; i++ {
		worldID.Step(restitutionTimeStep, subStepCount)
		r.step(worldID)
	}
	var speed float32 = boxID.LinearVelocity().Y
	*spin = boxID.AngularVelocity()
	r.end(worldID)
	worldID.Close()
	r.F32("MeasureFlatBounce/spin", *spin)
	r.F32("MeasureFlatBounce/return", (speed / restitutionImpactSpeed))
	return (speed / restitutionImpactSpeed)
}

func restitutionMeasureDrop(r *restitutionOracle, dropHeight float32, apexes []float32, capacity int) int {
	var worldID *World = restitutionMakeWorld(r, -float32(10.0))
	var groundDef BodyDef = DefaultBodyDef()
	var groundID Body = worldID.CreateBody(&groundDef)
	var groundShape ShapeDef = DefaultShapeDef()
	var segment Segment = Segment{Vec2{-float32(60.0), float32(0.0)}, Vec2{float32(60.0), float32(0.0)}}
	groundID.CreateSegmentShape(&groundShape, &segment)
	var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), dropHeight, float32(0.0), float32(1.0))
	var apexCount int = 0
	var previousSpeed float32 = float32(0.0)
	for i := 0; (i < 4000) && (apexCount < capacity); i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var speed float32 = ballID.LinearVelocity().Y
		if (previousSpeed > float32(0.0)) && (speed <= float32(0.0)) {
			apexes[apexCount] = float32(ballID.Position().Y)
			r.F32("MeasureDrop/apex", apexes[apexCount])
			apexCount += 1
		}
		previousSpeed = speed
	}
	r.end(worldID)
	worldID.Close()
	r.I32("MeasureDrop/return", int32(apexCount))
	return apexCount
}

func restitutionHeadOnTest(r *restitutionOracle) int {
	var restitutions []float32 = []float32{float32(0.0), float32(0.25), float32(0.5), float32(0.75), float32(1.0)}
	var densities []float32 = []float32{float32(1.0), float32(10.0), float32(100.0)}
	var tolerance float32 = float32(0.02)
	var failed int = 0
	for i := 0; i < len(densities); i++ {
		var worstError float32 = float32(0.0)
		var worstMomentum float32 = float32(0.0)
		for j := 0; j < len(restitutions); j++ {
			var momentumError float32 = float32(0.0)
			var measured float32 = restitutionMeasureHeadOn(r, restitutions[j], densities[i], &momentumError)
			var error float32 = absFloat((measured - restitutions[j]))
			if error > worstError {
				worstError = error
			}
			if absFloat(momentumError) > absFloat(worstMomentum) {
				worstMomentum = momentumError
			}
			r.F32("check/277/value0", error)
			r.F32("check/277/value1", tolerance)
			r.Bool("check/277", (error > tolerance))
			if error > tolerance {
				failed = 1
			}
		}
		r.F32("check/286/value0", worstError)
		r.F32("check/286/value1", worstMomentum)
		r.Bool("check/286", (absFloat(worstMomentum) > float32(1.0e-4)))
		if absFloat(worstMomentum) > float32(1.0e-4) {
			failed = 1
		}
	}
	r.I32("HeadOnTest/return", int32(failed))
	return failed
}

func restitutionPhaseTest(r *restitutionOracle) int {
	var restitutions []float32 = []float32{float32(0.25), float32(0.5), float32(0.9)}
	var sampleCount int = 16
	var tolerance float32 = float32(0.05)
	var failed int = 0
	for j := 0; j < len(restitutions); j++ {
		var minimum float32 = float32(math.MaxFloat32)
		var maximum float32 = -float32(math.MaxFloat32)
		var sum float32 = float32(0.0)
		for i := 0; i < sampleCount; i++ {
			var gap float32 = float32(float32((float32(i)/float32(sampleCount))*restitutionImpactSpeed) * restitutionTimeStep)
			var measured float32 = restitutionMeasureGroundBounce(r, restitutions[j], gap)
			minimum = minFloat(minimum, measured)
			maximum = maxFloat(maximum, measured)
			sum += measured
		}
		var mean float32 = (sum / float32(sampleCount))
		var spread float32 = (maximum - minimum)
		r.F32("check/323/value0", minimum)
		r.F32("check/323/value1", maximum)
		r.F32("check/323/value2", mean)
		r.F32("check/323/value3", spread)
		r.F32("check/323/value4", tolerance)
		r.Bool("check/323", ((spread > tolerance) || (absFloat((mean - restitutions[j])) > tolerance)))
		if (spread > tolerance) || (absFloat((mean - restitutions[j])) > tolerance) {
			failed = 1
		}
	}
	r.I32("PhaseTest/return", int32(failed))
	return failed
}

func restitutionSupportedTest(r *restitutionOracle) int {
	var restitutions []float32 = []float32{float32(0.5), float32(0.9)}
	var failed int = 0
	for propagate := 0; propagate <= 1; propagate++ {
		var floor float32 = 0.5
		if propagate != 0 {
			floor = 0.75
		}
		for j := 0; j < len(restitutions); j++ {
			for n := 0; n <= 3; n++ {
				var measured float32 = restitutionMeasureSupportedBounce(r, restitutions[j], n, (propagate == 1))
				r.F32("check/357/value0", measured)
				r.F32("check/357/value1", restitutions[j])
				r.F32("check/357/value2", floor)
				r.Bool("check/357", ((measured > (restitutions[j] + float32(0.01))) || (measured < float32(floor*restitutions[j]))))
				if (measured > (restitutions[j] + float32(0.01))) || (measured < float32(floor*restitutions[j])) {
					failed = 1
				}
			}
		}
	}
	r.I32("SupportedTest/return", int32(failed))
	return failed
}

func restitutionTwoPointTest(r *restitutionOracle) int {
	var restitutions []float32 = []float32{float32(0.5), float32(0.9)}
	var subStepCounts []int = []int{4, 8}
	var speedTolerance float32 = float32(0.01)
	var spinTolerance float32 = float32(0.03)
	var failed int = 0
	for j := 0; j < len(restitutions); j++ {
		for k := 0; k < len(subStepCounts); k++ {
			var spin float32 = float32(0.0)
			var measured float32 = restitutionMeasureFlatBounce(r, restitutions[j], subStepCounts[k], &spin)
			if subStepCounts[k] != restitutionSubStepCount {
				continue
			}
			r.F32("check/396/value0", measured)
			r.F32("check/396/value1", restitutions[j])
			r.F32("check/396/value2", spin)
			r.F32("check/396/value3", speedTolerance)
			r.F32("check/396/value4", spinTolerance)
			r.Bool("check/396", ((absFloat((measured - restitutions[j])) > speedTolerance) || (absFloat(spin) > spinTolerance)))
			if (absFloat((measured - restitutions[j])) > speedTolerance) || (absFloat(spin) > spinTolerance) {
				failed = 1
			}
		}
	}
	r.I32("TwoPointTest/return", int32(failed))
	return failed
}

func restitutionSpinTest(r *restitutionOracle) int {
	var spins []float32 = []float32{float32(0.0), float32(0.5), float32(1.0), float32(2.0)}
	var failed int = 0
	for j := 0; j < len(spins); j++ {
		var worldID *World = restitutionMakeWorldIterations(r, float32(0.0), restitutionRestitutionIterations)
		restitutionMakeGround(r, worldID, float32(0.0))
		var bodyDef BodyDef = DefaultBodyDef()
		bodyDef.Type = DynamicBody
		bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real((float32(0.25) + float32(float32(float32(0.5)*restitutionImpactSpeed)*restitutionTimeStep)))}
		bodyDef.LinearVelocity = Vec2{float32(0.0), -restitutionImpactSpeed}
		bodyDef.AngularVelocity = spins[j]
		var boxID Body = worldID.CreateBody(&bodyDef)
		var shapeDef ShapeDef = DefaultShapeDef()
		shapeDef.Density = float32(1.0)
		shapeDef.Material.Friction = float32(0.0)
		shapeDef.Material.Restitution = float32(1.0)
		var box Polygon = MakeBox(float32(1.0), float32(0.25))
		boxID.CreatePolygonShape(&shapeDef, &box)
		for i := 0; i < 60; i++ {
			worldID.Step(restitutionTimeStep, restitutionSubStepCount)
			r.step(worldID)
		}
		var speed float32 = boxID.LinearVelocity().Y
		var spin float32 = boxID.AngularVelocity()
		r.end(worldID)
		worldID.Close()
		r.F32("SpinTest/speed", speed)
		r.F32("SpinTest/spin", spin)
		r.Bool("check/446", ((absFloat((speed - restitutionImpactSpeed)) > float32(0.05)) || (absFloat((spin + spins[j])) > float32(0.05))))
		if (absFloat((speed - restitutionImpactSpeed)) > float32(0.05)) || (absFloat((spin + spins[j])) > float32(0.05)) {
			failed = 1
		}
	}
	r.I32("SpinTest/return", int32(failed))
	return failed
}

func restitutionDropTest(r *restitutionOracle) int {
	var heights []float32 = []float32{float32(40.0), float32(20.0), float32(10.0)}
	var failed int = 0
	for j := 0; j < len(heights); j++ {
		var apexes [6]float32 = [6]float32{}
		var apexCount int = restitutionMeasureDrop(r, heights[j], apexes[:], len(apexes))

		r.I32("check/478/value0", int32(apexCount))
		r.Bool("check/478", (apexCount < len(apexes)))
		if apexCount < len(apexes) {
			failed = 1
			continue
		}
		var highest float32 = apexes[0]
		for i := 1; i < apexCount; i++ {
			highest = maxFloat(highest, apexes[i])
		}
		r.F32("check/491/value0", highest)
		r.F32("check/491/value1", apexes[0])
		r.F32("check/491/value2", heights[j])
		r.F32("check/491/value3", apexes[apexCount-1])
		r.Bool("check/491", (((highest > float32(float32(1.001)*apexes[0])) || (apexes[0] > float32(float32(1.001)*heights[j]))) || (apexes[(apexCount-1)] < float32(float32(0.8)*apexes[0]))))
		if ((highest > float32(float32(1.001)*apexes[0])) || (apexes[0] > float32(float32(1.001)*heights[j]))) || (apexes[(apexCount-1)] < float32(float32(0.8)*apexes[0])) {
			failed = 1
		}
	}
	r.I32("DropTest/return", int32(failed))
	return failed
}

func restitutionSingleBoxTest(r *restitutionOracle) int {
	var worldID *World = restitutionMakeWorldIterations(r, -float32(10.0), restitutionRestitutionIterations)
	var groundDef BodyDef = DefaultBodyDef()
	var groundID Body = worldID.CreateBody(&groundDef)
	var groundShape ShapeDef = DefaultShapeDef()
	var segment Segment = Segment{Vec2{-float32(20.0), float32(0.0)}, Vec2{float32(20.0), float32(0.0)}}
	groundID.CreateSegmentShape(&groundShape, &segment)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Restitution = float32(1.0)
	shapeDef.Material.Friction = float32(0.0)
	var box Polygon = MakeBox(float32(0.5), float32(0.5))
	var dropHeight float32 = float32(10.0)
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(dropHeight)}
	bodyDef.SafetyFactor = float32(0.01)
	var boxID Body = worldID.CreateBody(&bodyDef)
	boxID.CreatePolygonShape(&shapeDef, &box)
	var firstSpin float32 = float32(0.0)
	var apexes [4]float32 = [4]float32{}
	var spins [4]float32 = [4]float32{}
	var apexCount int = 0
	var previousSpeed float32 = float32(0.0)
	for i := 0; (i < 1200) && (apexCount < len(apexes)); i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var speed float32 = boxID.LinearVelocity().Y
		if ((apexCount == 0) && (previousSpeed <= float32(0.0))) && (speed > float32(0.0)) {
			firstSpin = absFloat(boxID.AngularVelocity())
		}
		if (previousSpeed > float32(0.0)) && (speed <= float32(0.0)) {
			apexes[apexCount] = float32(boxID.Position().Y)
			spins[apexCount] = absFloat(boxID.AngularVelocity())
			r.F32("SingleBoxTest/apex", apexes[apexCount])
			r.F32("SingleBoxTest/spin", spins[apexCount])
			apexCount += 1
		}
		previousSpeed = speed
	}
	r.end(worldID)
	worldID.Close()

	r.F32("SingleBoxTest/firstSpin", firstSpin)
	r.Bool("check/568", (apexCount == len(apexes)))
	if !(apexCount == len(apexes)) {
		r.t.Fatal("condition false")
	}
	var failed int = 0
	r.Bool("check/572", (firstSpin > float32(0.01)))
	if firstSpin > float32(0.01) {
		failed = 1
	}
	for i := 0; i < apexCount; i++ {
		r.F32("check/579/value0", apexes[i])
		r.F32("check/579/value1", dropHeight)
		r.Bool("check/579", ((apexes[i] < float32(float32(0.95)*dropHeight)) || (apexes[i] > float32(float32(1.001)*dropHeight))))
		if (apexes[i] < float32(float32(0.95)*dropHeight)) || (apexes[i] > float32(float32(1.001)*dropHeight)) {
			failed = 1
		}
	}
	r.I32("SingleBoxTest/return", int32(failed))
	return failed
}

func restitutionThresholdTest(r *restitutionOracle) int {
	var worldDef WorldDef = DefaultWorldDef()
	var threshold float32 = worldDef.RestitutionThreshold
	var scales []float32 = []float32{float32(0.5), float32(0.9), float32(1.5), float32(4.0)}
	var failed int = 0
	for j := 0; j < len(scales); j++ {
		var speed float32 = float32(scales[j] * threshold)
		var worldID *World = restitutionMakeWorld(r, float32(0.0))
		restitutionMakeGround(r, worldID, float32(0.0))
		var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), (float32(0.5) + float32(float32(float32(0.5)*speed)*restitutionTimeStep)), -speed, float32(1.0))
		for i := 0; i < 60; i++ {
			worldID.Step(restitutionTimeStep, restitutionSubStepCount)
			r.step(worldID)
		}
		var ratio float32 = (ballID.LinearVelocity().Y / speed)
		r.end(worldID)
		worldID.Close()
		r.F32("ThresholdTest/ratio", ratio)
		r.F32("check/615/value0", scales[j])
		r.F32("check/615/value1", threshold)
		r.F32("check/615/value2", ratio)
		r.Bool("check/615", ((scales[j] < float32(1.0)) && (ratio > float32(0.05))))
		if (scales[j] < float32(1.0)) && (ratio > float32(0.05)) {
			failed = 1
		}
		r.Bool("check/620", ((scales[j] > float32(1.0)) && (ratio < float32(0.9))))
		if (scales[j] > float32(1.0)) && (ratio < float32(0.9)) {
			failed = 1
		}
	}
	r.I32("ThresholdTest/return", int32(failed))
	return failed
}

func restitutionRestingTest(r *restitutionOracle) int {
	var worldID *World = restitutionMakeWorld(r, -float32(10.0))
	restitutionMakeGround(r, worldID, float32(0.9))
	var boxCount int = 10
	var boxIDs [10]Body
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = float32(0.6)
	shapeDef.Material.Restitution = float32(0.9)
	var box Polygon = MakeBox(float32(0.5), float32(0.5))
	for i := 0; i < boxCount; i++ {
		var bodyDef BodyDef = DefaultBodyDef()
		bodyDef.Type = DynamicBody
		bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real((float32(0.5) + float32(float32(1.0)*float32(i))))}
		boxIDs[i] = worldID.CreateBody(&bodyDef)
		boxIDs[i].CreatePolygonShape(&shapeDef, &box)
	}
	for i := 0; i < 120; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
	}
	var settled float32 = float32(boxIDs[(boxCount - 1)].Position().Y)
	var drift float32 = float32(0.0)
	var peakSpeed float32 = float32(0.0)
	for i := 0; i < 300; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var y float32 = float32(boxIDs[(boxCount - 1)].Position().Y)
		drift = maxFloat(drift, absFloat((y - settled)))
		for j := 0; j < boxCount; j++ {
			peakSpeed = maxFloat(peakSpeed, boxIDs[j].LinearVelocity().Length())
		}
	}
	r.end(worldID)
	worldID.Close()
	r.F32("RestingTest/drift", drift)
	r.F32("RestingTest/peakSpeed", peakSpeed)
	r.Bool("check/679", (drift < float32(0.01)))
	if !(drift < float32(0.01)) {
		r.t.Fatal("condition false")
	}
	r.Bool("check/680", (peakSpeed < float32(0.05)))
	if !(peakSpeed < float32(0.05)) {
		r.t.Fatal("condition false")
	}
	r.I32("RestingTest/return", int32(0))
	return 0
}

func restitutionMeasureFirstTouchNormalVelocity(r *restitutionOracle, enableHitEvents bool) float32 {
	var worldID *World = restitutionMakeWorld(r, float32(0.0))
	restitutionMakeGround(r, worldID, float32(0.0))
	var bodyDef BodyDef = DefaultBodyDef()
	bodyDef.Type = DynamicBody
	bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(float32(0.6))}
	var ballID Body = worldID.CreateBody(&bodyDef)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Material.Friction = float32(0.0)
	shapeDef.Material.Restitution = float32(0.0)
	shapeDef.EnableHitEvents = enableHitEvents
	var circle Circle = Circle{Vec2{float32(0.0), float32(0.0)}, float32(0.5)}
	ballID.CreateCircleShape(&shapeDef, &circle)
	var normalVelocity float32 = float32(math.MaxFloat32)
	for i := 0; i < 20; i++ {
		ballID.SetLinearVelocity(Vec2{float32(0.0), -restitutionImpactSpeed})
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var contactData ContactData
		contactCount := 0
		for data := range ballID.ContactData() {
			contactData = data
			contactCount = 1
			break
		}
		if contactCount == 1 && contactData.Manifold.PointCount > 0 {
			normalVelocity = contactData.Manifold.Points[0].NormalVelocity
			break
		}
	}
	r.end(worldID)
	worldID.Close()
	r.F32("MeasureFirstTouchNormalVelocity/return", normalVelocity)
	return normalVelocity
}

func restitutionNormalVelocityTest(r *restitutionOracle) int {
	var quiet float32 = restitutionMeasureFirstTouchNormalVelocity(r, false)
	var published float32 = restitutionMeasureFirstTouchNormalVelocity(r, true)
	var failed int = 0
	r.F32("check/734/value0", quiet)
	r.F32("check/734/value1", published)
	r.Bool("check/734", (quiet != float32(0.0)))
	if quiet != float32(0.0) {
		failed = 1
	}
	r.Bool("check/739", (published > float32(-float32(0.8)*restitutionImpactSpeed)))
	if published > float32(-float32(0.8)*restitutionImpactSpeed) {
		failed = 1
	}
	r.I32("NormalVelocityTest/return", int32(failed))
	return failed
}

func restitutionMeasureEnergy(r *restitutionOracle, bodyID Body, worldID *World) float32 {
	var massData MassData = bodyID.MassData()
	var v Vec2 = bodyID.LinearVelocity()
	var w float32 = bodyID.AngularVelocity()
	var gravity Vec2 = worldID.Gravity().Scale(bodyID.GravityScale())
	var h float32 = (restitutionTimeStep / float32(restitutionSubStepCount))
	var kinetic float32 = (float32(float32(float32(0.5)*massData.Mass)*v.Dot(v)) + float32(float32(float32(float32(0.5)*massData.RotationalInertia)*w)*w))
	var potential float32 = float32(-massData.Mass * gravity.Dot(bodyID.WorldCenter().Vec2()))
	var phase float32 = float32(float32(float32(float32(0.5)*h)*massData.Mass) * v.Dot(gravity))
	return ((kinetic + potential) + phase)
}

type restitutionOvershootResult struct {
	firstApex        float32
	firstEnergyRatio float32
	peakEnergyRatio  float32
	speedRatio       float32
	bounceCount      int
}

func restitutionMeasureOvershoot(r *restitutionOracle, continuous bool) restitutionOvershootResult {
	var worldID *World = restitutionMakeWorldIterations(r, -float32(10.0), restitutionRestitutionIterations)
	var floorDef BodyDef = DefaultBodyDef()
	floorDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(-float32(0.25))}
	var floorID Body = worldID.CreateBody(&floorDef)
	var floorShape ShapeDef = DefaultShapeDef()
	var floor Polygon = MakeBox(float32(0.375), float32(0.25))
	floorID.CreatePolygonShape(&floorShape, &floor)
	var dropHeight float32 = float32(10.0)
	var boxDef BodyDef = DefaultBodyDef()
	boxDef.Type = DynamicBody
	boxDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(dropHeight)}
	if continuous {
		boxDef.SafetyFactor = float32(0.1)
	}
	var boxID Body = worldID.CreateBody(&boxDef)
	var boxShape ShapeDef = DefaultShapeDef()
	boxShape.Material.Restitution = float32(1.0)
	var box Polygon = MakeBox(float32(0.5), float32(0.5))
	boxID.CreatePolygonShape(&boxShape, &box)
	var result restitutionOvershootResult = restitutionOvershootResult{}
	var startEnergy float32 = restitutionMeasureEnergy(r, boxID, worldID)
	var impactSpeed float32 = float32(0.0)
	var reboundSpeed float32 = float32(0.0)
	var bounceHeight float32 = float32(0.0)
	var previousSpeed float32 = float32(0.0)
	for i := 0; (i < 600) && (result.bounceCount < 2); i++ {
		var before float32 = boxID.LinearVelocity().Y
		var heightBefore float32 = float32(boxID.Position().Y)
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var speed float32 = boxID.LinearVelocity().Y
		if ((reboundSpeed == float32(0.0)) && (before < float32(0.0))) && (speed > float32(0.0)) {
			impactSpeed = -before
			reboundSpeed = speed
			bounceHeight = heightBefore
		}
		if reboundSpeed == float32(0.0) {
			continue
		}
		var ratio float32 = (restitutionMeasureEnergy(r, boxID, worldID) / startEnergy)
		if ratio > result.peakEnergyRatio {
			result.peakEnergyRatio = ratio
		}
		if (previousSpeed > float32(0.0)) && (speed <= float32(0.0)) {
			if result.bounceCount == 0 {
				result.firstApex = float32(boxID.Position().Y)
				result.firstEnergyRatio = ratio
			}
			result.bounceCount += 1
		}
		previousSpeed = speed
	}
	r.end(worldID)
	worldID.Close()
	result.speedRatio = (reboundSpeed / impactSpeed)
	r.F32("MeasureOvershoot/bounceHeight", bounceHeight)
	r.F32("MeasureOvershoot/firstApex", result.firstApex)
	r.F32("MeasureOvershoot/firstEnergyRatio", result.firstEnergyRatio)
	r.F32("MeasureOvershoot/peakEnergyRatio", result.peakEnergyRatio)
	r.F32("MeasureOvershoot/speedRatio", result.speedRatio)
	r.I32("MeasureOvershoot/bounceCount", int32(result.bounceCount))
	return result
}

func restitutionOvershootTest(r *restitutionOracle) int {
	var dropHeight float32 = float32(10.0)
	var discrete restitutionOvershootResult = restitutionMeasureOvershoot(r, false)
	var continuous restitutionOvershootResult = restitutionMeasureOvershoot(r, true)
	var speedTolerance float32 = float32(1.0e-5)
	var failed int = 0
	r.Bool("check/897", ((discrete.speedRatio > (float32(1.0) + speedTolerance)) || (continuous.speedRatio > (float32(1.0) + speedTolerance))))
	if (discrete.speedRatio > (float32(1.0) + speedTolerance)) || (continuous.speedRatio > (float32(1.0) + speedTolerance)) {
		failed = 1
	}
	r.Bool("check/902", ((continuous.firstApex > dropHeight) || (continuous.peakEnergyRatio > float32(1.001))))
	if (continuous.firstApex > dropHeight) || (continuous.peakEnergyRatio > float32(1.001)) {
		failed = 1
	}
	r.Bool("check/907", ((discrete.firstApex > (dropHeight + float32(0.15))) || (discrete.firstEnergyRatio > float32(1.02))))
	if (discrete.firstApex > (dropHeight + float32(0.15))) || (discrete.firstEnergyRatio > float32(1.02)) {
		failed = 1
	}
	r.Bool("check/912", ((discrete.bounceCount < 2) || (continuous.bounceCount < 2)))
	if (discrete.bounceCount < 2) || (continuous.bounceCount < 2) {
		failed = 1
	}
	r.I32("OvershootTest/return", int32(failed))
	return failed
}

type restitutionImpulseResult struct {
	worstError      float32
	approachSpeed   float32
	firstSeparation float32
	bounceImpulse   float32
	contactSteps    int
	toiSteps        int
	toiImpulseSteps int
}

func restitutionMeasureDropImpulse(r *restitutionOracle, restitution float32, dropHeight float32) restitutionImpulseResult {
	var worldID *World = restitutionMakeWorld(r, -float32(10.0))
	restitutionMakeGround(r, worldID, float32(0.0))
	var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), (float32(0.5) + dropHeight), float32(0.0), restitution)
	var world *World = worldID
	var mass float32 = ballID.Mass()
	var gravityY float32 = worldID.Gravity().Y
	var result restitutionImpulseResult = restitutionImpulseResult{}
	var touched bool = false
	var bouncing bool = false
	for i := 0; i < 240; i++ {
		var speedBefore float32 = ballID.LinearVelocity().Y
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var speedAfter float32 = ballID.LinearVelocity().Y
		var measured float32 = float32(0.0)
		var separation float32 = float32(0.0)
		var contactData [4]ContactData
		contactCount := 0
		for data := range ballID.ContactData() {
			contactData[contactCount] = data
			contactCount++
			if contactCount == len(contactData) {
				break
			}
		}
		for c := 0; c < contactCount; c++ {
			var manifold *Manifold = &contactData[c].Manifold
			for p := 0; p < manifold.PointCount; p++ {
				measured += manifold.Points[p].TotalNormalImpulse
				separation = manifold.Points[p].Separation
			}
		}
		var ball *body = getBodyFullId(world, ballID)
		var ballSim *bodySim = getBodySim(world, ball)
		if (ballSim.flags & hadTimeOfImpact) != 0 {
			result.toiSteps += 1
			if measured != float32(0.0) {
				result.toiImpulseSteps += 1
			}
			continue
		}
		var expected float32 = (float32(mass*(speedAfter-speedBefore)) - float32(float32(mass*gravityY)*restitutionTimeStep))
		result.worstError = maxFloat(result.worstError, absFloat((measured - expected)))
		if contactCount > 0 {
			if touched == false {
				touched = true
				bouncing = true
				result.approachSpeed = -speedBefore
				result.firstSeparation = separation
			}
			if bouncing {
				result.bounceImpulse += measured
				result.contactSteps += 1
			}
		} else {
			bouncing = false
		}
	}
	r.end(worldID)
	worldID.Close()
	r.F32("MeasureDropImpulse/worstError", result.worstError)
	r.F32("MeasureDropImpulse/approachSpeed", result.approachSpeed)
	r.F32("MeasureDropImpulse/firstSeparation", result.firstSeparation)
	r.F32("MeasureDropImpulse/bounceImpulse", result.bounceImpulse)
	r.I32("MeasureDropImpulse/contactSteps", int32(result.contactSteps))
	r.I32("MeasureDropImpulse/toiSteps", int32(result.toiSteps))
	r.I32("MeasureDropImpulse/toiImpulseSteps", int32(result.toiImpulseSteps))
	return result
}

func restitutionImpulseTest(r *restitutionOracle) int {
	var restitutions []float32 = []float32{float32(0.0), float32(0.25), float32(0.5), float32(0.75), float32(1.0)}
	var heights []float32 = []float32{float32(1.0), float32(5.0), float32(20.0), float32(45.0)}
	var mass float32
	var fastSpeed float32
	{
		var worldID *World = restitutionMakeWorld(r, -float32(10.0))
		var ballID Body = restitutionMakeBall(r, worldID, float32(0.0), float32(0.5), float32(0.0), float32(0.0))
		mass = ballID.Mass()
		fastSpeed = (float32(ballID.SafetyFactor()*ballID.MinExtent()) / restitutionTimeStep)
		r.end(worldID)
		worldID.Close()
	}
	var weightImpulse float32 = float32(float32(mass*float32(10.0)) * restitutionTimeStep)
	var failed int = 0
	for j := 0; j < len(heights); j++ {
		var impactSpeed float32 = float32(math.Sqrt(float64(float32(float32(20.0) * heights[j]))))
		var expectToi bool = (impactSpeed > fastSpeed)
		var tolerance float32 = float32(float32(float32(1e-5)*mass) * (float32(10.0) + impactSpeed))
		for k := 0; k < len(restitutions); k++ {
			var result restitutionImpulseResult = restitutionMeasureDropImpulse(r, restitutions[k], heights[j])
			var reversal float32 = float32(float32((float32(1.0)+restitutions[k])*mass) * result.approachSpeed)
			var lower float32 = (reversal - weightImpulse)
			var upper float32 = (reversal + float32(weightImpulse*float32(result.contactSteps)))
			r.F32("check/1062/value0", tolerance)
			r.Bool("check/1062", (result.worstError > tolerance))
			if result.worstError > tolerance {
				failed = 1
			}
			r.Bool("check/1067", ((result.contactSteps == 0) || (result.toiImpulseSteps > 0)))
			if (result.contactSteps == 0) || (result.toiImpulseSteps > 0) {
				failed = 1
			}
			var slack float32 = float32(float32(result.contactSteps) * tolerance)
			r.F32("check/1073/value0", lower)
			r.F32("check/1073/value1", upper)
			r.F32("check/1073/value2", slack)
			r.Bool("check/1073", ((result.bounceImpulse < (lower - slack)) || (result.bounceImpulse > (upper + slack))))
			if (result.bounceImpulse < (lower - slack)) || (result.bounceImpulse > (upper + slack)) {
				failed = 1
			}
			r.F32("check/1078/value0", fastSpeed)
			r.F32("check/1078/value1", impactSpeed)
			r.Bool("check/1078/value2", expectToi)
			r.Bool("check/1078", ((result.toiSteps > 0) != expectToi))
			if (result.toiSteps > 0) != expectToi {
				failed = 1
			}
		}
	}
	r.I32("ImpulseTest/return", int32(failed))
	return failed
}

type restitutionEnergyScene struct {
	name         string
	safetyFactor float32
	friction     float32
	angle        float32
	spin         float32
	height       float32
	velocityY    float32
	halfWidth    float32
	halfHeight   float32
	bullet       bool
	stack        bool
	stepCount    int
}

func restitutionMeasurePeakEnergy(r *restitutionOracle, scene *restitutionEnergyScene) float32 {
	var worldID *World = restitutionMakeWorld(r, -float32(10.0))
	var groundDef BodyDef = DefaultBodyDef()
	var groundID Body = worldID.CreateBody(&groundDef)
	var groundShape ShapeDef = DefaultShapeDef()
	groundShape.Material.Friction = scene.friction
	var segment Segment = Segment{Vec2{-float32(40.0), float32(0.0)}, Vec2{float32(40.0), float32(0.0)}}
	groundID.CreateSegmentShape(&groundShape, &segment)
	var shapeDef ShapeDef = DefaultShapeDef()
	shapeDef.Density = float32(1.0)
	shapeDef.Material.Friction = scene.friction
	shapeDef.Material.Restitution = float32(1.0)
	var box Polygon = MakeBox(scene.halfWidth, scene.halfHeight)
	var bodyIDs [4]Body
	var bodyCount int = 0
	if scene.stack {
		var bodyDef BodyDef = DefaultBodyDef()
		bodyDef.Type = DynamicBody
		bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(scene.halfHeight)}
		bodyDef.SafetyFactor = scene.safetyFactor
		bodyIDs[bodyCount] = worldID.CreateBody(&bodyDef)
		bodyIDs[bodyCount].CreatePolygonShape(&shapeDef, &box)
		bodyCount += 1
	}
	{
		var bodyDef BodyDef = DefaultBodyDef()
		bodyDef.Type = DynamicBody
		bodyDef.Position = Pos{X: oracle.Real(float32(0.0)), Y: oracle.Real(scene.height)}
		bodyDef.Rotation = MakeRot(scene.angle)
		bodyDef.AngularVelocity = scene.spin
		bodyDef.LinearVelocity = Vec2{float32(0.0), scene.velocityY}
		bodyDef.SafetyFactor = scene.safetyFactor
		bodyDef.IsBullet = scene.bullet
		bodyIDs[bodyCount] = worldID.CreateBody(&bodyDef)
		bodyIDs[bodyCount].CreatePolygonShape(&shapeDef, &box)
		bodyCount += 1
	}
	var start float32 = float32(0.0)
	for k := 0; k < bodyCount; k++ {
		start += restitutionMeasureEnergy(r, bodyIDs[k], worldID)
	}
	var peak float32 = start
	var peakStep int = -1
	for i := 0; i < scene.stepCount; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
		var energy float32 = float32(0.0)
		for k := 0; k < bodyCount; k++ {
			energy += restitutionMeasureEnergy(r, bodyIDs[k], worldID)
		}
		if energy > peak {
			peak = energy
			peakStep = i
		}
	}
	r.end(worldID)
	worldID.Close()
	var ratio float32 = (peak / start)
	r.I32("MeasurePeakEnergy/peakStep", int32(peakStep))
	r.F32("MeasurePeakEnergy/return", ratio)
	return ratio
}

func restitutionEnergyTest(r *restitutionOracle) int {
	scenes := []restitutionEnergyScene{
		{"spinning corner", float32(0.01), float32(0.6), -float32(1.43), float32(7.9), float32(0.7), -float32(1.9), float32(0.5), float32(0.5), false, false, 300},
		{"spinning corner fast", float32(0.01), float32(0.6), -float32(0.8), float32(12.0), float32(3.0), -float32(8.0), float32(0.5), float32(0.5), false, false, 300},
		{"tumble", float32(0.01), float32(0.6), -float32(0.3), float32(3.0), float32(8.0), float32(0.0), float32(0.5), float32(0.5), false, false, 600},
		{"tumble discrete", float32(0.5), float32(0.6), -float32(0.3), float32(3.0), float32(8.0), float32(0.0), float32(0.5), float32(0.5), false, false, 600},
		{"flat with spin", float32(0.01), float32(0.6), float32(0.0), float32(2.0), float32(2.0), -float32(5.0), float32(0.5), float32(0.5), false, false, 300},
		{"flat frictionless", float32(0.01), float32(0.0), float32(0.0), float32(0.0), float32(10.0), float32(0.0), float32(0.5), float32(0.5), false, false, 1200},
		{"corner", float32(0.01), float32(0.6), -float32(0.2), float32(0.0), float32(5.0), float32(0.0), float32(0.5), float32(0.5), false, false, 600},
		{"two box", float32(0.02), float32(0.6), float32(0.0), float32(0.0), float32(4.0), float32(0.0), float32(0.5), float32(0.5), true, true, 1500},
		{"two box discrete", float32(0.5), float32(0.6), float32(0.0), float32(0.0), float32(4.0), float32(0.0), float32(0.5), float32(0.5), false, true, 1500},
		{"tall flat frictionless", float32(0.01), float32(0.0), float32(0.0), float32(0.0), float32(10.0), float32(0.0), float32(0.5), float32(1.0), false, false, 1200},
		{"tall flat friction", float32(0.01), float32(0.6), float32(0.0), float32(0.0), float32(10.0), float32(0.0), float32(0.5), float32(1.0), false, false, 1200},
		{"narrow flat frictionless", float32(0.01), float32(0.0), float32(0.0), float32(0.0), float32(10.0), float32(0.0), float32(0.25), float32(1.0), false, false, 1200},
		{"narrow flat friction", float32(0.01), float32(0.6), float32(0.0), float32(0.0), float32(10.0), float32(0.0), float32(0.25), float32(1.0), false, false, 1200},
	}
	var tolerance float32 = float32(1.0e-3)
	var discreteTolerance float32 = float32(2.0e-2)
	var failed int = 0
	for i := 0; i < len(scenes); i++ {
		var ratio float32 = restitutionMeasurePeakEnergy(r, &scenes[i])
		var limit float32 = discreteTolerance
		if scenes[i].safetyFactor < 0.1 {
			limit = tolerance
		}
		r.F32("check/1227/value0", ratio)
		r.F32("check/1227/value1", limit)
		r.Bool("check/1227", (ratio > (float32(1.0) + limit)))
		if ratio > (float32(1.0) + limit) {
			failed = 1
		}
	}
	r.I32("EnergyTest/return", int32(failed))
	return failed
}

func restitutionRunWorkerScene(r *restitutionOracle, workerCount int) uint64 {
	var worldDef WorldDef = DefaultWorldDef()
	worldDef.Gravity = Vec2{float32(0.0), -float32(10.0)}
	worldDef.EnableSleep = false
	worldDef.WorkerCount = workerCount
	var worldID *World = NewWorld(&worldDef)
	BTWorldCleanup(r.t, worldID)
	restitutionMakeGround(r, worldID, float32(0.6))
	for i := 0; i < 20; i++ {
		restitutionMakeBall(r, worldID, (-float32(10.0) + float32(float32(1.05)*float32(i))), (float32(3.0) + float32(float32(0.13)*float32(i))), float32(0.0), float32(0.6))
	}
	for i := 0; i < 200; i++ {
		worldID.Step(restitutionTimeStep, restitutionSubStepCount)
		r.step(worldID)
	}
	var hash uint64 = worldID.StateHash()
	r.end(worldID)
	worldID.Close()
	r.U64("RunWorkerScene/return", hash)
	return hash
}

func restitutionWorkerParityTest(r *restitutionOracle) int {
	var hash0 uint64 = restitutionRunWorkerScene(r, 0)
	var hash1 uint64 = restitutionRunWorkerScene(r, 1)
	var hash4 uint64 = restitutionRunWorkerScene(r, 4)
	r.Bool("check/1272", (hash0 == hash1))
	if !(hash0 == hash1) {
		r.t.Fatal("condition false")
	}
	r.Bool("check/1273", (hash0 == hash4))
	if !(hash0 == hash4) {
		r.t.Fatal("condition false")
	}
	r.I32("WorkerParityTest/return", int32(0))
	return 0
}

type restitutionOracle struct {
	*oracle.Reader
	t                  *testing.T
	steps              []byte
	transformSteps     []byte
	fullStates         []byte
	previousTransforms []restitutionTransform
	snapshots          []restitutionSnapshot
	nextSnapshot       int
	richExpected       [][][36]uint64
	richRows           [][36]uint64
	richScenario       int
}

type restitutionTransform struct {
	bodyID int
	values [4]uint64
}

type restitutionSnapshot struct {
	key  string
	full []restitutionTransform
}

type restitutionRecord struct {
	Key   string `json:"k"`
	Type  string `json:"t"`
	Value string `json:"v"`
}

func restitutionRecords(t *testing.T, name string) []restitutionRecord {
	t.Helper()
	config := "float"
	if oracle.DoublePrecision {
		config = "double"
	}
	data, err := os.ReadFile("testdata/oracle/" + config + "/restitution/" + name + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var records []restitutionRecord
	for {
		var record restitutionRecord
		err := decoder.Decode(&record)
		if err == io.EOF {
			return records
		}
		if err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
}

func restitutionOpen(t *testing.T, name string) *restitutionOracle {
	t.Helper()
	r := &restitutionOracle{Reader: oracle.Open(t, "restitution/"+name), t: t}
	for _, record := range restitutionRecords(t, name) {
		if record.Key == "dynamics/rle-u32-36u64-le" {
			data, err := hex.DecodeString(record.Value)
			if err != nil {
				t.Fatal(err)
			}
			r.richExpected = append(r.richExpected, BTDynamicsUnpack(t, data))
		}
		if record.Key == "transforms/full" {
			data, err := hex.DecodeString(record.Value)
			if err != nil {
				t.Fatal(err)
			}
			var previous []restitutionTransform
			for step := 60; len(data) > 0; step += 60 {
				full := restitutionDecodeSnapshot(t, &data, previous)
				r.snapshots = append(r.snapshots, restitutionSnapshot{key: fmt.Sprintf("transforms/step/%d", step), full: full})
				previous = full
			}
		}
	}
	return r
}

func restitutionTransformWidth(component int) int {
	if component < 2 && oracle.DoublePrecision {
		return 8
	}
	return 4
}

func restitutionPackSnapshot(full, previous []restitutionTransform) []byte {
	var data []byte
	if previous == nil {
		data = binary.LittleEndian.AppendUint32(data, uint32(len(full)))
	}
	for i, body := range full {
		var mask byte
		for component, value := range body.values {
			if i >= len(previous) || previous[i].bodyID != body.bodyID || previous[i].values[component] != value {
				mask |= 1 << component
			}
		}
		if previous == nil {
			data = binary.LittleEndian.AppendUint32(data, uint32(body.bodyID))
		}
		data = append(data, mask)
		for component, value := range body.values {
			if mask&(1<<component) == 0 {
				continue
			}
			if restitutionTransformWidth(component) == 8 {
				data = binary.LittleEndian.AppendUint64(data, value)
			} else {
				data = binary.LittleEndian.AppendUint32(data, uint32(value))
			}
		}
	}
	return data
}

func restitutionDecodeSnapshot(t *testing.T, data *[]byte, previous []restitutionTransform) []restitutionTransform {
	t.Helper()
	take := func(width int) []byte {
		if len(*data) < width {
			t.Fatal("truncated restitution transform snapshot")
		}
		value := (*data)[:width]
		*data = (*data)[width:]
		return value
	}
	count := len(previous)
	if previous == nil {
		count = int(binary.LittleEndian.Uint32(take(4)))
	}
	if !(count > 0 && count <= 64) {
		t.Fatal("condition false")
	}
	full := make([]restitutionTransform, count)
	for i := range full {
		id := 0
		if previous == nil {
			id = int(binary.LittleEndian.Uint32(take(4)))
		} else {
			id = previous[i].bodyID
		}
		mask := take(1)[0]
		if !(mask&0xf0 == 0) {
			t.Fatal("condition false")
		}
		if i < len(previous) && previous[i].bodyID == id {
			full[i] = previous[i]
		} else {
			if !(mask == 0x0f) {
				t.Fatal("condition false")
			}
		}
		full[i].bodyID = id
		if i > 0 {
			if !(full[i-1].bodyID < id) {
				t.Fatal("condition false")
			}
		}
		for component := range full[i].values {
			if mask&(1<<component) == 0 {
				continue
			}
			if restitutionTransformWidth(component) == 8 {
				full[i].values[component] = binary.LittleEndian.Uint64(take(8))
			} else {
				full[i].values[component] = uint64(binary.LittleEndian.Uint32(take(4)))
			}
		}
	}
	return full
}

func (r *restitutionOracle) snapshot(step int, full []restitutionTransform) {
	r.t.Helper()
	key := fmt.Sprintf("transforms/step/%d", step)
	if r.nextSnapshot >= len(r.snapshots) || r.snapshots[r.nextSnapshot].key != key {
		r.t.Fatalf("missing restitution transform snapshot at step %d", step)
	}
	want := r.snapshots[r.nextSnapshot].full
	r.nextSnapshot++
	if len(want) != len(full) {
		r.t.Fatalf("restitution step %d: body count C=%d Go=%d", step, len(want), len(full))
	}
	components := [4]string{"position.x", "position.y", "rotation.c", "rotation.s"}
	for i, body := range full {
		if want[i].bodyID != body.bodyID {
			r.t.Fatalf("restitution step %d: body order %d C=%d Go=%d", step, i, want[i].bodyID, body.bodyID)
		}
		for component, value := range body.values {
			if want[i].values[component] != value {
				r.t.Fatalf("restitution step %d, body %d, %s: C=0x%x Go=0x%x", step, body.bodyID,
					components[component], want[i].values[component], value)
			}
		}
	}
	r.fullStates = append(r.fullStates, restitutionPackSnapshot(full, r.previousTransforms)...)
	r.previousTransforms = full
}

func (r *restitutionOracle) step(worldID *World) {
	row := BTObserveDynamics(r.Reader, worldID, fmt.Sprintf("observation/%d/", len(r.steps)/8+1), len(r.steps) == 0 || (len(r.steps)/8+1)%60 == 0)
	if r.richScenario >= len(r.richExpected) {
		r.t.Fatal("unexpected dynamics scenario")
	}
	BTDynamicsCompare(r.t, fmt.Sprintf("restitution/%s/scenario/%d", r.t.Name(), r.richScenario+1), len(r.steps)/8+1, row, r.richExpected[r.richScenario])
	r.richRows = append(r.richRows, row)
	world := worldID
	hash := uint64(14695981039346656037)
	transformHash := uint64(14695981039346656037)
	step := len(r.steps)/8 + 1
	sample := step%60 == 0
	var full []restitutionTransform
	mix := func(value uint64, size int) {
		for i := 0; i < size; i++ {
			hash ^= uint64(byte(value >> (8 * i)))
			hash *= 1099511628211
		}
	}
	f32 := func(value float32) { mix(uint64(math.Float32bits(value)), 4) }
	for i := range world.bodies {
		if world.bodies[i].id == nullIndex {
			continue
		}
		body := makeBodyId(world, i)
		p := body.Position()
		if oracle.DoublePrecision {
			mix(math.Float64bits(float64(p.X)), 8)
			mix(math.Float64bits(float64(p.Y)), 8)
		} else {
			f32(float32(p.X))
			f32(float32(p.Y))
		}
		q := body.Rotation()
		f32(q.C)
		f32(q.S)
		values := [4]uint64{uint64(math.Float32bits(float32(p.X))), uint64(math.Float32bits(float32(p.Y))),
			uint64(math.Float32bits(q.C)), uint64(math.Float32bits(q.S))}
		if oracle.DoublePrecision {
			values[0] = math.Float64bits(float64(p.X))
			values[1] = math.Float64bits(float64(p.Y))
		}
		for component, value := range values {
			for b := 0; b < restitutionTransformWidth(component); b++ {
				transformHash ^= uint64(byte(value >> (8 * b)))
				transformHash *= 1099511628211
			}
		}
		if sample {
			full = append(full, restitutionTransform{bodyID: i, values: values})
		}
		v := body.LinearVelocity()
		f32(v.X)
		f32(v.Y)
		f32(body.AngularVelocity())
		var contacts [64]ContactData
		count := 0
		for data := range body.ContactData() {
			contacts[count] = data
			count++
			if count == len(contacts) {
				break
			}
		}
		mix(uint64(count), 4)
		for c := 0; c < count; c++ {
			m := &contacts[c].Manifold
			mix(uint64(m.PointCount), 4)
			for p := 0; p < m.PointCount; p++ {
				f32(m.Points[p].Separation)
				f32(m.Points[p].NormalVelocity)
				f32(m.Points[p].TotalNormalImpulse)
			}
		}
	}
	if sample {
		r.snapshot(step, full)
	}
	r.steps = binary.LittleEndian.AppendUint64(r.steps, hash)
	r.transformSteps = binary.LittleEndian.AppendUint64(r.transformSteps, transformHash)
}

func (r *restitutionOracle) end(worldID *World) {
	if r.richScenario >= len(r.richExpected) || len(r.richRows) != len(r.richExpected[r.richScenario]) {
		r.t.Fatal("dynamics scenario step count differs")
	}
	BTDynamicsEnd(r.Reader, r.richRows)
	r.richRows = nil
	r.richScenario++
	r.Bytes("transforms/steps", r.transformSteps)
	r.Bytes("transforms/full", r.fullStates)
	r.Bytes("steps", r.steps)
	r.steps = r.steps[:0]
	r.transformSteps = r.transformSteps[:0]
	r.fullStates = r.fullStates[:0]
	r.previousTransforms = nil
}

func TestRestitution(t *testing.T) {

	tests := []struct {
		name string
		run  func(*restitutionOracle) int
	}{
		{"HeadOnTest", restitutionHeadOnTest},
		{"PhaseTest", restitutionPhaseTest},
		{"SupportedTest", restitutionSupportedTest},
		{"TwoPointTest", restitutionTwoPointTest},
		{"SpinTest", restitutionSpinTest},
		{"DropTest", restitutionDropTest},
		{"SingleBoxTest", restitutionSingleBoxTest},
		{"ThresholdTest", restitutionThresholdTest},
		{"RestingTest", restitutionRestingTest},
		{"NormalVelocityTest", restitutionNormalVelocityTest},
		{"OvershootTest", restitutionOvershootTest},
		{"EnergyTest", restitutionEnergyTest},
		{"ImpulseTest", restitutionImpulseTest},
		{"WorkerParityTest", restitutionWorkerParityTest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			r := restitutionOpen(t, test.name)
			if !(test.run(r) == 0) {
				t.Fatal("condition false")
			}
		})
	}
}

func TestRestitutionTraceFixtures(t *testing.T) {

	names := []string{"HeadOnTest", "PhaseTest", "SupportedTest", "TwoPointTest", "SpinTest", "DropTest", "SingleBoxTest", "ThresholdTest", "RestingTest", "NormalVelocityTest", "OvershootTest", "EnergyTest", "ImpulseTest", "WorkerParityTest"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {

			var trace []byte
			fullSeen := false
			scenarios := 0
			for _, record := range restitutionRecords(t, name) {
				switch record.Key {
				case "transforms/steps":
					if !(!fullSeen && trace == nil && record.Type == "bytes") {
						t.Fatal("condition false")
					}
					var err error
					trace, err = hex.DecodeString(record.Value)
					if err != nil {
						t.Fatal(err)
					}
					if !(len(trace)%8 == 0) {
						t.Fatal("condition false")
					}
				case "transforms/full":
					if !(trace != nil && !fullSeen && record.Type == "bytes") {
						t.Fatal("condition false")
					}
					encoded, err := hex.DecodeString(record.Value)
					if err != nil {
						t.Fatal(err)
					}
					data := encoded
					var previous []restitutionTransform
					var repacked []byte
					frames := 0
					for len(data) > 0 {
						full := restitutionDecodeSnapshot(t, &data, previous)
						frames++
						step := frames * 60
						if step > len(trace)/8 {
							t.Fatalf("snapshot step %d exceeds trace", step)
						}
						hash := oracle.HashInit
						for _, body := range full {
							for component, value := range body.values {
								if restitutionTransformWidth(component) == 8 {
									hash = oracle.HashU64(hash, value)
								} else {
									hash = oracle.HashU32(hash, uint32(value))
								}
							}
						}
						want := binary.LittleEndian.Uint64(trace[(step-1)*8 : step*8])
						if hash != want {
							t.Fatalf("snapshot at step %d hashes to 0x%x; transform trace has 0x%x", step, hash, want)
						}
						repacked = append(repacked, restitutionPackSnapshot(full, previous)...)
						previous = full
					}
					if !(frames == (len(trace)/8)/60) {
						t.Fatal("condition false")
					}
					if !(bytes.Equal(encoded, repacked)) {
						t.Fatal("condition false")
					}
					fullSeen = true
				case "steps":
					if !(fullSeen && record.Type == "bytes") {
						t.Fatal("condition false")
					}
					if !(len(record.Value) == len(trace)*2) {
						t.Fatal("condition false")
					}
					scenarios++
					trace = nil
					fullSeen = false
				}
			}
			if !(scenarios > 0 && trace == nil && !fullSeen) {
				t.Fatal("condition false")
			}
		})
	}
}
