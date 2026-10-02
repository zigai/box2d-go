// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"encoding/binary"
	"fmt"
	"github.com/zigai/box2d/internal/oracle"
	"github.com/zigai/box2d/internal/porttest"
	"os"
	"path/filepath"
	"testing"
)

type snapshotSnapshotIDs struct {
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
func snapshotBuildScene(workerCount int, outIds *snapshotSnapshotIDs) World {
	heldJoint := Joint{}
	heldChain := Chain{}
	def := DefaultWorldDef()
	def.WorkerCount = workerCount
	worldId := CreateWorld(&def)
	{
		bd := DefaultBodyDef()
		bd.Position = Pos{X: 0.0, Y: -1.0}
		groundId := worldId.CreateBody(&bd)
		groundBox := MakeBox(40.0, 1.0)
		sd := DefaultShapeDef()
		groundId.CreatePolygonShape(&sd, &groundBox)
	}
	stackTop := Body{}
	{
		sd := DefaultShapeDef()
		box := MakeBox(0.5, 0.5)
		for i := 0; i < 8; i++ {
			bd := DefaultBodyDef()
			bd.Type = DynamicBody
			bd.Position = Pos{X: 0.0, Y: oracle.Real(float32(0.5) + float32(float32(i)*float32(1.1)))}
			bodyId := worldId.CreateBody(&bd)
			bodyId.CreatePolygonShape(&sd, &box)
			stackTop = bodyId
		}
	}
	jbDef := DefaultBodyDef()
	jbDef.Type = DynamicBody
	jbDef.Position = Pos{X: 5.0, Y: 2.0}
	jbA := worldId.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 7.0, Y: 2.0}
	jbB := worldId.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 9.0, Y: 2.0}
	jbC := worldId.CreateBody(&jbDef)
	jbDef.Position = Pos{X: 11.0, Y: 2.0}
	jbD := worldId.CreateBody(&jbDef)
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
		rd.Base.BodyIDA = jbA
		rd.Base.BodyIDB = jbB
		rd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		rd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		heldJoint = worldId.CreateRevoluteJoint(&rd).Joint
	}
	{
		pd := DefaultPrismaticJointDef()
		pd.EnableLimit = true
		pd.LowerTranslation = -0.5
		pd.UpperTranslation = 0.5
		pd.Base.BodyIDA = jbB
		pd.Base.BodyIDB = jbC
		pd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		pd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		worldId.CreatePrismaticJoint(&pd)
	}
	{
		dd := DefaultDistanceJointDef()
		dd.Length = 2.0
		dd.Base.BodyIDA = jbC
		dd.Base.BodyIDB = jbD
		dd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		dd.Base.LocalFrameB.P = Vec2{X: -0.3, Y: 0.0}
		worldId.CreateDistanceJoint(&dd)
	}
	{
		wd := DefaultWeldJointDef()
		wd.LinearHertz = 5.0
		wd.LinearDampingRatio = 0.7
		wd.Base.BodyIDA = jbD
		wd.Base.BodyIDB = stackTop
		wd.Base.LocalFrameA.P = Vec2{X: 0.3, Y: 0.0}
		wd.Base.LocalFrameB.P = Vec2{X: 0.0, Y: 0.0}
		worldId.CreateWeldJoint(&wd)
	}
	{
		cbd := DefaultBodyDef()
		cbd.Position = Pos{X: -10.0, Y: 0.0}
		chainBodyId := worldId.CreateBody(&cbd)
		chainPoints := []Vec2{{X: -2}, {}, {X: 2}}
		chainMat := DefaultSurfaceMaterial()
		chainMat.Friction = 0.4
		chainDef := DefaultChainDef()
		chainDef.Points = chainPoints
		chainDef.PointCount = 3
		chainDef.Ghost1 = Vec2{X: -4.0, Y: 0.0}
		chainDef.Ghost2 = Vec2{X: 4.0, Y: 2.0}
		chainDef.Materials = []SurfaceMaterial{chainMat}
		chainDef.MaterialCount = 1
		chainDef.IsLoop = false
		heldChain = chainBodyId.CreateChain(&chainDef)
	}
	{
		sbd := DefaultBodyDef()
		sbd.Position = Pos{X: 0.0, Y: 5.0}
		sensorBodyId := worldId.CreateBody(&sbd)
		sensorBox := MakeBox(3.0, 3.0)
		sensorDef := DefaultShapeDef()
		sensorDef.IsSensor = true
		sensorDef.EnableSensorEvents = true
		sensorBodyId.CreatePolygonShape(&sensorDef, &sensorBox)
	}
	{
		sd := DefaultShapeDef()
		box := MakeBox(0.5, 0.5)
		for i := 0; i < 6; i++ {
			bd := DefaultBodyDef()
			bd.Type = DynamicBody
			bd.Position = Pos{X: 40.0, Y: oracle.Real(float32(0.5) + float32(float32(i)*float32(1.1)))}
			bodyId := worldId.CreateBody(&bd)
			bodyId.CreatePolygonShape(&sd, &box)
		}
	}
	{
		bd := DefaultBodyDef()
		bd.Position = Pos{X: -30.0, Y: 10.0}
		scrapId := worldId.CreateBody(&bd)
		scrapBox := MakeBox(0.5, 0.5)
		sd := DefaultShapeDef()
		scrapId.CreatePolygonShape(&sd, &scrapBox)
		scrapId.Destroy()
	}
	if outIds != nil {
		outIds.body = stackTop
		outIds.shape = heldShape
		outIds.joint = heldJoint
		outIds.chain = heldChain
	}
	return worldId
}
func snapshotStepUntilSleep(r *oracle.Reader, worldId World) int {
	dt := float32(1.0) / 60.0
	subSteps := 4
	maxSteps := 500
	for step := 0; step < maxSteps; step++ {
		worldId.Step(dt, subSteps)
		snapshotTransformTrace(r, worldId, fmt.Sprintf("L227/step/%d", step), step)
		awake := worldId.AwakeBodyCount()
		if awake == 0 {
			return step + 1
		}
	}
	return maxSteps
}
func TestSnapshot(t *testing.T) {
	defer porttest.Guard(t)
	r := oracle.Open(t, "snapshot/SnapshotTest")
	snapPath := filepath.Join(t.TempDir(), "test_snapshot_midstream.b2rec")
	worldAId := snapshotBuildScene(1, nil)
	snapshotStepUntilSleep(r, worldAId)
	worldA := getWorldFromId(worldAId)
	porttest.Ensure(t, len(worldA.solverSets) > 3)
	r.Bool("L247", len(worldA.solverSets) > 3)
	r.I32("L247/value0", int32(len(worldA.solverSets)))
	porttest.Ensure(t, worldA.broadPhase.trees[StaticBody].pairFreeList != int32(nullIndex))
	r.Bool("L250", worldA.broadPhase.trees[StaticBody].pairFreeList != int32(nullIndex))
	r.I32("L250/value0", int32(worldA.broadPhase.trees[StaticBody].pairFreeList))
	r.I32("L250/value1", int32(nullIndex))
	buf := recBuffer{}
	serializeWorld(worldA, &buf)
	porttest.Ensure(t, buf.size > 0)
	r.Bool("L255", buf.size > 0)
	worldBId, err := CreateWorldFromSnapshot(buf.data, 1)
	porttest.Ensure(t, err == nil && worldBId.IsValid())
	r.Bool("L259", worldBId.IsValid())
	worldB := getWorldFromId(worldBId)
	hashA0 := hashWorldState(worldA)
	r.U64("L264/value", hashA0)
	hashB0 := hashWorldState(worldB)
	r.U64("L265/value", hashB0)
	porttest.Ensure(t, hashA0 == hashB0)
	r.Bool("L266", hashA0 == hashB0)
	deepA0 := hashWorldStateDeep(worldA)
	r.U64("L268/value", deepA0)
	deepB0 := hashWorldStateDeep(worldB)
	r.U64("L269/value", deepB0)
	porttest.Ensure(t, deepA0 == deepB0)
	r.Bool("L270", deepA0 == deepB0)
	for treeType := 0; treeType < int(BodyTypeCount); treeType++ {
		treeA := &worldA.broadPhase.trees[treeType]
		treeB := &worldB.broadPhase.trees[treeType]
		porttest.Ensure(t, treeA.nodeEnd == treeB.nodeEnd)
		r.Bool("L278", treeA.nodeEnd == treeB.nodeEnd)
		r.I32("L278/value0", int32(treeA.nodeEnd))
		r.I32("L278/value1", int32(treeB.nodeEnd))
		porttest.Ensure(t, treeA.pairFreeList == treeB.pairFreeList)
		r.Bool("L279", treeA.pairFreeList == treeB.pairFreeList)
		r.I32("L279/value0", int32(treeA.pairFreeList))
		r.I32("L279/value1", int32(treeB.pairFreeList))
		for pair := treeA.pairFreeList; pair != int32(nullIndex); pair = treeA.parents[pair] {
			porttest.Ensure(t, treeA.parents[pair] == treeB.parents[pair])
			r.Bool("L283", treeA.parents[pair] == treeB.parents[pair])
			r.I32("L283/value0", int32(treeA.parents[pair]))
			r.I32("L283/value1", int32(treeB.parents[pair]))
			porttest.Ensure(t, treeA.parents[pair+1] == treeB.parents[pair+1])
			r.Bool("L284", treeA.parents[pair+1] == treeB.parents[pair+1])
			r.I32("L284/value0", int32(treeA.parents[pair+1]))
			r.I32("L284/value1", int32(treeB.parents[pair+1]))
		}
	}
	dt := float32(1.0) / 60.0
	subSteps := 4
	for step := 0; step < 120; step++ {
		worldAId.Step(dt, subSteps)
		snapshotTransformTrace(r, worldAId, fmt.Sprintf("L293/step/%d", step), step)
		worldBId.Step(dt, subSteps)
		snapshotTransformTrace(r, worldBId, fmt.Sprintf("L294/step/%d", step), step)
		sA := hashWorldState(worldA)
		r.U64("L296/value", sA)
		sB := hashWorldState(worldB)
		r.U64("L297/value", sB)
		if sA != sB {
			r.Bool("L302", false)
			porttest.Ensure(t, false)
		}
		dA := hashWorldStateDeep(worldA)
		r.U64("L305/value", dA)
		dB := hashWorldStateDeep(worldB)
		r.U64("L306/value", dB)
		if dA != dB {
			r.Bool("L311", false)
			porttest.Ensure(t, false)
		}
	}
	recBufFree(&buf)
	buf = recBuffer{}
	serializeWorld(worldA, &buf)
	worldA1Id, err := CreateWorldFromSnapshot(buf.data, 1)
	porttest.Ensure(t, err == nil && worldA1Id.IsValid())
	r.Bool("L322", worldA1Id.IsValid())
	worldA1 := getWorldFromId(worldA1Id)
	worldCId, err := CreateWorldFromSnapshot(buf.data, 4)
	porttest.Ensure(t, err == nil && worldCId.IsValid())
	r.Bool("L326", worldCId.IsValid())
	worldC := getWorldFromId(worldCId)
	porttest.Ensure(t, hashWorldState(worldA1) == hashWorldState(worldC))
	r.Bool("L329", hashWorldState(worldA1) == hashWorldState(worldC))
	for step := 0; step < 120; step++ {
		worldA1Id.Step(dt, subSteps)
		snapshotTransformTrace(r, worldA1Id, fmt.Sprintf("L333/step/%d", step), step)
		worldCId.Step(dt, subSteps)
		snapshotTransformTrace(r, worldCId, fmt.Sprintf("L334/step/%d", step), step)
		sA1 := hashWorldState(worldA1)
		r.U64("L336/value", sA1)
		sC := hashWorldState(worldC)
		r.U64("L337/value", sC)
		if sA1 != sC {
			r.Bool("L342", false)
			porttest.Ensure(t, false)
		}
	}
	var ids snapshotSnapshotIDs
	rId := snapshotBuildScene(1, &ids)
	snapshotStepUntilSleep(r, rId)
	rWorld := getWorldFromId(rId)
	image := rId.Snapshot()
	imageSize := len(image)
	porttest.Ensure(t, imageSize > 0)
	r.Bool("L354", imageSize > 0)
	image = rId.Snapshot()
	written := len(image)
	porttest.Ensure(t, written == imageSize)
	r.Bool("L357", written == imageSize)
	snapHash := hashWorldStateDeep(rWorld)
	r.U64("L359/value", snapHash)
	ids.body.SetLinearVelocity(Vec2{X: 3.0, Y: 6.0})
	ids.body.SetAwake(true)
	postDef := DefaultBodyDef()
	postDef.Type = DynamicBody
	postDef.Position = Pos{X: 20.0, Y: 20.0}
	postBody := rId.CreateBody(&postDef)
	for step := 0; step < 30; step++ {
		rId.Step(dt, subSteps)
		snapshotTransformTrace(r, rId, fmt.Sprintf("L370/step/%d", step), step)
	}
	porttest.Ensure(t, hashWorldStateDeep(rWorld) != snapHash)
	r.Bool("L372", hashWorldStateDeep(rWorld) != snapHash)
	porttest.Ensure(t, rId.Restore(image) == nil)
	r.Bool("L374", true)
	porttest.Ensure(t, hashWorldStateDeep(rWorld) == snapHash)
	r.Bool("L377", hashWorldStateDeep(rWorld) == snapHash)
	porttest.Ensure(t, rId.IsValid())
	r.Bool("L380", rId.IsValid())
	porttest.Ensure(t, ids.body.IsValid())
	r.Bool("L381", ids.body.IsValid())
	porttest.Ensure(t, ids.shape.IsValid())
	r.Bool("L382", ids.shape.IsValid())
	porttest.Ensure(t, ids.joint.IsValid())
	r.Bool("L383", ids.joint.IsValid())
	porttest.Ensure(t, ids.chain.IsValid())
	r.Bool("L384", ids.chain.IsValid())
	porttest.Ensure(t, postBody.IsValid() == false)
	r.Bool("L387", postBody.IsValid() == false)
	preBadHash := hashWorldStateDeep(rWorld)
	r.U64("L390/value", preBadHash)
	porttest.Ensure(t, rId.Restore(nil) != nil)
	r.Bool("L391", true)
	corrupt := make([]byte, imageSize)
	copy(corrupt, image)
	corrupt[0] ^= 0xFF
	porttest.Ensure(t, rId.Restore(corrupt) != nil)
	r.Bool("L395", true)
	porttest.Ensure(t, hashWorldStateDeep(rWorld) == preBadHash)
	r.Bool("L396", hashWorldStateDeep(rWorld) == preBadHash)
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
	porttest.Ensure(t, treeOffset >= 0)
	r.Bool("L416", treeOffset >= 0)
	greedy := make([]byte, imageSize)
	copy(greedy, image)
	hugeCapacity := 2147483647 / 64
	binary.LittleEndian.PutUint32(greedy[treeOffset+4:], uint32(hugeCapacity))
	porttest.Ensure(t, rId.Restore(greedy) == nil)
	r.Bool("L422", true)
	porttest.Ensure(t, hashWorldStateDeep(rWorld) == snapHash)
	r.Bool("L423", hashWorldStateDeep(rWorld) == snapHash)
	porttest.Ensure(t, staticTree.nodeCapacity <= 2*staticTree.nodeEnd)
	r.Bool("L424", staticTree.nodeCapacity <= 2*staticTree.nodeEnd)
	r.I32("L424/value0", int32(staticTree.nodeCapacity))
	r.I32("L424/value1", int32(2*staticTree.nodeEnd))
	for i := 0; i < 3; i++ {
		porttest.Ensure(t, rId.Restore(image) == nil)
		r.Bool("L430", true)
	}
	porttest.Ensure(t, hashWorldStateDeep(rWorld) == snapHash)
	r.Bool("L432", hashWorldStateDeep(rWorld) == snapHash)
	def4 := DefaultWorldDef()
	def4.WorkerCount = 4
	sId := CreateWorld(&def4)
	porttest.Ensure(t, sId.Restore(image) == nil)
	r.Bool("L439", true)
	sWorld := getWorldFromId(sId)
	freshId, err := CreateWorldFromSnapshot(image, 1)
	porttest.Ensure(t, err == nil && freshId.IsValid())
	r.Bool("L443", freshId.IsValid())
	freshWorld := getWorldFromId(freshId)
	porttest.Ensure(t, hashWorldState(sWorld) == hashWorldState(freshWorld))
	r.Bool("L446", hashWorldState(sWorld) == hashWorldState(freshWorld))
	for step := 0; step < 120; step++ {
		sId.Step(dt, subSteps)
		snapshotTransformTrace(r, sId, fmt.Sprintf("L449/step/%d", step), step)
		freshId.Step(dt, subSteps)
		snapshotTransformTrace(r, freshId, fmt.Sprintf("L450/step/%d", step), step)
		if hashWorldState(sWorld) != hashWorldState(freshWorld) {
			r.Bool("L454", false)
			porttest.Ensure(t, false)
		}
	}
	{
		wId := snapshotBuildScene(1, nil)
		for step := 0; step < 20; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L469/step/%d", step), step)
		}
		rec := CreateRecording(0)
		wId.StartRecording(rec)
		for step := 0; step < 60; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L476/step/%d", step), step)
		}
		wId.StopRecording()
		wId.Destroy()
		recData := rec.Data()
		recSize := rec.Size()
		porttest.Ensure(t, ValidateReplay(recData[:recSize], 0))
		r.Bool("L483", true)
		porttest.Ensure(t, ValidateReplay(recData[:recSize], 4))
		r.Bool("L484", true)
		porttest.Ensure(t, rec.SaveToFile(snapPath) == nil)
		r.Bool("L487", true)
		loaded, err := LoadRecordingFromFile(snapPath)
		porttest.Ensure(t, err == nil && loaded != nil)
		r.Bool("L489", loaded != nil)
		porttest.Ensure(t, ValidateReplay(loaded.Data(), 0))
		r.Bool("L490", true)
		loaded.Destroy()
		player, err := CreateReplay(recData, 0)
		porttest.Ensure(t, err == nil && player != nil)
		r.Bool("L495", player != nil)
		pid0 := player.WorldID()
		frames := 0
		for player.StepFrame() {
			frames += 1
			snapshotTransformTrace(r, player.WorldID(), fmt.Sprintf("L501/step/%d", frames), frames-1)
		}
		porttest.Ensure(t, frames == 60)
		r.Bool("L503", frames == 60)
		r.I32("L503/value0", int32(frames))
		r.I32("L503/value0", int32(frames))
		porttest.Ensure(t, player.HasDiverged() == false)
		r.Bool("L504", player.HasDiverged() == false)
		player.Restart()
		pid1 := player.WorldID()
		porttest.Ensure(t, pid0.index1 == pid1.index1 && pid0.generation == pid1.generation)
		r.Bool("L508", pid0.index1 == pid1.index1 && pid0.generation == pid1.generation)
		porttest.Ensure(t, player.Frame() == 0)
		r.Bool("L509", player.Frame() == 0)
		r.I32("L509/value0", int32(player.Frame()))
		r.I32("L509/value0", int32(player.Frame()))
		frames2 := 0
		for player.StepFrame() {
			frames2 += 1
			snapshotTransformTrace(r, player.WorldID(), fmt.Sprintf("L514/step/%d", frames2), frames2-1)
		}
		porttest.Ensure(t, frames2 == 60)
		r.Bool("L516", frames2 == 60)
		r.I32("L516/value0", int32(frames2))
		r.I32("L516/value0", int32(frames2))
		porttest.Ensure(t, player.HasDiverged() == false)
		r.Bool("L517", player.HasDiverged() == false)
		player.Destroy()
		rec.Destroy()
	}
	{
		wId := snapshotBuildScene(1, nil)
		for step := 0; step < 20; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L531/step/%d", step), step)
		}
		snap := wId.Snapshot()
		origin := getWorldFromId(wId)
		const tailSteps = 90
		var realTail [tailSteps]uint64
		for step := 0; step < tailSteps; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L547/step/%d", step), step)
			realTail[step] = hashWorldState(origin)
		}
		cId, err := CreateWorldFromSnapshot(snap, 1)
		porttest.Ensure(t, err == nil && cId.IsValid())
		r.Bool("L552", cId.IsValid())
		clone := getWorldFromId(cId)
		for step := 0; step < tailSteps; step++ {
			cId.Step(dt, subSteps)
			snapshotTransformTrace(r, cId, fmt.Sprintf("L556/step/%d", step), step)
			if realTail[step] != hashWorldState(clone) {
				r.Bool("L560", false)
				porttest.Ensure(t, false)
			}
		}
		wId.Destroy()
		cId.Destroy()
	}
	{
		wd := DefaultWorldDef()
		wd.WorkerCount = 1
		wId := CreateWorld(&wd)
		rec := CreateRecording(0)
		wId.StartRecording(rec)
		bd := DefaultBodyDef()
		bd.Type = DynamicBody
		bd.Position = Pos{X: 0.0, Y: 8.0}
		fallingBody := wId.CreateBody(&bd)
		box := MakeBox(0.5, 0.5)
		sdef := DefaultShapeDef()
		fallingBody.CreatePolygonShape(&sdef, &box)
		for step := 0; step < 30; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L589/step/%d", step), step)
		}
		wId.StopRecording()
		wId.Destroy()
		player, err := CreateReplay(rec.Data(), 0)
		porttest.Ensure(t, err == nil && player != nil)
		r.Bool("L595", player != nil)
		pid0 := player.WorldID()
		for i := 0; i < 5; i++ {
			player.StepFrame()
			snapshotTransformTrace(r, player.WorldID(), fmt.Sprintf("L599/step/%d", i), i)
		}
		player.Restart()
		pid1 := player.WorldID()
		porttest.Ensure(t, pid0.index1 == pid1.index1 && pid0.generation == pid1.generation)
		r.Bool("L603", pid0.index1 == pid1.index1 && pid0.generation == pid1.generation)
		porttest.Ensure(t, player.Frame() == 0)
		r.Bool("L604", player.Frame() == 0)
		r.I32("L604/value0", int32(player.Frame()))
		r.I32("L604/value0", int32(player.Frame()))
		player.Destroy()
		rec.Destroy()
	}
	{
		wd := DefaultWorldDef()
		wd.WorkerCount = 1
		wId := CreateWorld(&wd)
		w := getWorldFromId(wId)
		gbd := DefaultBodyDef()
		gbd.Position = Pos{X: 0.0, Y: -1.0}
		ground := wId.CreateBody(&gbd)
		groundBox := MakeBox(10.0, 1.0)
		gsd := DefaultShapeDef()
		gsd.EnableContactEvents = true
		ground.CreatePolygonShape(&gsd, &groundBox)
		bd := DefaultBodyDef()
		bd.Type = DynamicBody
		bd.Position = Pos{X: 0.0, Y: 0.5}
		box := wId.CreateBody(&bd)
		boxPoly := MakeBox(0.5, 0.5)
		bsd := DefaultShapeDef()
		bsd.EnableContactEvents = true
		box.CreatePolygonShape(&bsd, &boxPoly)
		for step := 0; step < 60; step++ {
			wId.Step(dt, subSteps)
			snapshotTransformTrace(r, wId, fmt.Sprintf("L637/step/%d", step), step)
		}
		porttest.Ensure(t, wId.StateHash() == hashWorldStateDeep(w))
		r.Bool("L641", wId.StateHash() == hashWorldStateDeep(w))
		snap := wId.Snapshot()
		box.Disable()
		porttest.Ensure(t, len(w.contactEndEvents[w.endEventArrayIndex]) >= 1)
		r.Bool("L650", len(w.contactEndEvents[w.endEventArrayIndex]) >= 1)
		r.I32("L650/value0", int32(len(w.contactEndEvents[w.endEventArrayIndex])))
		porttest.Ensure(t, wId.Restore(snap) == nil)
		r.Bool("L652", true)
		porttest.Ensure(t, len(w.contactEndEvents[0]) == 0)
		r.Bool("L655", len(w.contactEndEvents[0]) == 0)
		r.I32("L655/value0", int32(len(w.contactEndEvents[0])))
		porttest.Ensure(t, len(w.contactEndEvents[1]) == 0)
		r.Bool("L656", len(w.contactEndEvents[1]) == 0)
		r.I32("L656/value0", int32(len(w.contactEndEvents[1])))
		wId.Step(dt, subSteps)
		snapshotTransformTrace(r, wId, "L659/step/0", 0)
		endCount := 0
		for range wId.ContactEndEvents() {
			endCount++
		}
		porttest.Ensure(t, endCount == 0)
		r.Bool("L661", endCount == 0)
		r.I32("L661/value0", int32(endCount))
		r.I32("L661/value0", int32(endCount))
		wId.Destroy()
	}
	_ = os.Remove(snapPath)
	worldAId.Destroy()
	worldBId.Destroy()
	worldA1Id.Destroy()
	worldCId.Destroy()
	rId.Destroy()
	sId.Destroy()
	freshId.Destroy()
	recBufFree(&buf)
	return
}
