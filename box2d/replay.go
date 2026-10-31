// SPDX-FileCopyrightText: 2026 Erin Catto
// SPDX-License-Identifier: MIT

package box2d

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"
)

const (
	recKeyframeIntervalDefault int  = 16
	recKeyframeBudgetDefault   uint = 512 * 1024 * 1024
)

const (
	ReplayQueryOverlapAABB ReplayQueryType = iota
	ReplayQueryOverlapShape
	ReplayQueryCastRay
	ReplayQueryCastShape
	ReplayQueryCollideMover
	ReplayQueryCastRayClosest
	ReplayQueryCastMover
	ReplayQueryShapeTestPoint
	ReplayQueryShapeRayCast
)

const (
	recqOverlapAABB recQueryKind = iota
	recqOverlapShape
	recqCastRay
	recqCastShape
	recqCollideMover
	recqCastRayClosest
	recqCastMover
	recqShapeTestPoint
	recqShapeRayCast
)

var recReplayDecoders = [256]func(*recReader){
	recOpDestroyWorld: func(rdr *recReader) {
		var a recArgsDestroyWorld
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDestroyWorld(&a, rdr)
		}
	},
	recOpStep: func(rdr *recReader) {
		var a recArgsStep

		a.dt = recRF32(rdr)

		a.subStepCount = recRI32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchStep(&a, rdr)
		}
	},
	recOpWorldEnableSleeping: func(rdr *recReader) {
		var a recArgsWorldEnableSleeping

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldEnableSleeping(&a, rdr)
		}
	},
	recOpWorldEnableContinuous: func(rdr *recReader) {
		var a recArgsWorldEnableContinuous

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldEnableContinuous(&a, rdr)
		}
	},
	recOpWorldSetRestitutionThreshold: func(rdr *recReader) {
		var a recArgsWorldSetRestitutionThreshold

		a.value = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetRestitutionThreshold(&a, rdr)
		}
	},
	recOpWorldSetHitEventThreshold: func(rdr *recReader) {
		var a recArgsWorldSetHitEventThreshold

		a.value = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetHitEventThreshold(&a, rdr)
		}
	},
	recOpWorldSetGravity: func(rdr *recReader) {
		var a recArgsWorldSetGravity

		a.gravity = recRVEC2(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetGravity(&a, rdr)
		}
	},
	recOpWorldExplode: func(rdr *recReader) {
		var a recArgsWorldExplode

		a.def = recREXPLOSIONDEF(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldExplode(&a, rdr)
		}
	},
	recOpWorldSetContactTuning: func(rdr *recReader) {
		var a recArgsWorldSetContactTuning

		a.hertz = recRF32(rdr)
		a.dampingRatio = recRF32(rdr)

		a.pushSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetContactTuning(&a, rdr)
		}
	},
	recOpWorldSetContactRecycleDistance: func(rdr *recReader) {
		var a recArgsWorldSetContactRecycleDistance

		a.recycleDistance = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetContactRecycleDistance(&a, rdr)
		}
	},
	recOpWorldSetMaximumLinearSpeed: func(rdr *recReader) {
		var a recArgsWorldSetMaximumLinearSpeed

		a.maximumLinearSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetMaximumLinearSpeed(&a, rdr)
		}
	},
	recOpWorldEnableWarmStarting: func(rdr *recReader) {
		var a recArgsWorldEnableWarmStarting

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldEnableWarmStarting(&a, rdr)
		}
	},
	recOpWorldRebuildStaticTree: func(rdr *recReader) {
		var a recArgsWorldRebuildStaticTree
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldRebuildStaticTree(&a, rdr)
		}
	},
	recOpWorldSetRestitutionIterations: func(rdr *recReader) {
		var a recArgsWorldSetRestitutionIterations

		a.iterations = recRI32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldSetRestitutionIterations(&a, rdr)
		}
	},
	recOpWorldEnableRestitutionPropagation: func(rdr *recReader) {
		var a recArgsWorldEnableRestitutionPropagation

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWorldEnableRestitutionPropagation(&a, rdr)
		}
	},
	recOpCreateBody: func(rdr *recReader) {
		var a recArgsCreateBody

		a.def = recRBODYDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateBody(&a, rdr)
		}
	},
	recOpDestroyBody: func(rdr *recReader) {
		var a recArgsDestroyBody

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDestroyBody(&a, rdr)
		}
	},
	recOpBodySetTransform: func(rdr *recReader) {
		var a recArgsBodySetTransform

		a.body = recRBODYID(rdr)
		a.position = recRPOSITION(rdr)

		a.rotation = recRROT(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetTransform(&a, rdr)
		}
	},
	recOpBodySetLinearVelocity: func(rdr *recReader) {
		var a recArgsBodySetLinearVelocity

		a.body = recRBODYID(rdr)

		a.v = recRVEC2(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetLinearVelocity(&a, rdr)
		}
	},
	recOpBodySetType: func(rdr *recReader) {
		var a recArgsBodySetType

		a.body = recRBODYID(rdr)

		a.bodyType = recRI32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetType(&a, rdr)
		}
	},
	recOpBodySetName: func(rdr *recReader) {
		var a recArgsBodySetName

		a.body = recRBODYID(rdr)

		a.name = recRSTR(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetName(&a, rdr)
		}
	},
	recOpBodySetAngularVelocity: func(rdr *recReader) {
		var a recArgsBodySetAngularVelocity

		a.body = recRBODYID(rdr)

		a.w = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetAngularVelocity(&a, rdr)
		}
	},
	recOpBodySetTargetTransform: func(rdr *recReader) {
		var a recArgsBodySetTargetTransform

		a.body = recRBODYID(rdr)
		a.target = recRWORLDXF(rdr)
		a.timeStep = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetTargetTransform(&a, rdr)
		}
	},
	recOpBodyApplyForce: func(rdr *recReader) {
		var a recArgsBodyApplyForce

		a.body = recRBODYID(rdr)
		a.force = recRVEC2(rdr)
		a.point = recRPOSITION(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyForce(&a, rdr)
		}
	},
	recOpBodyApplyForceToCenter: func(rdr *recReader) {
		var a recArgsBodyApplyForceToCenter

		a.body = recRBODYID(rdr)
		a.force = recRVEC2(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyForceToCenter(&a, rdr)
		}
	},
	recOpBodyApplyTorque: func(rdr *recReader) {
		var a recArgsBodyApplyTorque

		a.body = recRBODYID(rdr)
		a.torque = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyTorque(&a, rdr)
		}
	},
	recOpBodyClearForces: func(rdr *recReader) {
		var a recArgsBodyClearForces

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyClearForces(&a, rdr)
		}
	},
	recOpBodyApplyLinearImpulse: func(rdr *recReader) {
		var a recArgsBodyApplyLinearImpulse

		a.body = recRBODYID(rdr)
		a.impulse = recRVEC2(rdr)
		a.point = recRPOSITION(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyLinearImpulse(&a, rdr)
		}
	},
	recOpBodyApplyLinearImpulseToCenter: func(rdr *recReader) {
		var a recArgsBodyApplyLinearImpulseToCenter

		a.body = recRBODYID(rdr)
		a.impulse = recRVEC2(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyLinearImpulseToCenter(&a, rdr)
		}
	},
	recOpBodyApplyAngularImpulse: func(rdr *recReader) {
		var a recArgsBodyApplyAngularImpulse

		a.body = recRBODYID(rdr)
		a.impulse = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyApplyAngularImpulse(&a, rdr)
		}
	},
	recOpBodySetMassData: func(rdr *recReader) {
		var a recArgsBodySetMassData

		a.body = recRBODYID(rdr)

		a.massData = recRMASSDATA(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetMassData(&a, rdr)
		}
	},
	recOpBodyUpdateMassFromShapes: func(rdr *recReader) {
		var a recArgsBodyUpdateMassFromShapes

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyUpdateMassFromShapes(&a, rdr)
		}
	},
	recOpBodySetLinearDamping: func(rdr *recReader) {
		var a recArgsBodySetLinearDamping

		a.body = recRBODYID(rdr)

		a.damping = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetLinearDamping(&a, rdr)
		}
	},
	recOpBodySetAngularDamping: func(rdr *recReader) {
		var a recArgsBodySetAngularDamping

		a.body = recRBODYID(rdr)

		a.damping = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetAngularDamping(&a, rdr)
		}
	},
	recOpBodySetGravityScale: func(rdr *recReader) {
		var a recArgsBodySetGravityScale

		a.body = recRBODYID(rdr)

		a.scale = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetGravityScale(&a, rdr)
		}
	},
	recOpBodySetAwake: func(rdr *recReader) {
		var a recArgsBodySetAwake

		a.body = recRBODYID(rdr)

		a.awake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetAwake(&a, rdr)
		}
	},
	recOpBodyWakeTouching: func(rdr *recReader) {
		var a recArgsBodyWakeTouching

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyWakeTouching(&a, rdr)
		}
	},
	recOpBodyEnableSleep: func(rdr *recReader) {
		var a recArgsBodyEnableSleep

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyEnableSleep(&a, rdr)
		}
	},
	recOpBodySetSleepThreshold: func(rdr *recReader) {
		var a recArgsBodySetSleepThreshold

		a.body = recRBODYID(rdr)

		a.threshold = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetSleepThreshold(&a, rdr)
		}
	},
	recOpBodyDisable: func(rdr *recReader) {
		var a recArgsBodyDisable

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyDisable(&a, rdr)
		}
	},
	recOpBodyEnable: func(rdr *recReader) {
		var a recArgsBodyEnable

		a.body = recRBODYID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyEnable(&a, rdr)
		}
	},
	recOpBodySetMotionLocks: func(rdr *recReader) {
		var a recArgsBodySetMotionLocks

		a.body = recRBODYID(rdr)

		a.locks = recRLOCKS(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetMotionLocks(&a, rdr)
		}
	},
	recOpBodySetBullet: func(rdr *recReader) {
		var a recArgsBodySetBullet

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetBullet(&a, rdr)
		}
	},
	recOpBodyEnableContactRecycling: func(rdr *recReader) {
		var a recArgsBodyEnableContactRecycling

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyEnableContactRecycling(&a, rdr)
		}
	},
	recOpBodyEnableContactEvents: func(rdr *recReader) {
		var a recArgsBodyEnableContactEvents

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyEnableContactEvents(&a, rdr)
		}
	},
	recOpBodyEnableHitEvents: func(rdr *recReader) {
		var a recArgsBodyEnableHitEvents

		a.body = recRBODYID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodyEnableHitEvents(&a, rdr)
		}
	},
	recOpBodySetSafetyFactor: func(rdr *recReader) {
		var a recArgsBodySetSafetyFactor

		a.body = recRBODYID(rdr)

		a.value = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchBodySetSafetyFactor(&a, rdr)
		}
	},
	recOpCreateCircleShape: func(rdr *recReader) {
		var a recArgsCreateCircleShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.circle = recRCIRCLE(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateCircleShape(&a, rdr)
		}
	},
	recOpCreateCapsuleShape: func(rdr *recReader) {
		var a recArgsCreateCapsuleShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.capsule = recRCAPSULE(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateCapsuleShape(&a, rdr)
		}
	},
	recOpCreateSegmentShape: func(rdr *recReader) {
		var a recArgsCreateSegmentShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.segment = recRSEGMENT(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateSegmentShape(&a, rdr)
		}
	},
	recOpCreatePolygonShape: func(rdr *recReader) {
		var a recArgsCreatePolygonShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.polygon = recRPOLYGON(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreatePolygonShape(&a, rdr)
		}
	},
	recOpCreateChainSegmentShape: func(rdr *recReader) {
		var a recArgsCreateChainSegmentShape

		a.body = recRBODYID(rdr)
		a.def = recRSHAPEDEF(rdr)

		a.chainSegment = recRCHAINSEG(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateChainSegmentShape(&a, rdr)
		}
	},
	recOpDestroyShape: func(rdr *recReader) {
		var a recArgsDestroyShape

		a.shape = recRSHAPEID(rdr)

		a.updateBodyMass = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDestroyShape(&a, rdr)
		}
	},
	recOpShapeSetDensity: func(rdr *recReader) {
		var a recArgsShapeSetDensity

		a.shape = recRSHAPEID(rdr)
		a.density = recRF32(rdr)

		a.updateBodyMass = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetDensity(&a, rdr)
		}
	},
	recOpShapeSetFriction: func(rdr *recReader) {
		var a recArgsShapeSetFriction

		a.shape = recRSHAPEID(rdr)

		a.friction = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetFriction(&a, rdr)
		}
	},
	recOpShapeSetRestitution: func(rdr *recReader) {
		var a recArgsShapeSetRestitution

		a.shape = recRSHAPEID(rdr)

		a.restitution = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetRestitution(&a, rdr)
		}
	},
	recOpShapeSetUserMaterial: func(rdr *recReader) {
		var a recArgsShapeSetUserMaterial

		a.shape = recRSHAPEID(rdr)

		a.material = recRU64(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetUserMaterial(&a, rdr)
		}
	},
	recOpShapeSetSurfaceMaterial: func(rdr *recReader) {
		var a recArgsShapeSetSurfaceMaterial

		a.shape = recRSHAPEID(rdr)

		a.material = recRMATERIAL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetSurfaceMaterial(&a, rdr)
		}
	},
	recOpShapeSetFilter: func(rdr *recReader) {
		var a recArgsShapeSetFilter

		a.shape = recRSHAPEID(rdr)

		a.filter = recRFILTER(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetFilter(&a, rdr)
		}
	},
	recOpShapeEnableSensorEvents: func(rdr *recReader) {
		var a recArgsShapeEnableSensorEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeEnableSensorEvents(&a, rdr)
		}
	},
	recOpShapeEnableContactEvents: func(rdr *recReader) {
		var a recArgsShapeEnableContactEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeEnableContactEvents(&a, rdr)
		}
	},
	recOpShapeEnablePreSolveEvents: func(rdr *recReader) {
		var a recArgsShapeEnablePreSolveEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeEnablePreSolveEvents(&a, rdr)
		}
	},
	recOpShapeEnableHitEvents: func(rdr *recReader) {
		var a recArgsShapeEnableHitEvents

		a.shape = recRSHAPEID(rdr)

		a.flag = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeEnableHitEvents(&a, rdr)
		}
	},
	recOpShapeSetCircle: func(rdr *recReader) {
		var a recArgsShapeSetCircle

		a.shape = recRSHAPEID(rdr)

		a.circle = recRCIRCLE(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetCircle(&a, rdr)
		}
	},
	recOpShapeSetCapsule: func(rdr *recReader) {
		var a recArgsShapeSetCapsule

		a.shape = recRSHAPEID(rdr)

		a.capsule = recRCAPSULE(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetCapsule(&a, rdr)
		}
	},
	recOpShapeSetSegment: func(rdr *recReader) {
		var a recArgsShapeSetSegment

		a.shape = recRSHAPEID(rdr)

		a.segment = recRSEGMENT(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetSegment(&a, rdr)
		}
	},
	recOpShapeSetPolygon: func(rdr *recReader) {
		var a recArgsShapeSetPolygon

		a.shape = recRSHAPEID(rdr)

		a.polygon = recRPOLYGON(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetPolygon(&a, rdr)
		}
	},
	recOpShapeSetChainSegment: func(rdr *recReader) {
		var a recArgsShapeSetChainSegment

		a.shape = recRSHAPEID(rdr)

		a.chainSegment = recRCHAINSEG(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeSetChainSegment(&a, rdr)
		}
	},
	recOpShapeApplyWind: func(rdr *recReader) {
		var a recArgsShapeApplyWind

		a.shape = recRSHAPEID(rdr)
		a.wind = recRVEC2(rdr)
		a.drag = recRF32(rdr)
		a.lift = recRF32(rdr)

		a.wake = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchShapeApplyWind(&a, rdr)
		}
	},
	recOpCreateChain: func(rdr *recReader) {
		var a recArgsCreateChain

		a.body = recRBODYID(rdr)

		a.def = recRCHAINDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateChain(&a, rdr)
		}
	},
	recOpDestroyChain: func(rdr *recReader) {
		var a recArgsDestroyChain

		a.chain = recRCHAINID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDestroyChain(&a, rdr)
		}
	},
	recOpChainSetSurfaceMaterial: func(rdr *recReader) {
		var a recArgsChainSetSurfaceMaterial

		a.chain = recRCHAINID(rdr)
		a.material = recRMATERIAL(rdr)

		a.segmentIndex = recRI32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchChainSetSurfaceMaterial(&a, rdr)
		}
	},
	recOpChainSetAllSurfaceMaterials: func(rdr *recReader) {
		var a recArgsChainSetAllSurfaceMaterials

		a.chain = recRCHAINID(rdr)

		a.material = recRMATERIAL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchChainSetAllSurfaceMaterials(&a, rdr)
		}
	},
	recOpDestroyJoint: func(rdr *recReader) {
		var a recArgsDestroyJoint

		a.joint = recRJOINTID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDestroyJoint(&a, rdr)
		}
	},
	recOpJointSetLocalFrameA: func(rdr *recReader) {
		var a recArgsJointSetLocalFrameA

		a.joint = recRJOINTID(rdr)

		a.localFrame = recRXF(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetLocalFrameA(&a, rdr)
		}
	},
	recOpJointSetLocalFrameB: func(rdr *recReader) {
		var a recArgsJointSetLocalFrameB

		a.joint = recRJOINTID(rdr)

		a.localFrame = recRXF(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetLocalFrameB(&a, rdr)
		}
	},
	recOpJointSetCollideConnected: func(rdr *recReader) {
		var a recArgsJointSetCollideConnected

		a.joint = recRJOINTID(rdr)

		a.shouldCollide = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetCollideConnected(&a, rdr)
		}
	},
	recOpJointWakeBodies: func(rdr *recReader) {
		var a recArgsJointWakeBodies

		a.joint = recRJOINTID(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointWakeBodies(&a, rdr)
		}
	},
	recOpJointSetConstraintTuning: func(rdr *recReader) {
		var a recArgsJointSetConstraintTuning

		a.joint = recRJOINTID(rdr)
		a.hertz = recRF32(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetConstraintTuning(&a, rdr)
		}
	},
	recOpJointSetForceThreshold: func(rdr *recReader) {
		var a recArgsJointSetForceThreshold

		a.joint = recRJOINTID(rdr)

		a.threshold = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetForceThreshold(&a, rdr)
		}
	},
	recOpJointSetTorqueThreshold: func(rdr *recReader) {
		var a recArgsJointSetTorqueThreshold

		a.joint = recRJOINTID(rdr)

		a.threshold = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchJointSetTorqueThreshold(&a, rdr)
		}
	},
	recOpCreateDistanceJoint: func(rdr *recReader) {
		var a recArgsCreateDistanceJoint

		a.def = recRDISTANCEJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateDistanceJoint(&a, rdr)
		}
	},
	recOpDistanceJointSetLength: func(rdr *recReader) {
		var a recArgsDistanceJointSetLength

		a.joint = recRJOINTID(rdr)

		a.length = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetLength(&a, rdr)
		}
	},
	recOpDistanceJointEnableSpring: func(rdr *recReader) {
		var a recArgsDistanceJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointEnableSpring(&a, rdr)
		}
	},
	recOpDistanceJointSetSpringForceRange: func(rdr *recReader) {
		var a recArgsDistanceJointSetSpringForceRange

		a.joint = recRJOINTID(rdr)
		a.lowerForce = recRF32(rdr)

		a.upperForce = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetSpringForceRange(&a, rdr)
		}
	},
	recOpDistanceJointSetSpringHertz: func(rdr *recReader) {
		var a recArgsDistanceJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetSpringHertz(&a, rdr)
		}
	},
	recOpDistanceJointSetSpringDampingRatio: func(rdr *recReader) {
		var a recArgsDistanceJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetSpringDampingRatio(&a, rdr)
		}
	},
	recOpDistanceJointEnableLimit: func(rdr *recReader) {
		var a recArgsDistanceJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointEnableLimit(&a, rdr)
		}
	},
	recOpDistanceJointSetLengthRange: func(rdr *recReader) {
		var a recArgsDistanceJointSetLengthRange

		a.joint = recRJOINTID(rdr)
		a.minLength = recRF32(rdr)

		a.maxLength = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetLengthRange(&a, rdr)
		}
	},
	recOpDistanceJointEnableMotor: func(rdr *recReader) {
		var a recArgsDistanceJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointEnableMotor(&a, rdr)
		}
	},
	recOpDistanceJointSetMotorSpeed: func(rdr *recReader) {
		var a recArgsDistanceJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetMotorSpeed(&a, rdr)
		}
	},
	recOpDistanceJointSetMaxMotorForce: func(rdr *recReader) {
		var a recArgsDistanceJointSetMaxMotorForce

		a.joint = recRJOINTID(rdr)

		a.force = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchDistanceJointSetMaxMotorForce(&a, rdr)
		}
	},
	recOpCreateFilterJoint: func(rdr *recReader) {
		var a recArgsCreateFilterJoint

		a.def = recRFILTERJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateFilterJoint(&a, rdr)
		}
	},
	recOpCreateMotorJoint: func(rdr *recReader) {
		var a recArgsCreateMotorJoint

		a.def = recRMOTORJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateMotorJoint(&a, rdr)
		}
	},
	recOpMotorJointSetLinearVelocity: func(rdr *recReader) {
		var a recArgsMotorJointSetLinearVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRVEC2(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetLinearVelocity(&a, rdr)
		}
	},
	recOpMotorJointSetAngularVelocity: func(rdr *recReader) {
		var a recArgsMotorJointSetAngularVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetAngularVelocity(&a, rdr)
		}
	},
	recOpMotorJointSetMaxVelocityForce: func(rdr *recReader) {
		var a recArgsMotorJointSetMaxVelocityForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetMaxVelocityForce(&a, rdr)
		}
	},
	recOpMotorJointSetMaxVelocityTorque: func(rdr *recReader) {
		var a recArgsMotorJointSetMaxVelocityTorque

		a.joint = recRJOINTID(rdr)

		a.maxTorque = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetMaxVelocityTorque(&a, rdr)
		}
	},
	recOpMotorJointSetLinearHertz: func(rdr *recReader) {
		var a recArgsMotorJointSetLinearHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetLinearHertz(&a, rdr)
		}
	},
	recOpMotorJointSetLinearDampingRatio: func(rdr *recReader) {
		var a recArgsMotorJointSetLinearDampingRatio

		a.joint = recRJOINTID(rdr)

		a.damping = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetLinearDampingRatio(&a, rdr)
		}
	},
	recOpMotorJointSetAngularHertz: func(rdr *recReader) {
		var a recArgsMotorJointSetAngularHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetAngularHertz(&a, rdr)
		}
	},
	recOpMotorJointSetAngularDampingRatio: func(rdr *recReader) {
		var a recArgsMotorJointSetAngularDampingRatio

		a.joint = recRJOINTID(rdr)

		a.damping = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetAngularDampingRatio(&a, rdr)
		}
	},
	recOpMotorJointSetMaxSpringForce: func(rdr *recReader) {
		var a recArgsMotorJointSetMaxSpringForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetMaxSpringForce(&a, rdr)
		}
	},
	recOpMotorJointSetMaxSpringTorque: func(rdr *recReader) {
		var a recArgsMotorJointSetMaxSpringTorque

		a.joint = recRJOINTID(rdr)

		a.maxTorque = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMotorJointSetMaxSpringTorque(&a, rdr)
		}
	},
	recOpCreateMoverJoint: func(rdr *recReader) {
		var a recArgsCreateMoverJoint

		a.def = recRMOVERJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateMoverJoint(&a, rdr)
		}
	},
	recOpMoverJointSetLinearVelocity: func(rdr *recReader) {
		var a recArgsMoverJointSetLinearVelocity

		a.joint = recRJOINTID(rdr)

		a.velocity = recRVEC2(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMoverJointSetLinearVelocity(&a, rdr)
		}
	},
	recOpMoverJointSetMaxVelocityForce: func(rdr *recReader) {
		var a recArgsMoverJointSetMaxVelocityForce

		a.joint = recRJOINTID(rdr)

		a.maxForce = recRVEC2(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchMoverJointSetMaxVelocityForce(&a, rdr)
		}
	},
	recOpCreatePogoJoint: func(rdr *recReader) {
		var a recArgsCreatePogoJoint

		a.def = recRPOGOJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreatePogoJoint(&a, rdr)
		}
	},
	recOpPogoJointSetRestLength: func(rdr *recReader) {
		var a recArgsPogoJointSetRestLength

		a.joint = recRJOINTID(rdr)

		a.length = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPogoJointSetRestLength(&a, rdr)
		}
	},
	recOpPogoJointSetSpringHertz: func(rdr *recReader) {
		var a recArgsPogoJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPogoJointSetSpringHertz(&a, rdr)
		}
	},
	recOpPogoJointSetSpringDampingRatio: func(rdr *recReader) {
		var a recArgsPogoJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPogoJointSetSpringDampingRatio(&a, rdr)
		}
	},
	recOpCreatePrismaticJoint: func(rdr *recReader) {
		var a recArgsCreatePrismaticJoint

		a.def = recRPRISMATICJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreatePrismaticJoint(&a, rdr)
		}
	},
	recOpPrismaticJointEnableSpring: func(rdr *recReader) {
		var a recArgsPrismaticJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointEnableSpring(&a, rdr)
		}
	},
	recOpPrismaticJointSetSpringHertz: func(rdr *recReader) {
		var a recArgsPrismaticJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetSpringHertz(&a, rdr)
		}
	},
	recOpPrismaticJointSetSpringDampingRatio: func(rdr *recReader) {
		var a recArgsPrismaticJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetSpringDampingRatio(&a, rdr)
		}
	},
	recOpPrismaticJointSetTargetTranslation: func(rdr *recReader) {
		var a recArgsPrismaticJointSetTargetTranslation

		a.joint = recRJOINTID(rdr)

		a.translation = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetTargetTranslation(&a, rdr)
		}
	},
	recOpPrismaticJointEnableLimit: func(rdr *recReader) {
		var a recArgsPrismaticJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointEnableLimit(&a, rdr)
		}
	},
	recOpPrismaticJointSetLimits: func(rdr *recReader) {
		var a recArgsPrismaticJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetLimits(&a, rdr)
		}
	},
	recOpPrismaticJointEnableMotor: func(rdr *recReader) {
		var a recArgsPrismaticJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointEnableMotor(&a, rdr)
		}
	},
	recOpPrismaticJointSetMotorSpeed: func(rdr *recReader) {
		var a recArgsPrismaticJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetMotorSpeed(&a, rdr)
		}
	},
	recOpPrismaticJointSetMaxMotorForce: func(rdr *recReader) {
		var a recArgsPrismaticJointSetMaxMotorForce

		a.joint = recRJOINTID(rdr)

		a.force = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchPrismaticJointSetMaxMotorForce(&a, rdr)
		}
	},
	recOpCreateRevoluteJoint: func(rdr *recReader) {
		var a recArgsCreateRevoluteJoint

		a.def = recRREVOLUTEJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateRevoluteJoint(&a, rdr)
		}
	},
	recOpRevoluteJointEnableSpring: func(rdr *recReader) {
		var a recArgsRevoluteJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointEnableSpring(&a, rdr)
		}
	},
	recOpRevoluteJointSetSpringHertz: func(rdr *recReader) {
		var a recArgsRevoluteJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetSpringHertz(&a, rdr)
		}
	},
	recOpRevoluteJointSetSpringDampingRatio: func(rdr *recReader) {
		var a recArgsRevoluteJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetSpringDampingRatio(&a, rdr)
		}
	},
	recOpRevoluteJointSetTargetAngle: func(rdr *recReader) {
		var a recArgsRevoluteJointSetTargetAngle

		a.joint = recRJOINTID(rdr)

		a.angle = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetTargetAngle(&a, rdr)
		}
	},
	recOpRevoluteJointEnableLimit: func(rdr *recReader) {
		var a recArgsRevoluteJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointEnableLimit(&a, rdr)
		}
	},
	recOpRevoluteJointSetLimits: func(rdr *recReader) {
		var a recArgsRevoluteJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetLimits(&a, rdr)
		}
	},
	recOpRevoluteJointEnableMotor: func(rdr *recReader) {
		var a recArgsRevoluteJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointEnableMotor(&a, rdr)
		}
	},
	recOpRevoluteJointSetMotorSpeed: func(rdr *recReader) {
		var a recArgsRevoluteJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetMotorSpeed(&a, rdr)
		}
	},
	recOpRevoluteJointSetMaxMotorTorque: func(rdr *recReader) {
		var a recArgsRevoluteJointSetMaxMotorTorque

		a.joint = recRJOINTID(rdr)

		a.torque = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRevoluteJointSetMaxMotorTorque(&a, rdr)
		}
	},
	recOpCreateWeldJoint: func(rdr *recReader) {
		var a recArgsCreateWeldJoint

		a.def = recRWELDJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateWeldJoint(&a, rdr)
		}
	},
	recOpWeldJointSetLinearHertz: func(rdr *recReader) {
		var a recArgsWeldJointSetLinearHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWeldJointSetLinearHertz(&a, rdr)
		}
	},
	recOpWeldJointSetLinearDampingRatio: func(rdr *recReader) {
		var a recArgsWeldJointSetLinearDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWeldJointSetLinearDampingRatio(&a, rdr)
		}
	},
	recOpWeldJointSetAngularHertz: func(rdr *recReader) {
		var a recArgsWeldJointSetAngularHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWeldJointSetAngularHertz(&a, rdr)
		}
	},
	recOpWeldJointSetAngularDampingRatio: func(rdr *recReader) {
		var a recArgsWeldJointSetAngularDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWeldJointSetAngularDampingRatio(&a, rdr)
		}
	},
	recOpCreateWheelJoint: func(rdr *recReader) {
		var a recArgsCreateWheelJoint

		a.def = recRWHEELJOINTDEF(rdr)
		if recRdrCheckEnd(rdr, 8) {
			recDispatchCreateWheelJoint(&a, rdr)
		}
	},
	recOpWheelJointEnableSpring: func(rdr *recReader) {
		var a recArgsWheelJointEnableSpring

		a.joint = recRJOINTID(rdr)

		a.enableSpring = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointEnableSpring(&a, rdr)
		}
	},
	recOpWheelJointSetSpringHertz: func(rdr *recReader) {
		var a recArgsWheelJointSetSpringHertz

		a.joint = recRJOINTID(rdr)

		a.hertz = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointSetSpringHertz(&a, rdr)
		}
	},
	recOpWheelJointSetSpringDampingRatio: func(rdr *recReader) {
		var a recArgsWheelJointSetSpringDampingRatio

		a.joint = recRJOINTID(rdr)

		a.dampingRatio = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointSetSpringDampingRatio(&a, rdr)
		}
	},
	recOpWheelJointEnableLimit: func(rdr *recReader) {
		var a recArgsWheelJointEnableLimit

		a.joint = recRJOINTID(rdr)

		a.enableLimit = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointEnableLimit(&a, rdr)
		}
	},
	recOpWheelJointSetLimits: func(rdr *recReader) {
		var a recArgsWheelJointSetLimits

		a.joint = recRJOINTID(rdr)
		a.lower = recRF32(rdr)

		a.upper = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointSetLimits(&a, rdr)
		}
	},
	recOpWheelJointEnableMotor: func(rdr *recReader) {
		var a recArgsWheelJointEnableMotor

		a.joint = recRJOINTID(rdr)

		a.enableMotor = recRBOOL(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointEnableMotor(&a, rdr)
		}
	},
	recOpWheelJointSetMotorSpeed: func(rdr *recReader) {
		var a recArgsWheelJointSetMotorSpeed

		a.joint = recRJOINTID(rdr)

		a.motorSpeed = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointSetMotorSpeed(&a, rdr)
		}
	},
	recOpWheelJointSetMaxMotorTorque: func(rdr *recReader) {
		var a recArgsWheelJointSetMaxMotorTorque

		a.joint = recRJOINTID(rdr)

		a.torque = recRF32(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchWheelJointSetMaxMotorTorque(&a, rdr)
		}
	},
	recOpQueryOverlapAABB: func(rdr *recReader) {
		var a recArgsQueryOverlapAABB

		a.origin = recRPOSITION(rdr)
		a.aabb = recRAABB(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryOverlapAABB(&a, rdr)
		}
	},
	recOpQueryOverlapShape: func(rdr *recReader) {
		var a recArgsQueryOverlapShape

		a.origin = recRPOSITION(rdr)
		a.proxy = recRSHAPEPROXY(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryOverlapShape(&a, rdr)
		}
	},
	recOpQueryCastRay: func(rdr *recReader) {
		var a recArgsQueryCastRay

		a.origin = recRPOSITION(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastRay(&a, rdr)
		}
	},
	recOpQueryCastShape: func(rdr *recReader) {
		var a recArgsQueryCastShape

		a.origin = recRPOSITION(rdr)
		a.proxy = recRSHAPEPROXY(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastShape(&a, rdr)
		}
	},
	recOpQueryCollideMover: func(rdr *recReader) {
		var a recArgsQueryCollideMover

		a.origin = recRPOSITION(rdr)
		a.mover = recRCAPSULE(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCollideMover(&a, rdr)
		}
	},
	recOpQueryCastRayClosest: func(rdr *recReader) {
		var a recArgsQueryCastRayClosest

		a.origin = recRPOSITION(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastRayClosest(&a, rdr)
		}
	},
	recOpQueryCastMover: func(rdr *recReader) {
		var a recArgsQueryCastMover

		a.origin = recRPOSITION(rdr)
		a.mover = recRCAPSULE(rdr)
		a.translation = recRVEC2(rdr)

		a.filter = recRQUERYFILTER(rdr)
		if rdr.ok {
			recDispatchQueryCastMover(&a, rdr)
		}
	},
	recOpShapeTestPoint: func(rdr *recReader) {
		var a recArgsShapeTestPoint

		a.shape = recRSHAPEID(rdr)

		a.point = recRPOSITION(rdr)
		if rdr.ok {
			recDispatchShapeTestPoint(&a, rdr)
		}
	},
	recOpShapeRayCast: func(rdr *recReader) {
		var a recArgsShapeRayCast

		a.shape = recRSHAPEID(rdr)
		a.origin = recRPOSITION(rdr)

		a.translation = recRVEC2(rdr)
		if rdr.ok {
			recDispatchShapeRayCast(&a, rdr)
		}
	},
	recOpStateHash: func(rdr *recReader) {
		var a recArgsStateHash

		a.hash = recRU64(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchStateHash(&a, rdr)
		}
	},
	recOpRecordingBounds: func(rdr *recReader) {
		var a recArgsRecordingBounds

		a.bounds = recRAABB(rdr)
		if recRdrCheckEnd(rdr, 0) {
			recDispatchRecordingBounds(&a, rdr)
		}
	},
}

// The kind of a recorded spatial query, matching the public query and cast functions.
type (
	ReplayQueryType int
	recQueryKind    int
	// Static metadata describing a recording, resolved once when the player opens the file.
	ReplayInfo struct {
		FrameCount   int
		WorkerCount  int
		TimeStep     float32
		SubStepCount int
		LengthScale  float32
		Bounds       AABB
	}
)

// A spatial query recorded during a replayed frame, exposed for inspection.
type ReplayQueryInfo struct {
	Type        ReplayQueryType
	Filter      QueryFilter
	AABB        AABB
	Origin      Pos
	Translation Vec2
	Shape       Shape
	HitCount    int
}

// One result of a recorded spatial query.
type ReplayQueryHit struct {
	Shape    Shape
	Point    Pos
	Normal   Vec2
	Fraction float32
}

type recRecordedHit struct {
	id          Shape
	point       Pos
	normal      Vec2
	fraction    float32
	plane       PlaneResult
	userReturnF float32
	userReturnB bool
}

type recDrawQuery struct {
	kind         int
	filter       QueryFilter
	aabb         AABB
	proxy        ShapeProxy
	mover        Capsule
	origin       Pos
	translation  Vec2
	boolResult   bool
	castFraction float32
	castOut      WorldCastOutput
	shape        Shape
	hitStart     int
	hitCount     int
}

type recReader struct {
	data             []byte
	size             int
	cursor           int
	replayWorldID    *World
	workerCount      int
	ok               bool
	diverged         bool
	chainPoints      []Vec2
	chainPointCap    int
	chainMaterials   []SurfaceMaterial
	chainMaterialCap int
	hits             []recRecordedHit
	hitCap           int
	owner            *Replay
}

type recKeyframe struct {
	image         []byte
	imageSize     int
	imageCapacity int
	frame         int
	cursor        int
	bodyIDs       []Body
	bodyIDCount   int
	divergeFrame  int
	diverged      bool
}

// Opaque handle for incremental playback of a recording.
type Replay struct {
	data                 []byte
	size                 int
	headerEnd            int
	lengthScale          float32
	previousLengthScale  float32
	frame                int
	frameCount           int
	recordedWorkerCount  int
	recordedDt           float32
	recordedSubStepCount int
	bounds               AABB
	divergeFrame         int
	atEnd                bool
	rdr                  recReader
	frameQueries         []recDrawQuery
	frameQueryCount      int
	frameQueryCap        int
	frameHits            []recRecordedHit
	frameHitCount        int
	frameHitCap          int
	bodyIDs              []Body
	bodyIDCount          int
	bodyIDCap            int
	frame0Image          []byte
	frame0Size           int
	frame0BodyIDs        []Body
	frame0BodyIDCount    int
	keyframes            []recKeyframe
	keyframeCount        int
	keyframeCapacity     int
	keyframeBudget       uint
	keyframeBytes        uint
	keyframeMinInterval  int
	keyframeInterval     int
	lastKeyframeFrame    int
}

type recReplayQueryCtx struct {
	rdr    *recReader
	hits   []recRecordedHit
	count  int
	cursor int
}

// Open a recording for incremental playback. The player copies the bytes,
// so you may free or destroy the source buffer immediately after this call.
func NewReplay(data []byte, workerCount int) (*Replay, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("%w: recording too small", ErrCorrupt)
	}
	// Validate the header before copying anything
	magic := binary.LittleEndian.Uint32(data)
	major := binary.LittleEndian.Uint16(data[4:])
	minor := binary.LittleEndian.Uint16(data[6:])
	lengthScale := math.Float32frombits(binary.LittleEndian.Uint32(data[12:]))

	if magic != recMagic {
		return nil, fmt.Errorf("%w: recording magic 0x%08X", ErrBadMagic, magic)
	}

	if major != recVersionMajor || minor != recVersionMinor {
		return nil, fmt.Errorf(
			"%w: recording %d.%d, runtime %d.%d",
			ErrVersion,
			major,
			minor,
			recVersionMajor,
			recVersionMinor,
		)
	}

	if data[18] != 0 {
		return nil, fmt.Errorf("%w: big-endian recording", ErrLayout)
	}

	snapshotSize := binary.LittleEndian.Uint64(data[24:])
	if snapshotSize == 0 || snapshotSize > uint64(len(data)-32) {
		return nil, fmt.Errorf("%w: missing or oversized snapshot", ErrCorrupt)
	}

	headerEnd := 32 + int(snapshotSize)

	owned := append([]byte(nil), data...)
	player := &Replay{
		data:                owned,
		size:                len(owned),
		headerEnd:           headerEnd,
		lengthScale:         lengthScale,
		previousLengthScale: LengthUnitsPerMeter(),
		divergeFrame:        -1,
		keyframeBudget:      recKeyframeBudgetDefault,
		keyframeMinInterval: recKeyframeIntervalDefault,
		keyframeInterval:    recKeyframeIntervalDefault,
	}
	player.rdr = recReader{
		data:        owned,
		size:        len(owned),
		cursor:      headerEnd,
		workerCount: workerCount,
		ok:          true,
		owner:       player,
	}
	// Override the global length scale with the recording's so replay reproduces the same constants.
	// This is global engine state and affects the caller's other worlds, so the previous value was
	// captured above and is restored in Destroy.
	if lengthScale > 0 {
		SetLengthUnitsPerMeter(lengthScale)
	}

	recScanFile(player)

	world, err := NewWorldFromSnapshot(owned[32:headerEnd], workerCount)

	player.rdr.replayWorldID = world
	if err != nil {
		player.Close()
		return nil, fmt.Errorf("box2d: recording snapshot: %w", err)
	}

	player.recordedWorkerCount = workerCount
	player.frame0Image = owned[32:headerEnd]
	player.frame0Size = int(snapshotSize)
	recSeedBodyIDs(player)

	player.frame0BodyIDCount = player.bodyIDCount
	if player.bodyIDCount > 0 {
		player.frame0BodyIDs = append([]Body(nil), player.bodyIDs[:player.bodyIDCount]...)
	}

	return player, nil
}

// Replay a recording by re-running the engine and checking that recorded ids and state match.
// It returns ErrDiverged when the replay diverges and ErrCorrupt when a frame cannot be read.
func ValidateReplay(data []byte, workerCount int) error {
	player, err := NewReplay(data, workerCount)
	if err != nil {
		return err
	}
	defer player.Close()

	for player.StepFrame() {
		if player.rdr.diverged {
			break
		}
	}

	if player.rdr.diverged {
		return ErrDiverged
	}

	if !player.rdr.ok {
		return ErrCorrupt
	}

	return nil
}

func recRdrCheck(rdr *recReader, size int) {
	if size < 0 || rdr.cursor < 0 || rdr.cursor > rdr.size || size > rdr.size-rdr.cursor ||
		rdr.size > len(rdr.data) {
		rdr.ok = false
	}
}

func recRdrCheckEnd(rdr *recReader, trailingBytes int) bool {
	rdr.ok = rdr.ok && rdr.size-rdr.cursor == trailingBytes
	return rdr.ok
}

func recRU8(rdr *recReader) uint8 {
	recRdrCheck(rdr, 1)

	if !rdr.ok {
		return 0
	}

	v := rdr.data[rdr.cursor+0] << 0
	rdr.cursor += 1

	return v
}

func recRU16(rdr *recReader) uint16 {
	recRdrCheck(rdr, 2)

	if !rdr.ok {
		return 0
	}

	v := uint16(rdr.data[rdr.cursor+0])<<0 | uint16(rdr.data[rdr.cursor+1])<<8
	rdr.cursor += 2

	return v
}

func recRU24(rdr *recReader) uint32 {
	recRdrCheck(rdr, 3)

	if !rdr.ok {
		return 0
	}

	v := uint32(
		rdr.data[rdr.cursor+0],
	)<<0 | uint32(
		rdr.data[rdr.cursor+1],
	)<<8 | uint32(
		rdr.data[rdr.cursor+2],
	)<<16
	rdr.cursor += 3

	return v
}

func recRU32(rdr *recReader) uint32 {
	recRdrCheck(rdr, 4)

	if !rdr.ok {
		return 0
	}

	v := uint32(
		rdr.data[rdr.cursor+0],
	)<<0 | uint32(
		rdr.data[rdr.cursor+1],
	)<<8 | uint32(
		rdr.data[rdr.cursor+2],
	)<<16 | uint32(
		rdr.data[rdr.cursor+3],
	)<<24
	rdr.cursor += 4

	return v
}

func recRU64(rdr *recReader) uint64 {
	recRdrCheck(rdr, 8)

	if !rdr.ok {
		return 0
	}

	v := uint64(
		rdr.data[rdr.cursor+0],
	)<<0 | uint64(
		rdr.data[rdr.cursor+1],
	)<<8 | uint64(
		rdr.data[rdr.cursor+2],
	)<<16 | uint64(
		rdr.data[rdr.cursor+3],
	)<<24 | uint64(
		rdr.data[rdr.cursor+4],
	)<<32 | uint64(
		rdr.data[rdr.cursor+5],
	)<<40 | uint64(
		rdr.data[rdr.cursor+6],
	)<<48 | uint64(
		rdr.data[rdr.cursor+7],
	)<<56
	rdr.cursor += 8

	return v
}

func recRI32(rdr *recReader) int32 {
	return int32(recRU32(rdr))
}

func recRF32(rdr *recReader) float32 {
	return math.Float32frombits(recRU32(rdr))
}

func recRBOOL(rdr *recReader) bool {
	return recRU8(rdr) != 0
}

func recRVEC2(rdr *recReader) Vec2 {
	var v Vec2

	v.X = recRF32(rdr)
	v.Y = recRF32(rdr)

	return v
}

func recRROT(rdr *recReader) Rot {
	var r Rot

	r.C = recRF32(rdr)
	r.S = recRF32(rdr)

	return r
}

func recRXF(rdr *recReader) Transform {
	var xf Transform

	xf.P = recRVEC2(rdr)
	xf.Q = recRROT(rdr)

	return xf
}

func recRF64(rdr *recReader) float64 {
	return math.Float64frombits(recRU64(rdr))
}

func recRPOSITION(rdr *recReader) Pos {
	if doublePrecision {
		return Pos{X: posScalar(recRF64(rdr)), Y: posScalar(recRF64(rdr))}
	}

	return Pos{X: posScalar(recRF32(rdr)), Y: posScalar(recRF32(rdr))}
}

func recRWORLDXF(rdr *recReader) WorldTransform {
	var t WorldTransform

	t.P = recRPOSITION(rdr)
	t.Q = recRROT(rdr)

	return t
}

func recRBODYID(rdr *recReader) Body {
	x := recRU64(rdr)
	return Body{index1: int32(x >> 32), generation: uint16(x)}
}

func recRSHAPEID(rdr *recReader) Shape {
	x := recRU64(rdr)
	return Shape{index1: int32(x >> 32), generation: uint16(x)}
}

func recRCHAINID(rdr *recReader) Chain {
	x := recRU64(rdr)
	return Chain{index1: int32(x >> 32), generation: uint16(x)}
}

func recRJOINTID(rdr *recReader) Joint {
	x := recRU64(rdr)
	return Joint{index1: int32(x >> 32), generation: uint16(x)}
}

func recRCIRCLE(rdr *recReader) Circle {
	recRdrCheck(rdr, 12)

	if !rdr.ok {
		return Circle{}
	}

	return Circle{Center: recRVEC2(rdr), Radius: recRF32(rdr)}
}

func recRCAPSULE(rdr *recReader) Capsule {
	recRdrCheck(rdr, 20)

	if !rdr.ok {
		return Capsule{}
	}

	return Capsule{Center1: recRVEC2(rdr), Center2: recRVEC2(rdr), Radius: recRF32(rdr)}
}

func recRSEGMENT(rdr *recReader) Segment {
	recRdrCheck(rdr, 16)

	if !rdr.ok {
		return Segment{}
	}

	return Segment{Point1: recRVEC2(rdr), Point2: recRVEC2(rdr)}
}

func recRPOLYGON(rdr *recReader) Polygon {
	recRdrCheck(rdr, 144)

	if !rdr.ok {
		return Polygon{}
	}

	var p Polygon
	for i := range p.Vertices {
		p.Vertices[i] = recRVEC2(rdr)
	}

	for i := range p.Normals {
		p.Normals[i] = recRVEC2(rdr)
	}

	p.Centroid = recRVEC2(rdr)
	p.Radius = recRF32(rdr)
	p.Count = int(recRI32(rdr))

	return p
}

func recRCHAINSEG(rdr *recReader) ChainSegment {
	recRdrCheck(rdr, 36)

	if !rdr.ok {
		return ChainSegment{}
	}

	return ChainSegment{
		Ghost1:  recRVEC2(rdr),
		Segment: recRSEGMENT(rdr),
		Ghost2:  recRVEC2(rdr),
		ChainID: int(recRI32(rdr)),
	}
}

func recRFILTER(rdr *recReader) Filter {
	var f Filter

	f.CategoryBits = recRU64(rdr)
	f.MaskBits = recRU64(rdr)
	f.GroupIndex = int(recRI32(rdr))

	return f
}

func recRMATERIAL(rdr *recReader) SurfaceMaterial {
	m := DefaultSurfaceMaterial()
	m.Friction = recRF32(rdr)
	m.Restitution = recRF32(rdr)
	m.RollingResistance = recRF32(rdr)
	m.TangentSpeed = recRF32(rdr)
	m.UserMaterialID = recRU64(rdr)
	m.CustomColor = recRU32(rdr)

	return m
}

func recRMASSDATA(rdr *recReader) MassData {
	var md MassData

	md.Mass = recRF32(rdr)
	md.Center = recRVEC2(rdr)
	md.RotationalInertia = recRF32(rdr)

	return md
}

func recRLOCKS(rdr *recReader) MotionLocks {
	var locks MotionLocks

	locks.LinearX = recRBOOL(rdr)
	locks.LinearY = recRBOOL(rdr)
	locks.AngularZ = recRBOOL(rdr)

	return locks
}

func recRSTR(rdr *recReader) string {
	length := int(recRU16(rdr))
	if length == 0xffff {
		return ""
	}

	n := min(length, NameLength)
	recRdrCheck(rdr, length)

	var name string

	if rdr.ok && n > 0 {
		value := rdr.data[rdr.cursor : rdr.cursor+n]
		if end := bytes.IndexByte(value, 0); end >= 0 {
			value = value[:end]
		}

		name = string(value)
	}

	rdr.cursor += length

	return name
}

func recRBODYDEF(rdr *recReader) BodyDef {
	def := DefaultBodyDef()
	def.Type = BodyType(recRI32(rdr))
	def.Position = recRPOSITION(rdr)
	def.Rotation = recRROT(rdr)
	def.LinearVelocity = recRVEC2(rdr)
	def.AngularVelocity = recRF32(rdr)
	def.LinearDamping = recRF32(rdr)
	def.AngularDamping = recRF32(rdr)
	def.GravityScale = recRF32(rdr)
	def.SleepThreshold = recRF32(rdr)
	def.SafetyFactor = recRF32(rdr)

	// recRSTR handles the over-length clamp and skips the full recorded length, so the
	// cursor stays aligned even for names longer than NameLength. Valid until the create call.
	def.Name = recRSTR(rdr)

	_ = recRU64(rdr) // userData (not preserved)
	def.MotionLocks = recRLOCKS(rdr)
	def.EnableSleep = recRBOOL(rdr)
	def.IsAwake = recRBOOL(rdr)
	def.IsBullet = recRBOOL(rdr)
	def.IsEnabled = recRBOOL(rdr)
	def.AllowFastRotation = recRBOOL(rdr)
	def.EnableContactRecycling = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recRSHAPEDEF(rdr *recReader) ShapeDef {
	def := DefaultShapeDef()
	_ = recRU64(rdr) // userData (not preserved)
	def.Material = recRMATERIAL(rdr)
	def.Density = recRF32(rdr)
	def.Filter = recRFILTER(rdr)
	def.EnableCustomFiltering = recRBOOL(rdr)
	def.IsSensor = recRBOOL(rdr)
	def.EnableSensorEvents = recRBOOL(rdr)
	def.EnableContactEvents = recRBOOL(rdr)
	def.EnableHitEvents = recRBOOL(rdr)
	def.EnablePreSolveEvents = recRBOOL(rdr)
	def.InvokeContactCreation = recRBOOL(rdr)
	def.UpdateBodyMass = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recRCHAINDEF(rdr *recReader) ChainDef {
	def := DefaultChainDef()
	_ = recRU64(rdr)

	pointCount := max(int(recRI32(rdr)), 0)

	if !recReserveScratch(rdr, &rdr.chainPoints, &rdr.chainPointCap, pointCount, 8) {
		pointCount = 0
	}

	for i := range pointCount {
		rdr.chainPoints[i] = recRVEC2(rdr)
	}

	def.Points = rdr.chainPoints[:pointCount]
	def.Ghost1 = recRVEC2(rdr)
	def.Ghost2 = recRVEC2(rdr)

	materialCount := max(int(recRI32(rdr)), 0)

	if !recReserveScratch(rdr, &rdr.chainMaterials, &rdr.chainMaterialCap, materialCount, 32) {
		materialCount = 0
	}

	for i := range materialCount {
		rdr.chainMaterials[i] = recRMATERIAL(rdr)
	}

	def.Materials = rdr.chainMaterials[:materialCount]
	def.Filter = recRFILTER(rdr)
	def.IsLoop = recRBOOL(rdr)
	def.EnableSensorEvents = recRBOOL(rdr)
	def.UserData = nil

	return def
}

func recREXPLOSIONDEF(rdr *recReader) ExplosionDef {
	def := DefaultExplosionDef()
	def.MaskBits = recRU64(rdr)
	def.Position = recRPOSITION(rdr)
	def.Radius = recRF32(rdr)
	def.Falloff = recRF32(rdr)
	def.ImpulsePerLength = recRF32(rdr)

	return def
}

func recRJointBase(rdr *recReader, base *JointDef) {
	_ = recRU64(rdr) // userData
	base.BodyA = recRBODYID(rdr)
	base.BodyB = recRBODYID(rdr)
	base.LocalFrameA = recRXF(rdr)
	base.LocalFrameB = recRXF(rdr)
	base.ForceThreshold = recRF32(rdr)
	base.TorqueThreshold = recRF32(rdr)
	base.ConstraintHertz = recRF32(rdr)
	base.ConstraintDampingRatio = recRF32(rdr)
	base.DrawScale = recRF32(rdr)
	base.CollideConnected = recRBOOL(rdr)
	base.UserData = nil
}

func recRDISTANCEJOINTDEF(rdr *recReader) DistanceJointDef {
	def := DefaultDistanceJointDef()
	recRJointBase(rdr, &def.Base)
	def.Length = recRF32(rdr)
	def.EnableSpring = recRBOOL(rdr)
	def.LowerSpringForce = recRF32(rdr)
	def.UpperSpringForce = recRF32(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.MinLength = recRF32(rdr)
	def.MaxLength = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorForce = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRFILTERJOINTDEF(rdr *recReader) FilterJointDef {
	def := DefaultFilterJointDef()
	recRJointBase(rdr, &def.Base)

	return def
}

func recRMOTORJOINTDEF(rdr *recReader) MotorJointDef {
	def := DefaultMotorJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearVelocity = recRVEC2(rdr)
	def.MaxVelocityForce = recRF32(rdr)
	def.AngularVelocity = recRF32(rdr)
	def.MaxVelocityTorque = recRF32(rdr)
	def.LinearHertz = recRF32(rdr)
	def.LinearDampingRatio = recRF32(rdr)
	def.MaxSpringForce = recRF32(rdr)
	def.AngularHertz = recRF32(rdr)
	def.AngularDampingRatio = recRF32(rdr)
	def.MaxSpringTorque = recRF32(rdr)

	return def
}

func recRMOVERJOINTDEF(rdr *recReader) MoverJointDef {
	def := DefaultMoverJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearVelocity = recRVEC2(rdr)
	def.MaxVelocityForce = recRVEC2(rdr)

	return def
}

func recRPRISMATICJOINTDEF(rdr *recReader) PrismaticJointDef {
	def := DefaultPrismaticJointDef()
	recRJointBase(rdr, &def.Base)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.TargetTranslation = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerTranslation = recRF32(rdr)
	def.UpperTranslation = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorForce = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRPOGOJOINTDEF(rdr *recReader) PogoJointDef {
	def := DefaultPogoJointDef()
	recRJointBase(rdr, &def.Base)
	def.Normal = recRVEC2(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.RestLength = recRF32(rdr)
	def.MaxTensionForce = recRF32(rdr)
	def.MaxCompressionForce = recRF32(rdr)
	def.Impulse = recRF32(rdr)
	def.Velocity = recRF32(rdr)

	return def
}

func recRREVOLUTEJOINTDEF(rdr *recReader) RevoluteJointDef {
	def := DefaultRevoluteJointDef()
	recRJointBase(rdr, &def.Base)
	def.TargetAngle = recRF32(rdr)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerAngle = recRF32(rdr)
	def.UpperAngle = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorTorque = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRWELDJOINTDEF(rdr *recReader) WeldJointDef {
	def := DefaultWeldJointDef()
	recRJointBase(rdr, &def.Base)
	def.LinearHertz = recRF32(rdr)
	def.AngularHertz = recRF32(rdr)
	def.LinearDampingRatio = recRF32(rdr)
	def.AngularDampingRatio = recRF32(rdr)

	return def
}

func recRWHEELJOINTDEF(rdr *recReader) WheelJointDef {
	def := DefaultWheelJointDef()
	recRJointBase(rdr, &def.Base)
	def.EnableSpring = recRBOOL(rdr)
	def.Hertz = recRF32(rdr)
	def.DampingRatio = recRF32(rdr)
	def.EnableLimit = recRBOOL(rdr)
	def.LowerTranslation = recRF32(rdr)
	def.UpperTranslation = recRF32(rdr)
	def.EnableMotor = recRBOOL(rdr)
	def.MaxMotorTorque = recRF32(rdr)
	def.MotorSpeed = recRF32(rdr)

	return def
}

func recRAABB(rdr *recReader) AABB {
	var v AABB

	v.LowerBound = recRVEC2(rdr)
	v.UpperBound = recRVEC2(rdr)

	return v
}

func recRQUERYFILTER(rdr *recReader) QueryFilter {
	var f QueryFilter

	f.CategoryBits = recRU64(rdr)
	f.MaskBits = recRU64(rdr)

	return f
}

func recRSHAPEPROXY(rdr *recReader) ShapeProxy {
	var p ShapeProxy

	count := min(max(int(recRI32(rdr)), 0), MaxPolygonVertices)

	p.Count = count
	for i := range count {
		p.Points[i] = recRVEC2(rdr)
	}

	p.Radius = recRF32(rdr)

	return p
}

func recRWORLDCASTOUTPUT(rdr *recReader) WorldCastOutput {
	var v WorldCastOutput

	v.Normal = recRVEC2(rdr)
	v.Point = recRPOSITION(rdr)
	v.Fraction = recRF32(rdr)
	v.Iterations = int(recRI32(rdr))
	v.Hit = recRBOOL(rdr)

	return v
}

func recRRAYRESULT(rdr *recReader) RayResult {
	var v RayResult
	// ShapeID has no world; recMakeShapeID is applied at compare time
	v.Shape = recRSHAPEID(rdr)
	v.Point = recRPOSITION(rdr)
	v.Normal = recRVEC2(rdr)
	v.Fraction = recRF32(rdr)
	v.NodeVisits = int(recRI32(rdr))
	v.LeafVisits = int(recRI32(rdr))
	v.Hit = recRBOOL(rdr)

	return v
}

func recRPLANERESULT(rdr *recReader) PlaneResult {
	var v PlaneResult

	v.Plane.Normal = recRVEC2(rdr)
	v.Plane.Offset = recRF32(rdr)
	v.Point = recRVEC2(rdr)
	v.Hit = recRBOOL(rdr)

	return v
}

func recRTREESTATS(rdr *recReader) {
	recRI32(rdr)
	recRI32(rdr)
}

func recReserveScratch[T any](
	rdr *recReader,
	data *[]T,
	capacity *int,
	need int,
	elemSize int,
) bool {
	remaining := rdr.size - rdr.cursor
	if need < 0 || remaining < 0 || need > remaining || elemSize <= 0 ||
		need > math.MaxInt32/elemSize {
		rdr.ok = false
		return false
	}

	if need <= *capacity {
		return true
	}

	newCap := need
	if need <= math.MaxInt32/elemSize-8 {
		newCap = need + 8
	}

	*data = make([]T, newCap)
	*capacity = newCap

	return true
}

func recGrow[T any](data *[]T, capacity *int, need int, keep int) {
	if need <= *capacity {
		return
	}

	newCap := 8
	if *capacity != 0 {
		newCap = 2 * *capacity
	}

	if newCap < need {
		newCap = need
	}

	grown := make([]T, newCap)
	if *data != nil && keep > 0 {
		copy(grown[:keep], (*data)[:keep])
	}

	*data = grown
	*capacity = newCap
}

func recEnsureHits(rdr *recReader, n int) {
	recReserveScratch(rdr, &rdr.hits, &rdr.hitCap, n, int(unsafe.Sizeof(recRecordedHit{})))
}

func recMakeBodyID(rdr *recReader, recorded Body) Body {
	return Body{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeShapeID(rdr *recReader, recorded Shape) Shape {
	return Shape{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeChainID(rdr *recReader, recorded Chain) Chain {
	return Chain{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recMakeJointID(rdr *recReader, recorded Joint) Joint {
	return Joint{index1: recorded.index1, world: rdr.replayWorldID, generation: recorded.generation}
}

func recCheckID(
	rdr *recReader,
	kind string,
	gotIndex int,
	gotGen uint32,
	recIndex int,
	recGen uint32,
) {
	if gotIndex != recIndex || gotGen != recGen {
		rdr.replayWorldID.logger.Warn(
			"box2d: replay id mismatch",
			"kind",
			kind,
			"recordedIndex",
			recIndex,
			"recordedGeneration",
			recGen,
			"index",
			gotIndex,
			"generation",
			gotGen,
		)
		rdr.ok = false
	}
}

func recCheckBodyID(rdr *recReader, got Body, rec Body) {
	recCheckID(
		rdr,
		"body",
		int(got.index1),
		uint32(got.generation),
		int(rec.index1),
		uint32(rec.generation),
	)
}

func recCheckShapeID(rdr *recReader, got Shape, rec Shape) {
	recCheckID(
		rdr,
		"shape",
		int(got.index1),
		uint32(got.generation),
		int(rec.index1),
		uint32(rec.generation),
	)
}

func recCheckChainID(rdr *recReader, got Chain, rec Chain) {
	recCheckID(
		rdr,
		"chain",
		int(got.index1),
		uint32(got.generation),
		int(rec.index1),
		uint32(rec.generation),
	)
}

func recCheckJointID(rdr *recReader, got Joint, rec Joint) {
	recCheckID(
		rdr,
		"joint",
		int(got.index1),
		uint32(got.generation),
		int(rec.index1),
		uint32(rec.generation),
	)
}

func recDispatchDestroyWorld(a *recArgsDestroyWorld, rdr *recReader) {
	_ = a
	_ = rdr
	// The recorded session ended here. The player owns the replay world's lifetime and tears it
	// down in Destroy/Restart, so a viewer can keep drawing the final step. There is
	// one world per recording and this is always the last record, so leaving it alive is safe.
}

func recDispatchStep(a *recArgsStep, rdr *recReader) {
	_ = a
	rdr.replayWorldID.Step(a.dt, int(a.subStepCount))
}

func recDispatchWorldEnableSleeping(a *recArgsWorldEnableSleeping, rdr *recReader) {
	rdr.replayWorldID.SetSleepingEnabled(a.flag)
}

func recDispatchWorldEnableContinuous(a *recArgsWorldEnableContinuous, rdr *recReader) {
	rdr.replayWorldID.SetContinuousEnabled(a.flag)
}

func recDispatchWorldSetRestitutionThreshold(
	a *recArgsWorldSetRestitutionThreshold,
	rdr *recReader,
) {
	rdr.replayWorldID.SetRestitutionThreshold(a.value)
}

func recDispatchWorldSetHitEventThreshold(a *recArgsWorldSetHitEventThreshold, rdr *recReader) {
	rdr.replayWorldID.SetHitEventThreshold(a.value)
}

func recDispatchWorldSetGravity(a *recArgsWorldSetGravity, rdr *recReader) {
	rdr.replayWorldID.SetGravity(a.gravity)
}

func recDispatchWorldExplode(a *recArgsWorldExplode, rdr *recReader) {
	rdr.replayWorldID.Explode(&a.def)
}

func recDispatchWorldSetContactTuning(a *recArgsWorldSetContactTuning, rdr *recReader) {
	rdr.replayWorldID.SetContactTuning(a.hertz, a.dampingRatio, a.pushSpeed)
}

func recDispatchWorldSetContactRecycleDistance(
	a *recArgsWorldSetContactRecycleDistance,
	rdr *recReader,
) {
	rdr.replayWorldID.SetContactRecycleDistance(a.recycleDistance)
}

func recDispatchWorldSetMaximumLinearSpeed(a *recArgsWorldSetMaximumLinearSpeed, rdr *recReader) {
	rdr.replayWorldID.SetMaximumLinearSpeed(a.maximumLinearSpeed)
}

func recDispatchWorldEnableWarmStarting(a *recArgsWorldEnableWarmStarting, rdr *recReader) {
	rdr.replayWorldID.SetWarmStartingEnabled(a.flag)
}

func recDispatchWorldRebuildStaticTree(a *recArgsWorldRebuildStaticTree, rdr *recReader) {
	_ = a

	rdr.replayWorldID.RebuildStaticTree()
}

func recDispatchWorldSetRestitutionIterations(
	a *recArgsWorldSetRestitutionIterations,
	rdr *recReader,
) {
	rdr.replayWorldID.SetRestitutionIterations(int(a.iterations))
}

func recDispatchWorldEnableRestitutionPropagation(
	a *recArgsWorldEnableRestitutionPropagation,
	rdr *recReader,
) {
	rdr.replayWorldID.SetRestitutionPropagationEnabled(a.flag)
}

func recTrackBodyCreate(player *Replay, id Body) {
	recGrow(&player.bodyIDs, &player.bodyIDCap, player.bodyIDCount+1, player.bodyIDCount)
	player.bodyIDs[player.bodyIDCount] = id
	player.bodyIDCount++
}

func recTrackBodyDestroy(player *Replay, id Body) {
	for i := range player.bodyIDCount {
		if player.bodyIDs[i] == id {
			player.bodyIDs[i] = Body{}
			return
		}
	}
}

func recSeedBodyIDs(player *Replay) {
	world := player.rdr.replayWorldID
	player.bodyIDCount = 0

	for i := range world.bodies {
		if world.bodies[i].id != i {
			continue
		}

		recTrackBodyCreate(player, makeBodyID(world, i))
	}
}

func recDispatchCreateBody(a *recArgsCreateBody, rdr *recReader) {
	// Recorded id is appended after args (written before recEndRecord)
	recID := recRBODYID(rdr)
	gotID := rdr.replayWorldID.CreateBody(&a.def)
	recCheckBodyID(rdr, gotID, recID)

	if rdr.owner != nil {
		recTrackBodyCreate(rdr.owner, gotID)
	}
}

func recDispatchDestroyBody(a *recArgsDestroyBody, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	if rdr.owner != nil {
		recTrackBodyDestroy(rdr.owner, id)
	}

	id.Destroy()
}

func recDispatchBodySetTransform(a *recArgsBodySetTransform, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	id.SetTransform(a.position, a.rotation)
}

func recDispatchBodySetLinearVelocity(a *recArgsBodySetLinearVelocity, rdr *recReader) {
	id := recMakeBodyID(rdr, a.body)
	id.SetLinearVelocity(a.v)
}

func recDispatchBodySetType(a *recArgsBodySetType, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetType(BodyType(a.bodyType))
}

func recDispatchBodySetName(a *recArgsBodySetName, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetName(a.name)
}

func recDispatchBodySetAngularVelocity(a *recArgsBodySetAngularVelocity, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAngularVelocity(a.w)
}

func recDispatchBodySetTargetTransform(a *recArgsBodySetTargetTransform, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetTargetTransform(a.target, a.timeStep, a.wake)
}

func recDispatchBodyApplyForce(a *recArgsBodyApplyForce, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyForce(a.force, a.point, a.wake)
}

func recDispatchBodyApplyForceToCenter(a *recArgsBodyApplyForceToCenter, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyForceToCenter(a.force, a.wake)
}

func recDispatchBodyApplyTorque(a *recArgsBodyApplyTorque, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyTorque(a.torque, a.wake)
}

func recDispatchBodyClearForces(a *recArgsBodyClearForces, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ClearForces()
}

func recDispatchBodyApplyLinearImpulse(a *recArgsBodyApplyLinearImpulse, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyLinearImpulse(a.impulse, a.point, a.wake)
}

func recDispatchBodyApplyLinearImpulseToCenter(
	a *recArgsBodyApplyLinearImpulseToCenter,
	rdr *recReader,
) {
	recMakeBodyID(rdr, a.body).ApplyLinearImpulseToCenter(a.impulse, a.wake)
}

func recDispatchBodyApplyAngularImpulse(a *recArgsBodyApplyAngularImpulse, rdr *recReader) {
	recMakeBodyID(rdr, a.body).ApplyAngularImpulse(a.impulse, a.wake)
}

func recDispatchBodySetMassData(a *recArgsBodySetMassData, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetMassData(a.massData)
}

func recDispatchBodyUpdateMassFromShapes(a *recArgsBodyUpdateMassFromShapes, rdr *recReader) {
	recMakeBodyID(rdr, a.body).UpdateMassFromShapes()
}

func recDispatchBodySetLinearDamping(a *recArgsBodySetLinearDamping, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetLinearDamping(a.damping)
}

func recDispatchBodySetAngularDamping(a *recArgsBodySetAngularDamping, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAngularDamping(a.damping)
}

func recDispatchBodySetGravityScale(a *recArgsBodySetGravityScale, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetGravityScale(a.scale)
}

func recDispatchBodySetAwake(a *recArgsBodySetAwake, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetAwake(a.awake)
}

func recDispatchBodyWakeTouching(a *recArgsBodyWakeTouching, rdr *recReader) {
	recMakeBodyID(rdr, a.body).WakeTouching()
}

func recDispatchBodyEnableSleep(a *recArgsBodyEnableSleep, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSleepEnabled(a.flag)
}

func recDispatchBodySetSleepThreshold(a *recArgsBodySetSleepThreshold, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSleepThreshold(a.threshold)
}

func recDispatchBodySetSafetyFactor(a *recArgsBodySetSafetyFactor, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetSafetyFactor(a.value)
}

func recDispatchBodyDisable(a *recArgsBodyDisable, rdr *recReader) {
	recMakeBodyID(rdr, a.body).Disable()
}

func recDispatchBodyEnable(a *recArgsBodyEnable, rdr *recReader) {
	recMakeBodyID(rdr, a.body).Enable()
}

func recDispatchBodySetMotionLocks(a *recArgsBodySetMotionLocks, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetMotionLocks(a.locks)
}

func recDispatchBodySetBullet(a *recArgsBodySetBullet, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetBullet(a.flag)
}

func recDispatchBodyEnableContactRecycling(a *recArgsBodyEnableContactRecycling, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetContactRecyclingEnabled(a.flag)
}

func recDispatchBodyEnableContactEvents(a *recArgsBodyEnableContactEvents, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetContactEventsEnabled(a.flag)
}

func recDispatchBodyEnableHitEvents(a *recArgsBodyEnableHitEvents, rdr *recReader) {
	recMakeBodyID(rdr, a.body).SetHitEventsEnabled(a.flag)
}

func recDispatchCreateCircleShape(a *recArgsCreateCircleShape, rdr *recReader) {
	recID := recRSHAPEID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreateCircleShape(&a.def, &a.circle)
	recCheckShapeID(rdr, gotID, recID)
}

func recDispatchCreateCapsuleShape(a *recArgsCreateCapsuleShape, rdr *recReader) {
	recID := recRSHAPEID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreateCapsuleShape(&a.def, &a.capsule)
	recCheckShapeID(rdr, gotID, recID)
}

func recDispatchCreateSegmentShape(a *recArgsCreateSegmentShape, rdr *recReader) {
	recID := recRSHAPEID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreateSegmentShape(&a.def, &a.segment)
	recCheckShapeID(rdr, gotID, recID)
}

func recDispatchCreatePolygonShape(a *recArgsCreatePolygonShape, rdr *recReader) {
	recID := recRSHAPEID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreatePolygonShape(&a.def, &a.polygon)
	recCheckShapeID(rdr, gotID, recID)
}

func recDispatchCreateChainSegmentShape(a *recArgsCreateChainSegmentShape, rdr *recReader) {
	recID := recRSHAPEID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreateChainSegmentShape(&a.def, &a.chainSegment)
	recCheckShapeID(rdr, gotID, recID)
}

func recDispatchDestroyShape(a *recArgsDestroyShape, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).Destroy(a.updateBodyMass)
}

func recDispatchShapeSetDensity(a *recArgsShapeSetDensity, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetDensity(a.density, a.updateBodyMass)
}

func recDispatchShapeSetFriction(a *recArgsShapeSetFriction, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetFriction(a.friction)
}

func recDispatchShapeSetRestitution(a *recArgsShapeSetRestitution, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetRestitution(a.restitution)
}

func recDispatchShapeSetUserMaterial(a *recArgsShapeSetUserMaterial, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetUserMaterial(a.material)
}

func recDispatchShapeSetSurfaceMaterial(a *recArgsShapeSetSurfaceMaterial, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSurfaceMaterial(&a.material)
}

func recDispatchShapeSetFilter(a *recArgsShapeSetFilter, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetFilter(a.filter)
}

func recDispatchShapeEnableSensorEvents(a *recArgsShapeEnableSensorEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSensorEventsEnabled(a.flag)
}

func recDispatchShapeEnableContactEvents(a *recArgsShapeEnableContactEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetContactEventsEnabled(a.flag)
}

func recDispatchShapeEnablePreSolveEvents(a *recArgsShapeEnablePreSolveEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetPreSolveEventsEnabled(a.flag)
}

func recDispatchShapeEnableHitEvents(a *recArgsShapeEnableHitEvents, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetHitEventsEnabled(a.flag)
}

func recDispatchShapeSetCircle(a *recArgsShapeSetCircle, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetCircle(&a.circle)
}

func recDispatchShapeSetCapsule(a *recArgsShapeSetCapsule, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetCapsule(&a.capsule)
}

func recDispatchShapeSetSegment(a *recArgsShapeSetSegment, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetSegment(&a.segment)
}

func recDispatchShapeSetPolygon(a *recArgsShapeSetPolygon, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetPolygon(&a.polygon)
}

func recDispatchShapeSetChainSegment(a *recArgsShapeSetChainSegment, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).SetChainSegment(&a.chainSegment)
}

func recDispatchShapeApplyWind(a *recArgsShapeApplyWind, rdr *recReader) {
	recMakeShapeID(rdr, a.shape).ApplyWind(a.wind, a.drag, a.lift, a.wake)
}

func recDispatchCreateChain(a *recArgsCreateChain, rdr *recReader) {
	recID := recRCHAINID(rdr)
	bodyID := recMakeBodyID(rdr, a.body)
	gotID := bodyID.CreateChain(&a.def)
	recCheckChainID(rdr, gotID, recID)
}

func recDispatchDestroyChain(a *recArgsDestroyChain, rdr *recReader) {
	recMakeChainID(rdr, a.chain).Destroy()
}

func recDispatchChainSetSurfaceMaterial(a *recArgsChainSetSurfaceMaterial, rdr *recReader) {
	recMakeChainID(rdr, a.chain).SetSurfaceMaterial(&a.material, int(a.segmentIndex))
}

func recDispatchChainSetAllSurfaceMaterials(a *recArgsChainSetAllSurfaceMaterials, rdr *recReader) {
	recMakeChainID(rdr, a.chain).SetAllSurfaceMaterials(&a.material)
}

func recDispatchCreateDistanceJoint(a *recArgsCreateDistanceJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateDistanceJoint(&def).Joint, recID)
}

func recDispatchCreateFilterJoint(a *recArgsCreateFilterJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateFilterJoint(&def).Joint, recID)
}

func recDispatchCreateMotorJoint(a *recArgsCreateMotorJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateMotorJoint(&def).Joint, recID)
}

func recDispatchCreateMoverJoint(a *recArgsCreateMoverJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateMoverJoint(&def).Joint, recID)
}

func recDispatchCreatePrismaticJoint(a *recArgsCreatePrismaticJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreatePrismaticJoint(&def).Joint, recID)
}

func recDispatchCreatePogoJoint(a *recArgsCreatePogoJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreatePogoJoint(&def).Joint, recID)
}

func recDispatchCreateRevoluteJoint(a *recArgsCreateRevoluteJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateRevoluteJoint(&def).Joint, recID)
}

func recDispatchCreateWeldJoint(a *recArgsCreateWeldJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateWeldJoint(&def).Joint, recID)
}

func recDispatchCreateWheelJoint(a *recArgsCreateWheelJoint, rdr *recReader) {
	recID := recRJOINTID(rdr)
	def := a.def
	def.Base.BodyA = recMakeBodyID(rdr, def.Base.BodyA)
	def.Base.BodyB = recMakeBodyID(rdr, def.Base.BodyB)
	recCheckJointID(rdr, rdr.replayWorldID.CreateWheelJoint(&def).Joint, recID)
}

func recDispatchDestroyJoint(a *recArgsDestroyJoint, rdr *recReader) {
	recMakeJointID(rdr, a.joint).Destroy()
}

func recDispatchJointSetLocalFrameA(a *recArgsJointSetLocalFrameA, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetLocalFrameA(a.localFrame)
}

func recDispatchJointSetLocalFrameB(a *recArgsJointSetLocalFrameB, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetLocalFrameB(a.localFrame)
}

func recDispatchJointSetCollideConnected(a *recArgsJointSetCollideConnected, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetCollideConnected(a.shouldCollide)
}

func recDispatchJointWakeBodies(a *recArgsJointWakeBodies, rdr *recReader) {
	recMakeJointID(rdr, a.joint).WakeBodies()
}

func recDispatchJointSetConstraintTuning(a *recArgsJointSetConstraintTuning, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetConstraintTuning(a.hertz, a.dampingRatio)
}

func recDispatchJointSetForceThreshold(a *recArgsJointSetForceThreshold, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetForceThreshold(a.threshold)
}

func recDispatchJointSetTorqueThreshold(a *recArgsJointSetTorqueThreshold, rdr *recReader) {
	recMakeJointID(rdr, a.joint).SetTorqueThreshold(a.threshold)
}

func recDispatchDistanceJointSetLength(a *recArgsDistanceJointSetLength, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLength(a.length)
}

func recDispatchDistanceJointEnableSpring(a *recArgsDistanceJointEnableSpring, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchDistanceJointSetSpringForceRange(
	a *recArgsDistanceJointSetSpringForceRange,
	rdr *recReader,
) {
	DistanceJoint{
		Joint: recMakeJointID(rdr, a.joint),
	}.SetSpringForceRange(
		a.lowerForce,
		a.upperForce,
	)
}

func recDispatchDistanceJointSetSpringHertz(a *recArgsDistanceJointSetSpringHertz, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchDistanceJointSetSpringDampingRatio(
	a *recArgsDistanceJointSetSpringDampingRatio,
	rdr *recReader,
) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchDistanceJointEnableLimit(a *recArgsDistanceJointEnableLimit, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchDistanceJointSetLengthRange(a *recArgsDistanceJointSetLengthRange, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLengthRange(a.minLength, a.maxLength)
}

func recDispatchDistanceJointEnableMotor(a *recArgsDistanceJointEnableMotor, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchDistanceJointSetMotorSpeed(a *recArgsDistanceJointSetMotorSpeed, rdr *recReader) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchDistanceJointSetMaxMotorForce(
	a *recArgsDistanceJointSetMaxMotorForce,
	rdr *recReader,
) {
	DistanceJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorForce(a.force)
}

func recDispatchMotorJointSetLinearVelocity(a *recArgsMotorJointSetLinearVelocity, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearVelocity(a.velocity)
}

func recDispatchMotorJointSetAngularVelocity(
	a *recArgsMotorJointSetAngularVelocity,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularVelocity(a.velocity)
}

func recDispatchMotorJointSetMaxVelocityForce(
	a *recArgsMotorJointSetMaxVelocityForce,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityForce(a.maxForce)
}

func recDispatchMotorJointSetMaxVelocityTorque(
	a *recArgsMotorJointSetMaxVelocityTorque,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityTorque(a.maxTorque)
}

func recDispatchMotorJointSetLinearHertz(a *recArgsMotorJointSetLinearHertz, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearHertz(a.hertz)
}

func recDispatchMotorJointSetLinearDampingRatio(
	a *recArgsMotorJointSetLinearDampingRatio,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearDampingRatio(a.damping)
}

func recDispatchMotorJointSetAngularHertz(a *recArgsMotorJointSetAngularHertz, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularHertz(a.hertz)
}

func recDispatchMotorJointSetAngularDampingRatio(
	a *recArgsMotorJointSetAngularDampingRatio,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularDampingRatio(a.damping)
}

func recDispatchMotorJointSetMaxSpringForce(a *recArgsMotorJointSetMaxSpringForce, rdr *recReader) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxSpringForce(a.maxForce)
}

func recDispatchMotorJointSetMaxSpringTorque(
	a *recArgsMotorJointSetMaxSpringTorque,
	rdr *recReader,
) {
	MotorJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxSpringTorque(a.maxTorque)
}

func recDispatchMoverJointSetLinearVelocity(a *recArgsMoverJointSetLinearVelocity, rdr *recReader) {
	MoverJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearVelocity(a.velocity)
}

func recDispatchMoverJointSetMaxVelocityForce(
	a *recArgsMoverJointSetMaxVelocityForce,
	rdr *recReader,
) {
	MoverJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxVelocityForce(a.maxForce)
}

func recDispatchPogoJointSetRestLength(a *recArgsPogoJointSetRestLength, rdr *recReader) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetRestLength(a.length)
}

func recDispatchPogoJointSetSpringHertz(a *recArgsPogoJointSetSpringHertz, rdr *recReader) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchPogoJointSetSpringDampingRatio(
	a *recArgsPogoJointSetSpringDampingRatio,
	rdr *recReader,
) {
	PogoJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchPrismaticJointEnableSpring(a *recArgsPrismaticJointEnableSpring, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchPrismaticJointSetSpringHertz(
	a *recArgsPrismaticJointSetSpringHertz,
	rdr *recReader,
) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchPrismaticJointSetSpringDampingRatio(
	a *recArgsPrismaticJointSetSpringDampingRatio,
	rdr *recReader,
) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchPrismaticJointSetTargetTranslation(
	a *recArgsPrismaticJointSetTargetTranslation,
	rdr *recReader,
) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetTargetTranslation(a.translation)
}

func recDispatchPrismaticJointEnableLimit(a *recArgsPrismaticJointEnableLimit, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchPrismaticJointSetLimits(a *recArgsPrismaticJointSetLimits, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchPrismaticJointEnableMotor(a *recArgsPrismaticJointEnableMotor, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchPrismaticJointSetMotorSpeed(a *recArgsPrismaticJointSetMotorSpeed, rdr *recReader) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchPrismaticJointSetMaxMotorForce(
	a *recArgsPrismaticJointSetMaxMotorForce,
	rdr *recReader,
) {
	PrismaticJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorForce(a.force)
}

func recDispatchRevoluteJointEnableSpring(a *recArgsRevoluteJointEnableSpring, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchRevoluteJointSetSpringHertz(a *recArgsRevoluteJointSetSpringHertz, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchRevoluteJointSetSpringDampingRatio(
	a *recArgsRevoluteJointSetSpringDampingRatio,
	rdr *recReader,
) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchRevoluteJointSetTargetAngle(a *recArgsRevoluteJointSetTargetAngle, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetTargetAngle(a.angle)
}

func recDispatchRevoluteJointEnableLimit(a *recArgsRevoluteJointEnableLimit, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchRevoluteJointSetLimits(a *recArgsRevoluteJointSetLimits, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchRevoluteJointEnableMotor(a *recArgsRevoluteJointEnableMotor, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchRevoluteJointSetMotorSpeed(a *recArgsRevoluteJointSetMotorSpeed, rdr *recReader) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchRevoluteJointSetMaxMotorTorque(
	a *recArgsRevoluteJointSetMaxMotorTorque,
	rdr *recReader,
) {
	RevoluteJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorTorque(a.torque)
}

func recDispatchWeldJointSetLinearHertz(a *recArgsWeldJointSetLinearHertz, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearHertz(a.hertz)
}

func recDispatchWeldJointSetLinearDampingRatio(
	a *recArgsWeldJointSetLinearDampingRatio,
	rdr *recReader,
) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLinearDampingRatio(a.dampingRatio)
}

func recDispatchWeldJointSetAngularHertz(a *recArgsWeldJointSetAngularHertz, rdr *recReader) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularHertz(a.hertz)
}

func recDispatchWeldJointSetAngularDampingRatio(
	a *recArgsWeldJointSetAngularDampingRatio,
	rdr *recReader,
) {
	WeldJoint{Joint: recMakeJointID(rdr, a.joint)}.SetAngularDampingRatio(a.dampingRatio)
}

func recDispatchWheelJointEnableSpring(a *recArgsWheelJointEnableSpring, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringEnabled(a.enableSpring)
}

func recDispatchWheelJointSetSpringHertz(a *recArgsWheelJointSetSpringHertz, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringHertz(a.hertz)
}

func recDispatchWheelJointSetSpringDampingRatio(
	a *recArgsWheelJointSetSpringDampingRatio,
	rdr *recReader,
) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetSpringDampingRatio(a.dampingRatio)
}

func recDispatchWheelJointEnableLimit(a *recArgsWheelJointEnableLimit, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimitEnabled(a.enableLimit)
}

func recDispatchWheelJointSetLimits(a *recArgsWheelJointSetLimits, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetLimits(a.lower, a.upper)
}

func recDispatchWheelJointEnableMotor(a *recArgsWheelJointEnableMotor, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorEnabled(a.enableMotor)
}

func recDispatchWheelJointSetMotorSpeed(a *recArgsWheelJointSetMotorSpeed, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMotorSpeed(a.motorSpeed)
}

func recDispatchWheelJointSetMaxMotorTorque(a *recArgsWheelJointSetMaxMotorTorque, rdr *recReader) {
	WheelJoint{Joint: recMakeJointID(rdr, a.joint)}.SetMaxMotorTorque(a.torque)
}

func recF32Differs(a float32, b float32) bool {
	return math.Float32bits(a) != math.Float32bits(b)
}

func recVec2Differs(a Vec2, b Vec2) bool {
	return recF32Differs(a.X, b.X) || recF32Differs(a.Y, b.Y)
}

func recGrowFrameQueries(player *Replay) {
	recGrow(
		&player.frameQueries,
		&player.frameQueryCap,
		player.frameQueryCount+1,
		player.frameQueryCount,
	)
}

func recGrowFrameHits(player *Replay, need int) {
	recGrow(&player.frameHits, &player.frameHitCap, player.frameHitCount+need, player.frameHitCount)
}

func recStashQueryBegin(
	player *Replay,
	kind int,
	hits []recRecordedHit,
	hitCount int,
) *recDrawQuery {
	recGrowFrameQueries(player)
	q := &player.frameQueries[player.frameQueryCount]
	*q = recDrawQuery{kind: kind, hitStart: player.frameHitCount, hitCount: hitCount}
	recGrowFrameHits(player, hitCount)
	copy(player.frameHits[player.frameHitCount:player.frameHitCount+hitCount], hits[:hitCount])
	player.frameHitCount += hitCount
	player.frameQueryCount++

	return q
}

func recReplayOverlapTrampoline(id Shape, ctx *recReplayQueryCtx) bool {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return false
	}

	h := &rc.hits[rc.cursor]

	rc.cursor++
	if id.index1 != h.id.index1 || id.generation != h.id.generation {
		rc.rdr.diverged = true
	}

	return h.userReturnB
}

func recDispatchQueryOverlapAABB(a *recArgsQueryOverlapAABB, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	recRTREESTATS(rdr)

	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.OverlapAABB(
		a.origin,
		a.aabb,
		a.filter,
		func(id Shape) bool { return recReplayOverlapTrampoline(id, &rc) },
	)

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqOverlapAABB), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.aabb = a.aabb
	}
}

func recDispatchQueryOverlapShape(a *recArgsQueryOverlapShape, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	recRTREESTATS(rdr)

	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.OverlapShape(
		a.origin,
		&a.proxy,
		a.filter,
		func(id Shape) bool { return recReplayOverlapTrampoline(id, &rc) },
	)

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqOverlapShape), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.proxy = a.proxy
	}
}

func recReplayCastTrampoline(
	id Shape,
	point Pos,
	normal Vec2,
	fraction float32,
	ctx *recReplayQueryCtx,
) float32 {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return 0
	}

	h := &rc.hits[rc.cursor]
	rc.cursor++
	// Compare positions through the full width delta, truncating both sides would pass vacuously
	// far from the origin
	if id.index1 != h.id.index1 || id.generation != h.id.generation ||
		recVec2Differs(
			point.Sub(h.point),
			Vec2{},
		) || recVec2Differs(normal, h.normal) || recF32Differs(fraction, h.fraction) {
		rc.rdr.diverged = true
	}

	return h.userReturnF
}

func recDispatchQueryCastRay(a *recArgsQueryCastRay, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].point = recRPOSITION(rdr)
		rdr.hits[i].normal = recRVEC2(rdr)
		rdr.hits[i].fraction = recRF32(rdr)
		rdr.hits[i].userReturnF = recRF32(rdr)
	}

	recRTREESTATS(rdr)

	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CastRay(
		a.origin,
		a.translation,
		a.filter,
		func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
			return recReplayCastTrampoline(id, point, normal, fraction, &rc)
		},
	)

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastRay), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.translation = a.translation
	}
}

func recDispatchQueryCastShape(a *recArgsQueryCastShape, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].point = recRPOSITION(rdr)
		rdr.hits[i].normal = recRVEC2(rdr)
		rdr.hits[i].fraction = recRF32(rdr)
		rdr.hits[i].userReturnF = recRF32(rdr)
	}

	recRTREESTATS(rdr)

	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CastShape(
		a.origin,
		&a.proxy,
		a.translation,
		a.filter,
		func(id Shape, point Pos, normal Vec2, fraction float32) float32 {
			return recReplayCastTrampoline(id, point, normal, fraction, &rc)
		},
	)

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastShape), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.proxy = a.proxy
		q.translation = a.translation
	}
}

func recReplayPlaneTrampoline(id Shape, plane *PlaneResult, ctx *recReplayQueryCtx) bool {
	rc := ctx
	if rc.cursor >= rc.count {
		rc.rdr.diverged = true
		return true
	}

	h := &rc.hits[rc.cursor]

	rc.cursor++
	if id.index1 != h.id.index1 || id.generation != h.id.generation ||
		recVec2Differs(
			plane.Plane.Normal,
			h.plane.Plane.Normal,
		) || recF32Differs(plane.Plane.Offset, h.plane.Plane.Offset) ||
		recVec2Differs(plane.Point, h.plane.Point) || plane.Hit != h.plane.Hit {
		rc.rdr.diverged = true
	}

	return h.userReturnB
}

func recDispatchQueryCollideMover(a *recArgsQueryCollideMover, rdr *recReader) {
	n := int(recRU32(rdr))
	recEnsureHits(rdr, n)

	if !rdr.ok {
		return
	}

	for i := range n {
		rdr.hits[i].id = recMakeShapeID(rdr, recRSHAPEID(rdr))
		rdr.hits[i].plane = recRPLANERESULT(rdr)
		rdr.hits[i].userReturnB = recRBOOL(rdr)
	}

	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	rc := recReplayQueryCtx{rdr: rdr, hits: rdr.hits, count: n}
	rdr.replayWorldID.CollideMover(
		a.origin,
		&a.mover,
		a.filter,
		func(id Shape, plane *PlaneResult) bool { return recReplayPlaneTrampoline(id, plane, &rc) },
	)

	if rc.cursor != n {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCollideMover), rdr.hits, n)
		q.filter = a.filter
		q.origin = a.origin
		q.mover = a.mover
	}
}

func recDispatchQueryCastRayClosest(a *recArgsQueryCastRayClosest, rdr *recReader) {
	rec := recRRAYRESULT(rdr)
	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	got := rdr.replayWorldID.CastRayClosest(a.origin, a.translation, a.filter)
	if got.Hit != rec.Hit ||
		(got.Hit && (got.Shape.index1 != rec.Shape.index1 || got.Shape.generation != rec.Shape.generation ||
			recVec2Differs(got.Point.Sub(rec.Point), Vec2{}) || recVec2Differs(got.Normal, rec.Normal) ||
			recF32Differs(got.Fraction, rec.Fraction))) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		h := recRecordedHit{
			id:       recMakeShapeID(rdr, rec.Shape),
			point:    rec.Point,
			normal:   rec.Normal,
			fraction: rec.Fraction,
		}

		n := 0
		if rec.Hit {
			n = 1
		}

		q := recStashQueryBegin(rdr.owner, int(recqCastRayClosest), []recRecordedHit{h}, n)
		q.filter = a.filter
		q.origin = a.origin
		q.translation = a.translation
	}
}

func recDispatchQueryCastMover(a *recArgsQueryCastMover, rdr *recReader) {
	rec := recRF32(rdr)
	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	got := rdr.replayWorldID.CastMover(a.origin, &a.mover, a.translation, a.filter)
	if recF32Differs(got, rec) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqCastMover), nil, 0)
		q.filter = a.filter
		q.origin = a.origin
		q.mover = a.mover
		q.translation = a.translation
		q.castFraction = rec
	}
}

func recDispatchShapeTestPoint(a *recArgsShapeTestPoint, rdr *recReader) {
	rec := recRBOOL(rdr)
	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	id := recMakeShapeID(rdr, a.shape)

	got := id.TestPoint(a.point)
	if got != rec {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqShapeTestPoint), nil, 0)
		q.shape = id
		q.origin = a.point
		q.boolResult = rec
	}
}

func recDispatchShapeRayCast(a *recArgsShapeRayCast, rdr *recReader) {
	rec := recRWORLDCASTOUTPUT(rdr)
	if !recRdrCheckEnd(rdr, 0) {
		return
	}

	id := recMakeShapeID(rdr, a.shape)

	got := id.RayCast(a.origin, a.translation)
	if got.Hit != rec.Hit || (got.Hit && (recVec2Differs(got.Normal, rec.Normal) ||
		recVec2Differs(got.Point.Sub(rec.Point), Vec2{}) || recF32Differs(got.Fraction, rec.Fraction))) {
		rdr.diverged = true
	}

	if rdr.owner != nil {
		q := recStashQueryBegin(rdr.owner, int(recqShapeRayCast), nil, 0)
		q.shape = id
		q.origin = a.origin
		q.translation = a.translation
		q.castOut = rec
	}
}

func recDispatchStateHash(a *recArgsStateHash, rdr *recReader) {
	world := rdr.replayWorldID

	computed := hashWorldState(world)
	if computed != a.hash {
		world.logger.Warn(
			"box2d: replay state hash mismatch",
			"recorded",
			a.hash,
			"computed",
			computed,
		)

		rdr.diverged = true
	}
}

func recDispatchRecordingBounds(a *recArgsRecordingBounds, rdr *recReader) {
	rdr.owner.bounds = a.bounds
}

func recDispatchOne(player *Replay) int {
	rdr := &player.rdr
	if rdr.cursor >= rdr.size || !rdr.ok {
		return -1
	}

	opcode := recRU8(rdr)
	if !rdr.ok {
		return -1
	}

	payloadSize := recRU24(rdr)
	if !rdr.ok {
		return -1
	}

	payloadStart := rdr.cursor
	if uint64(payloadSize) > uint64(rdr.size-payloadStart) {
		rdr.ok = false
		return -1
	}

	payloadEnd := payloadStart + int(payloadSize)

	if decode := recReplayDecoders[opcode]; decode != nil {
		fileSize := rdr.size
		rdr.size = payloadEnd
		decode(rdr)
		rdr.size = fileSize
	} else {
		rdr.replayWorldID.logger.Warn(
			"box2d: replay skipped unknown opcode",
			"opcode",
			opcode,
			"bytes",
			payloadSize,
		)
		rdr.cursor = payloadEnd
	}

	if !rdr.ok {
		return -1
	}

	return int(opcode)
}

func recScanFile(player *Replay) {
	data := player.data
	size := player.size
	cursor := player.headerEnd
	frameCount := 0
	gotStep := false

	for cursor <= size-4 {
		opcode := data[cursor]
		payloadSize := uint32(
			data[cursor+1],
		) | uint32(
			data[cursor+2],
		)<<8 | uint32(
			data[cursor+3],
		)<<16

		payloadStart := cursor + 4
		if int(payloadSize) > size-payloadStart {
			break
		}

		if opcode == 0x80 {
			frameCount++

			if !gotStep && payloadSize >= 8 {
				player.recordedDt = math.Float32frombits(
					binary.LittleEndian.Uint32(data[payloadStart:]),
				)
				player.recordedSubStepCount = int(
					int32(binary.LittleEndian.Uint32(data[payloadStart+4:])),
				)
				gotStep = true
			}
		} else if opcode == 0xf2 && payloadSize >= 16 {
			rdr := recReader{data: data, size: size, cursor: payloadStart, ok: true}
			player.bounds = recRAABB(&rdr)
		}

		cursor = payloadStart + int(payloadSize)
	}

	player.frameCount = frameCount
}

func freeKeyframe(kf *recKeyframe) {
	kf.image = nil
	kf.bodyIDs = nil
}

func recCaptureKeyframe(player *Replay) {
	world := player.rdr.replayWorldID

	var buf recBuffer
	serializeWorld(world, &buf)

	bodyBytes := uint(player.bodyIDCount) * 8
	newBytes := uint(buf.capacity) + bodyBytes
	// Make room under the budget: doubling the spacing drops the off-grid keyframes, roughly halving
	// the bytes, until the new keyframe fits or only it remains. The budget is soft in the corner
	// where a single snapshot already exceeds it.
	for player.keyframeCount > 0 && player.keyframeBytes+newBytes > player.keyframeBudget {
		player.keyframeInterval *= 2
		kept := 0

		var keptBytes uint

		for i := range player.keyframeCount {
			kf := &player.keyframes[i]
			if kf.frame%player.keyframeInterval == 0 {
				player.keyframes[kept] = *kf
				keptBytes += uint(kf.imageCapacity) + uint(kf.bodyIDCount)*8
				kept++
			} else {
				freeKeyframe(kf)
			}
		}

		progress := kept < player.keyframeCount
		clear(player.keyframes[kept:player.keyframeCount])
		player.keyframeCount = kept
		player.keyframeBytes = keptBytes

		if !progress {
			break
		}
	}

	recGrow(
		&player.keyframes,
		&player.keyframeCapacity,
		player.keyframeCount+1,
		player.keyframeCount,
	)
	kf := &player.keyframes[player.keyframeCount]

	*kf = recKeyframe{
		image:         buf.data,
		imageSize:     buf.size,
		imageCapacity: buf.capacity,
		frame:         player.frame,
		cursor:        player.rdr.cursor,
		divergeFrame:  player.divergeFrame,
		diverged:      player.rdr.diverged,
		bodyIDCount:   player.bodyIDCount,
	}
	if bodyBytes > 0 {
		kf.bodyIDs = append([]Body(nil), player.bodyIDs[:player.bodyIDCount]...)
	}

	player.keyframeBytes += newBytes
	player.keyframeCount++
	player.lastKeyframeFrame = player.frame
}

func replayRestoreKeyframe(player *Replay, kf *recKeyframe) {
	if player.rdr.replayWorldID.UnmarshalBinary(kf.image[:kf.imageSize]) != nil {
		player.rdr.ok = false
		return
	}

	player.rdr.cursor = kf.cursor
	player.rdr.ok = true
	player.rdr.diverged = kf.diverged
	player.frame = kf.frame
	player.divergeFrame = kf.divergeFrame
	player.atEnd = false
	recGrow(&player.bodyIDs, &player.bodyIDCap, kf.bodyIDCount, 0)

	player.bodyIDCount = kf.bodyIDCount
	if kf.bodyIDCount > 0 {
		copy(player.bodyIDs[:kf.bodyIDCount], kf.bodyIDs[:kf.bodyIDCount])
	}
}

// Advance the replay by one recorded step.
func (r *Replay) StepFrame() bool {
	if r.atEnd {
		return false
	}

	r.frameQueryCount = 0
	r.frameHitCount = 0
	// Run this frame's Step, then consume the records that trail it (StateHash, queries, any
	// between-frame mutators) up to the next Step. The queries and hash for a frame are recorded
	// after its Step, so grouping them with that Step keeps them paired with the world state they
	// were computed against. Stopping before the next Step is what advances exactly one frame.
	stepped := false

	for {
		if r.rdr.cursor >= r.rdr.size || !r.rdr.ok {
			r.atEnd = true
			return stepped
		}

		if stepped && r.rdr.data[r.rdr.cursor] == 0x80 {
			r.captureKeyframeIfDue()

			return true
		}

		opcode := recDispatchOne(r)
		if opcode < 0 {
			r.atEnd = true
			return stepped
		}

		if opcode == 0x80 {
			r.frame++
			stepped = true
		}

		if r.divergeFrame < 0 && r.rdr.diverged {
			r.divergeFrame = r.frame
		}
	}
}

// Rewind the player to the first step, recreating the replay world from the file.
func (r *Replay) Restart() {
	if r.rdr.replayWorldID.UnmarshalBinary(r.frame0Image[:r.frame0Size]) != nil {
		r.rdr.ok = false
		return
	}

	r.rdr.cursor = r.headerEnd
	r.rdr.ok = true
	r.rdr.diverged = false
	r.frame = 0
	r.divergeFrame = -1
	r.atEnd = false
	// Frame 0 is the pre-step snapshot, so it has no recorded queries. Clear the per-frame store so
	// the last stepped frame's queries do not linger on a load or a backward scrub to the start.
	r.frameQueryCount = 0
	r.frameHitCount = 0

	r.bodyIDCount = r.frame0BodyIDCount
	if r.frame0BodyIDCount > 0 {
		copy(r.bodyIDs[:r.frame0BodyIDCount], r.frame0BodyIDs[:r.frame0BodyIDCount])
	}
}

// Get the replayed world.
func (r *Replay) World() *World {
	if r == nil {
		return nil
	}

	return r.rdr.replayWorldID
}

// Get the number of steps replayed so far.
func (r *Replay) Frame() int {
	if r == nil {
		return 0
	}

	return r.frame
}

// Seek to a recorded step. Seeking backward rewinds and re-runs from the start, so the
// cost grows with the target frame. Clamps to the recording bounds.
func (r *Replay) SeekFrame(targetFrame int) {
	if r == nil {
		return
	}

	if targetFrame < 0 {
		targetFrame = 0
	}
	// Resume from the nearest keyframe strictly below the target when it beats the current cursor.
	// A backward seek must restore since the cursor cannot rewind. A forward seek restores only when
	// a keyframe sits ahead of the cursor, capping a long forward fling at one keyframe interval of
	// replay instead of every intervening frame. Strictly below so the step loop still runs the
	// target frame and regenerates its per-frame query store, body list, and divergence latch
	// exactly as a plain forward replay would.
	best := r.keyframeBefore(targetFrame)

	if targetFrame < r.frame {
		if best != nil {
			replayRestoreKeyframe(r, best)
		} else {
			r.Restart()
		}
	} else if best != nil && best.frame > r.frame {
		replayRestoreKeyframe(r, best)
	}

	for r.frame < targetFrame && r.StepFrame() {
	}
}

// Get static metadata for the recording (frame count, recorded tuning, time).
func (r *Replay) Info() ReplayInfo {
	if r == nil {
		return ReplayInfo{}
	}

	return ReplayInfo{
		FrameCount: r.frameCount, WorkerCount: r.recordedWorkerCount, TimeStep: r.recordedDt,
		SubStepCount: r.recordedSubStepCount, LengthScale: r.lengthScale, Bounds: r.bounds,
	}
}

// Returns true once the end of the recording has been reached.
func (r *Replay) IsAtEnd() bool {
	if r == nil {
		return true
	}

	return r.atEnd
}

// Returns true if a recorded state hash failed to reproduce, meaning replay diverged.
func (r *Replay) HasDiverged() bool {
	if r == nil {
		return false
	}

	return r.rdr.diverged
}

// Get the first step at which replay diverged, or -1 if it has not diverged.
func (r *Replay) DivergeFrame() int {
	if r == nil {
		return -1
	}

	return r.divergeFrame
}

// Tune the keyframe ring used to speed up backward seeking. A keyframe is a periodic snapshot the
// player restores from instead of replaying from the start, trading memory for seek speed.
// A zero budget or a non-positive interval keeps that value. Clears the existing ring, so call
// Restart afterward to repopulate it under the new policy.
func (r *Replay) SetKeyframePolicy(budgetBytes uint, minIntervalFrames int) {
	if r == nil {
		return
	}

	if budgetBytes > 0 {
		r.keyframeBudget = budgetBytes
	}

	if minIntervalFrames > 0 {
		r.keyframeMinInterval = minIntervalFrames
	}

	for i := range r.keyframeCount {
		freeKeyframe(&r.keyframes[i])
	}

	r.keyframeCount = 0
	r.keyframeBytes = 0
	r.keyframeInterval = r.keyframeMinInterval
	r.lastKeyframeFrame = 0
}

// Get the keyframe memory budget in bytes.
func (r *Replay) KeyframeBudget() uint {
	if r == nil {
		return 0
	}

	return r.keyframeBudget
}

// Get the finest keyframe spacing in frames.
func (r *Replay) KeyframeMinInterval() int {
	if r == nil {
		return 0
	}

	return r.keyframeMinInterval
}

// Get the current keyframe spacing in frames. Starts at the min interval and doubles as the ring
// evicts to stay under budget, so it reflects the effective backward-seek granularity right now.
func (r *Replay) KeyframeInterval() int {
	if r == nil {
		return 0
	}

	return r.keyframeInterval
}

// Get the memory currently held by keyframe snapshots, in bytes.
func (r *Replay) KeyframeBytes() uint {
	if r == nil {
		return 0
	}

	return r.keyframeBytes
}

// Close a player and free its replay world and file buffer.
func (r *Replay) Close() {
	if r == nil {
		return
	}

	if r.rdr.replayWorldID.IsValid() {
		r.rdr.replayWorldID.Close()
	}

	r.data = nil
	r.rdr.data = nil
	r.rdr.chainPoints = nil
	r.rdr.chainMaterials = nil
	r.rdr.hits = nil
	r.frameQueries = nil
	r.frameHits = nil
	r.bodyIDs = nil
	r.frame0Image = nil

	r.frame0BodyIDs = nil
	for i := range r.keyframeCount {
		freeKeyframe(&r.keyframes[i])
	}

	r.keyframes = nil
	SetLengthUnitsPerMeter(r.previousLengthScale)
}

func recDrawHitAABBs(player *Replay, q *recDrawQuery, draw *DebugDraw) {
	for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
		id := player.frameHits[hi].id
		if !id.IsValid() {
			continue
		}

		b := id.AABB()
		lower, upper := b.LowerBound, b.UpperBound
		vs := []Vec2{lower, {X: upper.X, Y: lower.Y}, upper, {X: lower.X, Y: upper.Y}}
		draw.DrawPolygon(WorldTransform{Q: Rot{C: 1}}, vs, ColorMagenta)
	}
}

// Draw spatial queries recorded during the most recently replayed frame.
// Call after World_Draw so queries are layered on top of the world.
func (r *Replay) DrawFrameQueries(draw *DebugDraw, queryIndex int) {
	if r == nil || draw == nil {
		return
	}

	checkInput(draw.Drawer != nil, "draw.Drawer != nil")
	// queryIndex < 0 draws all queries, otherwise just the one selected in the viewer
	for qi := range r.frameQueryCount {
		if queryIndex >= 0 && qi != queryIndex {
			continue
		}

		r.drawFrameQuery(&r.frameQueries[qi], draw)
	}
}

func recDrawQueryProxy(q *recDrawQuery, draw *DebugDraw) {
	// The proxy points are relative to the query origin
	if q.proxy.Count == 1 {
		draw.DrawCircle(q.origin.Add(q.proxy.Points[0]), q.proxy.Radius, ColorLimeGreen)
	} else if q.proxy.Count >= 2 {
		draw.DrawPolygon(
			WorldTransform{P: q.origin, Q: Rot{C: 1}},
			q.proxy.Points[:q.proxy.Count],
			ColorLimeGreen,
		)
	}
}

// Get the number of spatial queries recorded for the most recently replayed frame.
func (r *Replay) FrameQueryCount() int {
	if r == nil {
		return 0
	}

	return r.frameQueryCount
}

// Get a recorded query from the most recently replayed frame by index.
func (r *Replay) FrameQuery(index int) ReplayQueryInfo {
	if r == nil || index < 0 || index >= r.frameQueryCount {
		return ReplayQueryInfo{}
	}

	q := &r.frameQueries[index]

	return ReplayQueryInfo{
		Type: ReplayQueryType(q.kind), Filter: q.filter, AABB: q.aabb, Origin: q.origin,
		Translation: q.translation, Shape: q.shape, HitCount: q.hitCount,
	}
}

// Get one result of a recorded query from the most recently replayed frame.
func (r *Replay) FrameQueryHit(queryIndex int, hitIndex int) ReplayQueryHit {
	if r == nil || queryIndex < 0 || queryIndex >= r.frameQueryCount {
		return ReplayQueryHit{}
	}

	q := &r.frameQueries[queryIndex]
	if hitIndex < 0 || hitIndex >= q.hitCount {
		return ReplayQueryHit{}
	}

	h := &r.frameHits[q.hitStart+hitIndex]

	return ReplayQueryHit{Shape: h.id, Point: h.point, Normal: h.normal, Fraction: h.fraction}
}

// Get the number of body slots tracked for the outliner. This is the creation-order span and
// includes holes for destroyed bodies, so it only grows as the replay advances.
func (r *Replay) BodyCount() int {
	if r == nil {
		return 0
	}

	return r.bodyIDCount
}

// Get a tracked body by creation ordinal. Returns the zero Body for a destroyed slot, an ordinal not
// yet reached at the current frame, or an out-of-range index. Validate with Body.IsValid.
func (r *Replay) Body(index int) Body {
	if r == nil || index < 0 || index >= r.bodyIDCount {
		return Body{}
	}

	return r.bodyIDs[index]
}

func (r *Replay) captureKeyframeIfDue() {
	if r.frame > r.lastKeyframeFrame && r.frame%r.keyframeInterval == 0 {
		recCaptureKeyframe(r)
	}
}

func (r *Replay) keyframeBefore(targetFrame int) *recKeyframe {
	var best *recKeyframe

	for i := range r.keyframeCount {
		kf := &r.keyframes[i]
		if kf.frame < targetFrame && (best == nil || kf.frame > best.frame) {
			best = kf
		}
	}

	return best
}

func (r *Replay) drawFrameQuery(q *recDrawQuery, draw *DebugDraw) {
	switch recQueryKind(q.kind) {
	case recqCastRay, recqCastRayClosest:
		// Ray origin to endpoint
		origin := q.origin
		end := origin.Add(q.translation)
		draw.DrawLine(origin, end, ColorYellow)
		// Per-hit point + short normal
		r.drawQueryHitNormals(q, draw, ColorYellow, ColorLightYellow)
	case recqCastShape:
		// Shape cast: draw per-hit points along the swept path
		r.drawQueryHitNormals(q, draw, ColorSkyBlue, ColorLightSkyBlue)
	case recqCastMover:
		// The mover capsule is relative to the query origin
		c1 := q.origin.Add(q.mover.Center1)
		c2 := q.origin.Add(q.mover.Center2)
		draw.DrawSolidCapsule(c1, c2, q.mover.Radius, ColorLightSkyBlue)
	case recqOverlapAABB:
		// The query box is relative to the query origin
		lower, upper := q.aabb.LowerBound, q.aabb.UpperBound
		vs := []Vec2{lower, {X: upper.X, Y: lower.Y}, upper, {X: lower.X, Y: upper.Y}}
		draw.DrawPolygon(WorldTransform{P: q.origin, Q: Rot{C: 1}}, vs, ColorLimeGreen)
		recDrawHitAABBs(r, q, draw)
	case recqOverlapShape:
		recDrawQueryProxy(q, draw)
		recDrawHitAABBs(r, q, draw)
	case recqCollideMover:
		// The mover capsule and the collision planes are relative to the query origin
		c1 := q.origin.Add(q.mover.Center1)
		c2 := q.origin.Add(q.mover.Center2)
		draw.DrawSolidCapsule(c1, c2, q.mover.Radius, ColorTan)
		r.drawQueryHitPlanes(q, draw)
	case recqShapeTestPoint:
		c := ColorRed
		if q.boolResult {
			c = ColorAqua
		}

		draw.DrawPoint(q.origin, 6, c)
	case recqShapeRayCast:
		origin := q.origin
		end := origin.Add(q.translation)
		draw.DrawLine(origin, end, ColorViolet)

		if q.castOut.Hit {
			draw.DrawPoint(q.castOut.Point, 4, ColorViolet)
		}
	}
}

func (r *Replay) drawQueryHitNormals(
	q *recDrawQuery,
	draw *DebugDraw,
	pointColor HexColor, normalColor HexColor,
) {
	for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
		h := &r.frameHits[hi]
		point := h.point
		draw.DrawPoint(point, 4, pointColor)
		np := point.Add(h.normal.Scale(0.2))
		draw.DrawLine(point, np, normalColor)
	}
}

func (r *Replay) drawQueryHitPlanes(q *recDrawQuery, draw *DebugDraw) {
	// Per-hit plane point and normal
	for hi := q.hitStart; hi < q.hitStart+q.hitCount; hi++ {
		h := &r.frameHits[hi]
		if h.plane.Hit {
			point := q.origin.Add(h.plane.Point)
			np := point.Add(h.plane.Plane.Normal.Scale(0.2))
			draw.DrawLine(point, np, ColorOrange)
		}
	}
}
