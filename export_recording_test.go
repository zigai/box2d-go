// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

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
