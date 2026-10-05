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
	world      *World
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
	world           *World
	seeds           []nodePair
	staticSeedCount int
}

func createBroadPhase(bp *broadPhase, capacity *Capacity) {
	bp.pairSet = createSet(max(32, 2*capacity.ContactCount))
	bp.trees[StaticBody] = NewDynamicTree(max(16, capacity.StaticShapeCount))
	bp.trees[KinematicBody] = NewDynamicTree(16)
	bp.trees[DynamicBody] = NewDynamicTree(max(16, capacity.DynamicShapeCount))
	bp.movedSiblings = nil
}

func destroyBroadPhase(bp *broadPhase) {
	for range int(BodyTypeCount) {
	}

	destroySet(&bp.pairSet)
	*bp = broadPhase{}
}

func (bp *broadPhase) createProxy(proxyType BodyType, aabb AABB, categoryBits uint64, shapeIndex int, forcePairCreation bool) int {
	assert(0 <= proxyType && proxyType < BodyTypeCount, "0 <= proxyType && proxyType < b2_bodyTypeCount")
	mark := proxyType != StaticBody || forcePairCreation
	id := createTreeProxyInternal(&bp.trees[proxyType], aabb, categoryBits, uint64(shapeIndex), mark)

	return proxyKey(id, proxyType)
}

func (bp *broadPhase) destroyProxy(proxyKey int) {
	kind := proxyType(proxyKey)
	id := proxyID(proxyKey)

	assert(0 <= kind && kind < BodyTypeCount, "0 <= proxyType && proxyType < b2_bodyTypeCount")
	bp.trees[kind].DestroyProxy(id)
}

func (bp *broadPhase) moveProxy(proxyKey int, aabb AABB) {
	kind := proxyType(proxyKey)
	id := proxyID(proxyKey)
	bp.trees[kind].moveProxyInternal(id, aabb, true)
}

func gatherMovedSiblings(tree *DynamicTree, pairIndices []int) int {
	nodes := tree.nodes
	count := 0

	for pair := 2; pair < int(tree.nodeEnd); pair += 2 {
		if (nodes[pair].FlagIndex|nodes[pair+1].FlagIndex)&movedNode != 0 {
			pairIndices[count] = pair
			count++
		}
	}

	return count
}

func prefetchHash(set *hashSet, hash uint64) {}

func flushCandidatePairs(context *pairContext) {
	w := context.world
	bp := &w.broadPhase
	count1 := context.batchCount
	context.batchCount = 0

	var keys, hashes [candidateBatch]uint64

	for i := range count1 {
		candidate := &context.batch[i]
		keys[i] = shapePairKey(candidate.shapeIDA, candidate.shapeIDB)
		hashes[i] = keyHash(keys[i])
		prefetchHash(&bp.pairSet, hashes[i])
	}

	var candidates [candidateBatch]candidatePair

	count2 := 0

	for i := range count1 {
		pairExists := containsHashedKey(&bp.pairSet, keys[i], hashes[i])
		if !pairExists {
			candidates[count2] = context.batch[i]
			count2++
		}
	}

	pairKeys := context.pairKeys

	for i := range count2 {
		shapeIDA, shapeIDB := candidates[i].shapeIDA, candidates[i].shapeIDB
		if !candidateShapesCollide(w, shapeIDA, shapeIDB) {
			continue
		}

		*pairKeys = arrayPush(*pairKeys, shapePairKey(shapeIDA, shapeIDB))
	}
}

func candidateShapesCollide(w *World, shapeIDA, shapeIDB int) bool {
	shapeA, shapeB := &w.shapes[shapeIDA], &w.shapes[shapeIDB]

	bodyIDA, bodyIDB := shapeA.bodyID, shapeB.bodyID
	if bodyIDA == bodyIDB {
		return false
	}

	if shapeA.sensorIndex != nullIndex || shapeB.sensorIndex != nullIndex {
		return false
	}

	if !shouldShapesCollide(shapeA.filter, shapeB.filter) {
		return false
	}

	if !canCollide(shapeA.shapeType, shapeB.shapeType) {
		return false
	}

	bodyA, bodyB := &w.bodies[bodyIDA], &w.bodies[bodyIDB]
	if !shouldBodiesCollide(w, bodyA, bodyB) {
		return false
	}

	if shapeA.enableCustomFiltering || shapeB.enableCustomFiltering {
		customFilter := w.customFilterFcn
		if customFilter != nil {
			idA := Shape{index1: int32(shapeIDA + 1), world: w, generation: shapeA.generation}
			idB := Shape{index1: int32(shapeIDB + 1), world: w, generation: shapeB.generation}

			return customFilter(idA, idB)
		}
	}

	return true
}

func addCandidatePair(shapeIDA, shapeIDB int, context *pairContext) {
	candidate := &context.batch[context.batchCount]
	candidate.shapeIDA = min(shapeIDA, shapeIDB)
	candidate.shapeIDB = max(shapeIDA, shapeIDB)

	context.batchCount++
	if context.batchCount == candidateBatch {
		flushCandidatePairs(context)
	}
}

func testPair(a, b *TreeNode) bool {
	if (a.FlagIndex|b.FlagIndex)&movedNode == 0 {
		return false
	}

	return a.AABB.Overlaps(b.AABB)
}

func collideProxyAndSubtree(proxy *TreeNode, nodes []TreeNode, pair int, context *pairContext) {
	proxyMark := proxy.FlagIndex & movedNode
	boxv := loadAABBV(&proxy.AABB)
	shapeID := int(proxy.ShapeIndex)

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = pair

	for stackCount > 0 {
		stackCount--

		pair = stack[stackCount]
		for i := range 2 {
			node := &nodes[pair+i]
			if (node.FlagIndex|proxyMark)&movedNode == 0 {
				continue
			}

			if !overlapNode(boxv, node) {
				continue
			}

			if isLeaf(node) {
				addCandidatePair(shapeID, int(node.ShapeIndex), context)
			} else {
				if stackCount < treeStackSize {
					stack[stackCount] = getLeftChild(node)
					stackCount++
				} else {
					assert(stackCount < treeStackSize, "stackCount < B2_TREE_STACK_SIZE")
				}
			}
		}
	}
}

func visitPair(arrayA, arrayB []TreeNode, nodeA, nodeB *TreeNode, stack []indexPair, stackCount *int, context *pairContext) {
	if !testPair(nodeA, nodeB) {
		return
	}

	leafA, leafB := isLeaf(nodeA), isLeaf(nodeB)
	switch {
	case leafA && leafB:
		addCandidatePair(int(nodeA.ShapeIndex), int(nodeB.ShapeIndex), context)
	case leafA:
		collideProxyAndSubtree(nodeA, arrayB, getLeftChild(nodeB), context)
	case leafB:
		collideProxyAndSubtree(nodeB, arrayA, getLeftChild(nodeA), context)
	default:
		if *stackCount < treeStackSize {
			stack[*stackCount] = indexPair{a: getLeftChild(nodeA), b: getLeftChild(nodeB)}
			*stackCount += 1
		} else {
			assert(*stackCount < treeStackSize, "*stackCount < B2_TREE_STACK_SIZE")
		}
	}
}

func collideCrossPairs(arrayA, arrayB []TreeNode, subtreeA, subtreeB *TreeNode, context *pairContext) {
	var stack [treeStackSize]indexPair

	stackCount := 0
	visitPair(arrayA, arrayB, subtreeA, subtreeB, stack[:], &stackCount, context)

	for stackCount > 0 {
		stackCount--
		pair := stack[stackCount]

		for i := range 2 {
			for j := range 2 {
				visitPair(arrayA, arrayB, &arrayA[pair.a+i], &arrayB[pair.b+j], stack[:], &stackCount, context)
			}
		}
	}
}

func selfPairsTask(startIndex, endIndex, workerIndex int, context *World) {
	w := context
	bp := &w.broadPhase
	nodes := bp.trees[DynamicBody].nodes
	siblingIndices := bp.movedSiblings
	pairs := pairContext{world: w, pairKeys: &w.taskContexts[workerIndex].pairKeys}

	for i := startIndex; i < endIndex; i++ {
		nodeIndex := siblingIndices[i]
		collideCrossPairs(nodes, nodes, &nodes[nodeIndex], &nodes[nodeIndex+1], &pairs)
	}

	flushCandidatePairs(&pairs)
}

func gatherCrossSeeds(treeA, treeB *DynamicTree, seeds []nodePair) int {
	nodesA, nodesB := treeA.nodes, treeB.nodes

	var queue [2 * crossSeedCount]nodePair

	mask := 2*crossSeedCount - 1
	head, tail := 0, 0

	rootA, rootB := &nodesA[rootNode], &nodesB[rootNode]
	if testPair(rootA, rootB) {
		queue[tail&mask] = nodePair{a: *rootA, b: *rootB}
		tail++
	}

	seedCount := 0
	for head < tail && seedCount+(tail-head)+3 < crossSeedCount {
		pair := queue[head&mask]
		head++

		if isLeaf(&pair.a) || isLeaf(&pair.b) {
			seeds[seedCount] = pair
			seedCount++

			continue
		}

		a := nodesA[getLeftChild(&pair.a):]
		b := nodesB[getLeftChild(&pair.b):]

		for i := range 2 {
			for j := range 2 {
				if testPair(&a[i], &b[j]) {
					queue[tail&mask] = nodePair{a: a[i], b: b[j]}
					tail++
				}
			}
		}
	}

	for head < tail {
		seeds[seedCount] = queue[head&mask]
		seedCount++
		head++
	}

	return seedCount
}

func crossPairsTask(startIndex, endIndex, workerIndex int, context *crossContext) {
	w := context.world
	bp := &w.broadPhase
	staticNodes := bp.trees[StaticBody].nodes
	kinematicNodes := bp.trees[KinematicBody].nodes
	dynamicNodes := bp.trees[DynamicBody].nodes
	pairs := pairContext{world: w, pairKeys: &w.taskContexts[workerIndex].pairKeys}

	for i := startIndex; i < endIndex; i++ {
		nodesB := kinematicNodes
		if i < context.staticSeedCount {
			nodesB = staticNodes
		}

		seed := context.seeds[i]
		collideCrossPairs(dynamicNodes, nodesB, &seed.a, &seed.b, &pairs)
	}

	flushCandidatePairs(&pairs)
}

func updateTreesTask(context *World) {
	context.broadPhase.trees[DynamicBody].rebuild(false)
	context.broadPhase.trees[KinematicBody].rebuild(false)
}

func enqueueTreeUpdate(w *World) {
	if w.taskCount < MaxTasks {
		w.userTreeTask = w.enqueueTaskFcn(func() { updateTreesTask(w) })

		w.taskCount++
		if w.userTreeTask != nil {
			w.activeTaskCount++
		}
	} else {
		w.userTreeTask = nil
		updateTreesTask(w)
	}
}

func updateBroadPhasePairs(w *World) {
	bp := &w.broadPhase
	needUpdate := hasTreeMoved(&bp.trees[StaticBody])
	needUpdate = needUpdate || needsRebuild(&bp.trees[KinematicBody])

	needUpdate = needUpdate || needsRebuild(&bp.trees[DynamicBody])
	if !needUpdate {
		return
	}

	for i := range w.workerCount {
		w.taskContexts[i].pairKeys = w.taskContexts[i].pairKeys[:0]
	}
	{
		dynamicTree := &bp.trees[DynamicBody]
		pairCapacity := int(dynamicTree.nodeEnd) / 2
		bp.movedSiblings = make([]int, pairCapacity)
		dynamicMoveCount := gatherMovedSiblings(dynamicTree, bp.movedSiblings)

		var crossSeeds [2 * crossSeedCount]nodePair

		staticSeedCount := gatherCrossSeeds(dynamicTree, &bp.trees[StaticBody], crossSeeds[:])
		assert(staticSeedCount <= crossSeedCount, "staticSeedCount <= B2_CROSS_SEED_COUNT")
		kinematicSeedCount := gatherCrossSeeds(dynamicTree, &bp.trees[KinematicBody], crossSeeds[staticSeedCount:])
		assert(kinematicSeedCount <= crossSeedCount, "kinematicSeedCount <= B2_CROSS_SEED_COUNT")
		crossMoveCount := staticSeedCount + kinematicSeedCount
		cross := crossContext{world: w, seeds: crossSeeds[:], staticSeedCount: staticSeedCount}
		parallelFor(w, func(start, end, worker int) { crossPairsTask(start, end, worker, &cross) }, crossMoveCount, 1)
		parallelFor(w, func(start, end, worker int) { selfPairsTask(start, end, worker, w) }, dynamicMoveCount, 64)
	}

	bp.trees[StaticBody].clearMoved()
	enqueueTreeUpdate(w)

	pairCount := 0
	for i := range w.workerCount {
		pairCount += len(w.taskContexts[i].pairKeys)
	}

	pairKeys := make([]uint64, max(pairCount, 1))
	keyCount := 0

	for i := range w.workerCount {
		workerKeys := w.taskContexts[i].pairKeys
		if len(workerKeys) > 0 {
			copy(pairKeys[keyCount:], workerKeys)
			keyCount += len(workerKeys)
		}
	}

	assert(keyCount == pairCount, "keyCount == pairCount")
	qsort(pairKeys[:pairCount], func(a, b uint64) bool { return a < b })

	for i := range keyCount {
		shapeIDA := int(pairKeys[i] >> 32)
		shapeIDB := int(pairKeys[i] & 0xFFFFFFFF)
		shapeA, shapeB := &w.shapes[shapeIDA], &w.shapes[shapeIDB]
		createContact(w, shapeA, shapeB)
	}

	bp.movedSiblings = nil

	validateSolverSets(w)
}

func validateBroadphase(bp *broadPhase) {
	bp.trees[DynamicBody].Validate()
	bp.trees[KinematicBody].Validate()
}

func validateNoEnlarged(bp *broadPhase) {
	if !validationEnabled {
		return
	}

	for j := range int(BodyTypeCount) {
		bp.trees[j].validateNoMoved()
	}
}

func (bp *broadPhase) markProxyMovedSerial(proxyKey int) {
	kind := proxyType(proxyKey)
	id := proxyID(proxyKey)
	bp.trees[kind].markProxyMovedSerial(id)
}

func (bp *broadPhase) markProxyMoved(proxyKey int, aabb AABB) {
	kind := proxyType(proxyKey)
	id := proxyID(proxyKey)
	bp.trees[kind].markProxyMoved(id, aabb)
}

func proxyType(key int) BodyType {
	return BodyType(key & 3)
}

func proxyID(key int) int {
	return key >> 2
}

func proxyKey(id int, proxyType BodyType) int {
	return id<<2 | int(proxyType)
}
