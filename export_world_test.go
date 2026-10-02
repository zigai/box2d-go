// SPDX-FileCopyrightText: 2023 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"sync"
	"testing"
)

func BTWorldCleanup(t testing.TB, id World) {
	t.Helper()
	t.Cleanup(func() {
		if id.IsValid() {
			id.Destroy()
		}
	})
}

func BTReplayCleanup(t testing.TB, player *Replay) func() {
	t.Helper()
	destroy := sync.OnceFunc(func() {
		if player != nil {
			player.Destroy()
		}
	})
	t.Cleanup(destroy)
	return destroy
}

const WorldStaticSet = staticSet
const WorldAwakeSet = awakeSet

type WorldWorld = world

func WorldGetWorldFromID(id World) *WorldWorld { return getWorldFromId(id) }
func WorldDynamicTree(id World) *DynamicTree {
	return &getWorldFromId(id).broadPhase.trees[DynamicBody]
}
func WorldJointSetIndex(w *WorldWorld, id Joint) int { return getJointFullId(w, id).setIndex }
func WorldCountEnlargedNodes(tree *DynamicTree) int {
	count := 0
	for i := int32(0); i < tree.nodeEnd; i++ {
		if isNodeMoved(&tree.nodes[i]) {
			count++
		}
	}
	return count
}

func WorldShapeIndex(s Shape) int32 { return s.index1 }

type WorldBodyTransform struct {
	Index     int32
	Transform WorldTransform
}

func WorldBodyTransforms(id World) []WorldBodyTransform {
	w := getWorldFromId(id)
	var transforms []WorldBodyTransform
	for i := range w.bodies {
		body := &w.bodies[i]
		if body.id != i {
			continue
		}
		transforms = append(transforms, WorldBodyTransform{int32(i), getBodySim(w, body).transform})
	}
	return transforms
}
