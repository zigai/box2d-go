// SPDX-License-Identifier: MIT

package box2d_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func recordingObserveShapeID(r *oracle.Reader, key string, id box2d.Shape, world *box2d.World) {
	r.I32(key+"/index", int32(id.ID()>>32))
	r.U32(key+"/generation", uint32(uint16(id.ID())))
	r.Bool(key+"/owner", id == (box2d.Shape{}) || id.World() == world)
}

func recordingObserveQueries(t *testing.T, r *oracle.Reader, p *box2d.Replay, prefix string) {
	t.Helper()

	count := p.FrameQueryCount()
	r.I32(prefix+"/queries", int32(count))

	for ordinal := range count + 2 {
		i := ordinal - 1
		key := fmt.Sprintf("%s/q%d", prefix, i)
		q := p.FrameQuery(i)
		r.I32(key+"/type", int32(q.Type))
		r.U64(key+"/category", q.Filter.CategoryBits)
		r.U64(key+"/mask", q.Filter.MaskBits)
		r.Vec2(key+"/lower", q.AABB.LowerBound.X, q.AABB.LowerBound.Y)
		r.Vec2(key+"/upper", q.AABB.UpperBound.X, q.AABB.UpperBound.Y)
		r.Pos(key+"/origin", oracle.Real(q.Origin.X), oracle.Real(q.Origin.Y))
		r.Vec2(key+"/translation", q.Translation.X, q.Translation.Y)
		recordingObserveShapeID(r, key+"/shape", q.Shape, p.World())
		r.I32(key+"/hits", int32(q.HitCount))

		recordingObserveQueryHits(r, p, key, i, count, q)

		if i >= 0 && i < count {
			b, fraction, out := box2d.RecordingQueryResults(p, i)
			r.Bool(key+"/bool", b)
			r.F32(key+"/castFraction", fraction)
			r.Bool(key+"/castHit", out.Hit)
			r.Pos(key+"/castPoint", oracle.Real(out.Point.X), oracle.Real(out.Point.Y))
			r.Vec2(key+"/castNormal", out.Normal.X, out.Normal.Y)
			r.F32(key+"/castOutputFraction", out.Fraction)
			r.I32(key+"/castIterations", int32(out.Iterations))
		}
	}

	draw := box2d.DefaultDebugDraw()
	drawFuncs := &box2d.FuncDrawer{}
	draw.Drawer = drawFuncs
	hash, calls := uint64(oracle.HashInit), 0
	pos := func(p box2d.Pos) { hash = oracle.HashPos(hash, oracle.Real(p.X), oracle.Real(p.Y)) }
	vec := func(v box2d.Vec2) { hash = oracle.HashF32(hash, v.X); hash = oracle.HashF32(hash, v.Y) }
	drawFuncs.Line = func(a, b box2d.Pos, c box2d.HexColor) {
		calls++
		hash = oracle.HashU32(hash, 1)

		pos(a)
		pos(b)

		hash = oracle.HashU32(hash, uint32(c))
	}
	drawFuncs.Point = func(p box2d.Pos, size float32, c box2d.HexColor) {
		calls++
		hash = oracle.HashU32(hash, 2)

		pos(p)

		hash = oracle.HashF32(hash, size)
		hash = oracle.HashU32(hash, uint32(c))
	}
	drawFuncs.Polygon = func(xf box2d.WorldTransform, vertices []box2d.Vec2, c box2d.HexColor) {
		calls++
		hash = oracle.HashU32(hash, 3)
		hash = oracle.HashTransform(hash, oracle.Real(xf.P.X), oracle.Real(xf.P.Y), xf.Q.C, xf.Q.S)

		hash = oracle.HashU32(hash, uint32(len(vertices)))
		for _, v := range vertices {
			vec(v)
		}

		hash = oracle.HashU32(hash, uint32(c))
	}
	drawFuncs.SolidCapsule = func(a, b box2d.Pos, radius float32, c box2d.HexColor) {
		calls++
		hash = oracle.HashU32(hash, 4)

		pos(a)
		pos(b)

		hash = oracle.HashF32(hash, radius)
		hash = oracle.HashU32(hash, uint32(c))
	}

	for ordinal := range count + 2 {
		selected := ordinal - 1
		hash, calls = oracle.HashInit, 0

		p.DrawFrameQueries(&draw, selected)
		r.I32(fmt.Sprintf("%s/draw%d/count", prefix, selected), int32(calls))
		r.U64(fmt.Sprintf("%s/draw%d/hash", prefix, selected), hash)
	}
}

func recordingObserveQueryHits(r *oracle.Reader, p *box2d.Replay, key string, i, count int, q box2d.ReplayQueryInfo) {
	for ordinal := range q.HitCount + 2 {
		j := ordinal - 1
		hk := fmt.Sprintf("%s/h%d", key, j)
		h := p.FrameQueryHit(i, j)
		recordingObserveShapeID(r, hk+"/shape", h.Shape, p.World())

		validHit := i >= 0 && i < count && j >= 0 && j < q.HitCount

		identityOnly := validHit && (q.Type == box2d.ReplayQueryOverlapAABB || q.Type == box2d.ReplayQueryOverlapShape || q.Type == box2d.ReplayQueryCollideMover)
		if !identityOnly {
			r.Pos(hk+"/point", oracle.Real(h.Point.X), oracle.Real(h.Point.Y))
			r.Vec2(hk+"/normal", h.Normal.X, h.Normal.Y)
			r.F32(hk+"/fraction", h.Fraction)
		}

		if validHit && q.Type == box2d.ReplayQueryCollideMover {
			plane := box2d.RecordingQueryPlane(p, i, j)
			r.Vec2(hk+"/planeNormal", plane.Plane.Normal.X, plane.Plane.Normal.Y)
			r.F32(hk+"/planeOffset", plane.Plane.Offset)
			r.Vec2(hk+"/planePoint", plane.Point.X, plane.Point.Y)
			r.Bool(hk+"/planeHit", plane.Hit)
		}
	}
}

func recordingObserve(t *testing.T, r *oracle.Reader, p *box2d.Replay, prefix string) {
	t.Helper()

	info := p.Info()
	r.I32(prefix+"/frame", int32(p.Frame()))
	r.Bool(prefix+"/end", p.IsAtEnd())
	r.Bool(prefix+"/diverged", p.HasDiverged())
	r.I32(prefix+"/divergeFrame", int32(p.DivergeFrame()))
	r.U64(prefix+"/state", p.World().StateHash())
	r.I32(prefix+"/info/frames", int32(info.FrameCount))
	r.I32(prefix+"/info/workers", int32(info.WorkerCount))
	r.F32(prefix+"/info/dt", info.TimeStep)
	r.I32(prefix+"/info/substeps", int32(info.SubStepCount))
	r.F32(prefix+"/info/scale", info.LengthScale)
	r.Vec2(prefix+"/info/lower", info.Bounds.LowerBound.X, info.Bounds.LowerBound.Y)
	r.Vec2(prefix+"/info/upper", info.Bounds.UpperBound.X, info.Bounds.UpperBound.Y)

	count := p.BodyCount()
	r.I32(prefix+"/bodies", int32(count))

	for ordinal := range count + 3 {
		i := ordinal - 1
		key := fmt.Sprintf("%s/body%d", prefix, i)
		id := p.Body(i)
		r.I32(key+"/index", int32(id.ID()>>32))
		r.U32(key+"/generation", uint32(uint16(id.ID())))
		r.Bool(key+"/valid", id.IsValid())
		r.Bool(key+"/owner", id == (box2d.Body{}) || id.World() == p.World())

		if id.IsValid() {
			r.Bytes(key+"/name", []byte(id.Name()))
		}
	}

	recordingObserveQueries(t, r, p, prefix)
}

func recordingLifecycleSource(t *testing.T, steps int) *box2d.Recording {
	t.Helper()

	wd := box2d.DefaultWorldDef()
	wd.Gravity = box2d.Vec2{}
	wd.WorkerCount = 1
	w := box2d.NewWorld(&wd)
	box2d.CleanupWorld(t, w)

	bd := box2d.DefaultBodyDef()
	bd.Type = box2d.DynamicBody
	bd.Name = "seed"
	bd.Position = box2d.Pos{X: 1, Y: 2}
	seed := w.CreateBody(&bd)
	sd := box2d.DefaultShapeDef()
	circle := box2d.Circle{Radius: 0.5}
	seed.CreateCircleShape(&sd, &circle)

	rec := box2d.NewRecording(0)
	w.StartRecording(rec)

	for frame := range steps {
		if frame == 1 {
			seed.Destroy()

			bd.Name = "reused"
			bd.Position.X = 3
			seed = w.CreateBody(&bd)
			seed.CreateCircleShape(&sd, &circle)
		}

		if frame == 2 {
			bd.Name = "future"
			bd.Position.X = 5
			body := w.CreateBody(&bd)
			body.CreateCircleShape(&sd, &circle)
		}

		w.Step(1.0/30.0, 2)

		filter := box2d.DefaultQueryFilter()
		filter.CategoryBits = 3
		w.OverlapAABB(box2d.Pos{X: 2, Y: -1}, box2d.AABB{LowerBound: box2d.Vec2{X: -3, Y: -3}, UpperBound: box2d.Vec2{X: 6, Y: 5}}, filter, func(box2d.Shape) bool { return true })
		filter.MaskBits = 0
		w.OverlapAABB(box2d.Pos{X: 2, Y: -1}, box2d.AABB{LowerBound: box2d.Vec2{X: -3, Y: -3}, UpperBound: box2d.Vec2{X: 6, Y: 5}}, filter, func(box2d.Shape) bool { return true })
	}

	w.StopRecording()
	seed.SetName("not recorded")
	w.Close()

	return rec
}

func TestRecordingLifecycleParity(t *testing.T) {
	r := oracle.Open(t, "recording/LifecycleParity")
	oldScale := box2d.LengthUnitsPerMeter()

	box2d.SetLengthUnitsPerMeter(2)
	t.Cleanup(func() { box2d.SetLengthUnitsPerMeter(oldScale) })

	for _, steps := range []int{0, 1, 16} {
		rec := recordingLifecycleSource(t, steps)
		for _, workers := range []int{0, 4} {
			p, destroy := recordingParityPlayer(t, rec.Data(), workers)
			prefix := fmt.Sprintf("n%d/w%d", steps, workers)
			recordingObserve(t, r, p, prefix+"/open")

			for action := range steps + 3 {
				r.Bool(fmt.Sprintf("%s/step%d/return", prefix, action), p.StepFrame())
				recordingObserve(t, r, p, fmt.Sprintf("%s/step%d", prefix, action))
			}

			world := p.World()
			p.Restart()
			r.Bool(prefix+"/restart/identity", world == p.World())
			recordingObserve(t, r, p, prefix+"/restart")
			p.SetKeyframePolicy(1<<26, 2)
			r.U64(prefix+"/policy/budget", uint64(p.KeyframeBudget()))
			r.I32(prefix+"/policy/min", int32(p.KeyframeMinInterval()))
			r.I32(prefix+"/policy/interval", int32(p.KeyframeInterval()))
			r.Bool(prefix+"/policy/empty", p.KeyframeBytes() == 0)
			p.SetKeyframePolicy(0, -1)
			r.U64(prefix+"/policy/retainedBudget", uint64(p.KeyframeBudget()))
			r.I32(prefix+"/policy/retainedMin", int32(p.KeyframeMinInterval()))

			for i, target := range []int{steps + 8, -3, 2, 2, steps, 1, 0, steps + 1} {
				p.SeekFrame(target)
				recordingObserve(t, r, p, fmt.Sprintf("%s/seek%d", prefix, i))
			}

			r.Bool(prefix+"/policy/populated", p.KeyframeBytes() > 0)
			r.Bool(prefix+"/policy/bounded", p.KeyframeBytes() <= p.KeyframeBudget())
			// A one-byte budget is native to each engine: no native snapshot byte charge is compared.
			p.SetKeyframePolicy(1, 1)
			p.Restart()

			for p.StepFrame() {
			}

			r.Bool(prefix+"/tight/bounded", p.KeyframeBytes() <= p.KeyframeBudget())
			r.Bool(prefix+"/tight/softBudget", p.KeyframeBytes() <= p.KeyframeBudget() || box2d.RecordingKeyframeCount(p) == 1)
			r.Bool(prefix+"/tight/coarsened", p.KeyframeInterval() > p.KeyframeMinInterval())
			p.SeekFrame(3)
			recordingObserve(t, r, p, prefix+"/tight/seek")
			destroy()
			r.Bool(prefix+"/destroy/worldValid", world.IsValid())
		}
	}
}

func TestRecordingOwnedSourceParity(t *testing.T) {
	r := oracle.Open(t, "recording/OwnedSourceParity")
	rec := recordingLifecycleSource(t, 16)
	source := rec.Data()

	p, destroy := recordingParityPlayer(t, source, 0)
	for i := range source {
		source[i] = 0xA5
	}

	recordingObserve(t, r, p, "open")

	for p.StepFrame() {
		recordingObserve(t, r, p, fmt.Sprintf("step%d", p.Frame()))
	}

	p.Restart()
	recordingObserve(t, r, p, "restart")

	for i, target := range []int{16, 1, 0, 7, 99} {
		p.SeekFrame(target)
		recordingObserve(t, r, p, fmt.Sprintf("seek%d", i))
	}

	w := p.World()

	destroy()
	r.Bool("destroy/worldValid", w.IsValid())
}

func TestRecordingDivergenceParity(t *testing.T) {
	r := oracle.Open(t, "recording/DivergenceParity")
	for _, kind := range []string{"state", "query"} {
		rec := recordingLifecycleSource(t, 16)
		data := box2d.RecordingMismatch(rec.Data(), kind)
		r.Bool(kind+"/validate", box2d.ValidateReplay(data, 0) == nil)
		p, destroy := recordingParityPlayer(t, data, 0)
		p.SetKeyframePolicy(1<<26, 2)
		recordingObserve(t, r, p, kind+"/open")

		for p.StepFrame() {
			recordingObserve(t, r, p, fmt.Sprintf("%s/step%d", kind, p.Frame()))
		}

		p.Restart()
		recordingObserve(t, r, p, kind+"/restart")

		for i, target := range []int{6, 7, 16, 5, 12, 0, 16} {
			p.SeekFrame(target)
			recordingObserve(t, r, p, fmt.Sprintf("%s/seek%d", kind, i))
		}

		destroy()
	}
}

func TestRecordingFileAndFramingParity(t *testing.T) {
	r := oracle.Open(t, "recording/FileAndFramingParity")
	rec := recordingLifecycleSource(t, 1)
	root := t.TempDir()
	err := box2d.RecordingSaveFile(rec, filepath.Join(root, "absent", "out.b2rec"))

	var pe *os.PathError

	r.Bool("save/missingParent", err != nil)

	if !errors.As(err, &pe) {
		t.Fatalf("save error category: %v", err)
	}

	loaded, err := box2d.RecordingLoadFile(filepath.Join(root, "missing.b2rec"))
	r.Bool("load/missing", loaded == nil && errors.Is(err, os.ErrNotExist))
	loaded, err = box2d.RecordingLoadFile(root)
	r.Bool("load/directory", loaded == nil && err != nil)

	if !errors.As(err, &pe) {
		t.Fatalf("directory error category: %v", err)
	}

	for _, name := range []string{"empty", "short", "magic", "major", "minor", "endian", "snapshotZero", "snapshotOversize", "incompleteFrame", "truncatedHash", "unknown", "unknownOverflow"} {
		data := box2d.RecordingMalformed(rec.Data(), name)
		p, err := box2d.NewReplay(data, 0)
		r.Bool(name+"/open", p != nil && err == nil)
		destroy := box2d.CleanupReplay(t, p)

		r.Bool(name+"/validate", box2d.ValidateReplay(data, 0) == nil)

		if p != nil {
			for p.StepFrame() {
			}

			r.I32(name+"/frame", int32(p.Frame()))
			r.Bool(name+"/diverged", p.HasDiverged())
			destroy()
		}
	}
	// Private Go pointer-width header acceptance is not claimed as native C ABI parity.
	data := append([]byte(nil), rec.Data()...)
	data[17] ^= 0xff
	_, destroy := recordingParityPlayer(t, data, 0)
	destroy()
}

// This ports the unregistered ReplayFileScrubDiag source diagnostic. It is not
// counted as another registered upstream main-suite test.
func TestRecordingDiagnosticUnregistered(t *testing.T) {
	r := oracle.Open(t, "recording/DiagnosticUnregistered")
	trace := newRecordingTrace(t, "DiagnosticUnregistered")
	path := filepath.Join(t.TempDir(), "diverge.b2rec")

	missing, err := box2d.RecordingLoadFile(path)
	if missing != nil || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing diagnostic file: %v", err)
	}

	r.I32("missing/result", 0)

	rec := recordingRecordScene(t, r, trace, recordingBuildPyramidScene, 1, 12)
	if err := box2d.RecordingSaveFile(rec, path); err != nil {
		t.Fatal(err)
	}

	loaded, err := box2d.RecordingLoadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	result := recordingScrubRecording(t, r, trace, loaded, 0)
	r.I32("present/result", int32(result))
}

func recordingParityPlayer(t *testing.T, data []byte, workers int) (*box2d.Replay, func()) {
	t.Helper()

	p, err := box2d.NewReplay(data, workers)
	if err != nil {
		t.Fatal(err)
	}

	return p, box2d.CleanupReplay(t, p)
}
