// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"

	"github.com/zigai/box2d-go/internal/scenes"
)

//go:linkname sceneRandomSeed github.com/zigai/box2d-go/internal/scenes.gRandomSeed
var sceneRandomSeed uint32

func isolateSceneRandomSeed(t *testing.T) {
	t.Helper()

	seed := sceneRandomSeed

	t.Cleanup(func() { sceneRandomSeed = seed })

	sceneRandomSeed = 12345
}

func TestSceneRandomSeedIsolation(t *testing.T) {
	seed := sceneRandomSeed

	t.Cleanup(func() { sceneRandomSeed = seed })

	var first uint32

	for i := range 2 {
		sceneRandomSeed = uint32(17 + i)
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			isolateSceneRandomSeed(t)

			if sceneRandomSeed != 12345 {
				t.Fatal("condition false")
			}

			bits := math.Float32bits(scenes.RandomFloatRange(-0.5, 0.5))
			if i == 0 {
				first = bits
			} else if bits != first {
				t.Fatal("condition false")
			}
		})

		if sceneRandomSeed != uint32(17+i) {
			t.Fatal("condition false")
		}
	}
}

func testScenes(t *testing.T, workers int) {
	t.Helper()

	cases := []sceneTestCase{
		{name: "JointGrid", create: scenes.CreateJointGrid, horizon: 500},
		{name: "LargePyramid", create: scenes.CreateLargePyramid, horizon: 500},
		{name: "ManyPyramids", create: scenes.CreateManyPyramids, horizon: 300},
		{name: "Rain", create: scenes.CreateRain, step: scenes.StepRain, horizon: 1000},
		{name: "Spinner", create: scenes.CreateSpinner, step: scenes.StepSpinner, horizon: 500},
		{name: "Smash", create: scenes.CreateSmash, horizon: 300},
		{name: "Tumbler", create: scenes.CreateTumbler, horizon: 750},
		{name: "Washer", create: scenes.CreateWasher, horizon: 500},
		{name: "Junkyard", create: scenes.CreateJunkyard, step: scenes.StepJunkyard, horizon: 800},
		{name: "Sleep", create: scenes.CreateSleep, step: scenes.StepSleep, horizon: 300},
		{name: "Compounds", create: scenes.CreateCompounds, horizon: 500},
		{
			name:    "Queries",
			create:  scenes.CreateQueries,
			step:    scenes.StepQueries,
			stats:   scenes.GetQueryBenchmarkStats,
			horizon: 300,
		},
		{
			name:    "TreeCast",
			create:  scenes.CreateTreeCast,
			step:    scenes.StepTreeCast,
			stats:   scenes.GetTreeCastBenchmarkStats,
			destroy: scenes.DestroyTreeCast,
			horizon: 300,
		},
		{
			name:    "TileWorld",
			create:  scenes.CreateTileWorld,
			step:    scenes.StepTileWorld,
			stats:   scenes.GetTileWorldBenchmarkStats,
			destroy: scenes.DestroyTileWorld,
			horizon: 300,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testBenchmarkScene(t, workers, tc)
		})
	}

	t.Run("FallingHinges", func(t *testing.T) {
		testFallingHingesScene(t, workers)
	})
	t.Run("Human", func(t *testing.T) {
		testHumanScene(t, workers)
	})
}

func checkScene(
	t *testing.T,
	r *oracle.Reader,
	full map[int][]byte, actual map[int][]byte,
	world *box2d.World,
	step int,
) {
	t.Helper()

	counters := world.Counters()
	r.I32(fmt.Sprintf("%d.b", step), int32(counters.BodyCount))
	r.I32(fmt.Sprintf("%d.s", step), int32(counters.ShapeCount))

	transforms := box2d.SceneBodyTransformsForTest(world)
	hash := uint64(14695981039346656037)

	var stateBytes []byte
	for _, body := range transforms {
		stateBytes = binary.LittleEndian.AppendUint32(stateBytes, uint32(body.Index))

		start := len(stateBytes)
		if unsafe.Sizeof(box2d.Pos{}) == 16 {
			stateBytes = binary.LittleEndian.AppendUint64(
				stateBytes,
				math.Float64bits(float64(body.Transform.P.X)),
			)
			stateBytes = binary.LittleEndian.AppendUint64(
				stateBytes,
				math.Float64bits(float64(body.Transform.P.Y)),
			)
		} else {
			stateBytes = binary.LittleEndian.AppendUint32(
				stateBytes,
				math.Float32bits(float32(body.Transform.P.X)),
			)
			stateBytes = binary.LittleEndian.AppendUint32(
				stateBytes,
				math.Float32bits(float32(body.Transform.P.Y)),
			)
		}

		stateBytes = binary.LittleEndian.AppendUint32(
			stateBytes,
			math.Float32bits(body.Transform.Q.C),
		)

		stateBytes = binary.LittleEndian.AppendUint32(
			stateBytes,
			math.Float32bits(body.Transform.Q.S),
		)
		for _, value := range stateBytes[start:] {
			hash ^= uint64(value)
			hash *= 1099511628211
		}
	}

	if want, ok := full[step]; ok {
		diagnoseSceneFull(t, step, want, stateBytes)
	}

	r.U64(fmt.Sprintf("%d.h", step), hash)
	checkSceneObservations(t, r, world, step)

	if step == -1 || (step < 300 && step%60 == 59) {
		actual[step] = stateBytes
	}
}

func checkSceneObservations(t *testing.T, r *oracle.Reader, world *box2d.World, step int) {
	t.Helper()

	for i, state := range box2d.SceneStateForTest(world) {
		lane := []string{"dynamic", "contact", "joint"}[i]
		r.I32(fmt.Sprintf("%d.%s.count", step, lane), int32(state.Count))
		r.U64(fmt.Sprintf("%d.%s.hash", step, lane), state.Hash)
	}

	for i, state := range box2d.SceneEventsForTest(world) {
		lane := []string{"move", "sensorBegin", "sensorEnd", "contactBegin", "contactEnd", "hit", "jointEvent"}[i]
		r.I32(fmt.Sprintf("%d.%s.count", step, lane), int32(state.Count))
		r.U64(fmt.Sprintf("%d.%s.hash", step, lane), state.Hash)
	}
}

func diagnoseSceneFull(t *testing.T, step int, want []byte, got []byte) {
	t.Helper()

	stride := 2*sceneWidth() + 12
	if len(want) != len(got) {
		t.Fatalf(
			"step %d: full-state body count: C %d, Go %d",
			step,
			len(want)/stride,
			len(got)/stride,
		)
	}

	for offset, value := range want {
		if got[offset] == value {
			continue
		}

		bodyOffset := offset / stride * stride
		body := binary.LittleEndian.Uint32(want[bodyOffset : bodyOffset+4])
		component := "body ID"
		byteInBody := offset % stride

		positionWidth := int(unsafe.Sizeof(box2d.Pos{})) / 2
		switch {
		case byteInBody < 4:
		case byteInBody < 4+positionWidth:
			component = "P.X"
		case byteInBody < 4+2*positionWidth:
			component = "P.Y"
		case byteInBody < 8+2*positionWidth:
			component = "Q.C"
		default:
			component = "Q.S"
		}

		t.Fatalf(
			"step %d: body %d, %s, byte %d: C 0x%02x, Go 0x%02x",
			step,
			body,
			component,
			byteInBody,
			value,
			got[offset],
		)
	}
}

var sceneSteps = []int{-1, 59, 119, 179, 239, 299}

func sceneWidth() int { return int(unsafe.Sizeof(box2d.Pos{})) / 2 }

func encodeSceneBundle(full map[int][]byte, width int) []byte {
	stride := 2*width + 12

	var data []byte

	for _, step := range sceneSteps {
		state := full[step]
		data = binary.LittleEndian.AppendUint32(data, uint32(int32(step)))
		data = binary.LittleEndian.AppendUint32(data, uint32(len(state)/stride))
		data = append(data, state...)
	}

	return data
}

//nolint:err113 // Test-only decoding diagnostics; consumers report the failure, not its identity.
func decodeSceneBundle(p []byte, width int) (map[int][]byte, error) {
	stride := 2*width + 12
	states := make(map[int][]byte, len(sceneSteps))

	for _, step := range sceneSteps {
		if len(p) < 8 {
			return nil, errors.New("short checkpoint header")
		}

		if got := int(int32(binary.LittleEndian.Uint32(p))); got != step {
			return nil, fmt.Errorf("checkpoint step %d, want %d", got, step)
		}

		count := int(binary.LittleEndian.Uint32(p[4:]))
		p = p[8:]

		if count > 100000 || len(p) < count*stride {
			return nil, fmt.Errorf("checkpoint %d: invalid body count %d", step, count)
		}

		state := p[:count*stride]
		p = p[count*stride:]

		for i := stride; i < len(state); i += stride {
			if binary.LittleEndian.Uint32(
				state[i:],
			) <= binary.LittleEndian.Uint32(
				state[i-stride:],
			) {
				return nil, fmt.Errorf("checkpoint %d: body IDs are not ascending", step)
			}
		}

		states[step] = state
	}

	if len(p) != 0 {
		return nil, errors.New("trailing checkpoint bytes")
	}

	return states, nil
}

type sceneRecord struct{ Key, Type, Value string }

func parseSceneRecord(line []byte) (sceneRecord, error) {
	var record sceneRecord

	dec := json.NewDecoder(bytes.NewReader(line))

	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return record, errors.New( //nolint:err113 // Test codec error; callers never match it.
			"invalid record object",
		)
	}

	seen := make(map[string]bool)

	for dec.More() {
		if err := decodeSceneRecordField(dec, &record, seen); err != nil {
			return record, err
		}
	}

	if _, err := dec.Token(); err != nil {
		return record, fmt.Errorf("close scene record object: %w", err)
	}

	if len(seen) != 3 {
		return record, errors.New( //nolint:err113 // Test codec error; callers never match it.
			"missing record key",
		)
	}

	if _, err := dec.Token(); err != io.EOF {
		return record, errors.New( //nolint:err113 // Test codec error; callers never match it.
			"trailing record data",
		)
	}

	if record.Key == "" {
		return record, errors.New( //nolint:err113 // Test codec error; callers never match it.
			"empty record key",
		)
	}

	return record, nil
}

func validateSceneValue(record sceneRecord, width int) error {
	fields := strings.Split(record.Value, ",")
	length, count := 8, 1

	switch record.Type {
	case "i32", "u32", "f32":
	case "u64":
		length = 16
	case "vec2", "rot":
		count = 2
	case "pos":
		count = 2
		length = 2 * width
	case "bool":
		if record.Value != "0" && record.Value != "1" {
			return errors.New( //nolint:err113 // Test codec error; callers never match it.
				"invalid bool",
			)
		}

		return nil
	case "bytes":
		if _, err := hex.DecodeString(record.Value); err != nil {
			return fmt.Errorf("decode scene byte value: %w", err)
		}

		return nil
	default:
		return errors.New( //nolint:err113 // Test codec error; callers never match it.
			"unknown primitive type",
		)
	}

	return validateScenePrimitive(fields, length, count)
}

func readSceneRecords(
	t *testing.T,
	config string,
	name string,
) ([]sceneRecord, []byte, map[int][]byte) {
	t.Helper()

	data, err := oracle.ReadFile(config, "scenes/"+name)
	if err != nil {
		t.Fatal(err)
	}

	width := 4
	if config == "double" {
		width = 8
	}

	records, packed := decodeSceneRecords(t, data, width)

	full, err := decodeSceneBundle(packed, width)
	if err != nil {
		t.Fatal(err)
	}

	return records, packed, full
}

func readSceneFull(t *testing.T, name string) map[int][]byte {
	t.Helper()

	config := "float"
	if sceneWidth() == 8 {
		config = "double"
	}

	_, _, full := readSceneRecords(t, config, name)

	return full
}

func TestSceneRecordParsing(t *testing.T) {
	checkSceneRecordCodec(t)
}

func sceneHash(state []byte, width int) uint64 {
	hash := uint64(14695981039346656037)

	stride := 2*width + 12
	for offset := 0; offset < len(state); offset += stride {
		for _, value := range state[offset+4 : offset+stride] {
			hash ^= uint64(value)
			hash *= 1099511628211
		}
	}

	return hash
}

type sceneFixtureCursor struct {
	t       *testing.T
	records []sceneRecord
	next    int
	full    map[int][]byte
	width   int
}

func (c *sceneFixtureCursor) take(key string, kind string) sceneRecord {
	c.t.Helper()

	if c.next >= len(c.records) {
		c.t.Fatalf("missing %s", key)
	}

	r := c.records[c.next]

	c.next++
	if r.Key != key || r.Type != kind {
		c.t.Fatalf("record %d: want %s/%s, got %s/%s", c.next, key, kind, r.Key, r.Type)
	}

	return r
}

func (c *sceneFixtureCursor) count(key string) int {
	c.t.Helper()

	r := c.take(key, "i32")

	value, err := strconv.ParseUint(r.Value, 16, 32)
	if err != nil || len(r.Value) != 8 {
		c.t.Fatalf("invalid count %s", key)
	}

	return int(int32(value))
}

func (c *sceneFixtureCursor) observations(step int) {
	c.t.Helper()

	for _, lane := range []string{
		"dynamic", "contact", "joint", "move", "sensorBegin",
		"sensorEnd", "contactBegin", "contactEnd", "hit", "jointEvent",
	} {
		if c.count(fmt.Sprintf("%d.%s.count", step, lane)) < 0 {
			c.t.Fatal("negative observation count")
		}

		c.take(fmt.Sprintf("%d.%s.hash", step, lane), "u64")
	}
}

func (c *sceneFixtureCursor) trace(step int) {
	c.t.Helper()

	bodies := c.count(fmt.Sprintf("%d.b", step))

	shapes := c.count(fmt.Sprintf("%d.s", step))
	if bodies < 0 || shapes < 0 {
		c.t.Fatal("negative body/shape count")
	}

	h := c.take(fmt.Sprintf("%d.h", step), "u64")

	value, err := strconv.ParseUint(h.Value, 16, 64)
	if err != nil || len(h.Value) != 16 {
		c.t.Fatal("invalid hash")
	}

	if state, ok := c.full[step]; ok {
		if len(state)/(2*c.width+12) != bodies {
			c.t.Fatalf("step %d count differs from full packet", step)
		}

		if got := sceneHash(state, c.width); got != value {
			c.t.Fatalf("step %d full packet FNV: %016x != %016x", step, got, value)
		}
	}

	c.observations(step)
}

func (c *sceneFixtureCursor) specialStep(name string, step int) {
	c.t.Helper()

	switch name {
	case "FallingHinges":
		c.take(fmt.Sprintf("%d.done", step), "bool")
		c.take(fmt.Sprintf("%d.hash", step), "u32")
		c.count(fmt.Sprintf("%d.sleep", step))
		c.count(fmt.Sprintf("%d.step", step))
	case "Human":
		c.humanStep(step)
	}
}

func checkSceneWorkerFixture(
	t *testing.T,
	config string,
	fixture string,
	name string,
	workers int, width int,
) int64 {
	t.Helper()

	records, packed, full := readSceneRecords(t, config, fixture)
	if !bytes.Equal(encodeSceneBundle(full, width), packed) {
		t.Fatal("C/Go deterministic encoding mismatch")
	}

	c := sceneFixtureCursor{t: t, records: records, full: full, width: width}
	c.trace(-1)

	c.initialMetadata(name)

	stepped, stats := false, false

	switch name {
	case "Rain", "Spinner", "Junkyard", "Sleep":
		stepped = true
	case "Queries", "TreeCast", "TileWorld":
		stepped, stats = true, true
	}

	horizon := map[string]int{
		"JointGrid": 500, "LargePyramid": 500, "ManyPyramids": 300, "Rain": 1000,
		"Spinner": 500, "Smash": 300, "Tumbler": 750, "Washer": 500,
		"Junkyard": 800, "Sleep": 300, "Compounds": 500, "Queries": 300,
		"TreeCast": 300, "TileWorld": 300, "FallingHinges": 300, "Human": 300,
	}[name]
	for step := range horizon {
		if stepped {
			c.take(fmt.Sprintf("%d.v", step), "f32")
		}

		c.specialStep(name, step)

		c.trace(step)

		if stats {
			c.count(fmt.Sprintf("%d.n", step))
			c.count(fmt.Sprintf("%d.l", step))
		}
	}

	if name == "Human" {
		c.trace(300)
	}

	c.take("full", "bytes")

	if c.next != len(records) {
		t.Fatal("extra fixture records")
	}

	info, err := os.Stat("../testdata/oracle/" + config + "/scenes/" + fixture + ".jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf(
		"%d bytes; %d ordered steps; six full checkpoints; workers %d; width %d",
		info.Size(),
		horizon,
		workers,
		width,
	)

	return info.Size()
}

func checkSceneFixtureDirectory(t *testing.T, config string, names []string) {
	t.Helper()

	entries, err := os.ReadDir("../testdata/oracle/" + config + "/scenes")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 43 {
		t.Fatalf(
			"want 43 area fixtures (32 worker traces, six query payloads, two extended helpers, "+
				"Utils/Trig/Helpers), got %d",
			len(entries),
		)
	}

	expectedNames := map[string]bool{
		"Utils.jsonl.gz":   true,
		"Trig.jsonl.gz":    true,
		"Helpers.jsonl.gz": true,
	}

	for _, workers := range []int{1, 4} {
		for _, name := range names {
			expectedNames[fmt.Sprintf("%s_workers%d.jsonl.gz", name, workers)] = true
		}

		for _, name := range []string{"QueriesPayload", "TreeCastPayload", "TileWorldPayload", "ExtendedHelpers"} {
			expectedNames[fmt.Sprintf("%s_workers%d.jsonl.gz", name, workers)] = true
		}
	}

	for _, entry := range entries {
		if !expectedNames[entry.Name()] || entry.IsDir() {
			t.Fatalf("unexpected scene fixture %s", entry.Name())
		}

		delete(expectedNames, entry.Name())
	}

	if len(expectedNames) != 0 {
		t.Fatalf("missing scene fixtures: %v", expectedNames)
	}
}

func checkSceneUtilityFixture(t *testing.T, config string, width int) {
	t.Helper()

	data, err := oracle.ReadFile(config, "scenes/Utils")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) > 1048576 {
		t.Fatal("utils exceeds default budget")
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	c := sceneLineCursor{t: t, lines: lines, width: width}

	for step := range 300 {
		take := func(suffix, kind string) sceneRecord {
			return c.utilityRecord(step, suffix, kind)
		}
		for _, field := range []struct{ key, kind string }{
			{"int", "i32"},
			{"range", "i32"},
			{"float", "f32"},
			{"floatRange", "f32"},
			{"vec", "vec2"},
			{"pos", "pos"},
			{"rot", "rot"},
		} {
			take(field.key, field.kind)
		}

		r := take("count", "i32")

		n, err := strconv.ParseUint(r.Value, 16, 32)
		if err != nil || n < 3 || n > 8 {
			t.Fatal("invalid polygon count")
		}

		take("radius", "f32")
		take("centroid", "vec2")

		for range int(n) {
			take("vertex", "vec2")
			take("normal", "vec2")
		}
	}

	c.checkUtilityBoundary()

	if c.next != len(lines) {
		t.Fatal("extra utility observations")
	}
}

type sceneTestCase struct {
	name    string
	create  func(*box2d.World)
	step    func(*box2d.World, int) float32
	stats   func() box2d.TreeStats
	destroy func()
	horizon int
}

func testBenchmarkScene(t *testing.T, workers int, tc sceneTestCase) {
	t.Helper()
	isolateSceneRandomSeed(t)

	def := box2d.DefaultWorldDef()
	def.WorkerCount = workers

	world := box2d.NewWorld(&def)
	defer world.Close()

	if tc.destroy != nil {
		defer tc.destroy()
	}

	tc.create(world)
	fixture := fmt.Sprintf("%s_workers%d", tc.name, workers)
	r := oracle.Open(t, "scenes/"+fixture)
	full := readSceneFull(t, fixture)
	actual := make(map[int][]byte)
	checkScene(t, r, full, actual, world, -1)

	if tc.name == "ManyPyramids" {
		c := scenes.GetManyPyramidsCapacity()
		r.I32("staticShapeCount", int32(c.StaticShapeCount))
		r.I32("staticBodyCount", int32(c.StaticBodyCount))
		r.I32("dynamicShapeCount", int32(c.DynamicShapeCount))
		r.I32("dynamicBodyCount", int32(c.DynamicBodyCount))
		r.I32("contactCount", int32(c.ContactCount))
	}

	if tc.name == "Queries" {
		r.I32("queryCount", int32(scenes.GetQueryBenchmarkCount()))
		r.F32("extent", scenes.GetQueryBenchmarkExtent())

		for _, index := range []int{0, 999} {
			origin, translation := scenes.GetQueryBenchmarkRay(index)
			r.Pos(fmt.Sprintf("origin%d", index), origin.X, origin.Y)
			r.Vec2(fmt.Sprintf("translation%d", index), translation.X, translation.Y)
		}
	}

	for step := range tc.horizon {
		if tc.step != nil {
			r.F32(fmt.Sprintf("%d.v", step), tc.step(world, step))
		}

		world.Step(float32(1)/float32(60), 4)
		checkScene(t, r, full, actual, world, step)

		if tc.stats != nil {
			stats := tc.stats()
			r.I32(fmt.Sprintf("%d.n", step), int32(stats.NodeVisits))
			r.I32(fmt.Sprintf("%d.l", step), int32(stats.LeafVisits))
		}
	}

	r.Bytes("full", encodeSceneBundle(actual, sceneWidth()))
}

func testFallingHingesScene(t *testing.T, workers int) {
	t.Helper()

	def := box2d.DefaultWorldDef()
	def.WorkerCount = workers

	world := box2d.NewWorld(&def)
	defer world.Close()

	data := scenes.CreateFallingHinges(world)
	fixture := fmt.Sprintf("FallingHinges_workers%d", workers)
	r := oracle.Open(t, "scenes/"+fixture)
	full := readSceneFull(t, fixture)
	actual := make(map[int][]byte)

	defer scenes.DestroyFallingHinges(&data)

	checkScene(t, r, full, actual, world, -1)

	for step := range 300 {
		world.Step(float32(1)/float32(60), 4)
		done := scenes.UpdateFallingHinges(world, &data)
		r.Bool(fmt.Sprintf("%d.done", step), done)
		r.U32(fmt.Sprintf("%d.hash", step), data.Hash)
		r.I32(fmt.Sprintf("%d.sleep", step), int32(data.SleepStep))
		r.I32(fmt.Sprintf("%d.step", step), int32(data.StepCount))
		checkScene(t, r, full, actual, world, step)
	}

	r.Bytes("full", encodeSceneBundle(actual, sceneWidth()))
}

func testHumanScene(t *testing.T, workers int) {
	t.Helper()
	isolateSceneRandomSeed(t)

	def := box2d.DefaultWorldDef()
	def.WorkerCount = workers

	world := box2d.NewWorld(&def)
	defer world.Close()

	var human scenes.Human
	scenes.CreateHuman(&human, world, box2d.Pos{X: 0, Y: 5}, 1, 0.05, 5, 0.5, 1, nil, true)

	fixture := fmt.Sprintf("Human_workers%d", workers)
	r := oracle.Open(t, "scenes/"+fixture)
	full := readSceneFull(t, fixture)
	actual := make(map[int][]byte)
	checkScene(t, r, full, actual, world, -1)

	for step := range 300 {
		applySceneHumanAction(&human, step)

		checkSceneHumanSensors(t, r, &human, step)

		if step == 0 || (step >= 20 && step <= 180 && step%20 == 0) {
			checkSceneObservations(t, r, world, 2000+step)
		}

		world.Step(float32(1)/float32(60), 4)
		checkScene(t, r, full, actual, world, step)
	}

	scenes.DestroyHuman(&human)
	checkScene(t, r, full, actual, world, 300)
	r.Bytes("full", encodeSceneBundle(actual, sceneWidth()))
}

func applySceneHumanAction(human *scenes.Human, step int) {
	switch step {
	case 0:
		human.SetVelocity(box2d.Vec2{X: 1})
	case 20:
		human.ApplyRandomAngularImpulse(0.5)
	case 40:
		human.SetJointFrictionTorque(0)
	case 60:
		human.SetJointFrictionTorque(0.1)
	case 80:
		human.SetJointSpringHertz(0)
	case 100:
		human.SetJointSpringHertz(2)
	case 120:
		human.SetJointDampingRatio(0.75)
	case 140:
		human.EnableSensorEvents(true)
	case 160:
		human.SetScale(1.5)
	case 180:
		human.EnableSensorEvents(false)
	}
}

func checkSceneHumanSensors(t *testing.T, r *oracle.Reader, human *scenes.Human, step int) {
	t.Helper()

	if step == 140 || step == 180 {
		count := int32(0)

		var enabled bool

		for shape := range human.Bones[int(scenes.BoneTorso)].BodyID.Shapes() {
			count++
			enabled = shape.SensorEventsEnabled()

			break
		}

		r.I32(fmt.Sprintf("%d.torsoShapes", step), count)
		r.Bool(fmt.Sprintf("%d.sensor", step), enabled)

		if count != 1 || enabled != (step == 140) {
			t.Fatal("condition false")
		}
	}
}

func checkSceneRecordCodec(t *testing.T) {
	t.Helper()

	for _, line := range []string{
		`{}`, `{"k":"full","t":"bytes"}`, `{"k":"full","k":"full","t":"bytes","v":""}`,
		`{"k":"full","t":"bytes","v":"","extra":""}`, `{"k":"full","t":"bytes","v":0}`,
		`{"k":"full","t":"bytes","v":""} {}`,
	} {
		if _, err := parseSceneRecord([]byte(line)); err == nil {
			t.Fatalf("accepted malformed record %s", line)
		}
	}
}

func decodeSceneRecordField(dec *json.Decoder, record *sceneRecord, seen map[string]bool) error {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("read scene record key: %w", err)
	}

	key, ok := token.(string)
	if !ok || seen[key] {
		return errors.New( //nolint:err113 // Test codec error; callers never match it.
			"duplicate record key",
		)
	}

	seen[key] = true

	var value string
	if err := dec.Decode(&value); err != nil {
		return fmt.Errorf("decode scene record value: %w", err)
	}

	switch key {
	case "k":
		record.Key = value
	case "t":
		record.Type = value
	case "v":
		record.Value = value
	default:
		return errors.New( //nolint:err113 // Test codec error; callers never match it.
			"unknown record key",
		)
	}

	return nil
}

func validateScenePrimitive(fields []string, length int, count int) error {
	if len(fields) != count {
		return errors.New( //nolint:err113 // Test codec error; callers never match it.
			"invalid primitive components",
		)
	}

	for _, field := range fields {
		if len(field) != length {
			return errors.New( //nolint:err113 // Test codec error; callers never match it.
				"invalid primitive width",
			)
		}

		if _, err := strconv.ParseUint(field, 16, 4*length); err != nil {
			return fmt.Errorf("decode scene primitive component: %w", err)
		}
	}

	return nil
}

func decodeSceneRecords(t *testing.T, data []byte, width int) ([]sceneRecord, []byte) {
	t.Helper()

	var (
		records []sceneRecord
		packed  []byte
	)

	for line := range bytes.SplitSeq(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}

		record, err := parseSceneRecord(line)
		if err != nil {
			t.Fatal(err)
		}

		if err := validateSceneValue(record, width); err != nil {
			t.Fatal(err)
		}

		records = append(records, record)
		if record.Key == "full" {
			if packed != nil || record.Type != "bytes" {
				t.Fatal("duplicate or mistyped full packet")
			}

			packed, err = hex.DecodeString(record.Value)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if len(records) == 0 || records[len(records)-1].Key != "full" || packed == nil {
		t.Fatal("missing final full packet")
	}

	return records, packed
}

type sceneLineCursor struct {
	t     *testing.T
	lines [][]byte
	next  int
	width int
}

func (c *sceneLineCursor) helperValue(key string, kind string) string {
	c.t.Helper()

	if c.next >= len(c.lines) {
		c.t.Fatalf("missing %s", key)
	}

	r, err := parseSceneRecord(c.lines[c.next])
	if err != nil {
		c.t.Fatal(err)
	}

	c.next++

	if r.Key != key || r.Type != kind {
		c.t.Fatalf("want %s/%s, got %s/%s", key, kind, r.Key, r.Type)
	}

	if err := validateSceneValue(r, c.width); err != nil {
		c.t.Fatal(err)
	}

	return r.Value
}

func (c *sceneLineCursor) utilityRecord(step int, suffix string, kind string) sceneRecord {
	c.t.Helper()

	if c.next >= len(c.lines) {
		c.t.Fatal("short utilities fixture")
	}

	r, err := parseSceneRecord(c.lines[c.next])
	c.next++

	if err == nil {
		err = validateSceneValue(r, c.width)
	}

	if err != nil || r.Key != fmt.Sprintf("%d.%s", step, suffix) || r.Type != kind {
		c.t.Fatalf("invalid utility record %d: %v", c.next, err)
	}

	return r
}

func (c *sceneLineCursor) checkGroupFlags(step int) {
	c.t.Helper()

	spawned, cleared := "1", "0"
	if step == 1 {
		spawned, cleared = "0", "1"
	}

	if c.helperValue(fmt.Sprintf("%d.spawned", step), "bool") != spawned ||
		c.helperValue(fmt.Sprintf("%d.cleared", step), "bool") != cleared {
		c.t.Fatal("group lifecycle flags")
	}

	if c.helperValue(fmt.Sprintf("%d.column", step), "i32") != "00000000" {
		c.t.Fatal("recycling index did not wrap")
	}

	if step > 0 {
		if c.helperValue(fmt.Sprintf("%d.oldBody", step), "bool") != "0" ||
			c.helperValue(fmt.Sprintf("%d.oldJoint", step), "bool") != "0" {
			c.t.Fatal("destroyed handles remain valid")
		}
	}

	if step == 1 && c.helperValue("2.v", "f32") != "00000000" {
		c.t.Fatal("StepRain return")
	}
}

func (c *sceneLineCursor) checkGroupLifecycle() {
	c.t.Helper()

	counts := make(map[string]string)

	for step := -1; step <= 2; step++ {
		for _, field := range []string{"b", "s", "j"} {
			key := fmt.Sprintf("%d.%s", step, field)

			counts[key] = c.helperValue(key, "i32")
			if step == 1 && counts[key] != counts["-1."+field] {
				c.t.Fatal("DestroyGroup did not restore world counts")
			}

			if step == 2 && counts[key] != counts["0."+field] {
				c.t.Fatal("recycling changed world counts")
			}
		}

		if step == -1 {
			continue
		}

		c.checkGroupFlags(step)
	}
}

func (c *sceneLineCursor) checkPaddleBoundaries() {
	c.t.Helper()

	for i, velocity := range []string{
		"c0a00000,00000000", "40a00000,00000000", "40a00000,00000000", "c0a00000,00000000",
	} {
		c.helperValue(fmt.Sprintf("%d.q", i), "f32")
		c.helperValue(fmt.Sprintf("%d.position", i), "pos")

		if c.helperValue(fmt.Sprintf("%d.velocity", i), "vec2") != velocity {
			c.t.Fatal("strict paddle boundary reversal")
		}

		c.helperValue(fmt.Sprintf("%d.n", i), "i32")
		c.helperValue(fmt.Sprintf("%d.l", i), "i32")
	}
}

func (c *sceneLineCursor) checkUtilityBoundary() {
	c.t.Helper()

	boundary := make([]struct{ key, kind string }, 0, 16)

	boundary = append(boundary, []struct{ key, kind string }{
		{"equal.int", "i32"},
		{"equal.float", "f32"},
		{"equal.vec", "vec2"},
		{"equal.pos", "pos"},
		{"equal.seed", "u32"},
		{"fallback.count", "i32"},
		{"fallback.radius", "f32"},
		{"fallback.centroid", "vec2"},
	}...)
	for range 4 {
		boundary = append(
			boundary,
			struct{ key, kind string }{"fallback.vertex", "vec2"},
			struct{ key, kind string }{"fallback.normal", "vec2"},
		)
	}

	for _, field := range boundary {
		if c.next >= len(c.lines) {
			c.t.Fatalf("missing utility boundary %s", field.key)
		}

		record, err := parseSceneRecord(c.lines[c.next])
		c.next++

		if err != nil {
			c.t.Fatal(err)
		}

		if record.Key != field.key || record.Type != field.kind {
			c.t.Fatalf(
				"utility boundary: want %s/%s, got %s/%s",
				field.key,
				field.kind,
				record.Key,
				record.Type,
			)
		}

		if err := validateSceneValue(record, c.width); err != nil {
			c.t.Fatal(err)
		}

		if field.key == "fallback.count" && record.Value != "00000004" {
			c.t.Fatal("C fallback is not a square")
		}
	}
}

func (c *sceneFixtureCursor) humanStep(step int) {
	c.t.Helper()

	if step == 140 || step == 180 {
		if c.count(fmt.Sprintf("%d.torsoShapes", step)) != 1 {
			c.t.Fatal("missing torso shape")
		}

		sensor := c.take(fmt.Sprintf("%d.sensor", step), "bool")

		expected := "0"
		if step == 140 {
			expected = "1"
		}

		if sensor.Value != expected {
			c.t.Fatal("sensor toggle differs")
		}
	}

	if step == 0 || (step >= 20 && step <= 180 && step%20 == 0) {
		c.observations(2000 + step)
	}
}

func (c *sceneFixtureCursor) initialMetadata(name string) {
	c.t.Helper()

	if name == "ManyPyramids" {
		for _, key := range []string{
			"staticShapeCount", "staticBodyCount", "dynamicShapeCount", "dynamicBodyCount", "contactCount",
		} {
			c.count(key)
		}
	}

	if name == "Queries" {
		c.count("queryCount")
		c.take("extent", "f32")

		for _, i := range []int{0, 999} {
			c.take(fmt.Sprintf("origin%d", i), "pos")
			c.take(fmt.Sprintf("translation%d", i), "vec2")
		}
	}
}
