package box2d

import (
	"fmt"
	"github.com/zigai/box2d/internal/oracle"
)

// BTObserveDynamics observes post-step public state, never pre-solve scratch impulses.
// Handle generations are retained; world membership is relative to the tested world.
func BTObserveDynamics(r *oracle.Reader, id World, prefix string, direct bool) [36]uint64 {
	var h [36]uint64
	for i := range h {
		h[i] = oracle.HashInit
	}
	w := getWorldFromId(id)

	h[0] = oracle.HashU32(h[0], uint32(int32(id.Counters().BodyCount)))
	if direct {
		r.I32(prefix+"bodyCount", int32(id.Counters().BodyCount))
	}
	for i := range w.bodies {
		if w.bodies[i].id != i {
			continue
		}
		b := makeBodyId(w, i)
		key := ""
		if direct {
			key = fmt.Sprintf("%sbody/%d/", prefix, i)
		}

		h[1] = oracle.HashU32(h[1], uint32(b.index1))
		if direct {
			r.I32(key+"index", b.index1)
		}

		h[2] = oracle.HashU32(h[2], uint32(int32(b.generation)))
		if direct {
			r.I32(key+"generation", int32(b.generation))
		}

		h[3] = btDynamicsBool(h[3], b.world0 == w.worldID)
		if direct {
			r.Bool(key+"world", b.world0 == w.worldID)
		}
		p, q, v := b.Position(), b.Rotation(), b.LinearVelocity()

		h[4] = oracle.HashPos(h[4], p.X, p.Y)
		if direct {
			r.Pos(key+"position", p.X, p.Y)
		}

		h[5] = oracle.HashRot(h[5], q.C, q.S)
		if direct {
			r.Rot(key+"rotation", q.C, q.S)
		}

		h[6] = oracle.HashRot(h[6], v.X, v.Y)
		if direct {
			r.Vec2(key+"velocity", v.X, v.Y)
		}

		h[7] = oracle.HashF32(h[7], b.AngularVelocity())
		if direct {
			r.F32(key+"angularVelocity", b.AngularVelocity())
		}

		h[8] = btDynamicsBool(h[8], b.IsAwake())
		if direct {
			r.Bool(key+"awake", b.IsAwake())
		}

		h[9] = btDynamicsBool(h[9], b.IsEnabled())
		if direct {
			r.Bool(key+"enabled", b.IsEnabled())
		}

		h[10] = oracle.HashU32(h[10], uint32(int32(b.Type())))
		if direct {
			r.I32(key+"type", int32(b.Type()))
		}
		count := 0
		for c := range b.ContactData() {
			ck := ""
			if direct {
				ck = fmt.Sprintf("%scontact/%d/", key, count)
			}
			btDynamicsContact(r, ck, c.ContactID, c.ShapeIDA, c.ShapeIDB, w.worldID, &h, direct)
			m := c.Manifold

			h[11] = oracle.HashRot(h[11], m.Normal.X, m.Normal.Y)
			if direct {
				r.Vec2(ck+"normal", m.Normal.X, m.Normal.Y)
			}

			h[12] = oracle.HashF32(h[12], m.RollingImpulse)
			if direct {
				r.F32(ck+"rollingImpulse", m.RollingImpulse)
			}

			h[13] = oracle.HashU32(h[13], uint32(int32(m.PointCount)))
			if direct {
				r.I32(ck+"pointCount", int32(m.PointCount))
			}
			for j := range m.PointCount {
				pk := ""
				if direct {
					pk = fmt.Sprintf("%spoint/%d/", ck, j)
				}
				p := m.Points[j]

				h[14] = oracle.HashRot(h[14], p.AnchorA.X, p.AnchorA.Y)
				if direct {
					r.Vec2(pk+"anchorA", p.AnchorA.X, p.AnchorA.Y)
				}

				h[15] = oracle.HashRot(h[15], p.AnchorB.X, p.AnchorB.Y)
				if direct {
					r.Vec2(pk+"anchorB", p.AnchorB.X, p.AnchorB.Y)
				}

				h[16] = oracle.HashF32(h[16], p.Separation)
				if direct {
					r.F32(pk+"separation", p.Separation)
				}

				h[17] = oracle.HashF32(h[17], p.NormalImpulse)
				if direct {
					r.F32(pk+"normalImpulse", p.NormalImpulse)
				}

				h[18] = oracle.HashF32(h[18], p.TangentImpulse)
				if direct {
					r.F32(pk+"tangentImpulse", p.TangentImpulse)
				}

				h[19] = oracle.HashF32(h[19], p.TotalNormalImpulse)
				if direct {
					r.F32(pk+"totalNormalImpulse", p.TotalNormalImpulse)
				}

				h[20] = oracle.HashF32(h[20], p.NormalVelocity)
				if direct {
					r.F32(pk+"normalVelocity", p.NormalVelocity)
				}

				h[21] = oracle.HashF32(h[21], p.BaseSeparation)
				if direct {
					r.F32(pk+"baseSeparation", p.BaseSeparation)
				}

				h[22] = oracle.HashU32(h[22], uint32(int32(p.ID)))
				if direct {
					r.I32(pk+"id", int32(p.ID))
				}

				h[23] = btDynamicsBool(h[23], p.Persisted)
				if direct {
					r.Bool(pk+"persisted", p.Persisted)
				}
			}
			count++
		}

		h[24] = oracle.HashU32(h[24], uint32(int32(count)))
		if direct {
			r.I32(key+"contactCount", int32(count))
		}
	}
	count := 0
	for e := range id.BodyMoveEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%smove/%d/", prefix, count)
		}

		h[1] = oracle.HashU32(h[1], uint32(e.BodyID.index1))
		if direct {
			r.I32(key+"index", e.BodyID.index1)
		}

		h[2] = oracle.HashU32(h[2], uint32(int32(e.BodyID.generation)))
		if direct {
			r.I32(key+"generation", int32(e.BodyID.generation))
		}

		h[3] = btDynamicsBool(h[3], e.BodyID.world0 == w.worldID)
		if direct {
			r.Bool(key+"world", e.BodyID.world0 == w.worldID)
		}

		h[4] = oracle.HashPos(h[4], e.Transform.P.X, e.Transform.P.Y)
		if direct {
			r.Pos(key+"position", e.Transform.P.X, e.Transform.P.Y)
		}

		h[5] = oracle.HashRot(h[5], e.Transform.Q.C, e.Transform.Q.S)
		if direct {
			r.Rot(key+"rotation", e.Transform.Q.C, e.Transform.Q.S)
		}

		h[25] = btDynamicsBool(h[25], e.FellAsleep)
		if direct {
			r.Bool(key+"fellAsleep", e.FellAsleep)
		}
		count++
	}

	h[26] = oracle.HashU32(h[26], uint32(int32(count)))
	if direct {
		r.I32(prefix+"moveCount", int32(count))
	}
	count = 0
	for e := range id.ContactBeginEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%sbegin/%d/", prefix, count)
		}
		btDynamicsContact(r, key, e.ContactID, e.ShapeIDA, e.ShapeIDB, w.worldID, &h, direct)
		count++
	}

	h[27] = oracle.HashU32(h[27], uint32(int32(count)))
	if direct {
		r.I32(prefix+"beginCount", int32(count))
	}
	count = 0
	for e := range id.ContactEndEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%send/%d/", prefix, count)
		}
		btDynamicsContact(r, key, e.ContactID, e.ShapeIDA, e.ShapeIDB, w.worldID, &h, direct)
		count++
	}

	h[28] = oracle.HashU32(h[28], uint32(int32(count)))
	if direct {
		r.I32(prefix+"endCount", int32(count))
	}
	count = 0
	for e := range id.ContactHitEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%shit/%d/", prefix, count)
		}
		btDynamicsContact(r, key, e.ContactID, e.ShapeIDA, e.ShapeIDB, w.worldID, &h, direct)

		h[29] = oracle.HashPos(h[29], e.Point.X, e.Point.Y)
		if direct {
			r.Pos(key+"point", e.Point.X, e.Point.Y)
		}

		h[11] = oracle.HashRot(h[11], e.Normal.X, e.Normal.Y)
		if direct {
			r.Vec2(key+"normal", e.Normal.X, e.Normal.Y)
		}

		h[30] = oracle.HashF32(h[30], e.ApproachSpeed)
		if direct {
			r.F32(key+"approachSpeed", e.ApproachSpeed)
		}
		count++
	}

	h[31] = oracle.HashU32(h[31], uint32(int32(count)))
	if direct {
		r.I32(prefix+"hitCount", int32(count))
	}
	count = 0
	for e := range id.SensorBeginEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%ssensorBegin/%d/", prefix, count)
		}
		btDynamicsShape(r, key+"sensor/", e.SensorShapeID, w.worldID, &h, direct)
		btDynamicsShape(r, key+"visitor/", e.VisitorShapeID, w.worldID, &h, direct)
		count++
	}

	h[32] = oracle.HashU32(h[32], uint32(int32(count)))
	if direct {
		r.I32(prefix+"sensorBeginCount", int32(count))
	}
	count = 0
	for e := range id.SensorEndEvents() {
		key := ""
		if direct {
			key = fmt.Sprintf("%ssensorEnd/%d/", prefix, count)
		}
		btDynamicsShape(r, key+"sensor/", e.SensorShapeID, w.worldID, &h, direct)
		btDynamicsShape(r, key+"visitor/", e.VisitorShapeID, w.worldID, &h, direct)
		count++
	}

	h[33] = oracle.HashU32(h[33], uint32(int32(count)))
	if direct {
		r.I32(prefix+"sensorEndCount", int32(count))
	}
	h[35] = id.StateHash()
	return h
}

func btDynamicsShape(r *oracle.Reader, key string, s Shape, worldID uint16, h *[36]uint64, direct bool) {

	h[1] = oracle.HashU32(h[1], uint32(s.index1))
	if direct {
		r.I32(key+"index", s.index1)
	}

	h[2] = oracle.HashU32(h[2], uint32(int32(s.generation)))
	if direct {
		r.I32(key+"generation", int32(s.generation))
	}

	h[3] = btDynamicsBool(h[3], s.world0 == uint16(worldID))
	if direct {
		r.Bool(key+"world", s.world0 == uint16(worldID))
	}

	h[34] = btDynamicsBool(h[34], s.IsValid())
	if direct {
		r.Bool(key+"valid", s.IsValid())
	}
}
func btDynamicsContact(r *oracle.Reader, key string, c Contact, a, b Shape, worldID uint16, h *[36]uint64, direct bool) {

	h[1] = oracle.HashU32(h[1], uint32(c.index1))
	if direct {
		r.I32(key+"index", c.index1)
	}

	h[2] = oracle.HashU64(h[2], uint64(c.generation))
	if direct {
		r.U64(key+"generation", uint64(c.generation))
	}

	h[3] = btDynamicsBool(h[3], c.world0 == uint16(worldID))
	if direct {
		r.Bool(key+"world", c.world0 == uint16(worldID))
	}

	h[34] = btDynamicsBool(h[34], c.IsValid())
	if direct {
		r.Bool(key+"valid", c.IsValid())
	}
	btDynamicsShape(r, key+"shapeA/", a, worldID, h, direct)
	btDynamicsShape(r, key+"shapeB/", b, worldID, h, direct)
}

func btDynamicsBool(h uint64, value bool) uint64 {
	if value {
		return oracle.HashU32(h, 1)
	}
	return oracle.HashU32(h, 0)
}
