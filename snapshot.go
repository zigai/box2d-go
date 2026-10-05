// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"errors"
	"math"
	"unsafe"
)

const (
	snapMagic               uint32 = 0x32534E42
	snapVersion             uint32 = 12
	snapFlagValidation      uint32 = 0x1
	snapFlagDoublePrecision uint32 = 0x2
)

var (
	ErrBadMagic = errors.New("box2d: invalid magic")
	ErrVersion  = errors.New("box2d: version mismatch")
	ErrLayout   = errors.New("box2d: layout mismatch")
	ErrCorrupt  = errors.New("box2d: corrupt data")
	ErrDiverged = errors.New("box2d: replay diverged")
)

type snapReader struct {
	data   []byte
	cursor int
	size   int
	ok     bool
}

type SnapshotError struct {
	Err error
}

// Create a new world from a snapshot image. Use this to load a saved world when there is
// no existing world to restore into. Handles held from the origin world do not reference
// the new world. Host wiring is reset to defaults.
// A rejected image returns a SnapshotError and releases the new world's resources.
func NewWorldFromSnapshot(image []byte, workerCount int) (*World, error) {
	if err := snapshotHeaderError(image, len(image)); err != nil {
		return nil, err
	}

	var r snapReader
	openSnapshotImage(image, len(image), &r)

	def := DefaultWorldDef()
	def.WorkerCount = workerCount

	world := NewWorld(&def)
	if !deserializeIntoShell(&r, world) {
		world.Close()
		return nil, &SnapshotError{Err: ErrCorrupt}
	}

	world.preSolveFcn = nil
	world.customFilterFcn = nil
	world.userData = nil

	return world, nil
}

func computeLayoutHash() uint32 {
	h := uint32(3498484001)
	if doublePrecision {
		h = (h ^ 8) * 16777619
	} else {
		h = (h ^ 4) * 16777619
	}

	return h
}

func snapRCheck(r *snapReader, need int) {
	if need < 0 || r.cursor < 0 || r.size < 0 || r.size > len(r.data) || r.cursor > r.size || need > r.size-r.cursor {
		r.ok = false
	}
}

func snapRBytes(r *snapReader, dst []byte, n int) {
	snapRCheck(r, n)

	if n > len(dst) {
		r.ok = false
	}

	if !r.ok {
		return
	}

	copy(dst[:n], r.data[r.cursor:r.cursor+n])
	r.cursor += n
}

func snapRI32(r *snapReader) int {
	return int(int32(snapRU32(r)))
}

func snapRU32(r *snapReader) uint32 {
	var b [4]byte
	snapRBytes(r, b[:], 4)

	return binary.LittleEndian.Uint32(b[:])
}

func snapWI32(buf *recBuffer, v int) {
	assert(int64(v) >= math.MinInt32 && int64(v) <= math.MaxInt32, "int32 range")
	recWI32(buf, int32(v))
}

func snapWU32(buf *recBuffer, v uint32) {
	recWU32(buf, v)
}

func snapCheckCount(r *snapReader, count int, memSize int, minStreamBytes int) bool {
	if count < 0 || memSize < 0 || minStreamBytes < 0 || (memSize > 0 && count > math.MaxInt32/memSize) || r.cursor < 0 || r.cursor > r.size {
		return false
	}

	return minStreamBytes == 0 || count <= (r.size-r.cursor)/minStreamBytes
}

func serIDPool(buf *recBuffer, pool *idPool) {
	snapWI32(buf, pool.nextIndex)
	serPodArray(buf, pool.freeArray)
}

func desIDPool(r *snapReader, pool *idPool) {
	pool.nextIndex = snapRI32(r)
	desPodArray(r, &pool.freeArray)

	if pool.nextIndex < 0 || len(pool.freeArray) > pool.nextIndex {
		r.ok = false
	}

	for _, id := range pool.freeArray {
		if id < 0 || id >= pool.nextIndex {
			r.ok = false
		}
	}
}

func serBitSet(buf *recBuffer, bs *bitSet) {
	snapWU32(buf, bs.blockCount)

	for _, v := range bs.bits[:bs.blockCount] {
		recWU64(buf, v)
	}
}

func desBitSet(r *snapReader, bs *bitSet) {
	n := snapRU32(r)

	destroyBitSet(bs)

	if !r.ok || !snapCheckCount(r, int(n), 8, 8) {
		r.ok = false
		return
	}

	bs.blockCount = n
	bs.blockCapacity = max(n, 1)

	bs.bits = make([]uint64, bs.blockCapacity)
	for i := range n {
		bs.bits[i] = snapReadU64(r)
	}
}

func serHashSet(buf *recBuffer, hs *hashSet) {
	snapWU32(buf, hs.capacity)
	snapWU32(buf, hs.count)

	for _, v := range hs.items[:hs.capacity] {
		recWU64(buf, v.key)
	}
}

func desHashSet(r *snapReader, hs *hashSet) {
	n := snapRU32(r)
	count := snapRU32(r)

	destroySet(hs)

	if !r.ok || !snapCheckCount(r, int(n), 8, 8) || n&(n-1) != 0 || count > n {
		r.ok = false
		return
	}

	hs.capacity = n
	hs.count = count

	hs.items = make([]setItem, n)
	for i := range n {
		hs.items[i].key = snapReadU64(r)
	}
}

func serTree(buf *recBuffer, tree *DynamicTree) {
	snapWI32(buf, int(tree.nodeEnd))
	snapWI32(buf, int(tree.nodeCapacity))
	snapWI32(buf, int(tree.pairFreeList))
	snapWI32(buf, int(tree.proxyCount))
	snapWI32(buf, int(tree.proxyCapacity))
	snapWI32(buf, int(tree.proxyFreeList))

	if tree.dfsOrdered {
		snapWI32(buf, 1)
	} else {
		snapWI32(buf, 0)
	}

	for _, v := range tree.nodes[:tree.nodeEnd] {
		snapWriteTreeNode(buf, v)
	}

	for _, v := range tree.parents[:tree.nodeEnd] {
		recWI32(buf, v)
	}

	for _, v := range tree.proxies[:tree.proxyCapacity] {
		snapWriteTreeProxy(buf, v)
	}
}

func desTree(r *snapReader, tree *DynamicTree) {
	nodeEnd := snapRI32(r)
	nodeCapacity := snapRI32(r)
	pairFreeList := snapRI32(r)
	proxyCount := snapRI32(r)
	proxyCapacity := snapRI32(r)
	proxyFreeList := snapRI32(r)
	ordered := snapRI32(r)

	valid := snapTreeCountsValid(r, nodeEnd, nodeCapacity, proxyCount, proxyCapacity)
	if !r.ok || !valid {
		r.ok = false
		*tree = NewDynamicTree(0)

		return
	}

	nodeCapacity = int(min(int64(nodeCapacity), 2*int64(nodeEnd)))
	tree.nodeEnd = int32(nodeEnd)
	tree.nodeCapacity = int32(nodeCapacity)
	tree.pairFreeList = int32(pairFreeList)
	tree.proxyCount = int32(proxyCount)
	tree.proxyCapacity = int32(proxyCapacity)
	tree.proxyFreeList = int32(proxyFreeList)
	tree.dfsOrdered = ordered != 0
	tree.nodes = make([]TreeNode, nodeCapacity)
	tree.parents = make([]int32, nodeCapacity)

	tree.proxies = make([]TreeProxy, proxyCapacity)
	for i := range nodeEnd {
		snapReadTreeNode(r, &tree.nodes[i])
	}

	for i := range nodeEnd {
		tree.parents[i] = int32(snapRI32(r))
	}

	for i := range proxyCapacity {
		snapReadTreeProxy(r, &tree.proxies[i])
	}
}

func serSolverSet(buf *recBuffer, set *solverSet) {
	snapWI32(buf, set.setIndex)
	serPodArray(buf, set.bodySims)
	serPodArray(buf, set.bodyStates)
	serPodArray(buf, set.jointSims)
	serPodArray(buf, set.contactSims)
	serPodArray(buf, set.islandSims)
}

func desSolverSet(r *snapReader, set *solverSet) {
	set.setIndex = snapRI32(r)
	desPodArray(r, &set.bodySims)
	desPodArray(r, &set.bodyStates)
	desPodArray(r, &set.jointSims)
	desPodArray(r, &set.contactSims)
	desPodArray(r, &set.islandSims)
}

func serGraphColor(buf *recBuffer, color *graphColor, isOverflow bool) {
	if !isOverflow {
		serBitSet(buf, &color.bodySet)
	}

	serPodArray(buf, color.contactSims)
	serPodArray(buf, color.jointSims)
}

func desGraphColor(r *snapReader, color *graphColor, isOverflow bool) {
	if !isOverflow {
		desBitSet(r, &color.bodySet)
	}

	desPodArray(r, &color.contactSims)
	desPodArray(r, &color.jointSims)
	color.wideConstraints = nil
	color.overflowConstraints = nil
	color.wideConstraintCount = 0
}

func serWorldConfig(buf *recBuffer, world *World) {
	snapWriteVec2(buf, world.gravity)
	recWF32(buf, world.hitEventThreshold)
	recWF32(buf, world.restitutionThreshold)
	snapWI32(buf, world.restitutionIterations)
	recWF32(buf, world.maxLinearSpeed)
	recWF32(buf, world.contactSpeed)
	recWF32(buf, world.contactHertz)
	recWF32(buf, world.contactDampingRatio)
	recWF32(buf, world.contactRecycleDistance)
	recWU64(buf, world.stepIndex)
	snapWI32(buf, world.splitIslandID)
	recWF32(buf, world.invH)
	recWF32(buf, world.invDT)
	snapWI32(buf, world.endEventArrayIndex)
	snapWriteCapacity(buf, world.maxCapacity)

	var flags uint8
	if world.enableSleep {
		flags |= 1
	}

	if world.enableWarmStarting {
		flags |= 2
	}

	if world.enableContinuous {
		flags |= 8
	}

	if world.enableRestitutionPropagation {
		flags |= 16
	}

	recWU8(buf, flags)
}

func desWorldConfig(r *snapReader, world *World) {
	snapReadVec2(r, &world.gravity)
	world.hitEventThreshold = math.Float32frombits(snapRU32(r))
	world.restitutionThreshold = math.Float32frombits(snapRU32(r))
	world.restitutionIterations = snapRI32(r)
	world.maxLinearSpeed = math.Float32frombits(snapRU32(r))
	world.contactSpeed = math.Float32frombits(snapRU32(r))
	world.contactHertz = math.Float32frombits(snapRU32(r))
	world.contactDampingRatio = math.Float32frombits(snapRU32(r))
	world.contactRecycleDistance = math.Float32frombits(snapRU32(r))
	world.stepIndex = snapReadU64(r)
	world.splitIslandID = snapRI32(r)
	world.invH = math.Float32frombits(snapRU32(r))
	world.invDT = math.Float32frombits(snapRU32(r))
	world.endEventArrayIndex = snapRI32(r)
	snapReadCapacity(r, &world.maxCapacity)
	flags := snapReadU8(r)
	world.enableSleep = flags&1 != 0
	world.enableWarmStarting = flags&2 != 0
	world.enableContinuous = flags&8 != 0
	world.enableRestitutionPropagation = flags&16 != 0
}

func serializeWorld(world *World, buf *recBuffer) {
	snapWU32(buf, snapMagic)
	snapWU32(buf, snapVersion)
	snapWU32(buf, computeLayoutHash())

	flags := uint32(0)
	if validationEnabled {
		flags |= snapFlagValidation
	}

	if doublePrecision {
		flags |= snapFlagDoublePrecision
	}

	snapWU32(buf, flags)
	serWorldConfig(buf, world)
	serIDPool(buf, &world.bodyIDPool)
	serIDPool(buf, &world.shapeIDPool)
	serIDPool(buf, &world.chainIDPool)
	serIDPool(buf, &world.contactIDPool)
	serIDPool(buf, &world.jointIDPool)
	serIDPool(buf, &world.islandIDPool)
	serIDPool(buf, &world.solverSetIDPool)
	snapWI32(buf, len(world.solverSets))

	for i := range world.solverSets {
		serSolverSet(buf, &world.solverSets[i])
	}

	serUserDataArray(buf, world.bodies, func(v *body) { v.userData = nil })
	serUserDataArray(buf, world.shapes, func(v *shape) { v.userData = nil })
	serPodArray(buf, world.fatAABBs)
	serPodArray(buf, world.contacts)
	serUserDataArray(buf, world.joints, func(v *joint) { v.userData = nil })
	snapWI32(buf, len(world.chainShapes))

	for _, chain := range world.chainShapes {
		snapWI32(buf, chain.id)
		snapWI32(buf, chain.bodyID)
		snapWI32(buf, chain.nextChainID)
		snapWI32(buf, chain.segmentCount)
		recWU16(buf, chain.generation)

		if chain.id != nullIndex {
			for _, id := range chain.shapeIndices[:chain.segmentCount] {
				snapWI32(buf, id)
			}
		}
	}

	snapWI32(buf, len(world.sensors))

	for _, s := range world.sensors {
		snapWI32(buf, s.shapeID)
		serPodArray(buf, s.hits)
		serPodArray(buf, s.overlaps1)
		serPodArray(buf, s.overlaps2)
	}

	snapWI32(buf, len(world.islands))

	for _, island := range world.islands {
		snapWI32(buf, island.setIndex)
		snapWI32(buf, island.localIndex)
		snapWI32(buf, island.islandID)
		snapWI32(buf, island.constraintRemoveCount)
		serPodArray(buf, island.bodies)
		serPodArray(buf, island.contacts)
		serPodArray(buf, island.joints)
	}

	for i := range world.broadPhase.trees {
		serTree(buf, &world.broadPhase.trees[i])
	}

	serHashSet(buf, &world.broadPhase.pairSet)

	for i := range world.constraintGraph.colors {
		serGraphColor(buf, &world.constraintGraph.colors[i], i == overflowIndex)
	}
}

func freeLiveSimElements(world *World) {
	for i := range world.chainShapes {
		world.chainShapes[i].shapeIndices = nil
	}

	for i := range world.sensors {
		world.sensors[i].hits = nil
		world.sensors[i].overlaps1 = nil
		world.sensors[i].overlaps2 = nil
	}

	for i := range world.islands {
		world.islands[i].bodies = nil
		world.islands[i].contacts = nil
		world.islands[i].joints = nil
	}
}

func deserializeIntoShell(r *snapReader, world *World) bool {
	desWorldConfig(r, world)
	desIDPool(r, &world.bodyIDPool)
	desIDPool(r, &world.shapeIDPool)
	desIDPool(r, &world.chainIDPool)
	desIDPool(r, &world.contactIDPool)
	desIDPool(r, &world.jointIDPool)
	desIDPool(r, &world.islandIDPool)
	desIDPool(r, &world.solverSetIDPool)

	n := snapReadCount(r, int(unsafe.Sizeof(solverSet{})), 24)
	if !r.ok {
		return false
	}

	world.solverSets = make([]solverSet, n)
	for i := range world.solverSets {
		desSolverSet(r, &world.solverSets[i])

		if !r.ok {
			return false
		}
	}

	desPodArray(r, &world.bodies)
	desPodArray(r, &world.shapes)
	desPodArray(r, &world.fatAABBs)
	desPodArray(r, &world.contacts)
	desPodArray(r, &world.joints)

	if !desChainShapes(r, world) {
		return false
	}

	if !desSensors(r, world) {
		return false
	}

	if !desIslands(r, world) {
		return false
	}

	for i := range world.broadPhase.trees {
		desTree(r, &world.broadPhase.trees[i])
	}

	desHashSet(r, &world.broadPhase.pairSet)

	for i := range world.constraintGraph.colors {
		desGraphColor(r, &world.constraintGraph.colors[i], i == overflowIndex)
	}

	world.bodyMoveEvents = world.bodyMoveEvents[:0]
	world.sensorBeginEvents = world.sensorBeginEvents[:0]

	world.contactBeginEvents = world.contactBeginEvents[:0]
	for i := range 2 {
		world.sensorEndEvents[i] = world.sensorEndEvents[i][:0]
		world.contactEndEvents[i] = world.contactEndEvents[i][:0]
	}

	world.contactHitEvents = world.contactHitEvents[:0]
	world.jointEvents = world.jointEvents[:0]

	return r.ok
}

func openSnapshotImage(image []byte, size int, r *snapReader) {
	if snapshotHeaderError(image, size) != nil {
		return
	}

	*r = snapReader{data: image[:size], cursor: 16, size: size, ok: true}
}

// UnmarshalBinary restores a world's simulation state from a snapshot image, in place. Handles
// held from this same world stay valid for objects that existed at the snapshot instant. Host
// wiring is preserved. Restore into the same world the snapshot came from to keep held handles
// valid. Must be called at a step boundary.
// On a rejected image (bad magic, version, or layout) the world is left unchanged.
// A corrupt payload detected after the rebuild begins returns an error and leaves
// the world unusable, so the caller must close it.
// Close remains safe even when only part of the payload was rebuilt.
func (w *World) UnmarshalBinary(image []byte) error {
	if err := snapshotHeaderError(image, len(image)); err != nil {
		return err
	}

	var r snapReader
	openSnapshotImage(image, len(image), &r)

	world := w
	checkInput(!world.locked, "world->locked == false")
	freeLiveSimElements(world)

	if !deserializeIntoShell(&r, world) {
		return &SnapshotError{Err: ErrCorrupt}
	}

	return nil
}

// AppendBinary appends a snapshot of the world's simulation state to b. Must be called at a
// step boundary.
func (w *World) AppendBinary(b []byte) ([]byte, error) {
	checkInput(!w.locked, "world->locked == false")

	buf := recBuffer{data: b[:cap(b)], capacity: cap(b), size: len(b)}
	serializeWorld(w, &buf)

	return buf.data[:buf.size], nil
}

// MarshalBinary returns a snapshot of the world's simulation state. Must be called at a step
// boundary.
func (w *World) MarshalBinary() ([]byte, error) {
	return w.AppendBinary(nil)
}

func fnvMixBytes(hash uint64, data []byte, n int) uint64 {
	for _, b := range data[:n] {
		hash = (hash ^ uint64(b)) * snapFNVPrime
	}

	return hash
}

func fnvMixFloat(hash uint64, f float32) uint64 {
	return (hash ^ uint64(math.Float32bits(f))) * snapFNVPrime
}

func fnvMixInt(hash uint64, v int) uint64 {
	return (hash ^ uint64(uint32(v))) * snapFNVPrime
}

func hashWorldStateDeep(world *World) uint64 {
	hash := snapFNVInit

	for i := range world.bodies {
		b := &world.bodies[i]
		if b.id != i {
			continue
		}

		sim := getBodySim(world, b)
		hash = fnvMixPosition(hash, sim.transform.P)
		hash = fnvMixFloat(hash, sim.transform.Q.C)

		hash = fnvMixFloat(hash, sim.transform.Q.S)
		if state := getBodyState(world, b); state != nil {
			hash = fnvMixFloat(hash, state.linearVelocity.X)
			hash = fnvMixFloat(hash, state.linearVelocity.Y)
			hash = fnvMixFloat(hash, state.angularVelocity)
		}

		hash = fnvMixInt(hash, b.setIndex)
		hash = fnvMixInt(hash, b.localIndex)
	}

	for i := range world.contacts {
		c := &world.contacts[i]
		if c.contactID != i {
			continue
		}

		hash = fnvMixInt(hash, c.setIndex)
		hash = fnvMixInt(hash, c.colorIndex)

		hash = fnvMixInt(hash, c.localIndex)
		if sim := getContactSim(world, c); sim != nil {
			m := &sim.manifold

			hash = fnvMixInt(hash, m.PointCount)
			for p := range m.PointCount {
				hash = fnvMixFloat(hash, m.Points[p].NormalImpulse)
				hash = fnvMixFloat(hash, m.Points[p].TangentImpulse)
				hash = fnvMixFloat(hash, m.Points[p].TotalNormalImpulse)
			}
		}
	}

	for i := range world.joints {
		j := &world.joints[i]
		if j.jointID != i {
			continue
		}

		hash = fnvMixInt(hash, j.setIndex)
		hash = fnvMixInt(hash, j.colorIndex)

		hash = fnvMixInt(hash, j.localIndex)
		if sim := getJointSim(world, j); sim != nil {
			hash = hashSnapshotJointImpulses(hash, sim)
		}
	}

	hash = fnvMixInt(hash, world.bodyIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.bodyIDPool))
	hash = fnvMixInt(hash, world.shapeIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.shapeIDPool))
	hash = fnvMixInt(hash, world.chainIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.chainIDPool))
	hash = fnvMixInt(hash, world.contactIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.contactIDPool))
	hash = fnvMixInt(hash, world.jointIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.jointIDPool))
	hash = fnvMixInt(hash, world.islandIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.islandIDPool))
	hash = fnvMixInt(hash, world.solverSetIDPool.nextIndex)
	hash = fnvMixInt(hash, getIDCount(&world.solverSetIDPool))
	hash = fnvMixInt(hash, len(world.solverSets))

	return hash
}

// Compute a deterministic hash of the simulation state: body transforms and velocities, contact and
// joint impulses, and the index bookkeeping that drives the solve. Reproduces exactly across worker
// counts and ignores struct padding and free slots, so two worlds that simulate identically always
// agree. Use this to detect desyncs (for example rollback resimulation) instead of comparing
// serialized world bytes, which carry non-canonical padding.
func (w *World) StateHash() uint64 {
	return hashWorldStateDeep(w)
}

func (e *SnapshotError) Error() string {
	return e.Err.Error()
}

func (e *SnapshotError) Unwrap() error {
	return e.Err
}

func serPodArray[T any](buf *recBuffer, arr []T) {
	snapWI32(buf, len(arr))

	for _, v := range arr {
		snapWriteValue(buf, v)
	}
}

func serUserDataArray[T any](buf *recBuffer, arr []T, clearUserData func(*T)) {
	snapWI32(buf, len(arr))

	for _, v := range arr {
		clearUserData(&v)
		snapWriteValue(buf, v)
	}
}

func desPodArray[T any](r *snapReader, arr *[]T) {
	n := snapRI32(r)
	size := snapMinSize[T]()

	if !r.ok {
		return
	}

	if !snapCheckCount(r, n, int(unsafe.Sizeof(*new(T))), size) {
		r.ok = false
		return
	}

	if n == 0 {
		*arr = (*arr)[:0]
		return
	}

	*arr = make([]T, n)
	for i := range *arr {
		snapReadValue(r, &(*arr)[i])

		if !r.ok {
			return
		}
	}
}

func snapWriteBody(buf *recBuffer, v body) {
	snapWI32(buf, v.setIndex)
	snapWI32(buf, v.localIndex)
	snapWI32(buf, v.headContactKey)
	snapWI32(buf, v.contactCount)
	snapWI32(buf, v.headShapeID)
	snapWI32(buf, v.shapeCount)
	snapWI32(buf, v.headChainID)
	snapWI32(buf, v.headJointKey)
	snapWI32(buf, v.jointCount)
	snapWI32(buf, v.islandID)
	snapWI32(buf, v.islandIndex)
	recWF32(buf, v.mass)
	recWF32(buf, v.inertia)
	recWF32(buf, v.sleepThreshold)
	recWF32(buf, v.sleepTime)
	recWF32(buf, v.safetyFactor)
	snapWI32(buf, v.bodyMoveIndex)
	snapWI32(buf, v.id)
	recWU32(buf, v.flags)
	snapWI32(buf, int(v.bodyType))
	recWU16(buf, v.generation)

	for i := range v.name {
		recWU8(buf, v.name[i])
	}
}

func snapReadBody(r *snapReader, v *body) {
	v.userData = nil
	v.setIndex = snapRI32(r)
	v.localIndex = snapRI32(r)
	v.headContactKey = snapRI32(r)
	v.contactCount = snapRI32(r)
	v.headShapeID = snapRI32(r)
	v.shapeCount = snapRI32(r)
	v.headChainID = snapRI32(r)
	v.headJointKey = snapRI32(r)
	v.jointCount = snapRI32(r)
	v.islandID = snapRI32(r)
	v.islandIndex = snapRI32(r)
	v.mass = math.Float32frombits(snapRU32(r))
	v.inertia = math.Float32frombits(snapRU32(r))
	v.sleepThreshold = math.Float32frombits(snapRU32(r))
	v.sleepTime = math.Float32frombits(snapRU32(r))
	v.safetyFactor = math.Float32frombits(snapRU32(r))
	v.bodyMoveIndex = snapRI32(r)
	v.id = snapRI32(r)
	v.flags = snapRU32(r)
	v.bodyType = BodyType(snapRI32(r))

	v.generation = snapReadU16(r)
	for i := range v.name {
		v.name[i] = snapReadU8(r)
	}
}

func snapWritePos(buf *recBuffer, v Pos) {
	if doublePrecision {
		recWF64(buf, float64(v.X))
	} else {
		recWF32(buf, float32(v.X))
	}

	if doublePrecision {
		recWF64(buf, float64(v.Y))
	} else {
		recWF32(buf, float32(v.Y))
	}
}

func snapReadPos(r *snapReader, v *Pos) {
	if doublePrecision {
		v.X = posScalar(math.Float64frombits(snapReadU64(r)))
	} else {
		v.X = posScalar(math.Float32frombits(snapRU32(r)))
	}

	if doublePrecision {
		v.Y = posScalar(math.Float64frombits(snapReadU64(r)))
	} else {
		v.Y = posScalar(math.Float32frombits(snapRU32(r)))
	}
}

func snapWriteRot(buf *recBuffer, v Rot) {
	recWF32(buf, v.C)
	recWF32(buf, v.S)
}

func snapReadRot(r *snapReader, v *Rot) {
	v.C = math.Float32frombits(snapRU32(r))
	v.S = math.Float32frombits(snapRU32(r))
}

func snapWriteWorldTransform(buf *recBuffer, v WorldTransform) {
	snapWritePos(buf, v.P)
	snapWriteRot(buf, v.Q)
}

func snapReadWorldTransform(r *snapReader, v *WorldTransform) {
	snapReadPos(r, &v.P)
	snapReadRot(r, &v.Q)
}

func snapWriteVec2(buf *recBuffer, v Vec2) {
	recWF32(buf, v.X)
	recWF32(buf, v.Y)
}

func snapReadVec2(r *snapReader, v *Vec2) {
	v.X = math.Float32frombits(snapRU32(r))
	v.Y = math.Float32frombits(snapRU32(r))
}

func snapWriteBodySim(buf *recBuffer, v bodySim) {
	snapWriteWorldTransform(buf, v.transform)
	snapWritePos(buf, v.center)
	snapWriteRot(buf, v.rotation0)
	snapWritePos(buf, v.center0)
	snapWriteVec2(buf, v.localCenter)
	snapWriteVec2(buf, v.force)
	recWF32(buf, v.torque)
	recWF32(buf, v.invMass)
	recWF32(buf, v.invInertia)
	recWF32(buf, v.minExtent)
	recWF32(buf, v.maxExtent)
	recWF32(buf, v.linearDamping)
	recWF32(buf, v.angularDamping)
	recWF32(buf, v.gravityScale)
	snapWI32(buf, v.bodyID)
	recWU32(buf, v.flags)
}

func snapReadBodySim(r *snapReader, v *bodySim) {
	snapReadWorldTransform(r, &v.transform)
	snapReadPos(r, &v.center)
	snapReadRot(r, &v.rotation0)
	snapReadPos(r, &v.center0)
	snapReadVec2(r, &v.localCenter)
	snapReadVec2(r, &v.force)
	v.torque = math.Float32frombits(snapRU32(r))
	v.invMass = math.Float32frombits(snapRU32(r))
	v.invInertia = math.Float32frombits(snapRU32(r))
	v.minExtent = math.Float32frombits(snapRU32(r))
	v.maxExtent = math.Float32frombits(snapRU32(r))
	v.linearDamping = math.Float32frombits(snapRU32(r))
	v.angularDamping = math.Float32frombits(snapRU32(r))
	v.gravityScale = math.Float32frombits(snapRU32(r))
	v.bodyID = snapRI32(r)
	v.flags = snapRU32(r)
}

func snapWriteBodyState(buf *recBuffer, v bodyState) {
	snapWriteVec2(buf, v.linearVelocity)
	recWF32(buf, v.angularVelocity)
	recWU32(buf, v.flags)
	snapWriteVec2(buf, v.deltaPosition)
	snapWriteRot(buf, v.deltaRotation)
}

func snapReadBodyState(r *snapReader, v *bodyState) {
	snapReadVec2(r, &v.linearVelocity)
	v.angularVelocity = math.Float32frombits(snapRU32(r))
	v.flags = snapRU32(r)
	snapReadVec2(r, &v.deltaPosition)
	snapReadRot(r, &v.deltaRotation)
}

func snapWriteSurfaceMaterial(buf *recBuffer, v SurfaceMaterial) {
	recWF32(buf, v.Friction)
	recWF32(buf, v.Restitution)
	recWF32(buf, v.RollingResistance)
	recWF32(buf, v.TangentSpeed)
	recWU64(buf, v.UserMaterialID)
	recWU32(buf, v.CustomColor)
}

func snapReadSurfaceMaterial(r *snapReader, v *SurfaceMaterial) {
	v.Friction = math.Float32frombits(snapRU32(r))
	v.Restitution = math.Float32frombits(snapRU32(r))
	v.RollingResistance = math.Float32frombits(snapRU32(r))
	v.TangentSpeed = math.Float32frombits(snapRU32(r))
	v.UserMaterialID = snapReadU64(r)
	v.CustomColor = snapRU32(r)
}

func snapWriteAABB(buf *recBuffer, v AABB) {
	snapWriteVec2(buf, v.LowerBound)
	snapWriteVec2(buf, v.UpperBound)
}

func snapReadAABB(r *snapReader, v *AABB) {
	snapReadVec2(r, &v.LowerBound)
	snapReadVec2(r, &v.UpperBound)
}

func snapWriteFilter(buf *recBuffer, v Filter) {
	recWU64(buf, v.CategoryBits)
	recWU64(buf, v.MaskBits)
	snapWI32(buf, v.GroupIndex)
}

func snapReadFilter(r *snapReader, v *Filter) {
	v.CategoryBits = snapReadU64(r)
	v.MaskBits = snapReadU64(r)
	v.GroupIndex = snapRI32(r)
}

func snapWriteCircle(buf *recBuffer, v Circle) {
	snapWriteVec2(buf, v.Center)
	recWF32(buf, v.Radius)
}

func snapReadCircle(r *snapReader, v *Circle) {
	snapReadVec2(r, &v.Center)
	v.Radius = math.Float32frombits(snapRU32(r))
}

func snapWriteCapsule(buf *recBuffer, v Capsule) {
	snapWriteVec2(buf, v.Center1)
	snapWriteVec2(buf, v.Center2)
	recWF32(buf, v.Radius)
}

func snapReadCapsule(r *snapReader, v *Capsule) {
	snapReadVec2(r, &v.Center1)
	snapReadVec2(r, &v.Center2)
	v.Radius = math.Float32frombits(snapRU32(r))
}

func snapWritePolygon(buf *recBuffer, v Polygon) {
	for i := range v.Vertices {
		snapWriteVec2(buf, v.Vertices[i])
	}

	for i := range v.Normals {
		snapWriteVec2(buf, v.Normals[i])
	}

	snapWriteVec2(buf, v.Centroid)
	recWF32(buf, v.Radius)
	snapWI32(buf, v.Count)
}

func snapReadPolygon(r *snapReader, v *Polygon) {
	for i := range v.Vertices {
		snapReadVec2(r, &v.Vertices[i])
	}

	for i := range v.Normals {
		snapReadVec2(r, &v.Normals[i])
	}

	snapReadVec2(r, &v.Centroid)
	v.Radius = math.Float32frombits(snapRU32(r))
	v.Count = snapRI32(r)
}

func snapWriteSegment(buf *recBuffer, v Segment) {
	snapWriteVec2(buf, v.Point1)
	snapWriteVec2(buf, v.Point2)
}

func snapReadSegment(r *snapReader, v *Segment) {
	snapReadVec2(r, &v.Point1)
	snapReadVec2(r, &v.Point2)
}

func snapWriteChainSegment(buf *recBuffer, v ChainSegment) {
	snapWriteVec2(buf, v.Ghost1)
	snapWriteSegment(buf, v.Segment)
	snapWriteVec2(buf, v.Ghost2)
	snapWI32(buf, v.ChainID)
}

func snapReadChainSegment(r *snapReader, v *ChainSegment) {
	snapReadVec2(r, &v.Ghost1)
	snapReadSegment(r, &v.Segment)
	snapReadVec2(r, &v.Ghost2)
	v.ChainID = snapRI32(r)
}

func snapWriteShape(buf *recBuffer, v shape) {
	snapWI32(buf, v.id)
	snapWI32(buf, v.bodyID)
	snapWI32(buf, v.prevShapeID)
	snapWI32(buf, v.nextShapeID)
	snapWI32(buf, v.sensorIndex)
	snapWI32(buf, int(v.shapeType))
	snapWriteSurfaceMaterial(buf, v.material)
	recWF32(buf, v.density)
	recWF32(buf, v.aabbMargin)
	snapWriteAABB(buf, v.aabb)
	snapWriteVec2(buf, v.localCentroid)
	snapWI32(buf, v.proxyKey)
	snapWriteFilter(buf, v.filter)
	recWU16(buf, v.generation)
	recWBOOL(buf, v.enableSensorEvents)
	recWBOOL(buf, v.enableContactEvents)
	recWBOOL(buf, v.enableCustomFiltering)
	recWBOOL(buf, v.enableHitEvents)
	recWBOOL(buf, v.enablePreSolveEvents)

	switch v.shapeType {
	case CircleShape:
		snapWriteCircle(buf, v.circle)
	case CapsuleShape:
		snapWriteCapsule(buf, v.capsule)
	case PolygonShape:
		snapWritePolygon(buf, v.polygon)
	case SegmentShape:
		snapWriteSegment(buf, v.segment)
	case ChainSegmentShape:
		snapWriteChainSegment(buf, v.chainSegment)
	case ShapeTypeCount:
		// The count sentinel has no geometry payload.
	}
}

func snapReadShape(r *snapReader, v *shape) {
	v.id = snapRI32(r)
	v.bodyID = snapRI32(r)
	v.prevShapeID = snapRI32(r)
	v.nextShapeID = snapRI32(r)
	v.sensorIndex = snapRI32(r)
	v.shapeType = ShapeType(snapRI32(r))
	snapReadSurfaceMaterial(r, &v.material)
	v.density = math.Float32frombits(snapRU32(r))
	v.aabbMargin = math.Float32frombits(snapRU32(r))
	snapReadAABB(r, &v.aabb)
	snapReadVec2(r, &v.localCentroid)
	v.proxyKey = snapRI32(r)
	snapReadFilter(r, &v.filter)
	v.userData = nil
	v.generation = snapReadU16(r)
	v.enableSensorEvents = snapReadU8(r) != 0
	v.enableContactEvents = snapReadU8(r) != 0
	v.enableCustomFiltering = snapReadU8(r) != 0
	v.enableHitEvents = snapReadU8(r) != 0

	v.enablePreSolveEvents = snapReadU8(r) != 0
	switch v.shapeType {
	case CircleShape:
		snapReadCircle(r, &v.circle)
	case CapsuleShape:
		snapReadCapsule(r, &v.capsule)
	case PolygonShape:
		snapReadPolygon(r, &v.polygon)
	case SegmentShape:
		snapReadSegment(r, &v.segment)
	case ChainSegmentShape:
		snapReadChainSegment(r, &v.chainSegment)
	case ShapeTypeCount:
		r.ok = false
	default:
		r.ok = false
	}
}

func snapWriteContactEdge(buf *recBuffer, v contactEdge) {
	snapWI32(buf, v.bodyID)
	snapWI32(buf, v.prevKey)
	snapWI32(buf, v.nextKey)
}

func snapReadContactEdge(r *snapReader, v *contactEdge) {
	v.bodyID = snapRI32(r)
	v.prevKey = snapRI32(r)
	v.nextKey = snapRI32(r)
}

func snapWriteContact(buf *recBuffer, v contact) {
	for i := range v.edges {
		snapWriteContactEdge(buf, v.edges[i])
	}

	snapWI32(buf, v.islandID)
	snapWI32(buf, v.islandIndex)
	snapWI32(buf, v.setIndex)
	snapWI32(buf, v.colorIndex)
	snapWI32(buf, v.localIndex)
	snapWI32(buf, v.shapeIDA)
	snapWI32(buf, v.shapeIDB)
	snapWI32(buf, v.contactID)
	recWU32(buf, v.flags)
	recWU32(buf, v.generation)
}

func snapReadContact(r *snapReader, v *contact) {
	for i := range v.edges {
		snapReadContactEdge(r, &v.edges[i])
	}

	v.islandID = snapRI32(r)
	v.islandIndex = snapRI32(r)
	v.setIndex = snapRI32(r)
	v.colorIndex = snapRI32(r)
	v.localIndex = snapRI32(r)
	v.shapeIDA = snapRI32(r)
	v.shapeIDB = snapRI32(r)
	v.contactID = snapRI32(r)
	v.flags = snapRU32(r)
	v.generation = snapRU32(r)
}

func snapWriteTransform(buf *recBuffer, v Transform) {
	snapWriteVec2(buf, v.P)
	snapWriteRot(buf, v.Q)
}

func snapReadTransform(r *snapReader, v *Transform) {
	snapReadVec2(r, &v.P)
	snapReadRot(r, &v.Q)
}

func snapWriteManifoldPoint(buf *recBuffer, v ManifoldPoint) {
	snapWriteVec2(buf, v.AnchorA)
	snapWriteVec2(buf, v.AnchorB)
	recWF32(buf, v.Separation)
	recWF32(buf, v.NormalImpulse)
	recWF32(buf, v.TangentImpulse)
	recWF32(buf, v.TotalNormalImpulse)
	recWF32(buf, v.NormalVelocity)
	recWF32(buf, v.BaseSeparation)
	recWU16(buf, v.ID)
	recWBOOL(buf, v.Persisted)
}

func snapReadManifoldPoint(r *snapReader, v *ManifoldPoint) {
	snapReadVec2(r, &v.AnchorA)
	snapReadVec2(r, &v.AnchorB)
	v.Separation = math.Float32frombits(snapRU32(r))
	v.NormalImpulse = math.Float32frombits(snapRU32(r))
	v.TangentImpulse = math.Float32frombits(snapRU32(r))
	v.TotalNormalImpulse = math.Float32frombits(snapRU32(r))
	v.NormalVelocity = math.Float32frombits(snapRU32(r))
	v.BaseSeparation = math.Float32frombits(snapRU32(r))
	v.ID = snapReadU16(r)
	v.Persisted = snapReadU8(r) != 0
}

func snapWriteManifold(buf *recBuffer, v Manifold) {
	snapWriteVec2(buf, v.Normal)
	recWF32(buf, v.RollingImpulse)

	for i := range v.Points {
		snapWriteManifoldPoint(buf, v.Points[i])
	}

	snapWI32(buf, v.PointCount)
}

func snapReadManifold(r *snapReader, v *Manifold) {
	snapReadVec2(r, &v.Normal)

	v.RollingImpulse = math.Float32frombits(snapRU32(r))
	for i := range v.Points {
		snapReadManifoldPoint(r, &v.Points[i])
	}

	v.PointCount = snapRI32(r)
}

func snapWriteSimplexCache(buf *recBuffer, v SimplexCache) {
	recWU16(buf, v.Count)

	for i := range v.IndexA {
		recWU8(buf, v.IndexA[i])
	}

	for i := range v.IndexB {
		recWU8(buf, v.IndexB[i])
	}
}

func snapReadSimplexCache(r *snapReader, v *SimplexCache) {
	v.Count = snapReadU16(r)
	for i := range v.IndexA {
		v.IndexA[i] = snapReadU8(r)
	}

	for i := range v.IndexB {
		v.IndexB[i] = snapReadU8(r)
	}
}

//nolint:dupl // Snapshot v12 contact and prismatic layouts encode unrelated structs in independently fixed field order.
func snapWriteContactSim(buf *recBuffer, v contactSim) {
	snapWI32(buf, v.contactID)
	snapWriteRot(buf, v.cachedRotationA)
	snapWriteRot(buf, v.cachedRotationB)
	snapWriteTransform(buf, v.cachedRelativePose)
	snapWI32(buf, v.bodyIDA)
	snapWI32(buf, v.bodyIDB)
	snapWI32(buf, v.encodedBodySimA)
	snapWI32(buf, v.encodedBodySimB)
	recWU32(buf, v.simFlags)
	snapWI32(buf, v.shapeIDA)
	snapWI32(buf, v.shapeIDB)
	recWF32(buf, v.invMassA)
	recWF32(buf, v.invIA)
	recWF32(buf, v.invMassB)
	recWF32(buf, v.invIB)
	snapWriteManifold(buf, v.manifold)
	recWF32(buf, v.friction)
	recWF32(buf, v.restitution)
	recWF32(buf, v.rollingResistance)
	recWF32(buf, v.tangentSpeed)
	snapWriteSimplexCache(buf, v.cache)
}

func snapReadContactSim(r *snapReader, v *contactSim) {
	v.contactID = snapRI32(r)
	snapReadRot(r, &v.cachedRotationA)
	snapReadRot(r, &v.cachedRotationB)
	snapReadTransform(r, &v.cachedRelativePose)
	v.bodyIDA = snapRI32(r)
	v.bodyIDB = snapRI32(r)
	v.encodedBodySimA = snapRI32(r)
	v.encodedBodySimB = snapRI32(r)
	v.simFlags = snapRU32(r)
	v.shapeIDA = snapRI32(r)
	v.shapeIDB = snapRI32(r)
	v.invMassA = math.Float32frombits(snapRU32(r))
	v.invIA = math.Float32frombits(snapRU32(r))
	v.invMassB = math.Float32frombits(snapRU32(r))
	v.invIB = math.Float32frombits(snapRU32(r))
	snapReadManifold(r, &v.manifold)
	v.friction = math.Float32frombits(snapRU32(r))
	v.restitution = math.Float32frombits(snapRU32(r))
	v.rollingResistance = math.Float32frombits(snapRU32(r))
	v.tangentSpeed = math.Float32frombits(snapRU32(r))
	snapReadSimplexCache(r, &v.cache)
}

func snapWriteJointEdge(buf *recBuffer, v jointEdge) {
	snapWI32(buf, v.bodyID)
	snapWI32(buf, v.prevKey)
	snapWI32(buf, v.nextKey)
}

func snapReadJointEdge(r *snapReader, v *jointEdge) {
	v.bodyID = snapRI32(r)
	v.prevKey = snapRI32(r)
	v.nextKey = snapRI32(r)
}

func snapWriteJoint(buf *recBuffer, v joint) {
	snapWI32(buf, v.setIndex)
	snapWI32(buf, v.colorIndex)
	snapWI32(buf, v.localIndex)

	for i := range v.edges {
		snapWriteJointEdge(buf, v.edges[i])
	}

	snapWI32(buf, v.jointID)
	snapWI32(buf, v.islandID)
	snapWI32(buf, v.islandIndex)
	recWF32(buf, v.drawScale)
	snapWI32(buf, int(v.jointType))
	recWU16(buf, v.generation)
	recWBOOL(buf, v.collideConnected)
}

func snapReadJoint(r *snapReader, v *joint) {
	v.userData = nil
	v.setIndex = snapRI32(r)
	v.colorIndex = snapRI32(r)

	v.localIndex = snapRI32(r)
	for i := range v.edges {
		snapReadJointEdge(r, &v.edges[i])
	}

	v.jointID = snapRI32(r)
	v.islandID = snapRI32(r)
	v.islandIndex = snapRI32(r)
	v.drawScale = math.Float32frombits(snapRU32(r))
	v.jointType = JointType(snapRI32(r))
	v.generation = snapReadU16(r)
	v.collideConnected = snapReadU8(r) != 0
}

func snapWriteSoftness(buf *recBuffer, v softness) {
	recWF32(buf, v.biasRate)
	recWF32(buf, v.massScale)
	recWF32(buf, v.impulseScale)
}

func snapReadSoftness(r *snapReader, v *softness) {
	v.biasRate = math.Float32frombits(snapRU32(r))
	v.massScale = math.Float32frombits(snapRU32(r))
	v.impulseScale = math.Float32frombits(snapRU32(r))
}

//nolint:dupl // Snapshot v12 distance, motor, and wheel layouts have distinct fields and independently fixed field order.
func snapWriteDistanceJoint(buf *recBuffer, v distanceJoint) {
	recWF32(buf, v.length)
	recWF32(buf, v.hertz)
	recWF32(buf, v.dampingRatio)
	recWF32(buf, v.lowerSpringForce)
	recWF32(buf, v.upperSpringForce)
	recWF32(buf, v.minLength)
	recWF32(buf, v.maxLength)
	recWF32(buf, v.maxMotorForce)
	recWF32(buf, v.motorSpeed)
	recWF32(buf, v.impulse)
	recWF32(buf, v.lowerImpulse)
	recWF32(buf, v.upperImpulse)
	recWF32(buf, v.motorImpulse)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteVec2(buf, v.anchorA)
	snapWriteVec2(buf, v.anchorB)
	snapWriteVec2(buf, v.deltaCenter)
	snapWriteSoftness(buf, v.distanceSoftness)
	recWF32(buf, v.axialMass)
	recWBOOL(buf, v.enableSpring)
	recWBOOL(buf, v.enableLimit)
	recWBOOL(buf, v.enableMotor)
}

func snapReadDistanceJoint(r *snapReader, v *distanceJoint) {
	v.length = math.Float32frombits(snapRU32(r))
	v.hertz = math.Float32frombits(snapRU32(r))
	v.dampingRatio = math.Float32frombits(snapRU32(r))
	v.lowerSpringForce = math.Float32frombits(snapRU32(r))
	v.upperSpringForce = math.Float32frombits(snapRU32(r))
	v.minLength = math.Float32frombits(snapRU32(r))
	v.maxLength = math.Float32frombits(snapRU32(r))
	v.maxMotorForce = math.Float32frombits(snapRU32(r))
	v.motorSpeed = math.Float32frombits(snapRU32(r))
	v.impulse = math.Float32frombits(snapRU32(r))
	v.lowerImpulse = math.Float32frombits(snapRU32(r))
	v.upperImpulse = math.Float32frombits(snapRU32(r))
	v.motorImpulse = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadVec2(r, &v.anchorA)
	snapReadVec2(r, &v.anchorB)
	snapReadVec2(r, &v.deltaCenter)
	snapReadSoftness(r, &v.distanceSoftness)
	v.axialMass = math.Float32frombits(snapRU32(r))
	v.enableSpring = snapReadU8(r) != 0
	v.enableLimit = snapReadU8(r) != 0
	v.enableMotor = snapReadU8(r) != 0
}

func snapWriteMat22(buf *recBuffer, v Mat22) {
	snapWriteVec2(buf, v.Cx)
	snapWriteVec2(buf, v.Cy)
}

func snapReadMat22(r *snapReader, v *Mat22) {
	snapReadVec2(r, &v.Cx)
	snapReadVec2(r, &v.Cy)
}

//nolint:dupl // Snapshot v12 motor, distance, and wheel layouts have distinct fields and independently fixed field order.
func snapWriteMotorJoint(buf *recBuffer, v motorJoint) {
	snapWriteVec2(buf, v.linearVelocity)
	recWF32(buf, v.maxVelocityForce)
	recWF32(buf, v.angularVelocity)
	recWF32(buf, v.maxVelocityTorque)
	recWF32(buf, v.linearHertz)
	recWF32(buf, v.linearDampingRatio)
	recWF32(buf, v.maxSpringForce)
	recWF32(buf, v.angularHertz)
	recWF32(buf, v.angularDampingRatio)
	recWF32(buf, v.maxSpringTorque)
	snapWriteVec2(buf, v.linearVelocityImpulse)
	recWF32(buf, v.angularVelocityImpulse)
	snapWriteVec2(buf, v.linearSpringImpulse)
	recWF32(buf, v.angularSpringImpulse)
	snapWriteSoftness(buf, v.linearSpring)
	snapWriteSoftness(buf, v.angularSpring)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	snapWriteMat22(buf, v.linearMass)
	recWF32(buf, v.angularMass)
}

func snapReadMotorJoint(r *snapReader, v *motorJoint) {
	snapReadVec2(r, &v.linearVelocity)
	v.maxVelocityForce = math.Float32frombits(snapRU32(r))
	v.angularVelocity = math.Float32frombits(snapRU32(r))
	v.maxVelocityTorque = math.Float32frombits(snapRU32(r))
	v.linearHertz = math.Float32frombits(snapRU32(r))
	v.linearDampingRatio = math.Float32frombits(snapRU32(r))
	v.maxSpringForce = math.Float32frombits(snapRU32(r))
	v.angularHertz = math.Float32frombits(snapRU32(r))
	v.angularDampingRatio = math.Float32frombits(snapRU32(r))
	v.maxSpringTorque = math.Float32frombits(snapRU32(r))
	snapReadVec2(r, &v.linearVelocityImpulse)
	v.angularVelocityImpulse = math.Float32frombits(snapRU32(r))
	snapReadVec2(r, &v.linearSpringImpulse)
	v.angularSpringImpulse = math.Float32frombits(snapRU32(r))
	snapReadSoftness(r, &v.linearSpring)
	snapReadSoftness(r, &v.angularSpring)
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	snapReadMat22(r, &v.linearMass)
	v.angularMass = math.Float32frombits(snapRU32(r))
}

func snapWriteMoverJoint(buf *recBuffer, v moverJoint) {
	snapWriteVec2(buf, v.linearVelocity)
	snapWriteVec2(buf, v.maxVelocityForce)
	snapWriteVec2(buf, v.linearVelocityImpulse)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	recWF32(buf, v.linearMass)
}

func snapReadMoverJoint(r *snapReader, v *moverJoint) {
	snapReadVec2(r, &v.linearVelocity)
	snapReadVec2(r, &v.maxVelocityForce)
	snapReadVec2(r, &v.linearVelocityImpulse)
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	v.linearMass = math.Float32frombits(snapRU32(r))
}

func snapWritePogoJoint(buf *recBuffer, v pogoJoint) {
	snapWriteVec2(buf, v.normal)
	recWF32(buf, v.restLength)
	recWF32(buf, v.hertz)
	recWF32(buf, v.dampingRatio)
	recWF32(buf, v.maxTensionForce)
	recWF32(buf, v.maxCompressionForce)
	recWF32(buf, v.impulse)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	recWF32(buf, v.linearMass)
	recWF32(buf, v.velocity)
}

func snapReadPogoJoint(r *snapReader, v *pogoJoint) {
	snapReadVec2(r, &v.normal)
	v.restLength = math.Float32frombits(snapRU32(r))
	v.hertz = math.Float32frombits(snapRU32(r))
	v.dampingRatio = math.Float32frombits(snapRU32(r))
	v.maxTensionForce = math.Float32frombits(snapRU32(r))
	v.maxCompressionForce = math.Float32frombits(snapRU32(r))
	v.impulse = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	v.linearMass = math.Float32frombits(snapRU32(r))
	v.velocity = math.Float32frombits(snapRU32(r))
}

//nolint:dupl // Snapshot v12 prismatic and contact layouts encode unrelated structs in independently fixed field order.
func snapWritePrismaticJoint(buf *recBuffer, v prismaticJoint) {
	snapWriteVec2(buf, v.impulse)
	recWF32(buf, v.springImpulse)
	recWF32(buf, v.motorImpulse)
	recWF32(buf, v.lowerImpulse)
	recWF32(buf, v.upperImpulse)
	recWF32(buf, v.hertz)
	recWF32(buf, v.dampingRatio)
	recWF32(buf, v.targetTranslation)
	recWF32(buf, v.maxMotorForce)
	recWF32(buf, v.motorSpeed)
	recWF32(buf, v.lowerTranslation)
	recWF32(buf, v.upperTranslation)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	snapWriteSoftness(buf, v.springSoftness)
	recWBOOL(buf, v.enableSpring)
	recWBOOL(buf, v.enableLimit)
	recWBOOL(buf, v.enableMotor)
}

func snapReadPrismaticJoint(r *snapReader, v *prismaticJoint) {
	snapReadVec2(r, &v.impulse)
	v.springImpulse = math.Float32frombits(snapRU32(r))
	v.motorImpulse = math.Float32frombits(snapRU32(r))
	v.lowerImpulse = math.Float32frombits(snapRU32(r))
	v.upperImpulse = math.Float32frombits(snapRU32(r))
	v.hertz = math.Float32frombits(snapRU32(r))
	v.dampingRatio = math.Float32frombits(snapRU32(r))
	v.targetTranslation = math.Float32frombits(snapRU32(r))
	v.maxMotorForce = math.Float32frombits(snapRU32(r))
	v.motorSpeed = math.Float32frombits(snapRU32(r))
	v.lowerTranslation = math.Float32frombits(snapRU32(r))
	v.upperTranslation = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	snapReadSoftness(r, &v.springSoftness)
	v.enableSpring = snapReadU8(r) != 0
	v.enableLimit = snapReadU8(r) != 0
	v.enableMotor = snapReadU8(r) != 0
}

func snapWriteRevoluteJoint(buf *recBuffer, v revoluteJoint) {
	snapWriteVec2(buf, v.linearImpulse)
	recWF32(buf, v.springImpulse)
	recWF32(buf, v.motorImpulse)
	recWF32(buf, v.lowerImpulse)
	recWF32(buf, v.upperImpulse)
	recWF32(buf, v.hertz)
	recWF32(buf, v.dampingRatio)
	recWF32(buf, v.targetAngle)
	recWF32(buf, v.maxMotorTorque)
	recWF32(buf, v.motorSpeed)
	recWF32(buf, v.lowerAngle)
	recWF32(buf, v.upperAngle)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	recWF32(buf, v.axialMass)
	snapWriteSoftness(buf, v.springSoftness)
	recWBOOL(buf, v.enableSpring)
	recWBOOL(buf, v.enableMotor)
	recWBOOL(buf, v.enableLimit)
}

func snapReadRevoluteJoint(r *snapReader, v *revoluteJoint) {
	snapReadVec2(r, &v.linearImpulse)
	v.springImpulse = math.Float32frombits(snapRU32(r))
	v.motorImpulse = math.Float32frombits(snapRU32(r))
	v.lowerImpulse = math.Float32frombits(snapRU32(r))
	v.upperImpulse = math.Float32frombits(snapRU32(r))
	v.hertz = math.Float32frombits(snapRU32(r))
	v.dampingRatio = math.Float32frombits(snapRU32(r))
	v.targetAngle = math.Float32frombits(snapRU32(r))
	v.maxMotorTorque = math.Float32frombits(snapRU32(r))
	v.motorSpeed = math.Float32frombits(snapRU32(r))
	v.lowerAngle = math.Float32frombits(snapRU32(r))
	v.upperAngle = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	v.axialMass = math.Float32frombits(snapRU32(r))
	snapReadSoftness(r, &v.springSoftness)
	v.enableSpring = snapReadU8(r) != 0
	v.enableMotor = snapReadU8(r) != 0
	v.enableLimit = snapReadU8(r) != 0
}

func snapWriteWeldJoint(buf *recBuffer, v weldJoint) {
	recWF32(buf, v.linearHertz)
	recWF32(buf, v.linearDampingRatio)
	recWF32(buf, v.angularHertz)
	recWF32(buf, v.angularDampingRatio)
	snapWriteSoftness(buf, v.linearSpring)
	snapWriteSoftness(buf, v.angularSpring)
	snapWriteVec2(buf, v.linearImpulse)
	recWF32(buf, v.angularImpulse)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	recWF32(buf, v.axialMass)
}

func snapReadWeldJoint(r *snapReader, v *weldJoint) {
	v.linearHertz = math.Float32frombits(snapRU32(r))
	v.linearDampingRatio = math.Float32frombits(snapRU32(r))
	v.angularHertz = math.Float32frombits(snapRU32(r))
	v.angularDampingRatio = math.Float32frombits(snapRU32(r))
	snapReadSoftness(r, &v.linearSpring)
	snapReadSoftness(r, &v.angularSpring)
	snapReadVec2(r, &v.linearImpulse)
	v.angularImpulse = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	v.axialMass = math.Float32frombits(snapRU32(r))
}

//nolint:dupl // Snapshot v12 wheel, distance, and motor layouts have distinct fields and independently fixed field order.
func snapWriteWheelJoint(buf *recBuffer, v wheelJoint) {
	recWF32(buf, v.perpImpulse)
	recWF32(buf, v.motorImpulse)
	recWF32(buf, v.springImpulse)
	recWF32(buf, v.lowerImpulse)
	recWF32(buf, v.upperImpulse)
	recWF32(buf, v.maxMotorTorque)
	recWF32(buf, v.motorSpeed)
	recWF32(buf, v.lowerTranslation)
	recWF32(buf, v.upperTranslation)
	recWF32(buf, v.hertz)
	recWF32(buf, v.dampingRatio)
	snapWI32(buf, v.indexA)
	snapWI32(buf, v.indexB)
	snapWriteTransform(buf, v.frameA)
	snapWriteTransform(buf, v.frameB)
	snapWriteVec2(buf, v.deltaCenter)
	recWF32(buf, v.perpMass)
	recWF32(buf, v.motorMass)
	recWF32(buf, v.axialMass)
	snapWriteSoftness(buf, v.springSoftness)
	recWBOOL(buf, v.enableSpring)
	recWBOOL(buf, v.enableMotor)
	recWBOOL(buf, v.enableLimit)
}

func snapReadWheelJoint(r *snapReader, v *wheelJoint) {
	v.perpImpulse = math.Float32frombits(snapRU32(r))
	v.motorImpulse = math.Float32frombits(snapRU32(r))
	v.springImpulse = math.Float32frombits(snapRU32(r))
	v.lowerImpulse = math.Float32frombits(snapRU32(r))
	v.upperImpulse = math.Float32frombits(snapRU32(r))
	v.maxMotorTorque = math.Float32frombits(snapRU32(r))
	v.motorSpeed = math.Float32frombits(snapRU32(r))
	v.lowerTranslation = math.Float32frombits(snapRU32(r))
	v.upperTranslation = math.Float32frombits(snapRU32(r))
	v.hertz = math.Float32frombits(snapRU32(r))
	v.dampingRatio = math.Float32frombits(snapRU32(r))
	v.indexA = snapRI32(r)
	v.indexB = snapRI32(r)
	snapReadTransform(r, &v.frameA)
	snapReadTransform(r, &v.frameB)
	snapReadVec2(r, &v.deltaCenter)
	v.perpMass = math.Float32frombits(snapRU32(r))
	v.motorMass = math.Float32frombits(snapRU32(r))
	v.axialMass = math.Float32frombits(snapRU32(r))
	snapReadSoftness(r, &v.springSoftness)
	v.enableSpring = snapReadU8(r) != 0
	v.enableMotor = snapReadU8(r) != 0
	v.enableLimit = snapReadU8(r) != 0
}

func snapWriteJointSim(buf *recBuffer, v jointSim) {
	snapWI32(buf, v.jointID)
	snapWI32(buf, v.bodyIDA)
	snapWI32(buf, v.bodyIDB)
	snapWI32(buf, int(v.jointType))
	snapWriteTransform(buf, v.localFrameA)
	snapWriteTransform(buf, v.localFrameB)
	recWF32(buf, v.invMassA)
	recWF32(buf, v.invMassB)
	recWF32(buf, v.invIA)
	recWF32(buf, v.invIB)
	recWF32(buf, v.constraintHertz)
	recWF32(buf, v.constraintDampingRatio)
	snapWriteSoftness(buf, v.constraintSoftness)
	recWF32(buf, v.forceThreshold)
	recWF32(buf, v.torqueThreshold)

	switch v.jointType {
	case DistanceJointType:
		snapWriteDistanceJoint(buf, v.distanceJoint)
	case MotorJointType:
		snapWriteMotorJoint(buf, v.motorJoint)
	case MoverJointType:
		snapWriteMoverJoint(buf, v.moverJoint)
	case PogoJointType:
		snapWritePogoJoint(buf, v.pogoJoint)
	case PrismaticJointType:
		snapWritePrismaticJoint(buf, v.prismaticJoint)
	case RevoluteJointType:
		snapWriteRevoluteJoint(buf, v.revoluteJoint)
	case WeldJointType:
		snapWriteWeldJoint(buf, v.weldJoint)
	case WheelJointType:
		snapWriteWheelJoint(buf, v.wheelJoint)
	case FilterJointType:
		// Filter joints have no solver payload.
	}
}

func snapReadJointSim(r *snapReader, v *jointSim) {
	v.jointID = snapRI32(r)
	v.bodyIDA = snapRI32(r)
	v.bodyIDB = snapRI32(r)
	v.jointType = JointType(snapRI32(r))
	snapReadTransform(r, &v.localFrameA)
	snapReadTransform(r, &v.localFrameB)
	v.invMassA = math.Float32frombits(snapRU32(r))
	v.invMassB = math.Float32frombits(snapRU32(r))
	v.invIA = math.Float32frombits(snapRU32(r))
	v.invIB = math.Float32frombits(snapRU32(r))
	v.constraintHertz = math.Float32frombits(snapRU32(r))
	v.constraintDampingRatio = math.Float32frombits(snapRU32(r))
	snapReadSoftness(r, &v.constraintSoftness)
	v.forceThreshold = math.Float32frombits(snapRU32(r))

	v.torqueThreshold = math.Float32frombits(snapRU32(r))
	switch v.jointType {
	case DistanceJointType:
		snapReadDistanceJoint(r, &v.distanceJoint)
	case MotorJointType:
		snapReadMotorJoint(r, &v.motorJoint)
	case MoverJointType:
		snapReadMoverJoint(r, &v.moverJoint)
	case PogoJointType:
		snapReadPogoJoint(r, &v.pogoJoint)
	case PrismaticJointType:
		snapReadPrismaticJoint(r, &v.prismaticJoint)
	case RevoluteJointType:
		snapReadRevoluteJoint(r, &v.revoluteJoint)
	case WeldJointType:
		snapReadWeldJoint(r, &v.weldJoint)
	case WheelJointType:
		snapReadWheelJoint(r, &v.wheelJoint)
	case FilterJointType:
		// Filter joints have no solver payload.
	default:
		r.ok = false
	}
}

func snapWriteIslandSim(buf *recBuffer, v islandSim) {
	snapWI32(buf, v.islandID)
}

func snapReadIslandSim(r *snapReader, v *islandSim) {
	v.islandID = snapRI32(r)
}

func snapWriteContactLink(buf *recBuffer, v contactLink) {
	snapWI32(buf, v.contactID)
	snapWI32(buf, v.bodyIDA)
	snapWI32(buf, v.bodyIDB)
}

func snapReadContactLink(r *snapReader, v *contactLink) {
	v.contactID = snapRI32(r)
	v.bodyIDA = snapRI32(r)
	v.bodyIDB = snapRI32(r)
}

func snapWriteJointLink(buf *recBuffer, v jointLink) {
	snapWI32(buf, v.jointID)
	snapWI32(buf, v.bodyIDA)
	snapWI32(buf, v.bodyIDB)
}

func snapReadJointLink(r *snapReader, v *jointLink) {
	v.jointID = snapRI32(r)
	v.bodyIDA = snapRI32(r)
	v.bodyIDB = snapRI32(r)
}

func snapWriteVisitor(buf *recBuffer, v visitor) {
	snapWI32(buf, v.shapeID)
	recWU16(buf, v.generation)
}

func snapReadVisitor(r *snapReader, v *visitor) {
	v.shapeID = snapRI32(r)
	v.generation = snapReadU16(r)
}

func snapWriteTreeNode(buf *recBuffer, v TreeNode) {
	snapWriteAABB(buf, v.AABB)
	recWU32(buf, v.FlagIndex)

	if v.FlagIndex&leafNode != 0 {
		recWI32(buf, v.ShapeIndex)
	} else {
		recWI32(buf, v.Height)
	}
}

func snapReadTreeNode(r *snapReader, v *TreeNode) {
	snapReadAABB(r, &v.AABB)

	v.FlagIndex = snapRU32(r)
	if v.FlagIndex&leafNode != 0 {
		v.ShapeIndex = int32(snapRI32(r))
	} else {
		v.Height = int32(snapRI32(r))
	}
}

func snapWriteTreeProxy(buf *recBuffer, v TreeProxy) {
	recWU64(buf, v.UserData)
	recWU64(buf, v.CategoryBits)
	recWI32(buf, v.Node)
	recWI32(buf, v.Next)
}

func snapReadTreeProxy(r *snapReader, v *TreeProxy) {
	v.UserData = snapReadU64(r)
	v.CategoryBits = snapReadU64(r)
	v.Node = int32(snapRI32(r))
	v.Next = int32(snapRI32(r))
}

func snapWriteCapacity(buf *recBuffer, v Capacity) {
	snapWI32(buf, v.StaticShapeCount)
	snapWI32(buf, v.DynamicShapeCount)
	snapWI32(buf, v.StaticBodyCount)
	snapWI32(buf, v.DynamicBodyCount)
	snapWI32(buf, v.ContactCount)
}

func snapReadCapacity(r *snapReader, v *Capacity) {
	v.StaticShapeCount = snapRI32(r)
	v.DynamicShapeCount = snapRI32(r)
	v.StaticBodyCount = snapRI32(r)
	v.DynamicBodyCount = snapRI32(r)
	v.ContactCount = snapRI32(r)
}

func snapWriteSetItem(buf *recBuffer, v setItem) {
	recWU64(buf, v.key)
}

func snapReadSetItem(r *snapReader, v *setItem) {
	v.key = snapReadU64(r)
}

func snapWriteValue[T any](buf *recBuffer, value T) {
	switch v := any(value).(type) {
	case AABB:
		snapWriteAABB(buf, v)
	case Capacity:
		snapWriteCapacity(buf, v)
	case TreeNode:
		snapWriteTreeNode(buf, v)
	case TreeProxy:
		snapWriteTreeProxy(buf, v)
	case body:
		snapWriteBody(buf, v)
	case bodySim:
		snapWriteBodySim(buf, v)
	case bodyState:
		snapWriteBodyState(buf, v)
	case contact:
		snapWriteContact(buf, v)
	case contactLink:
		snapWriteContactLink(buf, v)
	case contactSim:
		snapWriteContactSim(buf, v)
	case int:
		recWI32(buf, int32(v))
	case int32:
		recWI32(buf, v)
	case islandSim:
		snapWriteIslandSim(buf, v)
	case joint:
		snapWriteJoint(buf, v)
	case jointLink:
		snapWriteJointLink(buf, v)
	case jointSim:
		snapWriteJointSim(buf, v)
	case setItem:
		snapWriteSetItem(buf, v)
	case shape:
		snapWriteShape(buf, v)
	case uint64:
		recWU64(buf, v)
	case visitor:
		snapWriteVisitor(buf, v)
	default:
		panic("box2d: unsupported snapshot element")
	}
}

func snapReadValue[T any](r *snapReader, value *T) {
	switch v := any(value).(type) {
	case *AABB:
		snapReadAABB(r, v)
	case *Capacity:
		snapReadCapacity(r, v)
	case *TreeNode:
		snapReadTreeNode(r, v)
	case *TreeProxy:
		snapReadTreeProxy(r, v)
	case *body:
		snapReadBody(r, v)
	case *bodySim:
		snapReadBodySim(r, v)
	case *bodyState:
		snapReadBodyState(r, v)
	case *contact:
		snapReadContact(r, v)
	case *contactLink:
		snapReadContactLink(r, v)
	case *contactSim:
		snapReadContactSim(r, v)
	case *int:
		*v = snapRI32(r)
	case *int32:
		*v = int32(snapRI32(r))
	case *islandSim:
		snapReadIslandSim(r, v)
	case *joint:
		snapReadJoint(r, v)
	case *jointLink:
		snapReadJointLink(r, v)
	case *jointSim:
		snapReadJointSim(r, v)
	case *setItem:
		snapReadSetItem(r, v)
	case *shape:
		snapReadShape(r, v)
	case *uint64:
		*v = snapReadU64(r)
	case *visitor:
		snapReadVisitor(r, v)
	default:
		panic("box2d: unsupported snapshot element")
	}
}

func snapMinSize[T any]() int {
	switch any(*new(T)).(type) {
	case AABB:
		return 16
	case Capacity:
		return 20
	case TreeNode:
		return 24
	case TreeProxy:
		return 24
	case body:
		return 93
	case bodySim:
		return 96
	case bodyState:
		return 32
	case contact:
		return 64
	case contactLink:
		return 12
	case contactSim:
		return 206
	case int:
		return 4
	case int32:
		return 4
	case islandSim:
		return 4
	case joint:
		return 59
	case jointLink:
		return 12
	case jointSim:
		return 92
	case setItem:
		return 8
	case shape:
		return 115
	case uint64:
		return 8
	case visitor:
		return 6
	default:
		panic("box2d: unsupported snapshot element")
	}
}

func snapReadU8(r *snapReader) uint8 { var b [1]byte; snapRBytes(r, b[:], 1); return b[0] }
func snapReadU16(r *snapReader) uint16 {
	var b [2]byte
	snapRBytes(r, b[:], 2)

	return binary.LittleEndian.Uint16(b[:])
}

func snapReadU64(r *snapReader) uint64 {
	var b [8]byte
	snapRBytes(r, b[:], 8)

	return binary.LittleEndian.Uint64(b[:])
}

func snapReadCount(r *snapReader, memSize, streamSize int) int {
	n := snapRI32(r)
	if !snapCheckCount(r, n, memSize, streamSize) {
		r.ok = false
	}

	return n
}

func snapshotHeaderError(image []byte, size int) error {
	if size < 16 || size > len(image) {
		return &SnapshotError{Err: ErrCorrupt}
	}

	if binary.LittleEndian.Uint32(image) != snapMagic {
		return &SnapshotError{Err: ErrBadMagic}
	}

	if binary.LittleEndian.Uint32(image[4:]) != snapVersion {
		return &SnapshotError{Err: ErrVersion}
	}

	expectedPrecision := uint32(0)
	if doublePrecision {
		expectedPrecision = snapFlagDoublePrecision
	}

	flags := binary.LittleEndian.Uint32(image[12:])
	if flags&snapFlagDoublePrecision != expectedPrecision || binary.LittleEndian.Uint32(image[8:]) != computeLayoutHash() {
		return &SnapshotError{Err: ErrLayout}
	}

	return nil
}

func snapTreeCountsValid(r *snapReader, nodeEnd, nodeCapacity, proxyCount, proxyCapacity int) bool {
	return nodeEnd >= 2 && nodeEnd&1 == 0 && nodeCapacity >= nodeEnd && proxyCount >= 0 && proxyCapacity >= 1 && proxyCapacity >= proxyCount && snapCheckCount(r, nodeEnd, int(unsafe.Sizeof(TreeNode{})), 28) && snapCheckCount(r, proxyCapacity, 24, 24)
}

func desChainShapes(r *snapReader, world *World) bool {
	n := snapReadCount(r, int(unsafe.Sizeof(chainShape{})), 18)
	if !r.ok {
		return false
	}

	world.chainShapes = make([]chainShape, n)
	for i := range world.chainShapes {
		c := &world.chainShapes[i]
		c.id = snapRI32(r)
		c.bodyID = snapRI32(r)
		c.nextChainID = snapRI32(r)
		c.segmentCount = snapRI32(r)

		c.generation = snapReadU16(r)
		if !r.ok {
			return false
		}

		if c.id != nullIndex {
			if !snapCheckCount(r, c.segmentCount, 8, 4) {
				r.ok = false
				return false
			}

			c.shapeIndices = make([]int, c.segmentCount)
			for j := range c.shapeIndices {
				c.shapeIndices[j] = snapRI32(r)
			}
		}
	}

	return true
}

func desSensors(r *snapReader, world *World) bool {
	n := snapReadCount(r, int(unsafe.Sizeof(sensor{})), 16)
	if !r.ok {
		return false
	}

	world.sensors = make([]sensor, n)
	for i := range world.sensors {
		s := &world.sensors[i]
		s.shapeID = snapRI32(r)
		desPodArray(r, &s.hits)
		desPodArray(r, &s.overlaps1)
		desPodArray(r, &s.overlaps2)

		if !r.ok {
			return false
		}
	}

	return true
}

func desIslands(r *snapReader, world *World) bool {
	n := snapReadCount(r, int(unsafe.Sizeof(island{})), 28)
	if !r.ok {
		return false
	}

	world.islands = make([]island, n)
	for i := range world.islands {
		s := &world.islands[i]
		s.setIndex = snapRI32(r)
		s.localIndex = snapRI32(r)
		s.islandID = snapRI32(r)
		s.constraintRemoveCount = snapRI32(r)
		desPodArray(r, &s.bodies)
		desPodArray(r, &s.contacts)
		desPodArray(r, &s.joints)

		if !r.ok {
			return false
		}
	}

	return true
}

func hashSnapshotJointImpulses(hash uint64, sim *jointSim) uint64 {
	switch sim.jointType {
	case DistanceJointType:
		hash = fnvMixFloat(hash, sim.distanceJoint.impulse)
		hash = fnvMixFloat(hash, sim.distanceJoint.lowerImpulse)
		hash = fnvMixFloat(hash, sim.distanceJoint.upperImpulse)
		hash = fnvMixFloat(hash, sim.distanceJoint.motorImpulse)
	case MoverJointType:
		hash = fnvMixFloat(hash, sim.moverJoint.linearVelocityImpulse.X)
		hash = fnvMixFloat(hash, sim.moverJoint.linearVelocityImpulse.Y)
	case MotorJointType:
		hash = fnvMixFloat(hash, sim.motorJoint.linearVelocityImpulse.X)
		hash = fnvMixFloat(hash, sim.motorJoint.linearVelocityImpulse.Y)
		hash = fnvMixFloat(hash, sim.motorJoint.angularVelocityImpulse)
		hash = fnvMixFloat(hash, sim.motorJoint.linearSpringImpulse.X)
		hash = fnvMixFloat(hash, sim.motorJoint.linearSpringImpulse.Y)
		hash = fnvMixFloat(hash, sim.motorJoint.angularSpringImpulse)
	case PogoJointType:
		hash = fnvMixFloat(hash, sim.pogoJoint.impulse)
		hash = fnvMixFloat(hash, sim.pogoJoint.velocity)
	case PrismaticJointType:
		{
			var b recBuffer
			recWVEC2(&b, sim.prismaticJoint.impulse)
			hash = fnvMixBytes(hash, b.data, 8)
		}

		hash = fnvMixFloat(hash, sim.prismaticJoint.springImpulse)
		hash = fnvMixFloat(hash, sim.prismaticJoint.motorImpulse)
		hash = fnvMixFloat(hash, sim.prismaticJoint.lowerImpulse)
		hash = fnvMixFloat(hash, sim.prismaticJoint.upperImpulse)
	case RevoluteJointType:
		{
			var b recBuffer
			recWVEC2(&b, sim.revoluteJoint.linearImpulse)
			hash = fnvMixBytes(hash, b.data, 8)
		}

		hash = fnvMixFloat(hash, sim.revoluteJoint.springImpulse)
		hash = fnvMixFloat(hash, sim.revoluteJoint.motorImpulse)
		hash = fnvMixFloat(hash, sim.revoluteJoint.lowerImpulse)
		hash = fnvMixFloat(hash, sim.revoluteJoint.upperImpulse)
	case WeldJointType:
		{
			var b recBuffer
			recWVEC2(&b, sim.weldJoint.linearImpulse)
			hash = fnvMixBytes(hash, b.data, 8)
		}

		hash = fnvMixFloat(hash, sim.weldJoint.angularImpulse)
	case WheelJointType:
		hash = fnvMixFloat(hash, sim.wheelJoint.perpImpulse)
		hash = fnvMixFloat(hash, sim.wheelJoint.motorImpulse)
		hash = fnvMixFloat(hash, sim.wheelJoint.springImpulse)
		hash = fnvMixFloat(hash, sim.wheelJoint.lowerImpulse)
		hash = fnvMixFloat(hash, sim.wheelJoint.upperImpulse)
	case FilterJointType:
		// Filter joints have no solver impulses.
	}

	return hash
}
