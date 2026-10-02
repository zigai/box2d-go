// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

const (
	candidateBatch = 32
	crossSeedCount = 64
)

type broadPhase struct {
	trees         [BodyTypeCount]DynamicTree
	movedSiblings []int
	pairSet       hashSet
}

type candidatePair struct {
	shapeIDA int
	shapeIDB int
}

type pairContext struct {
	world      *world
	pairKeys   *[]uint64
	batch      [candidateBatch]candidatePair
	batchCount int
}

type nodePair struct {
	a TreeNode
	b TreeNode
}

type indexPair struct {
	a int
	b int
}

type crossContext struct {
	world           *world
	seeds           []nodePair
	staticSeedCount int
}

func createBroadPhase(bp *broadPhase, capacity *Capacity) {
	todo()
}

func destroyBroadPhase(bp *broadPhase) {
	todo()
}

func (bp *broadPhase) createProxy(proxyType BodyType, aabb AABB, categoryBits uint64, shapeIndex int, forcePairCreation bool) int {
	todo()
	return 0
}

func (bp *broadPhase) destroyProxy(proxyKey int) {
	todo()
}

func (bp *broadPhase) moveProxy(proxyKey int, aabb AABB) {
	todo()
}

func gatherMovedSiblings(tree *DynamicTree, pairIndices []int) int {
	todo()
	return 0
}

func prefetchHash(set *hashSet, hash uint64) {
	todo()
}

func flushCandidatePairs(context *pairContext) {
	todo()
}

func addCandidatePair(shapeIDA, shapeIDB int, context *pairContext) {
	todo()
}

func testPair(a, b *TreeNode) bool {
	todo()
	return false
}

func collideProxyAndSubtree(proxy *TreeNode, nodes []TreeNode, pair int, context *pairContext) {
	todo()
}

func visitPair(arrayA, arrayB []TreeNode, nodeA, nodeB *TreeNode, stack []indexPair, stackCount *int, context *pairContext) {
	todo()
}

func collideCrossPairs(arrayA, arrayB []TreeNode, subtreeA, subtreeB *TreeNode, context *pairContext) {
	todo()
}

func selfPairsTask(startIndex, endIndex, workerIndex int, context *world) {
	todo()
}

func gatherCrossSeeds(treeA, treeB *DynamicTree, seeds []nodePair) int {
	todo()
	return 0
}

func crossPairsTask(startIndex, endIndex, workerIndex int, context *crossContext) {
	todo()
}

func updateTreesTask(context *world) {
	todo()
}

func enqueueTreeUpdate(w *world) {
	todo()
}

func updateBroadPhasePairs(w *world) {
	todo()
}

func validateBroadphase(bp *broadPhase) {
	todo()
}

func validateNoEnlarged(bp *broadPhase) {
	todo()
}

func (bp *broadPhase) markProxyMovedSerial(proxyKey int) {
	todo()
}

func (bp *broadPhase) markProxyMoved(proxyKey int, aabb AABB) {
	todo()
}

func proxyType(key int) BodyType {
	todo()
	return 0
}

func proxyID(key int) int {
	todo()
	return 0
}

func proxyKey(id int, proxyType BodyType) int {
	todo()
	return 0
}
