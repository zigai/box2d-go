// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

type contactConstraintPoint struct {
	anchorA            Vec2
	anchorB            Vec2
	baseSeparation     float32
	relativeVelocity   float32
	normalImpulse      float32
	tangentImpulse     float32
	totalNormalImpulse float32
	restitutionImpulse float32
	normalMass         float32
	tangentMass        float32
}

type contactConstraint struct {
	indexA            int
	indexB            int
	points            [2]contactConstraintPoint
	normal            Vec2
	invMassA          float32
	invMassB          float32
	invIA             float32
	invIB             float32
	friction          float32
	restitution       float32
	tangentSpeed      float32
	rollingResistance float32
	rollingMass       float32
	rollingImpulse    float32
	softness          softness
	pointCount        int
}

type floatW struct {
	x float32
	y float32
	z float32
	w float32
}

type vec2W struct {
	x floatW
	y floatW
}

type rotW struct {
	c floatW
	s floatW
}

type contactConstraintWide struct {
	indexA              [simdWidth]int
	indexB              [simdWidth]int
	invMassA            floatW
	invMassB            floatW
	invIA               floatW
	invIB               floatW
	normal              vec2W
	anchorA1            vec2W
	anchorB1            vec2W
	anchorA2            vec2W
	anchorB2            vec2W
	normalMass1         floatW
	normalMass2         floatW
	baseSeparation1     floatW
	baseSeparation2     floatW
	normalImpulse1      floatW
	normalImpulse2      floatW
	totalNormalImpulse1 floatW
	totalNormalImpulse2 floatW
	tangentImpulse1     floatW
	tangentImpulse2     floatW
	rollingImpulse      floatW
	friction            floatW
	tangentSpeed        floatW
	rollingResistance   floatW
	tangentMass1        floatW
	tangentMass2        floatW
	restitution         floatW
	relativeVelocity1   floatW
	relativeVelocity2   floatW
	restitutionImpulse1 floatW
	restitutionImpulse2 floatW
}

type bodyStateW struct {
	v     vec2W
	w     floatW
	flags floatW
	dp    vec2W
	dq    rotW
}

var zeroContactSim contactSim

func prepareContactsOverflow(context *stepContext) {
	todo()
}

func warmStartContactsOverflow(context *stepContext) {
	todo()
}

func pushContactsOverflow(context *stepContext) {
	todo()
}

func solveContactsOverflow(context *stepContext) {
	todo()
}

func applyRestitutionOverflow(context *stepContext) {
	todo()
}

func storeImpulsesOverflow(context *stepContext) {
	todo()
}

func zeroW() floatW {
	todo()
	return floatW{}
}

func splatW(scalar float32) floatW {
	todo()
	return floatW{}
}

func setW(a float32, b float32, c float32, d float32) floatW {
	todo()
	return floatW{}
}

func negW(a floatW) floatW {
	todo()
	return floatW{}
}

func addW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func subW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func mulW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func divW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func mulAddW(a floatW, b floatW, c floatW) floatW {
	todo()
	return floatW{}
}

func mulSubW(a floatW, b floatW, c floatW) floatW {
	todo()
	return floatW{}
}

func minW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func maxW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func symClampW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func andW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func greaterThanW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func lessThanW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func anyTrueW(mask floatW) bool {
	todo()
	return false
}

func allZeroW(a floatW) bool {
	todo()
	return false
}

func blendW(a floatW, b floatW, mask floatW) floatW {
	todo()
	return floatW{}
}

func loadU(data []float32) floatW {
	todo()
	return floatW{}
}

func storeW(data []float32, a floatW) {
	todo()
}

func softMaskW(indexA []int, indexB []int) floatW {
	todo()
	return floatW{}
}

func unpackLoW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func unpackHiW(a floatW, b floatW) floatW {
	todo()
	return floatW{}
}

func subVW(a vec2W, b vec2W) vec2W {
	todo()
	return vec2W{}
}

func rightPerpW(a vec2W) vec2W {
	todo()
	return vec2W{}
}

func dotW(a vec2W, b vec2W) floatW {
	todo()
	return floatW{}
}

func crossW(a vec2W, b vec2W) floatW {
	todo()
	return floatW{}
}

func rotateVectorW(q rotW, v vec2W) vec2W {
	todo()
	return vec2W{}
}

func getWideContactConstraintByteCount() int {
	todo()
	return 0
}

func gatherBodies(states []bodyState, indices []int) bodyStateW {
	todo()
	return bodyStateW{}
}

func scatterBodies(states []bodyState, indices []int, simdBody *bodyStateW) {
	todo()
}

func prepareContactsWide(block solverBlock, context *stepContext) {
	todo()
}

func warmStartContactsWide(block solverBlock, context *stepContext) {
	todo()
}

func pushContactsWide(block solverBlock, context *stepContext) {
	todo()
}

func solveContactsWide(block solverBlock, context *stepContext) {
	todo()
}

func applyRestitutionWide(block solverBlock, context *stepContext) {
	todo()
}

func storeImpulsesWide(block solverBlock, context *stepContext, workerIndex int) {
	todo()
}
