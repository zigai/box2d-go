package box2d

import (
	"encoding/binary"
	"fmt"
	"github.com/zigai/box2d/internal/oracle"
	"strings"
	"testing"
)

const BTDynamicsSchema = "bodyCount,index,generation,world,position,rotation,velocity,angularVelocity,awake,enabled,type,normal,rollingImpulse,pointCount,anchorA,anchorB,separation,normalImpulse,tangentImpulse,totalNormalImpulse,normalVelocity,baseSeparation,id,persisted,contactCount,fellAsleep,moveCount,beginCount,endCount,point,approachSpeed,hitCount,sensorBeginCount,sensorEndCount,valid,canonicalState"

var btDynamicsFields = strings.Split(BTDynamicsSchema, ",")

func BTDynamicsPack(rows [][36]uint64) []byte {
	var data []byte
	for i := 0; i < len(rows); {
		end := i + 1
		for end < len(rows) && rows[end] == rows[i] {
			end++
		}
		data = binary.LittleEndian.AppendUint32(data, uint32(end-i))
		for _, v := range rows[i] {
			data = binary.LittleEndian.AppendUint64(data, v)
		}
		i = end
	}
	return data
}
func BTDynamicsUnpack(t testing.TB, data []byte) [][36]uint64 {
	t.Helper()
	var rows [][36]uint64
	for len(data) != 0 {
		if len(data) < 292 {
			t.Fatal("incomplete dynamics RLE row")
		}
		count := binary.LittleEndian.Uint32(data[:4])
		data = data[4:]
		if count == 0 || count > 1048576 || len(rows)+int(count) > 1048576 {
			t.Fatal("invalid dynamics RLE run count")
		}
		var row [36]uint64
		for i := range row {
			row[i] = binary.LittleEndian.Uint64(data[:8])
			data = data[8:]
		}
		for range count {
			rows = append(rows, row)
		}
	}
	return rows
}
func BTDynamicsCompare(t testing.TB, label string, step int, got [36]uint64, want [][36]uint64) {
	t.Helper()
	if step > len(want) {
		t.Fatalf("%s: unexpected observation %d", label, step)
	}
	for i, v := range got {
		if v != want[step-1][i] {
			t.Fatalf("%s: first divergent observation %d field %s: C=%016x Go=%016x", label, step, btDynamicsFields[i], want[step-1][i], v)
		}
	}
}
func BTDynamicsEnd(r *oracle.Reader, rows [][36]uint64) {
	r.Bytes("dynamics/schema", []byte(BTDynamicsSchema))
	r.Bytes("dynamics/rle-u32-36u64-le", BTDynamicsPack(rows))
}
func btDynamicsShort(r *dynamicsOracle, w World, prefix string, direct bool) {
	row := BTObserveDynamics(r.Reader, w, prefix, direct)
	r.observation++
	BTDynamicsCompare(r.t, prefix, r.observation, row, r.expected)
	r.U64(prefix+"canonicalState", row[35])
	data := make([]byte, 0, 36*8)
	for _, v := range row {
		data = binary.LittleEndian.AppendUint64(data, v)
	}
	r.Bytes(fmt.Sprintf("%sdigests", prefix), data)
}
