// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"

	"github.com/zigai/box2d/internal/scenes"
)

//go:linkname sceneRandomSeed github.com/zigai/box2d/internal/scenes.gRandomSeed
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
	for i := 0; i < 2; i++ {
		sceneRandomSeed = uint32(17 + i)
		t.Run(strconv.Itoa(i), func(t *testing.T) {

			isolateSceneRandomSeed(t)
			if !(sceneRandomSeed == 12345) {
				t.Fatal("condition false")
			}
			bits := math.Float32bits(scenes.RandomFloatRange(-0.5, 0.5))
			if i == 0 {
				first = bits
			} else {
				if !(bits == first) {
					t.Fatal("condition false")
				}
			}
		})
		if !(sceneRandomSeed == uint32(17+i)) {
			t.Fatal("condition false")
		}
	}
}

func TestScenes(t *testing.T) { testScenes(t, 0) }

func TestParallelScenes(t *testing.T) { testScenes(t, 4) }

func testScenes(t *testing.T, workers int) {

	cases := []struct {
		name    string
		create  func(box2d.World)
		step    func(box2d.World, int) float32
		stats   func() box2d.TreeStats
		destroy func()
	}{
		{name: "JointGrid", create: scenes.CreateJointGrid},
		{name: "LargePyramid", create: scenes.CreateLargePyramid},
		{name: "ManyPyramids", create: scenes.CreateManyPyramids},
		{name: "Rain", create: scenes.CreateRain, step: scenes.StepRain},
		{name: "Spinner", create: scenes.CreateSpinner, step: scenes.StepSpinner},
		{name: "Smash", create: scenes.CreateSmash},
		{name: "Tumbler", create: scenes.CreateTumbler},
		{name: "Washer", create: scenes.CreateWasher},
		{name: "Junkyard", create: scenes.CreateJunkyard, step: scenes.StepJunkyard},
		{name: "Sleep", create: scenes.CreateSleep, step: scenes.StepSleep},
		{name: "Compounds", create: scenes.CreateCompounds},
		{name: "Queries", create: scenes.CreateQueries, step: scenes.StepQueries, stats: scenes.GetQueryBenchmarkStats},
		{name: "TreeCast", create: scenes.CreateTreeCast, step: scenes.StepTreeCast, stats: scenes.GetTreeCastBenchmarkStats, destroy: scenes.DestroyTreeCast},
		{name: "TileWorld", create: scenes.CreateTileWorld, step: scenes.StepTileWorld, stats: scenes.GetTileWorldBenchmarkStats, destroy: scenes.DestroyTileWorld},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			def := box2d.DefaultWorldDef()
			def.WorkerCount = workers
			world := box2d.CreateWorld(&def)
			defer world.Destroy()
			if tc.destroy != nil {
				defer tc.destroy()
			}
			tc.create(world)
			r := oracle.Open(t, "scenes/"+tc.name)
			full := readSceneFull(t, tc.name)
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
			for step := 0; step < 300; step++ {
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
			r.Bytes("full", packSceneBundle(actual, sceneWidth()))
		})
	}
	t.Run("FallingHinges", func(t *testing.T) {

		def := box2d.DefaultWorldDef()
		def.WorkerCount = workers
		world := box2d.CreateWorld(&def)
		defer world.Destroy()
		data := scenes.CreateFallingHinges(world)
		r := oracle.Open(t, "scenes/FallingHinges")
		full := readSceneFull(t, "FallingHinges")
		actual := make(map[int][]byte)
		defer scenes.DestroyFallingHinges(&data)
		checkScene(t, r, full, actual, world, -1)
		for step := 0; step < 300; step++ {
			world.Step(float32(1)/float32(60), 4)
			done := scenes.UpdateFallingHinges(world, &data)
			r.Bool(fmt.Sprintf("%d.done", step), done)
			r.U32(fmt.Sprintf("%d.hash", step), data.Hash)
			r.I32(fmt.Sprintf("%d.sleep", step), int32(data.SleepStep))
			r.I32(fmt.Sprintf("%d.step", step), int32(data.StepCount))
			checkScene(t, r, full, actual, world, step)
		}
		r.Bytes("full", packSceneBundle(actual, sceneWidth()))
	})
	t.Run("Human", func(t *testing.T) {

		isolateSceneRandomSeed(t)
		def := box2d.DefaultWorldDef()
		def.WorkerCount = workers
		world := box2d.CreateWorld(&def)
		defer world.Destroy()
		var human scenes.Human
		scenes.CreateHuman(&human, world, box2d.Pos{X: 0, Y: 5}, 1, 0.05, 5, 0.5, 1, nil, true)
		r := oracle.Open(t, "scenes/Human")
		full := readSceneFull(t, "Human")
		actual := make(map[int][]byte)
		checkScene(t, r, full, actual, world, -1)
		for step := 0; step < 300; step++ {
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
			if step == 140 || step == 180 {
				count := int32(0)
				var enabled bool
				for shape := range human.Bones[int(scenes.BoneTorso)].BodyID.Shapes() {
					count++
					enabled = shape.AreSensorEventsEnabled()
					break
				}
				r.I32(fmt.Sprintf("%d.torsoShapes", step), count)
				r.Bool(fmt.Sprintf("%d.sensor", step), enabled)
				if !(count == 1 && enabled == (step == 140)) {
					t.Fatal("condition false")
				}
			}
			world.Step(float32(1)/float32(60), 4)
			checkScene(t, r, full, actual, world, step)
		}
		scenes.DestroyHuman(&human)
		checkScene(t, r, full, actual, world, 300)
		r.Bytes("full", packSceneBundle(actual, sceneWidth()))
	})
}

func checkScene(t *testing.T, r *oracle.Reader, full, actual map[int][]byte, world box2d.World, step int) {
	t.Helper()
	counters := world.Counters()
	r.I32(fmt.Sprintf("%d.b", step), int32(counters.BodyCount))
	r.I32(fmt.Sprintf("%d.s", step), int32(counters.ShapeCount))
	transforms := box2d.SceneBodyTransformsForTest(world)
	hash := uint64(14695981039346656037)
	var bytes []byte
	for _, body := range transforms {
		bytes = binary.LittleEndian.AppendUint32(bytes, uint32(body.Index))
		start := len(bytes)
		if unsafe.Sizeof(box2d.Pos{}) == 16 {
			bytes = binary.LittleEndian.AppendUint64(bytes, math.Float64bits(float64(body.Transform.P.X)))
			bytes = binary.LittleEndian.AppendUint64(bytes, math.Float64bits(float64(body.Transform.P.Y)))
		} else {
			bytes = binary.LittleEndian.AppendUint32(bytes, math.Float32bits(float32(body.Transform.P.X)))
			bytes = binary.LittleEndian.AppendUint32(bytes, math.Float32bits(float32(body.Transform.P.Y)))
		}
		bytes = binary.LittleEndian.AppendUint32(bytes, math.Float32bits(body.Transform.Q.C))
		bytes = binary.LittleEndian.AppendUint32(bytes, math.Float32bits(body.Transform.Q.S))
		for _, value := range bytes[start:] {
			hash ^= uint64(value)
			hash *= 1099511628211
		}
	}
	if want, ok := full[step]; ok {
		diagnoseSceneFull(t, step, want, bytes)
	}
	r.U64(fmt.Sprintf("%d.h", step), hash)
	if step == -1 || step%60 == 59 {
		actual[step] = bytes
	}
}

func diagnoseSceneFull(t *testing.T, step int, want, got []byte) {
	t.Helper()
	stride := 2*sceneWidth() + 12
	if len(want) != len(got) {
		t.Fatalf("step %d: full-state body count: C %d, Go %d", step, len(want)/stride, len(got)/stride)
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
		t.Fatalf("step %d: body %d, %s, byte %d: C 0x%02x, Go 0x%02x", step, body, component, byteInBody, value, got[offset])
	}
}

const sceneLimit = 32 << 20
const sceneWindow = 1 << 20

var sceneSteps = []int{-1, 59, 119, 179, 239, 299}

func sceneWidth() int { return int(unsafe.Sizeof(box2d.Pos{})) / 2 }

func sceneDictionaryKey(p []byte) uint32 {
	return binary.LittleEndian.Uint32(p) * 2654435761 >> 12
}

func packSceneLZ(raw []byte) []byte {
	if len(raw) > sceneLimit {
		panic("scene packet exceeds limit")
	}
	out := binary.LittleEndian.AppendUint32(nil, uint32(len(raw)))
	table := make([]uint32, sceneWindow)
	start, i := 0, 0
	for i+4 <= len(raw) {
		key := sceneDictionaryKey(raw[i:])
		entry := table[key]
		table[key] = uint32(i) + 1
		count, previous := 0, int(entry)-1
		if entry != 0 && i-previous <= sceneWindow && bytes.Equal(raw[previous:previous+4], raw[i:i+4]) {
			count = 4
			for count < 65535 && i+count < len(raw) && raw[previous+count] == raw[i+count] {
				count++
			}
		}
		if count == 0 {
			i++
			continue
		}
		if i > start {
			out = binary.AppendUvarint(out, uint64((i-start-1)<<1))
			out = append(out, raw[start:i]...)
		}
		out = binary.AppendUvarint(out, uint64((count-4)<<1|1))
		out = binary.AppendUvarint(out, uint64(i-previous))
		for j := i + 1; j < i+count && j+4 <= len(raw); j++ {
			table[sceneDictionaryKey(raw[j:])] = uint32(j) + 1
		}
		i += count
		start = i
	}
	if start < len(raw) {
		out = binary.AppendUvarint(out, uint64((len(raw)-start-1)<<1))
		out = append(out, raw[start:]...)
	}
	return out
}

func sceneVarint(p *[]byte) (uint32, error) {
	var value uint32
	for i := 0; i < 5; i++ {
		if len(*p) == 0 {
			return 0, fmt.Errorf("truncated varint")
		}
		b := (*p)[0]
		*p = (*p)[1:]
		if i == 4 && b > 15 {
			return 0, fmt.Errorf("overflowing varint")
		}
		value |= uint32(b&127) << (7 * i)
		if b < 128 {
			if i > 0 && b == 0 {
				return 0, fmt.Errorf("noncanonical varint")
			}
			return value, nil
		}
	}
	return 0, fmt.Errorf("overflowing varint")
}

func unpackSceneLZ(p []byte) ([]byte, error) {
	if len(p) < 4 {
		return nil, fmt.Errorf("short packet")
	}
	declared := binary.LittleEndian.Uint32(p)
	p = p[4:]
	if declared > sceneLimit {
		return nil, fmt.Errorf("packet exceeds limit")
	}
	size := int(declared)
	out := make([]byte, 0, size)
	for len(out) < size {
		command, err := sceneVarint(&p)
		if err != nil {
			return nil, err
		}
		if command>>1 >= sceneLimit {
			return nil, fmt.Errorf("command exceeds limit")
		}
		count := int(command>>1) + 1
		if command&1 == 0 {
			if count > size-len(out) || count > len(p) {
				return nil, fmt.Errorf("invalid literal length")
			}
			out = append(out, p[:count]...)
			p = p[count:]
		} else {
			count += 3
			distance, err := sceneVarint(&p)
			if err != nil {
				return nil, err
			}
			if count > 65535 || count > size-len(out) || distance == 0 || distance > sceneWindow || int(distance) > len(out) {
				return nil, fmt.Errorf("invalid match")
			}
			for range count {
				out = append(out, out[len(out)-int(distance)])
			}
		}
	}
	if len(p) != 0 {
		return nil, fmt.Errorf("trailing packet bytes")
	}
	return out, nil
}

func packSceneBundle(full map[int][]byte, width int) []byte {
	components, stride := 2*width+8, 2*width+12
	raw := []byte{'S', 'C', 'N', 'L', byte(width), 6}
	total := 0
	for _, step := range sceneSteps {
		state := full[step]
		if len(state)%stride != 0 || len(state)/stride > 100000 {
			panic("invalid scene state")
		}
		raw = binary.LittleEndian.AppendUint32(raw, uint32(int32(step)))
		raw = binary.LittleEndian.AppendUint32(raw, uint32(len(state)/stride))
		total += len(state) / stride
	}
	planes := make([]byte, total*components)
	history := make(map[uint32][]byte)
	record := 0
	for _, step := range sceneSteps {
		state := full[step]
		preceding := uint32(math.MaxUint32)
		for i := 0; i < len(state); i += stride {
			id := binary.LittleEndian.Uint32(state[i:])
			if id >= sceneWindow || preceding != math.MaxUint32 && id <= preceding {
				panic("invalid scene ID")
			}
			raw = binary.AppendUvarint(raw, uint64(id-preceding))
			preceding = id
			previous := history[id]
			if previous == nil {
				previous = make([]byte, components)
				history[id] = previous
			}
			for lane := range components {
				planes[lane*total+record] = state[i+4+lane] ^ previous[lane]
				previous[lane] = state[i+4+lane]
			}
			record++
		}
	}
	return packSceneLZ(append(raw, planes...))
}

func unpackSceneBundle(p []byte, width int) (map[int][]byte, error) {
	raw, err := unpackSceneLZ(p)
	if err != nil {
		return nil, err
	}
	if len(raw) < 54 || string(raw[:4]) != "SCNL" || width != 4 && width != 8 || int(raw[4]) != width || raw[5] != 6 {
		return nil, fmt.Errorf("invalid bundle header")
	}
	counts := make([]int, 6)
	total := 0
	for k, step := range sceneSteps {
		if int32(binary.LittleEndian.Uint32(raw[6+8*k:])) != int32(step) {
			return nil, fmt.Errorf("invalid checkpoint cadence")
		}
		count := binary.LittleEndian.Uint32(raw[10+8*k:])
		if count > 100000 {
			return nil, fmt.Errorf("body count exceeds limit")
		}
		counts[k] = int(count)
		total += counts[k]
	}
	components, stride := 2*width+8, 2*width+12
	p = raw[54:]
	if len(p) < total*(components+1) {
		return nil, fmt.Errorf("short body planes")
	}
	states := make(map[int][]byte)
	maxID := uint32(0)
	for k, step := range sceneSteps {
		state := make([]byte, counts[k]*stride)
		preceding := uint32(math.MaxUint32)
		for i := range counts[k] {
			gap, err := sceneVarint(&p)
			if err != nil {
				return nil, err
			}
			id := preceding + gap
			if gap == 0 || id >= sceneWindow || preceding != math.MaxUint32 && id <= preceding {
				return nil, fmt.Errorf("invalid body order")
			}
			binary.LittleEndian.PutUint32(state[i*stride:], id)
			preceding = id
			maxID = max(maxID, id)
		}
		states[step] = state
	}
	if len(p) != total*components {
		return nil, fmt.Errorf("body plane length mismatch")
	}
	history := make([]byte, (int(maxID)+1)*components)
	record := 0
	for _, step := range sceneSteps {
		state := states[step]
		for i := 0; i < len(state); i += stride {
			id := int(binary.LittleEndian.Uint32(state[i:]))
			for lane := range components {
				value := p[lane*total+record] ^ history[id*components+lane]
				state[i+4+lane] = value
				history[id*components+lane] = value
			}
			record++
		}
	}
	return states, nil
}

type sceneRecord struct{ Key, Type, Value string }

func parseSceneRecord(line []byte) (sceneRecord, error) {
	var record sceneRecord
	dec := json.NewDecoder(bytes.NewReader(line))
	token, err := dec.Token()
	if err != nil || token != json.Delim('{') {
		return record, fmt.Errorf("invalid record object")
	}
	seen := make(map[string]bool)
	for dec.More() {
		token, err := dec.Token()
		if err != nil {
			return record, err
		}
		key, ok := token.(string)
		if !ok || seen[key] {
			return record, fmt.Errorf("duplicate record key")
		}
		seen[key] = true
		var value string
		if err := dec.Decode(&value); err != nil {
			return record, err
		}
		switch key {
		case "k":
			record.Key = value
		case "t":
			record.Type = value
		case "v":
			record.Value = value
		default:
			return record, fmt.Errorf("unknown record key")
		}
	}
	if _, err := dec.Token(); err != nil {
		return record, err
	}
	if len(seen) != 3 {
		return record, fmt.Errorf("missing record key")
	}
	if _, err := dec.Token(); err != io.EOF {
		return record, fmt.Errorf("trailing record data")
	}
	if record.Key == "" {
		return record, fmt.Errorf("empty record key")
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
			return fmt.Errorf("invalid bool")
		}
		return nil
	case "bytes":
		_, err := hex.DecodeString(record.Value)
		return err
	default:
		return fmt.Errorf("unknown primitive type")
	}
	if len(fields) != count {
		return fmt.Errorf("invalid primitive components")
	}
	for _, field := range fields {
		if len(field) != length {
			return fmt.Errorf("invalid primitive width")
		}
		if _, err := strconv.ParseUint(field, 16, 4*length); err != nil {
			return err
		}
	}
	return nil
}

func readSceneRecords(t *testing.T, config, name string) ([]sceneRecord, []byte, map[int][]byte) {
	t.Helper()
	data, err := os.ReadFile("testdata/oracle/" + config + "/scenes/" + name + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	var records []sceneRecord
	var packed []byte
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		record, err := parseSceneRecord(line)
		if err != nil {
			t.Fatal(err)
		}
		width := 4
		if config == "double" {
			width = 8
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
	width := 4
	if config == "double" {
		width = 8
	}
	full, err := unpackSceneBundle(packed, width)
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

func TestSceneCodec(t *testing.T) {

	for _, width := range []int{4, 8} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {

			states := make(map[int][]byte)
			stride := 2*width + 12
			for k, step := range sceneSteps {
				ids := []uint32{0, 2, 7}
				if k == 1 {
					ids = []uint32{2, 7, 9}
				}
				if k == 2 {
					ids = nil
				}
				for _, id := range ids {
					body := make([]byte, stride)
					binary.LittleEndian.PutUint32(body, id)
					for i := 4; i < stride; i++ {
						body[i] = byte(int(id)*31 + i + k)
					}
					if id == 0 {
						if width == 8 {
							binary.LittleEndian.PutUint64(body[4:], 0x8000000000000000)
						} else {
							binary.LittleEndian.PutUint32(body[4:], 0x80000000)
						}
						binary.LittleEndian.PutUint32(body[stride-4:], 0x7fc01234)
					}
					states[step] = append(states[step], body...)
				}
			}
			packet := packSceneBundle(states, width)
			got, err := unpackSceneBundle(packet, width)
			if err != nil {
				t.Fatal(err)
			}
			for _, step := range sceneSteps {
				if !bytes.Equal(got[step], states[step]) {
					t.Fatalf("step %d bits changed", step)
				}
			}
			if _, err := unpackSceneBundle(packet, 12-width); err == nil {
				t.Fatal("accepted wrong precision")
			}
			raw, err := unpackSceneLZ(packet)
			if err != nil {
				t.Fatal(err)
			}
			for _, offset := range []int{0, 4, 5, 6, 10, 54} {
				bad := bytes.Clone(raw)
				bad[offset] ^= 255
				if _, err := unpackSceneBundle(packSceneLZ(bad), width); err == nil {
					t.Fatalf("accepted malformed field %d", offset)
				}
			}
			for _, change := range []func([]byte){
				func(p []byte) { binary.LittleEndian.PutUint32(p[10:], 100001) },
				func(p []byte) { p[54] = 0 },
				func(p []byte) { p[55] = 0 },
				func(p []byte) { binary.LittleEndian.PutUint32(p[14:], uint32(math.MaxUint32)) },
			} {
				bad := bytes.Clone(raw)
				change(bad)
				if _, err := unpackSceneBundle(packSceneLZ(bad), width); err == nil {
					t.Fatal("accepted invalid counts, IDs or cadence")
				}
			}
			for end := 0; end < len(packet); end++ {
				if _, err := unpackSceneBundle(packet[:end], width); err == nil {
					t.Fatalf("accepted truncation %d", end)
				}
			}
		})
	}
	for _, raw := range [][]byte{nil, []byte("abcdabcdabcdabcdabcd"), bytes.Repeat([]byte{0}, 70000), bytes.Repeat([]byte{1, 2, 3, 4, 5}, 200000)} {
		got, err := unpackSceneLZ(packSceneLZ(raw))
		if err != nil || !bytes.Equal(got, raw) {
			t.Fatalf("LZ round trip: %v", err)
		}
	}
	malformed := [][]byte{
		{1, 0, 0, 0, 0x80}, {1, 0, 0, 0, 0x80, 0}, {1, 0, 0, 0, 0xff, 0xff, 0xff, 0xff, 0x10},
		{4, 0, 0, 0, 1, 0}, {4, 0, 0, 0, 0xfe, 0xff, 0xff, 0xff, 0x0f}, {4, 0, 0, 0, 1, 1}, {1, 0, 0, 0, 2, 0, 0}, {0, 0, 0, 0, 0}, {0xff, 0xff, 0xff, 0xff},
		{5, 0, 0, 0, 0, 'a', 1, 2}, {5, 0, 0, 0, 0, 'a', 1, 0x81, 0x80, 0x40},
	}
	for i, p := range malformed {
		if _, err := unpackSceneLZ(p); err == nil {
			t.Fatalf("accepted malformed LZ %d", i)
		}
	}
	for _, line := range []string{
		`{}`, `{"k":"full","t":"bytes"}`, `{"k":"full","k":"full","t":"bytes","v":""}`,
		`{"k":"full","t":"bytes","v":"","extra":""}`, `{"k":"full","t":"bytes","v":0}`, `{"k":"full","t":"bytes","v":""} {}`,
	} {
		if _, err := parseSceneRecord([]byte(line)); err == nil {
			t.Fatalf("accepted malformed record %s", line)
		}
	}
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

func TestSceneFixtureInventory(t *testing.T) {

	names := []string{"JointGrid", "LargePyramid", "ManyPyramids", "Rain", "Spinner", "Smash", "Tumbler", "Washer", "Junkyard", "Sleep", "Compounds", "Queries", "TreeCast", "TileWorld", "FallingHinges", "Human"}
	for _, config := range []string{"float", "double"} {
		t.Run(config, func(t *testing.T) {

			width := 4
			if config == "double" {
				width = 8
			}
			benchmarks := int64(0)
			for index, name := range names {
				t.Run(name, func(t *testing.T) {

					records, packed, full := readSceneRecords(t, config, name)
					if !bytes.Equal(packSceneBundle(full, width), packed) {
						t.Fatal("C/Go deterministic encoding mismatch")
					}
					next := 0
					take := func(key, kind string) sceneRecord {
						t.Helper()
						if next >= len(records) {
							t.Fatalf("missing %s", key)
						}
						r := records[next]
						next++
						if r.Key != key || r.Type != kind {
							t.Fatalf("record %d: want %s/%s, got %s/%s", next, key, kind, r.Key, r.Type)
						}
						return r
					}
					count := func(key string) int {
						t.Helper()
						r := take(key, "i32")
						value, err := strconv.ParseUint(r.Value, 16, 32)
						if err != nil || len(r.Value) != 8 {
							t.Fatalf("invalid count %s", key)
						}
						return int(int32(value))
					}
					trace := func(step int) {
						t.Helper()
						bodies := count(fmt.Sprintf("%d.b", step))
						shapes := count(fmt.Sprintf("%d.s", step))
						if bodies < 0 || shapes < 0 {
							t.Fatal("negative body/shape count")
						}
						h := take(fmt.Sprintf("%d.h", step), "u64")
						value, err := strconv.ParseUint(h.Value, 16, 64)
						if err != nil || len(h.Value) != 16 {
							t.Fatal("invalid hash")
						}
						if state, ok := full[step]; ok {
							if len(state)/(2*width+12) != bodies {
								t.Fatalf("step %d count differs from full packet", step)
							}
							if got := sceneHash(state, width); got != value {
								t.Fatalf("step %d full packet FNV: %016x != %016x", step, got, value)
							}
						}
					}
					trace(-1)
					if name == "ManyPyramids" {
						for _, key := range []string{"staticShapeCount", "staticBodyCount", "dynamicShapeCount", "dynamicBodyCount", "contactCount"} {
							count(key)
						}
					}
					if name == "Queries" {
						count("queryCount")
						take("extent", "f32")
						for _, i := range []int{0, 999} {
							take(fmt.Sprintf("origin%d", i), "pos")
							take(fmt.Sprintf("translation%d", i), "vec2")
						}
					}
					stepped := name == "Rain" || name == "Spinner" || name == "Junkyard" || name == "Sleep" || name == "Queries" || name == "TreeCast" || name == "TileWorld"
					stats := name == "Queries" || name == "TreeCast" || name == "TileWorld"
					for step := 0; step < 300; step++ {
						if stepped {
							take(fmt.Sprintf("%d.v", step), "f32")
						}
						if name == "FallingHinges" {
							take(fmt.Sprintf("%d.done", step), "bool")
							take(fmt.Sprintf("%d.hash", step), "u32")
							count(fmt.Sprintf("%d.sleep", step))
							count(fmt.Sprintf("%d.step", step))
						}
						if name == "Human" && (step == 140 || step == 180) {
							if count(fmt.Sprintf("%d.torsoShapes", step)) != 1 {
								t.Fatal("missing torso shape")
							}
							sensor := take(fmt.Sprintf("%d.sensor", step), "bool")
							expected := "0"
							if step == 140 {
								expected = "1"
							}
							if sensor.Value != expected {
								t.Fatal("sensor toggle differs")
							}
						}
						trace(step)
						if stats {
							count(fmt.Sprintf("%d.n", step))
							count(fmt.Sprintf("%d.l", step))
						}
					}
					if name == "Human" {
						trace(300)
					}
					take("full", "bytes")
					if next != len(records) {
						t.Fatal("extra fixture records")
					}
					info, err := os.Stat("testdata/oracle/" + config + "/scenes/" + name + ".jsonl")
					if err != nil {
						t.Fatal(err)
					}
					if index < 14 {
						benchmarks += info.Size()
					} else if info.Size() > 1048576 {
						t.Fatal("area exceeds default budget")
					}
					t.Logf("%d bytes; 300 ordered steps; six full checkpoints; width %d", info.Size(), width)
				})
			}
			if benchmarks > 16777216 {
				t.Fatalf("benchmark fixtures exceed named budget: %d", benchmarks)
			}
			entries, err := os.ReadDir("testdata/oracle/" + config + "/scenes")
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 19 {
				t.Fatalf("want 19 area fixtures, got %d", len(entries))
			}
			data, err := os.ReadFile("testdata/oracle/" + config + "/scenes/Utils.jsonl")
			if err != nil {
				t.Fatal(err)
			}
			if len(data) > 1048576 {
				t.Fatal("utils exceeds default budget")
			}
			lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
			next := 0
			for step := 0; step < 300; step++ {
				take := func(suffix, kind string) sceneRecord {
					t.Helper()
					if next >= len(lines) {
						t.Fatal("short utilities fixture")
					}
					r, err := parseSceneRecord(lines[next])
					next++
					if err == nil {
						err = validateSceneValue(r, width)
					}
					if err != nil || r.Key != fmt.Sprintf("%d.%s", step, suffix) || r.Type != kind {
						t.Fatalf("invalid utility record %d: %v", next, err)
					}
					return r
				}
				for _, field := range []struct{ key, kind string }{{"int", "i32"}, {"range", "i32"}, {"float", "f32"}, {"floatRange", "f32"}, {"vec", "vec2"}, {"pos", "pos"}, {"rot", "rot"}} {
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
			if next != len(lines) {
				t.Fatal("extra utility observations")
			}
		})
	}
}

func TestSceneHelperFixtureInventory(t *testing.T) {

	for _, config := range []string{"float", "double"} {
		t.Run(config, func(t *testing.T) {

			data, err := os.ReadFile("testdata/oracle/" + config + "/scenes/Helpers.jsonl")
			if err != nil {
				t.Fatal(err)
			}
			if len(data) > 1048576 {
				t.Fatal("helpers exceed default budget")
			}
			lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
			next := 0
			take := func(key, kind string) string {
				t.Helper()
				if next >= len(lines) {
					t.Fatalf("missing %s", key)
				}
				r, err := parseSceneRecord(lines[next])
				if err != nil {
					t.Fatal(err)
				}
				next++
				if r.Key != key || r.Type != kind {
					t.Fatalf("want %s/%s, got %s/%s", key, kind, r.Key, r.Type)
				}
				width := 4
				if config == "double" {
					width = 8
				}
				if err := validateSceneValue(r, width); err != nil {
					t.Fatal(err)
				}
				return r.Value
			}
			counts := make(map[string]string)
			for step := -1; step <= 2; step++ {
				for _, field := range []string{"b", "s", "j"} {
					key := fmt.Sprintf("%d.%s", step, field)
					counts[key] = take(key, "i32")
					if step == 1 && counts[key] != counts["-1."+field] {
						t.Fatal("DestroyGroup did not restore world counts")
					}
					if step == 2 && counts[key] != counts["0."+field] {
						t.Fatal("recycling changed world counts")
					}
				}
				if step == -1 {
					continue
				}
				spawned, cleared := "1", "0"
				if step == 1 {
					spawned, cleared = "0", "1"
				}
				if take(fmt.Sprintf("%d.spawned", step), "bool") != spawned || take(fmt.Sprintf("%d.cleared", step), "bool") != cleared {
					t.Fatal("group lifecycle flags")
				}
				if take(fmt.Sprintf("%d.column", step), "i32") != "00000000" {
					t.Fatal("recycling index did not wrap")
				}
				if step > 0 {
					if take(fmt.Sprintf("%d.oldBody", step), "bool") != "0" || take(fmt.Sprintf("%d.oldJoint", step), "bool") != "0" {
						t.Fatal("destroyed handles remain valid")
					}
				}
				if step == 1 && take("2.v", "f32") != "00000000" {
					t.Fatal("StepRain return")
				}
			}
			for i, velocity := range []string{"c0a00000,00000000", "40a00000,00000000", "40a00000,00000000", "c0a00000,00000000"} {
				take(fmt.Sprintf("%d.q", i), "f32")
				take(fmt.Sprintf("%d.position", i), "pos")
				if take(fmt.Sprintf("%d.velocity", i), "vec2") != velocity {
					t.Fatal("strict paddle boundary reversal")
				}
				take(fmt.Sprintf("%d.n", i), "i32")
				take(fmt.Sprintf("%d.l", i), "i32")
			}
			if next != len(lines) {
				t.Fatal("extra helper records")
			}
			t.Logf("%d bytes; group destruction/recycling and four strict paddle boundaries", len(data))
		})
	}
}
