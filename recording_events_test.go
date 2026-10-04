// SPDX-License-Identifier: MIT
package box2d_test

import (
	"github.com/zigai/box2d"
	"github.com/zigai/box2d/internal/oracle"
)

func recordingEventValues(w box2d.World) [11]uint64 {
	var row [11]uint64
	hash := uint64(oracle.HashInit)
	id := func(packed uint64) {
		hash = oracle.HashU32(hash, uint32(packed>>32))
		hash = oracle.HashU32(hash, uint32(uint16(packed)))
		owner := uint32(0)
		if uint16(packed>>16) == uint16(w.Store()>>16)-1 {
			owner = 1
		}
		hash = oracle.HashU32(hash, owner)
	}
	contact := func(c box2d.Contact) {
		packed := c.Store()
		hash = oracle.HashU32(hash, packed[0])
		hash = oracle.HashU32(hash, packed[2])
		owner := uint32(0)
		if uint16(packed[1]) == uint16(w.Store()>>16)-1 {
			owner = 1
		}
		hash = oracle.HashU32(hash, owner)
	}
	count := 0
	for e := range w.BodyMoveEvents() {
		count++
		id(e.BodyID.Store())
		hash = oracle.HashTransform(hash, oracle.Real(e.Transform.P.X), oracle.Real(e.Transform.P.Y), e.Transform.Q.C, e.Transform.Q.S)
		asleep := uint32(0)
		if e.FellAsleep {
			asleep = 1
		}
		hash = oracle.HashU32(hash, asleep)
	}
	row[0] = uint64(count)
	row[1] = uint64(hash)
	hash = oracle.HashInit
	begin, end := 0, 0
	for e := range w.SensorBeginEvents() {
		begin++
		id(e.SensorShapeID.Store())
		id(e.VisitorShapeID.Store())
	}
	for e := range w.SensorEndEvents() {
		end++
		id(e.SensorShapeID.Store())
		id(e.VisitorShapeID.Store())
	}
	row[2] = uint64(begin)
	row[3] = uint64(end)
	row[4] = uint64(hash)
	hash = oracle.HashInit
	begin, end, hits := 0, 0, 0
	for e := range w.ContactBeginEvents() {
		begin++
		id(e.ShapeIDA.Store())
		id(e.ShapeIDB.Store())
		contact(e.ContactID)
	}
	for e := range w.ContactEndEvents() {
		end++
		id(e.ShapeIDA.Store())
		id(e.ShapeIDB.Store())
		contact(e.ContactID)
	}
	for e := range w.ContactHitEvents() {
		hits++
		id(e.ShapeIDA.Store())
		id(e.ShapeIDB.Store())
		contact(e.ContactID)
		hash = oracle.HashPos(hash, oracle.Real(e.Point.X), oracle.Real(e.Point.Y))
		hash = oracle.HashF32(hash, e.Normal.X)
		hash = oracle.HashF32(hash, e.Normal.Y)
		hash = oracle.HashF32(hash, e.ApproachSpeed)
	}
	row[5] = uint64(begin)
	row[6] = uint64(end)
	row[7] = uint64(hits)
	row[8] = uint64(hash)
	hash = oracle.HashInit
	count = 0
	for e := range w.JointEvents() {
		count++
		id(e.JointID.Store())
	}
	row[9] = uint64(count)
	row[10] = uint64(hash)
	return row
}

func recordingEvents(r *oracle.Reader, w box2d.World, prefix string) {
	row := recordingEventValues(w)
	r.I32(prefix+"/events/moveCount", int32(row[0]))
	r.U64(prefix+"/events/moveHash", row[1])
	r.I32(prefix+"/events/sensorBeginCount", int32(row[2]))
	r.I32(prefix+"/events/sensorEndCount", int32(row[3]))
	r.U64(prefix+"/events/sensorHash", row[4])
	r.I32(prefix+"/events/contactBeginCount", int32(row[5]))
	r.I32(prefix+"/events/contactEndCount", int32(row[6]))
	r.I32(prefix+"/events/hitCount", int32(row[7]))
	r.U64(prefix+"/events/contactHash", row[8])
	r.I32(prefix+"/events/jointCount", int32(row[9]))
	r.U64(prefix+"/events/jointHash", row[10])
}
