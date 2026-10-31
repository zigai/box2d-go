// SPDX-License-Identifier: MIT

package box2d_test

import (
	"github.com/zigai/box2d-go/box2d"
	"github.com/zigai/box2d-go/internal/oracle"
)

func recordingEventValues(w *box2d.World) [11]uint64 {
	var row [11]uint64

	hash := oracle.HashInit
	id := func(packed uint64, world *box2d.World) {
		hash = oracle.HashU32(hash, uint32(packed>>32))
		hash = oracle.HashU32(hash, uint32(uint16(packed)))

		owner := uint32(0)
		if world == w {
			owner = 1
		}

		hash = oracle.HashU32(hash, owner)
	}
	contact := func(c box2d.Contact) {
		packed := c.ID()
		hash = oracle.HashU32(hash, uint32(packed>>32))
		hash = oracle.HashU32(hash, uint32(packed))

		owner := uint32(0)
		if c.World() == w {
			owner = 1
		}

		hash = oracle.HashU32(hash, owner)
	}

	count := 0
	for e := range w.BodyMoveEvents() {
		count++

		id(e.Body.ID(), e.Body.World())
		hash = oracle.HashTransform(
			hash,
			oracle.Real(e.Transform.P.X),
			oracle.Real(e.Transform.P.Y),
			e.Transform.Q.C,
			e.Transform.Q.S,
		)

		asleep := uint32(0)
		if e.FellAsleep {
			asleep = 1
		}

		hash = oracle.HashU32(hash, asleep)
	}

	row[0] = uint64(count)
	row[1] = hash
	hash = oracle.HashInit

	begin, end := 0, 0
	for e := range w.SensorBeginEvents() {
		begin++

		id(e.SensorShape.ID(), e.SensorShape.World())
		id(e.VisitorShape.ID(), e.VisitorShape.World())
	}

	for e := range w.SensorEndEvents() {
		end++

		id(e.SensorShape.ID(), e.SensorShape.World())
		id(e.VisitorShape.ID(), e.VisitorShape.World())
	}

	row[2] = uint64(begin)
	row[3] = uint64(end)
	row[4] = hash
	hash = oracle.HashInit

	begin, end, hits := 0, 0, 0
	for e := range w.ContactBeginEvents() {
		begin++

		id(e.ShapeA.ID(), e.ShapeA.World())
		id(e.ShapeB.ID(), e.ShapeB.World())
		contact(e.Contact)
	}

	for e := range w.ContactEndEvents() {
		end++

		id(e.ShapeA.ID(), e.ShapeA.World())
		id(e.ShapeB.ID(), e.ShapeB.World())
		contact(e.Contact)
	}

	for e := range w.ContactHitEvents() {
		hits++

		id(e.ShapeA.ID(), e.ShapeA.World())
		id(e.ShapeB.ID(), e.ShapeB.World())
		contact(e.Contact)
		hash = oracle.HashPos(hash, oracle.Real(e.Point.X), oracle.Real(e.Point.Y))
		hash = oracle.HashF32(hash, e.Normal.X)
		hash = oracle.HashF32(hash, e.Normal.Y)
		hash = oracle.HashF32(hash, e.ApproachSpeed)
	}

	row[5] = uint64(begin)
	row[6] = uint64(end)
	row[7] = uint64(hits)
	row[8] = hash
	hash = oracle.HashInit

	count = 0
	for e := range w.JointEvents() {
		count++

		id(e.Joint.ID(), e.Joint.World())
	}

	row[9] = uint64(count)
	row[10] = hash

	return row
}

func recordingEvents(r *oracle.Reader, w *box2d.World, prefix string) {
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
