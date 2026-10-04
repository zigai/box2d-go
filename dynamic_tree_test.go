// SPDX-FileCopyrightText: 2025 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"github.com/zigai/box2d/internal/oracle"

	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

const treeGridCount = 20
const treeProxyCount = 200

func TestDynamicTree(t *testing.T) {

	t.Run("TreeCreateDestroy", func(t *testing.T) {

		s := treeOpen(t, "TreeCreateDestroy", 12345)
		s.r.U32("rng/start", s.randomState)
		a := AABB{
			LowerBound: Vec2{-1.0, -1.0},
			UpperBound: Vec2{2.0, 2.0},
		}
		tree := NewDynamicTree(16)
		s.tree("step/22", &tree)
		createTreeProxyInternal(&tree, a, 1, 0, false)
		s.tree("step/23", &tree)
		s.check(25, uint64(tree.nodeEnd), tree.nodeEnd == 2)
		if !(tree.nodeEnd == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/25")
		s.tree("check-state/25", &tree)
		s.check(26, uint64(tree.proxyCount), tree.proxyCount == 1)
		if !(tree.proxyCount == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/26")
		s.tree("check-state/26", &tree)
		tree = DynamicTree{}
		s.tree("step/28", &tree)
		s.check(30, uint64(tree.nodeEnd), tree.nodeEnd == 0)
		if !(tree.nodeEnd == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/30")
		s.tree("check-state/30", &tree)
		s.check(31, uint64(tree.proxyCount), tree.proxyCount == 0)
		if !(tree.proxyCount == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/31")
		s.tree("check-state/31", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeRayCastTest", func(t *testing.T) {

		s := treeOpen(t, "TreeRayCastTest", 12345)
		s.r.U32("rng/start", s.randomState)
		// Test AABB centered at origin with bounds [-1, -1] to [1, 1]
		a := AABB{LowerBound: Vec2{-1.0, -1.0}, UpperBound: Vec2{1.0, 1.0}}
		tree := NewDynamicTree(16)
		s.tree("step/50", &tree)
		proxyId := createTreeProxyInternal(&tree, a, 1, 0, false)
		s.tree("step/51", &tree)
		input := RayCastInput{}
		input.MaxFraction = 1.0
		// Test 1: Ray hits AABB from left side
		{
			p1 := Vec2{-3.0, 0.0}
			p2 := Vec2{3.0, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/65", &tree)
			s.check(67, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/67")
			s.tree("check-state/67", &tree)
		}
		// Test 2: Ray hits AABB from right side
		{
			p1 := Vec2{3.0, 0.0}
			p2 := Vec2{-3.0, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/78", &tree)
			s.check(80, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/80")
			s.tree("check-state/80", &tree)
		}
		// Test 3: Ray hits AABB from bottom
		{
			p1 := Vec2{0.0, -3.0}
			p2 := Vec2{0.0, 3.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/91", &tree)
			s.check(93, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/93")
			s.tree("check-state/93", &tree)
		}
		// Test 4: Ray hits AABB from top
		{
			p1 := Vec2{0.0, 3.0}
			p2 := Vec2{0.0, -3.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/104", &tree)
			s.check(106, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/106")
			s.tree("check-state/106", &tree)
		}
		// Test 5: Ray misses AABB completely (parallel to x-axis)
		{
			p1 := Vec2{-3.0, 2.0}
			p2 := Vec2{3.0, 2.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/117", &tree)
			s.check(119, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/119")
			s.tree("check-state/119", &tree)
		}
		// Test 6: Ray misses AABB completely (parallel to y-axis)
		{
			p1 := Vec2{2.0, -3.0}
			p2 := Vec2{2.0, 3.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/130", &tree)
			s.check(132, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/132")
			s.tree("check-state/132", &tree)
		}
		// Test 7: Ray starts inside AABB
		{
			p1 := Vec2{0.0, 0.0}
			p2 := Vec2{2.0, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/143", &tree)
			s.check(145, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/145")
			s.tree("check-state/145", &tree)
		}
		// Test 8: Ray hits corner of AABB (diagonal ray)
		{
			p1 := Vec2{-2.0, -2.0}
			p2 := Vec2{2.0, 2.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/156", &tree)
			s.check(158, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/158")
			s.tree("check-state/158", &tree)
		}
		// Test 9: Ray parallel to AABB edge but outside
		{
			p1 := Vec2{-2.0, 1.5}
			p2 := Vec2{2.0, 1.5}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/169", &tree)
			s.check(171, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/171")
			s.tree("check-state/171", &tree)
		}
		// Test 10: Ray parallel to AABB edge and exactly on boundary
		{
			p1 := Vec2{-2.0, 1.0}
			p2 := Vec2{2.0, 1.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/182", &tree)
			s.check(184, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/184")
			s.tree("check-state/184", &tree)
		}
		// Test 11: Very short ray that doesn't reach AABB
		{
			p1 := Vec2{-3.0, 0.0}
			p2 := Vec2{-2.5, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/195", &tree)
			s.check(197, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/197")
			s.tree("check-state/197", &tree)
		}
		// Test 12: Zero-length ray (degenerate case)
		{
			p1 := Vec2{0.0, 0.0}
			p2 := Vec2{0.0, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/208", &tree)
			s.check(210, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/210")
			s.tree("check-state/210", &tree)
		}
		// Test 13: Ray hits AABB at exact boundary condition (t = 1.0)
		{
			p1 := Vec2{-2.0, 0.0}
			p2 := Vec2{-1.0, 0.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/221", &tree)
			s.check(223, uint64(proxyHit), proxyHit == proxyId)
			if !(proxyHit == proxyId) {
				t.Fatal("condition false")
			}
			s.flush("check/223")
			s.tree("check-state/223", &tree)
		}
		// An off center box catches an axis aligned ray tested against the wrong axis, which the origin
		// centered box above cannot. Two leaves also give the root a child pair.
		b := AABB{LowerBound: Vec2{10.0, 4.0}, UpperBound: Vec2{11.0, 6.0}}
		proxyIdB := createTreeProxyInternal(&tree, b, 1, 1, false)
		s.tree("step/229", &tree)
		// Test 14: Horizontal ray through the off center box
		{
			p1 := Vec2{0.0, 5.0}
			p2 := Vec2{21.0, 5.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/239", &tree)
			s.check(241, uint64(proxyHit), proxyHit == proxyIdB)
			if !(proxyHit == proxyIdB) {
				t.Fatal("condition false")
			}
			s.flush("check/241")
			s.tree("check-state/241", &tree)
		}
		// Test 15: Vertical ray through the off center box
		{
			p1 := Vec2{10.5, 0.0}
			p2 := Vec2{10.5, 21.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/252", &tree)
			s.check(254, uint64(proxyHit), proxyHit == proxyIdB)
			if !(proxyHit == proxyIdB) {
				t.Fatal("condition false")
			}
			s.flush("check/254")
			s.tree("check-state/254", &tree)
		}
		// Test 16: Horizontal ray passing above the off center box
		{
			p1 := Vec2{0.0, 7.0}
			p2 := Vec2{21.0, 7.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/265", &tree)
			s.check(267, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/267")
			s.tree("check-state/267", &tree)
		}
		// Test 17: Vertical ray passing beside the off center box
		{
			p1 := Vec2{12.0, 0.0}
			p2 := Vec2{12.0, 21.0}
			input.Origin = p1
			input.Translation = p2.Sub(p1)
			proxyHit := -1
			tree.CastRay(&input, 1, func(_ *RayCastInput, proxyID int, _ uint64) float32 { proxyHit = proxyID; return 0 })
			s.tree("step/278", &tree)
			s.check(280, uint64(proxyHit), proxyHit == -1)
			if !(proxyHit == -1) {
				t.Fatal("condition false")
			}
			s.flush("check/280")
			s.tree("check-state/280", &tree)
		}
		tree = DynamicTree{}
		s.tree("step/283", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeMultipleProxiesTest", func(t *testing.T) {

		s := treeOpen(t, "TreeMultipleProxiesTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/308", &tree)
		a1 := AABB{LowerBound: Vec2{-5.0, -1.0}, UpperBound: Vec2{-3.0, 1.0}}
		a2 := AABB{LowerBound: Vec2{-1.0, -1.0}, UpperBound: Vec2{1.0, 1.0}}
		a3 := AABB{LowerBound: Vec2{3.0, -1.0}, UpperBound: Vec2{5.0, 1.0}}
		id1 := createTreeProxyInternal(&tree, a1, 0x1, 42, false)
		s.tree("step/314", &tree)
		id2 := createTreeProxyInternal(&tree, a2, 0x2, 43, false)
		s.tree("step/315", &tree)
		id3 := createTreeProxyInternal(&tree, a3, 0x4, 44, false)
		s.tree("step/316", &tree)
		s.check(318, uint64(tree.ProxyCount()), tree.ProxyCount() == 3)
		if !(tree.ProxyCount() == 3) {
			t.Fatal("condition false")
		}
		s.flush("check/318")
		s.tree("check-state/318", &tree)
		s.check(320, uint64(tree.UserData(id1)), tree.UserData(id1) == 42)
		if !(tree.UserData(id1) == 42) {
			t.Fatal("condition false")
		}
		s.flush("check/320")
		s.tree("check-state/320", &tree)
		s.check(321, uint64(tree.UserData(id2)), tree.UserData(id2) == 43)
		if !(tree.UserData(id2) == 43) {
			t.Fatal("condition false")
		}
		s.flush("check/321")
		s.tree("check-state/321", &tree)
		s.check(322, uint64(tree.UserData(id3)), tree.UserData(id3) == 44)
		if !(tree.UserData(id3) == 44) {
			t.Fatal("condition false")
		}
		s.flush("check/322")
		s.tree("check-state/322", &tree)
		s.check(324, uint64(tree.CategoryBits(id1)), tree.CategoryBits(id1) == 0x1)
		if !(tree.CategoryBits(id1) == 0x1) {
			t.Fatal("condition false")
		}
		s.flush("check/324")
		s.tree("check-state/324", &tree)
		s.check(325, uint64(tree.CategoryBits(id2)), tree.CategoryBits(id2) == 0x2)
		if !(tree.CategoryBits(id2) == 0x2) {
			t.Fatal("condition false")
		}
		s.flush("check/325")
		s.tree("check-state/325", &tree)
		s.check(326, uint64(tree.CategoryBits(id3)), tree.CategoryBits(id3) == 0x4)
		if !(tree.CategoryBits(id3) == 0x4) {
			t.Fatal("condition false")
		}
		s.flush("check/326")
		s.tree("check-state/326", &tree)
		tree = DynamicTree{}
		s.tree("step/328", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeQueryTest", func(t *testing.T) {

		s := treeOpen(t, "TreeQueryTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/334", &tree)
		a1 := AABB{LowerBound: Vec2{-5.0, -1.0}, UpperBound: Vec2{-3.0, 1.0}}
		a2 := AABB{LowerBound: Vec2{-1.0, -1.0}, UpperBound: Vec2{1.0, 1.0}}
		a3 := AABB{LowerBound: Vec2{3.0, -1.0}, UpperBound: Vec2{5.0, 1.0}}
		id1 := createTreeProxyInternal(&tree, a1, 0xFF, 0, false)
		s.tree("step/340", &tree)
		id2 := createTreeProxyInternal(&tree, a2, 0xFF, 0, false)
		s.tree("step/341", &tree)
		id3 := createTreeProxyInternal(&tree, a3, 0xFF, 0, false)
		s.tree("step/342", &tree)
		queryA := AABB{LowerBound: Vec2{-2.0, -2.0}, UpperBound: Vec2{2.0, 2.0}}
		var foundFlags [32]int
		stats := tree.Query(queryA, 0xFFFFFFFF, treeCollectFlags(foundFlags[:]))
		s.tree("step/347", &tree)
		// We expect at least the middle proxy to be visited.
		s.check(350, uint64(foundFlags[id2]), foundFlags[id2] == 1)
		if !(foundFlags[id2] == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/350")
		s.tree("check-state/350", &tree)
		s.check(351, uint64(stats.LeafVisits), stats.LeafVisits >= 1)
		if !(stats.LeafVisits >= 1) {
			t.Fatal("condition false")
		}
		s.flush("check/351")
		s.tree("check-state/351", &tree)
		// Test QueryAll using list collector
		var list [16]int // list[0] holds count, following entries are ids
		allStats := tree.QueryAll(queryA, treeCollectList(list[:]))
		s.tree("step/355", &tree)
		s.check(356, uint64(list[0]), list[0] >= 1)
		if !(list[0] >= 1) {
			t.Fatal("condition false")
		} // at least one proxy should be collected
		s.flush("check/356")
		s.tree("check-state/356", &tree)
		s.check(357, uint64(allStats.LeafVisits), allStats.LeafVisits >= 1)
		if !(allStats.LeafVisits >= 1) {
			t.Fatal("condition false")
		}
		s.flush("check/357")
		s.tree("check-state/357", &tree)
		tree = DynamicTree{}
		s.tree("step/359", &tree)
		_ = id1
		_ = id3
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeMoveAndEnlargeTest", func(t *testing.T) {

		s := treeOpen(t, "TreeMoveAndEnlargeTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/366", &tree)
		a := AABB{LowerBound: Vec2{0.0, 0.0}, UpperBound: Vec2{1.0, 1.0}}
		id := createTreeProxyInternal(&tree, a, 0x1, 100, false)
		s.tree("step/369", &tree)
		// Move proxy to a new place
		moved := AABB{LowerBound: Vec2{10.0, 10.0}, UpperBound: Vec2{11.0, 11.0}}
		tree.moveProxyInternal(id, moved, false)
		s.tree("step/373", &tree)
		got := tree.AABB(id)
		s.check(376, uint64(math.Float32bits(got.LowerBound.X)), got.LowerBound.X == moved.LowerBound.X)
		if !(got.LowerBound.X == moved.LowerBound.X) {
			t.Fatal("condition false")
		}
		s.flush("check/376")
		s.tree("check-state/376", &tree)
		s.check(377, uint64(math.Float32bits(got.LowerBound.Y)), got.LowerBound.Y == moved.LowerBound.Y)
		if !(got.LowerBound.Y == moved.LowerBound.Y) {
			t.Fatal("condition false")
		}
		s.flush("check/377")
		s.tree("check-state/377", &tree)
		s.check(378, uint64(math.Float32bits(got.UpperBound.X)), got.UpperBound.X == moved.UpperBound.X)
		if !(got.UpperBound.X == moved.UpperBound.X) {
			t.Fatal("condition false")
		}
		s.flush("check/378")
		s.tree("check-state/378", &tree)
		s.check(379, uint64(math.Float32bits(got.UpperBound.Y)), got.UpperBound.Y == moved.UpperBound.Y)
		if !(got.UpperBound.Y == moved.UpperBound.Y) {
			t.Fatal("condition false")
		}
		s.flush("check/379")
		s.tree("check-state/379", &tree)
		// Now enlarge the proxy
		enlarge := AABB{LowerBound: Vec2{9.5, 9.5}, UpperBound: Vec2{11.5, 11.5}}
		tree.enlargeProxy(id, enlarge)
		s.tree("step/383", &tree)
		got2 := tree.AABB(id)
		s.check(386, uint64(math.Float32bits(got2.LowerBound.X)), got2.LowerBound.X <= enlarge.LowerBound.X+1e-6)
		if !(got2.LowerBound.X <= enlarge.LowerBound.X+1e-6) {
			t.Fatal("condition false")
		}
		s.flush("check/386")
		s.tree("check-state/386", &tree)
		s.check(387, uint64(math.Float32bits(got2.UpperBound.X)), got2.UpperBound.X >= enlarge.UpperBound.X-1e-6)
		if !(got2.UpperBound.X >= enlarge.UpperBound.X-1e-6) {
			t.Fatal("condition false")
		}
		s.flush("check/387")
		s.tree("check-state/387", &tree)
		tree = DynamicTree{}
		s.tree("step/389", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeRebuildAndValidateTest", func(t *testing.T) {

		s := treeOpen(t, "TreeRebuildAndValidateTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/395", &tree)
		// Create a number of proxies to make rebuild meaningful
		for i := 0; i < 12; i++ {
			x := float32(float32(i) * 2.0)
			a := AABB{LowerBound: Vec2{x - 0.5, -0.5}, UpperBound: Vec2{x + 0.5, 0.5}}
			createTreeProxyInternal(&tree, a, 0xFF, uint64(i), false)
			s.tree("step/402", &tree)
		}
		sorted := tree.rebuild(true)
		s.tree("step/405", &tree)
		s.check(407, uint64(sorted), sorted >= 0)
		if !(sorted >= 0) {
			t.Fatal("condition false")
		}
		s.flush("check/407")
		s.tree("check-state/407", &tree)
		s.check(408, uint64(treeBool(tree.ByteCount() > 0)), tree.ByteCount() > 0)
		if !(tree.ByteCount() > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/408")
		s.tree("check-state/408", &tree)
		s.check(409, uint64(tree.Height()), tree.Height() > 0)
		if !(tree.Height() > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/409")
		s.tree("check-state/409", &tree)
		tree = DynamicTree{}
		s.tree("step/411", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeRowHeightTest", func(t *testing.T) {

		s := treeOpen(t, "TreeRowHeightTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/417", &tree)
		columnCount := 200
		for i := 0; i < columnCount; i++ {
			x := float32(float32(1) * float32(i))
			a := AABB{LowerBound: Vec2{x, 0.0}, UpperBound: Vec2{x + 1.0, 1.0}}
			createTreeProxyInternal(&tree, a, 1, uint64(i), false)
			s.tree("step/424", &tree)
		}
		minHeight := s.fixtureFloats("min-height", 0)[0]
		s.r.F32("min-height", minHeight)
		tree.Validate()
		s.check(430, uint64(tree.Height()), float32(tree.Height()) < 2.0*minHeight)
		if !(float32(tree.Height()) < 2.0*minHeight) {
			t.Fatal("condition false")
		}
		s.flush("check/430")
		s.tree("check-state/430", &tree)
		tree = DynamicTree{}
		s.tree("step/432", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeGridHeightTest", func(t *testing.T) {

		s := treeOpen(t, "TreeGridHeightTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/438", &tree)
		columnCount := 20
		rowCount := 20
		for i := 0; i < columnCount; i++ {
			x := float32(float32(1) * float32(i))
			for j := 0; j < rowCount; j++ {
				y := float32(float32(1) * float32(j))
				a := AABB{LowerBound: Vec2{x, y}, UpperBound: Vec2{x + 1.0, y + 1.0}}
				createTreeProxyInternal(&tree, a, 1, uint64(i), false)
				s.tree("step/449", &tree)
			}
		}
		minHeight := s.fixtureFloats("min-height", 0)[0]
		s.r.F32("min-height", minHeight)
		tree.Validate()
		s.check(456, uint64(tree.Height()), float32(tree.Height()) < 2.0*minHeight)
		if !(float32(tree.Height()) < 2.0*minHeight) {
			t.Fatal("condition false")
		}
		s.flush("check/456")
		s.tree("check-state/456", &tree)
		tree = DynamicTree{}
		s.tree("step/458", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeGridMovementTest", func(t *testing.T) {

		s := treeOpen(t, "TreeGridMovementTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/466", &tree)
		var proxyIds [treeGridCount * treeGridCount]int
		index := 0
		for i := 0; i < treeGridCount; i++ {
			x := float32(float32(1) * float32(i))
			for j := 0; j < treeGridCount; j++ {
				y := float32(float32(1) * float32(j))
				a := AABB{LowerBound: Vec2{x, y}, UpperBound: Vec2{x + 1.0, y + 1.0}}
				proxyIds[index] = createTreeProxyInternal(&tree, a, 1, uint64(i), false)
				s.tree("step/477", &tree)
				index += 1
			}
		}
		s.check(482, uint64(index), index == treeGridCount*treeGridCount)
		if !(index == treeGridCount*treeGridCount) {
			t.Fatal("condition false")
		}
		s.flush("check/482")
		s.tree("check-state/482", &tree)
		minHeight := s.fixtureFloats("min-height", 0)[0]
		s.r.F32("min-height", minHeight)
		tree.Validate()
		height1 := tree.Height()
		s.check(488, uint64(height1), float32(height1) < 2.0*minHeight)
		if !(float32(height1) < 2.0*minHeight) {
			t.Fatal("condition false")
		}
		s.flush("check/488")
		s.tree("check-state/488", &tree)
		offset := Vec2{10.0, 20.0}
		index = 0
		for i := 0; i < treeGridCount; i++ {
			for j := 0; j < treeGridCount; j++ {
				a := tree.AABB(proxyIds[index])
				a.LowerBound = a.LowerBound.Add(offset)
				a.UpperBound = a.UpperBound.Add(offset)
				tree.moveProxyInternal(proxyIds[index], a, false)
				s.tree("step/499", &tree)
				index += 1
			}
		}
		tree.Validate()
		height2 := tree.Height()
		s.check(506, uint64(height2), float32(height2) < 3.0*minHeight)
		if !(float32(height2) < 3.0*minHeight) {
			t.Fatal("condition false")
		}
		s.flush("check/506")
		s.tree("check-state/506", &tree)
		tree.rebuild(true)
		s.tree("step/508", &tree)
		tree.Validate()
		height3 := tree.Height()
		s.check(512, uint64(height3), float32(height3) < 2.0*minHeight)
		if !(float32(height3) < 2.0*minHeight) {
			t.Fatal("condition false")
		}
		s.flush("check/512")
		s.tree("check-state/512", &tree)
		tree = DynamicTree{}
		s.tree("step/514", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeOneProxyTest", func(t *testing.T) {

		s := treeOpen(t, "TreeOneProxyTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(1)
		s.tree("step/798", &tree)
		box := AABB{Vec2{0.0, 0.0}, Vec2{1.0, 1.0}}
		proxyId := createTreeProxyInternal(&tree, box, 1, 7, false)
		s.tree("step/800", &tree)
		proxyIds := [1]int{proxyId}
		tree.Validate()
		// One proxy is a leaf at the root beside the empty node
		s.check(805, uint64(tree.nodeEnd), tree.nodeEnd == 2)
		if !(tree.nodeEnd == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/805")
		s.tree("check-state/805", &tree)
		s.check(806, uint64(tree.proxyCount), tree.proxyCount == 1)
		if !(tree.proxyCount == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/806")
		s.tree("check-state/806", &tree)
		s.check(807, uint64(treeBool(isLeaf(&tree.nodes[rootNode]))), isLeaf(&tree.nodes[rootNode]))
		if !(isLeaf(&tree.nodes[rootNode])) {
			t.Fatal("condition false")
		}
		s.flush("check/807")
		s.tree("check-state/807", &tree)
		s.check(808, uint64(tree.Height()), tree.Height() == 0)
		if !(tree.Height() == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/808")
		s.tree("check-state/808", &tree)
		s.check(809, uint64(tree.UserData(proxyId)), tree.UserData(proxyId) == 7)
		if !(tree.UserData(proxyId) == 7) {
			t.Fatal("condition false")
		}
		s.flush("check/809")
		s.tree("check-state/809", &tree)
		s.check(810, uint64(tree.CategoryBits(proxyId)), tree.CategoryBits(proxyId) == 1)
		if !(tree.CategoryBits(proxyId) == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/810")
		s.tree("check-state/810", &tree)
		s.check(811, uint64(treeBool(treeBoxEquals(tree.AABB(proxyId), box))), treeBoxEquals(tree.AABB(proxyId), box))
		if !(treeBoxEquals(tree.AABB(proxyId), box)) {
			t.Fatal("condition false")
		}
		s.flush("check/811")
		s.tree("check-state/811", &tree)
		s.check(812, uint64(treeBool(treeBoxEquals(tree.RootBounds(), box))), treeBoxEquals(tree.RootBounds(), box))
		if !(treeBoxEquals(tree.RootBounds(), box)) {
			t.Fatal("condition false")
		}
		s.flush("check/812")
		s.tree("check-state/812", &tree)
		inside := AABB{Vec2{0.4, 0.4}, Vec2{0.6, 0.6}}
		outside := AABB{Vec2{5.0, 5.0}, Vec2{6.0, 6.0}}
		result816 := treeCompareQuery(t, s, &tree, proxyIds[:], 1, inside)
		s.check(816, uint64(result816), result816 == 0)
		if !(result816 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/816")
		s.tree("check-state/816", &tree)
		result817 := treeCompareQuery(t, s, &tree, proxyIds[:], 1, outside)
		s.check(817, uint64(result817), result817 == 0)
		if !(result817 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/817")
		s.tree("check-state/817", &tree)
		origin := Vec2{-1.0, 0.5}
		translation := Vec2{3.0, 0.0}
		hit := treeCastHit{&tree, nullIndex, 1.0}
		input := RayCastInput{Origin: origin, Translation: translation, MaxFraction: 1.0}
		tree.CastRay(&input, DefaultMaskBits, treeClosestRay(&hit))
		s.tree("step/823", &tree)
		s.check(824, uint64(hit.proxyID), hit.proxyID == proxyId)
		if !(hit.proxyID == proxyId) {
			t.Fatal("condition false")
		}
		s.flush("check/824")
		s.tree("check-state/824", &tree)
		result825 := treeCompareRayCast(t, s, &tree, proxyIds[:], 1, origin, translation)
		s.check(825, uint64(result825), result825 == 0)
		if !(result825 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/825")
		s.tree("check-state/825", &tree)
		missOrigin := Vec2{-1.0, 5.0}
		result828 := treeCompareRayCast(t, s, &tree, proxyIds[:], 1, missOrigin, translation)
		s.check(828, uint64(result828), result828 == 0)
		if !(result828 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/828")
		s.tree("check-state/828", &tree)
		castBox := AABB{Vec2{-3.0, 0.25}, Vec2{-2.0, 0.75}}
		castTranslation := Vec2{10.0, 0.0}
		result832 := treeCompareBoxCast(t, s, &tree, proxyIds[:], 1, castBox, castTranslation)
		s.check(832, uint64(result832), result832 == 0)
		if !(result832 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/832")
		s.tree("check-state/832", &tree)
		// A full build keeps the leaf at the root
		rebuilt835 := tree.rebuild(true)
		s.check(835, uint64(rebuilt835), rebuilt835 == 1)
		if !(rebuilt835 == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/835")
		s.tree("check-state/835", &tree)
		tree.Validate()
		s.check(837, uint64(tree.nodeEnd), tree.nodeEnd == 2)
		if !(tree.nodeEnd == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/837")
		s.tree("check-state/837", &tree)
		s.check(838, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/838")
		s.tree("check-state/838", &tree)
		result839 := treeCompareQuery(t, s, &tree, proxyIds[:], 1, inside)
		s.check(839, uint64(result839), result839 == 0)
		if !(result839 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/839")
		s.tree("check-state/839", &tree)
		// The leaf is the root, so a mark and refit show up in the root bounds directly
		enlarged := treeEnlargeBox(box, 0.5)
		tree.markProxyMoved(proxyId, enlarged)
		s.tree("step/843", &tree)
		s.check(844, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree))
		if !(hasTreeMoved(&tree)) {
			t.Fatal("condition false")
		}
		s.flush("check/844")
		s.tree("check-state/844", &tree)
		tree.refit()
		s.tree("step/845", &tree)
		tree.Validate()
		s.check(847, uint64(treeBool(treeBoxEquals(tree.AABB(proxyId), enlarged))), treeBoxEquals(tree.AABB(proxyId), enlarged))
		if !(treeBoxEquals(tree.AABB(proxyId), enlarged)) {
			t.Fatal("condition false")
		}
		s.flush("check/847")
		s.tree("check-state/847", &tree)
		s.check(848, uint64(treeBool(treeBoxEquals(tree.RootBounds(), enlarged))), treeBoxEquals(tree.RootBounds(), enlarged))
		if !(treeBoxEquals(tree.RootBounds(), enlarged)) {
			t.Fatal("condition false")
		}
		s.flush("check/848")
		s.tree("check-state/848", &tree)
		rebuilt849 := tree.rebuild(false)
		s.check(849, uint64(rebuilt849), rebuilt849 == 1)
		if !(rebuilt849 == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/849")
		s.tree("check-state/849", &tree)
		s.check(850, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/850")
		s.tree("check-state/850", &tree)
		tree.validateNoMoved()
		tree.DestroyProxy(proxyId)
		s.tree("step/853", &tree)
		tree.Validate()
		s.check(855, uint64(tree.proxyCount), tree.proxyCount == 0)
		if !(tree.proxyCount == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/855")
		s.tree("check-state/855", &tree)
		s.check(856, uint64(tree.nodeEnd), tree.nodeEnd == 2)
		if !(tree.nodeEnd == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/856")
		s.tree("check-state/856", &tree)
		s.check(857, uint64(treeBool(isEmptyNode(&tree.nodes[rootNode]))), isEmptyNode(&tree.nodes[rootNode]))
		if !(isEmptyNode(&tree.nodes[rootNode])) {
			t.Fatal("condition false")
		}
		s.flush("check/857")
		s.tree("check-state/857", &tree)
		s.check(858, uint64(tree.Height()), tree.Height() == 0)
		if !(tree.Height() == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/858")
		s.tree("check-state/858", &tree)
		result859 := treeCompareQuery(t, s, &tree, proxyIds[:], 0, inside)
		s.check(859, uint64(result859), result859 == 0)
		if !(result859 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/859")
		s.tree("check-state/859", &tree)
		tree = DynamicTree{}
		s.tree("step/861", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeLeafRootTest", func(t *testing.T) {

		s := treeOpen(t, "TreeLeafRootTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/869", &tree)
		boxA := AABB{Vec2{0.0, 0.0}, Vec2{1.0, 1.0}}
		boxB := AABB{Vec2{10.0, 0.0}, Vec2{11.0, 1.0}}
		boxC := AABB{Vec2{20.0, 0.0}, Vec2{21.0, 1.0}}
		boxD := AABB{Vec2{30.0, 0.0}, Vec2{31.0, 1.0}}
		all := AABB{Vec2{-1.0, -1.0}, Vec2{40.0, 2.0}}
		idA := createTreeProxyInternal(&tree, boxA, 1, 0, false)
		s.tree("step/876", &tree)
		idB := createTreeProxyInternal(&tree, boxB, 1, 1, false)
		s.tree("step/877", &tree)
		tree.Validate()
		s.check(879, uint64(tree.nodeEnd), tree.nodeEnd == 4)
		if !(tree.nodeEnd == 4) {
			t.Fatal("condition false")
		}
		s.flush("check/879")
		s.tree("check-state/879", &tree)
		s.check(880, uint64(tree.proxyCount), tree.proxyCount == 2)
		if !(tree.proxyCount == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/880")
		s.tree("check-state/880", &tree)
		s.check(881, uint64(treeBool(isLeaf(&tree.nodes[rootNode]))), isLeaf(&tree.nodes[rootNode]) == false)
		if !(isLeaf(&tree.nodes[rootNode]) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/881")
		s.tree("check-state/881", &tree)
		s.check(882, uint64(tree.Height()), tree.Height() == 1)
		if !(tree.Height() == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/882")
		s.tree("check-state/882", &tree)
		tree.DestroyProxy(idA)
		s.tree("step/884", &tree)
		tree.Validate()
		s.check(886, uint64(treeBool(isLeaf(&tree.nodes[rootNode]))), isLeaf(&tree.nodes[rootNode]))
		if !(isLeaf(&tree.nodes[rootNode])) {
			t.Fatal("condition false")
		}
		s.flush("check/886")
		s.tree("check-state/886", &tree)
		s.check(887, uint64(tree.nodeEnd), tree.nodeEnd == 4)
		if !(tree.nodeEnd == 4) {
			t.Fatal("condition false")
		}
		s.flush("check/887")
		s.tree("check-state/887", &tree)
		s.check(888, uint64(tree.pairFreeList), tree.pairFreeList == 2)
		if !(tree.pairFreeList == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/888")
		s.tree("check-state/888", &tree)
		s.check(889, uint64(tree.proxyCount), tree.proxyCount == 1)
		if !(tree.proxyCount == 1) {
			t.Fatal("condition false")
		}
		s.flush("check/889")
		s.tree("check-state/889", &tree)
		idsB := [1]int{idB}
		result891 := treeCompareQuery(t, s, &tree, idsB[:], 1, all)
		s.check(891, uint64(result891), result891 == 0)
		if !(result891 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/891")
		s.tree("check-state/891", &tree)
		result892 := treeCompareQuery(t, s, &tree, idsB[:], 1, boxA)
		s.check(892, uint64(result892), result892 == 0)
		if !(result892 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/892")
		s.tree("check-state/892", &tree)
		// The freed pair is reused
		idC := createTreeProxyInternal(&tree, boxC, 1, 2, false)
		s.tree("step/895", &tree)
		tree.Validate()
		s.check(897, uint64(tree.nodeEnd), tree.nodeEnd == 4)
		if !(tree.nodeEnd == 4) {
			t.Fatal("condition false")
		}
		s.flush("check/897")
		s.tree("check-state/897", &tree)
		s.check(898, uint64(tree.pairFreeList), tree.pairFreeList == int32(nullIndex))
		if !(tree.pairFreeList == int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/898")
		s.tree("check-state/898", &tree)
		s.check(899, uint64(tree.proxyCount), tree.proxyCount == 2)
		if !(tree.proxyCount == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/899")
		s.tree("check-state/899", &tree)
		idsBC := [2]int{idB, idC}
		result901 := treeCompareQuery(t, s, &tree, idsBC[:], 2, all)
		s.check(901, uint64(result901), result901 == 0)
		if !(result901 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/901")
		s.tree("check-state/901", &tree)
		// A third proxy extends the array
		idD := createTreeProxyInternal(&tree, boxD, 1, 3, false)
		s.tree("step/904", &tree)
		tree.Validate()
		s.check(906, uint64(tree.nodeEnd), tree.nodeEnd == 6)
		if !(tree.nodeEnd == 6) {
			t.Fatal("condition false")
		}
		s.flush("check/906")
		s.tree("check-state/906", &tree)
		s.check(907, uint64(tree.proxyCount), tree.proxyCount == 3)
		if !(tree.proxyCount == 3) {
			t.Fatal("condition false")
		}
		s.flush("check/907")
		s.tree("check-state/907", &tree)
		idsBCD := [3]int{idB, idC, idD}
		result909 := treeCompareQuery(t, s, &tree, idsBCD[:], 3, all)
		s.check(909, uint64(result909), result909 == 0)
		if !(result909 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/909")
		s.tree("check-state/909", &tree)
		result910 := treeCompareQuery(t, s, &tree, idsBCD[:], 3, boxC)
		s.check(910, uint64(result910), result910 == 0)
		if !(result910 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/910")
		s.tree("check-state/910", &tree)
		// A remove leaves a hole that the full build compacts
		tree.DestroyProxy(idB)
		s.tree("step/913", &tree)
		tree.Validate()
		s.check(915, uint64(tree.nodeEnd), tree.nodeEnd == 6)
		if !(tree.nodeEnd == 6) {
			t.Fatal("condition false")
		}
		s.flush("check/915")
		s.tree("check-state/915", &tree)
		s.check(916, uint64(tree.pairFreeList), tree.pairFreeList != int32(nullIndex))
		if !(tree.pairFreeList != int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/916")
		s.tree("check-state/916", &tree)
		idsCD := [2]int{idC, idD}
		result918 := treeCompareQuery(t, s, &tree, idsCD[:], 2, all)
		s.check(918, uint64(result918), result918 == 0)
		if !(result918 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/918")
		s.tree("check-state/918", &tree)
		rebuilt919 := tree.rebuild(true)
		s.check(919, uint64(rebuilt919), rebuilt919 == 2)
		if !(rebuilt919 == 2) {
			t.Fatal("condition false")
		}
		s.flush("check/919")
		s.tree("check-state/919", &tree)
		tree.Validate()
		s.check(921, uint64(tree.nodeEnd), tree.nodeEnd == 4)
		if !(tree.nodeEnd == 4) {
			t.Fatal("condition false")
		}
		s.flush("check/921")
		s.tree("check-state/921", &tree)
		s.check(922, uint64(tree.pairFreeList), tree.pairFreeList == int32(nullIndex))
		if !(tree.pairFreeList == int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/922")
		s.tree("check-state/922", &tree)
		result923 := treeCompareQuery(t, s, &tree, idsCD[:], 2, all)
		s.check(923, uint64(result923), result923 == 0)
		if !(result923 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/923")
		s.tree("check-state/923", &tree)
		tree.DestroyProxy(idC)
		s.tree("step/925", &tree)
		tree.Validate()
		s.check(927, uint64(treeBool(isLeaf(&tree.nodes[rootNode]))), isLeaf(&tree.nodes[rootNode]))
		if !(isLeaf(&tree.nodes[rootNode])) {
			t.Fatal("condition false")
		}
		s.flush("check/927")
		s.tree("check-state/927", &tree)
		tree.DestroyProxy(idD)
		s.tree("step/929", &tree)
		tree.Validate()
		s.check(931, uint64(tree.proxyCount), tree.proxyCount == 0)
		if !(tree.proxyCount == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/931")
		s.tree("check-state/931", &tree)
		s.check(932, uint64(treeBool(isEmptyNode(&tree.nodes[rootNode]))), isEmptyNode(&tree.nodes[rootNode]))
		if !(isEmptyNode(&tree.nodes[rootNode])) {
			t.Fatal("condition false")
		}
		s.flush("check/932")
		s.tree("check-state/932", &tree)
		s.check(933, uint64(tree.Height()), tree.Height() == 0)
		if !(tree.Height() == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/933")
		s.tree("check-state/933", &tree)
		result934 := treeCompareQuery(t, s, &tree, idsBCD[:], 0, all)
		s.check(934, uint64(result934), result934 == 0)
		if !(result934 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/934")
		s.tree("check-state/934", &tree)
		tree = DynamicTree{}
		s.tree("step/936", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeQueryBruteForceTest", func(t *testing.T) {

		s := treeOpen(t, "TreeQueryBruteForceTest", 12345)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(treeProxyCount)
		s.tree("step/942", &tree)
		var proxyIds [treeProxyCount]int
		for i := 0; i < treeProxyCount; i++ {
			proxyIds[i] = createTreeProxyInternal(&tree, s.randomBox(2.0), 1, uint64(i), false)
			s.tree("step/946", &tree)
		}
		tree.Validate()
		for i := 0; i < 32; i++ {
			result952 := treeCompareQuery(t, s, &tree, proxyIds[:], treeProxyCount, s.randomBox(8.0))
			s.check(952, uint64(result952), result952 == 0)
			if !(result952 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/952")
			s.tree("check-state/952", &tree)
		}
		rebuilt955 := tree.rebuild(true)
		s.check(955, uint64(rebuilt955), rebuilt955 == treeProxyCount)
		if !(rebuilt955 == treeProxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/955")
		s.tree("check-state/955", &tree)
		tree.Validate()
		for i := 0; i < 32; i++ {
			result959 := treeCompareQuery(t, s, &tree, proxyIds[:], treeProxyCount, s.randomBox(8.0))
			s.check(959, uint64(result959), result959 == 0)
			if !(result959 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/959")
			s.tree("check-state/959", &tree)
		}
		// Moves leave the built order behind
		for i := 0; i < treeProxyCount; i += 2 {
			tree.moveProxyInternal(proxyIds[i], s.randomBox(2.0), false)
			s.tree("step/965", &tree)
		}
		tree.Validate()
		for i := 0; i < 32; i++ {
			result971 := treeCompareQuery(t, s, &tree, proxyIds[:], treeProxyCount, s.randomBox(8.0))
			s.check(971, uint64(result971), result971 == 0)
			if !(result971 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/971")
			s.tree("check-state/971", &tree)
		}
		tree = DynamicTree{}
		s.tree("step/974", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeCastBruteForceTest", func(t *testing.T) {

		s := treeOpen(t, "TreeCastBruteForceTest", 1416758761)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(treeProxyCount)
		s.tree("step/980", &tree)
		var proxyIds [treeProxyCount]int
		for i := 0; i < treeProxyCount; i++ {
			proxyIds[i] = createTreeProxyInternal(&tree, s.randomBox(2.0), 1, uint64(i), false)
			s.tree("step/984", &tree)
		}
		rebuilt987 := tree.rebuild(true)
		s.check(987, uint64(rebuilt987), rebuilt987 == treeProxyCount)
		if !(rebuilt987 == treeProxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/987")
		s.tree("check-state/987", &tree)
		tree.Validate()
		for i := 0; i < 64; i++ {
			origin := Vec2{s.randomFloat(-60.0, 60.0), s.randomFloat(-60.0, 60.0)}
			angle := float32(s.randomFloat(-Pi, Pi))
			length := float32(s.randomFloat(40.0, 120.0))
			s.r.F32("angle", angle)
			s.r.F32("length", length)
			components := s.fixtureFloats("translation", i)
			translation := Vec2{components[0], components[1]}
			s.r.Vec2("translation", translation.X, translation.Y)
			result996 := treeCompareRayCast(t, s, &tree, proxyIds[:], treeProxyCount, origin, translation)
			s.check(996, uint64(result996), result996 == 0)
			if !(result996 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/996")
			s.tree("check-state/996", &tree)
			halfExtent := Vec2{s.randomFloat(0.25, 2.0), s.randomFloat(0.25, 2.0)}
			box := AABB{origin.Sub(halfExtent), origin.Add(halfExtent)}
			result1000 := treeCompareBoxCast(t, s, &tree, proxyIds[:], treeProxyCount, box, translation)
			s.check(1000, uint64(result1000), result1000 == 0)
			if !(result1000 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/1000")
			s.tree("check-state/1000", &tree)
		}
		tree = DynamicTree{}
		s.tree("step/1003", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeMarkAndRefitTest", func(t *testing.T) {

		s := treeOpen(t, "TreeMarkAndRefitTest", 2736551817)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(treeProxyCount)
		s.tree("step/1010", &tree)
		var proxyIds [treeProxyCount]int
		for i := 0; i < treeProxyCount; i++ {
			proxyIds[i] = createTreeProxyInternal(&tree, s.randomBox(2.0), 1, uint64(i), false)
			s.tree("step/1014", &tree)
		}
		rebuilt1017 := tree.rebuild(true)
		s.check(1017, uint64(rebuilt1017), rebuilt1017 == treeProxyCount)
		if !(rebuilt1017 == treeProxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/1017")
		s.tree("check-state/1017", &tree)
		s.check(1018, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1018")
		s.tree("check-state/1018", &tree)
		s.check(1019, uint64(tree.nodeEnd), tree.nodeEnd == 2*treeProxyCount)
		if !(tree.nodeEnd == 2*treeProxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/1019")
		s.tree("check-state/1019", &tree)
		s.check(1020, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1020")
		s.tree("check-state/1020", &tree)
		var marked [treeProxyCount]bool
		var enlarged [treeProxyCount]AABB
		for i := 0; i < treeProxyCount; i += 3 {
			proxyId := proxyIds[i]
			enlarged[proxyId] = treeEnlargeBox(tree.AABB(proxyId), 0.5)
			tree.markProxyMoved(proxyId, enlarged[proxyId])
			s.tree("step/1028", &tree)
			marked[proxyId] = true
		}
		markedBelow := false
		result1033 := treeCheckMarks(t, s, &tree, rootNode, marked[:], &markedBelow)
		s.check(1033, uint64(result1033), result1033 == 0)
		if !(result1033 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1033")
		s.tree("check-state/1033", &tree)
		s.check(1034, uint64(treeBool(markedBelow)), markedBelow)
		if !(markedBelow) {
			t.Fatal("condition false")
		}
		s.flush("check/1034")
		s.tree("check-state/1034", &tree)
		s.check(1035, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree))
		if !(hasTreeMoved(&tree)) {
			t.Fatal("condition false")
		}
		s.flush("check/1035")
		s.tree("check-state/1035", &tree)
		// The leaf records took the fat boxes before the refit
		for i := 0; i < treeProxyCount; i++ {
			if marked[i] {
				s.check(1042, uint64(treeBool(treeBoxEquals(tree.AABB(i), enlarged[i]))), treeBoxEquals(tree.AABB(i), enlarged[i]))
				if !(treeBoxEquals(tree.AABB(i), enlarged[i])) {
					t.Fatal("condition false")
				}
				s.flush("check/1042")
				s.tree("check-state/1042", &tree)
			}
		}
		tree.refit()
		s.tree("step/1046", &tree)
		tree.Validate()
		var bounds AABB
		result1049 := treeCheckBounds(t, s, &tree, rootNode, &bounds)
		s.check(1049, uint64(result1049), result1049 == 0)
		if !(result1049 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1049")
		s.tree("check-state/1049", &tree)
		s.check(1050, uint64(treeBool(treeBoxEquals(bounds, tree.RootBounds()))), treeBoxEquals(bounds, tree.RootBounds()))
		if !(treeBoxEquals(bounds, tree.RootBounds())) {
			t.Fatal("condition false")
		}
		s.flush("check/1050")
		s.tree("check-state/1050", &tree)
		// Marks survive the refit
		result1053 := treeCheckMarks(t, s, &tree, rootNode, marked[:], &markedBelow)
		s.check(1053, uint64(result1053), result1053 == 0)
		if !(result1053 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1053")
		s.tree("check-state/1053", &tree)
		s.check(1054, uint64(treeBool(markedBelow)), markedBelow)
		if !(markedBelow) {
			t.Fatal("condition false")
		}
		s.flush("check/1054")
		s.tree("check-state/1054", &tree)
		for i := 0; i < 16; i++ {
			result1057 := treeCompareQuery(t, s, &tree, proxyIds[:], treeProxyCount, s.randomBox(8.0))
			s.check(1057, uint64(result1057), result1057 == 0)
			if !(result1057 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/1057")
			s.tree("check-state/1057", &tree)
		}
		// The rebuild clears every mark and restores the order
		rebuilt1061 := tree.rebuild(false)
		s.check(1061, uint64(rebuilt1061), rebuilt1061 > 0)
		if !(rebuilt1061 > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1061")
		s.tree("check-state/1061", &tree)
		s.check(1062, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1062")
		s.tree("check-state/1062", &tree)
		clear(marked[:])
		result1064 := treeCheckMarks(t, s, &tree, rootNode, marked[:], &markedBelow)
		s.check(1064, uint64(result1064), result1064 == 0)
		if !(result1064 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1064")
		s.tree("check-state/1064", &tree)
		s.check(1065, uint64(treeBool(markedBelow)), markedBelow == false)
		if !(markedBelow == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1065")
		s.tree("check-state/1065", &tree)
		s.check(1066, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1066")
		s.tree("check-state/1066", &tree)
		result1067 := treeCheckBounds(t, s, &tree, rootNode, &bounds)
		s.check(1067, uint64(result1067), result1067 == 0)
		if !(result1067 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1067")
		s.tree("check-state/1067", &tree)
		for i := 0; i < 16; i++ {
			result1070 := treeCompareQuery(t, s, &tree, proxyIds[:], treeProxyCount, s.randomBox(8.0))
			s.check(1070, uint64(result1070), result1070 == 0)
			if !(result1070 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/1070")
			s.tree("check-state/1070", &tree)
		}
		// A single deep leaf still marks the whole path to the root
		loneId := proxyIds[treeProxyCount/2]
		enlarged[loneId] = treeEnlargeBox(tree.AABB(loneId), 0.5)
		tree.markProxyMoved(loneId, enlarged[loneId])
		s.tree("step/1076", &tree)
		marked[loneId] = true
		result1078 := treeCheckMarks(t, s, &tree, rootNode, marked[:], &markedBelow)
		s.check(1078, uint64(result1078), result1078 == 0)
		if !(result1078 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1078")
		s.tree("check-state/1078", &tree)
		s.check(1079, uint64(treeBool(markedBelow)), markedBelow)
		if !(markedBelow) {
			t.Fatal("condition false")
		}
		s.flush("check/1079")
		s.tree("check-state/1079", &tree)
		tree.refit()
		s.tree("step/1080", &tree)
		result1081 := treeCheckBounds(t, s, &tree, rootNode, &bounds)
		s.check(1081, uint64(result1081), result1081 == 0)
		if !(result1081 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1081")
		s.tree("check-state/1081", &tree)
		rebuilt1082 := tree.rebuild(false)
		s.check(1082, uint64(rebuilt1082), rebuilt1082 > 0)
		if !(rebuilt1082 > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1082")
		s.tree("check-state/1082", &tree)
		tree.validateNoMoved()
		tree = DynamicTree{}
		s.tree("step/1085", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeDfsOrderTest", func(t *testing.T) {

		s := treeOpen(t, "TreeDfsOrderTest", 4061065257)
		s.r.U32("rng/start", s.randomState)
		proxyCount := 50
		tree := NewDynamicTree(treeProxyCount)
		s.tree("step/1094", &tree)
		var proxyIds [treeProxyCount]int
		for i := 0; i < proxyCount; i++ {
			proxyIds[i] = createTreeProxyInternal(&tree, s.randomBox(2.0), 1, uint64(i), false)
			s.tree("step/1098", &tree)
		}
		rebuilt1101 := tree.rebuild(true)
		s.check(1101, uint64(rebuilt1101), rebuilt1101 == proxyCount)
		if !(rebuilt1101 == proxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/1101")
		s.tree("check-state/1101", &tree)
		s.check(1102, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1102")
		s.tree("check-state/1102", &tree)
		s.check(1103, uint64(tree.nodeEnd), tree.nodeEnd == int32(2*proxyCount))
		if !(tree.nodeEnd == int32(2*proxyCount)) {
			t.Fatal("condition false")
		}
		s.flush("check/1103")
		s.tree("check-state/1103", &tree)
		s.check(1104, uint64(tree.pairFreeList), tree.pairFreeList == int32(nullIndex))
		if !(tree.pairFreeList == int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/1104")
		s.tree("check-state/1104", &tree)
		// A fresh pair lands above everything, so the order survives when the sibling is a leaf.
		// Validate checks the flag against the array either way.
		proxyIds[proxyCount] = createTreeProxyInternal(&tree, s.randomBox(2.0), 1, uint64(proxyCount), false)
		s.tree("step/1108", &tree)
		proxyCount += 1
		tree.Validate()
		rebuilt1112 := tree.rebuild(true)
		s.check(1112, uint64(rebuilt1112), rebuilt1112 == proxyCount)
		if !(rebuilt1112 == proxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/1112")
		s.tree("check-state/1112", &tree)
		s.check(1113, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1113")
		s.tree("check-state/1113", &tree)
		tree.moveProxyInternal(proxyIds[7], s.randomBox(2.0), false)
		s.tree("step/1115", &tree)
		tree.Validate()
		rebuilt1118 := tree.rebuild(true)
		s.check(1118, uint64(rebuilt1118), rebuilt1118 == proxyCount)
		if !(rebuilt1118 == proxyCount) {
			t.Fatal("condition false")
		}
		s.flush("check/1118")
		s.tree("check-state/1118", &tree)
		s.check(1119, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1119")
		s.tree("check-state/1119", &tree)
		orderedEnd := tree.nodeEnd
		// Removes leave holes but every survivor stays above its parent
		tree.DestroyProxy(proxyIds[3])
		s.tree("step/1123", &tree)
		tree.DestroyProxy(proxyIds[20])
		s.tree("step/1124", &tree)
		proxyIds[3] = proxyIds[proxyCount-1]
		proxyIds[20] = proxyIds[proxyCount-2]
		proxyCount -= 2
		s.check(1128, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1128")
		s.tree("check-state/1128", &tree)
		s.check(1129, uint64(tree.nodeEnd), tree.nodeEnd == orderedEnd)
		if !(tree.nodeEnd == orderedEnd) {
			t.Fatal("condition false")
		}
		s.flush("check/1129")
		s.tree("check-state/1129", &tree)
		s.check(1130, uint64(tree.pairFreeList), tree.pairFreeList != int32(nullIndex))
		if !(tree.pairFreeList != int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/1130")
		s.tree("check-state/1130", &tree)
		tree.Validate()
		enlarged := treeEnlargeBox(tree.AABB(proxyIds[10]), 0.5)
		tree.markProxyMoved(proxyIds[10], enlarged)
		s.tree("step/1134", &tree)
		s.check(1135, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree))
		if !(hasTreeMoved(&tree)) {
			t.Fatal("condition false")
		}
		s.flush("check/1135")
		s.tree("check-state/1135", &tree)
		tree.refit()
		s.tree("step/1136", &tree)
		tree.Validate()
		var bounds AABB
		result1139 := treeCheckBounds(t, s, &tree, rootNode, &bounds)
		s.check(1139, uint64(result1139), result1139 == 0)
		if !(result1139 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1139")
		s.tree("check-state/1139", &tree)
		s.check(1140, uint64(treeBool(treeBoxEquals(bounds, tree.RootBounds()))), treeBoxEquals(bounds, tree.RootBounds()))
		if !(treeBoxEquals(bounds, tree.RootBounds())) {
			t.Fatal("condition false")
		}
		s.flush("check/1140")
		s.tree("check-state/1140", &tree)
		for i := 0; i < 16; i++ {
			result1143 := treeCompareQuery(t, s, &tree, proxyIds[:], proxyCount, s.randomBox(8.0))
			s.check(1143, uint64(result1143), result1143 == 0)
			if !(result1143 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/1143")
			s.tree("check-state/1143", &tree)
		}
		rebuilt1146 := tree.rebuild(false)
		s.check(1146, uint64(rebuilt1146), rebuilt1146 > 0)
		if !(rebuilt1146 > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1146")
		s.tree("check-state/1146", &tree)
		s.check(1147, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1147")
		s.tree("check-state/1147", &tree)
		s.check(1148, uint64(tree.nodeEnd), tree.nodeEnd == int32(2*proxyCount))
		if !(tree.nodeEnd == int32(2*proxyCount)) {
			t.Fatal("condition false")
		}
		s.flush("check/1148")
		s.tree("check-state/1148", &tree)
		s.check(1149, uint64(tree.pairFreeList), tree.pairFreeList == int32(nullIndex))
		if !(tree.pairFreeList == int32(nullIndex)) {
			t.Fatal("condition false")
		}
		s.flush("check/1149")
		s.tree("check-state/1149", &tree)
		tree.Validate()
		// A box around everything makes the root the best sibling, so the root's subtree moves into
		// the new pair below its own children and the order is lost with nothing marked
		huge := AABB{Vec2{-100.0, -100.0}, Vec2{100.0, 100.0}}
		proxyIds[proxyCount] = createTreeProxyInternal(&tree, huge, 1, uint64(proxyCount), false)
		s.tree("step/1155", &tree)
		proxyCount += 1
		s.check(1157, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered == false)
		if !(tree.dfsOrdered == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1157")
		s.tree("check-state/1157", &tree)
		s.check(1158, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1158")
		s.tree("check-state/1158", &tree)
		tree.Validate()
		// The refit falls back to a descent on an unordered array
		enlarged = treeEnlargeBox(tree.AABB(proxyIds[10]), 0.5)
		tree.markProxyMoved(proxyIds[10], enlarged)
		s.tree("step/1163", &tree)
		tree.refit()
		s.tree("step/1164", &tree)
		tree.Validate()
		result1166 := treeCheckBounds(t, s, &tree, rootNode, &bounds)
		s.check(1166, uint64(result1166), result1166 == 0)
		if !(result1166 == 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1166")
		s.tree("check-state/1166", &tree)
		tree.clearMoved()
		s.tree("step/1167", &tree)
		s.check(1168, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1168")
		s.tree("check-state/1168", &tree)
		// The rebuild runs for the order alone
		rebuilt1171 := tree.rebuild(false)
		s.check(1171, uint64(rebuilt1171), rebuilt1171 > 0)
		if !(rebuilt1171 > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1171")
		s.tree("check-state/1171", &tree)
		s.check(1172, uint64(treeBool(tree.dfsOrdered)), tree.dfsOrdered)
		if !(tree.dfsOrdered) {
			t.Fatal("condition false")
		}
		s.flush("check/1172")
		s.tree("check-state/1172", &tree)
		s.check(1173, uint64(tree.nodeEnd), tree.nodeEnd == int32(2*proxyCount))
		if !(tree.nodeEnd == int32(2*proxyCount)) {
			t.Fatal("condition false")
		}
		s.flush("check/1173")
		s.tree("check-state/1173", &tree)
		tree.Validate()
		for i := 0; i < 16; i++ {
			result1177 := treeCompareQuery(t, s, &tree, proxyIds[:], proxyCount, s.randomBox(8.0))
			s.check(1177, uint64(result1177), result1177 == 0)
			if !(result1177 == 0) {
				t.Fatal("condition false")
			}
			s.flush("check/1177")
			s.tree("check-state/1177", &tree)
		}
		tree = DynamicTree{}
		s.tree("step/1180", &tree)
		s.r.U32("rng/end", s.randomState)
	})
	t.Run("TreeStaleAfterInsertRemoveTest", func(t *testing.T) {

		s := treeOpen(t, "TreeStaleAfterInsertRemoveTest", 1530990073)
		s.r.U32("rng/start", s.randomState)
		tree := NewDynamicTree(16)
		s.tree("step/1188", &tree)
		var ids [16]int
		for i := 0; i < 16; i++ {
			box := AABB{Vec2{float32(2) * float32(i), 0.0}, Vec2{float32(float32(2)*float32(i)) + 1.0, 1.0}}
			ids[i] = createTreeProxyInternal(&tree, box, 1, uint64(i), true)
			s.tree("step/1193", &tree)
		}
		rebuilt1196 := tree.rebuild(false)
		s.check(1196, uint64(rebuilt1196), rebuilt1196 == 16)
		if !(rebuilt1196 == 16) {
			t.Fatal("condition false")
		}
		s.flush("check/1196")
		s.tree("check-state/1196", &tree)
		// Control: enlarge and refit on the ordered tree
		big := AABB{Vec2{0.0, 0.0}, Vec2{40.0, 40.0}}
		tree.markProxyMoved(ids[0], big)
		s.tree("step/1200", &tree)
		tree.refit()
		s.tree("step/1201", &tree)
		s.check(1202, uint64(treeBool(tree.RootBounds().Contains(big))), tree.RootBounds().Contains(big))
		if !(tree.RootBounds().Contains(big)) {
			t.Fatal("condition false")
		}
		s.flush("check/1202")
		s.tree("check-state/1202", &tree)
		rebuilt1203 := tree.rebuild(false)
		s.check(1203, uint64(rebuilt1203), rebuilt1203 > 0)
		if !(rebuilt1203 > 0) {
			t.Fatal("condition false")
		}
		s.flush("check/1203")
		s.tree("check-state/1203", &tree)
		// Insert and remove between steps leaves no marks
		extra := AABB{Vec2{100.0, 100.0}, Vec2{101.0, 101.0}}
		extraId := createTreeProxyInternal(&tree, extra, 1, 99, true)
		s.tree("step/1207", &tree)
		tree.DestroyProxy(extraId)
		s.tree("step/1208", &tree)
		s.check(1209, uint64(treeBool(hasTreeMoved(&tree))), hasTreeMoved(&tree) == false)
		if !(hasTreeMoved(&tree) == false) {
			t.Fatal("condition false")
		}
		s.flush("check/1209")
		s.tree("check-state/1209", &tree)
		// The world's rebuild call, then a proxy enlarges during the step
		tree.rebuild(false)
		s.tree("step/1212", &tree)
		bigger := AABB{Vec2{-10.0, -10.0}, Vec2{60.0, 60.0}}
		tree.markProxyMoved(ids[0], bigger)
		s.tree("step/1214", &tree)
		tree.refit()
		s.tree("step/1215", &tree)
		s.check(1216, uint64(treeBool(tree.RootBounds().Contains(bigger))), tree.RootBounds().Contains(bigger))
		if !(tree.RootBounds().Contains(bigger)) {
			t.Fatal("condition false")
		}
		s.flush("check/1216")
		s.tree("check-state/1216", &tree)
		tree.Validate()
		tree = DynamicTree{}
		s.tree("step/1219", &tree)
		s.r.U32("rng/end", s.randomState)
	})
}

const treeHashInit uint64 = 14695981039346656037

type treeOracle struct {
	r           *oracle.Reader
	randomState uint32
	checks      []byte
	t           *testing.T
	expected    []treeObservation
	checkpoint  int
	occurrences map[uint32]int
	packed      map[string][][]byte
	packedNext  map[string]int
	inputs      map[string][]string
}

func treeHashU32(hash uint64, value uint32) uint64 {
	for i := 0; i < 4; i++ {
		hash ^= uint64(byte(value >> (8 * i)))
		hash *= 1099511628211
	}
	return hash
}

func treeHashU64(hash, value uint64) uint64 {
	return treeHashU32(treeHashU32(hash, uint32(value)), uint32(value>>32))
}

func treeHashBox(hash uint64, box AABB) uint64 {
	for _, value := range [...]float32{box.LowerBound.X, box.LowerBound.Y, box.UpperBound.X, box.UpperBound.Y} {
		hash = treeHashU32(hash, math.Float32bits(value))
	}
	return hash
}

func treeBool(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}

func treeHashNode(hash uint64, tree *DynamicTree, index int) uint64 {
	node := &tree.nodes[index]
	hash = treeHashU32(hash, uint32(index))
	hash = treeHashU32(hash, uint32(tree.parents[index]))
	hash = treeHashU32(hash, node.FlagIndex)
	hash = treeHashBox(hash, node.AABB)
	if node.FlagIndex == emptyNode {
		return hash
	}
	if node.FlagIndex&leafNode != 0 {
		hash = treeHashU32(hash, uint32(node.ShapeIndex))
		proxy := &tree.proxies[node.FlagIndex&nodeIndexMask]
		hash = treeHashU64(hash, proxy.UserData)
		hash = treeHashU64(hash, proxy.CategoryBits)
		return treeHashU32(hash, uint32(proxy.Node))
	}
	hash = treeHashU32(hash, uint32(node.Height))
	pair := int(node.FlagIndex & nodeIndexMask)
	hash = treeHashNode(hash, tree, pair)
	return treeHashNode(hash, tree, pair+1)
}

func (s *treeOracle) tree(key string, tree *DynamicTree) {
	hash := treeHashInit
	for _, value := range [...]int32{tree.nodeEnd, tree.nodeCapacity, tree.pairFreeList, tree.proxyCount, tree.proxyCapacity, tree.proxyFreeList} {
		hash = treeHashU32(hash, uint32(value))
	}
	hash = treeHashU32(hash, treeBool(tree.dfsOrdered))
	if tree.nodeEnd > 0 {
		hash = treeHashNode(hash, tree, rootNode)
	}
	s.r.U64(key, hash)
}

type treeObservation struct {
	line    uint32
	value   uint64
	outcome bool
}

func treeOpen(t *testing.T, name string, seed uint32) *treeOracle {
	t.Helper()
	config := "float"
	if _, double := any(Pos{}.X).(float64); double {
		config = "double"
	}
	data, err := os.ReadFile("testdata/oracle/" + config + "/dynamic_tree/" + name + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	s := &treeOracle{r: oracle.Open(t, "dynamic_tree/"+name), t: t, randomState: seed,
		occurrences: make(map[uint32]int), packed: make(map[string][][]byte), packedNext: make(map[string]int), inputs: make(map[string][]string)}
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var rec struct {
			Key   string `json:"k"`
			Type  string `json:"t"`
			Value string `json:"v"`
		}
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatal(err)
		}
		s.inputs[rec.Key] = append(s.inputs[rec.Key], rec.Value)
		if rec.Type != "bytes" {
			continue
		}
		packed, err := hex.DecodeString(rec.Value)
		if err != nil {
			t.Fatal(err)
		}
		s.packed[rec.Key] = append(s.packed[rec.Key], packed)
		if !strings.HasPrefix(rec.Key, "check/") && !strings.HasPrefix(rec.Key, "return/") {
			continue
		}
		for len(packed) > 0 {
			if len(packed) < 3 {
				t.Fatalf("truncated checkpoint in %s", rec.Key)
			}
			source := uint32(binary.LittleEndian.Uint16(packed))
			value, n := binary.Uvarint(packed[2:])
			if n <= 0 || len(packed) < 3+n {
				t.Fatalf("invalid checkpoint in %s", rec.Key)
			}
			s.expected = append(s.expected, treeObservation{source, value, packed[2+n] != 0})
			packed = packed[3+n:]
		}
	}
	t.Cleanup(func() {
		if !t.Skipped() && !t.Failed() && s.checkpoint != len(s.expected) {
			t.Errorf("read %d of %d upstream observations", s.checkpoint, len(s.expected))
		}
	})
	return s
}

func (s *treeOracle) check(line uint32, value uint64, outcome bool) {
	s.t.Helper()
	s.occurrences[line]++
	if s.checkpoint >= len(s.expected) {
		s.t.Fatalf("extra upstream test_dynamic_tree.c:%d observation", line)
	}
	want := s.expected[s.checkpoint]
	if want.line != line || want.value != value || want.outcome != outcome {
		s.t.Fatalf("first differing upstream observation: test_dynamic_tree.c:%d occurrence %d, checkpoint %d; C line %d value 0x%x, Go line %d value 0x%x; C outcome %t, Go outcome %t",
			want.line, s.occurrences[line], s.checkpoint+1, want.line, want.value, line, value, want.outcome, outcome)
	}
	s.checkpoint++
	s.checks = binary.LittleEndian.AppendUint16(s.checks, uint16(line))
	s.checks = binary.AppendUvarint(s.checks, value)
	s.checks = append(s.checks, byte(treeBool(outcome)))
}

func (s *treeOracle) flush(key string) {
	s.t.Helper()
	s.r.Bytes(key, s.checks)
	s.checks = s.checks[:0]
}

func (s *treeOracle) packedBytes(key string, got []byte) {
	s.t.Helper()
	index := s.packedNext[key]
	if index >= len(s.packed[key]) {
		s.t.Fatalf("extra packed %s observation", key)
	}
	want := s.packed[key][index]
	s.packedNext[key]++
	if !bytes.Equal(want, got) {
		if key == "query/hits" && len(want) == 30 && len(got) == 30 {
			for i := 0; i < treeProxyCount; i++ {
				if (want[i/8]^got[i/8])&(1<<(i%8)) != 0 {
					s.t.Fatalf("first differing query/hits observation at upstream CompareQuery, call %d, proxy %d", index+1, i)
				}
			}
		}
		s.t.Fatalf("first differing packed %s observation, call %d: C %x, Go %x", key, index+1, want, got)
	}
	s.r.Bytes(key, got)
}

// Small generator so the random tests repeat run to run
func (s *treeOracle) randomFloat(lower, upper float32) float32 {
	s.randomState = 1664525*s.randomState + 1013904223
	unit := float32(s.randomState>>8) * (float32(1) / float32(16777216))
	return lower + float32((upper-lower)*unit)
}

func (s *treeOracle) randomBox(maxHalfExtent float32) AABB {
	x := s.randomFloat(-50, 50)
	y := s.randomFloat(-50, 50)
	hx := s.randomFloat(0.25, maxHalfExtent)
	hy := s.randomFloat(0.25, maxHalfExtent)
	return AABB{Vec2{x - hx, y - hy}, Vec2{x + hx, y + hy}}
}

func treeBoxEquals(a, b AABB) bool {
	return a.LowerBound.X == b.LowerBound.X && a.LowerBound.Y == b.LowerBound.Y && a.UpperBound.X == b.UpperBound.X && a.UpperBound.Y == b.UpperBound.Y
}

func treeEnlargeBox(a AABB, margin float32) AABB {
	return AABB{Vec2{a.LowerBound.X - margin, a.LowerBound.Y - margin}, Vec2{a.UpperBound.X + margin, a.UpperBound.Y + margin}}
}

func treeCollectFlags(out []int) TreeQueryCallbackFunc {
	return func(proxyID int, _ uint64) bool { out[proxyID] = 1; return true }
}

func treeCollectList(list []int) TreeQueryCallbackFunc {
	return func(proxyID int, _ uint64) bool {
		count := list[0]
		list[count+1] = proxyID
		list[0] = count + 1
		return true
	}
}

type treeHitSet struct {
	hit   [treeProxyCount]bool
	count int
	bad   bool
}

func treeCollectHit(hits *treeHitSet) TreeQueryCallbackFunc {
	return func(proxyID int, _ uint64) bool {
		if proxyID < 0 || treeProxyCount <= proxyID || hits.hit[proxyID] {
			// Out of range or reported twice
			hits.bad = true
			return true
		}
		hits.hit[proxyID] = true
		hits.count++
		return true
	}
}

// Query the tree and compare against every live proxy box
func treeCompareQuery(t *testing.T, s *treeOracle, tree *DynamicTree, proxyIDs []int, proxyCount int, box AABB) int {
	var hits treeHitSet
	tree.Query(box, DefaultMaskBits, treeCollectHit(&hits))
	s.check(578, uint64(treeBool(hits.bad)), hits.bad == false)
	if !(hits.bad == false) {
		t.Fatal("condition false")
	}
	expectedCount := 0
	for i := 0; i < proxyCount; i++ {
		proxyID := proxyIDs[i]
		overlaps := tree.AABB(proxyID).Overlaps(box)
		s.check(585, uint64(treeBool(hits.hit[proxyID])), hits.hit[proxyID] == overlaps)
		if !(hits.hit[proxyID] == overlaps) {
			t.Fatal("condition false")
		}
		if overlaps {
			expectedCount++
		}
	}
	s.check(589, uint64(hits.count), hits.count == expectedCount)
	if !(hits.count == expectedCount) {
		t.Fatal("condition false")
	}
	packed := make([]byte, 30)
	for i, hit := range hits.hit {
		if hit {
			packed[i/8] |= 1 << (i % 8)
		}
	}
	binary.LittleEndian.PutUint32(packed[25:], uint32(hits.count))
	packed[29] = byte(treeBool(hits.bad))
	s.packedBytes("query/hits", packed)
	s.flush("return/590")
	return 0
}

type treeCastHit struct {
	tree     *DynamicTree
	proxyID  int
	fraction float32
}

// Entry fraction along the full translation, negative for a miss
func treeRayBoxFraction(box AABB, origin, translation Vec2) float32 {
	output := aabbRayCast(box, origin, origin.Add(translation))
	if output.Hit {
		return output.Fraction
	}
	return -1
}

// A box sweep against a box is a ray against the box grown by the swept extents
func treeBoxCastFraction(box AABB, input *BoxCastInput) float32 {
	extents := input.Box.Extents()
	grown := AABB{box.LowerBound.Sub(extents), box.UpperBound.Add(extents)}
	return treeRayBoxFraction(grown, input.Box.Center(), input.Translation)
}

// The fraction is computed against the full translation rather than the clipped input so it
// matches the brute force bit for bit
func treeClosestRay(hit *treeCastHit) TreeRayCastCallbackFunc {
	return func(input *RayCastInput, proxyID int, _ uint64) float32 {
		fraction := treeRayBoxFraction(hit.tree.AABB(proxyID), input.Origin, input.Translation)
		if fraction < 0 {
			return -1
		}
		if fraction < hit.fraction {
			hit.proxyID = proxyID
			hit.fraction = fraction
		}
		return fraction
	}
}

func treeClosestBox(hit *treeCastHit) TreeBoxCastCallbackFunc {
	return func(input *BoxCastInput, proxyID int, _ uint64) float32 {
		fraction := treeBoxCastFraction(hit.tree.AABB(proxyID), input)
		if fraction < 0 {
			return -1
		}
		if fraction < hit.fraction {
			hit.proxyID = proxyID
			hit.fraction = fraction
		}
		return fraction
	}
}

// Closest ray hit from the tree against the closest over every live proxy
func treeCompareRayCast(t *testing.T, s *treeOracle, tree *DynamicTree, proxyIDs []int, proxyCount int, origin, translation Vec2) int {
	hit := treeCastHit{tree, nullIndex, 1}
	input := RayCastInput{Origin: origin, Translation: translation, MaxFraction: 1}
	tree.CastRay(&input, DefaultMaskBits, treeClosestRay(&hit))
	bestID := nullIndex
	bestFraction := float32(1)
	unique := true
	for i := 0; i < proxyCount; i++ {
		proxyID := proxyIDs[i]
		fraction := treeRayBoxFraction(tree.AABB(proxyID), origin, translation)
		if fraction < 0 {
			continue
		}
		if fraction < bestFraction {
			bestID = proxyID
			bestFraction = fraction
			unique = true
		} else if fraction == bestFraction {
			unique = false
		}
	}
	s.check(686, uint64(math.Float32bits(hit.fraction)), hit.fraction == bestFraction)
	if !(hit.fraction == bestFraction) {
		t.Fatal("condition false")
	}
	s.r.F32("fraction/686", hit.fraction)
	if unique {
		s.check(689, uint64(hit.proxyID), hit.proxyID == bestID)
		if !(hit.proxyID == bestID) {
			t.Fatal("condition false")
		}
	}
	s.flush("return/692")
	return 0
}

func treeCompareBoxCast(t *testing.T, s *treeOracle, tree *DynamicTree, proxyIDs []int, proxyCount int, box AABB, translation Vec2) int {
	hit := treeCastHit{tree, nullIndex, 1}
	input := BoxCastInput{Box: box, Translation: translation, MaxFraction: 1}
	tree.CastBox(&input, DefaultMaskBits, treeClosestBox(&hit))
	bestID := nullIndex
	bestFraction := float32(1)
	unique := true
	for i := 0; i < proxyCount; i++ {
		proxyID := proxyIDs[i]
		fraction := treeBoxCastFraction(tree.AABB(proxyID), &input)
		if fraction < 0 {
			continue
		}
		if fraction < bestFraction {
			bestID = proxyID
			bestFraction = fraction
			unique = true
		} else if fraction == bestFraction {
			unique = false
		}
	}
	s.check(725, uint64(math.Float32bits(hit.fraction)), hit.fraction == bestFraction)
	if !(hit.fraction == bestFraction) {
		t.Fatal("condition false")
	}
	s.r.F32("fraction/725", hit.fraction)
	if unique {
		s.check(728, uint64(hit.proxyID), hit.proxyID == bestID)
		if !(hit.proxyID == bestID) {
			t.Fatal("condition false")
		}
	}
	s.flush("return/731")
	return 0
}

// A node is marked exactly when a leaf below it is marked. Leaves are checked against the
// set the test marked itself.
func treeCheckMarks(t *testing.T, s *treeOracle, tree *DynamicTree, nodeIndex int, markedProxies []bool, markedBelow *bool) int {
	node := &tree.nodes[nodeIndex]
	marked := isNodeMoved(node)
	if isLeaf(node) {
		s.check(742, uint64(treeBool(marked)), marked == markedProxies[getProxyID(node)])
		if !(marked == markedProxies[getProxyID(node)]) {
			t.Fatal("condition false")
		}
	} else {
		pair := getLeftChild(node)
		below1 := false
		below2 := false
		if treeCheckMarks(t, s, tree, pair, markedProxies, &below1) != 0 {
			return 1
		}
		if treeCheckMarks(t, s, tree, pair+1, markedProxies, &below2) != 0 {
			return 1
		}
		s.check(759, uint64(treeBool(marked)), marked == (below1 || below2))
		if !(marked == (below1 || below2)) {
			t.Fatal("condition false")
		}
	}
	*markedBelow = marked
	s.flush("return/763")
	return 0
}

// Every internal node holds exactly the union of its children
func treeCheckBounds(t *testing.T, s *treeOracle, tree *DynamicTree, nodeIndex int, bounds *AABB) int {
	node := &tree.nodes[nodeIndex]
	if isLeaf(node) == false {
		pair := getLeftChild(node)
		var below1, below2 AABB
		if treeCheckBounds(t, s, tree, pair, &below1) != 0 {
			return 1
		}
		if treeCheckBounds(t, s, tree, pair+1, &below2) != 0 {
			return 1
		}
		s.check(784, uint64(treeBool(treeBoxEquals(below1.Union(below2), node.AABB))), treeBoxEquals(below1.Union(below2), node.AABB))
		if !(treeBoxEquals(below1.Union(below2), node.AABB)) {
			t.Fatal("condition false")
		}
	}
	*bounds = node.AABB
	packed := make([]byte, 0, 16)
	for _, value := range [...]float32{bounds.LowerBound.X, bounds.LowerBound.Y, bounds.UpperBound.X, bounds.UpperBound.Y} {
		packed = binary.LittleEndian.AppendUint32(packed, math.Float32bits(value))
	}
	s.packedBytes("bounds", packed)
	s.flush("return/788")
	return 0
}

func (s *treeOracle) fixtureFloats(key string, index int) []float32 {
	s.t.Helper()
	if index >= len(s.inputs[key]) {
		s.t.Fatalf("missing oracle input %s/%d", key, index)
	}
	fields := strings.Split(s.inputs[key][index], ",")
	values := make([]float32, len(fields))
	for i, field := range fields {
		bits, err := strconv.ParseUint(field, 16, 32)
		if err != nil {
			s.t.Fatal(err)
		}
		values[i] = math.Float32frombits(uint32(bits))
	}
	return values
}
