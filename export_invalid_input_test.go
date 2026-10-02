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
