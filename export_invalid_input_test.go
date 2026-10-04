// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

func InputSaveInputHooks() func() {
	oldAssert, oldLog := assertHandler, logHandler
	return func() {
		assertHandler, logHandler = oldAssert, oldLog
	}
}

func InputBodyTransforms(worldID World) []WorldTransform {
	w := getWorldFromId(worldID)
	transforms := make([]WorldTransform, 0, len(w.bodies))
	for i := range w.bodies {
		if w.bodies[i].id == nullIndex {
			continue
		}
		transforms = append(transforms, makeBodyId(w, i).Transform())
	}
	return transforms
}

// InputBodies preserves creation-slot order for the original upstream scene traces.
func InputBodies(worldID World) []Body {
	w := getWorldFromId(worldID)
	bodies := make([]Body, 0, len(w.bodies))
	for i := range w.bodies {
		if w.bodies[i].id != nullIndex {
			bodies = append(bodies, makeBodyId(w, i))
		}
	}
	return bodies
}

// The public API has no contact-tuning getter; observe the rejected stored settings.
func InputWorldTuning(worldID World) (float32, float32, float32) {
	w := getWorldFromId(worldID)
	return w.contactHertz, w.contactDampingRatio, w.contactSpeed
}

func InputChainCount(worldID World) int {
	return getIDCount(&getWorldFromId(worldID).chainIDPool)
}

func InputAssertsEnabled() bool { return assertsEnabled }
