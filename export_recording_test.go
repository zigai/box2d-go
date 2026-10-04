// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import "encoding/binary"

var RecordingGetWorldFromID = getWorldFromId
var RecordingHashWorldStateDeep = hashWorldStateDeep

type RecordingWorld = world

type RecordingBodyTransform struct {
	Index     int
	Transform WorldTransform
}

func RecordingBodyTransforms(id World) []RecordingBodyTransform {
	w := getWorldFromId(id)
	transforms := make([]RecordingBodyTransform, 0, len(w.bodies))
	for i := range w.bodies {
		if w.bodies[i].id == i {
			transforms = append(transforms, RecordingBodyTransform{Index: i, Transform: getBodyTransform(w, i)})
		}
	}
	return transforms
}

// RecordingQueryResults exposes results absent from the public inspector, not an
// alternate query implementation.
func RecordingQueryResults(p *Replay, index int) (bool, float32, WorldCastOutput) {
	q := p.frameQueries[index]
	return q.boolResult, q.castFraction, q.castOut
}

func RecordingQueryPlane(p *Replay, query, hit int) PlaneResult {
	return p.frameHits[p.frameQueries[query].hitStart+hit].plane
}

func RecordingKeyframeCount(p *Replay) int { return p.keyframeCount }

// Each implementation locates records with its own native reader. No C offsets
// or native snapshots are imported into the Go recording format.
func RecordingMismatch(source []byte, kind string) []byte {
	data := append([]byte(nil), source...)
	cursor := 32 + int(binary.LittleEndian.Uint64(data[24:]))
	frame, changed := 0, 0
	for cursor <= len(data)-4 {
		op := data[cursor]
		size := int(data[cursor+1]) | int(data[cursor+2])<<8 | int(data[cursor+3])<<16
		start := cursor + 4
		if size > len(data)-start {
			panic("invalid test recording framing")
		}
		if op == 0x80 {
			frame++
		}
		if frame == 7 || frame == 12 {
			rdr := recReader{data: data, size: start + size, cursor: start, ok: true}
			if kind == "state" && op == 0xf1 {
				_ = recRWORLDID(&rdr)
				at := rdr.cursor
				hash := recRU64(&rdr)
				if !rdr.ok {
					panic("invalid test hash")
				}
				binary.LittleEndian.PutUint64(data[at:], hash^1)
				changed++
			}
			if kind == "query" && op == 0xe0 {
				_ = recRWORLDID(&rdr)
				_ = recRPOSITION(&rdr)
				_ = recRAABB(&rdr)
				_ = recRQUERYFILTER(&rdr)
				n := recRU32(&rdr)
				if n > 0 {
					at := rdr.cursor
					id := recRSHAPEID(&rdr)
					if !rdr.ok {
						panic("invalid test hit")
					}
					id.generation ^= 1
					binary.LittleEndian.PutUint64(data[at:], id.Store())
					changed++
				}
			}
		}
		cursor = start + size
	}
	if changed != 2 {
		panic("mismatch probe did not alter exactly two relevant records")
	}
	return data
}

func RecordingMalformed(source []byte, name string) []byte {
	data := append([]byte(nil), source...)
	end := 32 + int(binary.LittleEndian.Uint64(data[24:]))
	switch name {
	case "empty":
		return nil
	case "short":
		return data[:31]
	case "magic":
		data[0] ^= 1
	case "major":
		data[4] ^= 1
	case "minor":
		data[6] ^= 1
	case "endian":
		data[18] = 1
	case "snapshotZero":
		binary.LittleEndian.PutUint64(data[24:], 0)
	case "snapshotOversize":
		binary.LittleEndian.PutUint64(data[24:], uint64(len(data)))
	case "incompleteFrame":
		return append(data[:end], 0xf1, 12)
	case "truncatedHash":
		return append(data[:end], 0xf1, 12, 0, 0, 0)
	case "unknown":
		return append(data[:end], 0xff, 1, 0, 0, 0x55)
	case "unknownOverflow":
		return append(data[:end], 0xff, 2, 0, 0, 0x55)
	default:
		panic("unknown malformed test case")
	}
	return data
}

func RecordingBodies(id World) []Body {
	w := getWorldFromId(id)
	var ids []Body
	for i := range w.bodies {
		if w.bodies[i].id == i {
			ids = append(ids, makeBodyId(w, i))
		}
	}
	return ids
}

func RecordingShapes(id World) []Shape {
	w := getWorldFromId(id)
	var ids []Shape
	for i := range w.shapes {
		if w.shapes[i].id == i {
			ids = append(ids, Shape{index1: int32(i + 1), world0: uint16(id.index1 - 1), generation: w.shapes[i].generation})
		}
	}
	return ids
}

func RecordingJoints(id World) []Joint {
	w := getWorldFromId(id)
	var ids []Joint
	for i := range w.joints {
		if w.joints[i].jointId == i {
			ids = append(ids, Joint{index1: int32(i + 1), world0: uint16(id.index1 - 1), generation: w.joints[i].generation})
		}
	}
	return ids
}

func RecordingContactTuning(id World) (float32, float32, float32) {
	w := getWorldFromId(id)
	return w.contactHertz, w.contactDampingRatio, w.contactSpeed
}
