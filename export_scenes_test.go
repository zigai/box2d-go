// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type SceneBodyTransformForTest struct {
	Index     int
	Transform WorldTransform
}

func SceneBodyTransformsForTest(worldID World) []SceneBodyTransformForTest {
	world := getWorldFromId(worldID)
	var transforms []SceneBodyTransformForTest
	for index := range world.bodies {
		body := &world.bodies[index]
		if body.setIndex == nullIndex {
			continue
		}
		transforms = append(transforms, SceneBodyTransformForTest{Index: index, Transform: getBodyTransformQuick(world, body)})
	}
	return transforms
}
