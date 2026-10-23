// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

func InputBodyTransforms(worldID *World) []WorldTransform {
	w := worldID

	transforms := make([]WorldTransform, 0, len(w.bodies))
	for i := range w.bodies {
		if w.bodies[i].id == nullIndex {
			continue
		}

		transforms = append(transforms, makeBodyID(w, i).Transform())
	}

	return transforms
}

// InputBodies preserves creation-slot order for the original upstream scene traces.
func InputBodies(worldID *World) []Body {
	w := worldID

	bodies := make([]Body, 0, len(w.bodies))
	for i := range w.bodies {
		if w.bodies[i].id != nullIndex {
			bodies = append(bodies, makeBodyID(w, i))
		}
	}

	return bodies
}

// The public API has no contact-tuning getter; observe the rejected stored settings.
func InputWorldTuning(worldID *World) (float32, float32, float32) {
	w := worldID
	return w.contactHertz, w.contactDampingRatio, w.contactSpeed
}

func InputChainCount(worldID *World) int {
	return worldID.chainIDPool.Count()
}
