// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/zigai/box2d-go/internal/core"
)

const (
	treeStackSize int    = 512
	movedNode     uint32 = 1 << 30
	leafNode      uint32 = 1 << 31
	nodeIndexMask uint32 = 0xFFFFFFFF & ^(movedNode | leafNode)
	emptyNode     uint32 = nodeIndexMask | leafNode
	rootNode      int    = 0
	treeHeuristic int    = 0
)

// A node in the dynamic tree. Siblings sit together at an even index so two nodes fit in
// a 64 byte cache line. The root is at index zero and index one always empty.
type TreeNode struct {
	// The node bounding box
	AABB AABB
	// In 3D this space is used by the AABB z components.
	Padding uint64
	// bit 31 : 1 for leaf node
	// bit 30 : 1 for moved flag
	// bits 0-29 : index of the sibling pair node or the proxy id for a leaf
	FlagIndex uint32
	// The height of an internal node. A leaf has zero height.
	Height int32
	// The shape index for a leaf. Truncated from proxy user data.
	ShapeIndex int32
}

// Separate storage for tree leaves.
type TreeProxy struct {
	// User data is an index instead of void* because it is used internally as a shape index.
	UserData uint64
	// Category bits for collision filtering.
	CategoryBits uint64
	// The leaf node. nullIndex for a free proxy.
	Node int32
	// Next free proxy.
	Next int32
}

// The dynamic tree structure. This should be considered private data.
// It is placed here for performance reasons.
type DynamicTree struct {
	// Array of nodes. The root is at index zero and index 1 is empty.
	// Otherwise siblings are paired at even indices. Has holes for free node pairs.
	nodes []TreeNode
	// Parent index per node. The free list is interweaved.
	parents []int32
	// Proxy data split from node array as cold data.
	proxies []TreeProxy
	// Every allocated node has a lower index than this.
	nodeEnd int32
	// The allocated node space
	nodeCapacity int32
	// Free pairs below nodeEnd
	pairFreeList int32
	// Number of proxies created
	proxyCount int32
	// The allocated proxy space
	proxyCapacity int32
	// Proxy free list
	proxyFreeList int32
	// Array of nodes for rebuild.
	swapNodes []TreeNode
	// Leaf indices for rebuild
	leafIndices []int32
	// Leaves for the rebuild. May represent a proxy or a retained subtree.
	leafNodes []TreeNode
	// Leaf bounding boxes for rebuild
	leafBoxes []AABB
	// Leaf bounding box centers for rebuild
	leafCenters []Vec2
	// Bins for sorting during rebuild
	binIndices []int32
	// Allocated space for rebuilding
	rebuildCapacity int32
	// Rebuild orders the nodes so the children follow parents. Cache friendly for queries
	// and refitting. The order can be disrupted by proxy creation.
	dfsOrdered bool
}

// These are performance results returned by dynamic tree queries.
type TreeStats struct {
	// Number of internal nodes visited during the query
	NodeVisits int
	// Number of leaf nodes visited during the query
	LeafVisits int
}

// This function receives proxies found in the AABB query.
// @return true if the query should continue.
type TreeQueryCallbackFunc func(proxyID int, userData uint64) bool

// This function receives clipped ray cast input for a proxy. The function
// returns the new ray fraction.
// - return a value of 0 to terminate the ray cast
// - return a value less than input.MaxFraction to clip the ray
// - return a value of input.MaxFraction to continue the ray cast without clipping.
type TreeRayCastCallbackFunc func(input *RayCastInput, proxyID int, userData uint64) float32

// Input for casting an AABB through a dynamic tree.
type BoxCastInput struct {
	// The AABB to cast, in the tree's frame.
	Box AABB
	// The sweep translation.
	Translation Vec2
	// The maximum fraction of the translation to consider, typically 1.
	MaxFraction float32
}

// This function receives clipped AABB cast input for a proxy. The function
// returns the new cast fraction.
// - return a value of 0 to terminate the cast
// - return a value less than input.MaxFraction to clip the cast
// - return a value of input.MaxFraction to continue the cast without clipping.
type TreeBoxCastCallbackFunc func(input *BoxCastInput, proxyID int, userData uint64) float32

type rebuildItem struct {
	nodeIndex  int
	pair       int
	childCount int
	startIndex int
	splitIndex int
	endIndex   int
}

type copyItem struct {
	oldPair  int
	newIndex int
}

// Constructing the tree initializes the node pool.
func NewDynamicTree(proxyCapacity int) DynamicTree {
	capacity := max(proxyCapacity, 16)
	tree := DynamicTree{}
	tree.nodeCapacity = int32(2 * capacity)
	tree.nodes = make([]TreeNode, tree.nodeCapacity)
	tree.parents = make([]int32, tree.nodeCapacity)
	tree.pairFreeList = int32(nullIndex)
	tree.nodes[rootNode] = makeEmptyNode()
	tree.nodes[rootNode+1] = makeEmptyNode()
	tree.parents[rootNode] = int32(nullIndex)
	tree.parents[rootNode+1] = int32(nullIndex)
	tree.nodeEnd = 2
	tree.dfsOrdered = true
	tree.proxyCapacity = int32(capacity)

	tree.proxies = make([]TreeProxy, capacity)
	for i := range capacity - 1 {
		tree.proxies[i].Node = int32(nullIndex)
		tree.proxies[i].Next = int32(i + 1)
	}

	tree.proxies[capacity-1].Node = int32(nullIndex)
	tree.proxies[capacity-1].Next = int32(nullIndex)

	return tree
}

func allocateSiblingPair(tree *DynamicTree) int {
	if tree.pairFreeList != int32(nullIndex) {
		pair := int(tree.pairFreeList)
		tree.pairFreeList = tree.parents[pair]

		return pair
	}

	if tree.nodeEnd+2 > tree.nodeCapacity {
		oldCapacity := tree.nodeCapacity
		newCapacity := oldCapacity + (oldCapacity >> 1)
		newCapacity += newCapacity & 1
		nodes := make([]TreeNode, newCapacity)
		copy(nodes, tree.nodes)
		tree.nodes = nodes
		parents := make([]int32, newCapacity)
		copy(parents, tree.parents)
		tree.parents = parents
		tree.nodeCapacity = newCapacity
		tree.swapNodes = nil
	}

	pair := int(tree.nodeEnd)
	tree.nodeEnd += 2

	return pair
}

func freePair(tree *DynamicTree, pair int) {
	core.Assert(
		pair&1 == 0 && 2 <= pair && pair < int(tree.nodeEnd),
		"( pair & 1 ) == 0 && 2 <= pair && pair < tree->nodeEnd",
	)
	tree.nodes[pair] = makeEmptyNode()
	tree.nodes[pair+1] = makeEmptyNode()
	tree.parents[pair] = tree.pairFreeList
	tree.parents[pair+1] = int32(nullIndex)
	tree.pairFreeList = int32(pair)
}

func allocateProxy(tree *DynamicTree) int {
	if tree.proxyFreeList == int32(nullIndex) {
		core.Assert(
			tree.proxyCount == tree.proxyCapacity,
			"tree->proxyCount == tree->proxyCapacity",
		)
		oldCapacity := tree.proxyCapacity
		tree.proxyCapacity += max(oldCapacity>>1, 1)
		proxies := make([]TreeProxy, tree.proxyCapacity)
		copy(proxies, tree.proxies)

		tree.proxies = proxies
		for i := oldCapacity; i < tree.proxyCapacity-1; i++ {
			tree.proxies[i].Node = int32(nullIndex)
			tree.proxies[i].Next = i + 1
		}

		tree.proxies[tree.proxyCapacity-1].Node = int32(nullIndex)
		tree.proxies[tree.proxyCapacity-1].Next = int32(nullIndex)
		tree.proxyFreeList = oldCapacity
	}

	proxyIndex := int(tree.proxyFreeList)
	tree.proxyFreeList = tree.proxies[proxyIndex].Next
	tree.proxies[proxyIndex] = TreeProxy{Node: int32(nullIndex), Next: int32(nullIndex)}
	tree.proxyCount++

	return proxyIndex
}

func freeProxy(tree *DynamicTree, proxyID int) {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	core.Assert(0 < tree.proxyCount, "0 < tree->proxyCount")
	tree.proxies[proxyID].Node = int32(nullIndex)
	tree.proxies[proxyID].Next = tree.proxyFreeList
	tree.proxyFreeList = int32(proxyID)
	tree.proxyCount--
}

func getNodeHeight(node *TreeNode) int {
	if isLeaf(node) {
		return 0
	}

	return int(node.Height)
}

func makeInternalNode(nodes []TreeNode, pair int) TreeNode {
	c1, c2 := &nodes[pair], &nodes[pair+1]

	return TreeNode{
		AABB:      c1.AABB.Union(c2.AABB),
		FlagIndex: uint32(pair) | ((c1.FlagIndex | c2.FlagIndex) & movedNode),
		Height:    int32(1 + max(getNodeHeight(c1), getNodeHeight(c2))),
	}
}

func makeLeafNode(aabb AABB, proxyID int, userData uint64, moved bool) TreeNode {
	node := TreeNode{AABB: aabb, FlagIndex: uint32(proxyID) | leafNode, ShapeIndex: int32(userData)}
	if moved {
		node.FlagIndex |= movedNode
	}

	return node
}

func linkChildren(tree *DynamicTree, nodeIndex int) {
	node := &tree.nodes[nodeIndex]
	if isLeaf(node) {
		tree.proxies[getProxyID(node)].Node = int32(nodeIndex)
	} else {
		pair := getLeftChild(node)
		tree.parents[pair] = int32(nodeIndex)
		tree.parents[pair+1] = int32(nodeIndex)
	}
}

func isNodeOrdered(nodes []TreeNode, nodeIndex int) bool {
	node := &nodes[nodeIndex]
	return isLeaf(node) || nodeIndex < getLeftChild(node)
}

func findBestSibling(tree *DynamicTree, boxD AABB) int {
	centerD := boxD.Center()
	areaD := perimeter(boxD)
	nodes := tree.nodes

	nodeIndex := rootNode
	if isLeaf(&nodes[nodeIndex]) {
		return nodeIndex
	}

	rootBox := nodes[nodeIndex].AABB
	areaBase := perimeter(rootBox)
	directCost := perimeter(rootBox.Union(boxD))
	inheritedCost := float32(0)
	bestSibling := nodeIndex
	bestCost := directCost

	for {
		child1 := getLeftChild(&nodes[nodeIndex])
		child2 := child1 + 1

		cost := directCost + inheritedCost
		if cost < bestCost {
			bestSibling = nodeIndex
			bestCost = cost
		}

		inheritedCost += directCost - areaBase
		leaf1 := isLeaf(&nodes[child1])
		leaf2 := isLeaf(&nodes[child2])
		lowerCost1 := float32(math.MaxFloat32)
		box1 := nodes[child1].AABB
		directCost1 := perimeter(box1.Union(boxD))
		area1 := float32(0)

		if leaf1 {
			cost1 := directCost1 + inheritedCost
			if cost1 < bestCost {
				bestSibling = child1
				bestCost = cost1
			}
		} else {
			area1 = perimeter(box1)
			lowerCost1 = inheritedCost + directCost1 + minFloat(areaD-area1, 0)
		}

		lowerCost2 := float32(math.MaxFloat32)
		box2 := nodes[child2].AABB
		directCost2 := perimeter(box2.Union(boxD))
		area2 := float32(0)

		if leaf2 {
			cost2 := directCost2 + inheritedCost
			if cost2 < bestCost {
				bestSibling = child2
				bestCost = cost2
			}
		} else {
			area2 = perimeter(box2)
			lowerCost2 = inheritedCost + directCost2 + minFloat(areaD-area2, 0)
		}

		if leaf1 && leaf2 {
			break
		}

		if bestCost <= lowerCost1 && bestCost <= lowerCost2 {
			break
		}

		if lowerCost1 == lowerCost2 && !leaf1 {
			core.Assert(lowerCost1 < math.MaxFloat32, "lowerCost1 < FLT_MAX")
			core.Assert(lowerCost2 < math.MaxFloat32, "lowerCost2 < FLT_MAX")

			d1 := box1.Center().Sub(centerD)
			d2 := box2.Center().Sub(centerD)
			lowerCost1 = d1.LengthSquared()
			lowerCost2 = d2.LengthSquared()
		}

		if lowerCost1 < lowerCost2 && !leaf1 {
			nodeIndex = child1
			areaBase = area1
			directCost = directCost1
		} else {
			nodeIndex = child2
			areaBase = area2
			directCost = directCost2
		}
	}

	return bestSibling
}

func swapNodes(tree *DynamicTree, iDown int, iUp int) {
	nodes := tree.nodes
	nodes[iDown], nodes[iUp] = nodes[iUp], nodes[iDown]
	linkChildren(tree, iDown)
	linkChildren(tree, iUp)

	if !isNodeOrdered(nodes, iDown) || !isNodeOrdered(nodes, iUp) {
		tree.dfsOrdered = false
	}

	iC := iDown ^ 1
	nodes[iC] = makeInternalNode(nodes, getLeftChild(&nodes[iC]))
}

func rotateNodes(tree *DynamicTree, iA int) {
	nodes := tree.nodes
	a := &nodes[iA]
	core.Assert(!isLeaf(a), "b2IsLeaf( A ) == false")
	iB := getLeftChild(a)
	iC := iB + 1
	b, c := &nodes[iB], &nodes[iC]

	leafB, leafC := isLeaf(b), isLeaf(c)
	if leafB && leafC {
		return
	}

	bestDown, bestUp := nullIndex, nullIndex
	bestDelta := float32(0)

	if !leafC {
		iF := getLeftChild(c)
		iG := iF + 1
		areaC := perimeter(c.AABB)

		deltaBF := perimeter(unionV(b.AABB, nodes[iG].AABB)) - areaC
		if deltaBF < bestDelta {
			bestDown = iB
			bestUp = iF
			bestDelta = deltaBF
		}

		deltaBG := perimeter(unionV(b.AABB, nodes[iF].AABB)) - areaC
		if deltaBG < bestDelta {
			bestDown = iB
			bestUp = iG
			bestDelta = deltaBG
		}
	}

	if !leafB {
		iD := getLeftChild(b)
		iE := iD + 1
		areaB := perimeter(b.AABB)

		deltaCD := perimeter(unionV(c.AABB, nodes[iE].AABB)) - areaB
		if deltaCD < bestDelta {
			bestDown = iC
			bestUp = iD
			bestDelta = deltaCD
		}

		deltaCE := perimeter(unionV(c.AABB, nodes[iD].AABB)) - areaB
		if deltaCE < bestDelta {
			bestDown = iC
			bestUp = iE
		}
	}

	if bestDown != nullIndex {
		swapNodes(tree, bestDown, bestUp)
	}
}

func insertLeaf(tree *DynamicTree, aabb AABB, proxyID int, moved bool, shouldRotate bool) {
	proxy := &tree.proxies[proxyID]

	leaf := makeLeafNode(aabb, proxyID, proxy.UserData, moved)
	if isEmptyNode(&tree.nodes[rootNode]) {
		tree.nodes[rootNode] = leaf
		proxy.Node = int32(rootNode)

		return
	}

	sibling := findBestSibling(tree, aabb)
	pair := allocateSiblingPair(tree)
	nodes, parents := tree.nodes, tree.parents
	nodes[pair] = nodes[sibling]
	nodes[pair+1] = leaf
	parents[pair] = int32(sibling)
	parents[pair+1] = int32(sibling)
	linkChildren(tree, pair)
	proxy.Node = int32(pair + 1)

	nodes[sibling] = makeInternalNode(nodes, pair)
	if !isNodeOrdered(nodes, sibling) || !isNodeOrdered(nodes, pair) {
		tree.dfsOrdered = false
	}

	index := sibling
	for index != nullIndex {
		if shouldRotate {
			rotateNodes(tree, index)
		}

		nodes[index] = makeInternalNode(nodes, getLeftChild(&nodes[index]))
		index = int(parents[index])
	}
}

func removeLeaf(tree *DynamicTree, proxyID int) {
	nodes, parents := tree.nodes, tree.parents
	leaf := int(tree.proxies[proxyID].Node)
	core.Assert(0 <= leaf && leaf < int(tree.nodeEnd), "0 <= leaf && leaf < tree->nodeEnd")
	core.Assert(
		isLeaf(&nodes[leaf]) && getProxyID(&nodes[leaf]) == proxyID,
		"b2IsLeaf( nodes + leaf ) && b2GetProxyId( nodes + leaf ) == proxyId",
	)

	if leaf == rootNode {
		nodes[rootNode] = makeEmptyNode()
		return
	}

	parent := int(parents[leaf])
	nodes[parent] = nodes[leaf^1]
	linkChildren(tree, parent)
	freePair(tree, leaf & ^1)

	index := int(parents[parent])
	for index != nullIndex {
		nodes[index] = makeInternalNode(nodes, getLeftChild(&nodes[index]))
		index = int(parents[index])
	}
}

func createTreeProxyInternal(
	tree *DynamicTree,
	aabb AABB,
	categoryBits uint64, userData uint64,
	markMoved bool,
) int {
	core.Validate(aabb.IsValid(), "b2IsValidAABB( aabb )")

	proxyID := allocateProxy(tree)
	proxy := &tree.proxies[proxyID]
	proxy.CategoryBits = categoryBits
	proxy.UserData = userData

	insertLeaf(tree, aabb, proxyID, markMoved, true)

	return proxyID
}

// Create a proxy. Provide an AABB and a userData value.
func (tree *DynamicTree) CreateProxy(aabb AABB, categoryBits uint64, userData uint64) int {
	return createTreeProxyInternal(tree, aabb, categoryBits, userData, false)
}

// Destroy a proxy. This asserts if the id is invalid.
func (tree *DynamicTree) DestroyProxy(proxyID int) {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	removeLeaf(tree, proxyID)
	freeProxy(tree, proxyID)
}

// Get the number of proxies created.
func (tree *DynamicTree) ProxyCount() int {
	return int(tree.proxyCount)
}

// Move a proxy to a new AABB by removing and reinserting into the tree.
func (tree *DynamicTree) MoveProxy(proxyID int, aabb AABB) {
	tree.moveProxyInternal(proxyID, aabb, false)
}

// Modify the category bits on a proxy.
func (tree *DynamicTree) SetCategoryBits(proxyID int, categoryBits uint64) {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	tree.proxies[proxyID].CategoryBits = categoryBits
}

// Get the category bits on a proxy.
func (tree *DynamicTree) CategoryBits(proxyID int) uint64 {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)

	return tree.proxies[proxyID].CategoryBits
}

// Get the height of the binary tree.
func (tree *DynamicTree) Height() int {
	return getNodeHeight(&tree.nodes[rootNode])
}

// Get the ratio of the sum of the internal node areas to the root area.
func (tree *DynamicTree) AreaRatio() float32 {
	if tree.proxyCount == 0 {
		return 0
	}

	nodes := tree.nodes

	rootArea := perimeter(nodes[rootNode].AABB)
	if rootArea <= 0 {
		return 0
	}

	internalArea := float32(0)

	for i := 2; i < int(tree.nodeEnd); i++ {
		if !isLeaf(&nodes[i]) {
			internalArea += perimeter(nodes[i].AABB)
		}
	}

	return internalArea / rootArea
}

// Get the bounding box that contains the entire tree.
func (tree *DynamicTree) RootBounds() AABB {
	if tree.proxyCount == 0 {
		return AABB{}
	}

	return tree.nodes[rootNode].AABB
}

func validateSubtree(tree *DynamicTree, nodeIndex int, leafCount *int) int {
	core.Assert(
		0 <= nodeIndex && nodeIndex < int(tree.nodeEnd),
		"0 <= nodeIndex && nodeIndex < tree->nodeEnd",
	)
	node := &tree.nodes[nodeIndex]
	core.Assert(!isEmptyNode(node), "b2IsEmptyNode( node ) == false")

	if isLeaf(node) {
		proxyID := getProxyID(node)
		core.Assert(
			0 <= proxyID && proxyID < int(tree.proxyCapacity),
			"0 <= proxyId && proxyId < tree->proxyCapacity",
		)
		core.Assert(
			tree.proxies[proxyID].Node == int32(nodeIndex),
			"tree->proxies[proxyId].node == nodeIndex",
		)
		core.Assert(
			int32(tree.proxies[proxyID].UserData) == node.ShapeIndex,
			"(int32_t)tree->proxies[proxyId].userData == node->shapeIndex",
		)

		*leafCount += 1

		return 0
	}

	pair := getLeftChild(node)
	core.Assert(
		pair&1 == 0 && 2 <= pair && pair < int(tree.nodeEnd),
		"( pair & 1 ) == 0 && 2 <= pair && pair < tree->nodeEnd",
	)
	core.Assert(tree.parents[pair] == int32(nodeIndex), "tree->parents[pair] == nodeIndex")
	core.Assert(tree.parents[pair+1] == int32(nodeIndex), "tree->parents[pair + 1] == nodeIndex")
	core.Assert(
		!tree.dfsOrdered || nodeIndex < pair,
		"tree->dfsOrdered == false || nodeIndex < pair",
	)
	c1, c2 := &tree.nodes[pair], &tree.nodes[pair+1]
	core.Assert(node.AABB.Contains(c1.AABB), "b2AABB_Contains( node->aabb, c1->aabb )")
	core.Assert(node.AABB.Contains(c2.AABB), "b2AABB_Contains( node->aabb, c2->aabb )")
	core.Assert(
		isNodeMoved(node) == (isNodeMoved(c1) || isNodeMoved(c2)),
		"b2IsNodeMoved( node ) == ( b2IsNodeMoved( c1 ) || b2IsNodeMoved( c2 ) )",
	)

	height1 := validateSubtree(tree, pair, leafCount)
	height2 := validateSubtree(tree, pair+1, leafCount)
	height := 1 + max(height1, height2)
	core.Assert(int(node.Height) == height, "node->height == height")

	return height
}

// Validate the tree.
func (tree *DynamicTree) Validate() {
	if !core.ValidationEnabled {
		return
	}

	core.Assert(
		2 <= tree.nodeEnd && tree.nodeEnd <= tree.nodeCapacity,
		"2 <= tree->nodeEnd && tree->nodeEnd <= tree->nodeCapacity",
	)
	core.Assert(tree.nodeEnd&1 == 0, "( tree->nodeEnd & 1 ) == 0")
	core.Assert(
		tree.parents[rootNode] == int32(nullIndex),
		"tree->parents[B2_ROOT_NODE] == B2_NULL_INDEX",
	)
	core.Assert(
		isEmptyNode(&tree.nodes[rootNode+1]),
		"b2IsEmptyNode( tree->nodes + B2_ROOT_NODE + 1 )",
	)

	freePairCount := 0

	pair := tree.pairFreeList
	for pair != int32(nullIndex) {
		core.Assert(
			pair&1 == 0 && 2 <= pair && pair < tree.nodeEnd,
			"( pair & 1 ) == 0 && 2 <= pair && pair < tree->nodeEnd",
		)
		core.Assert(isEmptyNode(&tree.nodes[pair]), "b2IsEmptyNode( tree->nodes + pair )")
		core.Assert(isEmptyNode(&tree.nodes[pair+1]), "b2IsEmptyNode( tree->nodes + pair + 1 )")
		pair = tree.parents[pair]
		freePairCount++
		core.Assert(2*freePairCount < int(tree.nodeEnd), "2 * freePairCount < tree->nodeEnd")
	}

	freeProxyCount := 0

	freeIndex := tree.proxyFreeList
	for freeIndex != int32(nullIndex) {
		core.Assert(
			0 <= freeIndex && freeIndex < tree.proxyCapacity,
			"0 <= freeIndex && freeIndex < tree->proxyCapacity",
		)
		core.Assert(
			tree.proxies[freeIndex].Node == int32(nullIndex),
			"tree->proxies[freeIndex].node == B2_NULL_INDEX",
		)
		freeIndex = tree.proxies[freeIndex].Next
		freeProxyCount++
	}

	core.Assert(
		int(tree.proxyCount)+freeProxyCount == int(tree.proxyCapacity),
		"tree->proxyCount + freeProxyCount == tree->proxyCapacity",
	)
	core.Assert(
		int(tree.nodeEnd) == 2*max(int(tree.proxyCount), 1)+2*freePairCount,
		"tree->nodeEnd == 2 * b2MaxInt( tree->proxyCount, 1 ) + 2 * freePairCount",
	)

	if tree.proxyCount == 0 {
		core.Assert(
			isEmptyNode(&tree.nodes[rootNode]),
			"b2IsEmptyNode( tree->nodes + B2_ROOT_NODE )",
		)

		return
	}

	leafCount := 0
	validateSubtree(tree, rootNode, &leafCount)
	core.Assert(leafCount == int(tree.proxyCount), "leafCount == tree->proxyCount")
}

// Get the number of bytes used by this tree.
func (tree *DynamicTree) ByteCount() int {
	size := int(unsafe.Sizeof(*tree))
	size += cap(tree.nodes) * int(unsafe.Sizeof(TreeNode{}))
	size += cap(tree.parents) * int(unsafe.Sizeof(int32(0)))
	size += cap(tree.proxies) * int(unsafe.Sizeof(TreeProxy{}))
	size += cap(tree.swapNodes) * int(unsafe.Sizeof(TreeNode{}))
	size += cap(tree.leafIndices) * int(unsafe.Sizeof(int32(0)))
	size += cap(tree.leafNodes) * int(unsafe.Sizeof(TreeNode{}))
	size += cap(tree.leafBoxes) * int(unsafe.Sizeof(AABB{}))
	size += cap(tree.leafCenters) * int(unsafe.Sizeof(Vec2{}))
	size += cap(tree.binIndices) * int(unsafe.Sizeof(int32(0)))

	return size
}

// Get proxy user data.
func (tree *DynamicTree) UserData(proxyID int) uint64 {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)

	return tree.proxies[proxyID].UserData
}

// Get the AABB of a proxy.
func (tree *DynamicTree) AABB(proxyID int) AABB {
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	nodeIndex := int(tree.proxies[proxyID].Node)
	core.Assert(
		0 <= nodeIndex && nodeIndex < int(tree.nodeEnd),
		"0 <= nodeIndex && nodeIndex < tree->nodeEnd",
	)

	return tree.nodes[nodeIndex].AABB
}

// Query an AABB for overlapping proxies. The callback class is called for each proxy that overlaps the supplied AABB.
//
//	@return performance data
func (tree *DynamicTree) Query(
	aabb AABB,
	maskBits uint64,
	callback TreeQueryCallbackFunc,
) TreeStats {
	result := TreeStats{}
	if tree.proxyCount == 0 {
		return result
	}

	nodes := tree.nodes

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getRootPair(nodes)
	boxv := loadAABBV(&aabb)

	for stackCount > 0 {
		stackCount--
		pair := stack[stackCount]
		result.NodeVisits++

		for i := range 2 {
			node := &nodes[pair+i]
			if !overlapNode(boxv, node) {
				continue
			}

			if !isLeaf(node) {
				if stackCount < treeStackSize-1 {
					stack[stackCount] = getLeftChild(node)
					stackCount++
				} else {
					core.Assert(stackCount < treeStackSize-1, "stackCount < B2_TREE_STACK_SIZE - 1")
				}

				continue
			}

			proxyID := getProxyID(node)

			proxy := &tree.proxies[proxyID]
			if proxy.CategoryBits&maskBits == 0 {
				continue
			}

			proceed := callback(proxyID, proxy.UserData)

			result.LeafVisits++
			if !proceed {
				return result
			}
		}
	}

	return result
}

// Query an AABB for overlapping proxies. The callback class is called for each proxy that overlaps the supplied AABB.
// No filtering is performed.
//
//	@return performance data
func (tree *DynamicTree) QueryAll(aabb AABB, callback TreeQueryCallbackFunc) TreeStats {
	result := TreeStats{}
	if tree.proxyCount == 0 {
		return result
	}

	nodes := tree.nodes

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getRootPair(nodes)
	boxv := loadAABBV(&aabb)

	for stackCount > 0 {
		stackCount--
		pair := stack[stackCount]
		result.NodeVisits++

		for i := range 2 {
			node := &nodes[pair+i]
			if !overlapNode(boxv, node) {
				continue
			}

			if !isLeaf(node) {
				if stackCount < treeStackSize-1 {
					stack[stackCount] = getLeftChild(node)
					stackCount++
				} else {
					core.Assert(stackCount < treeStackSize-1, "stackCount < B2_TREE_STACK_SIZE - 1")
				}

				continue
			}

			proxyID := getProxyID(node)
			proxy := &tree.proxies[proxyID]
			proceed := callback(proxyID, proxy.UserData)

			result.LeafVisits++
			if !proceed {
				return result
			}
		}
	}

	return result
}

// Ray cast against the proxies in the tree. This relies on the callback
// to perform a exact ray cast in the case were the proxy contains a shape.
// The callback also performs the any collision filtering. This has performance
// roughly equal to k * log(n), where k is the number of collisions and n is the
// number of proxies in the tree.
// Bit-wise filtering using mask bits can greatly improve performance in some scenarios.
//
//	However, this filtering may be approximate, so the user should still apply filtering to results.
//
// @param input the ray cast input data. The ray extends from p1 to p1 + maxFraction * (p2 - p1)
// @param maskBits mask bit hint: `bool accept = (maskBits & proxy.CategoryBits) != 0;`
// @param callback a callback class that is called for each proxy that is hit by the ray
//
//	@return performance data
func (tree *DynamicTree) CastRay(
	input *RayCastInput,
	maskBits uint64,
	callback TreeRayCastCallbackFunc,
) TreeStats {
	result := TreeStats{}
	if tree.proxyCount == 0 {
		return result
	}

	p1 := input.Origin
	d := input.Translation
	r := d.Normalize()
	v := r.ScalarCross(1)
	absV := v.Abs()
	maxFraction := input.MaxFraction
	p2 := p1.MulAdd(maxFraction, d)
	segmentAABB := AABB{LowerBound: p1.Min(p2), UpperBound: p1.Max(p2)}
	boxv := loadAABBV(&segmentAABB)
	nodes := tree.nodes

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getRootPair(nodes)
	subInput := *input

	for stackCount > 0 {
		stackCount--
		pair := stack[stackCount]
		result.NodeVisits++

		var (
			hit  [2]*TreeNode
			leaf [2]bool
		)

		hitCount := 0

		for i := range 2 {
			node := &nodes[pair+i]
			if !overlapNode(boxv, node) {
				continue
			}

			nodeAABB := node.AABB
			c := nodeAABB.Center()
			h := nodeAABB.Extents()
			term1 := absFloat(v.Dot(p1.Sub(c)))

			term2 := absV.Dot(h)
			if term2 < term1 {
				continue
			}

			leaf[hitCount] = isLeaf(node)
			hit[hitCount] = node
			hitCount++
		}

		orderCastHits(&hit, &leaf, hitCount, p1)

		for i := range hitCount {
			if !leaf[i] {
				if stackCount < treeStackSize-1 {
					stack[stackCount] = getLeftChild(hit[i])
					stackCount++
				} else {
					core.Assert(stackCount < treeStackSize-1, "stackCount < B2_TREE_STACK_SIZE - 1")
				}

				continue
			}

			proxyID := getProxyID(hit[i])

			proxy := &tree.proxies[proxyID]
			if proxy.CategoryBits&maskBits == 0 {
				continue
			}

			subInput.MaxFraction = maxFraction
			value := callback(&subInput, proxyID, proxy.UserData)

			result.LeafVisits++
			if value == 0 {
				return result
			}

			if 0 < value && value <= maxFraction {
				maxFraction = value
				p2 = p1.MulAdd(maxFraction, d)
				segmentAABB.LowerBound = p1.Min(p2)
				segmentAABB.UpperBound = p1.Max(p2)
				boxv = loadAABBV(&segmentAABB)
			}
		}
	}

	return result
}

// Cast a swept AABB through the tree. This has performance roughly equal to k * log(n),
// where k is the number of collisions and n is the number of proxies in the tree.
// @param input the AABB cast input. The box sweeps from its origin to origin + maxFraction * translation.
// @param maskBits filter bits: `bool accept = (maskBits & proxy.CategoryBits) != 0;`
// @param callback a callback that is called for each proxy the swept box may hit
//
//	@return performance data
func (tree *DynamicTree) CastBox(
	input *BoxCastInput,
	maskBits uint64,
	callback TreeBoxCastCallbackFunc,
) TreeStats {
	result := TreeStats{}
	if tree.proxyCount == 0 {
		return result
	}

	originAABB := input.Box
	p1 := originAABB.Center()
	extension := originAABB.Extents()
	r := input.Translation
	v := r.ScalarCross(1)
	absV := v.Abs()
	maxFraction := input.MaxFraction
	t := Vec2{
		X: float32(maxFraction * input.Translation.X),
		Y: float32(maxFraction * input.Translation.Y),
	}
	totalAABB := AABB{
		LowerBound: originAABB.LowerBound.Min(originAABB.LowerBound.Add(t)),
		UpperBound: originAABB.UpperBound.Max(originAABB.UpperBound.Add(t)),
	}
	boxv := loadAABBV(&totalAABB)
	nodes := tree.nodes

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getRootPair(nodes)
	subInput := *input

	for stackCount > 0 {
		stackCount--
		pair := stack[stackCount]
		result.NodeVisits++

		var (
			hit  [2]*TreeNode
			leaf [2]bool
		)

		hitCount := 0

		for i := range 2 {
			node := &nodes[pair+i]
			if !overlapNode(boxv, node) {
				continue
			}

			nodeAABB := node.AABB
			c := nodeAABB.Center()
			h := nodeAABB.Extents().Add(extension)
			term1 := absFloat(v.Dot(p1.Sub(c)))

			term2 := absV.Dot(h)
			if term2 < term1 {
				continue
			}

			leaf[hitCount] = isLeaf(node)
			hit[hitCount] = node
			hitCount++
		}

		orderCastHits(&hit, &leaf, hitCount, p1)

		for i := range hitCount {
			if !leaf[i] {
				if stackCount < treeStackSize-1 {
					stack[stackCount] = getLeftChild(hit[i])
					stackCount++
				} else {
					core.Assert(stackCount < treeStackSize-1, "stackCount < B2_TREE_STACK_SIZE - 1")
				}

				continue
			}

			proxyID := getProxyID(hit[i])

			proxy := &tree.proxies[proxyID]
			if proxy.CategoryBits&maskBits == 0 {
				continue
			}

			subInput.MaxFraction = maxFraction
			value := callback(&subInput, proxyID, proxy.UserData)

			result.LeafVisits++
			if value == 0 {
				return result
			}

			if 0 < value && value < maxFraction {
				maxFraction = value
				t = Vec2{
					X: float32(maxFraction * input.Translation.X),
					Y: float32(maxFraction * input.Translation.Y),
				}
				totalAABB.LowerBound = originAABB.LowerBound.Min(originAABB.LowerBound.Add(t))
				totalAABB.UpperBound = originAABB.UpperBound.Max(originAABB.UpperBound.Add(t))
				boxv = loadAABBV(&totalAABB)
			}
		}
	}

	return result
}

func (tree *DynamicTree) moveProxyInternal(proxyID int, aabb AABB, markMoved bool) {
	core.Validate(aabb.IsValid(), "b2IsValidAABB( aabb )")
	core.Validate(
		aabb.UpperBound.X-aabb.LowerBound.X < Huge(),
		"aabb.upperBound.x - aabb.lowerBound.x < B2_HUGE",
	)
	core.Validate(
		aabb.UpperBound.Y-aabb.LowerBound.Y < Huge(),
		"aabb.upperBound.y - aabb.lowerBound.y < B2_HUGE",
	)
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	removeLeaf(tree, proxyID)
	insertLeaf(tree, aabb, proxyID, markMoved, false)
}

func (tree *DynamicTree) enlargeProxy(proxyID int, aabb AABB) {
	core.Validate(aabb.IsValid(), "b2IsValidAABB( aabb )")
	core.Validate(
		aabb.UpperBound.X-aabb.LowerBound.X < Huge(),
		"aabb.upperBound.x - aabb.lowerBound.x < B2_HUGE",
	)
	core.Validate(
		aabb.UpperBound.Y-aabb.LowerBound.Y < Huge(),
		"aabb.upperBound.y - aabb.lowerBound.y < B2_HUGE",
	)
	core.Assert(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	nodes, parents := tree.nodes, tree.parents
	index := int(tree.proxies[proxyID].Node)
	node := &nodes[index]
	core.Assert(isLeaf(node), "b2IsLeaf( node )")
	core.Validate(!node.AABB.Contains(aabb), "b2AABB_Contains( node->aabb, aabb ) == false")
	node.AABB = aabb
	node.FlagIndex |= movedNode

	index = int(parents[index])
	for index != nullIndex {
		node = &nodes[index]
		changed := enlargeAABB(&node.AABB, aabb)
		node.FlagIndex |= movedNode
		index = int(parents[index])

		if !changed {
			break
		}
	}

	for index != nullIndex {
		node = &nodes[index]
		if node.FlagIndex&movedNode != 0 {
			break
		}

		node.FlagIndex |= movedNode
		index = int(parents[index])
	}
}

func (tree *DynamicTree) validateNoMoved() {
	if !core.ValidationEnabled {
		return
	}

	for i := range tree.nodeEnd {
		core.Assert(!isNodeMoved(&tree.nodes[i]), "b2IsNodeMoved( nodes + i ) == false")
	}
}

func orderCastHits(hit *[2]*TreeNode, leaf *[2]bool, hitCount int, origin Vec2) {
	if hitCount == 2 {
		orderCastPair(hit, leaf, origin)
	}
}

func orderCastPair(hit *[2]*TreeNode, leaf *[2]bool, origin Vec2) {
	if !leaf[0] && !leaf[1] {
		center1 := hit[0].AABB.Center()
		center2 := hit[1].AABB.Center()
		d1 := center1.DistanceSquared(origin)

		d2 := center2.DistanceSquared(origin)
		if d1 < d2 {
			hit[0], hit[1] = hit[1], hit[0]
		}
	}
}

func partitionMid(indices []int32, centers []Vec2, count int) int {
	if count <= 2 {
		return count / 2
	}

	lowerBound, upperBound := centers[0], centers[0]
	for i := 1; i < count; i++ {
		lowerBound = lowerBound.Min(centers[i])
		upperBound = upperBound.Max(centers[i])
	}

	d := upperBound.Sub(lowerBound)
	c := Vec2{X: 0.5 * (lowerBound.X + upperBound.X), Y: 0.5 * (lowerBound.Y + upperBound.Y)}

	var split int
	if d.X > d.Y {
		split = partitionCentersX(indices, centers, count, c.X)
	} else {
		split = partitionCentersY(indices, centers, count, c.Y)
	}

	return split
}

func partitionCentersX(indices []int32, centers []Vec2, count int, pivot float32) int {
	i1, i2 := 0, count
	for i1 < i2 {
		for i1 < i2 && centers[i1].X < pivot {
			i1++
		}

		for i1 < i2 && centers[i2-1].X >= pivot {
			i2--
		}

		if i1 < i2 {
			indices[i1], indices[i2-1] = indices[i2-1], indices[i1]
			centers[i1], centers[i2-1] = centers[i2-1], centers[i1]
			i1++
			i2--
		}
	}

	core.Assert(i1 == i2, "i1 == i2")

	if i1 > 0 && i1 < count {
		return i1
	}

	return count / 2
}

func partitionCentersY(indices []int32, centers []Vec2, count int, pivot float32) int {
	i1, i2 := 0, count
	for i1 < i2 {
		for i1 < i2 && centers[i1].Y < pivot {
			i1++
		}

		for i1 < i2 && centers[i2-1].Y >= pivot {
			i2--
		}

		if i1 < i2 {
			indices[i1], indices[i2-1] = indices[i2-1], indices[i1]
			centers[i1], centers[i2-1] = centers[i2-1], centers[i1]
			i1++
			i2--
		}
	}

	core.Assert(i1 == i2, "i1 == i2")

	if i1 > 0 && i1 < count {
		return i1
	}

	return count / 2
}

func bumpPair(tree *DynamicTree, parent int) int {
	pair := int(tree.nodeEnd)
	core.Assert(pair+2 <= int(tree.nodeCapacity), "pair + 2 <= tree->nodeCapacity")
	tree.nodeEnd += 2
	tree.parents[pair] = int32(parent)
	tree.parents[pair+1] = int32(parent)

	return pair
}

func setLeftChild(node *TreeNode, pair int) {
	node.FlagIndex = (node.FlagIndex & ^nodeIndexMask) | uint32(pair)
}

func copySubtree(tree *DynamicTree, node TreeNode, newIndex int) {
	oldNodes, newNodes, proxies := tree.nodes, tree.swapNodes, tree.proxies
	if isLeaf(&node) {
		newNodes[newIndex] = node
		proxies[getProxyID(&node)].Node = int32(newIndex)

		return
	}

	var stack [treeStackSize]copyItem

	stackCount := 0
	oldPair := getLeftChild(&node)
	newPair := bumpPair(tree, newIndex)
	setLeftChild(&node, newPair)

	newNodes[newIndex] = node
	for {
		pair := newNodes[newPair:]
		pair[0] = oldNodes[oldPair]

		pair[1] = oldNodes[oldPair+1]
		if isLeaf(&pair[1]) {
			proxyID := getProxyID(&pair[1])
			proxies[proxyID].Node = int32(newPair + 1)
		} else {
			core.Assert(stackCount < treeStackSize, "stackCount < B2_TREE_STACK_SIZE")

			leftChild := getLeftChild(&pair[1])
			stack[stackCount] = copyItem{oldPair: leftChild, newIndex: newPair + 1}
			stackCount++
		}

		if !isLeaf(&pair[0]) {
			leftIndex := newPair
			oldPair = getLeftChild(&pair[0])
			newPair = bumpPair(tree, leftIndex)
			setLeftChild(&newNodes[leftIndex], newPair)

			continue
		}

		proxies[getProxyID(&pair[0])].Node = int32(newPair)

		if stackCount == 0 {
			break
		}

		stackCount--
		item := stack[stackCount]
		oldPair = item.oldPair
		newPair = bumpPair(tree, item.newIndex)
		setLeftChild(&newNodes[item.newIndex], newPair)
	}
}

func placeLeaf(tree *DynamicTree, leaf TreeNode, newIndex int) {
	if isLeaf(&leaf) {
		tree.swapNodes[newIndex] = leaf
		tree.proxies[getProxyID(&leaf)].Node = int32(newIndex)
	} else {
		copySubtree(tree, leaf, newIndex)
	}
}

func buildTree(tree *DynamicTree, leafCount int) {
	nodes, leaves, leafIndices, leafCenters := tree.swapNodes, tree.leafNodes, tree.leafIndices, tree.leafCenters
	tree.nodeEnd = 2
	nodes[rootNode+1] = makeEmptyNode()
	tree.parents[rootNode] = int32(nullIndex)

	tree.parents[rootNode+1] = int32(nullIndex)
	if leafCount == 1 {
		placeLeaf(tree, leaves[leafIndices[0]], rootNode)
		return
	}

	var stack [treeStackSize]rebuildItem

	top := 0

	stack[0] = rebuildItem{
		nodeIndex:  rootNode,
		pair:       bumpPair(tree, rootNode),
		childCount: -1,
		startIndex: 0,
		endIndex:   leafCount,
		splitIndex: partitionMid(leafIndices, leafCenters, leafCount),
	}
	for {
		item := &stack[top]

		item.childCount++
		if item.childCount == 2 {
			nodes[item.nodeIndex] = makeInternalNode(nodes, item.pair)

			if top == 0 {
				break
			}

			top--

			continue
		}

		slot := item.childCount

		startIndex, endIndex := item.splitIndex, item.endIndex
		if slot == 0 {
			startIndex = item.startIndex
			endIndex = item.splitIndex
		}

		count := endIndex - startIndex
		core.Assert(count > 0, "count > 0")

		nodeIndex := item.pair + slot
		if count == 1 {
			placeLeaf(tree, leaves[leafIndices[startIndex]], nodeIndex)
			continue
		}

		core.Assert(top < treeStackSize-1, "top < B2_TREE_STACK_SIZE - 1")
		top++
		newItem := &stack[top]
		newItem.nodeIndex = nodeIndex
		newItem.pair = bumpPair(tree, nodeIndex)
		newItem.childCount = -1
		newItem.startIndex = startIndex
		newItem.endIndex = endIndex
		newItem.splitIndex = startIndex
		newItem.splitIndex += partitionMid(
			leafIndices[startIndex:],
			leafCenters[startIndex:],
			count,
		)
	}
}

// Rebuild the tree while retaining subtrees that haven't changed. Returns the number of boxes sorted.
func (tree *DynamicTree) rebuild(fullBuild bool) int {
	proxyCount := int(tree.proxyCount)
	if proxyCount == 0 {
		return 0
	}

	nodes := tree.nodes

	root := &nodes[rootNode]
	if !fullBuild && !isNodeMoved(root) && tree.dfsOrdered {
		return 0
	}

	if tree.swapNodes == nil {
		tree.swapNodes = make([]TreeNode, tree.nodeCapacity)
	}

	if proxyCount > int(tree.rebuildCapacity) {
		newCapacity := proxyCount + proxyCount/2
		tree.leafIndices = make([]int32, newCapacity)
		tree.leafNodes = make([]TreeNode, newCapacity)
		tree.leafCenters = make([]Vec2, newCapacity)
		tree.rebuildCapacity = int32(newCapacity)
	}

	leafCount := tree.collectRebuildLeaves(fullBuild)

	core.Assert(
		0 < leafCount && leafCount <= proxyCount,
		"0 < leafCount && leafCount <= proxyCount",
	)
	buildTree(tree, leafCount)
	tree.nodes, tree.swapNodes = tree.swapNodes, tree.nodes
	tree.pairFreeList = int32(nullIndex)
	tree.dfsOrdered = true
	tree.Validate()
	tree.validateNoMoved()

	return leafCount
}

func (tree *DynamicTree) collectRebuildLeaves(fullBuild bool) int {
	nodes := tree.nodes
	root := &nodes[rootNode]
	leafIndices, leaves, leafCenters := tree.leafIndices, tree.leafNodes, tree.leafCenters
	leafCount := 0

	var stack [treeStackSize]int

	stackCount := 0
	if !isLeaf(root) && (fullBuild || isNodeMoved(root)) {
		stack[stackCount] = getLeftChild(root)
		stackCount++
	} else {
		node := *root
		node.FlagIndex &= ^movedNode
		leafIndices[0] = 0
		leaves[0] = node
		leafCenters[0] = node.AABB.Center()
		leafCount = 1
	}

	for stackCount > 0 {
		stackCount--

		pair := stack[stackCount]
		for i := range 2 {
			node := nodes[pair+i]
			if !isLeaf(&node) && (fullBuild || isNodeMoved(&node)) {
				core.Assert(stackCount < treeStackSize, "stackCount < B2_TREE_STACK_SIZE")
				stack[stackCount] = getLeftChild(&node)
				stackCount++

				continue
			}

			node.FlagIndex &= ^movedNode
			leafIndices[leafCount] = int32(leafCount)
			leaves[leafCount] = node
			leafCenters[leafCount] = node.AABB.Center()
			leafCount++
		}
	}

	return leafCount
}

func (tree *DynamicTree) markProxyMovedSerial(proxyID int) {
	core.Validate(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	nodes, parents := tree.nodes, tree.parents
	index := int(tree.proxies[proxyID].Node)
	core.Validate(0 <= index && index < int(tree.nodeEnd), "0 <= index && index < tree->nodeEnd")
	core.Validate(isLeaf(&nodes[index]), "b2IsLeaf( nodes + index )")

	for index != nullIndex {
		nodes[index].FlagIndex |= movedNode
		index = int(parents[index])
	}
}

func (tree *DynamicTree) markProxyMoved(proxyID int, aabb AABB) {
	core.Validate(
		0 <= proxyID && proxyID < int(tree.proxyCapacity),
		"0 <= proxyId && proxyId < tree->proxyCapacity",
	)
	nodes, parents := tree.nodes, tree.parents
	index := int(tree.proxies[proxyID].Node)
	core.Validate(0 <= index && index < int(tree.nodeEnd), "0 <= index && index < tree->nodeEnd")
	node := &nodes[index]
	core.Validate(isLeaf(node), "b2IsLeaf( node )")
	core.Validate(!node.AABB.Contains(aabb), "b2AABB_Contains( node->aabb, aabb ) == false")
	node.AABB = aabb
	node.FlagIndex |= movedNode

	index = int(parents[index])
	for index != nullIndex {
		node = &nodes[index]
		if atomic.LoadUint32(&node.FlagIndex)&movedNode != 0 {
			break
		}

		previousFlags := atomic.OrUint32(&node.FlagIndex, movedNode)
		if previousFlags&movedNode != 0 {
			break
		}

		index = int(parents[index])
	}
}

func (tree *DynamicTree) clearMoved() {
	nodes := tree.nodes

	root := &nodes[rootNode]
	if !isNodeMoved(root) {
		return
	}

	root.FlagIndex &= ^movedNode
	if isLeaf(root) {
		return
	}

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getLeftChild(root)

	for stackCount > 0 {
		stackCount--

		pair := stack[stackCount]
		for i := range 2 {
			node := &nodes[pair+i]
			if node.FlagIndex&movedNode != 0 {
				node.FlagIndex &= ^movedNode
				if !isLeaf(node) {
					if stackCount < treeStackSize {
						stack[stackCount] = getLeftChild(node)
						stackCount++
					} else {
						core.Assert(stackCount < treeStackSize, "stackCount < B2_TREE_STACK_SIZE")
					}
				}
			}
		}
	}
}

func (tree *DynamicTree) gatherMovedProxies(proxyIDs []int) int {
	nodes := tree.nodes

	root := &nodes[rootNode]
	if !isNodeMoved(root) {
		return 0
	}

	if isLeaf(root) {
		proxyIDs[0] = getProxyID(root)
		return 1
	}

	count := 0

	var stack [treeStackSize]int

	stackCount := 1
	stack[0] = getLeftChild(root)

	for stackCount > 0 {
		stackCount--

		pair := stack[stackCount]
		for i := range 2 {
			node := &nodes[pair+i]
			if !isNodeMoved(node) {
				continue
			}

			if isLeaf(node) {
				proxyIDs[count] = getProxyID(node)
				count++
			} else {
				if stackCount < treeStackSize {
					stack[stackCount] = getLeftChild(node)
					stackCount++
				} else {
					core.Assert(stackCount < treeStackSize, "stackCount < B2_TREE_STACK_SIZE")
				}
			}
		}
	}

	return count
}

func refitSubtree(nodes []TreeNode, rootIndex int) {
	root := &nodes[rootIndex]
	if isLeaf(root) || !isNodeMoved(root) {
		return
	}

	var stack [treeStackSize]int

	stackCount := 3
	stack[0] = ^rootIndex
	stack[1] = getLeftChild(root)
	stack[2] = getLeftChild(root) + 1

	for stackCount > 0 {
		stackCount--

		item := stack[stackCount]
		if item < 0 {
			node := &nodes[^item]
			pair := getLeftChild(node)
			node.AABB = unionV(nodes[pair].AABB, nodes[pair+1].AABB)

			continue
		}

		node := &nodes[item]
		if isLeaf(node) || !isNodeMoved(node) {
			continue
		}

		if stackCount+3 <= treeStackSize {
			pair := getLeftChild(node)
			stack[stackCount] = ^item
			stack[stackCount+1] = pair
			stack[stackCount+2] = pair + 1
			stackCount += 3
		} else {
			core.Assert(stackCount+3 <= treeStackSize, "stackCount + 3 <= B2_TREE_STACK_SIZE")
		}
	}
}

func (tree *DynamicTree) refit() {
	if !hasTreeMoved(tree) {
		return
	}

	nodes := tree.nodes
	if !tree.dfsOrdered {
		refitSubtree(nodes, rootNode)
		return
	}

	for pair := int(tree.nodeEnd) - 2; pair >= 0; pair -= 2 {
		node := nodes[pair:]

		flags1, flags2 := node[0].FlagIndex, node[1].FlagIndex
		if (flags1|flags2)&movedNode == 0 {
			continue
		}

		refit1 := flags1&(leafNode|movedNode) == movedNode
		refit2 := flags2&(leafNode|movedNode) == movedNode

		children1, children2 := pair, pair
		if refit1 {
			children1 = int(flags1 & nodeIndexMask)
		}

		if refit2 {
			children2 = int(flags2 & nodeIndexMask)
		}

		storeAABBV(&node[0].AABB, unionPairV(nodes[children1:]), refit1)
		storeAABBV(&node[1].AABB, unionPairV(nodes[children2:]), refit2)
	}
}

func isLeaf(node *TreeNode) bool {
	return node.FlagIndex&leafNode == leafNode
}

func isNodeMoved(node *TreeNode) bool {
	return node.FlagIndex&movedNode == movedNode
}

func isEmptyNode(node *TreeNode) bool {
	return node.FlagIndex == emptyNode
}

func getLeftChild(node *TreeNode) int {
	return int(node.FlagIndex & nodeIndexMask)
}

func getProxyID(node *TreeNode) int {
	return int(node.FlagIndex & nodeIndexMask)
}

func getRootPair(nodes []TreeNode) int {
	root := &nodes[rootNode]
	if isLeaf(root) {
		return rootNode
	}

	return getLeftChild(root)
}

func makeEmptyNode() TreeNode {
	inf := float32(math.Inf(1))

	return TreeNode{
		AABB:      AABB{LowerBound: Vec2{X: inf, Y: inf}, UpperBound: Vec2{X: -inf, Y: -inf}},
		FlagIndex: emptyNode,
	}
}

func hasTreeMoved(tree *DynamicTree) bool {
	return isNodeMoved(&tree.nodes[rootNode])
}

func needsRebuild(tree *DynamicTree) bool {
	return isNodeMoved(&tree.nodes[rootNode]) || !tree.dfsOrdered
}
