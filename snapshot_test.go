// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zigai/box2d/internal/oracle"
)

type snapshotIDs struct {
	body  Body
	shape Shape
	joint Joint
	chain Chain
}

func snapshotTreeTailMatches(data []byte, tail []int32) bool {
	for i, v := range tail {
		if binary.LittleEndian.Uint32(data[4*i:]) != uint32(v) {
			return false
		}
	}

	return true
}

func snapshotBuildScene(t *testing.T, outIDs *snapshotIDs) *World {
	t.Helper()

	var heldJoint Joint

	var heldChain Chain

	def := DefaultWorldDef()
	def.WorkerCount = 1
	worldID := NewWorld(&def)
	CleanupWorld(t, worldID)
	{
		bd := DefaultBodyDef()
		bd.Position = Pos{X: 0.0, Y: -1.0}
		groundID := worldID.CreateBody(&bd)
		groundBox := MakeBox(40.0, 1.0)
		sd := DefaultShapeDef()
		groundID.CreatePolygonShape(&sd, &groundBox)
	}

	stackTop := Body{}
	{
		sd := DefaultShapeDef()
		box := MakeBox(0.5, 0.5)

		for i := range 8 {
			bd := DefaultBodyDef()
			bd.Type = DynamicBody
			bd.Position = Pos{X: 0.0, Y: oracle.Real(float32(0.5) + float32(float32(i)*float32(1.1)))}
			bodyID := worldID.CreateBody(&bd)
			bodyID.CreatePolygonShape(&sd, &box)
			stackTop = bodyID
		}
	}

	jbDef := DefaultBodyDef()
	jbDef.Type = DynamicBody
	jbDef.Position = Pos{X: 5.0, Y: 2.0}
	jbA := worldID.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 7.0, Y: 2.0}
	jbB := worldID.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 9.0, Y: 2.0}
	jbC := worldID.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 11.0, Y: 2.0}
	jbD := worldID.CreateBody(&jbDef)
	jbox := MakeBox(0.3, 0.3)
	jsd := DefaultShapeDef()
	heldShape := jbA.CreatePolygonShape(&jsd, &jbox)
	jbB.CreatePolygonShape(&jsd, &jbox)
	jbC.CreatePolygonShape(&jsd, &jbox)
	jbD.CreatePolygonShape(&jsd, &jbox)
	{
		rd := DefaultRevoluteJointDef()
		rd.EnableLimit = true
		rd.LowerAngle = float32(-0.1) * Pi
		rd.UpperAngle = float32(0.2) * Pi
		rd.EnableSpring = true
		rd.Hertz = 1.0
		rd.DampingRatio = 1.0
		rd.EnableMotor = true
		rd.MaxMotorTorque = 0.5
		rd.Base.BodyA = jbA
		rd.Base.BodyB = jbB
		rd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		rd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		heldJoint = worldID.CreateRevoluteJoint(&rd).Joint
	}
	{
		pd := DefaultPrismaticJointDef()
		pd.EnableLimit = true
		pd.LowerTranslation = -0.5
		pd.UpperTranslation = 0.5
		pd.Base.BodyA = jbB
		pd.Base.BodyB = jbC
		pd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		pd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		worldID.CreatePrismaticJoint(&pd)
	}
	{
		dd := DefaultDistanceJointDef()
		dd.Length = 2.0
		dd.Base.BodyA = jbC
		dd.Base.BodyB = jbD
		dd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		dd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		worldID.CreateDistanceJoint(&dd)
	}
	{
		wd := DefaultWeldJointDef()
		wd.LinearHertz = 5.0
		wd.LinearDampingRatio = 0.7
		wd.Base.BodyA = jbD
		wd.Base.BodyB = stackTop
		wd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		wd.Base.LocalFrameB.P = Vec2{X: 0.0, Y: 0.0}
		worldID.CreateWeldJoint(&wd)
	}
	{
		cbd := DefaultBodyDef()
		cbd.Position = Pos{X: -10.0, Y: 0.0}
		chainBodyID := worldID.CreateBody(&cbd)
		chainPoints := []Vec2{{X: -2}, {}, {X: 2}}
		chainMat := DefaultSurfaceMaterial()
		chainMat.Friction = 0.4
		chainDef := DefaultChainDef()
		chainDef.Points = chainPoints
		chainDef.Ghost1 = Vec2{X: -4.0, Y: 0.0}
		chainDef.Ghost2 = Vec2{X: 4.0, Y: 2.0}
		chainDef.Materials = []SurfaceMaterial{chainMat}
		chainDef.IsLoop = false
		heldChain = chainBodyID.CreateChain(&chainDef)
	}
	{
		sbd := DefaultBodyDef()
		sbd.Position = Pos{X: 0.0, Y: 5.0}
		sensorBodyID := worldID.CreateBody(&sbd)
		sensorBox := MakeBox(3.0, 3.0)
		sensorDef := DefaultShapeDef()
		sensorDef.IsSensor = true
		sensorDef.EnableSensorEvents = true
		sensorBodyID.CreatePolygonShape(&sensorDef, &sensorBox)
	}
	{
		sd := DefaultShapeDef()
		box := MakeBox(0.5, 0.5)

		for i := range 6 {
			bd := DefaultBodyDef()
			bd.Type = DynamicBody
			bd.Position = Pos{X: 40.0, Y: oracle.Real(float32(0.5) + float32(float32(i)*float32(1.1)))}
			bodyID := worldID.CreateBody(&bd)
			bodyID.CreatePolygonShape(&sd, &box)
		}
	}
	{
		bd := DefaultBodyDef()
		bd.Position = Pos{X: -30.0, Y: 10.0}
		scrapID := worldID.CreateBody(&bd)
		scrapBox := MakeBox(0.5, 0.5)
		sd := DefaultShapeDef()
		scrapID.CreatePolygonShape(&sd, &scrapBox)
		scrapID.Destroy()
	}

	if outIDs != nil {
		outIDs.body = stackTop
		outIDs.shape = heldShape
		outIDs.joint = heldJoint
		outIDs.chain = heldChain
	}

	return worldID
}

func snapshotStepUntilSleep(r *oracle.Reader, worldID *World) {
	dt := float32(1.0) / 60.0
	subSteps := 4

	maxSteps := 500
	for step := range maxSteps {
		worldID.Step(dt, subSteps)
		traceTransforms(r, worldID, fmt.Sprintf("L227/step/%d", step), step)

		awake := worldID.AwakeBodyCount()
		if awake == 0 {
			return
		}
	}
}

func TestSnapshot(t *testing.T) {
	r := oracle.Open(t, "snapshot/SnapshotTest")
	snapPath := filepath.Join(t.TempDir(), "test_snapshot_midstream.b2rec")

	var buf recBuffer

	worldAId, worldBId, worldA1Id, worldCId := snapshotCloneParity(t, r, &buf)
	rID, sID, freshID := snapshotRestoreParity(t, r)
	snapshotRecording(t, r, snapPath)
	snapshotMovingContinuation(t, r)
	snapshotReplayRestart(t, r)
	snapshotClearsEndEvents(t, r)

	_ = os.Remove(snapPath)

	worldAId.Close()
	worldBId.Close()
	worldA1Id.Close()
	worldCId.Close()
	rID.Close()
	sID.Close()
	freshID.Close()
	recBufFree(&buf)
}

func snapshotCloneParity(t *testing.T, r *oracle.Reader, buf *recBuffer) (*World, *World, *World, *World) {
	t.Helper()
	worldAId := snapshotBuildScene(t, nil)
	snapshotStepUntilSleep(r, worldAId)

	worldA := worldAId
	if len(worldA.solverSets) <= 3 {
		t.Fatal("condition false")
	}

	r.Bool("L247", len(worldA.solverSets) > 3)
	r.I32("L247/value0", int32(len(worldA.solverSets)))

	if worldA.broadPhase.trees[StaticBody].pairFreeList == int32(nullIndex) {
		t.Fatal("condition false")
	}

	r.Bool("L250", worldA.broadPhase.trees[StaticBody].pairFreeList != int32(nullIndex))
	r.I32("L250/value0", int32(worldA.broadPhase.trees[StaticBody].pairFreeList))
	r.I32("L250/value1", int32(nullIndex))

	serializeWorld(worldA, buf)

	if buf.size <= 0 {
		t.Fatal("condition false")
	}

	r.Bool("L255", buf.size > 0)
	worldBId, err := NewWorldFromSnapshot(buf.data, 1)
	CleanupWorld(t, worldBId)

	if err != nil || !worldBId.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L259", worldBId.IsValid())
	worldB := worldBId
	snapshotCloneHashes(t, r, worldA, worldB)

	snapshotCloneTrees(t, r, worldA, worldB)

	snapshotCloneContinuation(t, r, worldAId, worldBId)

	recBufFree(buf)
	*buf = recBuffer{}
	serializeWorld(worldA, buf)
	worldA1Id, err := NewWorldFromSnapshot(buf.data, 1)
	CleanupWorld(t, worldA1Id)

	if err != nil || !worldA1Id.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L322", worldA1Id.IsValid())
	worldA1 := worldA1Id
	worldCId, err := NewWorldFromSnapshot(buf.data, 4)
	CleanupWorld(t, worldCId)

	if err != nil || !worldCId.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L326", worldCId.IsValid())

	worldC := worldCId
	if hashWorldState(worldA1) != hashWorldState(worldC) {
		t.Fatal("condition false")
	}

	r.Bool("L329", hashWorldState(worldA1) == hashWorldState(worldC))

	snapshotWorkerContinuation(t, r, worldA1Id, worldCId)

	return worldAId, worldBId, worldA1Id, worldCId
}

func snapshotRestoreParity(t *testing.T, r *oracle.Reader) (*World, *World, *World) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4

	var ids snapshotIDs

	rID := snapshotBuildScene(t, &ids)
	snapshotStepUntilSleep(r, rID)
	rWorld := rID
	image := MarshalSnapshot(rID)

	imageSize := len(image)
	if imageSize <= 0 {
		t.Fatal("condition false")
	}

	r.Bool("L354", imageSize > 0)

	image = MarshalSnapshot(rID)

	written := len(image)
	if written != imageSize {
		t.Fatal("condition false")
	}

	r.Bool("L357", written == imageSize)

	snapHash := hashWorldStateDeep(rWorld)
	r.U64("L359/value", snapHash)
	ids.body.SetLinearVelocity(Vec2{X: 3.0, Y: 6.0})
	ids.body.SetAwake(true)

	postDef := DefaultBodyDef()
	postDef.Type = DynamicBody
	postDef.Position = Pos{X: 20.0, Y: 20.0}

	postBody := rID.CreateBody(&postDef)
	for step := range 30 {
		rID.Step(dt, subSteps)
		traceTransforms(r, rID, fmt.Sprintf("L370/step/%d", step), step)
	}

	if hashWorldStateDeep(rWorld) == snapHash {
		t.Fatal("condition false")
	}

	r.Bool("L372", hashWorldStateDeep(rWorld) != snapHash)

	if rID.UnmarshalBinary(image) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L374", true)

	if hashWorldStateDeep(rWorld) != snapHash {
		t.Fatal("condition false")
	}

	r.Bool("L377", hashWorldStateDeep(rWorld) == snapHash)

	snapshotRestoredHandles(t, r, rID, &ids, postBody)
	snapshotRejectedCorruption(t, r, rID, image)

	snapshotTreeCapacity(t, r, rID, image, snapHash)

	sID, freshID := snapshotRestoredContinuation(t, r, image)

	return rID, sID, freshID
}

func snapshotRecording(t *testing.T, r *oracle.Reader, snapPath string) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4
	wID := snapshotBuildScene(t, nil)

	for step := range 20 {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L469/step/%d", step), step)
	}

	rec := NewRecording(0)
	wID.StartRecording(rec)

	for step := range 60 {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L476/step/%d", step), step)
	}

	wID.StopRecording()
	wID.Close()

	recData := rec.Data()

	recSize := rec.Size()
	if ValidateReplay(recData[:recSize], 0) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L483", true)

	if ValidateReplay(recData[:recSize], 4) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L484", true)

	if RecordingSaveFile(rec, snapPath) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L487", true)

	loaded, err := RecordingLoadFile(snapPath)
	if err != nil || loaded == nil {
		t.Fatal("condition false")
	}

	r.Bool("L489", loaded != nil)

	if ValidateReplay(loaded.Data(), 0) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L490", true)

	snapshotReplayPlayback(t, r, recData)
}

func snapshotReplayPlayback(t *testing.T, r *oracle.Reader, recData []byte) {
	t.Helper()

	player, err := NewReplay(recData, 0)

	destroyPlayer := CleanupReplay(t, player)
	if err != nil || player == nil {
		t.Fatal("condition false")
	}

	r.Bool("L495", player != nil)
	pid0 := player.World()

	frames := 0
	for player.StepFrame() {
		frames += 1
		traceTransforms(r, player.World(), fmt.Sprintf("L501/step/%d", frames), frames-1)
	}

	if frames != 60 {
		t.Fatal("condition false")
	}

	r.Bool("L503", frames == 60)
	r.I32("L503/value0", int32(frames))
	r.I32("L503/value0", int32(frames))

	if player.HasDiverged() != false {
		t.Fatal("condition false")
	}

	r.Bool("L504", player.HasDiverged() == false)
	player.Restart()

	pid1 := player.World()
	if pid0 != pid1 {
		t.Fatal("condition false")
	}

	r.Bool("L508", pid0 == pid1)

	if player.Frame() != 0 {
		t.Fatal("condition false")
	}

	r.Bool("L509", player.Frame() == 0)
	r.I32("L509/value0", int32(player.Frame()))
	r.I32("L509/value0", int32(player.Frame()))

	frames2 := 0
	for player.StepFrame() {
		frames2 += 1
		traceTransforms(r, player.World(), fmt.Sprintf("L514/step/%d", frames2), frames2-1)
	}

	if frames2 != 60 {
		t.Fatal("condition false")
	}

	r.Bool("L516", frames2 == 60)
	r.I32("L516/value0", int32(frames2))
	r.I32("L516/value0", int32(frames2))

	if player.HasDiverged() != false {
		t.Fatal("condition false")
	}

	r.Bool("L517", player.HasDiverged() == false)
	destroyPlayer()
}

func snapshotMovingContinuation(t *testing.T, r *oracle.Reader) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4
	wID := snapshotBuildScene(t, nil)

	for step := range 20 {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L531/step/%d", step), step)
	}

	snap := MarshalSnapshot(wID)
	origin := wID

	const tailSteps = 90

	var (
		realTail     [tailSteps]uint64
		realDeepTail [tailSteps]uint64
	)
	for step := range tailSteps {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L547/step/%d", step), step)
		r.U64(fmt.Sprintf("moving/origin/%d/state", step), wID.StateHash())
		realTail[step] = hashWorldState(origin)
		realDeepTail[step] = wID.StateHash()
	}

	cID, err := NewWorldFromSnapshot(snap, 1)
	CleanupWorld(t, cID)

	if err != nil || !cID.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L552", cID.IsValid())

	clone := cID
	for step := range tailSteps {
		cID.Step(dt, subSteps)
		traceTransforms(r, cID, fmt.Sprintf("L556/step/%d", step), step)
		r.U64(fmt.Sprintf("moving/clone/%d/state", step), cID.StateHash())

		if cID.StateHash() != realDeepTail[step] {
			t.Fatalf("moving continuation canonical state at step %d: got %x, uninterrupted %x", step, cID.StateHash(), realDeepTail[step])
		}

		if realTail[step] != hashWorldState(clone) {
			r.Bool("L560", false)

			if !false {
				t.Fatal("condition false")
			}
		}
	}

	wID.Close()
	cID.Close()
}

func snapshotReplayRestart(t *testing.T, r *oracle.Reader) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4
	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	wID := NewWorld(&wd)
	CleanupWorld(t, wID)

	rec := NewRecording(0)
	wID.StartRecording(rec)

	bd := DefaultBodyDef()
	bd.Type = DynamicBody
	bd.Position = Pos{X: 0.0, Y: 8.0}
	fallingBody := wID.CreateBody(&bd)
	box := MakeBox(0.5, 0.5)
	sdef := DefaultShapeDef()
	fallingBody.CreatePolygonShape(&sdef, &box)

	for step := range 30 {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L589/step/%d", step), step)
	}

	wID.StopRecording()
	wID.Close()

	player, err := NewReplay(rec.Data(), 0)

	destroyPlayer := CleanupReplay(t, player)
	if err != nil || player == nil {
		t.Fatal("condition false")
	}

	r.Bool("L595", player != nil)

	pid0 := player.World()
	for i := range 5 {
		player.StepFrame()
		traceTransforms(r, player.World(), fmt.Sprintf("L599/step/%d", i), i)
	}

	player.Restart()

	pid1 := player.World()
	if pid0 != pid1 {
		t.Fatal("condition false")
	}

	r.Bool("L603", pid0 == pid1)

	if player.Frame() != 0 {
		t.Fatal("condition false")
	}

	r.Bool("L604", player.Frame() == 0)
	r.I32("L604/value0", int32(player.Frame()))
	r.I32("L604/value0", int32(player.Frame()))
	destroyPlayer()
}

func snapshotClearsEndEvents(t *testing.T, r *oracle.Reader) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4
	wd := DefaultWorldDef()
	wd.WorkerCount = 1
	wID := NewWorld(&wd)
	CleanupWorld(t, wID)
	w := wID
	gbd := DefaultBodyDef()
	gbd.Position = Pos{X: 0.0, Y: -1.0}
	ground := wID.CreateBody(&gbd)
	groundBox := MakeBox(10.0, 1.0)
	gsd := DefaultShapeDef()
	gsd.EnableContactEvents = true
	ground.CreatePolygonShape(&gsd, &groundBox)

	bd := DefaultBodyDef()
	bd.Type = DynamicBody
	bd.Position = Pos{X: 0.0, Y: 0.5}
	box := wID.CreateBody(&bd)
	boxPoly := MakeBox(0.5, 0.5)
	bsd := DefaultShapeDef()
	bsd.EnableContactEvents = true
	box.CreatePolygonShape(&bsd, &boxPoly)

	for step := range 60 {
		wID.Step(dt, subSteps)
		traceTransforms(r, wID, fmt.Sprintf("L637/step/%d", step), step)
	}

	if wID.StateHash() != hashWorldStateDeep(w) {
		t.Fatal("condition false")
	}

	r.Bool("L641", wID.StateHash() == hashWorldStateDeep(w))
	snap := MarshalSnapshot(wID)

	box.Disable()

	if len(w.contactEndEvents[w.endEventArrayIndex]) < 1 {
		t.Fatal("condition false")
	}

	r.Bool("L650", len(w.contactEndEvents[w.endEventArrayIndex]) >= 1)
	r.I32("L650/value0", int32(len(w.contactEndEvents[w.endEventArrayIndex])))

	if wID.UnmarshalBinary(snap) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L652", true)

	if len(w.contactEndEvents[0]) != 0 {
		t.Fatal("condition false")
	}

	r.Bool("L655", len(w.contactEndEvents[0]) == 0)
	r.I32("L655/value0", int32(len(w.contactEndEvents[0])))

	if len(w.contactEndEvents[1]) != 0 {
		t.Fatal("condition false")
	}

	r.Bool("L656", len(w.contactEndEvents[1]) == 0)
	r.I32("L656/value0", int32(len(w.contactEndEvents[1])))
	wID.Step(dt, subSteps)
	traceTransforms(r, wID, "L659/step/0", 0)

	endCount := 0
	for range wID.ContactEndEvents() {
		endCount++
	}

	if endCount != 0 {
		t.Fatal("condition false")
	}

	r.Bool("L661", endCount == 0)
	r.I32("L661/value0", int32(endCount))
	r.I32("L661/value0", int32(endCount))
	wID.Close()
}

func snapshotCloneTrees(t *testing.T, r *oracle.Reader, worldA, worldB *World) {
	t.Helper()

	for treeType := range int(BodyTypeCount) {
		treeA := &worldA.broadPhase.trees[treeType]

		treeB := &worldB.broadPhase.trees[treeType]
		if treeA.nodeEnd != treeB.nodeEnd {
			t.Fatal("condition false")
		}

		r.Bool("L278", treeA.nodeEnd == treeB.nodeEnd)
		r.I32("L278/value0", int32(treeA.nodeEnd))
		r.I32("L278/value1", int32(treeB.nodeEnd))

		if treeA.pairFreeList != treeB.pairFreeList {
			t.Fatal("condition false")
		}

		r.Bool("L279", treeA.pairFreeList == treeB.pairFreeList)
		r.I32("L279/value0", int32(treeA.pairFreeList))
		r.I32("L279/value1", int32(treeB.pairFreeList))

		for pair := treeA.pairFreeList; pair != int32(nullIndex); pair = treeA.parents[pair] {
			if treeA.parents[pair] != treeB.parents[pair] {
				t.Fatal("condition false")
			}

			r.Bool("L283", treeA.parents[pair] == treeB.parents[pair])
			r.I32("L283/value0", int32(treeA.parents[pair]))
			r.I32("L283/value1", int32(treeB.parents[pair]))

			if treeA.parents[pair+1] != treeB.parents[pair+1] {
				t.Fatal("condition false")
			}

			r.Bool("L284", treeA.parents[pair+1] == treeB.parents[pair+1])
			r.I32("L284/value0", int32(treeA.parents[pair+1]))
			r.I32("L284/value1", int32(treeB.parents[pair+1]))
		}
	}
}

func snapshotTreeCapacity(t *testing.T, r *oracle.Reader, rID *World, image []byte, snapHash uint64) {
	t.Helper()

	rWorld := rID
	imageSize := len(image)
	staticTree := &rWorld.broadPhase.trees[StaticBody]
	treeHead := int32(staticTree.nodeEnd)

	treeTail := []int32{staticTree.pairFreeList, staticTree.proxyCount, staticTree.proxyCapacity, staticTree.proxyFreeList, 0}
	if staticTree.dfsOrdered {
		treeTail[4] = 1
	}

	treeOffset := -1

	for i := 0; i+28 <= imageSize; i++ {
		if binary.LittleEndian.Uint32(image[i:]) == uint32(treeHead) && snapshotTreeTailMatches(image[i+8:i+28], treeTail) {
			treeOffset = i
			break
		}
	}

	if treeOffset < 0 {
		t.Fatal("condition false")
	}

	r.Bool("L416", treeOffset >= 0)

	greedy := make([]byte, imageSize)
	copy(greedy, image)

	hugeCapacity := 2147483647 / 64
	binary.LittleEndian.PutUint32(greedy[treeOffset+4:], uint32(hugeCapacity))

	if rID.UnmarshalBinary(greedy) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L422", true)

	if hashWorldStateDeep(rWorld) != snapHash {
		t.Fatal("condition false")
	}

	r.Bool("L423", hashWorldStateDeep(rWorld) == snapHash)

	if staticTree.nodeCapacity > 2*staticTree.nodeEnd {
		t.Fatal("condition false")
	}

	r.Bool("L424", staticTree.nodeCapacity <= 2*staticTree.nodeEnd)
	r.I32("L424/value0", int32(staticTree.nodeCapacity))
	r.I32("L424/value1", int32(2*staticTree.nodeEnd))

	for range 3 {
		if rID.UnmarshalBinary(image) != nil {
			t.Fatal("condition false")
		}

		r.Bool("L430", true)
	}

	if hashWorldStateDeep(rWorld) != snapHash {
		t.Fatal("condition false")
	}

	r.Bool("L432", hashWorldStateDeep(rWorld) == snapHash)
}

func snapshotCloneHashes(t *testing.T, r *oracle.Reader, worldA, worldB *World) {
	t.Helper()

	hashA0 := hashWorldState(worldA)
	r.U64("L264/value", hashA0)

	hashB0 := hashWorldState(worldB)
	r.U64("L265/value", hashB0)

	if hashA0 != hashB0 {
		t.Fatal("condition false")
	}

	r.Bool("L266", hashA0 == hashB0)

	deepA0 := hashWorldStateDeep(worldA)
	r.U64("L268/value", deepA0)

	deepB0 := hashWorldStateDeep(worldB)
	r.U64("L269/value", deepB0)

	if deepA0 != deepB0 {
		t.Fatal("condition false")
	}

	r.Bool("L270", deepA0 == deepB0)
}

func snapshotCloneContinuation(t *testing.T, r *oracle.Reader, worldAId, worldBId *World) {
	t.Helper()

	worldA, worldB := worldAId, worldBId
	dt := float32(1.0) / 60.0

	subSteps := 4
	for step := range 120 {
		worldAId.Step(dt, subSteps)
		traceTransforms(r, worldAId, fmt.Sprintf("L293/step/%d", step), step)
		worldBId.Step(dt, subSteps)
		traceTransforms(r, worldBId, fmt.Sprintf("L294/step/%d", step), step)

		sA := hashWorldState(worldA)
		r.U64("L296/value", sA)

		sB := hashWorldState(worldB)
		r.U64("L297/value", sB)

		if sA != sB {
			r.Bool("L302", false)

			if !false {
				t.Fatal("condition false")
			}
		}

		dA := hashWorldStateDeep(worldA)
		r.U64("L305/value", dA)

		dB := hashWorldStateDeep(worldB)
		r.U64("L306/value", dB)

		if dA != dB {
			r.Bool("L311", false)

			if !false {
				t.Fatal("condition false")
			}
		}
	}
}

func snapshotWorkerContinuation(t *testing.T, r *oracle.Reader, worldA1Id, worldCId *World) {
	t.Helper()

	worldA1, worldC := worldA1Id, worldCId
	dt := float32(1.0) / 60.0
	subSteps := 4

	for step := range 120 {
		worldA1Id.Step(dt, subSteps)
		traceTransforms(r, worldA1Id, fmt.Sprintf("L333/step/%d", step), step)
		worldCId.Step(dt, subSteps)
		traceTransforms(r, worldCId, fmt.Sprintf("L334/step/%d", step), step)

		sA1 := hashWorldState(worldA1)
		r.U64("L336/value", sA1)

		sC := hashWorldState(worldC)
		r.U64("L337/value", sC)

		if sA1 != sC {
			r.Bool("L342", false)

			if !false {
				t.Fatal("condition false")
			}
		}
	}
}

func snapshotRestoredHandles(t *testing.T, r *oracle.Reader, rID *World, ids *snapshotIDs, postBody Body) {
	t.Helper()

	if !rID.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L380", rID.IsValid())

	if !ids.body.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L381", ids.body.IsValid())

	if !ids.shape.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L382", ids.shape.IsValid())

	if !ids.joint.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L383", ids.joint.IsValid())

	if !ids.chain.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L384", ids.chain.IsValid())

	if postBody.IsValid() != false {
		t.Fatal("condition false")
	}

	r.Bool("L387", postBody.IsValid() == false)
}

func snapshotRejectedCorruption(t *testing.T, r *oracle.Reader, rID *World, image []byte) {
	t.Helper()

	rWorld := rID
	imageSize := len(image)
	preBadHash := hashWorldStateDeep(rWorld)
	r.U64("L390/value", preBadHash)

	if rID.UnmarshalBinary(nil) == nil {
		t.Fatal("condition false")
	}

	r.Bool("L391", true)

	corrupt := make([]byte, imageSize)
	copy(corrupt, image)

	corrupt[0] ^= 0xFF
	if rID.UnmarshalBinary(corrupt) == nil {
		t.Fatal("condition false")
	}

	r.Bool("L395", true)

	if hashWorldStateDeep(rWorld) != preBadHash {
		t.Fatal("condition false")
	}

	r.Bool("L396", hashWorldStateDeep(rWorld) == preBadHash)
}

func snapshotRestoredContinuation(t *testing.T, r *oracle.Reader, image []byte) (*World, *World) {
	t.Helper()

	dt := float32(1.0) / 60.0
	subSteps := 4
	def4 := DefaultWorldDef()
	def4.WorkerCount = 4
	sID := NewWorld(&def4)
	CleanupWorld(t, sID)

	if sID.UnmarshalBinary(image) != nil {
		t.Fatal("condition false")
	}

	r.Bool("L439", true)

	sWorld := sID
	freshID, err := NewWorldFromSnapshot(image, 1)
	CleanupWorld(t, freshID)

	if err != nil || !freshID.IsValid() {
		t.Fatal("condition false")
	}

	r.Bool("L443", freshID.IsValid())

	freshWorld := freshID
	if hashWorldState(sWorld) != hashWorldState(freshWorld) {
		t.Fatal("condition false")
	}

	r.Bool("L446", hashWorldState(sWorld) == hashWorldState(freshWorld))

	for step := range 120 {
		sID.Step(dt, subSteps)
		traceTransforms(r, sID, fmt.Sprintf("L449/step/%d", step), step)
		freshID.Step(dt, subSteps)
		traceTransforms(r, freshID, fmt.Sprintf("L450/step/%d", step), step)
		r.U64(fmt.Sprintf("continuation/in-place/%d/state", step), sID.StateHash())
		r.U64(fmt.Sprintf("continuation/fresh/%d/state", step), freshID.StateHash())

		if sID.StateHash() != freshID.StateHash() {
			t.Fatalf("in-place versus fresh canonical state at step %d", step)
		}

		if hashWorldState(sWorld) != hashWorldState(freshWorld) {
			r.Bool("L454", false)

			if !false {
				t.Fatal("condition false")
			}
		}
	}

	return sID, freshID
}
